package main

import (
	"common"
	"container/list"
	"db"
	"excel"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strings"
	"time"
	"zeus/entity"
	"zeus/global"
	"zeus/timer"

	"github.com/cihub/seelog"
)

// MatchMgr 匹配管理器
type MatchMgr struct {
	entity.Entity

	*timer.Timer

	wsList map[uint32]*list.List

	maxTeamNum int

	expendTime int64
	expectTime int64

	scenesC chan *Scene
	index   uint32
}

// Init 初始化调用
func (mgr *MatchMgr) Init(initParam interface{}) {
	mgr.RegMsgProc(&MatchMgrMsgProc{mgr: mgr})
	mgr.wsList = make(map[uint32]*list.List)

	proxy := entity.NewEntityProxy(GetSrvInst().GetSrvID(), 0, mgr.GetID())
	entity.RegEntityProxy(fmt.Sprintf("%s:%d", initParam.(string), GetSrvInst().GetSrvID()), proxy)

	switch initParam {
	case common.MatchMgrSolo:
		mgr.maxTeamNum = 1
		mgr.expectTime = 10
		mgr.index = 10000
	case common.MatchMgrDuo:
		mgr.maxTeamNum = 2
		mgr.index = 20000
	case common.MatchMgrSquad:
		mgr.maxTeamNum = 4
		mgr.index = 30000
	}

	mgr.scenesC = make(chan *Scene, 1000)
	mgr.Timer = timer.NewTimer()
	mgr.RegTimer(mgr.printInfo, 5*time.Second)
}

// Destroy 框架层回调
func (mgr *MatchMgr) Destroy() {
	name := fmt.Sprintf("%s:%d", mgr.GetInitParam().(string), GetSrvInst().GetSrvID())
	global.GetGlobalInst().RemoveGlobal(name)
}

// Loop 逻辑帧
func (mgr *MatchMgr) Loop() {

	mgr.Timer.Loop()
	for _, l := range mgr.wsList {
		for e := l.Front(); e != nil; e = e.Next() {
			ws := e.Value.(IWaitingScene)

			if ws.IsWaitSceneInit() {
				continue
			}

			if ws.IsNeedRemove() {
				l.Remove(e)
				mgr.UnregTimerByObj(ws)
				continue
			}

			if ws.IsReady() {
				mgr.expendTime = ws.GetExpendTime()
				ws.Go()
			}
		}
	}

	for {
		select {
		case s := <-mgr.scenesC:
			mgr.initScene(s)
		default:
			return
		}
	}
}

func (mgr *MatchMgr) pushScene(scene *Scene) {
	select {
	case mgr.scenesC <- scene:
	default:
		seelog.Error("match manager had been block ", scene)
		return
	}
}

func (mgr *MatchMgr) initScene(scene *Scene) {
	for _, l := range mgr.wsList {
		for e := l.Front(); e != nil; e = e.Next() {
			ws := e.Value.(*WaitingScene)

			if ws.GetSpaceID() == scene.GetID() {
				l.Remove(e)
				mgr.UnregTimerByObj(ws)

				scene.objs = ws.objs
				scene.isSingleMatch = ws.singleMatch
				scene.teamType = ws.teamType
				scene.matchMode = ws.GetMatchMode()
				scene.uniqueId = ws.uniqueId
				scene.mapid = ws.mapid
				scene.haveai = ws.aiNum

				rand.Seed(time.Now().UnixNano())
				base, ok := excel.GetMaps(uint64(ws.mapid))
				if ok {
					randstr := strings.Split(base.Skybox, ",")
					if len(randstr) > 0 {
						index := rand.Intn(len(randstr))
						scene.skybox = common.StringToUint32(randstr[index])
					}
				}
				if scene.skybox == 0 {
					scene.skybox = 1
				}

				var mode uint32 = 0
				if ws.singleMatch {
					mode = 0
				} else if ws.teamType == 0 {
					mode = 1
				} else if ws.teamType == 1 {
					mode = 2
				}

				for _, obj := range ws.objs {
					obj.onMatchSuccess(scene.GetID(), scene.mapid, scene.skybox, mode, ws.getNum())
				}

				//待优化
				if !ws.singleMatch {
					GetTeamMgr().expendTime = ws.GetExpendTime()

					// 保存场景内所有队伍信息
					spaceTeam := &db.SpaceTeamInfo{
						Id:       scene.GetID(),
						Teamtype: uint32(ws.teamType),
					}

					for _, objs := range ws.objs {
						//玩家切换至房间中
						teamInfo := db.TeamInfo{
							Id: objs.GetID(),
						}

						teamInfo.MemList = append(teamInfo.MemList, objs.GetDBIDs()...)

						spaceTeam.Teams = append(spaceTeam.Teams, teamInfo)
					}

					db.SpaceTeamUtil(scene.GetID()).SaveSpaceTeamInfo(spaceTeam)

					//删除队伍
					//for _, objs := range ws.objs {
					//	GetTeamMgr().teams.Delete(objs.GetID())
					//}
				}

				return
			}
		}
	}
}

func (mgr *MatchMgr) getMatchScene(obj IMatcher) IWaitingScene {
	l, ok := mgr.wsList[obj.GetMapID()]
	if !ok {
		l = list.New()
		mgr.wsList[obj.GetMapID()] = l
	}

	// for e := l.Front(); e != nil; e = e.Next() {
	// 	ws := e.Value.(IWaitingScene)
	// 	if ws.IsMatching(obj) {
	// 		return ws
	// 	}
	// }

	var mss MatchingScenes
	for e := l.Front(); e != nil; e = e.Next() {
		ws := e.Value.(IWaitingScene)
		degree := ws.GetMatchingDegree(obj)
		if degree != 0 {
			mss = append(mss, &MatchingScene{
				IWaitingScene:  ws,
				matchingDegree: degree,
			})
		}
	}

	mssLen := mss.Len()
	if mssLen != 0 {
		if mssLen == 1 {
			return mss[0].IWaitingScene
		}

		sort.Sort(mss)
		ms := mss[mssLen-1]
		return ms.IWaitingScene
	}

	return mgr.newScene(obj.GetMapID(), obj.IsSingleMatch(), obj.GetTeamType(), obj.GetMatchMode())
}

func (mgr *MatchMgr) newScene(mapid uint32, singleMatch bool, teamType uint8, matchMode uint32) IWaitingScene {
	var minWait, maxWait int64
	if matchMode == common.MatchModeArcade {
		minWait = int64(common.GetTBSystemValue(common.System_ArcadeMatchMinTime))
		maxWait = int64(common.GetTBSystemValue(common.System_ArcadeMatchMaxTime))
	} else if matchMode == common.MatchModeTankWar {
		minWait = int64(common.GetTBSystemValue(common.System_TankWarMatchMinTime))
		maxWait = int64(common.GetTBSystemValue(common.System_TankWarMatchMaxTime))
	} else {
		minWait = int64(common.GetTBSystemValue(common.System_MatchMinWait))
		maxWait = int64(common.GetTBSystemValue(common.System_MatchWait))
	}

	if matchMode == common.MatchModeNormal && mapid == 2 {
		minWait = int64(common.GetTBSystemValue(4004))
		maxWait = int64(common.GetTBSystemValue(4002))
	} else if matchMode == common.MatchModeEndless && mapid == 7 {
		minWait = int64(common.GetTBSystemValue(4010))
		maxWait = int64(common.GetTBSystemValue(4008))
	}

	min := mgr.getMinMatchSum(mapid, matchMode)
	max := mgr.getMaxMatchSum(mapid, matchMode)
	ws := NewWaitingScene(min, max, minWait, maxWait, mapid, singleMatch, teamType, matchMode)
	l, ok := mgr.wsList[mapid]
	if !ok {
		l = list.New()
		mgr.wsList[mapid] = l
	}
	l.PushBack(ws)

	mgr.index++
	ws.index = mgr.index

	// 注册通知人数定时器
	mgr.RegTimerByObj(ws, ws.AiSpeed, 1*time.Second) //定时提升ai速度
	mgr.RegTimerByObj(ws, ws.addAiNum, 5*time.Second)
	mgr.RegTimerByObj(ws, ws.NotifyWaitingNums, 5*time.Second)
	// mgr.RegTimerByObj(ws, ws.PrintInfo, 5*time.Second)
	return ws
}

func (mgr *MatchMgr) getMinMatchSum(mapid, matchMode uint32) uint32 {
	var sysID uint64
	if matchMode == common.MatchModeArcade {
		sysID = common.System_ArcadeMatchMinSum
	} else if matchMode == common.MatchModeTankWar {
		sysID = common.System_TankWarMatchMinSum
	} else {
		sysID = common.System_MatchMinSum
	}

	if matchMode == common.MatchModeNormal && mapid == 2 {
		sysID = 4003
	} else if matchMode == common.MatchModeEndless && mapid == 7 {
		sysID = 4009
	}

	minSum := common.GetTBSystemValue(sysID)
	if minSum == 0 {
		return 1
	}

	if mgr.GetInitParam() == common.MatchMgrSolo {
		return uint32(minSum)
	}

	var ret float64
	if mgr.GetInitParam() == common.MatchMgrDuo {
		ret = float64(minSum) / 2
	} else if mgr.GetInitParam() == common.MatchMgrSquad {
		ret = float64(minSum) / 4
	}

	return uint32(math.Floor(ret))
}

// 获取最大匹配成员数
func (mgr *MatchMgr) getMaxMatchSum(mapid, matchMode uint32) uint32 {
	var sysID uint64
	if matchMode == common.MatchModeArcade {
		sysID = common.System_ArcadeMatchMaxSum
	} else if matchMode == common.MatchModeTankWar {
		sysID = common.System_TankWarMatchMaxSum
	} else {
		sysID = common.System_RoomUserLimit
	}

	if matchMode == common.MatchModeNormal && mapid == 2 {
		sysID = 4001
	} else if matchMode == common.MatchModeEndless && mapid == 7 {
		sysID = 4007
	}

	maxSum := common.GetTBSystemValue(sysID)
	if maxSum == 0 {
		return 1
	}

	if mgr.GetInitParam() == common.MatchMgrSolo {
		return uint32(maxSum)
	}

	var ret float64
	if mgr.GetInitParam() == common.MatchMgrDuo {
		ret = float64(maxSum) / 2
	} else if mgr.GetInitParam() == common.MatchMgrSquad {
		ret = float64(maxSum) / 4
	}

	return uint32(math.Floor(ret))
}

func (mgr *MatchMgr) calExpectTime() uint64 {
	var agingFactor float64 = float64(common.GetTBSystemValue(19)) / 100
	mgr.expectTime = int64(float64(mgr.expectTime)*agingFactor + (1-agingFactor)*float64(mgr.expendTime))
	if mgr.expectTime == 0 {
		mgr.expectTime = 1
	}

	return uint64(mgr.expectTime)
}

func (mgr *MatchMgr) printInfo() {
	for _, l := range mgr.wsList {
		for e := l.Front(); e != nil; e = e.Next() {
			ws := e.Value.(IWaitingScene)
			ws.PrintInfo()
		}
	}
}

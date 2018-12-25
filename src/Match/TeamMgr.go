package main

import (
	"common"
	"container/list"
	"fmt"
	"sync"
	"zeus/entity"
	"zeus/global"
	"zeus/iserver"

	"db"

	log "github.com/cihub/seelog"
)

const (
	// TwoTeamType 双排
	TwoTeamType uint8 = 0
	// FourTeamType 四排
	FourTeamType uint8 = 1
)

const (
	// TeamNotMatch 队伍在组队中
	TeamNotMatch = 0
	// TeamMatching 队伍在匹配中
	TeamMatching = 1
	// TeamWatingScene 准备进入地图
	TeamWatingScene = 2
)

const (
	// MemberNotReady 成员未准备
	MemberNotReady = 0
	// MemberReadying 成员准备中
	MemberReadying = 1
)

// TeamMgr 组队管理器
type TeamMgr struct {
	entity.Entity

	teams   sync.Map
	cpLists map[uint32]*list.List //每种匹配模式对应一个key

	expendTime int64
	expectTime int64
}

//组队管理器
var teamMgrPtr *TeamMgr

// GetTeamMgr 获取组队管理器
func GetTeamMgr() *TeamMgr {
	return teamMgrPtr
}

// Init 初始化
func (mgr *TeamMgr) Init(initParam interface{}) {
	mgr.RegMsgProc(&TeamMgrMsgProc{mgr: mgr})
	mgr.cpLists = make(map[uint32]*list.List)
	mgr.expectTime = 10

	proxy := entity.NewEntityProxy(GetSrvInst().GetSrvID(), 0, mgr.GetID())
	entity.RegEntityProxy(fmt.Sprintf("%s:%d", initParam.(string), GetSrvInst().GetSrvID()), proxy)

	teamMgrPtr = mgr

	log.Info("TeamMgr Init ", initParam)
}

// Destroy 框架层回调
func (mgr *TeamMgr) Destroy() {
	//清理redis数据
	mgr.teams.Range(func(key, value interface{}) bool {
		//mgr.RemoveTeam(key.(uint64))
		mgr.teams.Delete(key)
		db.PlayerTeamUtil(key.(uint64)).Remove()
		return true
	})

	name := fmt.Sprintf("%s:%d", mgr.GetInitParam().(string), GetSrvInst().GetSrvID())
	global.GetGlobalInst().RemoveGlobal(name)
}

// boradExpectTime 广播期望匹配时间
func (mgr *TeamMgr) boradExpectTime(tmModelPtr *MatchTeam) {
	if tmModelPtr == nil {
		return
	}

	var agingFactor float64 = float64(common.GetTBSystemValue(19)) / 100
	mgr.expectTime = int64(float64(mgr.expectTime)*agingFactor + (1-agingFactor)*float64(mgr.expendTime))
	if mgr.expectTime == 0 {
		mgr.expectTime = 1
	}

	// for _, pUser := range tmModelPtr.members {
	// 	if pUser == nil {
	// 		continue
	// 	}

	// 	pUser.RPC(iserver.ServerTypeClient, "ExpectTime", uint64(mgr.expectTime))
	// }

	tmModelPtr.members.Range(
		func(k, v interface{}) bool {
			pUser := v.(*MatchMember)
			pUser.RPC(iserver.ServerTypeClient, "ExpectTime", uint64(mgr.expectTime))
			return true
		})
}

func (mgr *TeamMgr) addMatch(team *MatchTeam) {
	if team.teamType == TwoTeamType {
		name := fmt.Sprintf("%s:%d", common.MatchMgrDuo, GetSrvInst().GetSrvID())
		match := entity.GetEntityProxy(name)
		if match == nil {
			log.Warn("获取匹配管理器失败!")
			return
		}

		if err := match.RPC(common.ServerTypeMatch, "EnterDuoQueue", team.teamid); err != nil {
			log.Error(err)
		}
	} else if team.teamType == FourTeamType {
		name := fmt.Sprintf("%s:%d", common.MatchMgrSquad, GetSrvInst().GetSrvID())
		match := entity.GetEntityProxy(name)
		if match == nil {
			log.Warn("获取匹配管理器失败!")
			return
		}

		if err := match.RPC(common.ServerTypeMatch, "EnterDuoQueue", team.teamid); err != nil {
			log.Error(err)
		}
	}
}

func (mgr *TeamMgr) delMatch(team *MatchTeam) {
	if team.teamType == TwoTeamType {
		name := fmt.Sprintf("%s:%d", common.MatchMgrDuo, GetSrvInst().GetSrvID())
		proxy := entity.GetEntityProxy(name)
		if err := proxy.RPC(common.ServerTypeMatch, "CancelQueue", team.teamid); err != nil {
			log.Error(err)
		}
	} else if team.teamType == FourTeamType {
		name := fmt.Sprintf("%s:%d", common.MatchMgrSquad, GetSrvInst().GetSrvID())
		proxy := entity.GetEntityProxy(name)
		if err := proxy.RPC(common.ServerTypeMatch, "CancelQueue", team.teamid); err != nil {
			log.Error(err)
		}
	}
}

func (mgr *TeamMgr) Loop() {

	for _, list := range mgr.cpLists {

		twoteam_1 := make(map[uint32][]uint64)
		fourteam_1 := make(map[uint32][]uint64)
		fourteam_2 := make(map[uint32][]uint64)
		fourteam_3 := make(map[uint32][]uint64)
		dellist := make(map[uint64]bool)

		for e := list.Front(); e != nil; e = e.Next() {
			team := e.Value.(*MatchTeam)

			if team.teamStatus != TeamMatching {
				list.Remove(e)
				continue
			}

			if team.teamType == TwoTeamType {
				// if len(team.members) == 1 {
				if team.GetNums() == 1 {
					twoteam_1[team.mapid] = append(twoteam_1[team.mapid], team.teamid)
				}
			} else if team.teamType == FourTeamType {
				// if len(team.members) == 1 {
				if team.GetNums() == 1 {
					fourteam_1[team.mapid] = append(fourteam_1[team.mapid], team.teamid)
				} else if team.GetNums() == 2 {
					fourteam_2[team.mapid] = append(fourteam_2[team.mapid], team.teamid)
				} else if team.GetNums() == 3 {
					fourteam_3[team.mapid] = append(fourteam_3[team.mapid], team.teamid)
				}
			}
		}

		for _, v := range twoteam_1 {
			for i := 0; i < len(v); i += 2 {
				if i+1 >= len(v) {
					break
				}

				dellist[v[i]] = true
				dellist[v[i+1]] = true
				mgr.compose(v[i], v[i+1])
			}
		}

		fourteam_22 := make(map[uint32]uint64)
		for mapid, v := range fourteam_2 {
			for i := 0; i < len(v); i += 2 {
				if i+1 >= len(v) {
					fourteam_22[mapid] = v[i]
					break
				}

				dellist[v[i]] = true
				dellist[v[i+1]] = true
				mgr.compose(v[i], v[i+1])
			}
		}

		fourteam_11 := make(map[uint32][]uint64)
		for mapid, v := range fourteam_1 {
			other, ok := fourteam_3[mapid]
			if ok {
				for i := 0; i < len(v) && i < len(other); i++ {
					dellist[v[i]] = true
					dellist[other[i]] = true
					mgr.compose(v[i], other[i])
				}

				if len(v) > len(other) {
					fourteam_11[mapid] = append(fourteam_11[mapid], v[len(other):]...)
				}
			} else {
				fourteam_11[mapid] = append(fourteam_11[mapid], v...)
			}
		}

		fourteam_111 := make(map[uint32][]uint64)
		for mapid, v := range fourteam_11 {
			for i := 0; i < len(v); i += 4 {
				if i+1 >= len(v) {
					fourteam_111[mapid] = append(fourteam_111[mapid], v[i])
					break
				}
				if i+2 >= len(v) {
					fourteam_111[mapid] = append(fourteam_111[mapid], v[i])
					fourteam_111[mapid] = append(fourteam_111[mapid], v[i+1])
					break
				}
				if i+3 >= len(v) {
					fourteam_111[mapid] = append(fourteam_111[mapid], v[i])
					fourteam_111[mapid] = append(fourteam_111[mapid], v[i+1])
					fourteam_111[mapid] = append(fourteam_111[mapid], v[i+2])
					break
				}

				dellist[v[i]] = true
				dellist[v[i+1]] = true
				dellist[v[i+2]] = true
				dellist[v[i+3]] = true
				mgr.composeFour(v[i], v[i+1], v[i+2], v[i+3])
			}
		}

		for mapid, v := range fourteam_22 {
			other, ok := fourteam_111[mapid]
			if ok {
				if len(other) < 2 {
					break
				}

				dellist[v] = true
				dellist[other[0]] = true
				dellist[other[1]] = true
				mgr.composeThree(v, other[0], other[1])
			}
		}

		for e := list.Front(); e != nil; e = e.Next() {
			team := e.Value.(*MatchTeam)

			if _, ok := dellist[team.teamid]; ok {
				list.Remove(e)
			}
		}
	}
}

func (mgr *TeamMgr) compose(team1, team2 uint64) uint64 {
	t1, ok1 := mgr.teams.Load(team1)
	t2, ok2 := mgr.teams.Load(team2)
	if !ok1 || !ok2 {
		return 0
	}

	t11 := t1.(*MatchTeam)
	t22 := t2.(*MatchTeam)

	// for k, v := range t22.members {
	// 	t11.members[k] = v
	// }
	t11.members.Range(
		func(k, v interface{}) bool {
			mm := v.(*MatchMember)
			if mm.stateEnterTeam == 0 {
				mm.stateEnterTeam = 1 //系统自动分配补充进入队伍
			}
			return true
		})
	t22.members.Range(
		func(k, v interface{}) bool {
			mm := v.(*MatchMember)
			mm.stateEnterTeam = 1 //系统自动分配补充进入队伍
			mm.originalTeam = t22.GetID()
			t11.Add(mm)
			return true
		})
	t22.teamStatus = TeamNotMatch

	mgr.RemoveTeam(team2)
	t11.BroadInfo()
	t11.calcAvgMMR()
	// mgr.boradExpectTime(t11)

	t11.teamStatus = TeamWatingScene
	mgr.addMatch(t11)
	return team1
}

func (mgr *TeamMgr) composeThree(team1, team2, team3 uint64) uint64 {
	t1, ok1 := mgr.teams.Load(team1)
	t2, ok2 := mgr.teams.Load(team2)
	t3, ok3 := mgr.teams.Load(team3)

	if !ok1 || !ok2 || !ok3 {
		return 0
	}

	t11 := t1.(*MatchTeam)
	t22 := t2.(*MatchTeam)
	t33 := t3.(*MatchTeam)

	// for k, v := range t22.members {
	// 	t11.members[k] = v
	// }
	// for k, v := range t33.members {
	// 	t11.members[k] = v
	// }
	t11.members.Range(
		func(k, v interface{}) bool {
			mm := v.(*MatchMember)
			if mm.stateEnterTeam == 0 {
				mm.stateEnterTeam = 1 //系统自动分配补充进入队伍
			}
			return true
		})
	t22.members.Range(
		func(k, v interface{}) bool {
			mm := v.(*MatchMember)
			mm.stateEnterTeam = 1 //系统自动分配补充进入队伍
			mm.originalTeam = t22.GetID()
			t11.Add(mm)
			return true
		})
	t22.teamStatus = TeamNotMatch
	t33.members.Range(
		func(k, v interface{}) bool {
			mm := v.(*MatchMember)
			mm.stateEnterTeam = 1 //系统自动分配补充进入队伍
			mm.originalTeam = t33.GetID()
			t11.Add(mm)
			return true
		})
	t33.teamStatus = TeamNotMatch

	mgr.RemoveTeam(team2)
	mgr.RemoveTeam(team3)

	t11.BroadInfo()
	t11.calcAvgMMR()
	// mgr.boradExpectTime(t11)

	t11.teamStatus = TeamWatingScene
	mgr.addMatch(t11)
	return team1
}

func (mgr *TeamMgr) composeFour(team1, team2, team3, team4 uint64) uint64 {
	t1, ok1 := mgr.teams.Load(team1)
	t2, ok2 := mgr.teams.Load(team2)
	t3, ok3 := mgr.teams.Load(team3)
	t4, ok4 := mgr.teams.Load(team4)
	if !ok1 || !ok2 || !ok3 || !ok4 {
		return 0
	}

	t11 := t1.(*MatchTeam)
	t22 := t2.(*MatchTeam)
	t33 := t3.(*MatchTeam)
	t44 := t4.(*MatchTeam)

	// for k, v := range t22.members {
	// 	t11.members[k] = v
	// }
	// for k, v := range t33.members {
	// 	t11.members[k] = v
	// }
	// for k, v := range t44.members {
	// 	t11.members[k] = v
	// }
	t11.members.Range(
		func(k, v interface{}) bool {
			mm := v.(*MatchMember)
			if mm.stateEnterTeam == 0 {
				mm.stateEnterTeam = 1 //系统自动分配补充进入队伍
			}
			return true
		})
	t22.members.Range(
		func(k, v interface{}) bool {
			mm := v.(*MatchMember)
			mm.stateEnterTeam = 1 //系统自动分配补充进入队伍
			mm.originalTeam = t22.GetID()
			t11.Add(mm)
			return true
		})
	t22.teamStatus = TeamNotMatch
	t33.members.Range(
		func(k, v interface{}) bool {
			mm := v.(*MatchMember)
			mm.stateEnterTeam = 1 //系统自动分配补充进入队伍
			mm.originalTeam = t33.GetID()
			t11.Add(mm)
			return true
		})
	t33.teamStatus = TeamNotMatch
	t44.members.Range(
		func(k, v interface{}) bool {
			mm := v.(*MatchMember)
			mm.stateEnterTeam = 1 //系统自动分配补充进入队伍
			mm.originalTeam = t44.GetID()
			t11.Add(mm)
			return true
		})
	t44.teamStatus = TeamNotMatch

	mgr.RemoveTeam(team2)
	mgr.RemoveTeam(team3)
	mgr.RemoveTeam(team4)

	t11.BroadInfo()
	t11.calcAvgMMR()
	// mgr.boradExpectTime(t11)

	t11.teamStatus = TeamWatingScene
	mgr.addMatch(t11)
	return team1
}

// RemoveTeam 移除队伍
func (mgr *TeamMgr) RemoveTeam(teamID uint64) {
	//mgr.teams.Delete(teamID)
	//db.PlayerTeamUtil(teamID).Remove()
}

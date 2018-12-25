package main

import (
	"common"
	"excel"
	"fmt"
	"math"
	"time"

	log "github.com/cihub/seelog"
)

// IWaitingScene 等待状态的房间接口
type IWaitingScene interface {
	Add(IMatcher)
	Remove(IMatcher)
	Get(uint64) IMatcher
	// IsMatching(IMatcher) bool
	GetMatchingDegree(IMatcher) uint32
	IsReady() bool
	IsNeedRemove() bool
	Go()
	GetCreateTime() int64
	GetExpendTime() int64
	IsWaitSceneInit() bool
	GetSpaceID() uint64
	GetMatchMode() uint32
	PrintInfo()
}

// WaitingScene 等待状态的房间
type WaitingScene struct {
	objs map[uint64]IMatcher

	start   int64
	maxWait int64
	minWait int64
	maxNum  uint32
	minNum  uint32

	avgMMR  uint32
	minMMR  uint32
	maxMMR  uint32
	avgRank uint32
	minRank uint32
	maxRank uint32
	mapid   uint32

	// 匹配范围相关
	initRange uint32
	initWait  int64
	timeout   int64 // 超时时间
	waitRange uint32
	initPow   float32
	maxExpand uint32

	singleMatch bool
	teamType    uint8
	matchMode   uint32 //匹配模式，具体定义在common/Const.go
	uniqueId    uint32 //matchMode表中对应匹配模式的唯一id

	spaceID uint64
	index   uint32

	//AI
	aiSpeedNext  int64   //提升速度的时间
	aiSpeed      float32 //ai速度
	aiNum        uint32  //ai数量
	aiMaxNum     uint32  //最大ai数量
	aiMaxSpeed   float32 //最大速度
	aiSpeedAdd   float32 //ai速度修正量
	aiSpeedStart int64   //ai开始修正速度的时间
	aiExtra      float32 //每次速度增量的小数累计

}

// NewWaitingScene 创建一个空房间
func NewWaitingScene(minNum, maxNum uint32, minWait, maxWait int64, mapid uint32, singleMatch bool, teamType uint8, matchMode uint32) *WaitingScene {
	ws := &WaitingScene{}
	ws.objs = make(map[uint64]IMatcher)
	ws.start = time.Now().Unix()
	ws.minWait = minWait
	ws.maxWait = maxWait
	ws.minNum = minNum
	ws.maxNum = maxNum
	ws.minMMR = math.MaxUint32
	ws.mapid = mapid
	ws.singleMatch = singleMatch
	ws.teamType = teamType
	ws.spaceID = 0
	ws.matchMode = matchMode

	ws.initRange = uint32(common.GetTBMatchValue(common.Match_InitRange))
	ws.initWait = int64(common.GetTBMatchValue(common.Match_InitWait))
	ws.timeout = int64(common.GetTBMatchValue(common.Match_Timeout))
	ws.waitRange = uint32(common.GetTBMatchValue(common.Match_WaitRange))
	ws.initPow = common.GetTBMatchValue(common.Match_InitPow)
	ws.maxExpand = uint32(common.GetTBMatchValue(common.Match_MaxExpand))

	indexId := uint64(ws.matchMode * 10)
	if ws.singleMatch {
		indexId += 1
	} else if ws.teamType == 0 {
		indexId += 2
	} else if ws.teamType == 1 {
		indexId += 4
	}
	excelData, ok := excel.GetAi_spawn(indexId)
	if !ok {
		excelData, ok = excel.GetAi_spawn(indexId % 10)

	}
	if ok {
		ws.aiSpeedStart = int64(excelData.Timetobeginacc)
		ws.aiSpeed = excelData.Vbase
		ws.aiSpeedAdd = excelData.Vadd
		ws.aiMaxSpeed = excelData.Vmax
		ws.aiMaxNum = uint32(excelData.N_ai_max)
		log.Debug("waiting scene init ", "start ", ws.aiSpeedStart, " init speed ", ws.aiSpeed, " add ", ws.aiSpeedAdd, " max ", ws.aiMaxSpeed, " max num ", ws.aiMaxNum)
	}

	return ws
}

// GetCreateTime 获取房间创建时间
func (ws *WaitingScene) GetCreateTime() int64 {
	return ws.start
}

// IsWaitSceneInit 场景是否初始化
func (ws *WaitingScene) IsWaitSceneInit() bool {
	return ws.spaceID != 0
}

// GetSpaceID 获取场景SpaceID
func (ws *WaitingScene) GetSpaceID() uint64 {
	return ws.spaceID
}

// GetMatchMode 获取匹配模式
func (ws *WaitingScene) GetMatchMode() uint32 {
	return ws.matchMode
}

// Add 加入等待房间
func (ws *WaitingScene) Add(obj IMatcher) {
	_, ok := ws.objs[obj.GetID()]
	if ok {
		log.Warn("已在等待队列中 ", obj.GetID())
		return
	}

	if ws.GetMatchMode() != obj.GetMatchMode() {
		log.Error("ws's matchMode is ", ws.GetMatchMode(), " while new comer's is ", obj.GetMatchMode())
		return
	}
	// l := uint32(len(ws.objs))
	// if l > ws.maxNum {
	// 	log.Warn("等待队列已满!")
	// 	return
	// }

	ws.objs[obj.GetID()] = obj
	ws.NotifyWaitingNums()

	/*
		mmr := obj.GetMMR()
		if mmr < ws.minMMR {
			ws.minMMR = mmr
		}
		if mmr > ws.maxMMR {
			ws.maxMMR = mmr
		}

		ws.avgMMR = (ws.avgMMR*l + mmr) / (l + 1)
	*/
	// l := uint32(len(ws.objs))
	rank := obj.GetRank()
	// if rank < ws.minRank {
	// 	ws.minRank = rank
	// }
	// if rank > ws.maxRank {
	// 	ws.maxRank = rank
	// }

	// var totalRank uint32
	// for _, o := range ws.objs {
	// 	totalRank += o.GetRank()
	// }

	// if l != 0 {
	// 	ws.avgRank = totalRank / l
	// } else {
	// 	ws.avgRank = rank
	// }

	//取第一个进入作为平均分
	if ws.avgRank == 0 {
		ws.avgRank = rank
	}

	// ws.avgRank = (ws.avgRank*l + rank) / (l + 1)
	// log.Error(ws.start, "新加入计算平均分数:", ws.avgRank, " 加入分数:", rank)
}

// Remove 移除
func (ws *WaitingScene) Remove(obj IMatcher) {
	_, ok := ws.objs[obj.GetID()]
	if !ok {
		log.Warn("不在等待队列中 ", obj.GetID())
		return
	}

	delete(ws.objs, obj.GetID())

	/*
		var totalMMR uint32
		for _, o := range ws.objs {
			totalMMR += o.GetMMR()
		}

		if len(ws.objs) != 0 {
			ws.avgMMR = totalMMR / uint32(len(ws.objs))
		}
	*/
	// var totalRank uint32
	// for _, o := range ws.objs {
	// 	totalRank += o.GetRank()
	// }

	// if len(ws.objs) != 0 {
	// 	ws.avgRank = totalRank / uint32(len(ws.objs))
	// } else {
	// 	ws.avgRank = 0
	// }

	// log.Debug(ws.start, "离开计算平均分数:", ws.avgRank, " 离开的分数:", obj.GetRank())
}

// Get 获取匹配对象
func (ws *WaitingScene) Get(id uint64) IMatcher {
	return ws.objs[id]
}

// GetMatchingDegree 获取匹配程度
func (ws *WaitingScene) GetMatchingDegree(obj IMatcher) uint32 {
	if ws.mapid != obj.GetMapID() {
		return 0
	}

	if ws.GetMatchMode() != obj.GetMatchMode() {
		return 0
	}

	if ws.spaceID != 0 {
		return 0
	}

	if ws.singleMatch {
		if !obj.IsSingleMatch() {
			return 0
		}
	} else {
		if obj.IsSingleMatch() {
			return 0
		}

		if ws.teamType != obj.GetTeamType() {
			return 0
		}
	}

	if obj.IsSingleMatch() {
		sys, ok := excel.GetSystem(common.System_RoomUserLimit)
		if !ok {
			return 0
		}

		if ws.getNum() >= uint32(sys.Value) {
			return 0
		}
	} else {
		if uint32(len(ws.objs))+ws.aiNumToObjs() >= ws.maxNum {
			return 0
		}
	}

	if ws.GetMatchMode() >= 1 && ws.GetMatchMode() <= 10 {
		return 100
	}

	// 匹配程度计算规则
	// 0为不匹配, 分数越高越匹配, 正好匹配上为100, 超过最大等待时间后匹配上为200, 每扩展一次匹配上减1
	rank := obj.GetRank()
	var diff uint32 //匹配分差值
	if rank > ws.avgRank {
		diff = rank - ws.avgRank
	} else {
		diff = ws.avgRank - rank
	}

	dur := time.Now().Unix() - ws.start
	// 不扩展或者扩展时间未到
	if ws.initWait == 0 || dur <= ws.initWait {
		if diff <= ws.initRange {
			return 100
		}
		return 0
	}

	times := uint32(dur / ws.initWait)
	if times > ws.maxExpand {
		times = ws.maxExpand
	}

	var rankRange uint32

	// 时间超过设定的最大值
	if dur >= ws.timeout {
		rankRange = ws.getExpandRange(times)
		if diff <= rankRange {
			return 200
		}
		return 0
	}

	var i uint32
	for i = 0; i <= times; i++ {
		rankRange = ws.getExpandRange(i)
		if diff <= rankRange {
			return 100 - i
		}
	}

	return 0
}

// IsMatching 匹配对象是否符合此房间要求
// func (ws *WaitingScene) IsMatching(obj IMatcher) bool {
// 	if ws.mapid != obj.GetMapID() {
// 		return false
// 	}

// 	if ws.GetMatchMode() != obj.GetMatchMode() {
// 		return false
// 	}

// 	if ws.spaceID != 0 {
// 		return false
// 	}

// 	if ws.singleMatch {
// 		if !obj.IsSingleMatch() {
// 			return false
// 		}
// 	} else {
// 		if obj.IsSingleMatch() {
// 			return false
// 		}

// 		if ws.teamType != obj.GetTeamType() {
// 			return false
// 		}
// 	}

// 	if obj.IsSingleMatch() {
// 		sys, ok := excel.GetSystem(common.System_RoomUserLimit)
// 		if !ok {
// 			return false
// 		}

// 		if ws.getNum() >= uint32(sys.Value) {
// 			return false
// 		}
// 	} else {
// 		if uint32(len(ws.objs)) >= ws.maxNum {
// 			return false
// 		}
// 	}

// 	if ws.GetMatchMode() >= 1 && ws.GetMatchMode() <= 10 {
// 		return true
// 	}

// 	// TODO 匹配规则
// 	waitTime := time.Now().Unix() - ws.start

// 	rankRange := ws.getRankRange(waitTime)
// 	rank := obj.GetRank()
// 	var diff uint32
// 	if rank > ws.avgRank {
// 		diff = rank - ws.avgRank
// 	} else {
// 		diff = ws.avgRank - rank
// 	}

// 	ok := false
// 	if diff <= rankRange {
// 		ok = true
// 	}

// 	log.Debugf("平均分数为%d的新队伍尝试加入房间%d 房间队伍数: %d 总人数: %d 平均分：%d 等待时间：%d 允许加入分数范围：%d--%d 新队伍的人数：%d 是否可加入房间：%t\n", rank, ws.index, uint32(len(ws.objs)), ws.getNum(), ws.avgRank, waitTime, ws.avgRank-rankRange, ws.avgRank+rankRange, obj.GetNums(), ok)

// 	return ok
// }

func (ws *WaitingScene) getTypeStr() string {
	if ws.singleMatch {
		return "单排"
	}

	if ws.teamType == TwoTeamType {
		return "双排"
	}

	return "四排"
}

func (ws *WaitingScene) PrintInfo() {
	l := uint32(len(ws.objs))
	dur := time.Now().Unix() - ws.start
	rng := ws.getRankRange(dur)
	max := ws.avgRank + rng
	min := uint32(0)

	if ws.avgRank > rng {
		min = ws.avgRank - rng
	}

	log.Debugf("模式: %s 房间编号: %d 队伍数: %d 平均分数: %d 等待时间: %d 上下分数范围: %d 允许加入分数范围: %d---%d\n", ws.getTypeStr(), ws.index, l, ws.avgRank, dur, rng, min, max)
	log.Debug("人数:", ws.getNum(), " 各人的分数", ws.getScoreStr())
}

// IsReady 检查房间状态, 符合条件就开始比赛
func (ws *WaitingScene) IsReady() bool {

	waitTime := time.Now().Unix() - ws.start
	total := uint32(len(ws.objs)) + ws.aiNumToObjs()

	// 人数为0时返回
	if total == 0 {
		return false
	}

	// 达到最大人数立即开始比赛
	if total >= ws.maxNum {
		ws.checkMaxNum()
		return true
	}

	// 达到最大等待时间立即开始比赛
	if waitTime >= ws.maxWait {
		return true
	}

	// 到达开启时间且满足最小开启人数则立刻开
	if ws.getNum() >= ws.minNum && waitTime >= ws.minWait {
		return true
	}

	// total := uint32(len(ws.objs))
	// if total < ws.minNum {
	// 	return false
	// } else if total >= ws.maxNum {
	// 	return true
	// }

	return false
}

// IsNeedRemove 检查房间状态, 是否需要删除
func (ws *WaitingScene) IsNeedRemove() bool {
	// waitTime := time.Now().Unix() - ws.start
	// if waitTime < ws.maxWait {
	// 	return false
	// }

	//强制删除
	if time.Now().Unix() > ws.start+ws.maxWait+180 {
		log.Error("异常强制删除ws:", ws.spaceID)
		return true
	}

	if ws.getTrueNum() > 0 {
		return false
	}

	return true
}

// GetExpendTime 获取预计等待时间
func (ws *WaitingScene) GetExpendTime() int64 {
	var ret int64
	curTime := time.Now().Unix()
	for _, obj := range ws.objs {
		ret += curTime - obj.GetMatchTime()
	}

	if len(ws.objs) != 0 {
		ret = ret / int64(len(ws.objs))
	}

	return ret
}

// Go 等待结束, 比赛开始
func (ws *WaitingScene) Go() {
	str := ""
	info := &common.ModeInfo{}

	if ws.singleMatch {
		str = common.MatchMgrSolo
		info = common.GetOpenModeInfo(ws.GetMatchMode(), 1)
	} else {
		if ws.teamType == TwoTeamType {
			str = common.MatchMgrDuo
			info = common.GetOpenModeInfo(ws.GetMatchMode(), 2)
		} else if ws.teamType == FourTeamType {
			str = common.MatchMgrSquad
			info = common.GetOpenModeInfo(ws.GetMatchMode(), 4)
		}
	}

	if info != nil {
		ws.uniqueId = info.UniqueId
	}

	initStr := fmt.Sprintf("%d:%d:%d:%s:%d", ws.mapid, ws.uniqueId, ws.GetMatchMode(), str, GetSrvInst().GetSrvID())
	// log.Error("开启一局比赛!", initStr)

	e, err := GetSrvInst().CreateEntityAll("Space", 0, initStr, false)
	if err != nil {
		log.Error(err)
		return
	}

	ws.spaceID = e.GetID()
	log.Info("Create space entity, mapid: ", ws.mapid, " uniqueId: ", ws.uniqueId, " matchMode: ", ws.GetMatchMode(), " matchTyp: ", str)
}

// NotifyWaitingNums 广播当前等待人数
func (ws *WaitingScene) NotifyWaitingNums() {
	totalNum := ws.getNum()
	for _, obj := range ws.objs {
		if err := obj.RPC(common.ServerTypeLobby, "NotifyWaitingNums", totalNum); err != nil {
			log.Error(err)
		}
	}
}

// 根据等待时间调整mmr范围
// TODO 添加更多规则
func (ws *WaitingScene) getMMRRange(dur int64) uint32 {
	return uint32(dur)
}

// 获取扩展i次之后的匹配范围
func (ws *WaitingScene) getExpandRange(i uint32) uint32 {
	return ws.waitRange*uint32(math.Pow(float64(i*uint32(ws.initWait)), float64(ws.initPow)))/uint32(ws.initWait) + ws.initRange
}

func (ws *WaitingScene) getRankRange(dur int64) uint32 {
	//勇者模式
	if ws.GetMatchMode() == common.MatchModeBrave {
		var modeInfo *common.ModeInfo

		if ws.singleMatch {
			modeInfo = common.GetOpenModeInfo(ws.GetMatchMode(), 1)
		} else {
			modeInfo = common.GetOpenModeInfo(ws.GetMatchMode(), (ws.teamType+1)*2)
		}

		if modeInfo == nil {
			log.Error("Failed to get opened brave mode info")
			return ws.initRange
		}

		ws.initRange = 0
		ws.initWait = int64(modeInfo.WaitTime)
		ws.waitRange = 1
	}

	if ws.initWait == 0 {
		return ws.initRange
	}

	if dur <= ws.initWait {
		return ws.initRange
	}

	times := uint32(dur / ws.initWait)

	if ws.GetMatchMode() == common.MatchModeBrave {
		return times*ws.waitRange + ws.initRange
	}

	if times > ws.maxExpand {
		times = ws.maxExpand
	}

	return ws.getExpandRange(times)
}

func (ws *WaitingScene) getTrueNum() uint32 {
	var ret uint32
	for _, v := range ws.objs {
		ret += uint32(v.GetNums())
	}
	return ret
}

func (ws *WaitingScene) getNum() uint32 {
	var ret uint32
	for _, v := range ws.objs {
		ret += uint32(v.GetNums())
	}
	// Ai的数量
	ret += ws.aiNum
	return ret
}

func (ws *WaitingScene) getScoreStr() string {
	ret := "("

	for _, v := range ws.objs {
		ret += v.getScoreStr()
		ret += ";"
	}

	ret += ")"
	return ret
}

func (ws *WaitingScene) addAiNum() {
	ws.aiExtra += ws.aiSpeed - float32(uint32(ws.aiSpeed))
	ws.aiNum += uint32(ws.aiSpeed) + uint32(ws.aiExtra)
	ws.aiExtra = ws.aiExtra - float32(uint32(ws.aiExtra))
	if ws.aiNum >= ws.aiMaxNum {
		ws.aiNum = ws.aiMaxNum
	}
	log.Debugf("Ai Num add speed %f aiNum %d extraNum %f", ws.aiSpeed, ws.aiNum, ws.aiExtra)
}

//AiSpeed 增加ai添加速度
func (ws *WaitingScene) AiSpeed() {
	if ws.aiNum >= ws.aiMaxNum {
		return
	}
	now := time.Now().Unix()
	if (ws.aiSpeedNext != 0 && now < ws.aiSpeedNext) ||
		(ws.aiSpeedNext == 0 && now-ws.start < ws.aiSpeedStart) {
		return
	}
	ws.aiSpeed += ws.aiSpeedAdd
	if ws.aiSpeed >= ws.aiMaxSpeed {
		ws.aiSpeed = ws.aiMaxSpeed
	}
	ws.aiSpeedNext = now + 5
}

func (ws *WaitingScene) aiNumToObjs() uint32 {
	if ws.singleMatch {
		return ws.aiNum
	}
	var ret float64
	if ws.teamType == TwoTeamType {
		ret = float64(ws.aiNum) / 2

	} else if ws.teamType == FourTeamType {
		ret = float64(ws.aiNum) / 4
	}

	return uint32(math.Ceil(ret))
}

// checkMaxNum 检测本局总人数是否超过上限，如果超过将ai数量减少至总人数上限值
func (ws *WaitingScene) checkMaxNum() {
	aiTeamNum := ws.aiNumToObjs()
	total := uint32(len(ws.objs)) + aiTeamNum
	reduceNum := total - ws.maxNum
	if reduceNum <= 0 {
		return
	}

	log.Debug("ws.aiNum:", ws.aiNum, " total:", total, " reduceNum:", reduceNum, " ws.aiNumToObjs():", ws.aiNumToObjs())

	if ws.singleMatch {
		if aiTeamNum >= reduceNum {
			ws.aiNum = aiTeamNum - reduceNum
		}
	}

	if ws.teamType == TwoTeamType {
		if aiTeamNum >= reduceNum {
			ws.aiNum = (aiTeamNum - reduceNum) * 2
		}
	} else if ws.teamType == FourTeamType {
		if aiTeamNum >= reduceNum {
			ws.aiNum = (aiTeamNum - reduceNum) * 4
		}
	}
}

package main

import (
	"common"
	"db"
	"excel"
	"protoMsg"
	"time"
	"zeus/iserver"

	log "github.com/cihub/seelog"
)

/*------------------------------------------节日活动-----------------------------------------------*/

// syncFestivalInfo 同步已经领取到任务的id
func (a *ActivityMgr) syncFestivalInfo() {
	a.user.Debug("syncFestivalInfo")

	state, index := a.checkOpenDate(festivalActivityID)
	if state == ConfigErr {
		return
	}

	info := &db.FestivalActivityInfo{}
	info.FinishNum = make(map[uint32]uint32)
	info.PickState = make(map[uint32]uint32)
	info.FinishTime = make(map[uint32]int64)

	if ok := a.activityUtil[festivalActivityID].IsActivity(); ok && state == 0 {
		if err := a.activityUtil[festivalActivityID].GetInfo(info); err != nil {
			a.user.Error("GetInfo fail:", err)
			return
		}

		for k, v := range excel.GetFestivalMap() {
			if v.Flush != 1 {
				continue
			}

			if info.ActStartTm != a.activeInfo[festivalActivityID].tmStart[index] || time.Unix(info.FinishTime[uint32(k)], 0).Format("2006-01-02") != time.Now().Format("2006-01-02") {
				info.FinishNum[uint32(k)] = 0
				info.PickState[uint32(k)] = 0
			}
		}
	} else {
		info.Id = festivalActivityID
		info.ActStartTm = a.activeInfo[festivalActivityID].tmStart[index]

		if ok := a.activityUtil[festivalActivityID].SetInfo(info); !ok {
			a.user.Error("SetInfo fail!")
			return
		}
	}

	msg := &protoMsg.FestivalList{}
	msg.ActState = state
	for k, _ := range excel.GetFestivalMap() {
		festivalInfo := &protoMsg.FestivalInfo{}
		festivalInfo.Id = uint32(k)
		festivalInfo.Num = info.FinishNum[uint32(k)]
		festivalInfo.State = info.PickState[uint32(k)]

		msg.Info = append(msg.Info, festivalInfo)
	}

	log.Debug("syncFestivalInfo msg:", msg)
	if err := a.user.RPC(iserver.ServerTypeClient, "syncFestivalInfo", msg); err != nil {
		a.user.Error(err)
	}
}

// pickFestivalReward 领取节日活动奖励
func (a *ActivityMgr) pickFestivalReward(id uint32) uint32 {

	state, index := a.checkOpenDate(festivalActivityID)
	if state != 0 {
		return 0
	}

	info := &db.FestivalActivityInfo{}
	info.FinishNum = make(map[uint32]uint32)
	info.PickState = make(map[uint32]uint32)
	info.FinishTime = make(map[uint32]int64)

	if ok := a.activityUtil[festivalActivityID].IsActivity(); ok {
		if err := a.activityUtil[festivalActivityID].GetInfo(info); err != nil {
			a.user.Error("GetInfo fail:", err)
			return 0
		}
	}
	info.Id = festivalActivityID
	info.Time = time.Now().Unix()
	info.ActStartTm = a.activeInfo[festivalActivityID].tmStart[index]

	if time.Unix(info.FinishTime[id], 0).Format("2006-01-02") != time.Now().Format("2006-01-02") {
		return 0
	}

	if info.PickState[id] == 1 {
		return 0
	}

	festivalData, ok := excel.GetFestival(uint64(id))
	if !ok {
		return 0
	}
	if uint64(info.FinishNum[id]) < festivalData.Requirenum {
		return 0
	}

	if ok1 := a.addFestivalRewardBag(id, uint32(index)); ok1 {
		info.PickState[id] = 1
		if ok2 := a.activityUtil[festivalActivityID].SetInfo(info); !ok2 {
			a.user.Error("SetInfo fail!")
			return 0
		}
	}

	a.user.Debug("pickFestivalReward info:", info)
	return 1
}

// addFestivalRewardBag 添加领取的物品到包裹
func (a *ActivityMgr) addFestivalRewardBag(id, index uint32) bool {
	var totalNum uint32 = 0
	for i := 0; i < int(index); i++ {
		totalNum += a.activeInfo[festivalActivityID].pickNum[i]
	}
	log.Info("totalNum:", totalNum, " id:", id, " index:", index)

	data, ok := excel.GetFestival(uint64(totalNum + id))
	if !ok {
		a.user.Warn("GetFestival fail!")
		return false
	}

	awardsMap, err := common.SplitReward(data.Awards)
	if err != nil {
		a.user.Error("splitReward err:", err)
		return false
	}

	for k, v := range awardsMap {
		a.user.storeMgr.GetGoods(k, v, common.RS_Festival, common.MT_NO, 0)
	}

	a.user.tlogFestivalFlow(id, awardsMap)
	return true
}

// updateActivityInfo 更新活动信息
func (a *ActivityMgr) updateActivityInfo(data map[uint64]int32) {
	state, index := a.checkOpenDate(festivalActivityID)
	if state != 0 {
		return
	}

	info := &db.FestivalActivityInfo{}
	info.FinishNum = make(map[uint32]uint32)
	info.PickState = make(map[uint32]uint32)
	info.FinishTime = make(map[uint32]int64)

	if ok := a.activityUtil[festivalActivityID].IsActivity(); ok {
		if err := a.activityUtil[festivalActivityID].GetInfo(info); err != nil {
			a.user.Error("GetInfo fail:", err)
			return
		}
	}

	tmp := false
	for k, v := range excel.GetFestivalMap() {
		if info.ActStartTm != a.activeInfo[festivalActivityID].tmStart[index] || time.Unix(info.FinishTime[uint32(k)], 0).Format("2006-01-02") != time.Now().Format("2006-01-02") {
			info.FinishNum[uint32(k)] = uint32(data[v.Taskpool])
			info.FinishTime[uint32(k)] = time.Now().Unix()
			info.PickState[uint32(k)] = 0
			tmp = true
		} else {
			if data[v.Taskpool] != 0 && uint64(info.FinishNum[uint32(k)]) < v.Requirenum {
				info.FinishNum[uint32(k)] += uint32(data[v.Taskpool])
				tmp = true
			}
		}
	}
	if !tmp {
		return
	}

	info.ActStartTm = a.activeInfo[festivalActivityID].tmStart[index]
	if ok := a.activityUtil[festivalActivityID].SetInfo(info); !ok {
		a.user.Error("SetInfo fail!")
		return
	}

	a.syncFestivalInfo()
}

package main

import (
	"common"
	"db"
	"excel"
	"strconv"
	"strings"
	"time"
	"zeus/iserver"

	log "github.com/cihub/seelog"
)

/*------------------------------------------重返光荣任务-----------------------------------------------*/

// checkVeteran 检测是否是老兵
func (a *ActivityMgr) checkVeteran() {
	if a.activeInfo[backBattleID] == nil || a.activityUtil[backBattleID] == nil {
		a.user.Error("a.activeInfo[backBattleID] == nil or a.activityUtil[backBattleID] == nil!")
		return
	}

	index := a.getCurActivityTime(a.activeInfo[VeteranRecallID])
	callbackData, ok := excel.GetCallback(uint64(index + 1))
	if !ok {
		a.user.Warn("GetCallback fail!")
		return
	}

	t, _ := time.ParseInLocation("2006-01-02", time.Now().Format("2006-01-02"), time.Local)
	info := &db.BackBattleInfo{
		Id:          backBattleID,
		TaskStartTm: t.Unix(),
		RoundNum:    1,
	}
	if ok := a.activityUtil[backBattleID].IsActivity(); ok {
		if err := a.activityUtil[backBattleID].GetInfo(info); err != nil {
			a.user.Warn("GetInfo fail:", err)
			return
		}
	}

	if info.RoundNum == 0 || info.RoundNum-1 >= uint32(len(a.activeInfo[backBattleID].pickNum)) {
		return
	}

	if a.user.GetVeteran() == 1 {
		if time.Now().Unix()-info.TaskStartTm >= int64(a.activeInfo[backBattleID].pickNum[info.RoundNum-1])*86400 {
			a.user.SetVeteran(0)
			a.user.SetVeteranDirty()

			info.RoundNum++
			if info.RoundNum > uint32(len(a.activeInfo[backBattleID].pickNum)+1) {
				info.RoundNum = uint32(len(a.activeInfo[backBattleID].pickNum) + 1)
			}
		}
	} else {
		if a.user.GetLogoutTime() != 0 && time.Now().Unix()-a.user.GetLogoutTime() >= int64(callbackData.Exittime)*86400 && info.RoundNum <= uint32(len(a.activeInfo[backBattleID].pickNum)) {
			a.user.SetVeteran(1)
			a.user.SetVeteranDirty()

			info.Time = 0
			info.GoodId = 0
			info.Sum = 0
			info.TaskStartTm = t.Unix()
		}
	}

	if ok := a.activityUtil[backBattleID].SetInfo(info); !ok {
		a.user.Warn("SetInfo fail!")
		return
	}
}

// syncBackBattleInfo 同步已经领取到任务的id
func (a *ActivityMgr) syncBackBattleInfo() {
	a.user.Debug("syncBackBattleInfo")
	if a.activeInfo[backBattleID] == nil || a.activityUtil[backBattleID] == nil {
		return
	}

	if time.Now().Unix() < a.activeInfo[backBattleID].tmSartShow || time.Now().Unix() > a.activeInfo[backBattleID].tmEndShow {
		a.user.Info("syncBackBattleInfo out of date!")
		return
	}

	if a.user.GetVeteran() == 0 {
		return
	}

	if ok := a.activityUtil[backBattleID].IsActivity(); !ok {
		return
	}

	var state uint32 = 0 //未签到
	info := &db.BackBattleInfo{}
	if err := a.activityUtil[backBattleID].GetInfo(info); err != nil {
		a.user.Error("GetInfo fail:", err)
		return
	}
	if info.RoundNum == 0 || info.RoundNum-1 >= uint32(len(a.activeInfo[backBattleID].pickNum)) {
		return
	}

	if info.Sum >= a.activeInfo[backBattleID].pickNum[info.RoundNum-1] || time.Unix(info.Time, 0).Format("2006-01-02") == time.Now().Format("2006-01-02") {
		state = 1 //已签到
	} else {
		info.GoodId++
	}

	// 返回当前活动可领取id
	log.Debug("SyncBackBattleInfo-info:", info, " state:", state)
	if err := a.user.RPC(iserver.ServerTypeClient, "SyncBackBattleInfo", info.GoodId, state, info.RoundNum); err != nil {
		a.user.Error(err)
	}
}

// pickBackBattleReward 领取重返光荣战场奖励
func (a *ActivityMgr) pickBackBattleReward(id uint32) uint32 {
	if a.activeInfo[backBattleID] == nil || a.activityUtil[backBattleID] == nil {
		return 0
	}

	if time.Now().Unix() < a.activeInfo[backBattleID].tmSartShow || time.Now().Unix() > a.activeInfo[backBattleID].tmEndShow {
		a.user.Info("syncBackBattleInfo out of date!")
		return 0
	}

	if a.user.GetVeteran() == 0 {
		return 0
	}

	info := &db.BackBattleInfo{}
	if ok := a.activityUtil[backBattleID].IsActivity(); ok {
		if err := a.activityUtil[backBattleID].GetInfo(info); err != nil {
			a.user.Error("GetInfo fail:", err)
			return 0
		}

		if info.RoundNum == 0 || info.RoundNum-1 >= uint32(len(a.activeInfo[backBattleID].pickNum)) {
			return 0
		}

		if info.Sum >= a.activeInfo[backBattleID].pickNum[info.RoundNum-1] {
			return 0
		}

		if time.Unix(info.Time, 0).Format("2006-01-02") == time.Now().Format("2006-01-02") {
			return 0
		}

		if id-info.GoodId != 1 {
			return 0
		}
	}

	info.Id = backBattleID
	info.Time = time.Now().Unix()
	info.GoodId = id
	info.Sum++
	if info.Sum > a.activeInfo[backBattleID].pickNum[info.RoundNum-1] {
		info.Sum = a.activeInfo[backBattleID].pickNum[info.RoundNum-1]
	}

	if ok := a.activityUtil[backBattleID].SetInfo(info); !ok {
		a.user.Warn("领取失败")
		return 0
	} else {
		if sucss := a.AddBackBattleRewardBag(id, info.Sum, info.RoundNum); !sucss {
			return 0
		}
	}

	return 1
}

// AddBackBattleRewardBag 添加领取的物品到包裹
func (a *ActivityMgr) AddBackBattleRewardBag(id, sum, index uint32) bool {
	var totalNum uint32 = 0
	for i := 0; i < int(index-1); i++ {
		totalNum += a.activeInfo[backBattleID].pickNum[i]
	}
	log.Info("totalNum + sum:", totalNum+sum, " id:", id, " sum:", sum)

	if id != sum {
		return false
	}

	backReward, ok := excel.GetBackreward(uint64(totalNum + sum))
	if !ok {
		a.user.Warn("GetBackreward fail!")
		return false
	}

	if backReward.ActivityNum != uint64(index) {
		return false
	}

	reward := strings.Split(backReward.Reward, ";")
	for _, v := range reward {
		rewardMsg := strings.Split(v, "|")
		if len(rewardMsg) != 2 {
			return false
		}

		rewardType, err := strconv.Atoi(rewardMsg[0])
		if err != nil {
			return false
		}
		rewardNum, err := strconv.Atoi(rewardMsg[1])
		if err != nil {
			return false
		}

		a.user.storeMgr.GetGoods(uint32(rewardType), uint32(rewardNum), common.RS_BackBattle, common.MT_NO, 0)
	}

	return true
}

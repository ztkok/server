package main

import (
	"common"
	"db"
	"excel"
	"time"
	"zeus/iserver"

	log "github.com/cihub/seelog"
)

/*--------------------------------首胜活动------------------------------------*/

// syncFirstWinActivity 同步首胜活动信息
func (a *ActivityMgr) syncFirstWinActivity() {
	a.user.Debug("syncFirstWinActivity")
	if a.activeInfo[firstWinActivityID] == nil || a.activityUtil[firstWinActivityID] == nil {
		return
	}

	if time.Now().Unix() < a.activeInfo[firstWinActivityID].tmSartShow || time.Now().Unix() > a.activeInfo[firstWinActivityID].tmEndShow {
		a.user.Info("syncFirstWinActivity out of date!")
		return
	}

	index := a.getCurActivityTime(a.activeInfo[firstWinActivityID])
	if index >= len(a.activeInfo[firstWinActivityID].tmStart) || index >= len(a.activeInfo[firstWinActivityID].tmEnd) || index >= len(a.activeInfo[firstWinActivityID].pickNum) {
		a.user.Warn("tmStart):", len(a.activeInfo[firstWinActivityID].tmStart), " tmEnd):", len(a.activeInfo[firstWinActivityID].tmEnd), " pickNum:", len(a.activeInfo[firstWinActivityID].pickNum), " index:", index+1)
		return
	}

	var recState uint32 = 0 //活动状态
	if time.Now().Unix() < a.activeInfo[firstWinActivityID].tmStart[index] {
		recState = 1 //未开始
	}
	if time.Now().Unix() > a.activeInfo[firstWinActivityID].tmEnd[index] {
		recState = 2 //已过期
	}

	var CheckinState uint32 = 0 //当天未获取首胜
	info := &db.SigActivityInfo{}
	info.GoodId = 1

	if ok := a.activityUtil[firstWinActivityID].IsActivity(); ok {
		err := a.activityUtil[firstWinActivityID].GetActivityInfo(info)
		if err != nil {
			a.user.Error("GetActivityInfo fail:", err)
			return
		}
		if info.ActStartTm != a.activeInfo[firstWinActivityID].tmStart[index] {
			info.GoodId = 1
			a.activityUtil[firstWinActivityID].ClearOwnActivity()
		} else {
			if time.Unix(info.Time, 0).Format("2006-01-02") == time.Now().Format("2006-01-02") {
				CheckinState = 1 //已签到
			} else {
				info.GoodId += 1
			}

			if info.Sum >= a.activeInfo[firstWinActivityID].pickNum[index] {
				CheckinState = 1 //已签到(签到已满)
				info.GoodId = a.activeInfo[firstWinActivityID].pickNum[index]
			}
		}
	}

	log.Debug("timeStart:", time.Unix(a.activeInfo[firstWinActivityID].tmStart[index], 0).Format("2006-01-02"), "timeEnd:",
		time.Unix(a.activeInfo[firstWinActivityID].tmEnd[index], 0).Format("2006-01-02"), " timeNow:", time.Now().Format("2006-01-02"))
	log.Info("InitChickenCheckIn-GoodId:", info.GoodId, " CheckinState:", CheckinState, " recState:", recState, " index:", index)
	// 返回当前活动可领取id
	if err := a.user.RPC(iserver.ServerTypeClient, "InitChickenCheckIn", info.GoodId, CheckinState, recState, uint32(index+1)); err != nil {
		a.user.Error(err)
	}
}

// PickFirstWinActivity 领取首胜活动物品
func (a *ActivityMgr) PickFirstWinActivity() {
	a.user.Debug("PickFirstWinActivity")
	if a.activeInfo[firstWinActivityID] == nil || a.activityUtil[firstWinActivityID] == nil {
		return
	}

	if time.Now().Unix() < a.activeInfo[firstWinActivityID].tmSartShow || time.Now().Unix() > a.activeInfo[firstWinActivityID].tmEndShow {
		a.user.Info("PickFirstWinActivity out of date!")
		return
	}

	index := a.getCurActivityTime(a.activeInfo[firstWinActivityID])
	if index >= len(a.activeInfo[firstWinActivityID].tmStart) || index >= len(a.activeInfo[firstWinActivityID].tmEnd) || index >= len(a.activeInfo[firstWinActivityID].pickNum) {
		a.user.Warn("len(activeinfo.tmStart):", len(a.activeInfo[firstWinActivityID].tmStart), " len(activeinfo.tmEnd):", len(a.activeInfo[firstWinActivityID].tmEnd), " len(activeinfo.pickNum):", len(a.activeInfo[firstWinActivityID].pickNum), " index:", index+1)
		return
	}

	if time.Now().Unix() < a.activeInfo[firstWinActivityID].tmStart[index] || time.Now().Unix() > a.activeInfo[firstWinActivityID].tmEnd[index] {
		a.user.Info("activity no start or expire")
		return
	}

	info := &db.SigActivityInfo{}
	info.GoodId = 1
	info.Sum = 1

	if ok := a.activityUtil[firstWinActivityID].IsActivity(); ok {
		if err := a.activityUtil[firstWinActivityID].GetActivityInfo(info); err != nil {
			a.user.Error("GetActivityInfo fail:", err)
			return
		}

		if info.Sum >= a.activeInfo[firstWinActivityID].pickNum[index] {
			a.user.Info("Sum:", info.Sum, " pickNum:", a.activeInfo[firstWinActivityID].pickNum[index])
			return
		}

		if time.Unix(info.Time, 0).Format("2006-01-02") != time.Now().Format("2006-01-02") {
			info.GoodId += 1
			info.Sum += 1
		} else {
			a.user.Info("已经吃过鸡")
			return
		}
	}

	info.Id = firstWinActivityID
	info.Time = time.Now().Unix()
	info.ActStartTm = a.activeInfo[firstWinActivityID].tmStart[index]

	var recState uint32 = 1 //有资格领取（领取成功）
	ok := a.activityUtil[firstWinActivityID].SetActivityInfo(info)
	if !ok {
		a.user.Error("领取失败")
		recState = 0 //领取失败
	} else {
		if sucss := a.AddFirstWinRewardToBag(info.GoodId, info.Sum, index); !sucss {
			recState = 0 //领取失败
		}
	}

	log.Debug("time1:", time.Unix(info.Time, 0).Format("2006-01-02"), " time2:", time.Now().Format("2006-01-02"))
	log.Info("ChickenCheckInReward-GoodId:", info.GoodId, " recState:", recState, " index:", index, " info.Sum:", info.Sum)
	if err := a.user.RPC(iserver.ServerTypeClient, "ChickenCheckInReward", info.GoodId, recState); err != nil {
		a.user.Error(err)
	}
}

// AddFirstWinRewardToBag 添加领取的物品到包裹
func (a *ActivityMgr) AddFirstWinRewardToBag(id, sum uint32, index int) bool {
	var totalNum uint32 = 0
	for i := 0; i < index; i++ {
		totalNum += a.activeInfo[firstWinActivityID].pickNum[i]
	}

	log.Info("totalNum + sum:", totalNum+sum, " id:", id, " sum:", sum)

	if id != sum {
		return false
	}

	activity, ok := excel.GetChickenCheckin(uint64(totalNum + sum))
	if ok {
		if activity.ActivityNum == uint64(index+1) {
			a.user.storeMgr.GetGoods(uint32(activity.BonusID), uint32(activity.BonusNum), common.RS_FirstWinAct, common.MT_NO, 0) //common.RS_FirstWinAct代表当天首次吃鸡活动
			return true
		}
	}

	return false
}

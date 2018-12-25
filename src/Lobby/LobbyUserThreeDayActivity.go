package main

import (
	"common"
	"db"
	"excel"
	"time"
	"zeus/iserver"

	log "github.com/cihub/seelog"
)

/*------------------------------------------3天签到活动-----------------------------------------------*/

// syncThreeDayActivityID 同步已经领取到的id
func (a *ActivityMgr) syncThreeDayActivityID() {
	a.user.Debug("syncThreeDayActivityID")
	if a.activeInfo[threeDayActivityID] == nil || a.activityUtil[threeDayActivityID] == nil {
		return
	}

	if time.Now().Unix() < a.activeInfo[threeDayActivityID].tmSartShow || time.Now().Unix() > a.activeInfo[threeDayActivityID].tmEndShow {
		a.user.Info("syncThreeDayActivity out of date!")
		return
	}

	index := a.getCurActivityTime(a.activeInfo[threeDayActivityID])
	if index >= len(a.activeInfo[threeDayActivityID].tmStart) || index >= len(a.activeInfo[threeDayActivityID].tmEnd) || index >= len(a.activeInfo[threeDayActivityID].pickNum) {
		a.user.Warn("len(tmStart):", len(a.activeInfo[threeDayActivityID].tmStart), " len(tmEnd):", len(a.activeInfo[threeDayActivityID].tmEnd), " len(pickNum):", len(a.activeInfo[threeDayActivityID].pickNum), " index:", index+1)
		return
	}

	var recState uint32 = 0 //活动状态
	if time.Now().Unix() < a.activeInfo[threeDayActivityID].tmStart[index] {
		recState = 1 //未开始
	}
	if time.Now().Unix() > a.activeInfo[threeDayActivityID].tmEnd[index] {
		recState = 2 //已过期
	}

	var CheckinState uint32 = 0 //当天未获取首胜
	info := &db.SigActivityInfo{}
	info.GoodId = 1

	if ok := a.activityUtil[threeDayActivityID].IsActivity(); ok {
		err := a.activityUtil[threeDayActivityID].GetActivityInfo(info)
		if err != nil {
			a.user.Error("GetActivityInfo fail:", err)
			return
		}
		if info.ActStartTm != a.activeInfo[threeDayActivityID].tmStart[index] {
			info.GoodId = 1
			a.activityUtil[threeDayActivityID].ClearOwnActivity()
		} else {
			if time.Unix(info.Time, 0).Format("2006-01-02") == time.Now().Format("2006-01-02") {
				CheckinState = 1 //已签到
			} else {
				info.GoodId += 1
			}

			if info.Sum >= a.activeInfo[threeDayActivityID].pickNum[index] {
				CheckinState = 1 //已签到(签到已满)
				info.GoodId = a.activeInfo[threeDayActivityID].pickNum[index]
			}
		}
	}

	log.Debug("timeStart:", time.Unix(a.activeInfo[threeDayActivityID].tmStart[index], 0).Format("2006-01-02"), "timeEnd:",
		time.Unix(a.activeInfo[threeDayActivityID].tmEnd[index], 0).Format("2006-01-02"), " timeNow:", time.Now().Format("2006-01-02"))
	log.Info("InitTDCheckIn-GoodId:", info.GoodId, " CheckinState:", CheckinState, " recState", recState, " index:", index)
	// 返回当前活动可领取id
	if err := a.user.RPC(iserver.ServerTypeClient, "InitTDCheckIn", info.GoodId, CheckinState, recState, uint32(index+1)); err != nil {
		a.user.Error(err)
	}
}

// SetThreeDayActivity 领取3天签到活动物品
func (a *ActivityMgr) SetThreeDayActivity(id uint32) {
	a.user.Debug("SetThreeDayActivity")
	if a.activeInfo[threeDayActivityID] == nil || a.activityUtil[threeDayActivityID] == nil {
		return
	}

	if time.Now().Unix() < a.activeInfo[threeDayActivityID].tmSartShow || time.Now().Unix() > a.activeInfo[threeDayActivityID].tmEndShow {
		a.user.Info("ThreeDayActivity out of date!")
		return
	}

	index := a.getCurActivityTime(a.activeInfo[threeDayActivityID])
	if index >= len(a.activeInfo[threeDayActivityID].tmStart) || index >= len(a.activeInfo[threeDayActivityID].tmEnd) || index >= len(a.activeInfo[threeDayActivityID].pickNum) {
		a.user.Warn("len(tmStart):", len(a.activeInfo[threeDayActivityID].tmStart), " len(tmEnd):", len(a.activeInfo[threeDayActivityID].tmEnd), " len(pickNum):", len(a.activeInfo[threeDayActivityID].pickNum), " index:", index+1)
		return
	}

	var recState, result uint32 = 0, 0 //有资格领取（领取成功）
	if time.Now().Unix() < a.activeInfo[threeDayActivityID].tmStart[index] {
		result = NoStart
	} else if time.Now().Unix() > a.activeInfo[threeDayActivityID].tmEnd[index] {
		result = OutDate
	}

	info := &db.SigActivityInfo{}

	if ok := a.activityUtil[threeDayActivityID].IsActivity(); ok && result == 0 {
		if err := a.activityUtil[threeDayActivityID].GetActivityInfo(info); err != nil {
			a.user.Error("GetActivityInfo fail:", err)
			return
		}
		if time.Unix(info.Time, 0).Format("2006-01-02") == time.Now().Format("2006-01-02") {
			result = HasChecked //今天已领取
		}
		if info.Sum >= a.activeInfo[threeDayActivityID].pickNum[index] {
			result = Failure //活动已领取完成（领取失败）
		}
		if id == info.GoodId {
			result = HasChecked //该物品已领取（领取失败）
		}
	}

	if result == 0 {
		info.Id = firstWinActivityID
		info.Time = time.Now().Unix()
		info.GoodId++
		info.Sum++
		info.ActStartTm = a.activeInfo[threeDayActivityID].tmStart[index]

		ok := a.activityUtil[threeDayActivityID].SetActivityInfo(info)
		if !ok {
			a.user.Error("领取失败")
			result = Failure //领取失败
		} else {
			if sucss := a.AddThreeDayToBag(info.GoodId, info.Sum, index); !sucss {
				result = HasChecked //领取失败
			}
		}

		if result == 0 {
			recState = 1
		}
	}

	log.Info("TDCheckInReward-GoodId:", info.GoodId, " id:", id, " recState:", recState, " result", result, " index:", index, " info.Sum:", info.Sum)
	if err := a.user.RPC(iserver.ServerTypeClient, "TDCheckInReward", info.GoodId, recState, result); err != nil {
		a.user.Error(err)
	}
}

// AddOwnBag 添加领取的物品到包裹
func (a *ActivityMgr) AddThreeDayToBag(id, sum uint32, index int) bool {
	var totalNum uint32 = 0
	for i := 0; i < index; i++ {
		totalNum += a.activeInfo[threeDayActivityID].pickNum[i]
	}
	log.Info("totalNum + sum:", totalNum+sum, " id:", id, " sum:", sum)

	if id != sum {
		return false
	}

	activity, ok := excel.GetThreedayCheckin(uint64(totalNum + sum))
	if ok {
		if activity.ActivityNum == uint64(index+1) {
			a.user.storeMgr.GetGoods(uint32(activity.BonusID), uint32(activity.BonusNum), common.RS_ThreeDayAct, common.MT_NO, 0)
			return true
		}
	}

	return false
}

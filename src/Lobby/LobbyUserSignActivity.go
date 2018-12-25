package main

import (
	"common"
	"db"
	"excel"
	"time"
	"zeus/iserver"

	log "github.com/cihub/seelog"
)

/*------------------------------------------签到活动-----------------------------------------------*/

// syncSignActivityID 同步已经领取到的id
func (a *ActivityMgr) syncSignActivityID() {
	a.user.Debug("syncActivityID")
	if a.activeInfo[signActivityID] == nil || a.activityUtil[signActivityID] == nil {
		return
	}

	if time.Now().Unix() < a.activeInfo[signActivityID].tmSartShow || time.Now().Unix() > a.activeInfo[signActivityID].tmEndShow {
		a.user.Info("SignActivity out of date!")
		return
	}

	index := a.getCurActivityTime(a.activeInfo[signActivityID])
	if index >= len(a.activeInfo[signActivityID].tmStart) || index >= len(a.activeInfo[signActivityID].tmEnd) || index >= len(a.activeInfo[signActivityID].pickNum) {
		a.user.Warn("tmStart):", len(a.activeInfo[signActivityID].tmStart), " tmEnd):", len(a.activeInfo[signActivityID].tmEnd), " pickNum):", len(a.activeInfo[signActivityID].pickNum), " index:", index+1)
		return
	}

	var CheckinState uint32 = 0 //未签到
	info := &db.SigActivityInfo{}

	if time.Now().Unix() < a.activeInfo[signActivityID].tmStart[index] {
		CheckinState = 3 //未开始
	}

	if ok := a.activityUtil[signActivityID].IsActivity(); ok && CheckinState == 0 {
		err := a.activityUtil[signActivityID].GetActivityInfo(info)
		if err != nil {
			a.user.Error("GetActivityInfo fail:", err)
			return
		}
		if info.ActStartTm != a.activeInfo[signActivityID].tmStart[index] {
			info.GoodId = 0
			a.activityUtil[signActivityID].ClearOwnActivity()
		} else {
			if time.Unix(info.Time, 0).Format("2006-01-02") == time.Now().Format("2006-01-02") {
				CheckinState = 1 //已签到
				info.GoodId -= 1
			}
			if info.Sum >= a.activeInfo[signActivityID].pickNum[index] {
				CheckinState = 1 //已签到(签到已满)
			}
		}
	}

	if time.Now().Unix() > a.activeInfo[signActivityID].tmEnd[index] {
		CheckinState = 2 //已过期
	}

	log.Debug("timeStart:", time.Unix(a.activeInfo[signActivityID].tmStart[index], 0).Format("2006-01-02"), "timeEnd:",
		time.Unix(a.activeInfo[signActivityID].tmEnd[index], 0).Format("2006-01-02"), " timeNow:", time.Now().Format("2006-01-02"))
	log.Info("respCrtCheckID-GoodId:", info.GoodId, " CheckinState:", CheckinState, " index:", index)
	// 返回当前活动可领取id
	if err := a.user.RPC(iserver.ServerTypeClient, "respCrtCheckID", info.GoodId, CheckinState, uint32(index+1)); err != nil {
		a.user.Error(err)
	}
}

// SetActivity 领取签到活动物品
func (a *ActivityMgr) SetSignActivity(id uint32) {
	a.user.Debug("SetActivity:", id)
	if a.activeInfo[signActivityID] == nil || a.activityUtil[signActivityID] == nil {
		return
	}

	index := a.getCurActivityTime(a.activeInfo[signActivityID])
	if index >= len(a.activeInfo[signActivityID].tmStart) || index >= len(a.activeInfo[signActivityID].tmEnd) || index >= len(a.activeInfo[signActivityID].pickNum) {
		a.user.Warn("len(tmStart):", len(a.activeInfo[signActivityID].tmStart), " len(tmEnd):", len(a.activeInfo[signActivityID].tmEnd), " len(pickNum):", len(a.activeInfo[signActivityID].pickNum), " index:", index+1)
		return
	}

	var recState uint32 = 0 //有资格领取（领取成功）
	info := &db.SigActivityInfo{}

	if time.Now().Unix() < a.activeInfo[signActivityID].tmStart[index] {
		recState = 4 //未开始
	}
	if time.Now().Unix() > a.activeInfo[signActivityID].tmEnd[index] {
		recState = 2 //已过期
	}

	if ok := a.activityUtil[signActivityID].IsActivity(); ok && recState == 0 {
		if err := a.activityUtil[signActivityID].GetActivityInfo(info); err != nil {
			a.user.Error("GetActivityInfo fail:", err)
			return
		}

		if time.Unix(info.Time, 0).Format("2006-01-02") == time.Now().Format("2006-01-02") {
			recState = 1 //今天已领取
		}
		if info.Sum >= a.activeInfo[signActivityID].pickNum[index] {
			recState = 3 //活动已领取7天（领取失败）
		}
		if id == info.GoodId {
			recState = 3 //该物品已领取（领取失败）
		}
	}

	if recState == 0 {
		info.Id = signActivityID
		info.Time = time.Now().Unix()
		info.GoodId = id
		info.Sum += 1
		info.ActStartTm = a.activeInfo[signActivityID].tmStart[index]

		ok := a.activityUtil[signActivityID].SetActivityInfo(info)
		if !ok {
			a.user.Error("领取失败")
			recState = 3 //领取失败
		} else {
			if sucss := a.AddOwnBag(id, info.Sum, index); !sucss {
				recState = 3 //领取失败
			}
		}
	}

	log.Debug("time1:", time.Unix(info.Time, 0).Format("2006-01-02"), " time2:", time.Now().Format("2006-01-02"))
	log.Info("receiveCheckinID-id:", id, " recState:", recState, " index:", index, " info.Sum:", info.Sum)
	if err := a.user.RPC(iserver.ServerTypeClient, "receiveCheckinID", id, recState); err != nil {
		a.user.Error(err)
	}
}

// AddOwnBag 添加领取的物品到包裹
func (a *ActivityMgr) AddOwnBag(id, sum uint32, index int) bool {
	var totalNum uint32 = 0
	for i := 0; i < index; i++ {
		totalNum += a.activeInfo[signActivityID].pickNum[i]
	}
	log.Info("totalNum + sum:", totalNum+sum, " id:", id, " sum:", sum)

	if id != sum {
		return false
	}

	activity, ok := excel.GetActivity(uint64(totalNum + sum))
	if ok {
		if activity.ActivityNum == uint64(index+1) {
			a.user.storeMgr.GetGoods(uint32(activity.BonusID), uint32(activity.BonusNum), common.RS_Activity, common.MT_NO, 0) //2代表签到活动
			return true
		}
	}

	return false
}

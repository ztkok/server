package main

import (
	"common"
	"db"
	"excel"
	"time"
	"zeus/iserver"

	log "github.com/cihub/seelog"
)

/*--------------------------------更新活动------------------------------------*/

// syncUpdataActivity 同步更新活动信息
func (a *ActivityMgr) syncUpdateActivity() {
	a.user.Debug("syncUpdateActivity")
	if a.activeInfo[updateAcitityID] == nil || a.activityUtil[updateAcitityID] == nil {
		return
	}

	if time.Now().Unix() < a.activeInfo[updateAcitityID].tmSartShow || time.Now().Unix() > a.activeInfo[updateAcitityID].tmEndShow {
		a.user.Info("UpdateActivity out of date!")
		return
	}

	index := a.getCurActivityTime(a.activeInfo[updateAcitityID])
	if index >= len(a.activeInfo[updateAcitityID].tmStart) || index >= len(a.activeInfo[updateAcitityID].tmEnd) {
		a.user.Info("len(tmStart) < index+1 || len(tmEnd) < index+1")
		return
	}

	var campaignState uint32 = 1 //活动进行中
	var rewardState uint32 = 0   //未领取更新活动物品

	if time.Now().Unix() < a.activeInfo[updateAcitityID].tmStart[index] {
		campaignState = 3 //未开始
	}
	if time.Now().Unix() > a.activeInfo[updateAcitityID].tmEnd[index] {
		campaignState = 2 //已过期
	}

	if a.activityUtil[updateAcitityID].IsUpdateActivity() {
		info := &db.UpdateActivityInfo{}
		err := a.activityUtil[updateAcitityID].GetUpdateActivityInfo(info)
		if err != nil {
			a.user.Error("GetUpdateActivityInfo fail:", err)
			return
		}
		if info.ActStartTm != a.activeInfo[updateAcitityID].tmStart[index] {
			a.activityUtil[updateAcitityID].ClearUpdateActivity()
		} else {
			rewardState = 1 //已经领取更新活动物品
		}
	}

	log.Info("rewardState:", rewardState, " campaignState:", campaignState)
	if err := a.user.RPC(iserver.ServerTypeClient, "InitVersionRewardState", rewardState, campaignState, uint32(index+1)); err != nil {
		a.user.Error(err)
	}
}

// SetUpdateActivity 领取更新活动物品
func (a *ActivityMgr) SetUpdateActivity() uint32 {
	a.user.Debug("SetUpdateActivity")
	if a.activeInfo[updateAcitityID] == nil || a.activityUtil[updateAcitityID] == nil {
		return 0
	}

	index := a.getCurActivityTime(a.activeInfo[updateAcitityID])
	if index >= len(a.activeInfo[updateAcitityID].tmStart) || index >= len(a.activeInfo[updateAcitityID].tmEnd) {
		a.user.Info("len(tmStart) < index+1 || len(tmEnd) < index+1")
		return 0
	}

	//未开始
	if time.Now().Unix() < a.activeInfo[updateAcitityID].tmStart[index] {
		return 4
	}

	//已过期
	if time.Now().Unix() > a.activeInfo[updateAcitityID].tmEnd[index] {
		return 5
	}

	//已经领取更新活动物品
	if a.activityUtil[updateAcitityID].IsUpdateActivity() {
		return 3
	}

	updateAct, ok := excel.GetReversionWard(1)
	if !ok {
		a.user.Error("GetReversionWard fail!")
		return 0
	}

	if ok := a.checkVersion(updateAct.IosrewardVersion, updateAct.AndroidrewardVersion); ok {
		info := &db.UpdateActivityInfo{}
		info.Id = updateAcitityID
		info.Time = time.Now().Unix()
		info.ActStartTm = a.activeInfo[updateAcitityID].tmStart[index]
		if ok := a.activityUtil[updateAcitityID].SetUpdateActivityInfo(info); ok {
			a.user.storeMgr.GetGoods(uint32(updateAct.BonusID), uint32(updateAct.BonusNum), common.RS_UpdateAct, common.MT_NO, 0) //更新活动
			log.Info("领取成功")
			return 1 //可领取
		} else {
			return 2
		}
	}

	return 2 //更新才可领取
}

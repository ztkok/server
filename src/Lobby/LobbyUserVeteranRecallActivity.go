package main

import (
	"common"
	"db"
	"excel"
	"protoMsg"
	"strconv"
	"strings"
	"time"
	"zeus/dbservice"
	"zeus/iserver"

	log "github.com/cihub/seelog"
	"github.com/garyburd/redigo/redis"
)

/*------------------------------------------老兵召回活动-----------------------------------------------*/

// syncVeteranRecall 同步老兵召回 好友列表 和 奖励领取信息
func (a *ActivityMgr) syncVeteranRecall() {
	a.syncVeteranRecallInfo()
	a.syncVeteranRecallReward()
}

// syncVeteranRecallInfo 同步老兵召回的平台好友信息列表
func (a *ActivityMgr) syncVeteranRecallInfo() {
	a.user.Debug("syncVeteranRecallInfo")
	if a.activeInfo[VeteranRecallID] == nil || a.activityUtil[VeteranRecallID] == nil {
		return
	}

	if time.Now().Unix() < a.activeInfo[VeteranRecallID].tmSartShow || time.Now().Unix() > a.activeInfo[VeteranRecallID].tmEndShow {
		a.user.Info("VeteranRecallID out of date!")
		return
	}

	index := a.getCurActivityTime(a.activeInfo[VeteranRecallID])
	if index >= len(a.activeInfo[VeteranRecallID].tmStart) || index >= len(a.activeInfo[VeteranRecallID].tmEnd) {
		a.user.Warn("tmStart):", len(a.activeInfo[VeteranRecallID].tmStart), " tmEnd):", len(a.activeInfo[VeteranRecallID].tmEnd), " index:", index+1)
		return
	}

	var state uint32 = 0 //活动开始中
	if time.Now().Unix() < a.activeInfo[VeteranRecallID].tmStart[index] {
		state = NoStart //未开始
	}
	if time.Now().Unix() > a.activeInfo[VeteranRecallID].tmEnd[index] {
		state = OutDate //已过期
	}

	info := &db.VeteranRecallInfo{}
	if ok := a.activityUtil[VeteranRecallID].IsActivity(); ok && state == 0 {
		if err := a.activityUtil[VeteranRecallID].GetInfo(info); err != nil {
			a.user.Error("GetInfo fail:", err)
			return
		}
		if info.ActStartTm != a.activeInfo[VeteranRecallID].tmStart[index] {
			a.activityUtil[VeteranRecallID].ClearOwnActivity()
		} else if time.Unix(info.Time, 0).Format("2006-01-02") != time.Now().Format("2006-01-02") {
			info.RecallDayNum = 0
			if ok := a.activityUtil[VeteranRecallID].SetInfo(info); !ok {
				log.Warn("SetInfo fail!")
				return
			}
		}
	}

	msg := &protoMsg.VeteranRecallList{}
	if state == Finish {
		msg = a.checkVeteranInfo(uint32(index))
	}
	msg.State = state
	msg.Index = uint32(index) + 1

	log.Debug("syncVeteranRecallInfo msg:", msg)
	if err := a.user.RPC(iserver.ServerTypeClient, "SyncVeteranRecallInfo", msg); err != nil {
		a.user.Error(err)
	}
}

// checkVeteranInfo 检测平台好友是否符合要求
func (a *ActivityMgr) checkVeteranInfo(index uint32) *protoMsg.VeteranRecallList {
	callbackData, ok := excel.GetCallback(uint64(index + 1))
	if !ok {
		a.user.Error("GetCallback fail!")
		return nil
	}

	info := &db.VeteranRecallInfo{}
	if ok := a.activityUtil[VeteranRecallID].IsActivity(); ok {
		err := a.activityUtil[VeteranRecallID].GetInfo(info)
		if err != nil {
			a.user.Error("GetInfo fail:", err)
			return nil
		}
	}

	// 平台好友信息
	veteranList := &protoMsg.VeteranRecallList{}
	platInfo := db.GetFriendUtil(a.user.GetDBID()).GetPlatFrientInfo()
	for _, platid := range platInfo {
		args := []interface{}{
			"Name",
			"LogoutTime",
			"Picture",
		}
		values, valueErr := dbservice.EntityUtil("Player", platid).GetValues(args)
		if valueErr != nil || len(values) != len(args) {
			log.Error("获取url、name、logoutTime错误")
			continue
		}
		tmpUserName, userNameErr := redis.String(values[0], nil)
		if userNameErr != nil {
			continue
		}
		logoutTime, logoutErr := redis.Int64(values[1], nil)
		if logoutErr != nil {
			continue
		}
		tmpUrl, urlErr := redis.String(values[2], nil)
		if urlErr != nil {
			continue
		}

		if logoutTime == 0 || time.Now().Unix()-logoutTime < int64(callbackData.Exittime)*86400 {
			continue
		}

		if time.Now().Unix()-info.RecallInfo[platid] < int64(callbackData.Repeattime)*3600 {
			continue
		}

		veteranInfo := &protoMsg.VeteranInfo{}
		veteranInfo.Uid = platid
		veteranInfo.UserName = tmpUserName
		veteranInfo.Url = tmpUrl
		veteranInfo.NameColor = common.GetPlayerNameColor(platid)

		veteranList.VeteranList = append(veteranList.VeteranList, veteranInfo)
	}

	a.user.Debug("checkVeteranInfo platInfo:", platInfo)
	return veteranList
}

// recallFriendRes 召回好友请求(非成功召回，只是发送请求)
func (a *ActivityMgr) recallFriendRes(uid uint64) bool {
	if a.activeInfo[VeteranRecallID] == nil || a.activityUtil[VeteranRecallID] == nil {
		return false
	}

	index := a.getCurActivityTime(a.activeInfo[VeteranRecallID])
	if index >= len(a.activeInfo[VeteranRecallID].tmStart) || index >= len(a.activeInfo[VeteranRecallID].tmEnd) {
		a.user.Warn("len(tmStart):", len(a.activeInfo[VeteranRecallID].tmStart), " len(tmEnd):", len(a.activeInfo[VeteranRecallID].tmEnd), " index:", index+1)
		return false
	}

	if time.Now().Unix() < a.activeInfo[VeteranRecallID].tmStart[index] || time.Now().Unix() > a.activeInfo[VeteranRecallID].tmEnd[index] {
		return false
	}

	callbackData, ok := excel.GetCallback(uint64(index + 1))
	if !ok {
		a.user.Warn("GetCallback fail!")
		return false
	}

	logoutTime, _ := redis.Int64(dbservice.EntityUtil("Player", uid).GetValue("LogoutTime"))
	if logoutTime == 0 || time.Now().Unix()-logoutTime < int64(callbackData.Exittime)*86400 {
		return false
	}

	info := &db.VeteranRecallInfo{}
	info.RecallInfo = make(map[uint64]int64)
	info.RecallReward = make(map[uint32]uint32)

	if ok := a.activityUtil[VeteranRecallID].IsActivity(); ok {
		if err := a.activityUtil[VeteranRecallID].GetInfo(info); err != nil {
			a.user.Error("GetInfo fail:", err)
			return false
		}

		if time.Now().Unix()-info.RecallInfo[uid] < int64(callbackData.Repeattime)*3600 {
			return false
		}

		if time.Unix(info.Time, 0).Format("2006-01-02") == time.Now().Format("2006-01-02") {
			info.RecallDayNum++
		}

		info.Time = time.Now().Unix()
		info.RecallInfo[uid] = time.Now().Unix()
	} else {
		info.Id = VeteranRecallID
		info.ActStartTm = a.activeInfo[VeteranRecallID].tmStart[index]
		info.Time = time.Now().Unix()
		info.RecallDayNum = 1
		info.RecallInfo[uid] = time.Now().Unix()
	}

	if ok := a.activityUtil[VeteranRecallID].SetInfo(info); !ok {
		a.user.Warn("SetInfo fail!")
		return false
	}

	a.recallFriendReward(uint64(index), uint64(info.RecallDayNum))
	return true
}

// recallFriendReward 召回奖励(非成功召回)
func (a *ActivityMgr) recallFriendReward(index, num uint64) {
	callbackData, ok := excel.GetCallback(index + 1)
	if !ok {
		a.user.Warn("GetCallback fail!")
		return
	}

	if num > callbackData.Clicktimes {
		return
	}

	// 添加邮件
	objs := make(map[uint32]uint32)
	objs[uint32(callbackData.ClickrewardType)] = uint32(callbackData.ClickrewardNum)
	sendObjMail(a.user.GetDBID(), "", 0, callbackData.ClickrewardMailTitle, callbackData.ClickrewardMailContent, "", "", objs)
	if err := a.user.RPC(iserver.ServerTypeClient, "AddNewMail"); err != nil {
		a.user.Error(err)
	}
}

// recallSuccessNotify 成功召回好友通知(成功召回)
func (a *ActivityMgr) recallSuccessNotify(uid uint64) {
	log.Debug("recallSuccessNotify:", uid)
	if a.activeInfo[VeteranRecallID] == nil || a.activityUtil[VeteranRecallID] == nil {
		return
	}

	index := a.getCurActivityTime(a.activeInfo[VeteranRecallID])
	if index >= len(a.activeInfo[VeteranRecallID].tmStart) || index >= len(a.activeInfo[VeteranRecallID].tmEnd) {
		a.user.Warn("len(tmStart):", len(a.activeInfo[VeteranRecallID].tmStart), " len(tmEnd):", len(a.activeInfo[VeteranRecallID].tmEnd), " index:", index+1)
		return
	}

	if time.Now().Unix() < a.activeInfo[VeteranRecallID].tmStart[index] || time.Now().Unix() > a.activeInfo[VeteranRecallID].tmEnd[index] {
		return
	}

	util := db.PlayerActivityUtil(uid, VeteranRecallID)
	if ok := util.IsActivity(); !ok {
		a.user.Error("IsActivity is nothing!")
		return
	}

	info := &db.VeteranRecallInfo{}
	if err := util.GetInfo(info); err != nil {
		a.user.Error("GetInfo fail:", err)
		return
	}

	info.RecallSusNum++
	if ok := util.SetInfo(info); !ok {
		a.user.Error("SetInfo fail!")
		return
	}

	a.recallSuccessReward(uid)

	// 同步成功召回的老兵奖励数量
	user, ok := GetSrvInst().GetEntityByDBID("Player", uid).(*LobbyUser)
	if ok {
		if err := user.RPC(iserver.ServerTypeClient, "AddNewMail"); err != nil {
			user.Error(err)
		}
		user.activityMgr.syncVeteranRecallReward()
	}
}

// recallSuccessReward 成功召回奖励(成功召回)
func (a *ActivityMgr) recallSuccessReward(uid uint64) {
	index := a.getCurActivityTime(a.activeInfo[VeteranRecallID])
	callbackData, ok := excel.GetCallback(uint64(index) + 1)
	if !ok {
		a.user.Warn("GetCallback fail!")
		return
	}

	// 添加邮件
	objs := make(map[uint32]uint32)
	objs[uint32(callbackData.CallrewardType)] = uint32(callbackData.CallrewardNum)
	sendObjMail(uid, "", 0, callbackData.CallrewardMailTitle, callbackData.CallrewardMailContent, "", "", objs)
}

// syncVeteranRecallReward 同步成功召回的奖励信息
func (a *ActivityMgr) syncVeteranRecallReward() {
	a.user.Debug("syncVeteranRecallReward")
	if a.activeInfo[VeteranRecallID] == nil || a.activityUtil[VeteranRecallID] == nil {
		return
	}

	index := a.getCurActivityTime(a.activeInfo[VeteranRecallID])
	if index >= len(a.activeInfo[VeteranRecallID].tmStart) || index >= len(a.activeInfo[VeteranRecallID].tmEnd) {
		a.user.Warn("len(tmStart):", len(a.activeInfo[VeteranRecallID].tmStart), " len(tmEnd):", len(a.activeInfo[VeteranRecallID].tmEnd), " index:", index+1)
		return
	}

	var state uint32 = 0 //活动开始中
	if time.Now().Unix() < a.activeInfo[VeteranRecallID].tmStart[index] {
		state = NoStart //未开始
	}
	if time.Now().Unix() > a.activeInfo[VeteranRecallID].tmEnd[index] {
		state = OutDate //已过期
	}

	msg := &protoMsg.VeteranRecallReward{}
	if state == 0 {
		info := &db.VeteranRecallInfo{}
		if ok := a.activityUtil[VeteranRecallID].IsActivity(); ok {
			if err := a.activityUtil[VeteranRecallID].GetInfo(info); err != nil {
				a.user.Error("GetInfo fail:", err)
				return
			}
		}

		msg.RecallSusNum = info.RecallSusNum
		for k, v := range info.RecallReward {
			if v == 0 {
				continue
			}
			msg.RewardIdList = append(msg.RewardIdList, k)
		}
	}

	log.Debug("syncVeteranRecallReward state:", state, " index:", index, " msg:", msg)
	if err := a.user.RPC(iserver.ServerTypeClient, "syncVeteranRecallReward", msg); err != nil {
		a.user.Error(err)
	}
}

// pickRecallReward 领取成功召回老兵的奖励
func (a *ActivityMgr) pickRecallRewardRes(id uint32) uint32 {
	if a.activeInfo[VeteranRecallID] == nil || a.activityUtil[VeteranRecallID] == nil {
		return 0
	}

	index := a.getCurActivityTime(a.activeInfo[VeteranRecallID])
	if index >= len(a.activeInfo[VeteranRecallID].tmStart) || index >= len(a.activeInfo[VeteranRecallID].tmEnd) || index >= len(a.activeInfo[VeteranRecallID].pickNum) {
		a.user.Warn("len(tmStart):", len(a.activeInfo[VeteranRecallID].tmStart), " len(tmEnd):", len(a.activeInfo[VeteranRecallID].tmEnd), " len(pickNum):", len(a.activeInfo[VeteranRecallID].pickNum), " index:", index+1)
		return 0
	}

	if time.Now().Unix() < a.activeInfo[VeteranRecallID].tmStart[index] || time.Now().Unix() > a.activeInfo[VeteranRecallID].tmEnd[index] {
		return 0
	}

	info := &db.VeteranRecallInfo{}
	info.RecallInfo = make(map[uint64]int64)
	info.RecallReward = make(map[uint32]uint32)

	info.Id = VeteranRecallID
	info.ActStartTm = a.activeInfo[VeteranRecallID].tmStart[index]

	if ok := a.activityUtil[VeteranRecallID].IsActivity(); ok {
		if err := a.activityUtil[VeteranRecallID].GetInfo(info); err != nil {
			a.user.Error("GetInfo fail:", err)
			return 0
		}
	}

	if ok1 := a.addRecallBag(id, index, info); ok1 {
		info.RecallReward[id] = 1
		if ok2 := a.activityUtil[VeteranRecallID].SetInfo(info); !ok2 {
			a.user.Error("SetInfo fail!")
			return 0
		}
	}

	a.user.Debug("pickRecallRewardRes info:", info)
	return 1
}

// AddRecallBag 添加领取的物品到包裹
func (a *ActivityMgr) addRecallBag(id uint32, index int, info *db.VeteranRecallInfo) bool {
	var totalNum uint32 = 0
	for i := 0; i < index; i++ {
		totalNum += a.activeInfo[VeteranRecallID].pickNum[i]
	}
	log.Debug("totalNum + id:", totalNum+id, " id:", id)

	callbackrewardData, ok := excel.GetCallbackreward(uint64(totalNum + id))
	if !ok {
		a.user.Error("GetCallbackreward fail!")
		return false
	}

	if callbackrewardData.ActivityNum != uint64(index)+1 {
		return false
	}

	if uint64(info.RecallSusNum) < callbackrewardData.CallNum {
		return false
	}

	if info.RecallReward[id] == 1 {
		return false
	}

	reward := strings.Split(callbackrewardData.Reward, ";")
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

		a.user.storeMgr.GetGoods(uint32(rewardType), uint32(rewardNum), common.RS_VeteranRecall, common.MT_NO, 0)
	}

	return true
}

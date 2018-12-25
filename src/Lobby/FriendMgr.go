package main

import (
	"common"
	"db"
	"protoMsg"
	"strconv"
	"time"
	"zeus/dbservice"
	"zeus/entity"
	"zeus/iserver"

	"github.com/garyburd/redigo/redis"
)

// FriendMgr 好友管理器
type FriendMgr struct {
	user       *LobbyUser
	platFrient map[uint64]string
	gameFriend uint32
}

// NewFriendMgr 获取好友管理器
func NewFriendMgr(user *LobbyUser) *FriendMgr {
	mgr := &FriendMgr{
		user:       user,
		platFrient: make(map[uint64]string),
		gameFriend: 0,
	}

	return mgr
}

func (mgr *FriendMgr) setFriendSum() {
	mgr.user.SetFriendsNum(uint32(len(mgr.platFrient)) + mgr.gameFriend)
}

// addFriend 添加好友
func (mgr *FriendMgr) addFriend(id uint64) {

	if !db.GetFriendUtil(mgr.user.GetDBID()).InApplyListByID(id) {
		// 请求申请id无效
		mgr.user.AdviceNotify(common.NotifyCommon, 2)
		mgr.syncApplyList()
		mgr.user.Warn("add friend failed, it is not in apply list, id: ", id)
		return
	}

	if db.GetFriendUtil(mgr.user.GetDBID()).IsReachLimit() {
		// 你的好友已达上限
		mgr.user.AdviceNotify(common.NotifyCommon, 3)
		mgr.user.Warn("add friend failed, my friend num reach limit, id: ", id)
		return
	}

	if db.GetFriendUtil(id).IsReachLimit() {
		// 对方好友已达上限
		mgr.user.AdviceNotify(common.NotifyCommon, 4)
		mgr.user.Warn("add friend failed, his friend num reach limit, id: ", id)
		return
	}

	applyInfo := db.GetFriendUtil(mgr.user.GetDBID()).GetSigleApplyReq(id)
	if applyInfo == nil {
		return
	}

	// 自己添加好友
	db.GetFriendUtil(mgr.user.GetDBID()).AddFriend(db.FriendInfo{ID: id, Name: applyInfo.Name})
	mgr.user.friendMgr.syncFriendList()

	// 删除申请请求
	mgr.delApplyReq(id)

	// 目标玩家添加好友
	db.GetFriendUtil(id).AddFriend(db.FriendInfo{ID: mgr.user.GetDBID(), Name: mgr.user.GetName()})

	// 删除目标玩家请求列表中申请信息
	db.GetFriendUtil(id).DelApply(mgr.user.GetDBID())

	// 更新目标玩家好友列表和申请列表
	mgr.SendProxyInfo(id, "SyncFriendList")
	mgr.SendProxyInfo(id, "SyncApplyList")

	mgr.user.Info("add friend success, id: ", id)
}

// SendProxyInfo 发送代理请求
func (mgr *FriendMgr) SendProxyInfo(targetID uint64, event string) {
	entityID, err := dbservice.SessionUtil(targetID).GetUserEntityID()
	if err != nil {
		return
	}

	srvID, spaceID, errSrv := dbservice.EntitySrvUtil(entityID).GetSrvInfo(iserver.ServerTypeGateway)
	if errSrv != nil {
		return
	}

	proxy := entity.NewEntityProxy(srvID, spaceID, entityID)
	proxy.RPC(common.ServerTypeLobby, event)
}

//WXAddFriend 微信添加好友
func (mgr *FriendMgr) WXAddFriend(targetUser *LobbyUser) {

	if targetUser == nil {
		return
	}

	// 你的好友已达上限
	if db.GetFriendUtil(mgr.user.GetDBID()).IsReachLimit() {
		mgr.friendApplyReq(targetUser.GetName())
		mgr.user.Warn("add weixin friend failed, my friend num reach limit, id: ", targetUser.GetDBID())
		return
	}

	// 对方好友已达上限
	if db.GetFriendUtil(targetUser.GetDBID()).IsReachLimit() {
		mgr.friendApplyReq(targetUser.GetName())
		mgr.user.Warn("add weixin friend failed, his friend num reach limit, id: ", targetUser.GetDBID())
		return
	}

	// 自己添加好友
	db.GetFriendUtil(mgr.user.GetDBID()).AddFriend(db.FriendInfo{ID: targetUser.GetDBID(), Name: targetUser.GetName()})
	mgr.user.friendMgr.syncFriendList()
	// 删除申请请求
	mgr.delApplyReq(targetUser.GetDBID())

	// 目标玩家添加好友
	db.GetFriendUtil(targetUser.GetDBID()).AddFriend(db.FriendInfo{ID: mgr.user.GetDBID(), Name: mgr.user.GetName()})

	// 更新目标玩家好友列表
	targetUser.friendMgr.delApplyReq(mgr.user.GetDBID())
	targetUser.friendMgr.syncFriendList()

	mgr.user.Info("add weixin friend success, id: ", targetUser.GetDBID())
}

// delFriend 删除好友
func (mgr *FriendMgr) delFriend(id uint64) {

	// 删除自己
	if db.GetFriendUtil(mgr.user.GetDBID()).DelFriend(id) {
		mgr.syncFriendList()

		//  删除对方
		if db.GetFriendUtil(id).DelFriend(mgr.user.GetDBID()) {
			mgr.SendProxyInfo(id, "SyncFriendList")
		}

		if db.PlayerInfoUtil(mgr.user.GetDBID()).IsTeacherPupil(id) {
			mgr.user.releaseTeacherPupil(id)
		}

		if mgr.user.isTmpBookedFriend(id) {
			db.PlayerTempUtil(id).DelTmpBookingFriend(mgr.user.GetDBID())
			db.PlayerTempUtil(mgr.user.GetDBID()).DelTmpBookedFriend(id)
		}

		if mgr.user.isBookedFriend(id) {
			proc := &LobbyUserMsgProc{user: mgr.user}
			proc.RPC_BattleBookingCancelReq(id)
		}
	} else {
		mgr.user.Warn("del friend failed, id: ", id)
	}

	mgr.user.Info("del friend success, id: ", id)
}

// getFriendInfo 获取好友的信息
func (mgr *FriendMgr) getFriendInfo(uid uint64) *protoMsg.FriendInfo {
	info := &protoMsg.FriendInfo{
		Id:        uid,
		Enterplat: "platform",
		Nickname:  "nickname",
	}

	args := []interface{}{
		"LogoutTime",
		"Picture",
		"QQVIP",
		"NickName",
		"GameEnter",
		"Name",
		"Gender",
		"Level",
		"Watchable",
	}

	values, err := dbservice.EntityUtil("Player", uid).GetValues(args)
	if err != nil {
		mgr.user.Error("GetValus err: ", err)
		return info
	}

	tmpUrl, err := redis.String(values[1], nil)
	if err == nil {
		info.Url = tmpUrl
	}

	tmpQQvip, err := redis.Int64(values[2], nil)
	if err == nil {
		info.Qqvip = uint32(tmpQQvip)
	}

	tmpNickname, err := redis.String(values[3], nil)
	if err == nil {
		info.Nickname = tmpNickname
	}

	tmpEnterplat, err := redis.String(values[4], nil)
	if err == nil {
		info.Enterplat = tmpEnterplat
	}

	name, err := redis.String(values[5], nil)
	if err == nil {
		info.Name_ = name
	}

	gender, err := redis.String(values[6], nil)
	if err == nil {
		if gender == "男" {
			info.Gender = 1
		} else if gender == "女" {
			info.Gender = 2
		}
	}

	level, err := redis.Uint64(values[7], nil)
	if err == nil {
		info.Level = uint32(level)
	}

	watchable, err := redis.Uint64(values[8], nil)
	if err == nil {
		info.Watchable = uint32(watchable)
	}

	// 根据是否存在Session表判断是否在线
	isOnline, err := dbservice.SessionUtil(uid).IsExisted()
	if err == nil && isOnline == false {
		info.State = common.StateOffline

		timeStr, err := redis.String(values[0], nil)
		if err == nil {
			tmpTime, err := strconv.ParseInt(timeStr, 10, 64)
			if err == nil {
				info.Time = uint32(tmpTime)
			}
		}
	} else if err == nil {
		info.State = uint32(db.PlayerTempUtil(uid).GetGameState())
		if info.State == common.StateGame {
			info.Time = uint32(db.PlayerTempUtil(uid).GetEnterGameTime())
		}
	}

	if db.PlayerInfoUtil(mgr.user.GetDBID()).IsTeacherPupil(uid) {
		info.Bound = true
	}

	info.NameColor = common.GetPlayerNameColor(uid)

	return info
}

// getPlatFriendInfo 获取平台好友的信息
func (mgr *FriendMgr) getPlatFriendInfo(openid string, uid uint64) *protoMsg.PlatFriendState {
	item := &protoMsg.PlatFriendState{
		Openid: openid,
		Uid:    uid,
		State:  common.StateOffline,
		Time:   0,
		Name_:  "",
		Level:  1,
	}

	args := []interface{}{
		"LogoutTime",
		"Name",
		"Level",
	}

	values, valueErr := dbservice.EntityUtil("Player", uid).GetValues(args)
	if valueErr != nil || len(values) != len(args) {
		return item
	}

	// 游戏中名称
	nameStr, nameErr := redis.String(values[1], nil)
	if nameErr == nil {
		item.Name_ = nameStr
	}

	if item.Name_ == "" {
		return item
	}

	level, err := redis.Uint64(values[2], nil)
	if err == nil {
		item.Level = uint32(level)
	}

	// 根据是否存在Session表判断是否在线
	isOnline, err := dbservice.SessionUtil(uid).IsExisted()
	if err == nil && isOnline == false {
		item.State = common.StateOffline

		// 登出时间
		timeStr, timeTmp := redis.String(values[0], nil)
		if timeTmp == nil {
			tmpTime, erro := strconv.ParseInt(timeStr, 10, 64)
			if erro == nil {
				item.Time = uint32(tmpTime)
			}
		}

	} else if err == nil {
		item.State = uint32(db.PlayerTempUtil(uid).GetGameState())

		// 游戏时间
		if item.State == common.StateGame {
			item.Time = uint32(db.PlayerTempUtil(uid).GetEnterGameTime())
		}
	}

	if db.PlayerInfoUtil(mgr.user.GetDBID()).IsTeacherPupil(uid) {
		item.Bound = true
	}

	item.NameColor = common.GetPlayerNameColor(uid)

	return item
}

// getBlackerInfo 获取拉黑对象的信息
func (mgr *FriendMgr) getBlackerInfo(uid uint64, stamp int64) *protoMsg.FriendInfo {
	info := &protoMsg.FriendInfo{
		Id:    uid,
		Stamp: uint64(stamp),
	}

	args := []interface{}{
		"Picture",
		"Name",
		"Gender",
	}

	values, err := dbservice.EntityUtil("Player", uid).GetValues(args)
	if err != nil {
		mgr.user.Error("GetValus err: ", err)
		return info
	}

	tmpUrl, err := redis.String(values[0], nil)
	if err == nil {
		info.Url = tmpUrl
	}

	name, err := redis.String(values[1], nil)
	if err == nil {
		info.Name_ = name
	}

	gender, err := redis.String(values[2], nil)
	if err == nil {
		if gender == "男" {
			info.Gender = 1
		} else if gender == "女" {
			info.Gender = 2
		}
	}

	info.NameColor = common.GetPlayerNameColor(uid)

	return info
}

// syncFriendList rpc同步好友信息
func (mgr *FriendMgr) syncFriendList() {
	retMsg := &protoMsg.SyncFriendList{}
	util := db.GetFriendUtil(mgr.user.GetDBID())

	for _, info := range util.GetFriendList() {
		retMsg.Item = append(retMsg.Item, mgr.getFriendInfo(info.ID))
	}

	for _, info := range util.GetBlackList() {
		retMsg.BlackList = append(retMsg.BlackList, mgr.getBlackerInfo(info.ID, info.Time))
	}

	mgr.gameFriend = uint32(len(retMsg.Item))
	mgr.setFriendSum()

	if err := mgr.user.RPC(iserver.ServerTypeClient, "SyncFriendList", retMsg); err != nil {
		mgr.user.Error("SyncFriendList err：", err)
	}
}

// friendApply 申请好友
func (mgr *FriendMgr) friendApply(targetID uint64) {
	result := 0
	if !db.PlayerInfoUtil(targetID).IsKeyExist() {
		targetID = 0
	}
	if targetID == 0 {
		result = 1
		mgr.user.AdviceNotify(common.NotifyCommon, 6)
	} else if targetID == mgr.user.GetDBID() {
		result = 1
		mgr.user.AdviceNotify(common.NotifyCommon, 54)
	} else {
		// 判断对方是否已经是好友
		if db.GetFriendUtil(mgr.user.GetDBID()).IsFriendByID(targetID) {
			result = 2
			mgr.user.AdviceNotify(common.NotifyCommon, 7)
		}

		// 已在对方申请列表中
		if db.GetFriendUtil(targetID).InApplyListByID(mgr.user.GetDBID()) {
			result = 3
			mgr.user.AdviceNotify(common.NotifyCommon, 8)
		}

		// 判断是否到达申请列表上线
		if db.GetFriendUtil(targetID).IsReachApplyLimit() {
			result = 4
			mgr.user.AdviceNotify(common.NotifyCommon, 10)
		}

	}

	if result == 0 {
		mgr.user.AdviceNotify(common.NotifyCommon, 5)

		info := db.ApplyInfo{
			ID:        mgr.user.GetDBID(),
			Name:      mgr.user.GetName(),
			ApplyTime: time.Now().Unix(),
		}

		// 申请信息添加至数据库
		db.GetFriendUtil(targetID).AddApplyInfo(info)

		// 更新目标玩家申请列表
		mgr.SendProxyInfo(targetID, "SyncApplyList")

		mgr.user.Info("apply friend success, name: ", info.Name)
	}

	// 好友申请请求结果 0申请成功 1用户名不存在 2已是好友 3已经申请 4达到申请列表上限
	mgr.user.RPC(iserver.ServerTypeClient, "FriendApplyReqRet", uint32(result))
}

// addApplyReq 添加申请请求
func (mgr *FriendMgr) friendApplyReq(name string) {
	targetID := db.GetIDByName(name)
	mgr.friendApply(targetID)
}

// delApplyReq 删除申请请求
func (mgr *FriendMgr) delApplyReq(id uint64) {

	if !db.GetFriendUtil(mgr.user.GetDBID()).DelApply(id) {
		mgr.user.Warn("del apply failed, id: ", id)
	}

	mgr.syncApplyList()
	mgr.user.Info("del apply success, id: ", id)
}

// syncApplyList 同步申请列表
func (mgr *FriendMgr) syncApplyList() {

	retMsg := &protoMsg.SyncFriendApplyList{}

	list := make([]*db.ApplyInfo, 0)
	list = db.GetFriendUtil(mgr.user.GetDBID()).GetApplyList()

	for _, info := range list {

		item := protoMsg.FriendApplyInfo{
			Id:        info.ID,
			Name_:     info.Name,
			ApplyTime: info.ApplyTime,
			Url:       "",
			Level:     1,
		}

		args := []interface{}{
			"Picture",
			"Name",
			"Level",
		}
		values, valueErr := dbservice.EntityUtil("Player", info.ID).GetValues(args)
		if valueErr != nil || len(values) != len(args) {
			continue
		}
		urlStr, urlErr := redis.String(values[0], nil)
		if urlErr == nil {
			item.Url = urlStr
		}
		name, nameErr := redis.String(values[1], nil)
		if nameErr == nil {
			item.Name_ = name
		}
		level, err := redis.Uint64(values[2], nil)
		if err == nil {
			item.Level = uint32(level)
		}

		item.NameColor = common.GetPlayerNameColor(info.ID)

		retMsg.Item = append(retMsg.Item, &item)
	}

	if err := mgr.user.RPC(iserver.ServerTypeClient, "SyncApplyList", retMsg); err != nil {
		mgr.user.Error("SyncApplyList err: ", err)
	}

}

// InitPlatFriendList 初始化好友平台好友
func (mgr *FriendMgr) InitPlatFriendList(msg *protoMsg.PlatFriendStateReq) {
	if msg == nil {
		return
	}

	retMsg := &protoMsg.PlatFriendStateRet{}
	platFrientID := make([]uint64, 0)

	for _, openid := range msg.Openid {

		//log.Debug("info  openid:", openid)
		uid, err := dbservice.GetUID(openid)
		if err != nil || uid == 0 {
			continue
		}

		retMsg.Data = append(retMsg.Data, mgr.getPlatFriendInfo(openid, uid))
		platFrientID = append(platFrientID, uid)

		mgr.platFrient[uid] = openid

		//log.Debugf("item info: %+v", item)
	}

	// 平台好友数据存入数据库
	db.GetFriendUtil(mgr.user.GetDBID()).UpdatePlatFrientInfo(platFrientID)

	// 设置好友数量
	mgr.setFriendSum()

	if err := mgr.user.RPC(iserver.ServerTypeClient, "PlatFriendStateRet", retMsg); err != nil {
		mgr.user.Error("RPC PlatFriendStateRet err: ", err)
	}

	curState := db.PlayerTempUtil(mgr.user.GetDBID()).GetGameState()
	mgr.syncFriendState(curState)
	//log.Debug("初始化好友平台好友数量 reqSum:", len(msg.Openid), " sum:", len(mgr.platFrient))
}

// syncPlatFriendList 同步平台好友列表
func (mgr *FriendMgr) syncPlatFriendList() {
	retMsg := &protoMsg.PlatFriendStateRet{}

	for uid, openid := range mgr.platFrient {
		retMsg.Data = append(retMsg.Data, mgr.getPlatFriendInfo(openid, uid))
	}

	mgr.user.RPC(iserver.ServerTypeClient, "PlatFriendStateRet", retMsg)
}

// syncFriendState 同步好友在线状态
func (mgr *FriendMgr) syncFriendState(state uint64) {
	list := db.GetFriendUtil(mgr.user.GetDBID()).GetFriendList()

	for _, info := range list {
		mgr.sendMsgToFriend(info.ID, common.ServerTypeLobby, "SyncFriendState", mgr.user.GetDBID(), state)
	}

	for id := range mgr.platFrient {
		mgr.sendMsgToFriend(id, common.ServerTypeLobby, "SyncFriendState", mgr.user.GetDBID(), state)
	}

}

// syncFriendName 同步好友的名字及颜色
func (mgr *FriendMgr) syncFriendName() {
	name := mgr.user.GetName()
	color := common.GetPlayerNameColor(mgr.user.GetDBID())
	list := db.GetFriendUtil(mgr.user.GetDBID()).GetFriendList()

	for _, info := range list {
		mgr.sendMsgToFriend(info.ID, iserver.ServerTypeClient, "SyncFriendName", mgr.user.GetDBID(), name, color)
	}

	for id, _ := range mgr.platFrient {
		mgr.sendMsgToFriend(id, iserver.ServerTypeClient, "SyncFriendName", mgr.user.GetDBID(), name, color)
	}
}

// syncWatchable 同步是否允许观战
func (mgr *FriendMgr) syncWatchable(watchable uint32) {
	list := db.GetFriendUtil(mgr.user.GetDBID()).GetFriendList()

	for _, info := range list {
		mgr.sendMsgToFriend(info.ID, iserver.ServerTypeClient, "SyncWatchable", mgr.user.GetDBID(), watchable)
	}

	for id, _ := range mgr.platFrient {
		mgr.sendMsgToFriend(id, iserver.ServerTypeClient, "SyncWatchable", mgr.user.GetDBID(), watchable)
	}
}

// sendMsgToFriend 向好友发送消息
func (mgr *FriendMgr) sendMsgToFriend(uid uint64, srvType uint8, method string, args ...interface{}) {
	entityID, err := dbservice.SessionUtil(uid).GetUserEntityID()
	if err != nil {
		return
	}

	srvID, spaceID, errSrv := dbservice.EntitySrvUtil(entityID).GetSrvInfo(iserver.ServerTypeGateway)
	if errSrv != nil {
		return
	}

	proxy := entity.NewEntityProxy(srvID, spaceID, entityID)
	proxy.RPC(srvType, method, args...)
}

//  recommendFriends 推荐好友
func (mgr *FriendMgr) recommendFriends() {
	retMsg := &protoMsg.SyncFriendList{}

	GetSrvInst().TravsalEntity("Player", func(e iserver.IEntity) {
		if e == nil {
			return
		}

		if len(retMsg.Item) >= 10 {
			return
		}

		if user, ok := e.(*LobbyUser); ok {
			if user.GetDBID() == mgr.user.GetDBID() {
				return
			}

			if user.GetName() == "" {
				return
			}

			if mgr.isFriend(user.GetDBID()) {
				return
			}

			if db.GetFriendUtil(user.GetDBID()).InApplyListByID(mgr.user.GetDBID()) {
				return
			}

			if mgr.getFriendState(user.GetDBID()) == common.StateOffline {
				return
			}

			retMsg.Item = append(retMsg.Item, mgr.getFriendInfo(user.GetDBID()))
		}
	})

	mgr.user.RPC(iserver.ServerTypeClient, "GetRecommendListRet", retMsg)
}

// isFriend 是否好友
func (mgr *FriendMgr) isFriend(uid uint64) bool {
	list := db.GetFriendUtil(mgr.user.GetDBID()).GetFriendList()

	for _, info := range list {
		if info.ID == uid {
			return true
		}
	}

	for id, _ := range mgr.platFrient {
		if id == uid {
			return true
		}
	}

	return false
}

// isGameFriend 是否游戏好友
func (mgr *FriendMgr) isGameFriend(uid uint64) bool {
	list := db.GetFriendUtil(mgr.user.GetDBID()).GetFriendList()

	for _, info := range list {
		if info.ID == uid {
			return true
		}
	}

	return false
}

// getFriendState 获取好友的在线状态
func (mgr *FriendMgr) getFriendState(uid uint64) int {
	var state int

	isOnline, err := dbservice.SessionUtil(uid).IsExisted()
	if err == nil && isOnline == false {
		state = common.StateOffline
	} else if err == nil {
		state = int(db.PlayerTempUtil(uid).GetGameState())
	}

	return state
}

// getFriendName 获取好友的名字
func (mgr *FriendMgr) getFriendName(uid uint64) string {
	name, err := redis.String(dbservice.EntityUtil("Player", uid).GetValue("Name"))
	if err != nil {
		mgr.user.Error("String err: ", err)
		return ""
	}

	return name
}

// getFriendLevel 获取好友的军衔等级
func (mgr *FriendMgr) getFriendLevel(uid uint64) uint32 {
	level, err := redis.Uint64(dbservice.EntityUtil("Player", uid).GetValue("Level"))
	if err != nil {
		mgr.user.Error("Uint64 err: ", err)
		return 0
	}

	return uint32(level)
}

// getFriendWatchable 获取好友的观战设置
func (mgr *FriendMgr) getFriendWatchable(uid uint64) uint32 {
	watchable, err := redis.Uint64(dbservice.EntityUtil("Player", uid).GetValue("Watchable"))
	if err != nil {
		mgr.user.Error("Uint64 err: ", err)
		return 0
	}

	return uint32(watchable)
}

// isBindableFriend 是否为可绑定好友
func (mgr *FriendMgr) isBindableFriend(uid uint64) bool {
	if !mgr.isFriend(uid) {
		return false
	}

	if db.PlayerInfoUtil(mgr.user.GetDBID()).IsTeacherPupil(uid) {
		return false
	}

	minLevel := uint32(common.GetTBSystemValue(common.System_RecruitInitLevel))
	return mgr.getFriendLevel(uid) >= minLevel
}

// recommendBindableFriends 推荐绑定好友
func (mgr *FriendMgr) recommendBindableFriends() {
	retMsg := &protoMsg.SyncFriendList{}
	list := db.GetFriendUtil(mgr.user.GetDBID()).GetFriendList()

	for _, info := range list {
		if mgr.isBindableFriend(info.ID) {
			retMsg.Item = append(retMsg.Item, mgr.getFriendInfo(info.ID))
		}
	}

	for id, _ := range mgr.platFrient {
		if mgr.isBindableFriend(id) {
			retMsg.Item = append(retMsg.Item, mgr.getFriendInfo(id))
		}
	}

	mgr.user.RPC(iserver.ServerTypeClient, "GetBindListRet", retMsg)
}

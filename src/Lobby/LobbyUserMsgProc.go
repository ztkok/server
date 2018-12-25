package main

import (
	"common"
	"db"
	"excel"
	"fmt"
	"math"
	"msdk"
	"protoMsg"
	"strings"
	"time"
	"zeus/dbservice"
	"zeus/entity"
	"zeus/iserver"
	"zeus/msgdef"
	"zeus/serverMgr"
	"zeus/tlog"
	"zeus/tsssdk"

	"github.com/garyburd/redigo/redis"
	"github.com/spf13/viper"
)

// LobbyUserMsgProc LobbyUser的消息处理函数
type LobbyUserMsgProc struct {
	user *LobbyUser
}

func (proc *LobbyUserMsgProc) RPC_CheckMatchOpen(param uint32) {
	single := viper.GetBool("Lobby.Solo")
	two := viper.GetBool("Lobby.Duo")
	four := viper.GetBool("Lobby.Squad")
	proc.user.RPC(iserver.ServerTypeClient, "CheckMatchOpen", param, single, two, four)
}

// RPC_SwitchMatchMode C->S 客户端切换匹配模式
func (proc *LobbyUserMsgProc) RPC_SwitchMatchModeReq(mode uint32) {
	if proc.user.matchMode == mode {
		return
	}

	var ret uint8 = 1

	//验证该匹配模式是否开放
	if !common.IsMatchModeOk(mode, 1) &&
		!common.IsMatchModeOk(mode, 2) &&
		!common.IsMatchModeOk(mode, 4) {
		ret = 0
	}

	if ret == 1 {
		proc.user.delMatch()
		proc.user.MemberQuitTeam()
		proc.user.matchMode = mode
		db.PlayerTempUtil(proc.user.GetDBID()).SetPlayerMatchMode(mode)
	}

	proc.user.RPC(iserver.ServerTypeClient, "SwitchMatchModeRet", ret)
	proc.user.Info("SwitchMatchModeReq success, matchMode: ", mode)
}

// RPC_EnterRoomReq C->S 客户端进入单人匹配队列
func (proc *LobbyUserMsgProc) RPC_EnterRoomReq(mapid uint32) {
	if proc.user.GetPlayerGameState() != common.StateFree {
		proc.user.Warn("Recv EnterRoomReq but user game state not free ", proc.user.GetPlayerGameState())
		return
	}

	if !common.IsMatchModeOk(proc.user.matchMode, 1) {
		proc.user.AdviceNotify(common.NotifyCommon, 22)
		proc.user.Warn("EnterRoomReq failed, match mode is not open, matchMode: ", proc.user.matchMode, " matchTyp: ", 1)
		return
	}

	//勇者战场
	if proc.user.matchMode == common.MatchModeBrave && !proc.user.CanMatchBraveGame(1) {
		return
	}

	// 获取匹配管理器
	if proc.user.matchMgrSoloProxy == nil {
		srvInfo, err := serverMgr.GetServerMgr().GetServerByType(common.ServerTypeMatch)
		if err != nil {
			proc.user.Error("EnterRoomReq failed, GetServerByType err: ", err)
			return
		}

		name := fmt.Sprintf("%s:%d", common.MatchMgrSolo, srvInfo.ServerID)
		proc.user.matchMgrSoloProxy = entity.GetEntityProxy(name)
		if proc.user.matchMgrSoloProxy == nil {
			proc.user.Error("EnterRoomReq failed, matchMgrSoloProxy is nil")
			return
		}
	}

	ranks := proc.user.GetRanksByMode(proc.user.matchMode, 1)
	if ranks == nil {
		return
	}

	//普通模式之外的其他模式使用的地图均从赛事模式配置表中读取
	if proc.user.matchMode == common.MatchModeNormal {
		if mapid != 1 && mapid != 2 {
			mapid = 1
		}
	} else {
		info := common.GetOpenModeInfo(proc.user.matchMode, 1)
		if info == nil {
			return
		}

		mapid = info.MapId
	}

	_, ok := excel.GetMaps(uint64(mapid))
	if !ok {
		proc.user.Error("EnterRoomReq failed, map doesn't exist, mapid: ", mapid)
		return
	}

	if err := db.PlayerTempUtil(proc.user.GetDBID()).SetPlayerMapID(mapid); err != nil {
		proc.user.Error("EnterRoomReq failed, SetPlayerMapID err: ", err)
		return
	}

	color := common.GetPlayerNameColor(proc.user.GetDBID())
	weapon := proc.user.GetOutsideWeapon()

	if err := proc.user.matchMgrSoloProxy.RPC(common.ServerTypeMatch, "EnterSoloQueue",
		GetSrvInst().GetSrvID(), proc.user.GetID(), mapid, uint32(1000), uint32(ranks[1]),
		proc.user.GetName(), proc.user.GetRoleModel(), proc.user.GetDBID(), proc.user.matchMode, proc.user.GetVeteran(), color, weapon); err != nil {
		proc.user.Error("RPC EnterSoloQueue err: ", err)
		return
	}

	proc.user.Info("Send EnterRoomReq to match success, mapid: ", mapid, " matchMode: ", proc.user.matchMode)
}

func (proc *LobbyUserMsgProc) RPC_EnterSoloQueueRet(result uint32, expectTime uint64) {
	if result == 0 {
		proc.user.RPC(iserver.ServerTypeClient, "ExpectTime", expectTime)
		proc.user.SetPlayerGameState(common.StateMatching)
		proc.user.Info("Enter solo queue success, matchMode: ", proc.user.matchMode, " expectTime: ", expectTime)
	} else {
		proc.user.Warn("EnterSoloQueueRet failed, result: ", result)
		proc.user.tlogMatchFlow(0, 1, 0, 0, 0, 0, 0)
	}
}

// RPC_CancelEnterRoom 取消单人匹配
func (proc *LobbyUserMsgProc) RPC_CancelEnterRoom() {
	teamID := proc.user.GetTeamID()
	if teamID != 0 {
		if proc.user.teamMgrProxy == nil {
			return
		}
		proc.user.teamMgrProxy.RPC(common.ServerTypeMatch, "CancelQueue", teamID)
	} else {
		if proc.user.matchMgrSoloProxy == nil {
			proc.user.Error("CancelEnterRoom failed, matchMgrSoloProxy is nil")
			return
		}
		if err := proc.user.matchMgrSoloProxy.RPC(common.ServerTypeMatch, "CancelSoloQueue",
			GetSrvInst().GetSrvID(), proc.user.GetID()); err != nil {
			proc.user.Error("RPC CancelSoloQueue err: ", err)
			return
		}
	}
	proc.user.Info("Send CancelEnterRoom to match success, matchMode: ", proc.user.matchMode)
}

func (proc *LobbyUserMsgProc) RPC_CancelSoloQueueRet(result uint32) {
	if result == 0 {
		// 如果玩家已经下线, 则不再切换状态
		if proc.user.gameState != common.StateOffline {
			proc.user.SetPlayerGameState(common.StateFree)
		}
		proc.user.RPC(iserver.ServerTypeClient, "CancelSingleResult", uint32(1)) // 0 失败 1 成功
		proc.user.Info("Cancel solo queue success, matchMode: ", proc.user.matchMode)

		util := db.PlayerTempUtil(proc.user.GetDBID())
		toJoinTeam := util.GetToJoinTeam()

		if toJoinTeam != 0 {
			util.SetToJoinTeam(0)
			proc.user.pullUserToTeam(proc.user.GetDBID(), toJoinTeam)
		}

		proc.user.tlogMatchFlow(0, 1, 0, 2, 0, 0, 0)
	} else {
		proc.user.RPC(iserver.ServerTypeClient, "CancelSingleResult", uint32(0))
		proc.user.Info("Cancel solo queue failed, matchMode: ", proc.user.matchMode)

		proc.user.tlogMatchFlow(0, 1, 0, 0, 0, 0, 0)
	}
}

// RPC_SetRoleModel 设置角色模型id
func (proc *LobbyUserMsgProc) RPC_SetRoleModel(id uint32) {
	if proc.user == nil {
		return
	}

	if proc.user.GetPlayerGameState() == common.StateGame {
		proc.user.Error("SetRoleModel failed, change role model is not allowed when gaming")
		return
	}

	// 商品信息
	goodsConfig, ok := excel.GetStore(uint64(id))
	if !ok {
		proc.user.Error("SetRoleModel failed, goods doesn't exist, id: ", id)
		return
	}

	// 商品类型
	if goodsConfig.Type != GoodsRoleType {
		proc.user.Error("SetRoleModel failed, goods doesn't support set role model, id: ", id)
		return
	}

	// 是否可使用
	if goodsConfig.State != GoodsFree {
		isBuy := db.PlayerGoodsUtil(proc.user.GetDBID()).IsOwnGoods(id)
		if isBuy == false {
			proc.user.Error("SetRoleModel failed, goods is not available, id: ", id)
			return
		}
	}

	proc.user.SetRoleModel(uint32(goodsConfig.RelationID))
	proc.user.SetGoodsRoleModel(id)

	if proc.user.teamMgrProxy != nil && proc.user.GetTeamID() != 0 {
		if err := proc.user.teamMgrProxy.RPC(common.ServerTypeMatch, "SetRoleModel",
			proc.user.GetID(), proc.user.GetTeamID(), uint32(goodsConfig.RelationID)); err != nil {
			proc.user.Error("RPC SetRoleModel err: ", err)
			return
		}
	}
	proc.user.setPreferenceSwitch(uint32(goodsConfig.Type), 1, false)

	proc.user.Info("Set role model success, id: ", id)
}

// RPC_SetParachuteID 设置伞包id
func (proc *LobbyUserMsgProc) RPC_SetParachuteID(id uint32) {

	if proc.user == nil {
		return
	}

	// 商品信息
	goodsConfig, ok := excel.GetStore(uint64(id))
	if !ok {
		proc.user.Error("SetParachuteID failed, goods doesn't exist, id: ", id)
		return
	}

	// 商品类型
	if goodsConfig.Type != GoodsParachuteType {
		proc.user.Error("SetParachuteID failed, goods is not a parachute, id: ", id)
		return
	}

	// 是否可使用
	if goodsConfig.State != GoodsFree {
		isBuy := db.PlayerGoodsUtil(proc.user.GetDBID()).IsOwnGoods(id)
		if isBuy == false {
			proc.user.Error("SetParachuteID failed, goods is not available, id: ", id)
			return
		}
	}

	proc.user.SetParachuteID(uint32(goodsConfig.RelationID))
	proc.user.SetGoodsParachuteID(id)
	proc.user.setPreferenceSwitch(uint32(goodsConfig.Type), 1, false)

	proc.user.Info("Set parachute id success, id: ", id)
}

// RPC_SetOutsideWeapon 设置用于在大厅展示的场景外的武器
func (proc *LobbyUserMsgProc) RPC_SetOutsideWeapon(id uint32) {
	proc.user.SetOutsideWeapon(id)

	if proc.user.teamMgrProxy != nil && proc.user.GetTeamID() != 0 {
		if err := proc.user.teamMgrProxy.RPC(common.ServerTypeMatch, "SetOutsideWeapon",
			proc.user.GetID(), proc.user.GetTeamID(), id); err != nil {
			proc.user.Error("RPC SetOutsideWeapon err: ", err)
			return
		}
	}
}

// RPC_ChangeName 修改玩家名称
func (proc *LobbyUserMsgProc) RPC_ChangeName(name string) {
	if proc.user == nil {
		return
	}
	if proc.user.GetName() == name {
		return
	}
	result := 0
	util := db.PlayerInfoUtil(proc.user.GetDBID())
	num, _ := util.GetChangeNameNum()
	var cost, day uint64
	excelData, err := excel.GetChangename(num)
	if !err {
		return
	}
	day = excelData.Length
	cost = excelData.Cost
	excelData, err = excel.GetChangename(num + 1)
	if err {
		cost = excelData.Cost
		num = num + 1
	}

	stamp, _ := util.GetChangeNameTime()
	cdTime := day * 86400
	if stamp+cdTime > uint64(time.Now().Unix()) {
		return
	}
	if proc.user.GetCoin() < cost {
		return
	}
	if ret, err := tsssdk.JudgeUserInputName(name); !ret {
		if err != nil {
			proc.user.Error("SetName failed, JudgeUserInputName err: ", err)
		}
		result = 2
	} else if !db.JudgeNameInUse(name) {
		db.DelUsedName(proc.user.GetName())
		proc.user.SetName(name)
		db.AddUsedName(name, proc.user.GetDBID())
		util.SetChangeNameNum(num)
		proc.user.storeMgr.reduceMoney(common.MT_MONEY, common.RS_ChangeName, cost)
		proc.user.ChangeNameInfo()
		if proc.user.teamMgrProxy != nil && proc.user.GetTeamID() != 0 {
			if err := proc.user.teamMgrProxy.RPC(common.ServerTypeMatch, "ChangeTeamMemberName",
				proc.user.GetID(), proc.user.GetTeamID(), name); err != nil {
				proc.user.Error("RPC ChangeTeamMemberName err: ", err)
				return
			}
		}
		proc.user.friendMgr.syncFriendName()
		result = 1
	}
	proc.user.RPC(iserver.ServerTypeClient, "ChangeNameResult", uint64(result))
}

// RPC_SetName 设置角色名称
func (proc *LobbyUserMsgProc) RPC_SetName(name string) {
	if proc.user == nil {
		return
	}

	result := 0
	if ret, err := tsssdk.JudgeUserInputName(name); !ret {
		if err != nil {
			proc.user.Error("SetName failed, JudgeUserInputName err: ", err)
		}
		result = 2
	} else if !db.JudgeNameInUse(name) {
		proc.user.SetName(name)
		db.AddUsedName(name, proc.user.GetDBID())

		util := db.PlayerInfoUtil(proc.user.GetDBID())
		util.SetChangeNameNum(1)

		proc.user.ChangeNameInfo()
		result = 1
		if proc.user.teamMgrProxy != nil && proc.user.GetTeamID() != 0 {
			if err := proc.user.teamMgrProxy.RPC(common.ServerTypeMatch, "ChangeTeamMemberName",
				proc.user.GetID(), proc.user.GetTeamID(), name); err != nil {
				proc.user.Error("RPC ChangeTeamMemberName err: ", err)
				return
			}
		}
	}

	//qqscorebatch 角色名称
	if proc.user.loginMsg != nil && proc.user.loginMsg.LoginChannel == 2 {
		msdk.QQScoreBatch(common.QQAppIDStr, common.MSDKKey, proc.user.loginMsg.VOpenID, proc.user.GetAccessToken(), proc.user.loginMsg.PlatID,
			name, proc.user.GetDBID(), 0, 0, "", "")
	}

	proc.user.RPC(iserver.ServerTypeClient, "SetNameResult", uint64(result)) // 0 失败(名字已被使用) 1 成功

	proc.user.Info("Set name,  name: ", name, " result: ", result)
}

// RPC_SetNoviceType 客户端在玩家选项界面勾选新手类型
func (proc *LobbyUserMsgProc) RPC_SetNoviceType(typ uint8) {
	db.PlayerInfoUtil(proc.user.GetDBID()).SetNoviceType(typ)
}

// RPC_SetNoviceProgress 客户端在进行新手引导的过程中，上报当前进度
func (proc *LobbyUserMsgProc) RPC_SetNoviceProgress(prog uint8) {
	proc.user.tlogGuideFlow(prog)
}

// RPC_SeasonRankInfoReq //赛季排行信息请求
func (proc *LobbyUserMsgProc) RPC_SeasonRankInfoReq() {
	proc.user.syncSeasonRankInfo()
}

// RPC_DrawSeasonAwardsReq 客户端领取赛季奖励
func (proc *LobbyUserMsgProc) RPC_DrawSeasonAwardsReq() {
	proc.user.DrawSeasonAwards()
}

// 设置跳过飞机
func (proc *LobbyUserMsgProc) RPC_JumpAir(set uint64) {
	db.PlayerTempUtil(proc.user.GetDBID()).SetPlayerJumpAir(set)
}

// RPC_EnterRoomSuccess RoomUser创建成功
func (proc *LobbyUserMsgProc) RPC_EnterRoomSuccess() {
	// 这条消息只在压测模式下用到,外网不需要给客户端发这条消息,节省流量
	if viper.GetBool("Config.Stress") {
		cost := time.Now().Sub(proc.user.enterRoomStamp).Nanoseconds() / 1000000
		proc.user.RPC(iserver.ServerTypeClient, "TransResult", "EnterSpace", true, cost)
	}

	// 进入房间成功, 停止清除观战信息的计时器, 发送观战目标信息
	if proc.user.watchTarget != 0 {
		proc.user.clearWatchTimer.Stop()
		proc.user.RPC(common.ServerTypeRoom, "WatchTargetBattle", proc.user.watchTarget)
		proc.user.SetPlayerGameState(common.StateWatch)
	} else {
		proc.user.RPC(common.ServerTypeRoom, "EnterUserBattle")
	}
}

// RPC_AddFriend rpc添加好友
func (proc *LobbyUserMsgProc) RPC_AddFriend(id uint64) {
	proc.user.friendMgr.addFriend(id)
}

// RPC_DelFriend rpc删除好友
func (proc *LobbyUserMsgProc) RPC_DelFriend(id uint64) {
	proc.user.friendMgr.delFriend(id)

	proc.user.RPC(iserver.ServerTypeClient, "DelOneFriend", id)
}

// RPC_SyncFriendList 同步好友列表
func (proc *LobbyUserMsgProc) RPC_SyncFriendList() {
	proc.user.friendMgr.syncFriendList()
}

// RPC_SyncPlatFriendList 同步平台好友列表
func (proc *LobbyUserMsgProc) RPC_SyncPlatFriendList() {
	proc.user.friendMgr.syncPlatFriendList()
}

// RPC_FriendApplyReq rpc添加申请请求
func (proc *LobbyUserMsgProc) RPC_FriendApplyReq(name string) {
	proc.user.friendMgr.friendApplyReq(name)
}

// RPC_FriendApplyReq rpc添加申请请求
func (proc *LobbyUserMsgProc) RPC_FriendApplyIdReq(id uint64) {
	proc.user.friendMgr.friendApply(id)
}

// RPC_DelApplyReq rpc删除申请请求
func (proc *LobbyUserMsgProc) RPC_DelApplyReq(id uint64) {
	proc.user.friendMgr.delApplyReq(id)
}

// RPC_SyncApplyList rpc同步申请列表
func (proc *LobbyUserMsgProc) RPC_SyncApplyList() {
	proc.user.friendMgr.syncApplyList()
}

// RPC_SyncFriendState 同步好友状态
func (proc *LobbyUserMsgProc) RPC_SyncFriendState(friendID uint64, state uint64) {
	curtime := time.Now().Unix()
	proc.user.RPC(iserver.ServerTypeClient, "SyncFriendState", friendID, uint32(state), uint32(curtime))
}

// RPC_GetRecommendListReq 获取好友推荐列表请求
func (proc *LobbyUserMsgProc) RPC_GetRecommendListReq() {
	proc.user.friendMgr.recommendFriends()
}

// RPC_InviteUpLineReport 邀请好友上线上报
func (proc *LobbyUserMsgProc) RPC_InviteUpLineReport(uid uint64) {
	util := db.PlayerInfoUtil(proc.user.GetDBID())
	today := common.GetTodayBeginStamp()

	if !util.IsFriendInvited(today, uid) {
		util.AddInviteUpLineFriend(today, uid)
		proc.user.updateDayTaskItems(common.TaskItem_InviteUpLine, 1)
		proc.user.taskMgr.updateTaskItemsAll(common.TaskItem_InviteUpLine, 1)
	}
}

// RPC_MatchSuccess 匹配成功, 通知客户端加载地图
func (proc *LobbyUserMsgProc) RPC_MatchSuccess(spaceID uint64, mapid, skybox uint32, mode, realNum uint32, matchTime int64, stateEnterTeam uint32) {
	if proc.user.gm.isUseGmSkyType != 0 {
		skybox = proc.user.gm.gmSkyType
	}

	proc.user.spaceID = spaceID
	proc.user.skyBox = skybox
	proc.user.mapId = mapid

	proc.user.RPC(iserver.ServerTypeClient, "MatchSuccess", mapid, skybox)
	proc.user.RPC(iserver.ServerTypeClient, "RecvMatchModeId", proc.user.matchMode)

	proc.user.SetPlayerGameState(common.StateGame)

	util := db.PlayerTempUtil(proc.user.GetDBID())
	util.SetEnterGameTime(uint64(time.Now().Unix()))
	util.SetPlayerSpaceID(spaceID)
	util.ClearTeamCustoms()

	//更新每日任务项(与好友一起参加比赛)进度
	if stateEnterTeam == 2 {
		proc.user.updateDayTaskItems(common.TaskItem_FriendGame, 1)
		proc.user.taskMgr.updateTaskItemsAll(common.TaskItem_FriendGame, 1)
	}

	if proc.user.matchMode == common.MatchModeBrave {
		proc.user.UpdateBraveRecord(mode)
	}

	//游戏期间不通知客户端过期的时限道具
	if proc.user.cronTask != nil {
		proc.user.cronTask.Stop()
	}

	proc.user.tlogMatchFlow(mode, 1, matchTime, 1, realNum, skybox, stateEnterTeam)
	proc.user.Info("Match success, matchMode: ", proc.user.matchMode, " mapid: ", mapid, " skybox: ", skybox, " spaceid: ", spaceID)
}

// RPC_EnterScene 通知客户端进入房间
func (proc *LobbyUserMsgProc) RPC_EnterScene(spaceid uint64) {
	proc.user.EnterSpace(spaceid)
	proc.user.Info("Enter scene success, matchMode: ", proc.user.matchMode)
}

func (proc *LobbyUserMsgProc) RPC_leaveSpace() {
	if proc.user.watchTarget != 0 {
		proc.user.watchTarget = 0
		proc.user.WatchFlow(WatchLeave, proc.user.watchTarget, proc.user.spaceID, 0, uint32(time.Now().Unix()-proc.user.watchStart), 1)
	}
	proc.user.spaceID = 0
	proc.user.Info("Leave space success, matchmode: ", proc.user.matchMode)
	if proc.user.GetTeamID() == 0 {
		proc.user.SetPlayerGameState(common.StateFree)
	} else {
		proc.user.SetPlayerGameState(common.StateMatchWaiting)
	}
	//玩家在游戏期间过期的时限道具，在游戏结束后通知给客户端
	proc.user.StartCrondForExpireCheck()

	proc.user.syncDataToMatch()
	proc.user.clearTmpBookingFriends()

	// 随机刷新偏好道具(角色和伞包)
	for i := 1; i <= 2; i++ {
		proc.user.preItemRandomCreate(uint32(i), 1)
	}
	// 随机刷新偏好道具(背包和头盔)
	for i := 5; i <= 6; i++ {
		for j := 1; j <= 3; j++ {
			proc.user.preItemRandomCreate(uint32(i), uint32(j))
		}
	}
}

// RPC_GmLobby Gm命令
func (proc *LobbyUserMsgProc) RPC_GmLobby(paras string) {
	proc.user.Debug("RPC_GmLobby: ", paras)
	proc.user.gm.exec(paras)
}

func (proc *LobbyUserMsgProc) RPC_AddCoin(num uint32) {
	proc.user.storeMgr.addMoney(common.MT_MONEY, common.RS_Battle, num)
}

func (proc *LobbyUserMsgProc) RPC_GmAddCoin(num uint32) {
	proc.user.SetCoin(proc.user.GetCoin() + uint64(num))
}

func (proc *LobbyUserMsgProc) RPC_AddDiam(num, reason uint32) {
	proc.user.storeMgr.limitAddDiam(num, reason)
}

func (proc *LobbyUserMsgProc) RPC_GmAddDiam(num uint32) {
	proc.user.SetDiam(proc.user.GetDiam() + uint64(num))
}

func (proc *LobbyUserMsgProc) RPC_AddBraveCoin(num uint32) {
	proc.user.storeMgr.addMoney(common.MT_BraveCoin, common.RS_Battle, num)
}

func (proc *LobbyUserMsgProc) RPC_GmAddBraveCoin(num uint32) {
	proc.user.SetBraveCoin(proc.user.GetBraveCoin() + uint64(num))
}

// RPC_AddExp 玩家在Room服对局结束后获得经验值
func (proc *LobbyUserMsgProc) RPC_AddExp(incr uint32) {
	proc.user.UpdateMilitaryRank(incr)
}

// RPC_UpdateAchievement 玩家在Room服对局结束后更新成就
func (proc *LobbyUserMsgProc) RPC_UpdateAchievement(msg *protoMsg.Uint32Array) {
	for _, v := range msg.GetList() {
		proc.user.announcementBroad(1, v, 0, 0)
	}
}

// RPC_QueryExpCardReq 请求查询经验卡使用情况
func (proc *LobbyUserMsgProc) RPC_QueryExpCardReq() {
	factor, leftTime := db.PlayerInfoUtil(proc.user.GetDBID()).GetUsingExpCardInfo()
	proc.user.RPC(iserver.ServerTypeClient, "QueryExpCardRet", factor, leftTime)
}

// RPC_SelectExpCardReq 请求选用经验卡
func (proc *LobbyUserMsgProc) RPC_SelectExpCardReq(id uint32) {
	ret := uint8(0)
	goodsUtil := db.PlayerGoodsUtil(proc.user.GetDBID())
	infoUtil := db.PlayerInfoUtil(proc.user.GetDBID())

	if goodsUtil.IsOwnGoods(id) {
		factor1, validTime := common.GetExpCardArgs(id)
		factor2, _ := infoUtil.GetUsingExpCardInfo()

		if math.Abs(float64(factor2)) < 0.1 || math.Abs(float64(factor2-factor1)) < 0.1 {
			proc.user.storeMgr.ReduceGoods(id, 1, common.RS_NormalUse)
			infoUtil.AddExpCardUseRecord(id, factor1, validTime)
			ret = 1
		}
	}

	factor, leftTime := infoUtil.GetUsingExpCardInfo()
	proc.user.RPC(iserver.ServerTypeClient, "SelectExpCardRet", ret, factor, leftTime)
}

// RPC_SelectMulExpCardReq 请求选用多张经验卡
func (proc *LobbyUserMsgProc) RPC_SelectMulExpCardReq(id, num uint32) {
	ret := uint8(0)
	goodsUtil := db.PlayerGoodsUtil(proc.user.GetDBID())
	infoUtil := db.PlayerInfoUtil(proc.user.GetDBID())

	var day uint32
	for i := 0; i < int(num); i++ {
		if goodsUtil.IsOwnGoods(id) {
			factor1, validTime := common.GetExpCardArgs(id)
			factor2, _ := infoUtil.GetUsingExpCardInfo()

			if math.Abs(float64(factor2)) < 0.1 || math.Abs(float64(factor2-factor1)) < 0.1 {
				proc.user.storeMgr.ReduceGoods(id, 1, common.RS_NormalUse)
				infoUtil.AddExpCardUseRecord(id, factor1, validTime)
				ret = 1
			}

			day = uint32(validTime / 24 / 60 / 60)
		}
	}

	factor, leftTime := infoUtil.GetUsingExpCardInfo()
	proc.user.RPC(iserver.ServerTypeClient, "SelectMulExpCardRet", ret, factor, leftTime, day, num)
}

// RPC_UpdateTaskItems 玩家在Room服对局结束后更新任务项完成进度
func (proc *LobbyUserMsgProc) RPC_UpdateTaskItems(data []byte, comrade bool, comrades uint32) {
	items := common.BytesToUint32s(data)
	proc.user.Debugf("UpdateTaskItems %+v", items)
	proc.user.updateDayTaskItems(items...)

	if comrade {
		proc.user.updateComradeTaskItems(comrades, items...)
	}

	proc.user.taskMgr.updateTaskItemsAll(items...)
}

// RPC_DrawDayTaskAwardsReq 玩家请求领取每日任务相关的奖励
func (proc *LobbyUserMsgProc) RPC_DrawDayTaskAwardsReq(typ uint8, id uint32) {
	proc.user.drawDayTaskAwards(typ, id)
}

// RPC_ComradeTaskReq 玩家请求战友任务的标记信息
func (proc *LobbyUserMsgProc) RPC_ComradeTaskReq() {
	proc.user.syncComradeTask()
}

// RPC_DrawComradeTaskAwardsReq 玩家请求领取战友任务相关的奖励
func (proc *LobbyUserMsgProc) RPC_DrawComradeTaskAwardsReq(id uint32) {
	proc.user.drawComradeTaskAwards(id)
}

// RPC_EnableTaskItemReq 玩家请求激活任务项
func (proc *LobbyUserMsgProc) RPC_EnableTaskItemReq(name string, taskId uint32) {
	proc.user.taskMgr.enableTaskItemByType(name, taskId)
}

// RPC_ReplaceTaskItemAwardsReq 玩家请求补领任务项奖励
func (proc *LobbyUserMsgProc) RPC_ReplaceTaskItemAwardsReq(name string, groupId uint32, taskId uint32) {
	proc.user.taskMgr.replaceTaskItemAwardsByType(name, uint8(groupId), taskId)
}

// RPC_DrawTaskItemAwardsReq 玩家请求领取任务项奖励
func (proc *LobbyUserMsgProc) RPC_DrawTaskItemAwardsReq(name string, groupId uint32, taskId uint32) {
	proc.user.taskMgr.drawTaskItemAwardsByType(name, uint8(groupId), taskId)
}

// RPC_DrawTaskAwardsReq 玩家请求领取任务奖励
func (proc *LobbyUserMsgProc) RPC_DrawTaskAwardsReq(name string, id uint32, typ uint8) {
	proc.user.taskMgr.drawTaskAwardsByType(name, id, typ)
}

// RPC_OldBringNewReq 玩家请求以老带新标记信息
func (proc *LobbyUserMsgProc) RPC_OldBringNewReq() {
	proc.user.syncOldBringNew()
}

// RPC_OldBringNewDetailReq 玩家请求以老带新详细信息
func (proc *LobbyUserMsgProc) RPC_OldBringNewDetailReq(typ uint8) {
	proc.user.syncOldBringNewDetail(typ)
}

// RPC_DrawOldBringNewAwardsReq 玩家请求领取以老带新相关奖励
func (proc *LobbyUserMsgProc) RPC_DrawOldBringNewAwardsReq(typ uint8, id uint32) {
	proc.user.drawOldBringNewAwards(typ, id)
}

// RPC_GetBindListReq 玩家请求推荐绑定列表
func (proc *LobbyUserMsgProc) RPC_GetBindListReq() {
	proc.user.friendMgr.recommendBindableFriends()
}

// RPC_TakeTeacherReq 玩家请求拜师
func (proc *LobbyUserMsgProc) RPC_TakeTeacherReq(uid uint64) {
	proc.user.takeTeacher(uid)

	if proc.user.friendMgr.isGameFriend(uid) {
		proc.user.friendMgr.syncFriendList()
		proc.user.friendMgr.SendProxyInfo(uid, "SyncFriendList")
	} else {
		proc.user.friendMgr.syncPlatFriendList()
		proc.user.friendMgr.SendProxyInfo(uid, "SyncPlatFriendList")
	}
}

// RPC_ReceivePupil 玩家收徒
func (proc *LobbyUserMsgProc) RPC_ReceivePupil() {
	proc.user.receivePupil()
}

// RPC_BattleBookingReq 玩家请求预约好友一起比赛
func (proc *LobbyUserMsgProc) RPC_BattleBookingReq(uid uint64) {
	ret := uint8(1)
	fmgr := proc.user.friendMgr

	if fmgr.isFriend(uid) && fmgr.getFriendState(uid) == common.StateGame && !proc.user.isTmpBookedFriend(uid) {
		fmgr.sendMsgToFriend(uid, iserver.ServerTypeClient, "BattleBookingNotify", proc.user.GetDBID())

		db.PlayerTempUtil(proc.user.GetDBID()).AddTmpBookedFriend(uid)
		db.PlayerTempUtil(uid).AddTmpBookingFriend(proc.user.GetDBID())

		ret = 0
	}

	proc.user.RPC(iserver.ServerTypeClient, "BattleBookingRet", ret, uid)
	proc.user.Info("BattleBookingRet, ret: ", ret, " uid: ", uid)
}

// RPC_BattleBookingRespReq 玩家响应好友的预约请求
func (proc *LobbyUserMsgProc) RPC_BattleBookingRespReq(resp uint8, uid uint64) {
	ret := uint8(1)
	fmgr := proc.user.friendMgr

	util1 := db.PlayerTempUtil(proc.user.GetDBID())
	util2 := db.PlayerTempUtil(uid)

	if fmgr.isFriend(uid) && proc.user.isTmpBookingFriend(uid) {
		fmgr.sendMsgToFriend(uid, iserver.ServerTypeClient, "BattleBookingRespNotify", resp, proc.user.GetDBID())
		ret = 0

		if resp == 1 {
			bookingFriend := util1.GetBookingFriend()
			if bookingFriend != 0 {
				proc.RPC_BattleBookingCancelReq(bookingFriend)
			}

			util1.SetBookingFriend(uid)
			util2.AddBookedFriend(proc.user.GetDBID())
		}
	}

	util1.DelTmpBookingFriend(uid)
	util2.DelTmpBookedFriend(proc.user.GetDBID())

	proc.user.RPC(iserver.ServerTypeClient, "BattleBookingRespRet", ret, resp, uid)
	proc.user.BattleBookingFlow(uid, proc.user.GetDBID(), uint32(resp))
	proc.user.Info("BattleBookingRespRet, ret: ", ret, " uid: ", uid)
}

// RPC_BattleBookingCancelReq 玩家取消好友的预约
func (proc *LobbyUserMsgProc) RPC_BattleBookingCancelReq(uid uint64) {
	fmgr := proc.user.friendMgr

	if fmgr.getFriendState(uid) != common.StateOffline {
		fmgr.sendMsgToFriend(uid, iserver.ServerTypeClient, "BattleBookingCancelNotify", proc.user.GetDBID())
	}

	proc.user.cancelBattleBooking(uid)
	proc.user.RPC(iserver.ServerTypeClient, "BattleBookingCancelRet", uint8(0), uid)
	proc.user.Info("BattleBookingCancelRet, ret: ", 0, " uid: ", uid)
}

// RPC_JoinBookingTeamReq 玩家请求加入之前预约的队伍
func (proc *LobbyUserMsgProc) RPC_JoinBookingTeamReq(uid uint64) {
	util1 := db.PlayerTempUtil(proc.user.GetDBID())
	util2 := db.PlayerTempUtil(uid)

	if util1.GetBookingFriend() != uid {
		return
	}

	ret := uint8(0)
	defer func() {
		proc.user.RPC(iserver.ServerTypeClient, "JoinBookingTeamRet", ret)
		proc.user.friendMgr.sendMsgToFriend(uid, iserver.ServerTypeClient, "JoinBookingTeamNotify", proc.user.GetDBID())
		proc.user.Info("JoinBookingTeamRet, ret: ", ret, " uid: ", uid)
	}()

	switch util2.GetGameState() {
	case common.StateOffline:
		ret = 1
	case common.StateMatching:
		ret = 2
	case common.StateGame:
		ret = 3
	case common.StateWatch:
		ret = 4
	case common.StateFree, common.StateMatchWaiting:
		{
			teamID := util2.GetPlayerTeamID()
			matchMode := util2.GetPlayerMatchMode()

			if isTeamFull(teamID) {
				ret = 5
			} else if !common.IsMatchModeOk(matchMode, 2) && !common.IsMatchModeOk(matchMode, 4) {
				ret = 6
			}
		}
	}

	if ret == 0 {
		proc.user.friendMgr.sendMsgToFriend(uid, common.ServerTypeLobby, "JoinBookingTeam", proc.user.GetDBID())
	} else {
		proc.user.cancelBattleBooking(uid)
	}
}

// RPC_JoinBookingTeam 预约好友请求加入队伍
func (proc *LobbyUserMsgProc) RPC_JoinBookingTeam(uid uint64) {
	if !proc.user.isBookedFriend(uid) {
		return
	}

	if proc.user.GetPlayerGameState() != common.StateFree &&
		proc.user.GetPlayerGameState() != common.StateMatchWaiting {
		return
	}

	teamID := proc.user.GetTeamID()
	if teamID != 0 {
		proc.user.friendMgr.sendMsgToFriend(uid, common.ServerTypeLobby, "InviteRsp", teamID)
		return
	}

	util := db.PlayerTempUtil(proc.user.GetDBID())
	util.AddReadyBookedFriend(uid)

	mapid := util.GetPlayerMapID()
	if mapid == 0 {
		mapid = 1
	}

	automatch := util.GetPlayerAutoMatch()

	if common.IsMatchModeOk(proc.user.matchMode, 2) {
		proc.RPC_QuickEnterTeam(mapid, 0, automatch)
	} else if common.IsMatchModeOk(proc.user.matchMode, 4) {
		proc.RPC_QuickEnterTeam(mapid, 1, automatch)
	}
}

// RPC_TeamCustomReq 玩家请求当队长或者找队伍
func (proc *LobbyUserMsgProc) RPC_TeamCustomReq(msg *protoMsg.TeamCustom) {
	var cdtime uint
	sv := common.GetTBSystemValue(common.System_TeamCustomCDTime)
	if sv > 5 {
		cdtime = sv - 5
	}

	if time.Since(proc.user.lastTeamCustomTime) < time.Duration(cdtime)*time.Second {
		return
	}

	if proc.user.GetPlayerGameState() == common.StateMatching {
		proc.user.AdviceNotify(common.NotifyCommon, 95) //匹配中无法发送招募信息
		return
	}

	teamID := proc.user.GetTeamID()
	if isCustomTeamFull(teamID, msg) {
		proc.user.AdviceNotify(common.NotifyCommon, 96) //队伍人数已达到招募要求
		return
	}

	proc.user.lastTeamCustomTime = time.Now()
	msg.Id = common.GetNewTeamCustomID()
	msg.Info = proc.user.getChaterInfo()

	msg.CurNum, _ = db.PlayerTeamUtil(teamID).GetTeamNum()
	if msg.CurNum == 0 {
		msg.CurNum = 1
	}

	db.PlayerTempUtil(proc.user.GetDBID()).AddTeamCustom(msg)
	GetSrvInst().FireEvent(iserver.RPCChannel, "TeamCustomNotify", msg)
}

// RPC_TeamCustomRespReq 玩家响应当队长或者找队伍通知
func (proc *LobbyUserMsgProc) RPC_TeamCustomRespReq(uid, id uint64) {
	ret := proc.user.doTeamCustomResp(uid, id)
	proc.user.RPC(iserver.ServerTypeClient, "TeamCustomRespRet", ret, uid, id)
	proc.user.Error("TeamCustomRespRet, ret: ", ret, " uid: ", uid, " id: ", id)
}

// RPC_ChatPushReq 玩家请求发布聊天内容
func (proc *LobbyUserMsgProc) RPC_ChatPushReq(msg *protoMsg.ChatDetail) {
	var cdtime uint
	sv := common.GetTBSystemValue(common.System_ChatPushCDTime)
	if sv > 5 {
		cdtime = sv - 5
	}

	if time.Since(proc.user.lastChatPushTime) < time.Duration(cdtime)*time.Second && !msg.Trumpet {
		return
	}

	var canTrumpet bool
	if msg.Trumpet {
		if db.PlayerGoodsUtil(proc.user.GetDBID()).IsOwnGoods(common.Item_ChatTrumpet) {
			proc.user.storeMgr.ReduceGoods(common.Item_ChatTrumpet, 1, common.RS_Chat)
			canTrumpet = true
		}
	}

	if !msg.Trumpet || canTrumpet {
		f := func(content string) {
			msg.Content = content
			msg.Info = proc.user.getChaterInfo()
			GetSrvInst().FireEvent(iserver.RPCChannel, "ChatPushNotify", msg)
		}

		uuid := common.GetChatUUID()
		GetSrvInst().chats.Store(uuid, f)

		loginMsg := proc.user.GetPlayerLogin()
		tsssdk.JudgeUserInputChat(loginMsg.GetVOpenID(), proc.user.GetName(), msg.GetContent(), uuid, loginMsg.GetPlatID(), proc.user.GetDBID(), proc.user.GetLevel())
	}

	if !msg.Trumpet {
		proc.user.lastChatPushTime = time.Now()
	}
}

// RPC_PrivateChatReq 玩家请求私聊好友
func (proc *LobbyUserMsgProc) RPC_PrivateChatReq(uid uint64, content string) {
	if time.Since(proc.user.lastPrivateChatTime) < 100*time.Millisecond {
		proc.user.RPC(iserver.ServerTypeClient, "PrivateChatRet", uint8(1), uid, content)
		return
	}

	proc.user.lastPrivateChatTime = time.Now()
	util := db.GetFriendUtil(uid)

	f := func(content string) {
		if !util.IsBlacker(proc.user.GetDBID()) {
			util.AddChat(&protoMsg.ChatInfo{
				Uid:     proc.user.GetDBID(),
				Stamp:   uint64(time.Now().Unix()),
				Content: content,
			})
			proc.user.friendMgr.sendMsgToFriend(uid, iserver.ServerTypeClient, "PrivateChatNotify", proc.user.GetDBID(), content)
		}
		proc.user.RPC(iserver.ServerTypeClient, "PrivateChatRet", uint8(0), uid, content)
	}

	uuid := common.GetChatUUID()
	GetSrvInst().chats.Store(uuid, f)

	loginMsg := proc.user.GetPlayerLogin()
	tsssdk.JudgeUserInputChat(loginMsg.GetVOpenID(), proc.user.GetName(), content, uuid, loginMsg.GetPlatID(), proc.user.GetDBID(), proc.user.GetLevel())
}

// RPC_PrivateChatReadReq 玩家请求查看私聊
func (proc *LobbyUserMsgProc) RPC_PrivateChatReadReq(uid uint64) {
	db.GetFriendUtil(proc.user.GetDBID()).DelChat(uid)
}

// RPC_PullBlackReq 玩家请求拉黑好友
func (proc *LobbyUserMsgProc) RPC_PullBlackReq(uid uint64) {
	ret := uint8(1)
	if db.GetFriendUtil(proc.user.GetDBID()).AddBlacker(db.FriendInfo{uid, "", time.Now().Unix()}) {
		ret = 0
	}
	proc.user.RPC(iserver.ServerTypeClient, "PullBlackRet", ret, uid)
}

// RPC_CancelBlackReq 玩家请求取消拉黑
func (proc *LobbyUserMsgProc) RPC_CancelBlackReq(uid uint64) {

	ret := uint8(1)
	if db.GetFriendUtil(proc.user.GetDBID()).DelBlacker(uid) {
		ret = 0
	}
	proc.user.RPC(iserver.ServerTypeClient, "CancelBlackRet", ret, uid)
}

// RPC_PlayerLogin 玩家登录消息
func (proc *LobbyUserMsgProc) RPC_PlayerLogin(msg *protoMsg.PlayerLogin) {
	if msg == nil {
		return
	}

	// 断线重连打印PlayerLogout
	if proc.user.loginMsg != nil {
		proc.user.tlogOnPlayerLogout()
		proc.user.loginTime = time.Now().Unix()
	}

	msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
	msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
	if msg.VGameAppID == "" {
		if msg.LoginChannel == 1 {
			msg.VGameAppID = "wxa916d09c4b4ef98f"
		} else if msg.LoginChannel == 2 {
			msg.VGameAppID = "1106393072"
		} else if msg.LoginChannel == 3 {
			msg.VGameAppID = "G_1106393072"
		}
	}
	proc.user.Info("Player login, channel: ", msg.LoginChannel, " token: ", proc.user.GetAccessToken(), " platid: ", msg.PlatID)

	r := strings.NewReplacer("|", "-", "｜", "+", ",", "=")

	msg.IZoneAreaID = 0
	if msg.VOpenID == "" {
		var err error
		msg.VOpenID, err = dbservice.Account(proc.user.GetDBID()).GetUsername()
		if err != nil {
			msg.VOpenID = "NoOpenID"
		}
		proc.user.Info("RPC_PlayerLogin VOpenID IS NULL")
	}

	msg.Level = proc.user.GetLevel()
	msg.PlayerFriendsNum = proc.user.GetFriendsNum()
	if msg.ClientVersion == "" {
		msg.ClientVersion = "1.0.0"
	}
	if msg.SystemHardware == "" {
		msg.SystemHardware = "NoSysHardware"
	} else {
		msg.SystemHardware = r.Replace(msg.SystemHardware)
	}

	if msg.TelecomOper == "" {
		msg.TelecomOper = "NoOp"
	} else {
		msg.TelecomOper = r.Replace(msg.TelecomOper)
	}

	if msg.Network == "" {
		msg.Network = "NoNetwork"
	} else {
		msg.Network = r.Replace(msg.Network)
	}

	msg.VRoleID = fmt.Sprintf("%d", proc.user.GetDBID())
	if msg.VRoleID == "" {
		msg.VRoleID = "NoRoleID"
	}

	msg.VRoleName = proc.user.GetName()
	if msg.VRoleName == "" {
		msg.VRoleName = "NoUserName"
	} else {
		msg.VRoleName = r.Replace(msg.VRoleName)
	}

	if msg.RegChannel == "" {
		msg.RegChannel = "NoRegChannel"
	} else {
		msg.RegChannel = r.Replace(msg.RegChannel)
	}

	proc.user.SetPlayerLogin(msg)
	proc.user.SetPlayerLoginDirty()
	proc.user.loginMsg = msg

	if proc.user.isReg {
		regMsg := &protoMsg.PlayerRegister{}
		regMsg.GameSvrID = msg.GameSvrID
		regMsg.DtEventTime = msg.DtEventTime
		regMsg.VGameAppID = msg.VGameAppID
		regMsg.PlatID = msg.PlatID
		regMsg.IZoneAreaID = msg.IZoneAreaID
		regMsg.VOpenID = msg.VOpenID
		regMsg.TelecomOper = msg.TelecomOper
		regMsg.RegChannel = msg.RegChannel
		regMsg.LoginChannel = msg.LoginChannel
		tlog.Format(regMsg)

		proc.user.onUserRegister()
	}

	if proc.user.loginMsg != nil && proc.user.loginMsg.LoginChannel == 2 {
		var lst []*msdk.Param
		if proc.user.isReg {
			//qqscorebatch 用户注册时间
			lst = append(lst, &msdk.Param{
				Tp:      25,
				BCover:  1,
				Data:    fmt.Sprintf("%v", proc.user.loginTime),
				Expires: "不过期",
			})
		}

		//qqscorebatch 最近登录时间
		lst = append(lst, &msdk.Param{
			Tp:      8,
			BCover:  1,
			Data:    fmt.Sprintf("%v", proc.user.loginTime),
			Expires: "不过期",
		})

		msdk.QQScoreBatchList(common.QQAppIDStr, common.MSDKKey, proc.user.loginMsg.VOpenID, proc.user.GetAccessToken(), proc.user.loginMsg.PlatID,
			proc.user.GetName(), proc.user.GetDBID(), lst)
	}

	tlog.Format(msg)

	if err := GetSrvInst().loginCnt(msg.VGameAppID, int(msg.PlatID)); err != nil {
		proc.user.Error(err, msg)
	}

	proc.user.tsssdkLogin()
	proc.user.tlogGoodRecordFlow(login)
}

func (proc *LobbyUserMsgProc) RPC_TssData(data []byte) {
	proc.user.tsssdkRecvData(data)
}

func (proc *LobbyUserMsgProc) RPC_CheckUserState() {
	state := db.PlayerTempUtil(proc.user.GetDBID()).GetGameState()
	proc.user.RPC(iserver.ServerTypeClient, "RetUserState", uint8(state))
}

// RPC_OpenChatFlow 聊天流水表(打开)
func (proc *LobbyUserMsgProc) RPC_OpenChatFlow() {
	proc.user.tlogChatFlow(0, 3, 1)
}

// RPC_CloseChatFlow 聊天流水表(关闭)
func (proc *LobbyUserMsgProc) RPC_CloseChatFlow() {
	proc.user.tlogChatFlow(1, 3, 1)
}

// RPC_OperFlow 操作模式表
func (proc *LobbyUserMsgProc) RPC_OperFlow(msg *protoMsg.OperFlow) {
	proc.user.tlogOperFlow(msg)
}

// RPC_SnsFlow 社交流水表
func (proc *LobbyUserMsgProc) RPC_SnsFlow(msg *protoMsg.SnsFlow) {
	proc.user.tlogSnsFlow(msg.Count, msg.SnsType, msg.AcceptOpenID)

	//玩家分享比赛战绩到朋友圈
	if msg.SnsType == 0 {
		proc.user.festivalDataFlush(common.Act_ShareNum)
	}
}

// RPC_ShareReport 客户端分享后，向服务器上报记录
func (proc *LobbyUserMsgProc) RPC_ShareReport(typ uint8) {
	switch typ {
	case 1, 2, 3, 4, 5:
		proc.user.updateDayTaskItems(common.TaskItem_ShareResult, 1)
		proc.user.taskMgr.updateTaskItemsAll(common.TaskItem_ShareResult, 1)
	}
}

// RPC_SyncFlatFriend 同步平台好友
func (proc *LobbyUserMsgProc) RPC_SyncFlatFriend(msg *protoMsg.PlatFriendStateReq) {
	if msg == nil {
		return
	}
	proc.user.friendMgr.InitPlatFriendList(msg)
}

// RPC_NotifyWaitingNums 通知客户端当前匹配人数
func (proc *LobbyUserMsgProc) RPC_NotifyWaitingNums(totalNum uint32) {
	if err := proc.user.RPC(iserver.ServerTypeClient, "NotifyWaitingNums", totalNum); err != nil {
		proc.user.Error("RPC NotifyWaitingNums err: ", err)
	}
}

// RPC_KickingPlayer 执行踢出角色下线请求
func (proc *LobbyUserMsgProc) RPC_KickingPlayer(banAccountReason string) {
	proc.user.RPC(iserver.ServerTypeClient, "KickingPlayerMsg", banAccountReason)

	proc.user.RPC(iserver.ServerTypeClient, "NoticeKicking")
	err := GetSrvInst().DestroyEntityAll(proc.user.GetID())
	if err != nil {
		proc.user.Error("KickingPlayer failed, DestroyEntityAll err: ", err)
	}
}

// RPC_BuyGoods 购买商品
func (proc *LobbyUserMsgProc) RPC_BuyGoods(goodsID uint32, num uint32, param []byte) {
	result := proc.user.storeMgr.BuyGoods(goodsID, num, param)
	proc.user.RPC(iserver.ServerTypeClient, "BuyGoodsResult", goodsID, result)
}

// RPC_UpdateGoodsState 更新购买物品状态
func (proc *LobbyUserMsgProc) RPC_UpdateGoodsState(goodsID uint32, state uint32) {
	proc.user.storeMgr.updateGoodsState(goodsID, state)
}

func (proc *LobbyUserMsgProc) RPC_AutoEnterTeam(teamID uint64) {
	if teamID == 0 {
		proc.user.RPC(iserver.ServerTypeClient, "AutoEnterTeam", teamID)
	} else {
		srvID, err := db.PlayerTeamUtil(teamID).GetMatchSrvID()
		if err != nil || srvID == 0 {
			proc.user.Error("AutoEnterTeam failed, GetMatchSrvID err: ", err)
			return
		}
		name := fmt.Sprintf("%s:%d", common.TeamMgr, srvID)
		proc.user.teamMgrProxy = entity.GetEntityProxy(name)
		if proc.user.teamMgrProxy == nil {
			proc.user.RPC(iserver.ServerTypeClient, "AutoEnterTeam", uint64(0))
			return
		}

		ranks := proc.user.GetRanksByMode(proc.user.matchMode, 2, 4)
		if ranks == nil {
			proc.user.Error("AutoEnterTeam failed, ranks is nil")
			return
		}

		color := common.GetPlayerNameColor(proc.user.GetDBID())
		weapon := proc.user.GetOutsideWeapon()

		if err := proc.user.teamMgrProxy.RPC(common.ServerTypeMatch, "AutoEnterTeam",
			GetSrvInst().GetSrvID(), proc.user.GetID(), teamID,
			proc.user.GetName(), proc.user.GetRoleModel(), proc.user.GetDBID(), proc.user.GetTeamID(), ranks[2], ranks[4], proc.user.GetVeteran(), color, weapon); err != nil {
			proc.user.Error("RPC AutoEnterTeam err: ", err)
		}
	}
}

// RPC_ShareRMBMoney 腾讯RMB分享
func (proc *LobbyUserMsgProc) RPC_ShareRMBMoney(shareRMBMoney *protoMsg.ShareRMBMoney) {
	proc.user.Debug("Share RMB Money, channel: ", shareRMBMoney.Channel)

	if shareRMBMoney.Channel == 1 {
		proc.user.WXShareRMB(shareRMBMoney)
	} else if shareRMBMoney.Channel == 2 {
		proc.user.QQShareRMB(shareRMBMoney)
	}
}

func (proc *LobbyUserMsgProc) RPC_SyncVoiceInfo(teamID uint64, memberId int32, entityId uint64) {
	if teamID == 0 {
		return
	}

	if proc.user.watchTarget != 0 {
		proc.user.RPC(common.ServerTypeRoom, "SyncVoiceInfo", teamID, memberId)
		return
	}

	srvID, err := db.PlayerTeamUtil(teamID).GetMatchSrvID()
	if err != nil {
		proc.user.Error("SyncVoiceInfo failed, GetMatchSrvID err: ", err)
		return
	}

	if proc.user.teamMgrProxy == nil {
		name := fmt.Sprintf("%s:%d", common.TeamMgr, srvID)
		proc.user.teamMgrProxy = entity.GetEntityProxy(name)
		if proc.user.teamMgrProxy == nil {
			proc.user.Error("SyncVoiceInfo failed, teamMgrProxy is nil")
			return
		}
	}

	if err := proc.user.teamMgrProxy.RPC(common.ServerTypeMatch, "SyncVoiceInfo", GetSrvInst().GetSrvID(), teamID, memberId, proc.user.GetID()); err != nil {
		proc.user.Error("RPC SyncVoiceInfo err: ", err)
	}
}

// RPC_HaveAchieved 玩家达成成就
func (proc *LobbyUserMsgProc) RPC_HaveAchieved(id uint32) {
	// 成就信息
	achieveConfig, ok := excel.GetAchievement(uint64(id))
	if !ok {
		proc.user.Warn("GetAchievement failed, id doesn't exist, id: ", id)
		return
	}

	// 添加邮件
	objs := make(map[uint32]uint32)
	objs[uint32(achieveConfig.BonusID)] = uint32(achieveConfig.BonusNum)
	sendObjMail(proc.user.GetDBID(), "", 0, achieveConfig.MailTitle, achieveConfig.MailContent, "", "", objs)
	proc.user.RPC(iserver.ServerTypeClient, "AddNewMail")
}

// RPC_AccessTokenChange 玩家的accessToken变化了
func (proc *LobbyUserMsgProc) RPC_AccessTokenChange(accessToken string) {
	proc.user.SetAccessToken(accessToken)
}

// RPC_TakeCheckinReward 领取签到活动奖励
func (proc *LobbyUserMsgProc) RPC_TakeCheckinReward(id uint32) {
	proc.user.Debug("RPC_TakeCheckinReward")
	proc.user.activityMgr.SetSignActivity(id)
}

// RPC_ReqCheckinActivity 同步已经领取到的天数
func (proc *LobbyUserMsgProc) RPC_ReqCheckinActivity() {
	proc.user.Debug("RPC_ReqCheckinActivity")
	proc.user.activityMgr.syncSignActivityID() // 同步已经领取到的id
}

func (proc *LobbyUserMsgProc) RPC_ChangeAccount() {
	//seelog.Debug("切换账号:", proc.user.GetName())

	proc.user.delMatch()
	proc.user.SetPlayerGameState(common.StateOffline)
	//proc.user.MemberQuitTeam()
}

// RPC_CheckPreGameState 检查上局游戏状态
func (proc *LobbyUserMsgProc) RPC_CheckPreGameState() {

	proc.user.Debug("RPC_CheckPreGameState proc.user.spaceID:", proc.user.spaceID)

	if proc.user.spaceID != 0 {
		// 客户端未直接调用Room的该RPC时，Lobby提供转发功能（测试）
		proc.user.RPC(common.ServerTypeRoom, "CheckPreGameState")
	} else {
		proc.user.RPC(iserver.ServerTypeClient, "NotifyContinuePreGame", uint64(0))
	}
}

// RPC_NotifyContinuePreGame Room->Lobby 这条消息需要由Gateway转发出去, 不能直接从Room转到Client, 处理顶号的情况
func (proc *LobbyUserMsgProc) RPC_NotifyContinuePreGame(state uint64) {
	proc.user.RPC(iserver.ServerTypeClient, "NotifyContinuePreGame", state)
}

// RPC_EnterRoomReq C->S 客户端重新进入上局游戏
func (proc *LobbyUserMsgProc) RPC_ContinuePreGame(yes string) {

	user := proc.user
	if yes == "1" {
		// 调用ReEnterSpace for RoomServer
		proc.user.Debug("重新进入上局游戏")
		proc.user.RPC(common.ServerTypeRoom, "ReEnterSpace")
		proc.user.RPC(iserver.ServerTypeClient, "RetContinuePreGame")
		proc.user.RPC(iserver.ServerTypeClient, "MatchSuccess", user.mapId, user.skyBox)
		proc.user.RPC(iserver.ServerTypeClient, "RecvMatchModeId", user.matchMode)
		proc.user.SetPlayerGameState(common.StateGame)
	} else {
		// 调用LeaveSpace for RoomServer
		proc.user.Debug("放弃进入上局游戏")
		proc.user.RPC(common.ServerTypeRoom, "leaveSpace")
	}
}

// RPC_OnDisconnect Gateway -> Lobby
func (proc *LobbyUserMsgProc) MsgProc_ConnEventNotify(imsg msgdef.IMsg) {
	msg := imsg.(*msgdef.ConnEventNotify)
	// proc.user.Debug("MsgProc_ConnEventNotify eventType:", msg.EvtType)
	if msg.EvtType == 1 {
		// 断开连接事件
		// proc.user.Debug("Gateway connection is disconnected")

		// matchstate := db.PlayerTempUtil(proc.user.GetDBID()).GetGameState()
		// if matchstate != common.StateMatchWaiting {
		// 	proc.user.Debug("RPC_OnDisconnect current matchstate:", matchstate)
		// 	return
		// }

		//客户端下线，取消检测
		if proc.user.cronTask != nil {
			proc.user.cronTask.Stop()
		}

		proc.user.delMatch()
		proc.user.SetPayOS("")

		// 玩家不在游戏中的时候掉线, 则切换状态至离线
		if proc.user.GetPlayerGameState() != common.StateGame {
			proc.user.SetPlayerGameState(common.StateOffline)
		}
	} else if msg.EvtType == 2 {
		// 重连事件
		proc.user.Debug("Gateway connection is reconnected")
		proc.user.tlogReConnetionFlow()

		proc.user.checkMatchOpen()
		proc.user.notifyServerMode()
		proc.user.notifyPlayerMode()

		proc.user.checkGlobalMail()
		checkMail(proc.user.GetDBID())
		proc.user.MailNotify()
		proc.user.checkModeLightFlag()

		friendMgr := proc.user.friendMgr
		if friendMgr != nil {
			friendMgr.syncFriendList()
			friendMgr.syncApplyList()
		} else {
			proc.user.Error("RPC_OnReconnnect friendMgr is nil")
		}

		//发送用于新手引导的数据
		noviceType, _ := db.PlayerInfoUtil(proc.user.GetDBID()).GetNoviceType()
		proc.user.RPC(iserver.ServerTypeClient, "NoviceDataNotify", noviceType)

		// 玩家重新上线,恢复在线状态
		if proc.user.GetPlayerGameState() == common.StateOffline {
			proc.user.SetPlayerGameState(common.StateFree)
		}

		proc.user.ChangeNameInfo()
		proc.user.AnnounceDataCenterAddr()          //通知客户端DataCenter服地址
		proc.user.CareerCurSeason()                 // 通知客户端赛季id
		proc.user.InsigniaInfo(proc.user.GetDBID()) //同步勋章信息

		proc.user.RPC(iserver.ServerTypeClient, "SrvTime", time.Now().UnixNano())
		activityMgr := proc.user.activityMgr
		if activityMgr != nil {
			activityMgr.syncActiveState()          //同步活动信息是否开放
			proc.user.activityMgr.syncActiveInfo() //同步活动领取状态信息
		} else {
			proc.user.Error("RPC_OnReconnnect activityMgr is nil")
		}

		if proc.user.treasureBoxMgr != nil {
			proc.user.treasureBoxMgr.syncTreasureBoxInfo() //同步宝箱开启状态
		} else {
			proc.user.Error("RPC_OnReconnnect treasureBoxMgr is nil")
		}

		proc.user.storeMgr.initPropInfo() //初始化购买物品信息

		if proc.user.GetPlayerGameState() != common.StateGame {
			proc.user.StartCrondForExpireCheck()
		}
		proc.user.initPreferenceInfo() // 初始化偏好信息

		proc.user.ExpAdjust()
		proc.user.RPC(iserver.ServerTypeClient, "SyncLevelAndExp", proc.user.GetLevel(), proc.user.GetExp())

		proc.user.dayTaskInfoNotify()
		proc.user.comradeTaskInfoNotify()
		proc.user.taskMgr.syncTaskDetailAll()

		proc.user.updateDayTaskItems(common.TaskItem_Login, 1)
		proc.user.taskMgr.updateTaskItemsAll(common.TaskItem_Login, 1)

		proc.user.festivalDataFlush(common.Act_Login)

		proc.user.syncWeaponEquipment()   // 同步武器装备信息
		proc.user.syncUnreadPrivateChat() // 同步私聊信息
		proc.user.syncBattleBooking()     // 同步预约关系

		proc.user.syncFirstPay() // 同步首充
		proc.user.syncPayLevel() // 同步首次充送

		proc.user.syncRedDotOnce() // 通知玩家一次行点击类红点

		if proc.user.teamInfo != nil {
			proc.user.RPC(iserver.ServerTypeClient, "SyncTeamInfoRet", proc.user.teamInfo)
		}

		//清除预约数据
		if proc.user.GetPlayerGameState() != common.StateGame {
			bookingFriend := db.PlayerTempUtil(proc.user.GetDBID()).GetBookingFriend()
			if bookingFriend != 0 {
				proc.RPC_BattleBookingCancelReq(bookingFriend)
			}
		}
	}
}

// RPC_StrangerList 前端获取当前关注的陌生人游戏状态
func (proc *LobbyUserMsgProc) RPC_StrangerList(list *protoMsg.StrangerList) {
	if len(list.List) > 20 {
		return
	}
	now := time.Now().Unix()
	for _, id := range list.List {
		state := db.PlayerTempUtil(id).GetGameState()

		tmpLevel, levelErr := dbservice.EntityUtil("Player", id).GetValue("Level")
		if levelErr != nil {
			proc.user.Error("RPC_StrangerList levelErr:", levelErr)
			continue
		}
		level, Err := redis.Uint64(tmpLevel, nil)
		if Err != nil {
			proc.user.Error("RPC_StrangerList Err:", Err)
			continue
		}

		proc.user.RPC(iserver.ServerTypeClient, "SyncStrangerState", id, uint32(state), uint32(now), uint32(level))
	}
}

// RPC_FriendRecommendList 最近一局游戏产生的陌生人推荐列表
func (proc *LobbyUserMsgProc) RPC_FriendRecommendList(uids ...uint64) {
	retMsg := &protoMsg.SyncFriendList{}
	util := db.GetFriendUtil(proc.user.GetDBID())
	for _, uid := range uids {
		if util.IsFriendByID(uid) {
			continue
		}

		retMsg.Item = append(retMsg.Item, proc.user.friendMgr.getFriendInfo(uid))
	}
	proc.user.RPC(iserver.ServerTypeClient, "FriendRecommendList", retMsg)
}

// RPC_PickVersionReward 更新活动物品领取
func (proc *LobbyUserMsgProc) RPC_PickVersionReward() {
	ret := proc.user.activityMgr.SetUpdateActivity() // 更新活动物品领取
	proc.user.Info("RPC_PickVersionReward:", ret)

	if err := proc.user.RPC(iserver.ServerTypeClient, "respVersionRewardState", ret); err != nil {
		proc.user.Error("err:", err, " ret:", ret)
	}
}

// RPC_PickFirstWin 领取首胜奖励
func (proc *LobbyUserMsgProc) RPC_PickFirstWin() {
	proc.user.Debug("RPC_PickFirstWin")
	proc.user.activityMgr.PickFirstWinActivity()
}

// RPC_ReportVersion 客户端上报版本号
func (proc *LobbyUserMsgProc) RPC_ReportVersion(version string, platID uint32) {
	proc.user.Info("RPC_ReportVersion version:", version, " platID:", platID)
	proc.user.version = version
	proc.user.platID = platID

	proc.user.activityMgr.syncActiveState() //同步活动信息是否开放
	proc.user.activityMgr.syncActiveInfo()  //同步活动领取状态信息

	if err := proc.user.RPC(iserver.ServerTypeClient, "ActivityFinish"); err != nil {
		proc.user.Error("err:", err)
	}
}

// RPC_TakeTDCheckinReward 3天签到活动
func (proc *LobbyUserMsgProc) RPC_TakeTDCheckinReward(id uint32) {
	proc.user.Debug("RPC_TakeTDCheckinReward")
	proc.user.activityMgr.SetThreeDayActivity(id)
}

// RPC_PickNewYearActivity 领取新年活动奖励
func (proc *LobbyUserMsgProc) RPC_PickNewYearActivity(id uint32) {
	proc.user.Debug("RPC_PickNewYearActivity:", id)

	proc.user.activityMgr.PickNewYearActivity(uint32(id)) //领取新年活动物品
}

// RPC_SetNewYearActivityInfo set新年活动info
func (proc *LobbyUserMsgProc) RPC_SetNewYearActivityInfo(content msgdef.IMsg) {
	msg := content.(*protoMsg.NewYearInfo)
	proc.user.Debug("RPC_SetNewYearActivityInfo:", msg)

	goodID := make(map[uint64]int32)
	for _, v := range msg.GoodID {
		goodID[uint64(v.Key)] = v.Num
	}
	proc.user.activityMgr.SetNewYearInfo(goodID)
	proc.user.activityMgr.updateActivityInfo(goodID)
}

// RPC_UpdateGoodsUse 更新购买物品状态
func (proc *LobbyUserMsgProc) RPC_UseGoods(goodsID uint32, state uint32) {
	if state == 1 {
		goodsConfig, ok := excel.GetStore(uint64(goodsID))
		if !ok {
			proc.user.Error(" goods doesn't exist, id: ", goodsID)
			return
		}

		switch goodsConfig.Type {
		case GoodsGameEquipPack, GoodsGameEquipHead:
			{
				itemData, ok := excel.GetItem(goodsConfig.RelationID)
				if !ok {
					return
				}
				proc.user.setPreferenceSwitch(uint32(goodsConfig.Type), uint32(itemData.Subtype), false)
			}
		case GoodsGiftPack:
			{
				if goodsConfig.UseType == 3 {
					proc.user.storeMgr.ReduceGoods(goodsID, 1, common.RS_NormalUse)
					proc.user.storeMgr.getGiftPack(goodsID, 1, common.RS_BuyGoods, common.MT_NO, 0)
				}
			}
		}
	}
	if goodsID == 5001 { //使用默认场景
		if state == 1 {
			proc.user.SetTheme(5001)
		} else {
			proc.user.SetTheme(0)
		}
		return
	}

	proc.user.storeMgr.useMayaItem(goodsID, state)
	proc.user.Info("UseGoods, goodsID: ", goodsID, " state: ", state)
}

// WeaponEquipReq 客户端请求装备武器
func (proc *LobbyUserMsgProc) RPC_WeaponEquipReq(weapon, addition uint32, action uint8) {
	proc.user.weaponEquip(weapon, addition, action)
}

// RPC_ReportQemu 客户端上报是否是模拟器  0不是模拟器， 1是模拟器
func (proc *LobbyUserMsgProc) RPC_ReportQemu(is uint32) {
	proc.user.Debug("RPC_ReportQemu:", is)

	proc.user.isQemu = is
}

// RPC_ReqOpenTreasureBox 客户端开宝箱
func (proc *LobbyUserMsgProc) RPC_ReqOpenTreasureBox(id uint32) {
	proc.user.Debug("RPC_ReqOpenTreasureBox:", id)

	ret, itemId, moneyType, money := proc.user.treasureBoxMgr.syncOpenTreasureBox(id)

	proc.user.Debug("syncOpenTreasureBox-ret:", ret, " itemId:", itemId, " moneyType:", moneyType, " money:", money)
	if err := proc.user.RPC(iserver.ServerTypeClient, "RspOpenTreasureBox", ret, itemId, moneyType, money); err != nil {
		proc.user.Error(err)
	}

	if ret == 1 {
		proc.user.updateDayTaskItems(common.TaskItem_OpenBox, 1)
		proc.user.taskMgr.updateTaskItemsAll(common.TaskItem_OpenBox, 1)
		proc.user.festivalDataFlush(common.Act_OpenBox)
		proc.user.AddAchievementData(common.AchievementOpenTreasureBox, 1)
	}
}

// RPC_SetInsigniaFlag 设置勋章是否有红点标记
func (proc *LobbyUserMsgProc) RPC_SetInsigniaFlag(id uint32) {
	proc.user.Debug("RPC_SetInsigniaFlag id:", id)

	util := db.PlayerInsigniaUtil(proc.user.GetDBID())
	util.SetInsigniaFlag(id, 0)                 // 设置新获得的勋章红点标记 1:新获得的 0:旧的
	proc.user.InsigniaInfo(proc.user.GetDBID()) //同步勋章信息
}

// 前端检测网络速度
func (proc *LobbyUserMsgProc) RPC_PingNetMsg() {
	proc.user.RPC(iserver.ServerTypeClient, "PingNetMsg")
}

// RPC_SetAchievementFlag 设置成就是否有高亮标记
func (proc *LobbyUserMsgProc) RPC_SetAchievementFlag(id uint32) {
	proc.user.Debug("RPC_SetAchievementFlag id:", id)

	util := db.PlayerAchievementUtil(proc.user.GetDBID())
	achieveProcess := util.GetAchieveInfo()
	oneAchieve := achieveProcess[uint64(id)]
	if oneAchieve == nil {
		proc.user.Warn("oneAchieve is nil!")
		return
	}
	oneAchieve.Flag = 0 // 1:代表是新获得的成就, 0:旧成就

	var needSave []*db.AchieveInfo
	needSave = append(needSave, oneAchieve)

	util.AddAchieve(needSave) // 设置成就是否有高亮标记 1:新获得的 0:旧的
}

// RPC_syncVeteranRecallInfo 同步老兵召回的平台好友信息列表
func (proc *LobbyUserMsgProc) RPC_SyncVeteranRecallInfo() {
	proc.user.activityMgr.syncVeteranRecall()
}

// RPC_RecallFriendRes 召回老兵请求
func (proc *LobbyUserMsgProc) RPC_RecallFriendRes(uid uint64) {
	var ret uint32 = 0 //失败

	if uid != 0 {
		if ok := proc.user.activityMgr.recallFriendRes(uid); ok {
			ret = 1 //成功

			veteranRecallList := &db.VeteranRecallList{
				List: make(map[uint64]uint32),
			}
			playerInfoUtil := db.PlayerInfoUtil(uid)
			if err := playerInfoUtil.GetVeteranRecallList(veteranRecallList); err != nil {
				proc.user.Error("GetVeteranRecallList err:", err)
				return
			}
			veteranRecallList.List[proc.user.GetDBID()] = 1
			if err := playerInfoUtil.SetVeteranRecallList(veteranRecallList); err != nil {
				proc.user.Error("SetVeteranRecallList err:", err)
				return
			}
		}
	}

	proc.user.Debug("RPC_RecallFriendRes ret:", ret, " uid:", uid)
	if err := proc.user.RPC(iserver.ServerTypeClient, "RecallFriendRsq", ret, uid); err != nil {
		proc.user.Error(err)
	}
}

// RPC_PickRecallRewardRes 领取成功召回奖励
func (proc *LobbyUserMsgProc) RPC_PickRecallRewardRes(id uint32) {
	state := proc.user.activityMgr.pickRecallRewardRes(id)

	proc.user.Debug("PickRecallRewardRsq id:", id, " state:", state)
	if err := proc.user.RPC(iserver.ServerTypeClient, "PickRecallRewardRsq", id, state); err != nil {
		proc.user.Error(err)
	}
}

// RPC_SyncBackBattleInfo 同步重返光荣战场领取到的id
func (proc *LobbyUserMsgProc) RPC_SyncBackBattleInfo() {
	proc.user.activityMgr.syncBackBattleInfo()
}

// RPC_PickBackBattleRewardRes 领取重返光荣战场奖励
func (proc *LobbyUserMsgProc) RPC_PickBackBattleRewardRes(id uint32) {
	state := proc.user.activityMgr.pickBackBattleReward(id)

	proc.user.Debug("PickBackBattleRewardRsq-id:", id, " state:", state)
	if err := proc.user.RPC(iserver.ServerTypeClient, "PickBackBattleRewardRsq", id, state); err != nil {
		proc.user.Error(err)
	}
}

// RPC_OpenRandomPreference 开启随机设置 (typ: 1角色，2伞包，5背包，6头盔)
func (proc *LobbyUserMsgProc) RPC_OpenRandomPreference(typ, level uint32) {
	proc.user.Debug("RPC_OpenRandomPreference typ:", typ, " level:", level)

	proc.user.setPreferenceSwitch(typ, level, true)

	if typ != 1 {
		proc.user.preItemRandomCreate(typ, level) // 刷新
	}
}

// RPC_SetAllOrPartPre 全体还是部分偏好设置开关(typ: 1角色，2伞包，5背包，6头盔) (result:0全体， 1偏好)
func (proc *LobbyUserMsgProc) RPC_AllOrPartPreSwitch(typ, level, result uint32) {
	ret, state := proc.user.allOrPartPreSwitch(typ, level, result)
	if ret {
		if typ != 1 {
			proc.user.preItemRandomCreate(typ, level)
		}

		proc.user.tlogPreferenceFlow(typ, result) // tlogPreferenceFlow 偏好流水表
	}

	proc.user.Debug("RPC_AllOrPartPreSwitch typ:", typ, " level:", level, " result:", result, " ret:", ret, " state:", state)
	if err := proc.user.RPC(iserver.ServerTypeClient, "AllOrPartPreSwitchRsp", ret, state, typ, level); err != nil {
		proc.user.Error(err)
	}
}

// RPC_SetItemPreference 设置某物品为偏好物品
func (proc *LobbyUserMsgProc) RPC_SetItemPreference(typ, id, preference uint32) {
	proc.user.Debug("RPC_SetItemPreference typ:", typ, " id:", id, " preference:", preference)

	// 是否已拥有商品
	if ok := db.PlayerGoodsUtil(proc.user.GetDBID()).IsOwnGoods(id); !ok {
		proc.user.Info("RPC_SetItemPreference IsOwnGoods nil!")
		return
	}

	if ok := proc.user.setItemPreference(typ, id, preference); !ok {
		return
	}
	proc.user.storeMgr.updateGoodsPreference(id, preference)

	if typ == 1 {
		return
	}

	switch typ {
	case 1, 2: //1:角色，2：伞包
		proc.user.preItemRandomCreate(typ, 1) // 刷新
	case 5, 6: //5：背包，6:头盔
		goodsConfig, ok := excel.GetStore(uint64(id))
		if !ok {
			proc.user.Error("GetStore does't exist, id: ", id)
			return
		}
		itemData, ok := excel.GetItem(goodsConfig.RelationID)
		if !ok {
			proc.user.Error("GetItem does't exist, goodsConfig.RelationID: ", goodsConfig.RelationID)
			return
		}

		proc.user.preItemRandomCreate(typ, uint32(itemData.Subtype)) // 刷新
	}
}

// RPC_RefreshPreference 角色偏好刷新请求
func (proc *LobbyUserMsgProc) RPC_RefreshPreference() {
	proc.user.Debug("RPC_RefreshPreference!")

	proc.user.preItemRandomCreate(1, 1)
}

// RPC_PickFestivalReward 领取节日活动奖励
func (proc *LobbyUserMsgProc) RPC_PickFestivalRewardRes(id uint32) {
	state := proc.user.activityMgr.pickFestivalReward(id)

	proc.user.Debug("RPC_PickFestivalReward id:", id, " state:", state)
	if err := proc.user.RPC(iserver.ServerTypeClient, "PickFestivalRewardRsq", id, state); err != nil {
		proc.user.Error(err)
	}
}

// RPC_ExchangeGoodsRes 兑换物品
func (proc *LobbyUserMsgProc) RPC_ExchangeGoodsRes(actId, exchangeId uint32) {
	state := proc.user.activityMgr.exchangeGoods(actId, exchangeId)

	proc.user.Debug("RPC_ExchangeGoodsRes actId:", actId, " exchangeId:", exchangeId, " state:", state)
	if err := proc.user.RPC(iserver.ServerTypeClient, "ExchangeGoodsRsq", actId, exchangeId, state); err != nil {
		proc.user.Error(err)
	}
}

// RPC_ExchangeCallRes 兑换活动提醒开关请求
func (proc *LobbyUserMsgProc) RPC_ExchangeCallRes(actId, exchangeId, state uint32) {
	ret, result := proc.user.activityMgr.exchangeCallState(actId, exchangeId, state)

	proc.user.Debug("RPC_ExchangeCallRes actId:", actId, " exchangeId:", exchangeId, " result:", result, " ret:", ret)
	if err := proc.user.RPC(iserver.ServerTypeClient, "ExchangeCallRsq", actId, exchangeId, ret, result); err != nil {
		proc.user.Error(err)
	}
}

// RPC_ActFirstRedDot 设置每次版本更新首次点击红点提示
func (proc *LobbyUserMsgProc) RPC_ActFirstRedDot(actId uint32) {
	proc.user.Debug("RPC_ActFirstRedDot actId:", actId)

	if proc.user.activityMgr.activityUtil[uint64(actId)] != nil {
		proc.user.activityMgr.activityUtil[uint64(actId)].SetRedDot(proc.user.version)
	}
}

// RPC_ClickBallStarRewardRes 抽一球成名活动奖励
func (proc *LobbyUserMsgProc) RPC_ClickBallStarRewardRes() {
	result, id, rewardType, rewardNum := proc.user.activityMgr.clickBallStarReward()

	proc.user.Debug("RPC_ClickBallStarRewardRes result:", result, " id:", id, " rewardType:", rewardType, " rewardNum:", rewardNum)
	if err := proc.user.RPC(iserver.ServerTypeClient, "ClickBallStarRewardRsq", result, id, rewardType, rewardNum); err != nil {
		proc.user.Error(err)
	}
}

// RPC_PickBallStarRewardRes 领取一球成名活动奖励
func (proc *LobbyUserMsgProc) RPC_PickBallStarRewardRes(id uint32) {
	result := proc.user.activityMgr.pickBallStarReward(id)

	proc.user.Debug("RPC_PickBallStarRewardRes result:", result, " id:", id)
	if err := proc.user.RPC(iserver.ServerTypeClient, "PickBallStarRewardRsq", result, id); err != nil {
		proc.user.Error(err)
	}
}

// RPC_SetRoleSkillInfo 设置角色技能
func (proc *LobbyUserMsgProc) RPC_SetRoleSkillInfo(roleID, modeID, skillType, skillID, position uint32) {
	// 针对模式是否开放目前做的特殊处理
	{
		openModeID := proc.user.getSkillOpenMode()
		if len(openModeID) == 0 {
			proc.user.Error("SkillSytem excel err!")
			return
		}
		modeID = openModeID[0]
	}

	ret := proc.user.setRoleSkillInfo(roleID, modeID, skillType, skillID, position)
	msg := proc.user.getOneRoleSkillInfo(roleID, modeID)

	proc.user.Debug("RPC_SetRoleSkillInfo roleID:", roleID, " modeID:", modeID, " skillType:", skillType, " skillID:", skillID, " position:", position, " ret:", ret, " msg:", msg)
	if err := proc.user.RPC(iserver.ServerTypeClient, "SetRoleSkillInfoRsq", ret, msg); err != nil {
		proc.user.Error(err)
	}
}

// RPC_SyncRoleSkill 同步角色技能信息
func (proc *LobbyUserMsgProc) RPC_SyncRoleSkillRes() {
	msg := proc.user.getAllRoleSkillInfo()

	//proc.user.Debug("RPC_SyncRoleSkillRes msg:", msg)
	if err := proc.user.RPC(iserver.ServerTypeClient, "SyncRoleSkillRsq", msg); err != nil {
		proc.user.Error(err)
	}
}

// RPC_NewBuyGood 新的购买商品Rpc
func (proc *LobbyUserMsgProc) RPC_NewBuyGood(channel, tracks, id, priceNum, roundType uint32, param []byte) {
	result := proc.user.storeMgr.NewBuyGoods(channel, tracks, id, priceNum, roundType, param)

	proc.user.Debug("RPC_NewBuyGood channel:", channel, " tracks:", tracks, " id:", id, " priceNum:", priceNum, " roundType:", roundType, " result:", result)
	proc.user.RPC(iserver.ServerTypeClient, "NewBuyGoodsResult", channel, tracks, id, result)
}

// RPC_SaleGoodRefreshTimeRes 同步特惠商品轮数
func (proc *LobbyUserMsgProc) RPC_SaleGoodRefreshTimeRes() {
	_, _, round7, round1 := proc.user.storeMgr.getSaleGoodRound()

	proc.user.Debug("RPC_SaleGoodRefreshTimeRes round7:", round7, " round1:", round1)
	proc.user.RPC(iserver.ServerTypeClient, "SaleGoodRefreshTimeReq", uint32(time.Now().Unix()), round7, round1)
}

// RPC_SetRedDotOnce 设置首次点击红点消失
func (proc *LobbyUserMsgProc) RPC_SetRedDotOnce(num uint32) {
	util := db.PlayerInfoUtil(proc.user.GetDBID())
	info := &db.RedDotOnce{}
	info.RedDot = make(map[uint32]uint32)

	if err := util.GetRedDotOnce(info); err != nil {
		proc.user.Error("RPC_SetRedDotOnce err:", err)
		return
	}

	info.RedDot[num] = 1
	if err := util.SetRedDotOnce(info); err != nil {
		proc.user.Error("RPC_SetRedDotOnce err:", err)
		return
	}

	proc.user.syncRedDotOnce()
}

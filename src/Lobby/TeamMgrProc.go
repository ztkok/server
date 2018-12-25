package main

import (
	"common"
	"db"
	"excel"
	"fmt"
	"protoMsg"
	"zeus/dbservice"
	"zeus/entity"
	"zeus/iserver"
	"zeus/serverMgr"
)

// RPC_QuickEnterTeam 快速进入队伍 c-->s
func (proc *LobbyUserMsgProc) RPC_QuickEnterTeam(mapid uint32, teamtype uint32, automatch uint32) {
	proc.user.Info("RPC_QuickEnterTeam, mapid: ", mapid, "	teamtype: ", teamtype, " automatch: ", automatch, " matchMode: ", proc.user.matchMode)
	if teamtype != TwoTeamType && teamtype != FourTeamType {
		proc.user.Warn("QuickEnterTeam failed, teamtype is not allowed, teamtype: ", teamtype)
		return
	}

	if teamtype == TwoTeamType && !common.IsMatchModeOk(proc.user.matchMode, 2) {
		proc.user.AdviceNotify(common.NotifyCommon, 24)
		proc.user.Warn("QuickEnterTeam failed, match mode is not open, matchMode: ", proc.user.matchMode, " matchTyp: ", 2)
		return
	}

	if teamtype == FourTeamType && !common.IsMatchModeOk(proc.user.matchMode, 4) {
		proc.user.Warn("QuickEnterTeam failed, match mode is not open, matchMode: ", proc.user.matchMode, " matchTyp: ", 4)
		proc.user.AdviceNotify(common.NotifyCommon, 25)
		return
	}

	if proc.user.teamMgrProxy == nil {
		srvInfo, err := serverMgr.GetServerMgr().GetServerByType(common.ServerTypeMatch)
		if err != nil {
			proc.user.Error("QuickEnterTeam failed, GetServerByType err: ", err)
			return
		}

		name := fmt.Sprintf("%s:%d", common.TeamMgr, srvInfo.ServerID)
		proc.user.teamMgrProxy = entity.GetEntityProxy(name)
		if proc.user.teamMgrProxy == nil {
			proc.user.Error("QuickEnterTeam failed, teamMgrProxy is nil ")
			return
		}
	}

	ranks := proc.user.GetRanksByMode(proc.user.matchMode, 2, 4)
	if ranks == nil {
		proc.user.Error("QuickEnterTeam failed, ranks is nil")
		return
	}

	//普通模式之外的其他模式使用的地图均从赛事模式配置表中读取
	if proc.user.matchMode == common.MatchModeNormal {
		if mapid != 1 && mapid != 2 {
			mapid = 1
		}
	} else {
		info := common.GetOpenModeInfo(proc.user.matchMode, uint8((teamtype+1)*2))
		if info == nil {
			return
		}

		mapid = info.MapId
	}

	_, ok := excel.GetMaps(uint64(mapid))
	if !ok {
		proc.user.Error("QuickEnterTeam failed, map doesn't exist, mapid: ", mapid)
		return
	}

	//勇者模式不支持自动匹配队友
	if proc.user.matchMode == common.MatchModeBrave {
		automatch = 0
	}

	//如果存在队伍 获取信息
	if teamID := proc.user.GetTeamID(); teamID != 0 {
		proc.user.teamMgrProxy.RPC(common.ServerTypeMatch, "GetTeamInfo", proc.user.GetID(), teamID)
		return
	}

	srvID := proc.user.teamMgrProxy.(*entity.EntityProxy).SrvID
	teamID := db.GetTeamGlobalID()
	teamUtil := db.PlayerTeamUtil(teamID)
	tempUtil := db.PlayerTempUtil(proc.user.GetDBID())

	if err := teamUtil.SetMatchSrvID(srvID); err != nil {
		proc.user.Error("QuickEnterTeam failed, SetMatchSrvID err: ", err)
		return
	}

	if err := teamUtil.SetMatchMode(proc.user.matchMode); err != nil {
		proc.user.Error("QuickEnterTeam failed, SetMatchMode err: ", err)
		return
	}

	if err := teamUtil.SetTeamType(uint8(teamtype)); err != nil {
		proc.user.Error("QuickEnterTeam failed, SetTeamType err: ", err)
		return
	}

	if err := teamUtil.SetTeamMap(mapid); err != nil {
		proc.user.Error("QuickEnterTeam failed, SetTeamMap err: ", err)
		return
	}

	if err := tempUtil.SetPlayerMapID(mapid); err != nil {
		proc.user.Error("QuickEnterTeam failed, SetPlayerMapID err: ", err)
		return
	}

	if err := tempUtil.SetPlayerAutoMatch(automatch); err != nil {
		proc.user.Error("QuickEnterTeam failed, SetPlayerAutoMatch err: ", err)
		return
	}

	color := common.GetPlayerNameColor(proc.user.GetDBID())
	weapon := proc.user.GetOutsideWeapon()

	if err := proc.user.teamMgrProxy.RPC(common.ServerTypeMatch, "EnterTeam",
		GetSrvInst().GetSrvID(), proc.user.GetID(), proc.user.GetTeamID(), teamID, mapid,
		uint32(0), uint32(ranks[2]), uint32(ranks[4]), uint32(1), proc.user.GetName(), proc.user.GetRoleModel(), proc.user.GetDBID(), teamtype, automatch, proc.user.matchMode, proc.user.GetVeteran(), color, weapon); err != nil {
		proc.user.Error("RPC EnterTeam err: ", err)
		return
	}

	proc.user.Info("Send enter team req to match success, teamID: ", teamID, " srvID: ", srvID, " automatch:", automatch, " matchMode: ", proc.user.matchMode)
}

func (proc *LobbyUserMsgProc) RPC_EnterTeamRet(ret uint32, teamID uint64) {
	if ret == 0 {
		// 设置玩家所在队伍id
		proc.user.SetPlayerGameState(common.StateMatchWaiting)
		proc.user.syncDataToMatch()
		proc.user.pullBookedFriendsToTeam()
		proc.user.Info("Enter team success, teamID: ", teamID, " matchMode: ", proc.user.matchMode)
	} else {
		proc.user.AdviceNotify(common.NotifyError, 1)
		proc.user.Info("Enter team failed, ret: ", ret)
	}
}

func (proc *LobbyUserMsgProc) RPC_SyncTeamInfoRet(msg *protoMsg.SyncTeamInfoRet) {
	if proc.user.GetPlayerGameState() != common.StateGame && proc.user.GetPlayerGameState() != common.StateOffline {
		if msg.TeamState == common.PlayerTeamMatching {
			proc.user.SetPlayerGameState(common.StateMatching)
		} else {
			proc.user.SetPlayerGameState(common.StateMatchWaiting)

			util := db.PlayerTempUtil(proc.user.GetDBID())
			toJoinTeam := util.GetToJoinTeam()
			toInviteUser := util.GetToInviteUser()

			if toJoinTeam != 0 {
				util.SetToJoinTeam(0)
				proc.user.pullUserToTeam(proc.user.GetDBID(), toJoinTeam)
			}

			if toInviteUser != 0 {
				util.SetToInviteUser(0)
				proc.user.pullUserToTeam(toInviteUser, msg.Id)
			}
		}
	}

	if proc.user.GetID() == msg.Leaderid {
		proc.user.isInTeamReady = false
	}

	proc.user.SetTeamID(msg.Id)
	proc.user.teamInfo = msg
	proc.user.RPC(iserver.ServerTypeClient, "SyncTeamInfoRet", msg)
	proc.user.Debugf("SyncTeamInfoRet: %+v\n", msg)

	util := db.PlayerTeamUtil(msg.Id)
	num, _ := util.GetTeamNum()
	if num != uint32(len(msg.Memberinfo)) {
		util.SetTeamNum(uint32(len(msg.Memberinfo)))
	}
}

func (proc *LobbyUserMsgProc) RPC_ChangeTeamType(teamtype uint32) {
	if proc.user == nil {
		return
	}
	teamId := proc.user.GetTeamID()
	if teamId == 0 {
		return
	}
	if teamtype != TwoTeamType && teamtype != FourTeamType {
		return
	}

	if teamtype == TwoTeamType && !common.IsMatchModeOk(proc.user.matchMode, 2) {
		proc.user.AdviceNotify(common.NotifyCommon, 24)
		proc.user.Warn("ChangeTeamType failed, match mode is not open, matchMode: ", proc.user.matchMode, " matchTyp: ", 2)
		return
	}

	if teamtype == FourTeamType && !common.IsMatchModeOk(proc.user.matchMode, 4) {
		proc.user.Warn("ChangeTeamType failed, match mode is not open, matchMode: ", proc.user.matchMode, " matchTyp: ", 4)
		proc.user.AdviceNotify(common.NotifyCommon, 25)
		return
	}

	if proc.user.teamMgrProxy == nil {
		proc.user.Error("ChangeTeamType failed, teamMgrProxy is nil")
		return
	}

	if err := proc.user.teamMgrProxy.RPC(common.ServerTypeMatch, "ChangeTeamType",
		GetSrvInst().GetSrvID(), proc.user.GetID(), teamId, teamtype); err != nil {
		proc.user.Error("RPC ChangeTeamType err: ", err)
		return
	}

	if err := db.PlayerTeamUtil(teamId).SetTeamType(uint8(teamtype)); err != nil {
		proc.user.Error("SetTeamType err: ", err)
	}

	proc.user.Info("Send change team type req to match success, matchmode: ", proc.user.matchMode)
}

func (proc *LobbyUserMsgProc) RPC_ChangeAutoMatch(automatch uint32) {
	if proc.user == nil || proc.user.GetTeamID() == 0 {
		return
	}

	if proc.user.teamMgrProxy == nil {
		proc.user.Error("ChangeAutoMatch failed, teamMgrProxy is nil")
		return
	}

	if proc.user.matchMode == common.MatchModeBrave {
		proc.user.Error("ChangeAutoMatch failed, automatch is not allowed")
		return
	}

	if err := proc.user.teamMgrProxy.RPC(common.ServerTypeMatch, "ChangeAutoMatch",
		GetSrvInst().GetSrvID(), proc.user.GetID(), proc.user.GetTeamID(), automatch); err != nil {
		proc.user.Error("RPC ChangeAutoMatch err: ", err)
		return
	}

	if err := db.PlayerTempUtil(proc.user.GetDBID()).SetPlayerAutoMatch(automatch); err != nil {
		proc.user.Error("ChangeAutoMatch failed, SetPlayerAutoMatch err: ", err)
		return
	}

	proc.user.Info("Send change auto match req to match success, matchmode: ", proc.user.matchMode)
}

func (proc *LobbyUserMsgProc) RPC_ChangeMap(mapid uint32) {
	_, ok := excel.GetMaps(uint64(mapid))
	if !ok {
		proc.user.Error("ChangeMap failed, map doesn't exist, mapid: ", mapid)
		return
	}

	if err := db.PlayerTempUtil(proc.user.GetDBID()).SetPlayerMapID(mapid); err != nil {
		proc.user.Error("SetPlayerMapID err: ", err)
		return
	}

	teamId := proc.user.GetTeamID()
	if teamId == 0 {
		return
	}

	if proc.user.teamMgrProxy == nil {
		proc.user.Error("ChangeMap failed, teamMgrProxy is nil")
		return
	}

	if err := proc.user.teamMgrProxy.RPC(common.ServerTypeMatch, "ChangeMap",
		GetSrvInst().GetSrvID(), proc.user.GetID(), teamId, mapid); err != nil {
		proc.user.Error("RPC ChangeMap err: ", err)
		return
	}

	if err := db.PlayerTeamUtil(teamId).SetTeamMap(mapid); err != nil {
		proc.user.Error("SetTeamMap err: ", err)
		return
	}

	proc.user.Info("Change Map success, teamID: ", teamId, " mapid: ", mapid)
}

// RPC_QuitTeam 退出队伍 c-->s
func (proc *LobbyUserMsgProc) RPC_QuitTeam() {
	if proc.user == nil {
		return
	}
	teamId := proc.user.GetTeamID()
	if teamId == 0 {
		return
	}
	if proc.user.teamMgrProxy == nil {
		proc.user.Error("Quit team failed, teamMgrProxy is nil")
		return
	}

	if err := proc.user.teamMgrProxy.RPC(common.ServerTypeMatch, "LeaveTeam",
		GetSrvInst().GetSrvID(), proc.user.GetID(), teamId, true); err != nil {
		proc.user.Error("RPC LeaveTeam err: ", err)
		return
	}

	proc.user.Info("Send quit team req to match success, teamID: ", teamId, " matchMode: ", proc.user.matchMode)
}

// RPC_QuitTeamRet 退出队伍结果 s-->c
func (proc *LobbyUserMsgProc) RPC_QuitTeamRet(result uint32) {
	if result == 0 {
		proc.user.RPC(iserver.ServerTypeClient, "QuitTeamRet", uint64(result))

		proc.user.Info("Quit team success, teamId: ", proc.user.GetTeamID(), " matchMode: ", proc.user.matchMode)
		proc.user.SetTeamID(0)
		proc.user.teamInfo = nil

		proc.user.isInTeamReady = false
		if proc.user.GetPlayerGameState() != common.StateOffline {
			proc.user.SetPlayerGameState(common.StateFree)
		}

		proc.user.tlogMatchFlow(1, 1, 0, 2, 0, 0, 0)
	} else {
		proc.user.Error("Quit team failed, teamId: ", proc.user.GetTeamID(), " matchMode: ", proc.user.matchMode)
	}
}

// RPC_ConfirmTeamMatch 确认开始匹配 c-->s
func (proc *LobbyUserMsgProc) RPC_ConfirmTeamMatch() {
	if proc.user.GetPlayerGameState() != common.StateMatchWaiting {
		proc.user.Warn("ConfirmTeamMatch but user game state not free ", proc.user.GetPlayerGameState())
		return
	}
	if proc.user == nil || proc.user.GetTeamID() == 0 || proc.user.watchTarget != 0 {
		return
	}

	if proc.user.teamMgrProxy == nil {
		proc.user.Error("ConfirmTeamMatch failed, teamMgrProxy is nil")
		return
	}

	teamtype, err := db.PlayerTeamUtil(proc.user.GetTeamID()).GetTeamType()
	if err != nil {
		proc.user.Error("ConfirmTeamMatch failed, GetTeamType err: ", err)
		return
	}

	if teamtype == TwoTeamType && !common.IsMatchModeOk(proc.user.matchMode, 2) {
		proc.user.AdviceNotify(common.NotifyCommon, 24)
		proc.user.Warn("ConfirmTeamMatch failed, match mode is not open, matchMode: ", proc.user.matchMode, " matchTyp: ", 2)
		return
	}

	if teamtype == FourTeamType && !common.IsMatchModeOk(proc.user.matchMode, 4) {
		proc.user.Warn("ConfirmTeamMatch failed, match mode is not open, matchMode: ", proc.user.matchMode, " matchTyp: ", 4)
		proc.user.AdviceNotify(common.NotifyCommon, 25)
		return
	}

	//勇者战场
	if proc.user.matchMode == common.MatchModeBrave && !proc.user.CanMatchBraveGame((teamtype+1)*2) {
		return
	}

	if err := proc.user.teamMgrProxy.RPC(common.ServerTypeMatch, "ConfirmTeamMatch",
		GetSrvInst().GetSrvID(), proc.user.GetID(), proc.user.GetTeamID()); err != nil {
		proc.user.Error("RPC ConfirmTeamMatch err: ", err)
		return
	}
	proc.user.isInTeamReady = true
	proc.user.Info("Confirm team match, teamID: ", proc.user.GetTeamID(), " matchMode: ", proc.user.matchMode)
}

func (proc *LobbyUserMsgProc) RPC_InviteReq(otherid uint64) {
	if proc.user.GetDBID() == otherid {
		return
	}

	state1 := proc.user.GetPlayerGameState()
	if state1 != common.StateFree && state1 != common.StateMatchWaiting {
		proc.user.AdviceNotify(common.NotifyCommon, 99)
		return
	}

	state2 := proc.user.friendMgr.getFriendState(otherid)
	if state2 != common.StateFree && state2 != common.StateMatchWaiting {
		proc.user.AdviceNotify(common.NotifyCommon, 97)
		return
	}

	teamId := proc.user.GetTeamID()
	if teamId == 0 {
		return
	}

	entityID, err := dbservice.SessionUtil(otherid).GetUserEntityID()
	if err != nil {
		return
	}

	srvID, spaceID, err := dbservice.EntitySrvUtil(entityID).GetSrvInfo(iserver.ServerTypeGateway)
	if err != nil {
		return
	}
	proxy := entity.NewEntityProxy(srvID, spaceID, entityID)
	proxy.RPC(iserver.ServerTypeClient, "Invite", proc.user.GetID(), teamId, proc.user.GetName())
	proxy.RPC(iserver.ServerTypeClient, "InviteType", proc.user.GetID(), teamId, proc.user.GetName(), proc.user.matchMode)

	acceptOpenID, _ := dbservice.Account(otherid).GetUsername()
	proc.user.tlogSnsFlow(0, SNSTYPE_INVITE, acceptOpenID)

	proc.user.Info("Send invite req to friend ", otherid, ", teamId: ", teamId, " matchMode: ", proc.user.matchMode)
}

func (proc *LobbyUserMsgProc) RPC_InviteRsp(teamID uint64) {
	if proc.user.GetPlayerGameState() == common.StateMatching {
		return
	}

	bookingFriend := db.PlayerTempUtil(proc.user.GetDBID()).GetBookingFriend()
	if bookingFriend != 0 {
		proc.user.cancelBattleBooking(bookingFriend)
	}

	if proc.user.GetTeamID() == teamID {
		proc.user.Error("InviteRsp failed, already in this team")
		return
	}

	srvID, err := db.PlayerTeamUtil(teamID).GetMatchSrvID()
	if err != nil || srvID == 0 {
		proc.user.AdviceNotify(common.NotifyCommon, common.ErrCodeIDInviteJoin)
		proc.user.Errorf("InviteRsp failed teamID:%d srvID:%d err:%s", teamID, srvID, err)
		return
	}

	name := fmt.Sprintf("%s:%d", common.TeamMgr, srvID)
	proc.user.teamMgrProxy = entity.GetEntityProxy(name)
	if proc.user.teamMgrProxy == nil {
		proc.user.Error("InviteRsp failed, teamID: ", teamID)
		return
	}

	matchMode, _ := db.PlayerTeamUtil(teamID).GetMatchMode()
	ranks := proc.user.GetRanksByMode(matchMode, 2, 4)
	if ranks == nil {
		proc.user.Error("InviteRsp failed, ranks is nil")
		return
	}

	color := common.GetPlayerNameColor(proc.user.GetDBID())
	weapon := proc.user.GetOutsideWeapon()

	if err := proc.user.teamMgrProxy.RPC(common.ServerTypeMatch, "InviteRsp",
		GetSrvInst().GetSrvID(), proc.user.GetID(), teamID,
		proc.user.GetName(), proc.user.GetRoleModel(), proc.user.GetDBID(), proc.user.GetTeamID(), ranks[2], ranks[4], proc.user.GetVeteran(), color, weapon); err != nil {
		proc.user.Error("RPC InviteRsp err: ", err)
		return
	}

	proc.user.Info("Send invite resp to match success, teamID: ", teamID, " srvID：", srvID, " matchMode: ", proc.user.matchMode)
}

func (proc *LobbyUserMsgProc) RPC_InviteRspRet(result uint32, newteam, oldteam uint64) {
	if result != 0 {
		if result == 1 { //队伍失效
			proc.user.AdviceNotify(common.NotifyCommon, common.ErrCodeIDInviteJoin)
		}
		if result == 2 { // 队伍人数已满
			proc.user.AdviceNotify(common.NotifyCommon, 84)
		}
	} else {
		if oldteam != 0 {
			oldSrvID, err := db.PlayerTeamUtil(oldteam).GetMatchSrvID()
			if err == nil {
				oldname := fmt.Sprintf("%s:%d", common.TeamMgr, oldSrvID)
				proxy := entity.GetEntityProxy(oldname)
				if proxy != nil {
					proxy.RPC(common.ServerTypeMatch, "InviteLeaveTeam",
						GetSrvInst().GetSrvID(), proc.user.GetID(), oldteam)
				}
			}
		}

		proc.user.SetPlayerGameState(common.StateMatchWaiting)
		proc.user.syncDataToMatch()

		//接受邀请成功入队后，切换MatchMode和MapID
		teamUtil := db.PlayerTeamUtil(newteam)
		tempUtil := db.PlayerTempUtil(proc.user.GetDBID())

		proc.user.matchMode, _ = teamUtil.GetMatchMode()
		tempUtil.SetPlayerMatchMode(proc.user.matchMode)
		proc.user.notifyPlayerMode()

		mapid, _ := teamUtil.GetTeamMap()
		tempUtil.SetPlayerMapID(mapid)

		acceptOpenID, _ := dbservice.Account(proc.user.GetDBID()).GetUsername()
		proc.user.tlogSnsFlow(0, SNSTYPE_INVITESUCCESS, acceptOpenID)
		proc.user.Info("Invite resp success, oldteam: ", oldteam, " newteam: ", newteam, " matchMode: ", proc.user.matchMode)
	}
}

func (proc *LobbyUserMsgProc) RPC_AutoEnterTeamRet(result uint32) {
	if result != 0 {
		proc.user.RPC(iserver.ServerTypeClient, "AutoEnterTeam", uint64(0))
	} else {
		proc.user.SetPlayerGameState(common.StateMatchWaiting)
		proc.user.syncDataToMatch()
	}
}

func (proc *LobbyUserMsgProc) RPC_CancelTeamReady() {
	if proc.user.GetPlayerGameState() != common.StateMatchWaiting {
		proc.user.Warn("ConfirmTeamMatch but user game state not free ", proc.user.GetPlayerGameState())
		return
	}

	if proc.user == nil || proc.user.GetTeamID() == 0 {
		return
	}

	if proc.user.teamMgrProxy == nil {
		proc.user.Error("ConfirmTeamMatch failed, teamMgrProxy is nil")
		return
	}

	teamtype, err := db.PlayerTeamUtil(proc.user.GetTeamID()).GetTeamType()
	if err != nil {
		proc.user.Error("ConfirmTeamMatch failed, GetTeamType err: ", err)
		return
	}

	if teamtype == TwoTeamType && !common.IsMatchModeOk(proc.user.matchMode, 2) {
		proc.user.AdviceNotify(common.NotifyCommon, 24)
		proc.user.Warn("ConfirmTeamMatch failed, match mode is not open, matchMode: ", proc.user.matchMode, " matchTyp: ", 2)
		return
	}

	if teamtype == FourTeamType && !common.IsMatchModeOk(proc.user.matchMode, 4) {
		proc.user.Warn("ConfirmTeamMatch failed, match mode is not open, matchMode: ", proc.user.matchMode, " matchTyp: ", 4)
		proc.user.AdviceNotify(common.NotifyCommon, 25)
		return
	}

	if err := proc.user.teamMgrProxy.RPC(common.ServerTypeMatch, "CancelTeamReady",
		proc.user.GetID(), proc.user.GetTeamID()); err != nil {
		proc.user.Error("RPC ConfirmTeamMatch err: ", err)
		return
	}
	proc.user.isInTeamReady = false

	proc.user.Info("Confirm team match, teamID: ", proc.user.GetTeamID(), " matchMode: ", proc.user.matchMode)
}

// RPC_SyncTeamID
func (proc *LobbyUserMsgProc) RPC_SyncTeamID(teamID uint64, state uint8) {
	//TODO 队伍保留 状态需要重新整理
	if state == 0 {
		proc.user.isInTeamReady = false
	}
	proc.user.SetTeamID(teamID)
}

func (proc *LobbyUserMsgProc) RPC_GetExpectTime() {
	teamID := proc.user.GetTeamID()
	if teamID == 0 {
		return
	}
	if proc.user.teamMgrProxy == nil {
		return
	}
	proc.user.teamMgrProxy.RPC(common.ServerTypeMatch, "GetExpectTime", proc.user.GetID(), teamID)
}

// RPC_KickTeamMemberReq 踢出队员请求
func (proc *LobbyUserMsgProc) RPC_KickTeamMemberReq(memID uint64) {
	if proc.user.GetTeamID() == 0 || proc.user.teamInfo == nil ||
		proc.user.GetID() != proc.user.teamInfo.GetLeaderid() {
		return
	}

	if proc.user.teamMgrProxy == nil {
		proc.user.Error("KickTeamMemberReq failed, teamMgrProxy is nil")
		return
	}

	if proc.user.GetPlayerGameState() == common.StateMatching {
		proc.user.teamMgrProxy.RPC(common.ServerTypeMatch, "CancelQueue", proc.user.GetTeamID())
	}

	if err := proc.user.teamMgrProxy.RPC(common.ServerTypeMatch, "KickTeamMember",
		proc.user.GetID(), proc.user.GetTeamID(), memID); err != nil {
		proc.user.Error("RPC KickTeamMemberReq err: ", err)
	}

	proc.user.Info("Send kick team member req to match success, memID: ", memID)
}

// RPC_KickTeamMember 踢出队员通知
func (proc *LobbyUserMsgProc) RPC_KickTeamMember(leaderID, memID uint64) {
	if proc.user.GetID() == memID {
		proc.user.SetTeamID(0)
		proc.user.teamInfo = nil
		state := proc.user.GetPlayerGameState()
		if state != common.StateGame && state != common.StateWatch {
			proc.user.SetPlayerGameState(common.StateFree)
		}
	}
	proc.user.RPC(iserver.ServerTypeClient, "KickTeamMemberNotify", memID)
}

// RPC_TransferTeamLeaderReq 转让队长请求
func (proc *LobbyUserMsgProc) RPC_TransferTeamLeaderReq(memID uint64) {
	if proc.user == nil || proc.user.GetTeamID() == 0 {
		return
	}

	if proc.user.teamMgrProxy == nil {
		proc.user.Error("TransferTeamLeaderReq failed, teamMgrProxy is nil")
		return
	}

	if err := proc.user.teamMgrProxy.RPC(common.ServerTypeMatch, "TransferTeamLeader",
		proc.user.GetID(), proc.user.GetTeamID(), memID); err != nil {
		proc.user.Error("RPC TransferTeamLeaderReq err: ", err)
	}

	proc.user.Info("Send transfer team leader req to match success, memID: ", memID)
}

// RPC_TransferTeamLeader 转让队长通知
func (proc *LobbyUserMsgProc) RPC_TransferTeamLeader(memID uint64) {
	proc.user.RPC(iserver.ServerTypeClient, "TransferTeamLeaderNotify", memID)
}

// RPC_QuickDialogReq 发送快捷对话请求
func (proc *LobbyUserMsgProc) RPC_QuickDialogReq(id uint32) {
	if proc.user == nil || proc.user.GetTeamID() == 0 {
		return
	}

	if proc.user.teamMgrProxy == nil {
		proc.user.Error("QuickDialogReq failed, teamMgrProxy is nil")
		return
	}

	if err := proc.user.teamMgrProxy.RPC(common.ServerTypeMatch, "QuickDialog",
		proc.user.GetID(), proc.user.GetTeamID(), id); err != nil {
		proc.user.Error("RPC QuickDialogReq err: ", err)
	}

	proc.user.Info("Send quick dialog req to match success, id: ", id)
}

// RPC_QuickDialog 快捷对话通知
func (proc *LobbyUserMsgProc) RPC_QuickDialog(memID uint64, id uint32) {
	proc.user.RPC(iserver.ServerTypeClient, "QuickDialogNotify", memID, id)
}

// RPC_ShowRecordReq 战绩展示请求
func (proc *LobbyUserMsgProc) RPC_ShowRecordReq() {
	if proc.user == nil || proc.user.GetTeamID() == 0 {
		return
	}

	if proc.user.teamMgrProxy == nil {
		proc.user.Error("ShowRecordReq failed, teamMgrProxy is nil")
		return
	}

	if err := proc.user.teamMgrProxy.RPC(common.ServerTypeMatch, "ShowRecord",
		proc.user.GetID(), proc.user.GetTeamID()); err != nil {
		proc.user.Error("RPC ShowRecordReq err: ", err)
	}

	proc.user.Info("Send show record req to match success")
}

// RPC_ShowRecord 战绩展示通知
func (proc *LobbyUserMsgProc) RPC_ShowRecord(msg *protoMsg.GameRecordDetail) {
	proc.user.RPC(iserver.ServerTypeClient, "ShowRecordNotify", msg)
}

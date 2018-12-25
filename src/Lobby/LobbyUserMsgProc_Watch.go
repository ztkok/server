package main

import (
	"common"
	"db"
	"time"
	"zeus/dbservice"
	"zeus/iserver"
)

// RPC_SetWatchable Client->Lobby 设置是否允许好友观战
func (proc *LobbyUserMsgProc) RPC_SetWatchable(watchable uint32) {
	proc.user.SetWatchable(watchable)
	proc.user.friendMgr.syncWatchable(watchable)
	proc.user.Info("SetWatchable: ", watchable)
}

// RPC_WatchTargetBattle Client->Lobby 观战目标, targetUID 目标的uid
func (proc *LobbyUserMsgProc) RPC_WatchTargetBattle(targetUID uint64) {
	targetEntityID, err := dbservice.SessionUtil(targetUID).GetUserEntityID()
	if err != nil {
		proc.user.Error("Watch target battle failed ", err)
		return
	}
	var errCode uint32
	var mapid, skybox, watchNum, teamWatchNum, matchMode uint32
	var spaceID uint64
	var waitTime uint64
	tempUtil := db.PlayerTempUtil(targetUID)

	state := db.PlayerTempUtil(proc.user.GetDBID()).GetGameState()

	if (state != common.StateFree && state != common.StateMatchWaiting) || proc.user.isInTeamReady {
		errCode = WatchBattleErrCodeCanNot

		goto END
	}

	if proc.user.friendMgr.getFriendWatchable(targetUID) != 0 {
		errCode = WatchBattleErrCodeNotAllow
		goto END
	}
	if tempUtil.GetGameState() != common.StateGame {
		errCode = WatchBattleErrCodeNotInGame
		goto END
	}

	if enterTime, now := tempUtil.GetEnterGameTime()+uint64(common.GetTBSystemValue(common.System_WatchWait)), uint64(time.Now().Unix()); enterTime > now {
		waitTime = enterTime - now
		errCode = WatchBattleErrCodeWaitTime
		goto END
	}

	if teamID := tempUtil.GetPlayerTeamID(); teamID != 0 {
		teamWatchNum = db.PlayerTeamUtil(teamID).GetTeamWatchNum()
	} else {
		teamWatchNum = db.SceneTempUtil(spaceID).GetSingleUserWatchNum(targetUID)
	}
	if teamWatchNum >= uint32(common.GetTBSystemValue(common.System_WatchTeamLimit)) {
		errCode = WatchBattleErrCodeNumLimit
		goto END
	}

	// 获取目标所在的SpaceID
	spaceID = tempUtil.GetPlayerSpaceID()
	if spaceID == 0 {
		return
	}

	mapid, skybox, watchNum, matchMode, err = db.SceneTempUtil(spaceID).GetInfo()
	if err != nil {
		proc.user.Error("Watch battle failed ", err)
		return
	}
	if watchNum >= uint32(common.GetTBSystemValue(common.System_WatchSpaceLimit)) {
		errCode = WatchBattleErrCodeNumLimit
		goto END
	}

	proc.user.watchTarget = targetUID
	proc.user.clearWatchTimer = time.AfterFunc(10*time.Second, func() {
		proc.user.watchTarget = 0
	})
	// 原流程, 客户端收到MatchSuccess才加载地图
	proc.user.RPC(iserver.ServerTypeClient, "EnterWatchSuccess", mapid, skybox, targetEntityID)
	proc.user.RPC(iserver.ServerTypeClient, "RecvMatchModeId", matchMode)
	proc.user.EnterSpace(spaceID)
	proc.user.WatchFlow(WatchEnter, proc.user.watchTarget, spaceID, teamWatchNum, 0, 0)
	proc.user.Info("Watch battle, target: ", targetUID)
END:
	proc.user.RPC(iserver.ServerTypeClient, "WatchTargetBattleRet", errCode, waitTime)
	proc.user.Info("WatchTargetBattleRet, errCode: ", errCode, " waitTime: ", waitTime)
}

// RPC_LeaveWatch 离开观战
func (proc *LobbyUserMsgProc) RPC_LeaveWatchTarget() {
	proc.user.watchTarget = 0
	proc.user.SetPlayerGameState(common.StateFree)
}

//RPC_WatchFlow 离开观战tlog
func (proc *LobbyUserMsgProc) RPC_WatchFlow(spaceId uint64, watchTime uint32, loading uint32, num uint32) {
	proc.user.WatchFlow(WatchLeave, proc.user.watchTarget, spaceId, num, watchTime, loading)
}

// RPC_SetPlayerWatchState 设置游戏 继续观战时观战者为观战状态
func (proc *LobbyUserMsgProc) RPC_SetPlayerWatchState() {
	proc.user.SetPlayerGameState(common.StateWatch)
}

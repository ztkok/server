package main

import (
	"excel"
	"math"
	"time"
	"zeus/iserver"
)

// doFriendWatch 处理好友观战
func (user *RoomUser) doFriendWatch(targetUser *RoomUser) {
	scene := user.GetSpace().(*Scene)
	if scene == nil {
		return
	}

	scene.UpdateWatchNum(user.watchingTarget, 1, false)

	//设置队友信息
	teamInfo := scene.teamMgr.GetInitRoomTeamInfo(targetUser.GetUserTeamID())
	if teamInfo != nil {
		user.RPC(iserver.ServerTypeClient, "InitRoomTeamInfoRet", teamInfo)
	}

	//同步当前游戏状态
	user.SendFullGameState()

	user.DisposeWatch(targetUser.GetID())
	user.SyncWatchTargetInfo()

	user.sumData.watchStartTime = time.Now().Unix()
}

// doLoadingDone 好友观战玩家loading结束进入
func (user *RoomUser) doLoadingDone(loading bool) {
	space := user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	if !loading && user.watchLoading == 0 {
		return
	}

	targetUser, ok := space.GetEntity(user.watchingTarget).(*RoomUser)
	if !ok {
		return
	}

	user.watchLoading = 1
	space.watchers[user.GetID()] = true

	space.updateAliveNum(user)
	space.updateTotalNum(user)
	space.NotifyWatchNum(user.watchingTarget)
	user.RPC(iserver.ServerTypeClient, "UpdateKillNum", uint32(targetUser.GetKillNum()))

	if targetUser.IsOffline() {
		user.RPC(iserver.ServerTypeClient, "WatchTargetOffline")
	}

	if !space.teamMgr.isTeam {
		if math.Abs(float64(targetUser.signPos.X)) > 0.1 || math.Abs(float64(targetUser.signPos.Y)) > 0.1 {
			user.RPC(iserver.ServerTypeClient, "SyncMapSign", targetUser.GetID(), float64(targetUser.signPos.X), float64(targetUser.signPos.Y))
		}
	} else {
		for _, v := range targetUser.GetTeamMembers() {
			mem, ok := space.GetEntity(v).(*RoomUser)
			if !ok {
				continue
			}

			baseInfo := space.teamMgr.membersBaseInfo[mem.GetDBID()]
			if baseInfo != nil {
				signPos := baseInfo.SignPos
				if math.Abs(float64(signPos.X)) > 0.1 || math.Abs(float64(signPos.Y)) > 0.1 {
					user.RPC(iserver.ServerTypeClient, "SyncMapSign", baseInfo.ID, float64(baseInfo.SignPos.X), float64(baseInfo.SignPos.Y))
				}
				if math.Abs(float64(baseInfo.DiePos.X)) > 0.1 || math.Abs(float64(baseInfo.DiePos.Y)) > 0.1 {
					user.RPC(iserver.ServerTypeClient, "ShowDieFlag", baseInfo.ID, float64(baseInfo.DiePos.X), float64(baseInfo.DiePos.Z))
				}
			}

			if mem.GetBaseState() == RoomPlayerBaseState_WillDie {
				if mem.stateMgr.downEndStamp != 0 && mem.stateMgr.downEndStamp > time.Now().Unix() {
					downTime, _ := excel.GetSystem(9)
					// 通知玩家濒死进度
					if mem.GetDBID() == targetUser.GetDBID() { //只需要被观战者的进度
						user.SyncProgressBar(USERDOWNTYPE, uint64(mem.stateMgr.downEndStamp), uint64(downTime.Value), uint64(mem.stateMgr.downEndStamp)-uint64(time.Now().Unix()))
					}
					user.RPC(iserver.ServerTypeClient, "SyncDownEndTime", mem.GetID(), uint32(mem.stateMgr.downEndStamp), uint32(mem.stateMgr.downEndStamp)-uint32(time.Now().Unix()), mem.stateMgr.isBerescue)
				}
			}
		}
	}
}

package main

import (
	"common"
	"db"
	"time"
	"zeus/iserver"
)

// RPC_WatchTargetBattle Lobby->Room 好友观战
func (p *RoomUserMsgProc) RPC_WatchTargetBattle(targetUID uint64) {
	scene, ok := p.user.GetSpace().(*Scene)
	if !ok {
		p.user.Error("Watch target battle failed, get scene failed")
		p.user.RPC(common.ServerTypeLobby, "LeaveWatchTarget")
		return
	}

	target, ok := scene.GetEntityByDBID("Player", targetUID).(*RoomUser)
	if !ok {
		p.user.Error("Watch target battle failed, find target failed")
		p.user.RPC(common.ServerTypeLobby, "LeaveWatchTarget")
		return
	}

	if target.GetState() == RoomPlayerBaseState_Dead || target.GetState() == RoomPlayerBaseState_Watch {
		p.user.RPC(common.ServerTypeLobby, "LeaveWatchTarget")
		target.RPC(iserver.ServerTypeClient, "WatchTargetLeave")
		return
	}

	scene.watchers[p.user.GetID()] = true
	p.user.userType = RoomUserTypeWatcher

	p.user.doFriendWatch(target)

	//为单人玩家初始化新的语音房间
	voiceRoomId := target.GetUserTeamID()
	if voiceRoomId == 0 {
		if target.watchVoice == 0 {
			target.watchVoice = db.GetTeamGlobalID()
			target.RPC(iserver.ServerTypeClient, "SyncWatcherVoiceRoom", target.watchVoice)
		}
		voiceRoomId = target.watchVoice
	}
	p.user.RPC(iserver.ServerTypeClient, "SyncWatcherVoiceRoom", voiceRoomId)

	p.user.Info("Watch target battle ", targetUID)
}

// RPC_WatchBattle Client->Room 观战 兼容旧包的观战
func (p *RoomUserMsgProc) RPC_WatchBattle() {
	space := p.user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	target := space.getWatchTarget(p.user, true)
	if target != 0 {
		p.user.DisposeWatch(target)
		p.user.SyncWatchTargetInfo()
		p.user.GetEntities().RemoveDelayCall(p.user.leaveSceneTimer)
		p.user.sumData.watchStartTime = time.Now().Unix()
	}

	p.user.Info("Watch battle success, target: ", target)
}

// RPC_WatchBattlePre Client->Room 新的观战流程 拆成两部分 1 处理AOI结构 解决前端无法获得最新的Entity数据问题
func (p *RoomUserMsgProc) RPC_WatchBattlePre(isReConn bool) {
	space := p.user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	target := space.getWatchTarget(p.user, !isReConn)
	if target != 0 {
		p.user.DisposeWatch(target)
		p.user.GetEntities().RemoveDelayCall(p.user.leaveSceneTimer)
		p.user.sumData.watchStartTime = time.Now().Unix()
	}

	p.user.RPC(iserver.ServerTypeClient, "WatchBattlePre", target)
	p.user.Info("Watch battle success, target: ", target)
}

// WatchBattleSync Client->Room
func (p *RoomUserMsgProc) RPC_WatchBattleSync() {
	p.user.SyncWatchTargetInfo()

	p.user.RPC(common.ServerTypeLobby, "SetPlayerWatchState")
}

// RPC_WatchBattleStart Client->Room 新的观战流程 拆成两部分 2 触发前端观战进程
func (p *RoomUserMsgProc) RPC_WatchBattleStart() {
	space := p.user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	// 在AOI范围内
	inMyAOI := false
	space.TravsalAOI(p.user, func(n iserver.ICoordEntity) {
		if n.GetID() == p.user.watchingTarget {
			inMyAOI = true
		}
	})

	if inMyAOI {
		p.user.SendFullAOIs()
	}
}

// RPC_CancelWatchBattle Client->Room 取消观战
func (p *RoomUserMsgProc) RPC_CancelWatchBattle() {
	p.user.sumData.watchType = 1 //1主动退出
	if p.user.sumData.watchStartTime != 0 {
		p.user.sumData.watchEndTime = time.Now().Unix()
	}

	if p.user.GetBaseState() == RoomPlayerBaseState_Watch {
		p.user.SetBaseState(RoomPlayerBaseState_Dead)
	}

	if p.user.IsWatching() {
		p.user.CancelWatch()
	}

	p.user.Info("Cancel watch battle")
}

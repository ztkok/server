package main

import (
	"common"
	"excel"
	"protoMsg"
	"time"
	"zeus/iserver"
	"zeus/linmath"
	"zeus/msgdef"
)

// 客户端断开连接
func (p *RoomUserMsgProc) MsgProc_SessClosed(content interface{}) {

	p.user.Info("MsgProc_SessClosed, player offline")
	p.user.SetOffline(true)

	space := p.user.GetSpace().(*Scene)
	if space == nil || space.isEnd() {
		return
	}

	if p.user.IsFollowedParachute() || p.user.IsFollowingParachute() {
		space.teamMgr.CancelFollowParachute(p.user)
	}

}

// 有客户端连进来，包含正常的初次连接和断线重连
func (p *RoomUserMsgProc) MsgProc_SpaceUserConnect(content interface{}) {
	if p.user == nil {
		return
	}

	p.user.SetOffline(false)

	tb := p.user.GetSpace().(*Scene)
	if tb == nil {
		return
	}

	p.user.lastFireTime = 0
	baseState := p.user.GetBaseState()

	switch baseState {
	case RoomPlayerBaseState_LoadingMap:
		// 初始状态，不做处理
		return
	case RoomPlayerBaseState_Dead, RoomPlayerBaseState_LeaveMap:
		// 玩家已退出场景，可能是吃鸡或死亡
		p.user.Post(iserver.ServerTypeClient, &msgdef.SpaceUserConnectGameEnd{})
		return
	case RoomPlayerBaseState_Inplane:
		p.user.SendFullAOIs()

		if tb.allowParachute {
			p.user.RPC(iserver.ServerTypeClient, "AllowParachute")
		}
	case RoomPlayerBaseState_Glide, RoomPlayerBaseState_Parachute:
		pos := p.user.GetPos()
		var err error
		if p.user.parachuteCtrl != nil {
			// 估算当前位置
			pos, err = p.user.parachuteCtrl.CalcCurPos()
			if err != nil {
				p.user.Warn(err)
				return
			}
		}

		p.user.GetPlayerState().SetTimeStamp(p.user.GetSpace().GetTimeStamp())
		p.user.SetPos(pos)
		p.user.GetSpace().UpdateCoord(p.user)
		p.user.SendFullAOIs()
		p.user.sendEmptySyncMsg()
		p.user.SendFullState()
		p.user.sendSafeZone()
		p.user.SendAllObj()
		//在空中获得枪支 断线重连后通知
		for _, v := range p.user.equips {
			if v.thisid == p.user.useweapon {
				p.user.RefreshGunNotifyAll(v)
				break
			}
		}
	default:
		p.user.sendAllGameData()
	}

	p.user.StateM.hasinit = true

	//同步队伍内的跟随关系
	tb.teamMgr.TeamFollowSync(p.user)

	tb.updateAliveNum(p.user)
	p.user.RPC(iserver.ServerTypeClient, "UpdateKillNum", uint32(p.user.GetKillNum()))

	tb.updateTotalNum(p.user)
	p.user.RPC(iserver.ServerTypeClient, "UpdateWatchNum", uint32(len(tb.watchers)))
	tb.resendData(p.user)

	// 发送给自己队友信息
	if tb.teamMgr != nil && tb.teamMgr.isTeam {
		tb.teamMgr.SendMyTeamInfoToMe(p.user)

		p.user.RPC(iserver.ServerTypeClient, "BreakProgressBar")
		if p.user.GetBaseState() == RoomPlayerBaseState_WillDie {
			if tb.teamMgr.isExistTeammate(p.user) && p.user.stateMgr.downEndStamp != 0 && p.user.stateMgr.downEndStamp > time.Now().Unix() {
				downTime, _ := excel.GetSystem(9)
				// 通知玩家濒死进度
				p.user.SyncProgressBar(USERDOWNTYPE, uint64(p.user.stateMgr.downEndStamp), uint64(downTime.Value), uint64(p.user.stateMgr.downEndStamp)-uint64(time.Now().Unix()))

				// 通知队友自己处于濒死状态
				tb.teamMgr.SyncDownEndTime(p.user, uint64(p.user.stateMgr.downEndStamp), uint64(p.user.stateMgr.downEndStamp)-uint64(time.Now().Unix()))
			}
		}
	}

	// // 同步当前观战对象
	// if baseState == RoomPlayerBaseState_Watch {
	// 	p.user.SyncWatchTargetInfo()
	// }

	// 通知玩家被观战
	if len(p.user.watchers) > 0 {
		p.user.RPC(iserver.ServerTypeClient, "BeWatched")
	}

	p.user.syncCurRoleSkillInfo() // 通知当前使用的技能
	p.user.refreshSkillEffect()
	p.user.UpdateCell()
}

func (p *RoomUserMsgProc) RPC_ReEnterSpace() {
	p.user.OnReEnterSpace()
}

func (user *RoomUser) IsOffline() bool {
	return user._offline
}

func (user *RoomUser) SetOffline(offline bool) {
	if user._offline && !offline {
		user._offline = offline
		// TODO:通知队友，该玩家上线
	} else if !user._offline && offline {
		user._offline = offline
		// TODO: 通知队友，该玩家离线
	}
}

// 客户端为正确处理恢复流程需要
func (user *RoomUser) sendEmptySyncMsg() {
	bytes := make([]byte, 1)
	msg := &msgdef.PropsSyncClient{
		EntityID: user.GetID(),
		Num:      uint32(0),
		Data:     bytes,
	}
	user.CastMsgToMe(msg)
}

// 通知区域信息(包括:毒圈、目标区、轰炸区等)
func (user *RoomUser) notifyZoneInfo(zonetype uint32, center linmath.Vector3, radius float32, interval uint32) {
	zoneproto := &protoMsg.ZoneNotify{}
	zoneproto.Type = zonetype
	zoneproto.Center = &protoMsg.Vector3{
		X: center.X,
		Y: center.Y,
		Z: center.Z,
	}
	zoneproto.Radius = radius
	zoneproto.Interval = interval
	user.CastRPCToMe("ZoneNotify", zoneproto)
}

func (user *RoomUser) sendSafeZone() {
	scene := user.GetSpace().(*Scene)
	sf := scene.refreshzone
	mininow := uint64(time.Now().UnixNano() / (1000 * 1000))
	secnow := time.Now().Unix()

	base, ok := excel.GetMaprule(sf.refreshcount)
	if ok {
		//先清除轰炸区
		user.RPC(iserver.ServerTypeClient, "BombDisapear")
		// 存在轰炸区
		for _, area := range GetRefreshZoneMgr(scene).bombAreas {
			if mininow < area.bombendtime && area.bombradius != 0 {
				pos := area.bombcenter
				user.RPC(iserver.ServerTypeClient, "BombRefresh", pos.X, pos.Y, pos.Z, area.bombradius)
			}
		}

		// 禁令区通知
		if sf.refreshstatus == common.Status_Init {
			// 未开始收缩圈
			user.notifyZoneInfo(3, scene.safecenter, scene.saferadius, 0)
			user.notifyZoneInfo(2, sf.nextsafecenter, sf.nextsaferadius, 0)

			countdown := int64(base.Shrinkarea)
			if secnow <= sf.refreshtime+int64(base.Shrinkarea) {
				countdown = sf.refreshtime + int64(base.Shrinkarea) - secnow
			}
			user.RPC(iserver.ServerTypeClient, "ShrinkRefresh", uint64(countdown))

		} else if sf.refreshstatus == common.Status_ShrinkBegin {
			user.notifyZoneInfo(3, scene.safecenter, scene.saferadius, uint32(sf.shrinkinterval))
			user.notifyZoneInfo(2, sf.nextsafecenter, sf.nextsaferadius, uint32(sf.shrinkinterval))

			// user.notifyZoneInfo(0, scene.safecenter, scene.saferadius, uint32(sf.shrinkinterval))
			// 收缩圈
			// user.RPC(iserver.ServerTypeClient, "ShrinkRefresh", uint64(base.Shrinkarea))
		}
	}
}

// SendFullGameState 发送全局的游戏状态数据
func (user *RoomUser) SendFullGameState() {

	scene := user.GetSpace().(*Scene)
	if scene == nil {
		return
	}

	user.sendSafeZone()

	// 当前门的状态
	if scene.doorMgr != nil {
		scene.doorMgr.SendAllDoorStateToUser(user)
	} else {
		scene.Error("scene doorMgr is nil")
	}

	user.SendAllObj()
	user.reconnectGunNotify()

	GetRefreshItemMgr(scene).resendBoxList(user)

	if user.GetBaseState() == RoomPlayerBaseState_Watch && user.dieNotify != nil {
		user.RPC(iserver.ServerTypeClient, "DieNotifyRet", user.dieNotify)
	}

	if user.GetBaseState() == RoomPlayerBaseState_Passenger {
		prop := user.GetVehicleProp()
		if prop != nil {

			//载具坐标
			var cood iserver.ICoordEntity
			if car, ok := scene.cars[prop.Thisid]; ok {
				roomUser := car.GetUser()
				if roomUser.GetID() != user.GetID() {
					cood = roomUser
				}
			}

			if cood == nil {
				cood = scene.GetEntity(prop.Thisid).(*SpaceVehicle)
			}

			if cood != nil {
				user.SetPos(cood.GetPos())
			} else {
				user.Error("SetPos Error! cood == nil! prop:", prop)
			}
		}
	}

	if user.GetActionState() == RoomPlayerActionState_Position {
		user.RPC(iserver.ServerTypeClient, "FinishPotion")
	}

	user.resendKillDrop()
}

// 发送游戏内所需要的所有游戏状态数据
func (user *RoomUser) sendAllGameData() {

	user.GetSpace().UpdateCoord(user)

	if err := user.SendFullAOIs(); err != nil {
		user.Error("SendFullAOIs err: ", err)
	}

	user.sendEmptySyncMsg()
	user.SendFullState()

	if err := user.SendFullProps(); err != nil {
		user.Error("SendFullProps err: ", err)
	}
	if user.isInTank() { //玩家在坦克上 通知客户端是否需要重新切换武器
		first := &protoMsg.T_Object{}
		second := &protoMsg.T_Object{}
		for k, v := range user.tmpequips {
			if k == user.tmpuseweapon {
				first = v.fillInfo()
			} else {
				second = v.fillInfo()
			}
			user.spareGunSightNotify(v)
		}
		user.RPC(iserver.ServerTypeClient, "OldGunInfoNotify", first, second)
	}
	// 发送毒圈、目标区域、门的状态、轰炸区域等全局性的数据
	user.SendFullGameState()

	//同步aoi范围内的坦克状态
	user.syncAOITanksToMe()

	return
}

////////////////////////以下为大断线处理区域////////////////////////////////////
func (user *RoomUser) OnReEnterSpace() {
	msg := &msgdef.EnterSpace{
		SpaceID:   user.GetSpace().GetID(),
		MapName:   user.GetSpace().GetInitParam().(string),
		EntityID:  user.GetID(),
		Addr:      iserver.GetSrvInst().GetCurSrvInfo().OuterAddress,
		TimeStamp: user.GetSpace().GetTimeStamp(),
	}

	user.SetClient(nil)
	if err := user.Post(iserver.ServerTypeClient, msg); err != nil {
		user.Error("Send ReEnterSpace failed ", err)
	}

	scene := user.GetSpace().(*Scene)
	if user.GetBaseState() == RoomPlayerBaseState_Inplane || user.GetBaseState() == RoomPlayerBaseState_LoadingMap {
		// 飞机已经起飞
		if scene.allowParachute {
			start := scene.airlineStart
			end := scene.airlineEnd
			fliedSeconds := time.Now().Sub(scene.allowParachuteTime).Seconds()
			if err := user.RPC(iserver.ServerTypeClient, "PlanePosForAirline", float64(start.X), float64(start.Z), float64(end.X), float64(end.Z), float64(fliedSeconds)); err != nil {
				user.Error(err)
			}
		} else {
			// 飞机尚未起飞
			user.notifyAirlineInfo()
		}
	} else {
		user.notifyAirline("ReconnectAirLine")
	}
	// user.Debug("OnReEnterSpace")
}

// MsgProc_ReqUserGameState
func (user *RoomUser) RPC_CheckPreGameState() {
	if user.userType == RoomUserTypeWatcher {
		return
	}

	space := user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	// 当前状态 0：未组队死亡（或组队全死亡），1：未组队未死亡，2：组队自己未死亡 3:组队自己已死亡
	var gameState uint64

	if !space.teamMgr.isTeam { //单排
		if !user.IsDead() {
			gameState = 1 // 还活着
		}
	} else { // 组队
		if space.teamMgr.isExistTeam(user) {
			if !user.IsDead() {
				gameState = 2
			} else if !user.IsWatching() {
				gameState = 3
			}
		}
	}

	user.RPC(common.ServerTypeLobby, "NotifyContinuePreGame", gameState)

	//1：未组队未死亡，2：组队自己未死亡
	if gameState == 0 {
		user.userLeave()
		user.LeaveScene()
	}
}

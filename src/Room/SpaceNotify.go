package main

import (
	"common"
	"protoMsg"
	"zeus/iserver"
	"zeus/linmath"
	"zeus/msgdef"
)

func (tb *Scene) zoneNotify(zonetype uint32, center linmath.Vector3, radius float32, interval uint32) {
	zoneproto := &protoMsg.ZoneNotify{}
	zoneproto.Type = zonetype
	zoneproto.Center = &protoMsg.Vector3{
		X: center.X,
		Y: center.Y,
		Z: center.Z,
	}
	zoneproto.Radius = radius
	zoneproto.Interval = interval
	tb.BroadCastMsg(zoneproto, "ZoneNotify")
}

func (tb *Scene) chatNotify(chattype uint32, str string) {
	proto := &protoMsg.ChatNotify{}
	proto.Type = chattype
	proto.Content = str
	tb.BroadCastMsg(proto, "ChatNotify")
}

// BroadCastMsg 广播消息
func (tb *Scene) BroadCastMsg(msg msgdef.IMsg, msgName string) {
	tb.TravsalEntity("Player", func(e iserver.IEntity) {
		if e == nil {
			return
		}

		if user, ok := e.(*RoomUser); ok {
			//user.CastMsgToMe(msg)
			user.RPC(iserver.ServerTypeClient, msgName, msg)
		}
	})
}

// BroadDeathDropBox 死亡掉落箱广播
func (tb *Scene) BroadDeathDropBox(eid, thisid uint64) {
	tb.TravsalEntity("Player", func(e iserver.IEntity) {
		if e == nil {
			return
		}

		if user, ok := e.(*RoomUser); ok {
			user.RPC(iserver.ServerTypeClient, "DeathDropBoxNotify", eid, thisid)
		}
	})
}

// BroadDieNotify 死亡通知
func (tb *Scene) BroadDieNotify(attackerid uint64, defenederid uint64, isHeadShot bool, injuredInfo InjuredInfo) {

	proto := &protoMsg.DieNotifyRet{}

	proto.Type = uint32(injuredInfo.injuredType)
	proto.Attackerid = attackerid
	proto.Defenderid = defenederid
	proto.CurAliveSum = uint32(len(tb.members))

	if injuredInfo.injuredType == headShotAttack {
		proto.Type = 0
		proto.IsHeadhost = 1
	} else {
		proto.IsHeadhost = 0
	}

	if attackerid != 0 {
		attacker, ok := tb.GetEntity(attackerid).(iAttacker)
		if ok {
			proto.Attackername = attacker.GetName()
			proto.Attackerinsignia = attacker.GetInsignia()
			if proto.GetType() == gunAttack {

				gunid := attacker.GetInUseWeapon()
				if gunid == 0 {
					proto.Type = fistAttack
				} else {
					proto.Gunid = uint64(gunid)
				}
			}

			proto.Attackercolor = common.GetPlayerNameColor(proto.Attackerid)
		}
	}

	if defenederid != 0 {
		defender, ok := tb.GetEntity(defenederid).GetRealPtr().(iDefender)
		if ok {
			proto.Defendername = defender.GetName()
			proto.Defenderinsignia = defender.GetInsignia()
			state := defender.GetState()
			if state == RoomPlayerBaseState_Watch || state == RoomPlayerBaseState_Dead {
				proto.Defenderstate = 0
			} else if state == RoomPlayerBaseState_WillDie {
				proto.Defenderstate = 1
			} else {
				tb.Error("BroadDieNotify failed, state: ", state)
			}

			proto.Defendercolor = common.GetPlayerNameColor(proto.Defenderid)

			defender.SetDieNotify(proto)
		}
	}

	//广播死亡通知,通知给攻击者的消息特殊处理
	tb.TravsalEntity("Player", func(e iserver.IEntity) {
		if e == nil {
			return
		}
		if user, ok := e.(*RoomUser); ok {
			if attackerid != defenederid && attackerid == e.GetID() && !tb.teamMgr.IsInOneTeamByID(attackerid, defenederid) {
				user.MedalNotify(proto.Defenderstate == 1, injuredInfo.injuredType == headShotAttack, injuredInfo.injuredType)
			}

			if injuredInfo.killDownInjured && attackerid == e.GetID() {
				p := *proto
				// p.Type = 100
				user.RPC(iserver.ServerTypeClient, "DieNotifyRet", &p)
				return
			}
			user.RPC(iserver.ServerTypeClient, "DieNotifyRet", proto)
		}
	})

	tb.Debugf("BroadDieNotify success, DieNotifyRet: %+v\n", proto)
}

// BroadAliveNum 广播当前存活人数
func (tb *Scene) BroadAliveNum() {
	tb.TravsalEntity("Player", func(e iserver.IEntity) {
		if e == nil {
			return
		}

		if user, ok := e.(*RoomUser); ok {
			//tb.Error("广播存活人数:", user.GetName(), "num:", len(tb.members))
			// user.RPC(iserver.ServerTypeClient, "UpdateAliveNum", uint32(len(tb.members)))
			tb.updateAliveNum(user)
		}
	})
}

//BroadAirLeft 广播机舱剩余人数
func (tb *Scene) BroadAirLeft() {
	var ret uint32
	tb.TravsalEntity("Player", func(e iserver.IEntity) {
		if e == nil {
			return
		}

		if user, ok := e.(*RoomUser); ok {
			if user.haveenter && user.GetBaseState() == RoomPlayerBaseState_LoadingMap || user.GetBaseState() == RoomPlayerBaseState_Inplane {
				ret++
			}
		}
	})

	ret += tb.onAirAiNum

	tb.TravsalEntity("Player", func(e iserver.IEntity) {
		if e == nil {
			return
		}

		if user, ok := e.(*RoomUser); ok {
			//tb.Debug("广播机舱人数:", user.GetName(), " num:", ret, " !!!:", tb.onAirAiNum)
			user.RPC(iserver.ServerTypeClient, "UpdateAirLeft", ret)
		}
	})
}

// BroadBattleOver 广播游戏结束
func (tb *Scene) BroadBattleOver(winner uint64) {
	tb.TravsalEntity("Player", func(e iserver.IEntity) {
		if e == nil {
			return
		}

		if user, ok := e.(*RoomUser); ok {
			user.RPC(iserver.ServerTypeClient, "BattleOverNotify", winner)
		}
	})
}

func (tb *Scene) rpcBombRefresh(pos linmath.Vector3, r float32, id uint32, caller string, color uint32) {
	tb.TravsalEntity("Player", func(e iserver.IEntity) {
		if e == nil {
			return
		}

		if user, ok := e.(*RoomUser); ok {
			user.RPC(iserver.ServerTypeClient, "BombRefresh", pos.X, pos.Y, pos.Z, r)
			user.RPC(iserver.ServerTypeClient, "BombRefreshSync", pos.X, pos.Y, pos.Z, r, id, caller, color)
		}
	})
}

func (tb *Scene) rpcBombDam(pos linmath.Vector3, r float32, id uint32) {
	tb.TravsalEntity("Player", func(e iserver.IEntity) {
		if e == nil {
			return
		}

		if user, ok := e.(*RoomUser); ok {
			user.RPC(iserver.ServerTypeClient, "BombDam", pos.X, pos.Y, pos.Z, r)
			user.RPC(iserver.ServerTypeClient, "BombDamSync", pos.X, pos.Y, pos.Z, r, id)
		}
	})
}

func (tb *Scene) rpcBombDisapear(id uint32) {
	tb.TravsalEntity("Player", func(e iserver.IEntity) {
		if e == nil {
			return
		}

		if user, ok := e.(*RoomUser); ok {
			user.RPC(iserver.ServerTypeClient, "BombDisapear")
			user.RPC(iserver.ServerTypeClient, "BombDisapearSync", id)
		}
	})
}

// rpcSpaceDoorState 同步门的状态
func (tb *Scene) rpcSpaceDoorState(id uint64, newState uint32) {
	tb.TravsalEntity("Player", func(e iserver.IEntity) {
		if e == nil {
			return
		}

		if user, ok := e.(*RoomUser); ok {
			//tb.Debug("同步门的状态 ", user.GetName(), user.GetDBID(), id, newState)
			user.RPC(iserver.ServerTypeClient, "SpaceDoorState", id, newState)
		}
	})
}

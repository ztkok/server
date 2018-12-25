package main

import (
	"common"
	"protoMsg"
	"time"
	"zeus/iserver"
	"zeus/linmath"
)

func (user *RoomUser) addKillDrop(thisid uint64, pos linmath.Vector3) {
	space := user.GetSpace().(*Scene)
	now := time.Now().Unix()
	killdroplast := int64(common.GetTBSystemValue(common.System_KillDropLast))
	killdrop := &protoMsg.KillDrop{}
	killdrop.Thisid = thisid
	killdrop.Pos = &protoMsg.Vector3{
		X: pos.X,
		Y: pos.Y,
		Z: pos.Z,
	}
	killdrop.Disappeartime = uint64(now + killdroplast)

	if !space.teamMgr.isTeam {
		user.killdrop[thisid] = killdrop

		item := &protoMsg.KillDrop{
			Thisid:        killdrop.Thisid,
			Pos:           killdrop.Pos,
			Disappeartime: uint64(killdroplast),
		}
		user.RPC(iserver.ServerTypeClient, "KillDrop", item)
		return
	}

	team, ok := space.teamMgr.teams[user.GetUserTeamID()]
	if !ok {
		return
	}

	for _, memid := range team {
		tmpUser, ok := space.GetEntityByDBID("Player", memid).(*RoomUser)
		if !ok {
			continue
		}

		tmpUser.killdrop[thisid] = killdrop

		item := &protoMsg.KillDrop{
			Thisid:        killdrop.Thisid,
			Pos:           killdrop.Pos,
			Disappeartime: uint64(killdroplast),
		}
		tmpUser.RPC(iserver.ServerTypeClient, "KillDrop", item)
	}
}

func (user *RoomUser) removeKillDrop(thisid uint64) {
	space := user.GetSpace().(*Scene)
	if !space.teamMgr.isTeam {
		if _, ok := user.killdrop[thisid]; ok {
			delete(user.killdrop, thisid)
			user.RPC(iserver.ServerTypeClient, "KillDropDisappear", thisid)
		}
		return
	}

	team, ok := space.teamMgr.teams[user.GetUserTeamID()]
	if !ok {
		return
	}

	for _, memid := range team {
		tmpUser, ok := space.GetEntityByDBID("Player", memid).(*RoomUser)
		if !ok {
			continue
		}

		if _, ok := tmpUser.killdrop[thisid]; ok {
			delete(tmpUser.killdrop, thisid)
			tmpUser.RPC(iserver.ServerTypeClient, "KillDropDisappear", thisid)
		}
	}
}

func (user *RoomUser) resendKillDrop() {
	proto := &protoMsg.KillDropList{}
	now := uint64(time.Now().Unix())
	for _, v := range user.killdrop {
		if now < v.Disappeartime {

			item := &protoMsg.KillDrop{
				Thisid:        v.Thisid,
				Pos:           v.Pos,
				Disappeartime: v.Disappeartime - now,
			}
			proto.Data = append(proto.Data, item)
		}
	}

	user.RPC(iserver.ServerTypeClient, "KillDropList", proto)
}

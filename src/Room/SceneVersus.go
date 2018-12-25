package main

import (
	"common"
	"excel"
	"fmt"
	"protoMsg"
	"time"
	"zeus/iserver"

	log "github.com/cihub/seelog"
)

// 红蓝军团对抗scene
type VersusScene struct {
	SceneData
	boxlist         *protoMsg.DropBoxPosList
	redOrBlueAttack bool // 是否开启 红军 或 蓝军 之间伤害
	normalAttack    bool // 是否开启 普通玩家 之间伤害
	initRedAndBlue  bool // 是否分配完红军和蓝军
	redMember       map[uint64]bool
	blueMember      map[uint64]bool
	redTotalNum     uint32 // 红方总人数
	blueTotalNum    uint32 // 蓝方总人数

	notifyNum      int64 // 初始化预期通知次数
	equipPosNotify int64 // 控制通知cli 红方 和 蓝方 的位置频率
}

func NewVersusScene(sc *Scene) *VersusScene {
	scenedata := &VersusScene{}
	d, ok := excel.GetSystem(common.System_RedAndBlueAttackS)
	if ok {
		scenedata.redOrBlueAttack = d.Value == 1
	}
	d, ok = excel.GetSystem(common.System_RedAndBlueAttackN)
	if ok {
		scenedata.normalAttack = d.Value == 1
	}
	scenedata.initRedAndBlue = false

	scenedata.redMember = make(map[uint64]bool)
	scenedata.blueMember = make(map[uint64]bool)

	scenedata.scene = sc
	scenedata.boxlist = &protoMsg.DropBoxPosList{}
	return scenedata
}

func (v *VersusScene) canAttack(user IRoomChracter, defendid uint64) bool {
	defender, ok := v.scene.GetEntity(defendid).(IRoomChracter)
	if !ok {
		return true
	}
	uElite := user.GetGamerType()
	dElite := defender.GetGamerType()

	// 普通之间玩家是否开启伤害
	if uElite == RoomUserTypeNormal && dElite == RoomUserTypeNormal {
		return v.normalAttack
	}

	if v.redOrBlueAttack {
		return true
	}

	// 红军之间是否开启伤害
	if uElite == RoomUserTypeRed && dElite == RoomUserTypeRed {
		if len(v.blueMember) == 0 { // 判断蓝军是否都死光光
			return true
		} else {
			return false
		}
	}

	// 蓝军之间是否开启伤害
	if uElite == RoomUserTypeBlue && dElite == RoomUserTypeBlue {
		if len(v.redMember) == 0 { // 判断红军是否都死光光
			return true
		} else {
			return false
		}
	}

	return true
}

// doLoop
func (v *VersusScene) doLoop() {
	if v.scene.parachuteTime == 0 {
		return
	}

	now := time.Now().Unix()
	// 刷新红军和蓝军
	v.redAndBlueFresh(now)
	// 通知刷新红方和蓝方
	v.redAndBlueFreshNotify(now)
	// 通知红方和蓝方位置
	if now >= v.equipPosNotify {
		v.notifyRedAndBluePos()
		l, ok := excel.GetSystem(common.System_RedAndBluePosFreshTime)
		if ok {
			v.equipPosNotify = now + int64(l.Value)
		} else {
			v.equipPosNotify = now + 60
		}
	}

	// 红蓝军团是否伤害刷新
	v.freshAttackState(now)
}

// redAndBlueFresh 刷新红军和蓝军
func (v *VersusScene) redAndBlueFresh(now int64) {
	if v.initRedAndBlue {
		return
	}
	d, ok := excel.GetSystem(common.System_RedAndBlueFreshTime)
	if !ok {
		return
	}

	if now >= v.scene.parachuteTime+int64(d.Value) {
		aiNum := v.scene.aiNumSurplus
		playerNum := uint32(v.scene.getMemSum()) - aiNum
		v.redTotalNum = (playerNum+1)/2 + aiNum/2
		v.blueTotalNum = playerNum/2 + (aiNum+1)/2

		v.setRedAndBluePlayer()
		v.setRedAndBlueAi()

		v.initRedAndBlue = true
		v.scene.TravsalEntity("Player", func(entity iserver.IEntity) {
			if user, ok := entity.(*RoomUser); ok {
				v.updateTotalNum(user) // 通知总人数
			}
		})
	}
}

// setRedAndBluePlayer 将真实玩家划分为红蓝军团
func (v *VersusScene) setRedAndBluePlayer() {
	var i int = 0
	playerNum := v.scene.getMemSum() - int(v.scene.aiNumSurplus)

	errorcode, ok := excel.GetErrorcode(2002)
	if !ok {
		return
	}
	sRed := fmt.Sprintf(errorcode.Content, "红", "蓝")
	sBlue := fmt.Sprintf(errorcode.Content, "蓝", "红")

	v.scene.TravsalEntity("Player", func(entity iserver.IEntity) {
		user, ok := entity.(*RoomUser)
		if !ok {
			return
		}

		msg := &protoMsg.ChatNotify{}
		msg.Type = 1
		if i < int(playerNum+1)/2 {
			if user.GetGamerType() != RoomUserTypeRed {
				v.broadVersusState(user, RoomUserTypeRed)
			}
			user.SetGamerType(RoomUserTypeRed)
			v.redMember[user.GetID()] = true
			i++

			msg.Content = sRed
			user.RPC(iserver.ServerTypeClient, "ChatNotify", msg)
		} else {
			if user.GetGamerType() != RoomUserTypeBlue {
				v.broadVersusState(user, RoomUserTypeBlue)
			}
			user.SetGamerType(RoomUserTypeBlue)
			v.blueMember[user.GetID()] = true

			msg.Content = sBlue
			user.RPC(iserver.ServerTypeClient, "ChatNotify", msg)
		}
	})
}

// setRedAndBlueAi 将Ai划分为红蓝军团
func (v *VersusScene) setRedAndBlueAi() {
	var i int = 0
	aiNum := v.scene.aiNumSurplus
	v.scene.TravsalEntity("AI", func(entity iserver.IEntity) {
		Ai, ok := entity.(*RoomAI)
		if !ok {
			return
		}

		if i < int(aiNum)/2 {
			if Ai.GetGamerType() != RoomUserTypeRed {
				v.broadVersusState(Ai, RoomUserTypeRed)
			}
			Ai.SetGamerType(RoomUserTypeRed)
			v.redMember[Ai.GetID()] = true
			i++
		} else {
			if Ai.GetGamerType() != RoomUserTypeBlue {
				v.broadVersusState(Ai, RoomUserTypeBlue)
			}
			Ai.SetGamerType(RoomUserTypeBlue)
			v.blueMember[Ai.GetID()] = true
		}
	})
}

// redAndBlueFreshNotify 通知红方和蓝方刷新的时间
func (v *VersusScene) redAndBlueFreshNotify(now int64) {
	if v.initRedAndBlue {
		return
	}

	d, ok := excel.GetSystem(common.System_RedAndBlueFreshTime)
	if !ok {
		return
	}

	freshTime := v.scene.parachuteTime + int64(d.Value) - 10
	if now >= freshTime+v.notifyNum {
		if errorcode, ok := excel.GetErrorcode(2001); ok {
			s := fmt.Sprintf(errorcode.Content, 10-v.notifyNum)
			v.scene.chatNotify(1, s)
			v.notifyNum++
		}
	}
}

// showRedAndBluePos 是否显示红蓝军团位置
func (v *VersusScene) showRedAndBluePos() bool {
	return !v.redOrBlueAttack
}

// notifyRedAndBluePos 通知红蓝军团各自的位置
func (v *VersusScene) notifyRedAndBluePos() {
	if !v.showRedAndBluePos() {
		return
	}

	redPosList := &protoMsg.PlayerPosList{}
	for uid := range v.redMember {
		user, ok := v.scene.GetEntity(uid).(IRoomChracter)
		if !ok {
			continue
		}

		if user.GetGamerType() != RoomUserTypeRed {
			continue
		}
		pos := &protoMsg.PlayerPos{
			Id:  uid,
			Typ: RoomUserTypeRed,
			Pos: &protoMsg.Vector3{
				X: user.GetPos().X,
				Y: user.GetPos().Y,
				Z: user.GetPos().Z,
			},
		}
		redPosList.PlayerPos = append(redPosList.PlayerPos, pos)
	}

	bluePosList := &protoMsg.PlayerPosList{}
	for uid := range v.blueMember {
		user, ok := v.scene.GetEntity(uid).(IRoomChracter)
		if !ok {
			continue
		}

		if user.GetGamerType() != RoomUserTypeBlue {
			continue
		}
		pos := &protoMsg.PlayerPos{
			Id:  uid,
			Typ: RoomUserTypeBlue,
			Pos: &protoMsg.Vector3{
				X: user.GetPos().X,
				Y: user.GetPos().Y,
				Z: user.GetPos().Z,
			},
		}
		bluePosList.PlayerPos = append(bluePosList.PlayerPos, pos)
	}

	v.scene.TravsalEntity("Player", func(entity iserver.IEntity) {
		user, ok := entity.(*RoomUser)
		if !ok {
			return
		}

		if user.GetGamerType() == RoomUserTypeRed {
			if len(redPosList.PlayerPos) == 0 {
				return
			}

			user.RPC(iserver.ServerTypeClient, "SyncRedAndBluePos", redPosList)
		} else if user.GetGamerType() == RoomUserTypeBlue {
			if len(bluePosList.PlayerPos) == 0 {
				return
			}

			user.RPC(iserver.ServerTypeClient, "SyncRedAndBluePos", bluePosList)
		}
	})
}

// onDeath 死亡回调
func (v *VersusScene) onDeath(user IRoomChracter) {
	if _, ok1 := v.redMember[user.GetID()]; !ok1 {
		if _, ok2 := v.blueMember[user.GetID()]; !ok2 {
			log.Debug("onDeath no player! ID:", user.GetID())
			return
		}
	}

	if user.GetGamerType() == RoomUserTypeRed {
		delete(v.redMember, user.GetID())
	} else if user.GetGamerType() == RoomUserTypeBlue {
		delete(v.blueMember, user.GetID())
	} else if user.GetGamerType() == RoomUserTypeNormal {
		delete(v.redMember, user.GetID())
		delete(v.blueMember, user.GetID())
	}

	v.scene.TravsalEntity("Player", func(e iserver.IEntity) {
		if e == nil {
			return
		}

		if roomUser, ok := e.(*RoomUser); ok {
			if roomUser.GetGamerType() == user.GetGamerType() {
				roomUser.RPC(iserver.ServerTypeClient, "ClearDeadPos", user.GetID())
			}
		}
	})
}

// clearRedAndBluePos 通知客户端清理红、蓝特效和位置
func (v *VersusScene) clearRedAndBluePos() {
	if !v.showRedAndBluePos() {
		v.scene.TravsalEntity("Player", func(e iserver.IEntity) {
			if e == nil {
				return
			}

			if user, ok := e.(*RoomUser); ok {
				if user.GetGamerType() != RoomUserTypeNormal {
					v.broadVersusState(user, RoomUserTypeNormal)
				}
				user.SetGamerType(RoomUserTypeNormal)
				user.RPC(iserver.ServerTypeClient, "ClearRedAndBluePos")
			}
		})

		v.scene.TravsalEntity("AI", func(e iserver.IEntity) {
			if e == nil {
				return
			}

			if ai, ok := e.(*RoomAI); ok {
				if ai.GetGamerType() != RoomUserTypeNormal {
					v.broadVersusState(ai, RoomUserTypeNormal)
				}
				ai.SetGamerType(RoomUserTypeNormal)
			}
		})
	}
}

// freshAttackState 刷新玩家之间的攻击状态
func (v *VersusScene) freshAttackState(now int64) {
	if !v.initRedAndBlue || v.redOrBlueAttack {
		return
	}

	d, ok := excel.GetSystem(common.System_RedAndBlueAttackOpen)
	if !ok {
		return
	}
	if now >= v.scene.parachuteTime+int64(d.Value) {
		v.redOrBlueAttack = true
	}

	if len(v.redMember) == 0 || len(v.blueMember) == 0 {
		v.redOrBlueAttack = true
	}

	if v.redOrBlueAttack {
		v.clearRedAndBluePos()

		log.Debug("redMember:", len(v.redMember), " blueMember:", len(v.blueMember))
		if errorcode, ok := excel.GetErrorcode(2003); ok {
			v.scene.chatNotify(1, errorcode.Content)
		}
	}
}

func (v *VersusScene) refreshSpecialBox(refreshcount uint64) {
	tb := v.scene
	base, ok := excel.GetMaprule(refreshcount)
	if !ok {
		return
	}

	for _, v := range v.boxlist.Items {
		tb.RemoveTinyEntity(v.Id)
	}

	msg := &protoMsg.DropBoxPosList{}
	for i := uint64(0); i < base.Elitebox && i < 10; i++ {
		tmp := &protoMsg.DropBoxPos{}
		droppos := tb.GetCircleRamdomIndexPos(i, GetRefreshZoneMgr(tb).nextsafecenter, GetRefreshZoneMgr(tb).nextsaferadius)
		droppos = getXZCanPutHeight(tb, droppos)
		boxid := uint32(common.GetTBSystemValue(common.System_RefreshSuperBox))
		entityid := GetRefreshItemMgr(tb).dropItemByID(boxid, droppos)

		tmp.Id = entityid
		tmp.Pos = &protoMsg.Vector3{
			X: droppos.X,
			Y: droppos.Y,
			Z: droppos.Z,
		}
		msg.Items = append(msg.Items, tmp)
		tb.Debug("生成超级特殊空投", " 坐标", droppos)
	}

	v.boxlist = msg
	tb.TravsalEntity("Player", func(e iserver.IEntity) {
		if e == nil {
			return
		}

		if user, ok := e.(*RoomUser); ok {
			user.RPC(iserver.ServerTypeClient, "DropBoxPosList", msg)
		}
	})
}

func (v *VersusScene) resendData(user *RoomUser) {
	v.SceneData.resendData(user)

	user.RPC(iserver.ServerTypeClient, "DropBoxPosList", v.boxlist)
}

// updateTotalNum 通知本局总人数
func (v *VersusScene) updateTotalNum(user *RoomUser) {
	if !v.initRedAndBlue {
		return
	}

	log.Debug("redTotalNum:", v.redTotalNum, " blueTotalNum:", v.blueTotalNum)

	user.RPC(iserver.ServerTypeClient, "UpdateVersusTotalNum", v.redTotalNum, v.blueTotalNum)
}

// updateAliveNum 通知本局存活人数
func (v *VersusScene) updateAliveNum(user *RoomUser) {
	if !v.initRedAndBlue {
		return
	}
	log.Debug("redMember:", uint32(len(v.redMember)), " blueMember:", uint32(len(v.blueMember)))

	user.RPC(iserver.ServerTypeClient, "UpdateVersusAliveNum", uint32(len(v.redMember)), uint32(len(v.blueMember)))
}

// clearSuperBox 当空投被捡完清空小地图中的标识
func (v *VersusScene) clearSuperBox(id uint64) {
	log.Debug("clearSuperBox id:", id, " v.boxlist:", v.boxlist)

	msg := &protoMsg.DropBoxPosList{}
	for _, vBox := range v.boxlist.Items {
		if vBox == nil {
			continue
		}

		if vBox.Id != id {
			msg.Items = append(msg.Items, vBox)
		}
	}
	v.boxlist = msg

	v.scene.TravsalEntity("Player", func(e iserver.IEntity) {
		if e == nil {
			return
		}

		if user, ok := e.(*RoomUser); ok {
			user.RPC(iserver.ServerTypeClient, "ClearSuperBox", id)
		}
	})
}

func (v *VersusScene) broadVersusState(user iserver.ICoordEntity, state int) {
	v.scene.TravsalAOI(user, func(ia iserver.ICoordEntity) {
		if ise, ok := ia.(iserver.IEntityStateGetter); ok {
			if ise.GetEntityState() != iserver.Entity_State_Loop {
				return
			}
			if ie, ok := ia.(iserver.IEntity); ok {
				ie.RPC(iserver.ServerTypeClient, "PlayerChangeGamerType", user.GetID(), uint32(state))
			}
		}
	})
}

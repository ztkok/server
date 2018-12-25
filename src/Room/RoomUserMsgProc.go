package main

import (
	"common"
	"db"
	"excel"
	"fmt"
	"protoMsg"
	"strings"
	"time"
	"zeus/iserver"
	"zeus/linmath"
	"zeus/msgdef"
)

// RoomUserMsgProc RoomUser的消息处理函数
type RoomUserMsgProc struct {
	user *RoomUser
}

// RPC_PickupItem C->S 捡道具
func (p *RoomUserMsgProc) RPC_PickupItem(id uint64) {
	space := p.user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	item, ok := space.GetTinyEntity(id).(*SpaceItem)
	if !ok {
		p.user.Warn("Item id is not ok, id: ", id)
		return
	}

	base, ok := excel.GetItem(uint64(item.item.GetBaseID()))
	if !ok {
		p.user.Warn("Item id not exist, id: ", id)
		return
	}

	//人质, 车屏蔽，雾隐技能烟雾
	if base.Id == 1503 || base.Type == ItemTypeCar || base.Id == 2105 {
		p.user.Warn("PickupItem failed due to car")
		return
	}

	if common.Distance(p.user.GetPos(), item.GetPos()) > 3.0 {
		p.user.Warn("PickupItem failed, item is so far, distance: ", common.Distance(p.user.GetPos(), item.GetPos()))
		return
	}

	canreform := p.user.autoreform && base.Type == ItemTypeWeaponReform && p.user.canAutoReformAll(uint32(base.Id))
	if base.Addpack != 0 && p.user.LeftCell() == 0 && !p.user.CanAddFullPack(base, item.item.count) && !canreform {
		p.user.SendChat("包裹已满")
		p.user.Warn("PickupItem failed, package is full")
		return
	}

	if base.Type == ItemTypePack && uint32(len(p.user.items)) > p.user.initCells+uint32(base.Addcell) {
		p.user.Warn("PickupItem failed, package space is not enough")
		return
	}

	if canreform {
		p.user.autoReform(uint32(base.Id))
	} else {
		if p.user.AddItem(item.item) == 0 {
			return
		}
	}

	if err := space.RemoveTinyEntity(item.GetID()); err != nil {
		p.user.Error("PickupItem failed, RemoveTinyEntity err: ", err)
	}
	p.user.RPC(iserver.ServerTypeClient, "PickupItem", id, uint32(base.Id))
	p.user.Debug("Pick up item success: itemid: ", id)
}

// MsgProc_AttackReq C->S 攻击
func (p *RoomUserMsgProc) MsgProc_AttackReq(content interface{}) {
	msg := content.(*protoMsg.AttackReq)
	if msg == nil {
		return
	}

	if !p.user.StateM.CanAttack() {
		return
	}

	if p.user.checkAttackCold() {
		return
	}

	if p.user.checkFireReplay(msg) {
		//p.user.Info("受到包重放攻击，攻击者ID：", p.user.GetID())
		return
	}

	//p.user.Debug(p.user.GetID(), "攻击", msg.Defendid, msg.Ishead)
	if !p.user.CanAttackAndConsume(msg.Firetime) {
		// p.user.Debug("没有子弹了", p.user.GetID())
		return
	}

	space := p.user.GetSpace().(*Scene)
	if !space.canAttack(p.user, msg.Defendid) {
		return
	}

	// 打断救援
	p.user.stateMgr.BreakRescue()

	if !p.user.isMeleeWeaponUse() {
		msg.Distance = p.user.GetWeaponDistance()
		msg.Attackid = p.user.GetID()
		p.user.CastMsgToAllClientExceptMe(msg)
		// p.user.CastRPCToAllClientExceptMe("AttackReq", msg)
	}

	color := common.GetPlayerNameColor(p.user.GetDBID())

	// 召唤空投
	if p.user.GetInUseWeapon() == common.Item_DropBoxGun {
		GetRefreshItemMgr(space).callUpDropBox(p.user.GetPos(), p.user.GetName(), color)
		p.user.sumData.signalGunUseNum++
		return
	}

	// 召唤空袭
	if p.user.GetInUseWeapon() == common.Item_BombAreaGun {
		GetRefreshZoneMgr(space).callUpBombArea(p.user.GetName(), color)
		p.user.sumData.signalGunUseNum++
		return
	}

	// 召唤坦克0
	if p.user.GetInUseWeapon() == common.Item_DropTank {
		itemId := uint32(common.GetTBSystemValue(common.System_CallBoxItem))
		GetRefreshItemMgr(space).callUpDropTank(p.user.GetPos(), p.user.GetName(), color, itemId)
		p.user.sumData.signalGunUseNum++
		return
	}

	// 召唤坦克1
	if p.user.GetInUseWeapon() == common.Item_DropTank1 {
		itemId := uint32(common.GetTBSystemValue(common.System_CallBoxItem1))
		GetRefreshItemMgr(space).callUpDropTank(p.user.GetPos(), p.user.GetName(), color, itemId)
		p.user.sumData.signalGunUseNum++
		return
	}

	// 坦克攻击
	if p.user.isInTank() {
		p.user.tankShellExplode(msg)
		return
	}

	// 火箭筒攻击
	if p.user.isBazookaWeaponUse() {
		p.user.shellExplodeDamage(msg, p.user.GetInUseWeapon(), gunAttack, 2)
		return
	}

	if msg.Defendid == 0 {
		return
	}

	defender, ok := p.user.GetSpace().GetEntity(msg.Defendid).(iDefender)
	if !ok {
		// p.user.Debug("寻找被攻击者", msg.Defendid, "失败", p.user)

		p.user.attackVehicle(msg)
		return
	}

	//散弹枪打中队友，至濒死状态停止
	if defenderUser, ok := defender.(*RoomUser); ok {
		if (defenderUser.shotGunToWillDie == 0 || defenderUser.shotGunToWillDie == msg.Firetime) &&
			len(p.user.shotgunbullet) > 0 && p.user.GetUserTeamID() != 0 &&
			defender.GetState() == RoomPlayerBaseState_WillDie &&
			p.user.GetUserTeamID() == defender.GetUserTeamID() {
			defenderUser.shotGunToWillDie = msg.Firetime
			//log.Debug("散弹枪打中队友 ", msg.Defendid, " 失败 ", p.user)
			return
		}
	}

	if p.user.isMeleeWeaponUse() {
		if common.Distance(p.user.GetPos(), defender.GetPos()) > 2.0 {
			// p.user.Info("拳头攻击距离过远:", common.Distance(p.user.GetPos(), defender.GetPos()))
			return
		}
	} else {
		if msg.Origion == nil || msg.Dir == nil {
			// p.user.Warn("客户端发送错误空数据")
			return
		}

		orgion := linmath.NewVector3(msg.Origion.X, msg.Origion.Y, msg.Origion.Z)
		dir := linmath.NewVector3(msg.Dir.X, msg.Dir.Y, msg.Dir.Z)
		dir.Normalize()

		if p.user.GetPos().Sub(orgion).Len() > 5 {
			//		if common.Distance(p.user.GetPos(), orgion) > 5 {
			//p.user.Debug("攻击发射点位置异常:", p.user.GetPos(), " 发射点:", orgion, " 距离:", p.user.GetPos().Sub(orgion).Len())
			return
		}

		// 检测障碍物
		rayDistance := orgion.Sub(defender.GetPos()).Len()
		_, _, _, hit, err := space.Raycast(orgion, dir, rayDistance, unityLayerGround|unityLayerBuilding|unityLayerFurniture)
		if hit {
			p.user.Error("Raycast err: ", err, " distance: ", rayDistance, " orgion:", orgion, " dir:", dir, " defender:")
			return
		}

		// 射线检测
		distance, canattack, err := space.SphereRaycast(defender.GetPos(), 3.0, orgion, dir, p.user.GetWeaponDistance())
		if !canattack {
			p.user.Error("SphereRaycast err: ", err, " distance: ", distance, " gun distance: ", p.user.GetWeaponDistance(), " orgion:", orgion, " dir:", dir, " defender:", defender.GetPos())
			return
		}
	}

	if msg.AttackPos == AttackPos_Head && p.user.checkHeadShot(defender.GetID()) {
		//p.user.Debug(p.user.GetID(), " 单位时间内枪枪爆头异常")
		msg.AttackPos = AttackPos_Body
	}

	AttackHandle(p.user, defender, msg.AttackPos)
	//p.user.Info("Request to attack success")
}

// MsgProc_ShootReq 开枪
func (p *RoomUserMsgProc) RPC_ShootReq(msg *protoMsg.ShootReq) {
	if !p.user.CanFakeAttackAndConsume() {
		//p.user.Warn("攻击动画没有子弹了", p.user.GetID())
		return
	}

	p.user.CastRPCToAllClientExceptMe("ShootReq", msg)
}

func (p *RoomUserMsgProc) RPC_Melee(left uint64) {
	p.user.CastRPCToAllClientExceptMe("Melee", left)
}

func (p *RoomUserMsgProc) RPC_AimPos(aimPos *protoMsg.Vector3) {
	p.user.SetAimPos(aimPos)
}

func (p *RoomUserMsgProc) RPC_HeadTilt(param float32) {
	p.user.CastRPCToAllClientExceptMe("HeadTilt", param)
}

func (p *RoomUserMsgProc) RPC_ExchangeGunReq(useweapon uint64) {
	state := p.user.GetBaseState()
	if state != RoomPlayerBaseState_Stand && state != RoomPlayerBaseState_Down && state != RoomPlayerBaseState_Crouch &&
		state != RoomPlayerBaseState_Fall && state != RoomPlayerBaseState_Jump {
		return
	}

	p.user.ExchangeGun(useweapon)

	// 打断救援
	p.user.stateMgr.BreakRescue()
}

func (p *RoomUserMsgProc) MsgProc_UseObjectReq(content interface{}) {
	msg := content.(*protoMsg.UseObjectReq)

	p.user.UseObject(msg.Thisid)
}

// RPC_GunReformEquipReq 装备枪支配件请求
func (p *RoomUserMsgProc) RPC_GunReformEquipReq(gunthisid, reformthisid uint64, pos uint8) {
	p.user.reformGun(gunthisid, reformthisid, pos)
}

func (p *RoomUserMsgProc) MsgProc_GunReformUnequipReq(content interface{}) {
	msg := content.(*protoMsg.GunReformUnequipReq)
	// p.user.Info("拆掉改装枪支装备", msg.Gunthisid)

	v, ok := p.user.equips[msg.Gunthisid]
	if !ok {
		return
	}

	var needcell uint32 = 1
	if v.isAddBullet(p.user, msg.Baseid) {
		needcell++
	}

	if p.user.LeftCell() < needcell {
		p.user.AdviceNotify(common.NotifyCommon, common.ErrCodeIDPackCell)
		return
	}

	p.user.unequipReform(msg.Baseid, msg.Gunthisid, false)
	p.user.tlogBattleFlow(behavetype_dismount, 0, uint64(msg.Baseid), 0, 0, uint32(len(p.user.items)))
}

func (p *RoomUserMsgProc) RPC_GunReformUnequipDrop(gunthisid uint64, baseid uint32) {
	_, ok := p.user.equips[gunthisid]
	if !ok {
		return
	}

	p.user.unequipReform(baseid, gunthisid, true)
	p.user.tlogBattleFlow(behavetype_dismount, 0, uint64(baseid), 0, 0, uint32(len(p.user.items)))
}

// RPC_SwitchGunSightReq 请求切换倍镜
func (p *RoomUserMsgProc) RPC_SwitchGunSightReq(gunthisid uint64, pos1, pos2 uint8, para uint8) {
	p.user.switchGunSight(gunthisid, int(pos1), int(pos2), para)
}

func (p *RoomUserMsgProc) MsgProc_ChangeBulletReq(content interface{}) {
	msg := content.(*protoMsg.ChangeBulletReq)
	p.user.changeBullet(msg.Full)
}

func (p *RoomUserMsgProc) RPC_CheckBox(id uint64) {
	// p.user.Info("rpc, 查看道具", id)
	space := p.user.GetSpace().(*Scene)
	sceneitem, ok := space.GetTinyEntity(id).(*SpaceItem)
	if !ok {
		//p.user.Warn("查看补给箱id异常", id)
		return
	}

	if len(sceneitem.haveobjs) == 0 {
		//p.user.Warn("查看补给箱内没有道具", sceneitem.item.GetBaseID())
		return
	}

	proto := &protoMsg.RefreshBoxObjNotify{
		Id: id,
	}
	for index, item := range sceneitem.haveobjs {
		if item == nil {
			//p.user.Warn("补给箱错误道具")
			continue
		}
		iteminfo := &protoMsg.ItemProp{}
		iteminfo.Id = uint64(index)
		iteminfo.Baseid = item.GetBaseID()
		iteminfo.Num = item.count
		proto.Data = append(proto.Data, iteminfo)
	}
	p.user.RPC(iserver.ServerTypeClient, "RefreshBoxObjNotify", proto)
	/*
		p.user.CastMsgToMe(proto)
	*/
	// p.user.Info("rpc, 返回查看补给箱道具", proto)
}

func (p *RoomUserMsgProc) RPC_GetBoxItem(id uint64, boxitem uint64) {
	// p.user.Debug("rpc, 领取补给箱内部道具", id, boxitem)
	space := p.user.GetSpace().(*Scene)
	sceneitem, ok := space.GetTinyEntity(id).(*SpaceItem)
	if !ok {
		p.user.Debug("查看补给箱id异常", id)
		return
	}

	baseid := sceneitem.item.GetBaseID()
	if sceneitem.item.base.Type == ItemTypeBox && sceneitem.item.base.Subtype == ItemSubtypeSupply && sceneitem.cangettime != 0 && time.Now().Unix() < sceneitem.cangettime {
		p.user.Debug("领取补给箱内道具，补给箱在坠落过程中")
		return
	}

	if item, ok := sceneitem.haveobjs[uint32(boxitem)]; ok {
		p.user.removeKillDrop(id)

		canreform := p.user.autoreform && item.base.Type == ItemTypeWeaponReform && p.user.canAutoReformAll(uint32(item.base.Id))
		if item.base.Addpack != 0 && p.user.LeftCell() == 0 && !p.user.CanAddFullPack(item.base, item.count) && !canreform {
			p.user.SendChat("包裹已满")
			return
		}

		if canreform {
			p.user.autoReform(item.GetBaseID())
		} else {
			if p.user.AddItem(item) == 0 {
				return
			}
		}

		delete(sceneitem.haveobjs, uint32(boxitem))

		if sceneitem.item.base.Type == ItemTypeBox && sceneitem.item.base.Subtype == ItemSubtypeSupply {
			p.user.sumData.getDropBoxs[id] = true
		}

		p.user.RPC(iserver.ServerTypeClient, "GetBoxItemVoice", id, boxitem, uint32(item.base.Id))

		space.TravsalEntity("Player", func(e iserver.IEntity) {
			if e == nil {
				return
			}

			if user, ok := e.(*RoomUser); ok {
				user.RPC(iserver.ServerTypeClient, "GetBoxItem", id, boxitem, uint32(item.base.Id))
			}
		})

		if len(sceneitem.haveobjs) == 0 {

			boxid := uint32(common.GetTBSystemValue(common.System_RefreshSpecialBox))
			if baseid != boxid {
				if err := space.RemoveTinyEntity(sceneitem.GetID()); err != nil {
					p.user.Error(err, p.user, sceneitem)
				}
			}
			superBoxId := uint32(common.GetTBSystemValue(common.System_RefreshSuperBox))
			if baseid == superBoxId {
				space.ISceneData.clearSuperBox(id)
			}

			p.user.RPC(iserver.ServerTypeClient, "PickupItem", id, baseid)

			space.TravsalEntity("Player", func(e iserver.IEntity) {
				if e == nil {
					return
				}

				if user, ok := e.(*RoomUser); ok {
					user.RPC(iserver.ServerTypeClient, "ClearBoxNotify", id)
				}
			})

			box, ok := GetRefreshItemMgr(space).boxlist[id]
			if ok {
				box.Havepick = true
			}
		}
	}
}

func (p *RoomUserMsgProc) RPC_RemoveObject(thisid uint64, dropSum uint32) {
	// p.user.Info("掉落道具", thisid)
	if dropSum == 0 {
		return
	}

	if !p.user.CanDrop() {
		return
	}

	p.user.BreakEffect(true)
	p.user.dropObj(thisid, dropSum)
}

func (p *RoomUserMsgProc) RPC_DropGun(thisid uint64) {
	// p.user.Info("丢弃武器", thisid)
	if !p.user.CanDrop() {
		return
	}

	p.user.dropGun(thisid)
}

func (p *RoomUserMsgProc) RPC_DropArmor(baseid uint64) {
	// p.user.Info("丢弃防具", baseid)
	if !p.user.CanDrop() {
		return
	}

	p.user.dropArmor(baseid)
	var level uint32 = 0
	if baseid == 1201 || baseid == 1202 || baseid == 1601 {
		level = 1
	} else if baseid == 1203 || baseid == 1204 || baseid == 1602 {
		level = 2
	} else if baseid == 1205 || baseid == 1206 {
		level = 3
	}
	p.user.tlogBattleFlow(behavetype_throw, 0, baseid, 0, level, uint32(len(p.user.items)))
}

// func (p *RoomUserMsgProc) RPC_ReleaseParachute() {
// 	// p.user.Info("客户端开伞", p.user)
// 	p.user.SetBaseState(RoomPlayerBaseState_Parachute)
// }

func (p *RoomUserMsgProc) RPC_SetMvSpeed(speed float32) {
	p.user.SetMvSpeed(speed)
}

func (p *RoomUserMsgProc) RPC_upVehicle(thisid uint64) {
	if !p.user.upVehicle(thisid) {
		p.user.CastRPCToMe("UpVehicleFail", thisid)
	}
}

func (p *RoomUserMsgProc) RPC_compilotUp(pilotid uint64) {
	p.user.compilotUp(pilotid)
}

func (p *RoomUserMsgProc) RPC_downVehicle(pos *protoMsg.Vector3, prop *protoMsg.VehiclePhysics, dir *protoMsg.Vector3) {
	if pos == nil || prop == nil || dir == nil {
		return
	}

	downpos := linmath.NewVector3(pos.X, pos.Y, pos.Z)
	downdir := linmath.NewVector3(dir.X, dir.Y, dir.Z)

	if common.Distance(p.user.GetPos(), downpos) > 6.0 {
		p.user.Debug("downVehicle failed, distance is so far, distance: ", common.Distance(p.user.GetPos(), downpos))
		return
	}

	p.user.downVehicle(downpos, downdir, prop, true)
}

func (p *RoomUserMsgProc) RPC_compilotDown(pos *protoMsg.Vector3, dir *protoMsg.Vector3, speed float64) {
	if pos == nil {
		return
	}

	downpos := linmath.NewVector3(pos.X, pos.Y, pos.Z)
	downdir := linmath.NewVector3(dir.X, dir.Y, dir.Z)

	if common.Distance(p.user.GetPos(), downpos) > 12.0 {
		p.user.Warn("compilotDown failed, distance is so far, distance: ", common.Distance(p.user.GetPos(), downpos))
		return
	}
	if p.user.stateSyncTime+7 < p.user.GetSpace().GetTimeStamp() {
		speed = 0
	}

	p.user.compilotDown(downpos, downdir, float32(speed))
}

func (p *RoomUserMsgProc) RPC_SyncVehiclePos(thisid uint64, pos, dir *protoMsg.Vector3, stop bool) {
	if pos == nil || dir == nil {
		p.user.Error("SyncVehiclePos failed, pos or dir is nil")
		return
	}

	space := p.user.GetSpace().(*Scene)
	sitem, ok := space.mapitem[thisid]
	if !ok {
		p.user.Error("SyncVehiclePos failed, thisid: ", thisid)
		return
	}

	base := sitem.item.base
	if base.Type != ItemTypeCar {
		p.user.Error("SyncVehiclePos failed, not ItemTypeCar, thisid: ", thisid)
		return
	}

	item, ok := space.GetEntity(thisid).(*SpaceVehicle)
	if !ok {
		p.user.Error("SyncVehiclePos failed, thisid: ", thisid)
		return
	}

	item.SetPos(linmath.NewVector3(pos.X, pos.Y, pos.Z))
	item.SetRota(linmath.NewVector3(dir.X, dir.Y, dir.Z))

	prop := item.GetVehicleProp()
	if prop.Reducedam == 0 {
		return
	}

	waterLevel := space.mapdata.Water_height
	isWater, err := space.IsWater(pos.X, pos.Z, waterLevel)
	if err != nil {
		p.user.Error("IsWater err: ", err)
		return
	}

	var broken bool

	//船驶到岸上
	if !isWater && base.Subtype == 1 {
		broken = true
	}

	//车辆驶入水中
	if isWater && base.Subtype == 0 && !p.user.isOnBridge(linmath.NewVector3(pos.X, pos.Y, pos.Z)) {
		broken = true
	}

	//将Reducedam置为0
	if broken {
		prop.Reducedam = 0
		item.SetVehicleProp(prop)
		item.SetVehiclePropDirty()
		p.user.Info("Vehicle enter into water or boat ashore, reducedam is clear, type: ", base.Subtype, " thisid: ", thisid)
	}
}

func (p *RoomUserMsgProc) RPC_SyncVehiclePhysics(thisid uint64, prop *protoMsg.VehiclePhysics) {
	if prop == nil || prop.Position == nil || prop.Rotation == nil {
		return
	}

	space := p.user.GetSpace().(*Scene)
	car, ok := space.cars[thisid]
	if !ok {
		p.user.Error("SyncVehiclePhysics failed, thisid: ", thisid)
		return
	}

	car.physics = prop
	// p.user.Error("同步车物理信息", prop.Position, prop.Rotation, prop)
}

// RPC_TankShootReady 玩家上坦克后，客户端完成初始化
func (proc *RoomUserMsgProc) RPC_TankShootReady() {
	if proc.user.isInTank() {
		proc.user.RPC(iserver.ServerTypeClient, "TankShootReady")
	}
}

// RPC_TankSubRotationSync 同步坦克炮台角度和炮管角度
func (proc *RoomUserMsgProc) RPC_TankSubRotationSync(rota1, rota2 *protoMsg.Vector3) {
	proc.user.SetSubRotation1(rota1)
	proc.user.SetSubRotation1Dirty()
	proc.user.SetSubRotation2(rota2)
	proc.user.SetSubRotation2Dirty()
}

// RPC_RescueTeammate 救援队友
func (proc *RoomUserMsgProc) RPC_RescueTeammate(targetid uint64) {
	// p.user.Info("救援队友")
	space := proc.user.GetSpace().(*Scene)
	targetuser, ok := space.GetEntity(targetid).(*RoomUser)
	if !ok {
		proc.user.Warn("Teammate not exist, targetid: ", targetid)
		return
	}

	// 距离验证
	distance := proc.user.GetPos().Sub(targetuser.GetPos()).Len()
	systemValue := float32(common.GetTBSystemValue(37)) / 100
	if distance > systemValue {
		proc.user.SetActionState(0)
		proc.user.Warn("Distance is so far, distance: ", distance)
		return
	}

	baseState := proc.user.GetBaseState()
	if baseState != RoomPlayerBaseState_Stand && baseState != RoomPlayerBaseState_Crouch {
		proc.user.Warn("baseState err:", baseState)
		return
	}
	actionState := proc.user.GetActionState()
	if actionState == Potion {
		proc.user.Warn("actionState err:", actionState)
		return
	}

	proc.user.stateMgr.setRescue(targetuser)
	proc.user.Info("Rescue teammate success, targetid: ", targetid)
}

func (proc *RoomUserMsgProc) RPC_leaveSpace() {
	proc.user.Info("Room user leave space by iteself")

	space := proc.user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	if proc.user.userType == RoomUserTypeWatcher {
		proc.user.LeaveSpace()
		return
	}

	if proc.user.stateMgr.GetState() == RoomPlayerBaseState_WillDie {
		// 终止玩家濒死状态
		proc.user.DisposeDeath(proc.user.stateMgr.injuredtype, proc.user.stateMgr.attactID, false)
		attacker, ok := space.GetEntity(proc.user.stateMgr.attactID).(*RoomUser)
		if ok {
			if !(space.teamMgr.IsInOneTeam(proc.user.GetDBID(), attacker.GetDBID()) || proc.user.GetID() == proc.user.stateMgr.attactID) {
				attacker.DisposeIncrKillNum()
				if proc.user.stateMgr.isHeadShot && proc.user.stateMgr.attactID == proc.user.stateMgr.downAttacker {
					attacker.IncrHeadShotNum()
				}
			}
		}
		space.BroadDieNotify(proc.user.stateMgr.attactID, proc.user.GetID(), false, InjuredInfo{injuredType: proc.user.stateMgr.injuredtype, isHeadshot: false})
	}

	if proc.user.sumData.deadType == 0 {
		proc.user.sumData.deadType = 10
	}

	proc.user.userLeave()

	if space.teamMgr.isTeam {
		teammate := space.teamMgr.getAliveTeammate(proc.user)
		if teammate == 0 {
			space.teamMgr.DisposeTeamSettle(proc.user.GetUserTeamID(), false)
		}
	}

	//离开地图
	proc.user.LeaveScene()
}

// RPC_SpaceLoadCompleteByTeam 客户端场景组队相关加载完成
func (proc *RoomUserMsgProc) RPC_SpaceLoadCompleteByTeam() {
	proc.user.Info("SpaceLoadCompleteByTeam")

	space := proc.user.GetSpace().(*Scene)
	strID := fmt.Sprintf("%d", space.GetID())
	proc.user.RPC(iserver.ServerTypeClient, "SyncSpaceType", uint32(space.GetSpaceType()), strID) // 1.组队场景 2.非组队场景

	if proc.user.userType == RoomUserTypeWatcher {
		proc.user.doLoadingDone(true)
	}
}

// RPC_MapSign 总览地图标记
func (p *RoomUserMsgProc) RPC_MapSign(x float64, y float64) {
	p.user.Info("Map sign, x: ", x, " y: ", y)
	if p.user.GetUserTeamID() == 0 {
		//单人标记
		p.user.signPos.X = float32(x)
		p.user.signPos.Y = float32(y)
		p.user.BroadcastToWatchers(RoomUserTypeWatcher, "SyncMapSign", p.user.GetID(), x, y)
		return
	}
	space := p.user.GetSpace().(*Scene)
	space.teamMgr.SyncMapSign(p.user, x, y)
}

// RPC_CancelMapSign 取消地图标记
func (p *RoomUserMsgProc) RPC_CancelMapSign() {
	p.user.Info("Cancel map sign")
	if p.user.GetUserTeamID() == 0 {
		//单人取消标记
		p.user.signPos.X = 0
		p.user.signPos.Y = 0
		p.user.BroadcastToWatchers(RoomUserTypeWatcher, "CancelSyncMapSign", p.user.GetID())
		return
	}
	space := p.user.GetSpace().(*Scene)
	space.teamMgr.CancelSyncMapSign(p.user)
}

// RPC_TeamDirectReq 请求广播指挥指令
func (p *RoomUserMsgProc) RPC_TeamDirectReq(id uint8, dir uint16) {
	p.user.Debug("Team Direct, id: ", id, " dir: ", dir)
	space := p.user.GetSpace().(*Scene)
	space.teamMgr.TeamDirectBroad(p.user, id, dir)
}

// RPC_VehicleFuelLevel 客户端向服务器发送车辆的油门档位
func (p *RoomUserMsgProc) RPC_VehicleFuelLevel(level float32) {
	if !(level >= 0 && level <= 1.0) {
		return
	}

	prop := p.user.GetVehicleProp()
	if prop == nil {
		p.user.Error("Vehicle prop is nil")
		return
	}

	if prop.GetPilotID() != p.user.GetID() {
		p.user.Error("Only calculate pilot's fuel consumption")
		return
	}

	carrierConfig, ok := excel.GetCarrier(prop.GetVehicleID())
	if !ok {
		p.user.Error("Can't get carrier config, baseid: ", prop.GetVehicleID())
		return
	}

	space := p.user.GetSpace().(*Scene)
	car, ok := space.cars[prop.GetThisid()]
	if !ok {
		p.user.Error("Car not exist, thisid: ", prop.GetThisid())
		return
	}

	minLevel := float32(common.GetTBSystemValue(169)) / 100.0

	if level < minLevel {
		if common.Distance(p.user.GetPos(), car.lastConsumePos) <= 0.15 {
			return
		}
		level = minLevel
	} else {
		strs := strings.Split(carrierConfig.Factor, "|")
		for _, str := range strs {
			ss := strings.Split(str, "-")
			if len(ss) != 3 {
				continue
			}

			//计算车辆耗油
			if level > float32(common.StringToInt(ss[0]))/100.0 && level <= float32(common.StringToInt(ss[1]))/100.0 {
				level = float32(common.StringToUint32(ss[2])) / 100.0
				break
			}
		}
	}

	per := p.user.SkillData.getReduceTypeDam(SE_OilLoss) //开车油耗减少百分比
	prop.FuelLeft -= carrierConfig.Consumption * level * per
	if prop.FuelLeft < 0 {
		prop.FuelLeft = 0
	}

	car.VehicleProp = prop
	car.RefreshInfo()
	car.lastConsumePos = p.user.GetPos()
}

// RPC_VehicleAddFuel 玩家给车辆加燃料
func (p *RoomUserMsgProc) RPC_VehicleAddFuel(thisid uint64) {
	//人在车上才能加油
	if p.user.GetBaseState() != RoomPlayerBaseState_Ride && p.user.GetBaseState() != RoomPlayerBaseState_Passenger {
		p.user.Error("User is not in a car")
		return
	}

	//不能重复加油
	if p.user.IsInAddFuelState() {
		p.user.Error("Can't add fuel to car again")
		return
	}

	prop := p.user.GetVehicleProp()
	if prop == nil {
		p.user.Error("Vehicle prop is nil")
		return
	}

	if prop.GetFuelLeft() == prop.GetFuelMax() {
		p.user.Warn("The car has full fuel")
		return
	}

	space := p.user.GetSpace().(*Scene)
	_, ok := space.cars[prop.GetThisid()]
	if !ok {
		p.user.Error("Car not exist, id: ", prop.GetThisid())
		return
	}

	item := p.user.GetItemByThisid(thisid)
	if item == nil {
		p.user.Error("User doesn't own this item, thisid: ", thisid)
		return
	}

	base, ok := excel.GetItem(uint64(item.GetBaseID()))
	if !ok {
		p.user.Error("Item id not exist, baseid: ", item.GetBaseID())
		return
	}

	if err := p.user.RPC(iserver.ServerTypeClient, "StartAddFuel", uint32(base.Id)); err != nil {
		p.user.Error("RPC StartAddFuel err: ", err)
		return
	}

	//调度定时器，开启加油过程
	p.user.addFuelTimer = p.user.GetEntities().AddDelayCall(p.user.SucceedAddFuel, time.Duration(base.Processtime)*time.Second)
	p.user.addFuelThisid = thisid

	p.user.Info("Start add fuel to the car, baseid: ", base.Id)
}

func (p *RoomUserMsgProc) RPC_VehicleHitCollision(speed float32) {
	space := p.user.GetSpace().(*Scene)

	if p.user.GetState() != RoomPlayerBaseState_Ride {
		p.user.Warn("User is not riding, state: ", p.user.GetState())
		return
	}

	system, ok := excel.GetSystem(uint64(common.System_VehicleCollisionA))
	if !ok {
		return
	}
	A := float32(system.Value)

	system, ok = excel.GetSystem(uint64(common.System_VehicleCollisionB))
	if !ok {
		return
	}
	B := float32(system.Value)

	system, ok = excel.GetSystem(uint64(common.System_VehicleCollisionC))
	if !ok {
		return
	}
	C := float32(system.Value)

	if speed > B {
		subhp := uint32(A * (speed - B))
		p.user.Debug("Hit enemy success, damage value: ", subhp)

		carid := p.user.GetVehicleProp().Thisid
		if carid != 0 {
			car, ok := space.cars[carid]
			if !ok {
				return
			}

			if car.Reducedam == 0 {
				return
			}

			if car.Reducedam > subhp {
				car.Reducedam -= subhp
			} else {
				car.Reducedam = 0
				car.Haveexplode = true
			}

			car.RefreshInfo()

			if car.Reducedam != 0 {
				subhp = subhp * uint32(C) / 100
				car.SubUserHp(subhp, carcollision)
			} else {
				car.onExplode()
				p.user.vehicleExplodeDamage(car.GetThisid())
			}
		}
	}
}

func (p *RoomUserMsgProc) RPC_CarHitUser(defendid uint64, speed float32) {
	space := p.user.GetSpace().(*Scene)
	defender, ok := space.GetEntity(defendid).(iDefender)
	if !ok {
		p.user.Warn("Defender id not exist, id: ", defendid)
		return
	}
	if space.isEliteScene() {
		if !space.ISceneData.canAttack(p.user, defendid) {
			return
		}
	}
	if space.isVersusScene() {
		if !space.ISceneData.canAttack(p.user, defendid) {
			return
		}
	}
	if p.user.GetID() == defendid {
		p.user.Warn("Car can't hit user self")
		return
	}

	if p.user.GetState() != RoomPlayerBaseState_Ride {
		p.user.Warn("User is not riding, state: ", p.user.GetState())
		return
	}

	if common.Distance(p.user.GetPos(), defender.GetPos()) > 4.0 {
		p.user.Warn("Distance is so far, distance: ", common.Distance(p.user.GetPos(), defender.GetPos()))
		return
	}

	system, ok := excel.GetSystem(uint64(common.System_VehicleHitA))
	if !ok {
		return
	}
	A := float32(system.Value)

	system, ok = excel.GetSystem(uint64(common.System_VehicleHitB))
	if !ok {
		return
	}
	B := float32(system.Value)

	if speed > B {
		subhp := uint32(A * (speed - B))

		if defenderUser, ok := space.GetEntity(defendid).(*RoomUser); ok {
			per := defenderUser.SkillData.getReduceTypeDam(SE_VehicleNoDam)
			subhp = uint32(float32(subhp) * per)
		}

		subhp -= dongReduceDam(defender, subhp)
		if subhp <= 0 {
			return
		}

		defender.DisposeSubHp(InjuredInfo{num: subhp, injuredType: carhit, isHeadshot: false, attackid: p.user.GetID(), attackdbid: p.user.GetDBID()})
		p.user.Debug("Hit enemy success, damage value: ", subhp)
	}
}

// RPC_GmSetLeaveSceneTime 设置结算后离开场景的倒计时时间
func (p *RoomUserMsgProc) RPC_GmSetLeaveSceneTime(t uint32) {
	p.user.leaveSceneTime = int(t)
}

// RPC_ParachuteReady 客户端通知, 场景加载完毕, 可以跳伞
func (p *RoomUserMsgProc) RPC_ParachuteReady() {
	//p.user.Debug("收到客户端加载完成", p.user.GetID(), " state:", p.user.GetBaseState())
	tb := p.user.GetSpace().(*Scene)
	if p.user.GetState() == RoomPlayerBaseState_LoadingMap {
		p.user.SetBaseState(RoomPlayerBaseState_Inplane)

		if !tb.allowParachute {
			tb.checkAllowParachuteState()
		}
	} else if p.user.GetState() == RoomPlayerBaseState_Inplane {
		if tb.allowParachute {
			fliedSeconds := time.Now().Sub(tb.allowParachuteTime).Seconds()
			if err := p.user.RPC(iserver.ServerTypeClient, "ReadyPosForAirline", float64(fliedSeconds)); err != nil {
				p.user.Error(err)
			}
		}
	}
}

func (p *RoomUserMsgProc) RPC_StartParachute() {
	if p.user.GetState() == RoomPlayerBaseState_Inplane {
		if p.user.IsFollowingParachute() {
			return
		}

		p.user.doParachute(false)
		if err := p.user.RPC(iserver.ServerTypeClient, "ParachutePos"); err != nil {
			p.user.Error("RPC ParachutePos err: ", err)
		}

		space := p.user.GetSpace().(*Scene)
		space.BroadAirLeft()

		p.user.Info("Start parachute")
		p.user.tlogBattleFlow(behavetype_jump, 0, 0, 0, 0, 0) // tlog战场流水表(6代表跳伞)
	}
}

// RPC_InviteFollowReq 邀请队友跟随跳伞
func (p *RoomUserMsgProc) RPC_InviteFollowReq() {
	space := p.user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	ret := uint8(0)

	if !space.teamMgr.InviteTeamFollow(p.user) {
		ret = 1
	}

	p.user.RPC(iserver.ServerTypeClient, "InviteFollowRet", ret)
}

// RPC_InviteFollowRespReq 被邀请者回复邀请请求
func (p *RoomUserMsgProc) RPC_InviteFollowRespReq(resp uint8, inviter uint64) {
	if resp == 1 {
		return
	}

	space := p.user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	ret := uint8(0)

	if !space.teamMgr.InviteFollowResp(p.user, inviter) {
		ret = 1
	}

	p.user.RPC(iserver.ServerTypeClient, "InviteFollowRespRet", ret, inviter)
}

// RPC_FollowReq 跟随跳伞请求
func (p *RoomUserMsgProc) RPC_FollowReq(target uint64) {
	space := p.user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	ret := uint8(0)

	if !space.teamMgr.FollowParachute(p.user, target) {
		ret = 1
	}

	p.user.RPC(iserver.ServerTypeClient, "FollowRet", ret)
}

// RPC_CancelFollowReq 取消跟随跳伞请求
func (p *RoomUserMsgProc) RPC_CancelFollowReq() {
	space := p.user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	space.teamMgr.CancelFollowParachute(p.user)
}

// RPC_SetGunSight 设置高倍镜状态
func (p *RoomUserMsgProc) RPC_SetGunSight(state uint64) {
	if p.user.canSetGunSightState() {
		p.user.SetGunSight(state)
	}
}

func (p *RoomUserMsgProc) RPC_onEndFall(speed float64) {
	return

	if !p.user.StateM.CanChangeState() {
		return
	}

	if p.user.GetState() == RoomPlayerBaseState_Ride || p.user.GetState() == RoomPlayerBaseState_Passenger {
		//p.user.Error("驾驶状态不掉落")
		return
	}

	pos := p.user.GetPos()
	waterLevel := p.user.GetSpace().(*Scene).mapdata.Water_height
	isWater, err := p.user.GetSpace().IsWater(pos.X, pos.Z, waterLevel)
	if err != nil {
		return
	}
	if isWater {
		// p.user.Debug("掉落水域不受伤害", p.user.GetID())
		return
	}

	system, ok := excel.GetSystem(uint64(common.System_FallDamageA))
	if !ok {
		return
	}
	A := float64(system.Value)

	system, ok = excel.GetSystem(uint64(common.System_FallDamageB))
	if !ok {
		return
	}
	B := float64(system.Value)

	system, ok = excel.GetSystem(uint64(common.System_FallDamage))
	if !ok {
		return
	}
	damage := float64(system.Value)

	//p.user.Debug("@@@@@@@@@@@坠落", " 参数A:", A, " 参数B:", B, "speed:", speed)
	var subhp uint32
	if B != 0 && speed > A {
		subhp = uint32(damage * (speed - A) / B)
		p.user.Debug("onEndFall A: ", A, " B: ", B, " damage:", damage, " subhp: ", subhp)
	}

	// p.user.DisposeSubHp(InjuredInfo{num: subhp, injuredType: falldam, isHeadshot: false})
}

// RPC_GmDeath 压测相关
func (p *RoomUserMsgProc) RPC_GmDeath(injuredType uint32) {

	p.user.DisposeDeath(injuredType, 0, false)
	p.user.DisposeSettle()

	//p.user.Death(injuredType)
}

// RPC_GmRoom Gm命令
func (p *RoomUserMsgProc) RPC_GmRoom(paras string) {
	p.user.Debug("RPC_GmRoom: ", paras)
	p.user.gm.exec(paras)
}

func (p *RoomUserMsgProc) RPC_ExchangeReformGun(gunid uint64, reformid uint32, othergun uint64) {
	p.user.exchangeReformGun(gunid, reformid, othergun)
}

// RPC_startThrow 玩家扔出投掷道具
func (p *RoomUserMsgProc) RPC_startThrow(pos *protoMsg.Vector3, speed *protoMsg.Vector3) {
	p.user.CastRPCToAllClientExceptMe("startThrow", pos, speed)
}

// RPC_triggerInHand 投掷道具掉在原地
func (p *RoomUserMsgProc) RPC_triggerInHand(pos *protoMsg.Vector3) {
	p.user.CastRPCToAllClientExceptMe("triggerInHand", pos)
}

func (p *RoomUserMsgProc) RPC_throwItemTime(time uint64) {
	p.user.CastRPCToAllClientExceptMe("throwItemTime", time)
}

// RPC_onThrowItem 玩家从背包中拿出投掷道具
func (p *RoomUserMsgProc) RPC_onThrowItem(baseid uint64) {
	base, ok := excel.GetItem(baseid)
	if !ok {
		return
	}

	if base.Type != ItemTypeThrow {
		return
	}

	if p.user.GetItemNum(uint32(baseid)) < 1 {
		return
	}

	// p.user.Debug("投资道具", p.user.GetID())
	p.user.CastRPCToAllClientExceptMe("onThrowItem", baseid)
	p.user.removeItem(uint32(baseid), 1)
	p.user.havethrownum[uint32(baseid)]++
	p.user.stateMgr.BreakRescue()
}

// RPC_onThrowItemGuid 客户端上传投掷道具的唯一标识
func (p *RoomUserMsgProc) RPC_onThrowItemGuid(guid string) {
	p.user.CastRPCToAllClientExceptMe("onThrowItemGuid", guid)
}

// RPC_onThrowTrigger 触发手雷爆炸
func (p *RoomUserMsgProc) RPC_onThrowTrigger(guid string) {
	p.user.CastRPCToAllClientExceptMe("onThrowTrigger", guid)
}

// RPC_onThrowDestroy 销毁失效的手雷
func (p *RoomUserMsgProc) RPC_onThrowDestroy(guid string) {
	p.user.CastRPCToAllClientExceptMe("onThrowDestroy", guid)
}

// RPC_onThrowDamage 客户端上传手雷的爆炸数据
func (p *RoomUserMsgProc) RPC_onThrowDamage(info *protoMsg.ThrowDamageInfo) {
	if info == nil || info.Center == nil {
		return
	}

	thrownum, ok := p.user.havethrownum[info.Baseid]
	if !ok || thrownum < 1 {
		p.user.Warn("Bomb not exist, id: ", info.Baseid)
		return
	}

	p.user.havethrownum[info.Baseid]--

	center := linmath.NewVector3(info.Center.X, info.Center.Y, info.Center.Z)
	if common.Distance(p.user.GetPos(), center) > 100 {
		p.user.Warn("Obj is too far, distance: ", common.Distance(p.user.GetPos(), center))
		return
	}

	base, baseok := excel.GetItem(uint64(info.Baseid))
	if !baseok || base.ThrowHurtRadius == 0 {
		p.user.Warn("Base id is not ok, id: ", info.Baseid)
		return
	}

	for i, j := range info.Defends {
		if j.Id == p.user.GetID() && i != (len(info.Defends)-1) {
			tmp := j
			info.Defends[i] = info.Defends[len(info.Defends)-1]
			info.Defends[len(info.Defends)-1] = tmp
			break
		}
	}
	space := p.user.GetSpace().(*Scene)
	if space == nil {
		p.user.Error("SetDoorState failed, can't get space")
	}
	for _, v := range info.Defends {
		defender, ok := p.user.GetSpace().GetEntity(v.Id).(iDefender)
		if !ok {
			p.user.Warn("Get defender failed, id: ", v.Id)
			p.user.explodeDamageVehicle(v.Id, v.Dam, center, base)
			continue
		}

		if defender.isInTank() {
			continue
		}

		if space.isEliteScene() {
			if v.Id != p.user.GetID() && !space.ISceneData.canAttack(p.user, v.Id) {
				continue
			}
		}

		if space.isVersusScene() {
			if v.Id != p.user.GetID() && !space.ISceneData.canAttack(p.user, v.Id) {
				continue
			}
		}

		s := common.Distance(defender.GetPos(), center)
		rate := s / base.ThrowHurtRadius
		if rate >= 1 {
			p.user.Warn("Distance is too far, rate: ", rate)
			continue
		}

		subhp := uint32(base.ThrowDamage * (1 - rate))
		if v.Dam > uint32(base.ThrowDamage) || v.Dam > 2*subhp {
			p.user.Warn("Damage is too large, damage: ", v.Dam)
			continue
		}

		v.Dam -= dongReduceDam(defender, v.Dam)
		if v.Dam <= 0 {
			continue
		}

		defender.DisposeSubHp(InjuredInfo{num: v.Dam, injuredType: throwdam, isHeadshot: false, attackid: p.user.GetID(), attackdbid: p.user.GetDBID()})
		p.user.Debug("Bomp damage to enemy, id: ", v.Id, " damage: ", v.Dam)
	}
}

func (p *RoomUserMsgProc) RPC_actionStateChange(oldaction, newaction uint64) {
	p.user.SetActionState(uint8(newaction))
	if newaction == ThrowPrepare || newaction == Throw {
		p.user.SetSlowMove(1)
	} else {
		p.user.SetSlowMove(0)
	}
}

// RPC_SetDoorState 客户端请求设置门的状态
func (p *RoomUserMsgProc) RPC_SetDoorState(doorID uint64, state uint32) {
	// p.user.Info("请求设置门的状态 ", p.user.GetName(), doorID, state)

	space := p.user.GetSpace().(*Scene)
	if space == nil {
		p.user.Error("SetDoorState failed, can't get space")
	}

	space.doorMgr.SetDoorState(p.user, doorID, state)
}

// RPC_KickingPlayer 执行踢出角色下线请求
func (p *RoomUserMsgProc) RPC_KickingPlayer(banAccountReason string) {
	p.user.RPC(iserver.ServerTypeClient, "KickingPlayerMsg", banAccountReason)

	p.user.RPC(iserver.ServerTypeClient, "NoticeKicking")
	err := GetSrvInst().DestroyEntityAll(p.user.GetID())
	if err != nil {
		p.user.Error("KickingPlayer failed, DestroyEntityAll err: ", err)
	}
}

func (p *RoomUserMsgProc) RPC_SetAutoReform(autoreform bool) {
	p.user.autoreform = autoreform
}

func (p *RoomUserMsgProc) RPC_OtherVehicleCollide(value float32, material uint8, position uint8) {
	p.user.CastRPCToAllClientExceptMe("OtherVehicleCollide", value, material, position)
}

func (p *RoomUserMsgProc) RPC_OtherVehicleBrake(speed float32) {
	p.user.CastRPCToAllClientExceptMe("OtherVehicleBrake", speed)
}

func (p *RoomUserMsgProc) RPC_OtherVehicleStopBrake() {
	p.user.CastRPCToAllClientExceptMe("OtherVehicleStopBrake")
}

func (p *RoomUserMsgProc) RPC_OtherVehicleSlip(speed float32) {
	p.user.CastRPCToAllClientExceptMe("OtherVehicleSlip", speed)
}

func (p *RoomUserMsgProc) RPC_OtherVehicleStopSlip() {
	p.user.CastRPCToAllClientExceptMe("OtherVehicleStopSlip")
}

func (p *RoomUserMsgProc) RPC_ReqMedalDataList() {
	// p.user.RPC(iserver.ServerTypeClient, "MedalDataList", p.user.medalhave)
}

func (p *RoomUserMsgProc) RPC_VehicleFullNotify(msg *protoMsg.VehicleFullNotify) {
	for _, v := range msg.Uid {
		user, ok := p.user.GetSpace().GetEntity(v).(*RoomUser)
		if !ok {
			continue
		}
		user.RPC(iserver.ServerTypeClient, "VehicleFullNotify", msg.Id)
	}
}

// RPC_CheaterReport 举报类型记录
func (p *RoomUserMsgProc) RPC_CheaterReport(uid uint64, id uint32) {
	p.user.Debug("RPC_CheaterReport uid:", uid, " id:", id)
	if uid == 0 {
		return
	}
	if p.user.reportCheck[uid] {
		return
	}

	info := &db.CheaterReportInfo{}
	info.ReportNum = make(map[uint32]uint32)
	if err := db.PlayerInfoUtil(uid).GetCheaterReportNum(info); err != nil {
		p.user.Error("GetCheaterReportNum failed, err:", err)
		return
	}
	info.ReportNum[id]++
	if err := db.PlayerInfoUtil(uid).SetCheaterReportNum(info); err != nil {
		p.user.Error("SetCheaterReportNum failed, err:", err)
		return
	}

	p.user.reportCheck[uid] = true
	p.user.tlogCheaterReportFlow(uid, id, info.ReportNum)
}

//RPC_EnterUserBattle 通知玩家进入成功
func (p *RoomUserMsgProc) RPC_EnterUserBattle() {
	p.user.enterBattle()
}

// MsgProc_ConnEventNotify 客户端Gateway连接断开事件
func (p *RoomUserMsgProc) MsgProc_ConnEventNotify(imsg msgdef.IMsg) {
	msg := imsg.(*msgdef.ConnEventNotify)

	if msg.EvtType == 1 {
		if p.user.userType == RoomUserTypeWatcher {
			//p.user.LeaveScene()
		} else {
			//p.user.BroadcastToWatchers(RoomUserTypeWatcher, "WatchTargetOffline")
		}
	}
	if msg.EvtType == 2 {
		//p.user.BroadcastToWatchers(RoomUserTypeWatcher, "WatchTargetReconnect")
	}
}

//RPC_SyncVoiceInfo 观战玩家进入 同步语言房间信息
func (p *RoomUserMsgProc) RPC_SyncVoiceInfo(voiceRoom uint64, memberId int32) {
	if p.user.userType != RoomUserTypeWatcher {
		return
	}
	if p.user.watchingTarget == 0 {
		return
	}
	scene, ok := p.user.GetSpace().(*Scene)
	if !ok {
		return
	}
	targetUser, ok := scene.GetEntity(p.user.watchingTarget).(*RoomUser)
	if !ok {
		return
	}
	p.user.voiceMemberId = memberId
	retMsg := &protoMsg.TeamVoiceInfo{}
	retMsg.MemberInfos = append(retMsg.MemberInfos, &protoMsg.MemVoiceInfo{
		MemberId: memberId,
		Uid:      p.user.GetID(),
	})

	targetUser.RPC(iserver.ServerTypeClient, "WatcherEnterVoice", retMsg)
	for _, v := range targetUser.GetTeamMembers() {
		if v == p.user.GetID() {
			continue
		}
		mem, ok := scene.GetEntity(v).(*RoomUser)
		if ok {
			mem.RPC(iserver.ServerTypeClient, "WatcherEnterVoice", retMsg)
		}
	}
}

// RPC_onExplodeDamage 爆炸物产生的伤害
func (p *RoomUserMsgProc) RPC_onExplodeDamage(info *protoMsg.ThrowDamageInfo) {
	if !p.user.StateM.CanAttack() {
		return
	}

	if p.user.checkAttackCold() {
		return
	}

	if !p.user.CanAttackAndConsume(0) {
		return
	}

	p.user.stateMgr.BreakRescue()

	// 坦克攻击
	if p.user.isInTank() {
		if p.user.IsInAddFuelState() {
			p.user.BreakAddFuel(6)
		}

		p.user.rpgExplodeDamage(info, tankShell, 1)
		return
	}

	// 火箭筒攻击
	if p.user.isBazookaWeaponUse() {
		p.user.rpgExplodeDamage(info, gunAttack, 2)
		return
	}
}

// RPC_SyncCurRoleSkillInfoReq 同步技能释放时间请求
func (p *RoomUserMsgProc) RPC_SyncCurRoleSkillTimeReq() {
	p.user.syncCurRoleSkillTime() // 通知当前使用的技能
}

// RPC_CheckStateSync 检测状态同步
func (p *RoomUserMsgProc) RPC_CheckStateSync(curState, curAction uint8) {
	state := p.user.GetBaseState()
	if curState == RoomPlayerBaseState_Stand && (state == RoomPlayerBaseState_Jump || state == RoomPlayerBaseState_Crouch || state == RoomPlayerBaseState_Down) {
		p.user.SetBaseState(curState)
	}

	// if p.user.GetActionState() != curAction {
	// 	p.user.SetActionState(curAction)
	// }
}

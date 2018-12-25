package main

import (
	"common"
	"excel"
	"math"
	"protoMsg"
	"strings"
	"time"
	"zeus/iserver"
	"zeus/linmath"
)

// isVehicleNoDamage 载具是否对武器的攻击免疫
func isVehicleNoDamage(vehicleId uint64, baseid uint32) bool {
	itemConfig, ok := excel.GetItem(uint64(baseid))
	if !ok {
		return false
	}

	carrierConfig, ok := excel.GetCarrier(vehicleId)
	if !ok {
		return false
	}

	strs := strings.Split(carrierConfig.NoDamage, ";")
	for _, str := range strs {
		if common.StringToUint64(str) == itemConfig.Type {
			return true
		}
	}

	return false
}

// isTank 载具是否是坦克
func isTank(vehicleId uint64) bool {
	carrierConfig, ok := excel.GetCarrier(vehicleId)
	if !ok {
		return false
	}

	return carrierConfig.Subtype == 2
}

func (user *RoomUser) upVehicle(thisid uint64) bool {
	space := user.GetSpace().(*Scene)
	item, ok := space.GetEntity(thisid).(*SpaceVehicle)
	if !ok {
		user.Warn("upVehicle failed, can't get vehicle, id: ", thisid)
		return false
	}

	if common.Distance(user.GetPos(), item.GetPos()) > 3.0 {
		user.Warn("upVehicle failed, distance is so far, distance: ", common.Distance(user.GetPos(), item.GetPos()))
		return false
	}

	if item.GetVehicleProp() == nil || item.GetVehicleProp().Reducedam == 0 {
		user.Warn("upVehicle failed, vehicle is broke")
		return false
	}

	if !user.StateM.CanUpVehicle() {
		user.Warn("upVehicle failed, user state not suit")
		return false
	}

	base, ok := excel.GetItem(uint64(item.GetVehicleProp().VehicleID))
	if !ok {
		user.Warn("upVehicle failed, vehicle not exist, id: ", item.GetVehicleProp().VehicleID)
		return false
	}

	if base.Type != ItemTypeCar {
		return false
	}
	if user.GetBaseState() == RoomPlayerBaseState_Swim && base.Subtype != 1 {
		return false
	}

	var vehProp *protoMsg.VehicleProp

	car, ok := space.cars[thisid]
	if ok {
		vehProp = car.VehicleProp
	} else {
		vehProp = item.GetVehicleProp()
	}

	vehProp.VehicleID = item.GetVehicleProp().VehicleID
	vehProp.PilotID = user.GetID()
	vehProp.Thisid = thisid
	vehProp.Enter = true

	if !item.temppos.IsEqual(linmath.Vector3_Zero()) {
		item.SetPos(item.temppos)
		item.SetRota(item.tempdir)
		item.temppos = linmath.Vector3_Zero()
		item.tempdir = linmath.Vector3_Zero()
	}
	user.SetRota(item.GetRota())
	user.SetPos(item.GetPos())

	if err := space.RemoveEntity(item.GetID()); err != nil {
		user.Error("RemoveEntity err: ", err)
	}

	user.stateMgr.BreakRescue()
	user.SetBaseState(RoomPlayerBaseState_Ride)
	user.SetActionState(ActionNone)
	// user.Info(user.GetID(), "上车")

	if car == nil {
		car = &Car{
			space:       space,
			VehicleProp: vehProp,
			physics:     nil,
		}
	}

	space.cars[car.Thisid] = car
	car.lastConsumePos = user.GetPos()
	car.RefreshInfo()
	car.haveobjs = item.haveobjs
	// user.Info("地图增加载具", car, len(space.cars))

	//强制关闭倍镜
	user.SetGunSight(0)

	if user.isInTank() {
		user.upTankInit(item)
	}

	user.sumData.IncrCarUserNum() //载具使用数量自增
	user.sumData.tmpcarDistance = user.GetPos()

	user.breakdownVehicle(thisid, false)
	user.AdviceNotify(common.NotifyCommon, 27) //通知进入驾驶状态
	user.tlogBattleFlow(behavetype_usecar, 0, 0, 0, 0, uint32(len(user.items)))
	user.Info("upVehicle success, thisid: ", thisid)

	// 上车清空防爆盾技能效果
	if user.SkillMgr.initiveEffect[SE_Shield] != nil {
		user.SkillData.skillEffectDam[SE_Shield] = 0
		delete(user.SkillMgr.initiveEffect, SE_Shield)
		user.SkillMgr.refreshSkillEffect()
	}

	user.BreakEffect(true)

	return true
}

func (user *RoomUser) compilotUp(pilotid uint64) {
	space := user.GetSpace().(*Scene)
	pilot, ok := space.GetEntity(pilotid).(*RoomUser)
	if !ok {
		user.Error("compilotUp failed, can't get pilot, id: ", pilotid)
		return
	}

	if common.Distance(user.GetPos(), pilot.GetPos()) > 3.0 {
		user.Warn("compilotUp failed, distance is so far, distance: ", common.Distance(user.GetPos(), pilot.GetPos()))
		return
	}

	if !user.StateM.CanUpVehicle() {
		return
	}

	vehProp := pilot.GetVehicleProp()
	if vehProp.VehicleID == 0 {
		user.Warn("compilotUp failed, no vehicle available")
		return
	}

	if vehProp == nil || vehProp.Reducedam == 0 {
		// user.Warn("车生命值未0， 不允许上车")
		return
	}

	itembase, ok := excel.GetItem(uint64(vehProp.VehicleID))
	if !ok {
		return
	}

	if itembase.Type != ItemTypeCar {
		return
	}
	if user.GetBaseState() == RoomPlayerBaseState_Swim && itembase.Subtype != 1 {
		return
	}

	base, ok := excel.GetCarrier(uint64(vehProp.VehicleID))
	if !ok {
		return
	}

	car, ok := space.cars[vehProp.Thisid]
	if !ok {
		user.Warn("compilotUp failed, can not find the vehicle, id: ", vehProp.Thisid)
		return
	}

	if car.GetUserNum() >= uint32(base.Seat) {
		// user.Warn("座位已满")
		return
	}

	if vehProp.PilotID == 0 {
		vehProp.PilotID = user.GetID()
		user.SetBaseState(RoomPlayerBaseState_Ride)
		user.SetActionState(ActionNone)
		// user.Info(user.GetID(), "副驾驶上车, 变为驾驶")
	} else {
		seatindex := make(map[uint32]uint64)
		for _, v := range vehProp.Copilots {
			seatindex[v.Index] = v.Id
		}

		var canuse uint32
		for i := 1; i <= int(base.Seat); i++ {
			if _, ok := seatindex[uint32(i)]; !ok {
				canuse = uint32(i)
				break
			}
		}
		if canuse == 0 {
			return
		}

		seat := &protoMsg.CopilotData{}
		seat.Index = canuse
		seat.Id = user.GetID()
		vehProp.Copilots = append(vehProp.Copilots, seat)
		user.SetBaseState(RoomPlayerBaseState_Passenger)
		user.SetActionState(ActionNone)
		// user.Info(user.GetID(), "副驾驶上车")
	}

	car.VehicleProp = vehProp
	car.RefreshInfo()

	//强制关闭倍镜
	user.SetGunSight(0)

	// 上车清空防爆盾技能效果
	if user.SkillMgr.initiveEffect[SE_Shield] != nil {
		user.SkillData.skillEffectDam[SE_Shield] = 0
		delete(user.SkillMgr.initiveEffect, SE_Shield)
		user.SkillMgr.refreshSkillEffect()
	}

	user.sumData.IncrCarUserNum() //载具使用数量自增
	user.sumData.tmpcarDistance = user.GetPos()

	user.breakdownVehicle(vehProp.Thisid, false)
	user.Info("compilotUp success, pilotid: ", pilotid)
	user.BreakEffect(true)
	user.stateMgr.BreakRescue()

	user.tlogBattleFlow(behavetype_usecar, 0, 0, 0, 0, uint32(len(user.items)))
}

// getRidingCar 获取玩家正在乘坐的载具
func (user *RoomUser) getRidingCar() *Car {
	space := user.GetSpace().(*Scene)
	if space == nil {
		return nil
	}

	prop := user.GetVehicleProp()
	if prop == nil {
		return nil
	}

	car, ok := space.cars[prop.GetThisid()]
	if !ok {
		return nil
	}

	return car
}

func (user *RoomUser) downVehicle(pos, dir linmath.Vector3, carprop *protoMsg.VehiclePhysics, reducedam bool) {
	space := user.GetSpace().(*Scene)
	if user.StateM.GetState() != RoomPlayerBaseState_Ride {
		user.Warn("downVehicle failed, not in a vehicle, state: ", user.StateM.GetState())
		return
	}

	prop := user.GetVehicleProp()
	if prop == nil {
		user.Warn("downVehicle failed, vehicle prop is nil")
		return
	}

	car, ok := space.cars[prop.Thisid]
	if !ok {
		user.Warn("downVehicle failed, can't find the vehicle, id: ", prop.Thisid)
		return
	}

	if user.IsInAddFuelState() {
		user.BreakAddFuel(4)
	}

	if user.isInTank() {
		user.SetIsInTank(0)
		user.downTankEnd()
	}

	car.PilotID = 0
	if car.PilotID == 0 && len(car.Copilots) == 0 {
		delete(space.cars, prop.Thisid)
	} else {
		car.RefreshInfo()
	}

	//强制关闭倍镜
	user.SetGunSight(0)

	isinboat := user.IsinBoat()

	//地图增加道具
	GetRefreshItemMgr(space).dropVehicle(uint32(prop.VehicleID), carprop, car.VehicleProp, car.haveobjs, user.GetID(), user.GetSubRotation1(), user.GetSubRotation2())

	vehProp := &protoMsg.VehicleProp{}
	user.SetVehicleProp(vehProp)
	user.SetVehiclePropDirty()

	user.downVehicleChangeState(pos)

	// user.Error("@@@@@@@@@, 传入方向", dir)
	userspeed := user.GetSpeed()
	userspeed.X += 0.1
	userspeed.Z += 0.1
	user.SetSpeed(userspeed)

	user.SetPos(pos)
	user.SetRota(dir)
	// user.Error(user.GetID(), "下车", "当前坐标: ", user.GetPos(), " 方向", user.GetRota())

	if reducedam && carprop != nil && carprop.Velocity != nil && !isinboat {
		velocity := linmath.NewVector3(carprop.Velocity.X, carprop.Velocity.Y, carprop.Velocity.Z)
		user.downVehicleDam(velocity.Len())
	}
	if d, ok := excel.GetItem(uint64(user.GetBodyProp().Baseid)); ok {
		if d.AddValue == "1" {
			user.CastRpcToAllClient("UpdateArmor", user.GetID(), uint32(2), true)
		}
	}

	user.sumData.CarDistance()
	user.sumData.tmpcarDistance = linmath.Vector3_Invalid()
	user.BreakEffect(true)
	user.Info("downVehicle success")
}

func (user *RoomUser) downVehicleDam(speed float32) {
	speed *= 3.6

	system, ok := excel.GetSystem(uint64(common.System_VehicleDamageA))
	if !ok {
		return
	}
	A := float32(system.Value)

	system, ok = excel.GetSystem(uint64(common.System_VehicleDamageB))
	if !ok {
		return
	}
	B := float32(system.Value)

	// user.Info("下车速度", speed, "参数A：", A, "参数B:", B, "伤害:", A*(speed-B))
	if speed > B {
		damage := A * (speed - B)
		user.DisposeSubHp(InjuredInfo{num: uint32(damage), injuredType: vehicledam, isHeadshot: false})
	}
}

func (user *RoomUser) compilotDown(pos, dir linmath.Vector3, downspeed float32) {
	// user.Error("副驾驶下车", user.GetID())
	space := user.GetSpace().(*Scene)
	if user.StateM.GetState() != RoomPlayerBaseState_Passenger {
		user.Warn("compilotDown failed, not in a vehicle, state: ", user.StateM.GetState())
		return
	}

	prop := user.GetVehicleProp()
	if prop == nil {
		user.Warn("compilotDown failed, vehicle prop is nil")
		return
	}

	car, ok := space.cars[prop.Thisid]
	if !ok {
		user.Warn("compilotDown failed, can't find the vehicle")
		return
	}

	if user.IsInAddFuelState() {
		user.BreakAddFuel(4)
	}

	for k, v := range car.Copilots {
		if v.Id == user.GetID() {
			car.Copilots = append(car.Copilots[:k], car.Copilots[k+1:]...)
			break
		}
	}

	if car.PilotID == 0 && len(car.Copilots) == 0 {
		delete(space.cars, prop.Thisid)
	} else {
		car.RefreshInfo()
	}

	//强制关闭倍镜
	user.SetGunSight(0)

	isinboat := user.IsinBoat()

	item, ok := space.GetEntity(car.Thisid).(*SpaceVehicle)
	if ok {
		item.SetVehicleProp(car.VehicleProp)
		item.SetVehiclePropDirty()
	} else {
		user.Warn("compilotDown failed, can't get the vehicle, id: ", car.Thisid)
	}

	vehProp := &protoMsg.VehicleProp{}
	user.SetVehicleProp(vehProp)
	user.SetVehiclePropDirty()

	user.downVehicleChangeState(pos)

	userspeed := user.GetSpeed()
	userspeed.X += 0.1
	userspeed.Z += 0.1
	user.SetSpeed(userspeed)

	// user.Error("副驾驶下车传入方向:", dir)
	user.SetRota(dir)
	user.SetPos(pos)
	// user.Error(user.GetID(), "副驾驶下车", "当前坐标: ", user.GetPos(), "方向:", user.GetRota())

	if !isinboat {
		user.downVehicleDam(downspeed)
	}
	if d, ok := excel.GetItem(uint64(user.GetBodyProp().Baseid)); ok {
		if d.AddValue == "1" {
			user.CastRpcToAllClient("UpdateArmor", user.GetID(), uint32(2), true)
		}
	}

	user.sumData.CarDistance()
	user.sumData.tmpcarDistance = linmath.Vector3_Invalid()
	user.BreakEffect(true)
	user.Info("compilotDown success")
}

// checkCanAttack 判断是否可以攻击载具
func (user *RoomUser) checkCanAttack(msg *protoMsg.AttackReq, targetPos linmath.Vector3) bool {
	space := user.GetSpace().(*Scene)
	orgion := linmath.NewVector3(msg.Origion.X, msg.Origion.Y, msg.Origion.Z)
	dir := linmath.NewVector3(msg.Dir.X, msg.Dir.Y, msg.Dir.Z)
	dir.Normalize()

	if common.Distance(user.GetPos(), orgion) > 5 {
		//log.Warn("攻击发射点位置异常:", attacker.GetPos(), " 发射点:", orgion, " 距离:", common.Distance(attacker.GetPos(), orgion))
		return false
	}

	distance, canattack, err := space.SphereRaycast(targetPos, 3.0, orgion, dir, user.GetWeaponDistance())
	if !canattack {
		user.Warn("SphereRaycast err: ", err, " distance: ", distance, " gun distance: ", user.GetWeaponDistance(), " orgion:", orgion, " dir:", dir, " defender:", targetPos)
		return false
	}

	return true
}

// attackVehicle 攻击载具
func (user *RoomUser) attackVehicle(msg *protoMsg.AttackReq) {
	if msg.Defendid == 0 {
		return
	}

	if user.isMeleeWeaponUse() {
		return
	}

	if msg.Origion == nil || msg.Dir == nil {
		//user.Warn("客户端发送错误空数据")
		return
	}

	space := user.GetSpace().(*Scene)
	car, ok := space.cars[msg.Defendid]
	if ok {
		if isVehicleNoDamage(car.GetVehicleID(), user.GetInUseWeapon()) {
			return
		}

		driver := car.GetUser()
		if driver == nil {
			// user.Error("未找到载具上乘客")
			return
		}

		attack := user.GetAttack(driver.GetPos(), false)
		if 0 == attack {
			// user.Error("获取到的伤害是0")
			return
		}

		if !user.checkCanAttack(msg, driver.GetPos()) {
			return
		}

		if car.Reducedam > attack {
			car.Reducedam -= attack
			// user.Debug("载具更新耐久", car.Reducedam)
		} else {
			car.Reducedam = 0
			car.Haveexplode = true
		}

		if car.IsInAddFuelState() {
			car.BreakAddFuelAll(3)
		}

		car.RefreshInfo()

		if car.Reducedam == 0 {
			car.onExplode()
			user.vehicleExplodeDamage(car.GetThisid())

			user.sumData.IncrCarDestoryNum() //载具摧毁自增
			user.breakdownVehicle(car.Thisid, true)
		}
	} else {
		item, ok := space.GetEntity(msg.Defendid).(*SpaceVehicle)
		if !ok {
			return
		}

		if isVehicleNoDamage(item.GetVehicleProp().GetVehicleID(), user.GetInUseWeapon()) {
			return
		}

		attack := user.GetAttack(item.GetPos(), false)
		if 0 == attack {
			//user.Error("获取到的伤害是0")
			return
		}

		if !user.checkCanAttack(msg, item.GetPos()) {
			return
		}

		prop := item.GetVehicleProp()
		if prop.Reducedam == 0 {
			return
		}

		if prop.Reducedam > attack {
			prop.Reducedam -= attack
			//user.Debug("载具更新耐久", prop.Reducedam)
		} else {
			prop.Reducedam = 0
			prop.Haveexplode = true
		}

		item.SetVehicleProp(prop)
		item.SetVehiclePropDirty()

		if prop.Reducedam == 0 {
			item.CastRPCToAllClientExceptMe("CarExplode", item.GetID())
			user.vehicleExplodeDamage(item.GetID())

			user.sumData.IncrCarDestoryNum() //载具摧毁自增
			user.breakdownVehicle(item.GetID(), true)
		}
	}
}

// explodeDamageVehicle 爆炸对载具造成伤害
func (user *RoomUser) explodeDamageVehicle(defendid uint64, dam uint32, center linmath.Vector3, base excel.ItemData) {
	if defendid == 0 {
		return
	}

	space := user.GetSpace().(*Scene)
	car, ok := space.cars[defendid]
	if ok {
		if isVehicleNoDamage(car.GetVehicleID(), uint32(base.Id)) {
			return
		}

		driver := car.GetUser()
		if driver == nil {
			//user.Warn("未找到载具上乘客")
			return
		}

		s := common.Distance(driver.GetPos(), center)
		rate := s / base.ThrowHurtRadius
		if rate >= 1 {
			//user.Warn("爆炸伤害距离太远")
			return
		}

		subhp := uint32(base.ThrowDamage * (1 - rate))
		if dam > uint32(base.ThrowDamage) || dam > 2*subhp {
			//user.Warn("爆炸伤害过大异常")
			return
		}

		if car.Reducedam > dam {
			car.Reducedam -= dam
			// user.Debug("载具更新耐久", car.Reducedam)
		} else {
			car.Reducedam = 0
			car.Haveexplode = true
		}

		car.RefreshInfo()

		if car.Reducedam == 0 {
			car.onExplode()
			user.vehicleExplodeDamage(car.GetThisid())

			user.sumData.IncrCarDestoryNum() //载具摧毁自增
		}

		user.Info("Explode damage to vehicle with person, thisid: ", defendid, " damage: ", dam, " reducedam: ", car.Reducedam)
	} else {
		item, ok := space.GetEntity(defendid).(*SpaceVehicle)
		if !ok {
			return
		}

		if isVehicleNoDamage(item.GetVehicleProp().GetVehicleID(), uint32(base.Id)) {
			return
		}

		s := common.Distance(item.GetPos(), center)
		rate := s / base.ThrowHurtRadius
		if rate >= 1 {
			//user.Warn("爆炸伤害距离太远")
			return
		}

		subhp := uint32(base.ThrowDamage * (1 - rate))
		if dam > uint32(base.ThrowDamage) || dam > 2*subhp {
			//user.Warn("爆炸伤害过大异常")
			return
		}

		prop := item.GetVehicleProp()
		if prop.Reducedam == 0 {
			return
		}

		if prop.Reducedam > dam {
			prop.Reducedam -= dam
			// user.Debug("载具更新耐久", prop.Reducedam)
		} else {
			prop.Reducedam = 0
			prop.Haveexplode = true
		}

		item.SetVehicleProp(prop)
		item.SetVehiclePropDirty()

		if prop.Reducedam == 0 {
			item.CastRPCToAllClientExceptMe("CarExplode", item.GetID())
			user.vehicleExplodeDamage(item.GetID())

			user.sumData.IncrCarDestoryNum() //载具摧毁自增
		}

		user.Info("Explode damage to empty vehicle, thisid: ", defendid, " damage: ", dam, " reducedam: ", prop.Reducedam)
	}
}

// vehicleExplodeDamage 载具爆炸，对车上乘客及周围玩家和载具造成伤害
func (user *RoomUser) vehicleExplodeDamage(thisid uint64) {
	space := user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	item, ok := space.mapitem[thisid]
	if !ok {
		user.Error("Can't get item in the map, thisid: ", thisid)
		return
	}

	itemConfig, ok := excel.GetItem(uint64(item.itemid))
	if !ok {
		user.Error("Can't get item config, baseid: ", item.itemid)
		return
	}

	//载具坐标
	var cood iserver.ICoordEntity

	car, ok := space.cars[thisid]
	if ok {
		cood = car.GetUser()
	} else {
		cood = space.GetEntity(thisid).(*SpaceVehicle)
	}

	if cood == nil {
		user.Error("Can't get car entity in the space, thisid: ", thisid)
		return
	}

	carPos := cood.GetPos()
	damagedCars := make(map[uint64]uint8) //伤害区域内的载具

	//遍历爆炸载具的aoi范围
	space.TravsalAOI(cood, func(o iserver.ICoordEntity) {
		target, ok := o.(iserver.ISpaceEntity)
		if !ok {
			return
		}

		if target.GetID() == thisid {
			return
		}

		targetPos := target.GetPos()

		dis := common.Distance(targetPos, carPos)
		if dis >= itemConfig.ThrowHurtRadius {
			return
		}

		//射线长度及方向
		diff := targetPos.Sub(carPos)
		rayDis := diff.Len()
		diff.Normalize()

		//检测障碍物
		_, _, _, hit, _ := space.Raycast(carPos, diff, rayDis, unityLayerBuilding|unityLayerFurniture)
		if hit {
			return
		}

		rate := dis / itemConfig.ThrowHurtRadius
		subhp := uint32(itemConfig.ThrowDamage * (1 - rate))

		defender, ok := space.GetEntity(o.GetID()).(iDefender)
		if ok {
			//对附近有人的载具造成伤害
			prop := defender.GetVehicleProp()
			carid := uint64(0)

			if prop != nil && prop.GetThisid() != thisid {
				carid = prop.GetThisid()
				if _, ok := damagedCars[carid]; !ok {
					user.explodeDamageVehicle(carid, subhp, carPos, itemConfig)
					damagedCars[carid] = 1
				}
			}

			if defender.isInTank() {
				return
			}

			if carid != 0 && defender.GetState() == RoomPlayerBaseState_WillDie {
				return
			}

			subhp -= dongReduceDam(defender, subhp)
			if subhp <= 0 {
				return
			}

			//对附近玩家造成伤害
			defender.DisposeSubHp(InjuredInfo{num: subhp, injuredType: carExplode, isHeadshot: false, attackid: user.GetID(), attackdbid: 0})
			user.Infof("Car %d explode damage to player %d, damage value: %d\n", thisid, defender.GetDBID(), subhp)

			return
		}

		//对附近的空载具造成伤害
		if damagedCars[target.GetID()] == 0 {
			user.explodeDamageVehicle(target.GetID(), subhp, carPos, itemConfig)
			damagedCars[target.GetID()] = 1
		}
	})

	//对乘客造成伤害
	if car != nil {
		for _, v := range car.Copilots {
			other, ok := space.GetEntity(v.Id).(*RoomUser)
			if ok {
				other.DisposeSubHp(InjuredInfo{num: other.GetHP(), injuredType: carExplode, attackid: user.GetID(), attackdbid: user.GetDBID(), isHeadshot: false})
			}
		}

		other, ok := space.GetEntity(car.PilotID).(*RoomUser)
		if ok {
			other.DisposeSubHp(InjuredInfo{num: other.GetHP(), injuredType: carExplode, attackid: user.GetID(), attackdbid: user.GetDBID(), isHeadshot: false})
		}
	}
}

//IsinBoat 在船上
func (user *RoomUser) IsinBoat() bool {
	vehProp := user.GetVehicleProp()
	if vehProp.VehicleID == 0 {
		return false
	}

	base, ok := excel.GetItem(uint64(vehProp.VehicleID))
	if !ok {
		return false
	}

	return base.Type == ItemTypeCar && base.Subtype == 1
}

func (user *RoomUser) vehicleEngineBroke(autodown bool) {
	car := user.getRidingCar()
	if car == nil {
		return
	}

	if car.Reducedam == 0 {
		return
	}

	car.Reducedam = 0
	car.RefreshInfo()
	pilot := car.GetUser()

	if autodown {
		car.autoDownVehicle()
	}

	if pilot != nil {
		pilot.breakdownVehicle(car.GetThisid(), true)
	}
}

func (user *RoomUser) deathOnVehicle() {
	space := user.GetSpace().(*Scene)
	if user.StateM.GetState() != RoomPlayerBaseState_Ride && user.StateM.GetState() != RoomPlayerBaseState_Passenger {
		user.Warn("deathOnVehicle failed, not in a vehicle, state: ", user.StateM.GetState())
		return
	}

	prop := user.GetVehicleProp()
	if prop == nil {
		user.Warn("deathOnVehicle failed, vehicle prop is nil")
		return
	}

	car, ok := space.cars[prop.Thisid]
	if !ok {
		user.Warn("deathOnVehicle failed, can't get the vehicle, id: ", prop.Thisid)
		return
	}

	if user.isInTank() {
		user.SetIsInTank(0)
		user.downTankEnd()
	}

	if car.PilotID == user.GetID() {
		car.PilotID = 0
	} else {
		for k, v := range car.Copilots {
			if v.Id == user.GetID() {
				car.Copilots = append(car.Copilots[:k], car.Copilots[k+1:]...)
				break
			}
		}
	}

	if car.PilotID == 0 && len(car.Copilots) == 0 {
		delete(space.cars, prop.Thisid)
	} else {
		car.RefreshInfo()
	}

	if user.StateM.GetState() == RoomPlayerBaseState_Ride {
		physics := car.getPhysics(user.GetPos(), user.GetRota())
		GetRefreshItemMgr(space).dropVehicle(uint32(prop.VehicleID), physics, car.VehicleProp, car.haveobjs, user.GetID(), user.GetSubRotation1(), user.GetSubRotation2())
	} else {
		item, ok := space.GetEntity(car.Thisid).(*SpaceVehicle)
		if ok {
			item.SetVehicleProp(car.VehicleProp)
			item.SetVehiclePropDirty()
		} else {
			user.Warn("deathOnVehicle failed, can't get the vehicle, id: ", car.Thisid)
		}
	}

	vehProp := &protoMsg.VehicleProp{}
	user.SetVehicleProp(vehProp)
	user.SetVehiclePropDirty()

	ppos := linmath.NewVector2(0, 1)
	qpos := linmath.NewVector2(1, 0)
	r := float64(user.GetRota().Y * math.Pi / 180)
	newP := ppos.Mul(float32(math.Cos(r)))
	newP.AddS(qpos.Mul(float32(math.Sin(r))))
	downforward := linmath.Vector3{
		X: newP.X,
		Y: 0,
		Z: newP.Y,
	}
	fdir := linmath.Vector3{
		X: downforward.Z,
		Y: 0,
		Z: -downforward.X,
	}
	fdir.Normalize()
	oldpos := user.GetPos()

	newpos := oldpos.Add(fdir.Mul(2))
	newpos = getCanPutHeight(space, newpos)
	// waterLevel := user.GetSpace().(*Scene).mapdata.Water_height
	// posy, err := user.GetSpace().GetHeight(newpos.X, newpos.Z)
	// if err == nil && newpos.Y > waterLevel {
	// 	newpos.Y = posy
	// }
	newpos.Y += 0.4
	user.SetPos(newpos)

	rota := user.GetRota()
	rota.Z = 0
	user.SetRota(rota)
	if d, ok := excel.GetItem(uint64(user.GetBodyProp().Baseid)); ok { //从车上下来同步吉利服
		if d.AddValue == "1" {
			user.CastRpcToAllClient("UpdateArmor", user.GetID(), uint32(2), true)
		}
	}

	user.sumData.CarDistance()
	user.sumData.tmpcarDistance = linmath.Vector3_Invalid()
	user.Infof("Death on vehicle, pos: %+v\n", user.GetPos())
}

func (user *RoomUser) autoDownVehicle() {
	if user.GetState() == RoomPlayerBaseState_Ride || user.GetState() == RoomPlayerBaseState_Passenger {
		physics := &protoMsg.VehiclePhysics{}
		physics.Position = &protoMsg.Vector3{
			X: user.GetPos().X,
			Y: user.GetPos().Y,
			Z: user.GetPos().Z,
		}
		physics.Rotation = &protoMsg.Vector3{
			X: user.GetRota().X,
			Y: user.GetRota().Y,
			Z: user.GetRota().Z,
		}

		// user.Error("自动下车!!!!!!!!!!!")
		user.downVehicle(user.GetPos(), user.GetRota(), physics, false)
		user.compilotDown(user.GetPos(), user.GetRota(), 0)
	}
}

func (user *RoomUser) downVehicleChangeState(pos linmath.Vector3) {
	if user.CanChangeState() {
		system, ok := excel.GetSystem(uint64(common.System_FallHeight))
		posy, err := user.GetSpace().GetHeight(pos.X, pos.Z)
		if err == nil && ok {
			fallheight := float32(system.Value) / 100.0
			if pos.Y > posy+fallheight {
				user.SetSpeed(linmath.Vector3_Zero())
				user.SetBaseState(RoomPlayerBaseState_Fall)
				user.SetActionState(ActionNone)
				user.lastfallpos = user.GetPos()
				// user.Error("角色下车掉落")
			} else {
				user.SetBaseState(RoomPlayerBaseState_Stand)
				user.SetActionState(ActionNone)
				pos.Y = posy
			}
		} else {
			user.SetBaseState(RoomPlayerBaseState_Stand)
			user.SetActionState(ActionNone)
		}
	}
}

// isInTank 判断玩家是否在坦克上
func (user *RoomUser) isInTank() bool {
	car := user.getRidingCar()
	if car == nil {
		return false
	}

	return isTank(car.GetVehicleID())
}

// upTankInit 玩家上坦克后初始工作
func (user *RoomUser) upTankInit(item *SpaceVehicle) {
	user.SetIsInTank(1)
	user.equipTankShell() //装备坦克炮

	user.SetSubRotation1(item.GetSubRotation1())
	user.SetSubRotation1Dirty()
	user.SetSubRotation2(item.GetSubRotation2())
	user.SetSubRotation2Dirty()

	user.sumData.lastUpTankTime = time.Now().Unix()
}

// downTankEnd 玩家下坦克后收尾工作
func (user *RoomUser) downTankEnd() {
	user.unequipTankShell() //卸载坦克炮

	useTime := time.Now().Unix() - user.sumData.lastUpTankTime
	user.sumData.tankUseTime += useTime
}

// tankShellExplode 坦克炮弹爆炸对周围玩家和载具造成伤害
func (user *RoomUser) tankShellExplode(msg *protoMsg.AttackReq) {
	if user.IsInAddFuelState() {
		user.BreakAddFuel(6)
	}

	//炮弹爆炸
	user.shellExplodeDamage(msg, common.Item_TankShell, tankShell, 1)
}

// autoDownTank 自动下坦克
func (user *RoomUser) autoDownTank() {
	space := user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	car := user.getRidingCar()
	if car == nil {
		return
	}

	physics := car.getPhysics(user.GetPos(), user.GetRota())
	if physics.Velocity == nil {
		physics.Velocity = &protoMsg.Vector3{0, 0, 0}
	}

	rota := user.GetRota()
	rota.X = 0
	rota.Z = 0

	pos := user.GetPos()
	carpos := linmath.NewVector3(physics.GetPosition().X, physics.GetPosition().Y, physics.GetPosition().Z)

	if common.Distance(pos, carpos) < 4 {
		pos.X += 4
		pos.Z += 4

		waterLevel := space.mapdata.Water_height
		isWater, err := space.IsWater(pos.X, pos.Z, waterLevel)

		if err == nil && !isWater {
			ground, err := space.GetHeight(pos.X, pos.Z)
			if err == nil {
				pos.Y = ground
			}
		}
	}

	user.downVehicle(pos, rota, physics, false)
}

// syncAOITanksToMe 断线重连后，将aoi范围内的坦克状态同步给客户端
func (user *RoomUser) syncAOITanksToMe() {
	space := user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	space.TravsalAOI(user, func(o iserver.ICoordEntity) {
		player, ok := space.GetEntity(o.GetID()).(*RoomUser)
		if !ok {
			return
		}

		if player.isInTank() {
			user.RPC(iserver.ServerTypeClient, "TankSubRotationSync", player.GetID(), player.GetSubRotation1(), player.GetSubRotation2())
		}

		return
	})
}

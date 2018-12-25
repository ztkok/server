package main

import (
	"common"
	"protoMsg"
	"zeus/linmath"
)

// Car 车辆信息
type Car struct {
	space *Scene

	*protoMsg.VehicleProp
	physics *protoMsg.VehiclePhysics

	haveobjs       map[uint32]*Item //武器装备道具
	lastConsumePos linmath.Vector3  //载具最近一次耗油时的位置
}

// RefreshInfo 刷新车辆信息
func (sf *Car) RefreshInfo() {
	for _, v := range sf.Copilots {
		other, ok := sf.space.GetEntity(v.Id).(*RoomUser)
		if ok {
			// log.Warn("副驾驶更新信息", v.Id)
			other.SetVehicleProp(sf.VehicleProp)
			other.SetVehiclePropDirty()
		}
	}

	other, ok := sf.space.GetEntity(sf.PilotID).(*RoomUser)
	if ok {
		// log.Warn("驾驶更新信息", sf.PilotID)
		other.SetVehicleProp(sf.VehicleProp)
		other.SetVehiclePropDirty()
	}
}

//GetUserNum 获取车上人数
func (sf *Car) GetUserNum() uint32 {
	var ret uint32

	if sf.PilotID != 0 {
		ret++
	}

	ret += uint32(len(sf.Copilots))
	return ret
}

// GetPassengers 获取车上乘客
func (sf *Car) GetPassengers() []uint64 {
	var ret []uint64

	if sf.PilotID != 0 {
		ret = append(ret, sf.PilotID)
	}

	for _, copilot := range sf.Copilots {
		ret = append(ret, copilot.Id)
	}

	return ret
}

func (sf *Car) getPhysics(pos, rota linmath.Vector3) *protoMsg.VehiclePhysics {
	physics := &protoMsg.VehiclePhysics{}
	if sf.physics == nil || sf.physics.Position == nil || sf.physics.Rotation == nil {
		physics.Position = &protoMsg.Vector3{
			X: pos.X,
			Y: pos.Y,
			Z: pos.Z,
		}
		physics.Rotation = &protoMsg.Vector3{
			X: rota.X,
			Y: rota.Y,
			Z: rota.Z,
		}
		// seelog.Error("下车使用乘客坐标:", pos, rota)
	} else {
		physics = sf.physics
		// seelog.Error("下车物理信息:", physics.Position, physics.Rotation)
	}

	return physics
}

func (sf *Car) autoDownVehicle() {
	other, ok := sf.space.GetEntity(sf.PilotID).(*RoomUser)
	if ok {
		physics := sf.getPhysics(other.GetPos(), other.GetRota())
		rota := other.GetRota()
		rota.X = 0
		rota.Z = 0

		if physics.Velocity == nil {
			physics.Velocity = &protoMsg.Vector3{
				X: 0,
				Y: 0,
				Z: 0,
			}
		}

		carpos := linmath.NewVector3(physics.GetPosition().X, physics.GetPosition().Y, physics.GetPosition().Z)
		pos := other.GetPos()
		if common.Distance(pos, carpos) < 2 {
			pos.X += 2
			pos.Z += 2
			waterLevel := sf.space.mapdata.Water_height
			isWater, err := sf.space.IsWater(pos.X, pos.Z, waterLevel)
			if err == nil && !isWater {
				ground, err := sf.space.GetHeight(pos.X, pos.Z)
				if err == nil && pos.Y < ground {
					pos.Y = ground
				}
			}
		}
		other.downVehicle(pos, rota, physics, false)
	}

	for _, v := range sf.Copilots {
		other, ok := sf.space.GetEntity(v.Id).(*RoomUser)
		if ok {
			rota := other.GetRota()
			rota.X = 0
			rota.Z = 0

			pos := other.GetPos()
			pos.X += 2
			pos.Z += 2
			waterLevel := sf.space.mapdata.Water_height
			isWater, err := sf.space.IsWater(pos.X, pos.Z, waterLevel)
			if err == nil && !isWater {
				ground, err := sf.space.GetHeight(pos.X, pos.Z)
				if err == nil && pos.Y < ground {
					pos.Y = ground
				}
			}
			other.compilotDown(pos, rota, 0)
		}
	}
}

//GetUser 获取一个乘客
func (sf *Car) GetUser() *RoomUser {
	other, ok := sf.space.GetEntity(sf.PilotID).(*RoomUser)
	if ok {
		return other
	}

	for _, v := range sf.Copilots {
		other, ok := sf.space.GetEntity(v.Id).(*RoomUser)
		if ok {
			return other
		}
	}

	return nil
}

func (sf *Car) onExplode() {
	user := sf.GetUser()
	if user == nil {
		return
	}

	user.CastRPCToAllClient("CarExplode", sf.Thisid)
}

//SubUserHp 对乘客造成伤害
func (sf *Car) SubUserHp(subhp uint32, injuredType uint32) {
	for _, v := range sf.Copilots {
		other, ok := sf.space.GetEntity(v.Id).(*RoomUser)
		if ok {
			other.DisposeSubHp(InjuredInfo{num: subhp, injuredType: injuredType, isHeadshot: false})
		}
	}

	other, ok := sf.space.GetEntity(sf.PilotID).(*RoomUser)
	if ok {
		other.DisposeSubHp(InjuredInfo{num: subhp, injuredType: injuredType, isHeadshot: false})
	}
}

//IsInAddFuelState 车辆是否正在加油
func (sf *Car) IsInAddFuelState() bool {
	user, ok := sf.space.GetEntity(sf.PilotID).(*RoomUser)
	if ok && user.IsInAddFuelState() {
		return true
	}

	for _, v := range sf.Copilots {
		user, ok := sf.space.GetEntity(v.Id).(*RoomUser)
		if ok && user.IsInAddFuelState() {
			return true
		}
	}

	return false
}

//BreakAddFuelAll 所有玩家加油失败
func (sf *Car) BreakAddFuelAll(reason uint8) {
	user, ok := sf.space.GetEntity(sf.PilotID).(*RoomUser)
	if ok {
		user.BreakAddFuel(reason)
	}

	for _, v := range sf.Copilots {
		user, ok := sf.space.GetEntity(v.Id).(*RoomUser)
		if ok {
			user.BreakAddFuel(reason)
		}
	}
}

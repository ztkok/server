package main

import (
	"common"
	"entitydef"
	"excel"
	"math/rand"
	"protoMsg"
	"strings"
	"time"
	"zeus/linmath"
	"zeus/space"
)

//SpaceVehicle 载具
type SpaceVehicle struct {
	entitydef.VehicleDef
	space.Entity

	temppos, tempdir linmath.Vector3
	haveobjs         map[uint32]*Item //武器装备道具
}

// Init 初始化底层回调
func (item *SpaceVehicle) Init(initParam interface{}) {
	space := item.GetSpace().(*Scene)
	if space == nil {
		return
	}

	mapitem, ok := space.mapitem[item.GetID()]
	if !ok {
		item.Error("Init failed, itemid: ", item.GetID())
		return
	}

	if mapitem.item == nil {
		item.Error("Init failed, item is nil")
		return
	}

	item.haveobjs = mapitem.haveobjs

	vehicle := &protoMsg.Vehicle{}
	prop := &protoMsg.VehicleProp{}

	zerovec := &protoMsg.Vector3{}
	physics := &protoMsg.VehiclePhysics{
		Position:        zerovec,
		Rotation:        zerovec,
		Velocity:        zerovec,
		AngularVelocity: zerovec,
	}

	prop.VehicleID = uint64(mapitem.item.GetBaseID())
	prop.Reducemax = uint32(mapitem.item.base.Reducedam)
	prop.Reducedam = uint32(mapitem.item.base.Reducedam)

	carrierConfig, ok := excel.GetCarrier(uint64(mapitem.itemid))
	if !ok {
		item.Error("Init failed, can't get carrier config")
		return
	}

	//初始化油量
	if carrierConfig.MaxFuel > 0 {
		strs := strings.Split(carrierConfig.InitFuelRange, ";")
		if len(strs) == 2 {
			beg := common.StringToUint32(strs[0])
			end := common.StringToUint32(strs[1])
			prop.FuelLeft = float32(beg + uint32(rand.Intn(int(end-beg+1))))
		}
		prop.FuelMax = carrierConfig.MaxFuel
	}

	if initParam != "" {
		if err := vehicle.Unmarshal(initParam.([]byte)); err != nil {
			item.Error("Init failed, Unmarshal err: ", err)
		} else {
			prop = vehicle.Prop
			physics = vehicle.Physics
		}
	}

	item.SetVehicleProp(prop)
	item.SetVehiclePropDirty()
	item.SetVehiclePhysics(physics)
	item.SetVehiclePhysicsDirty()
	item.Setownerid(vehicle.Ownerid)

	item.SetPos(mapitem.pos)

	rota := mapitem.dir
	if rota.IsEqual(linmath.Vector3_Zero()) {
		rand.Seed(time.Now().UnixNano())
		rota.Y = rand.Float32() * 360
	}
	item.SetRota(rota)
}

// Destroy 销毁的时候底层回调
// func (item *SpaceVehicle) Destroy() {
// 	item.Info("销毁车", item)
// }

// CreateNewEntityState 创建一个新的状态快照，由框架层调用
func (item *SpaceVehicle) CreateNewEntityState() space.IEntityState {
	return NewRoomVehicleState()
}

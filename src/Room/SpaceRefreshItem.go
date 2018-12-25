package main

import (
	"common"
	"excel"
	"math/rand"
	"protoMsg"
	"strings"
	"time"
	"zeus/iserver"
	"zeus/linmath"
)

//GetRefreshItemMgr 刷新道具
func GetRefreshItemMgr(scene *Scene) *RefreshItemMgr {
	if scene.refreshitem == nil {
		scene.refreshitem = &RefreshItemMgr{
			tb:         scene,
			refreshbox: 0,
			droppos:    linmath.Vector3_Zero(),
			dropstart:  linmath.Vector3_Zero(),
			boxlist:    make(map[uint64]*protoMsg.DropBoxInfo, 0),
		}
	}

	return scene.refreshitem
}

//RefreshItemMgr 刷新道具
type RefreshItemMgr struct {
	tb             *Scene
	refreshbox     int64
	refreshhostage bool

	mapItemRateIds map[int]uint8 //地图道具刷新概率表中对应配置项id

	droppos     linmath.Vector3
	dropstart   linmath.Vector3
	dropboxtime int64
	boxlist     map[uint64]*protoMsg.DropBoxInfo
}

//SetRefreshBox 刷新补给箱时间
func (sf *RefreshItemMgr) SetRefreshBox(t int64) {
	sf.refreshbox = t
}

//InitMapItemRateIds 初始化地图道具刷新规则
func (sf *RefreshItemMgr) InitMapItemRateIds() {
	sf.mapItemRateIds = map[int]uint8{
		1:  1,
		2:  1,
		3:  1,
		4:  1,
		5:  1,
		6:  1,
		7:  1,
		8:  1,
		9:  1,
		10: 1,
	}

	info, ok := excel.GetMatchmode(uint64(sf.tb.uniqueId))
	if !ok {
		return
	}

	itemrule := strings.Split(info.Itemrule, ";")
	if len(itemrule) == 0 {
		return
	}
	index := rand.Intn(len(itemrule))

	strs := strings.Split(itemrule[index], "|")
	if len(strs) != 7 {
		return
	}

	sf.mapItemRateIds = map[int]uint8{
		common.StringToInt(strs[0]): 1,
		common.StringToInt(strs[1]): 1,
		common.StringToInt(strs[2]): 1,
		common.StringToInt(strs[3]): 1,
		common.StringToInt(strs[4]): 1,
		common.StringToInt(strs[5]): 1,
		common.StringToInt(strs[6]): 1,
	}
}

//InitSceneItem 初始化刷新地图道具
func (sf *RefreshItemMgr) InitSceneItem() {
	sf.InitMapItemRateIds()
	sf.tb.Infof("mapItemRateIds: %+v\n", sf.mapItemRateIds)

	tb := sf.tb
	mapranges := tb.GetRanges()
	if mapranges == nil {
		return
	}

	mapitemrate := GetMapItemRate(tb).mapitemrate
	rand.Seed(time.Now().UnixNano())

	for k, v := range mapitemrate {
		if sf.mapItemRateIds[int(k)] == 0 {
			continue
		}

		typeranges, err := mapranges.GetRangeList(int(k)%100 - 1)
		if err != nil {
			continue
		}

		for _, typerange := range typeranges {

			for _, config := range v {
				rate := rand.Intn(10000) + 1
				if rate > config.createrate {
					continue
				}

				for index := 0; index < config.num; index++ {
					randnum := rand.Intn(config.total) + 1
					var percent int

					for _, itemrate := range config.itemrate {
						percent += itemrate.rate
						if percent >= randnum {

							for _, itembaseid := range itemrate.items {
								base, ok := excel.GetItem(uint64(itembaseid))
								if !ok {
									sf.tb.Error("InitSceneItem failed, id: ", itembaseid)
									continue
								}

								droppos := typerange.GetRandomPos()
								var entityid uint64
								if base.Type == ItemTypeCar {
									entityid = GetSrvInst().FetchTempID()
									sceneitem := &SceneItem{
										id:       entityid,
										pos:      droppos,
										itemid:   itembaseid,
										item:     NewItem(itembaseid, 1),
										haveobjs: sf.dropVehicleWeapons(uint64(itembaseid)),
									}

									tb.mapitem[sceneitem.id] = sceneitem

									tb.AddEntity("Vehicle", entityid, 0, "", false, true)
								} else {
									entityid = tb.GetEntityTempID()
									sceneitem := &SceneItem{
										id:       entityid,
										pos:      droppos,
										itemid:   itembaseid,
										item:     NewItem(itembaseid, 1),
										haveobjs: make(map[uint32]*Item),
									}

									if base.Type == ItemTypeBox && base.Subtype == ItemSubtypeSupply {
										reward := GetMapItemRate(tb).randRuleItem(tb.r, base.Reward)
										for _, obj := range reward {
											tb.mapindex++
											itemptr := NewItem(obj, 1)
											if itemptr != nil {
												sceneitem.haveobjs[tb.mapindex] = itemptr
												//sf.tb.Debug("生成补给箱内部道具", sf.tb.mapindex, obj, len(dropitem.haveobjs))
											}
										}
									}

									tb.mapitem[sceneitem.id] = sceneitem

									tb.AddTinyEntity("Item", entityid, "")
								}
							}
							break
						}
					}
				}
			}
		}
	}
}

// dropVehicleWeapons 生成载具装备的武器道具
func (sf *RefreshItemMgr) dropVehicleWeapons(vehicleID uint64) map[uint32]*Item {
	weapons := make(map[uint32]*Item)

	carrierConfig, ok := excel.GetCarrier(vehicleID)
	if !ok {
		return weapons
	}

	if carrierConfig.MaxShells == 0 {
		return weapons
	}

	var shells uint32

	strs := strings.Split(carrierConfig.InitShellsRange, ";")
	if len(strs) == 2 {
		beg := common.StringToUint32(strs[0])
		end := common.StringToUint32(strs[1])
		shells = beg + uint32(rand.Intn(int(end-beg+1)))
	}

	thisid := GetSrvInst().FetchTempID()
	weapon := NewItem(uint32(carrierConfig.Weapon), 1)
	weapon.thisid = thisid
	weapon.bullet = shells
	weapon.gunreform = append(weapon.gunreform, uint32(carrierConfig.Telescope))

	sf.tb.mapindex++
	weapons[sf.tb.mapindex] = weapon

	return weapons
}

func (sf *RefreshItemMgr) dropBox() {
	if sf.dropboxtime == 0 {
		return
	}

	now := time.Now().UnixNano() / (1000 * 1000)
	if now >= sf.dropboxtime {
		sf.dropboxtime = 0
		sf.droppos = getXZCanPutHeight(sf.tb, sf.droppos)
		droppos := &protoMsg.Vector3{
			X: sf.droppos.X,
			Y: sf.droppos.Y,
			Z: sf.droppos.Z,
		}

		boxid := uint32(common.GetTBSystemValue(common.System_RefreshBoxID))
		entityid := sf.dropItemByID(boxid, sf.droppos)
		//sf.tb.Debug("生成补给箱", sf.droppos, boxid, len(sf.tb.mapitem))

		fakeboxid := uint32(common.GetTBSystemValue(common.System_RefreshFakeBoxID))
		sf.tb.TravsalEntity("Player", func(e iserver.IEntity) {
			if e == nil {
				return
			}

			if user, ok := e.(*RoomUser); ok {
				user.RPC(iserver.ServerTypeClient, "RefreshBoxNotify", droppos, entityid, fakeboxid)
				user.AdviceNotify(common.NotifyCommon, 21)
				// sf.tb.Debug("补给箱生成广播成功")
			}
		})

		boxinfo := &protoMsg.DropBoxInfo{}
		boxinfo.Thisid = entityid
		boxinfo.Fakebox = fakeboxid
		boxinfo.Pos = &protoMsg.Vector3{X: sf.droppos.X, Y: sf.droppos.Y, Z: sf.droppos.Z}
		boxinfo.Havepick = false
		sf.boxlist[entityid] = boxinfo
	}
}

//RefreshBox 刷新补给箱
func (sf *RefreshItemMgr) RefreshBox(now int64) {
	tb := sf.tb
	refreshcount := GetRefreshZoneMgr(sf.tb).GetRefreshCount()
	base, ok := excel.GetMaprule(refreshcount)
	if !ok {
		return
	}

	if sf.refreshbox != 0 && base.Boxrefresh != 0 && now >= sf.refreshbox+int64(base.Boxrefresh) {
		sf.refreshbox = 0

		sf.droppos = tb.GetCircleRamdomPos(GetRefreshZoneMgr(tb).nextsafecenter, GetRefreshZoneMgr(tb).nextsaferadius)
		//sf.droppos = linmath.NewVector3(6556, 200, 3802.6)

		droppos := sf.droppos
		droppos.Y = float32(common.GetTBSystemValue(common.System_RefreshBoxHeight))

		sf.dropstart = sf.randAirStart()
		sf.dropstart.Y = float32(common.GetTBSystemValue(common.System_RefreshBoxHeight))

		dropstart := sf.dropstart
		flyspeed := float32(common.GetTBSystemValue(common.System_RefreshBoxSpeed))
		direction := droppos.Sub(dropstart)
		distance := flyspeed * float32(sf.getFlyTime())
		dropend := dropstart.Add(direction.Mul(distance / direction.Len()))

		dropstartproto := &protoMsg.Vector3{
			X: sf.dropstart.X,
			Y: sf.dropstart.Y,
			Z: sf.dropstart.Z,
		}
		dropendproto := &protoMsg.Vector3{
			X: dropend.X,
			Y: dropend.Y,
			Z: dropend.Z,
		}

		now := time.Now().UnixNano() / (1000 * 1000)
		dist := droppos.Sub(sf.dropstart).Len()
		droptime := (dist / flyspeed) * 1000
		sf.dropboxtime = now + int64(droptime)

		//sf.tb.Debug("生成补给箱", dropstart, sf.droppos, sf.dropboxtime, droptime)
		tb.TravsalEntity("Player", func(e iserver.IEntity) {
			if e == nil {
				return
			}

			if user, ok := e.(*RoomUser); ok {
				user.RPC(iserver.ServerTypeClient, "RefreshBoxStart", dropstartproto, dropendproto)
				//user.AdviceNotify(common.NotifyCommon, 21)
				// sf.tb.Debug("补给箱生成广播成功")
			}
		})
	}
}

//RefreshBox GM刷新补给箱
func (sf *RefreshItemMgr) GmRefreshBox(droppos linmath.Vector3) {
	tb := sf.tb

	sf.droppos = droppos
	droppos.Y = float32(common.GetTBSystemValue(common.System_RefreshBoxHeight))

	sf.dropstart = sf.randAirStart()
	sf.dropstart.Y = float32(common.GetTBSystemValue(common.System_RefreshBoxHeight))

	dropstart := sf.dropstart
	flyspeed := float32(common.GetTBSystemValue(common.System_RefreshBoxSpeed))
	direction := droppos.Sub(dropstart)
	distance := flyspeed * float32(sf.getFlyTime())
	dropend := dropstart.Add(direction.Mul(distance / direction.Len()))

	dropstartproto := &protoMsg.Vector3{
		X: sf.dropstart.X,
		Y: sf.dropstart.Y,
		Z: sf.dropstart.Z,
	}
	dropendproto := &protoMsg.Vector3{
		X: dropend.X,
		Y: dropend.Y,
		Z: dropend.Z,
	}

	now := time.Now().UnixNano() / (1000 * 1000)
	dist := droppos.Sub(sf.dropstart).Len()
	droptime := (dist / flyspeed) * 1000
	sf.dropboxtime = now + int64(droptime)

	sf.tb.Debug("GM生成补给箱", dropstart, sf.droppos, sf.dropboxtime, droptime)
	tb.TravsalEntity("Player", func(e iserver.IEntity) {
		if e == nil {
			return
		}

		if user, ok := e.(*RoomUser); ok {
			user.RPC(iserver.ServerTypeClient, "RefreshBoxStart", dropstartproto, dropendproto)
			//user.AdviceNotify(common.NotifyCommon, 21)
			// sf.tb.Debug("补给箱生成广播成功")
		}
	})
}

func (sf *RefreshItemMgr) dropItemByID(baseid uint32, droppos linmath.Vector3) uint64 {
	base, ok := excel.GetItem(uint64(baseid))
	if !ok {
		sf.tb.Error("dropItemByID failed, id: ", baseid)
		return 0
	}

	//entityid := server.GetEntityTempID()
	entityid := sf.tb.GetEntityTempID()
	dropitem := &SceneItem{
		id:       entityid,
		pos:      droppos,
		itemid:   baseid,
		item:     NewItem(baseid, 1),
		haveobjs: make(map[uint32]*Item),
	}

	if base.Type == ItemTypeBox && base.Subtype == ItemSubtypeSupply {
		add := common.GetTBSystemValue(common.System_GetBoxCold)
		if baseid == 1508 {
			add = common.GetTBSystemValue(common.System_SuperDropBoxCold)
		}
		dropitem.cangettime = time.Now().Unix() + int64(add)
		reward := GetMapItemRate(sf.tb).randRuleItem(sf.tb.r, base.Reward)
		for _, obj := range reward {
			sf.tb.mapindex++
			itemptr := NewItem(obj, 1)
			if itemptr != nil {
				dropitem.haveobjs[sf.tb.mapindex] = itemptr
				//sf.tb.Debug("生成补给箱内部道具", sf.tb.mapindex, obj, len(dropitem.haveobjs))
			}
		}
	}

	sf.tb.mapitem[dropitem.id] = dropitem
	sf.tb.AddTinyEntity("Item", entityid, "")
	sf.tb.Info("AddTinyEntity, itemid: ", entityid)
	return entityid
}

func (sf *RefreshItemMgr) dropVehicle(baseid uint32, carprop *protoMsg.VehiclePhysics, prop *protoMsg.VehicleProp, haveobjs map[uint32]*Item, ownerid uint64, subRota1, subRota2 *protoMsg.Vector3) {
	if sf.tb == nil || sf.tb.isEnd() {
		return
	}

	if carprop.Position == nil {
		carprop.Position = &protoMsg.Vector3{}
	}
	if carprop.Rotation == nil {
		carprop.Rotation = &protoMsg.Vector3{}
	}

	_, ok := excel.GetItem(uint64(baseid))
	if !ok {
		sf.tb.Error("dropVehicle failed, id: ", baseid)
		return
	}

	// entityid := sf.tb.GetEntityTempID()
	entityid := prop.Thisid
	dropitem := &SceneItem{
		id:       entityid,
		pos:      linmath.NewVector3(carprop.Position.X, carprop.Position.Y, carprop.Position.Z),
		dir:      linmath.NewVector3(carprop.Rotation.X, carprop.Rotation.Y, carprop.Rotation.Z),
		itemid:   baseid,
		item:     NewItem(baseid, 1),
		haveobjs: haveobjs,
	}

	sf.tb.mapitem[dropitem.id] = dropitem

	vehicle := &protoMsg.Vehicle{}
	vehicle.Prop = prop
	vehicle.Physics = carprop
	vehicle.Ownerid = ownerid

	data, err := vehicle.Marshal()
	if err != nil {
		sf.tb.Error("dropVehicle failed, Marshal err: ", err)
		return
	}
	sf.tb.AddEntity("Vehicle", entityid, 0, data, true, true)

	sv := sf.tb.GetEntity(entityid).(*SpaceVehicle)
	sv.SetSubRotation1(subRota1)
	sv.SetSubRotation1Dirty()
	sv.SetSubRotation2(subRota2)
	sv.SetSubRotation2Dirty()
}

func getCanPutHeight(space *Scene, droppos linmath.Vector3) linmath.Vector3 {
	origin := linmath.Vector3{
		X: droppos.X,
		Y: droppos.Y + 1,
		Z: droppos.Z,
	}
	direction := linmath.Vector3{
		X: 0,
		Y: -1,
		Z: 0,
	}

	waterLevel := space.mapdata.Water_height
	_, pos, _, hit, _ := space.Raycast(origin, direction, 100, unityLayerGround|unityLayerBuilding|unityLayerFurniture|unityLayerPlantAndRock)
	if hit && pos.Y > waterLevel {
		return pos
	}

	return droppos
}

func getXZCanPutHeight(space *Scene, droppos linmath.Vector3) linmath.Vector3 {
	origin := linmath.Vector3{
		X: droppos.X,
		Y: 1000,
		Z: droppos.Z,
	}
	direction := linmath.Vector3{
		X: 0,
		Y: -1,
		Z: 0,
	}

	_, pos, _, hit, _ := space.Raycast(origin, direction, 2000, unityLayerGround|unityLayerBuilding|unityLayerFurniture|unityLayerPlantAndRock)
	if hit {
		return pos
	}

	return droppos
}

func (sf *RefreshItemMgr) dropItem(id uint64, sceneItem *Item, droppos linmath.Vector3) {
	if sceneItem == nil {
		return
	}

	//还原成原始道具
	itemRestore(id, sceneItem)

	droppos = getCanPutHeight(sf.tb, droppos)

	itembaseid := sceneItem.GetBaseID()
	//entityid := server.GetEntityTempID()
	entityid := sf.tb.GetEntityTempID()
	dropitem := &SceneItem{
		id:       entityid,
		pos:      droppos,
		itemid:   itembaseid,
		item:     NewCopyItem(sceneItem),
		haveobjs: make(map[uint32]*Item),
	}

	sf.tb.mapitem[dropitem.id] = dropitem
	sf.tb.AddTinyEntity("Item", entityid, "")
}

func (sf *RefreshItemMgr) dropArmor(baseid uint32, reducedam uint32, droppos linmath.Vector3) {
	if baseid == 0 {
		return
	}

	droppos = getCanPutHeight(sf.tb, droppos)

	_, ok := excel.GetItem(uint64(baseid))
	if !ok {
		sf.tb.Error("dropArmor failed, id: ", baseid)
		return
	}

	//entityid := server.GetEntityTempID()
	entityid := sf.tb.GetEntityTempID()
	dropitem := &SceneItem{
		id:       entityid,
		pos:      droppos,
		itemid:   baseid,
		item:     NewArmorItem(baseid, reducedam),
		haveobjs: make(map[uint32]*Item),
	}

	sf.tb.mapitem[dropitem.id] = dropitem
	sf.tb.AddTinyEntity("Item", entityid, "")
}

func (sf *RefreshItemMgr) RefreshHostage() {
	if sf.refreshhostage {
		return
	}

	if len(sf.tb.members) <= 10 {
		sf.refreshhostage = true

		/*
			droppos := sf.tb.GetCircleRamdomPos(sf.tb.safecenter, sf.tb.saferadius)
			sf.dropItemByID(1503, droppos)
			//sf.tb.Debug("生成人质:", droppos)
		*/
	}
}

func (sf *RefreshItemMgr) randAirStart() linmath.Vector3 {
	tb := sf.tb
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	style := r.Intn(4)
	switch style {
	case 0:
		return linmath.Vector3{
			X: 0,
			Y: 0,
			Z: float32(int(tb.mapdata.W_start) + r.Intn(int(tb.mapdata.W_end-tb.mapdata.W_start))),
		}

	case 1:
		return linmath.Vector3{
			X: tb.mapdata.Width,
			Y: 0,
			Z: float32(int(tb.mapdata.E_start) + r.Intn(int(tb.mapdata.E_end-tb.mapdata.E_start))),
		}

	case 2:
		return linmath.Vector3{
			X: float32(int(tb.mapdata.S_start) + r.Intn(int(tb.mapdata.S_end-tb.mapdata.S_start))),
			Y: 0,
			Z: 0,
		}

	case 3:
		return linmath.Vector3{
			X: float32(int(tb.mapdata.N_start) + r.Intn(int(tb.mapdata.N_end-tb.mapdata.N_start))),
			Y: 0,
			Z: tb.mapdata.Height,
		}

	}

	return linmath.Vector3{
		X: 0,
		Y: 0,
		Z: float32(int(tb.mapdata.W_start) + r.Intn(int(tb.mapdata.W_end-tb.mapdata.W_start))),
	}
}

func (sf *RefreshItemMgr) getFlyTime() uint32 {
	flyspeed := float32(common.GetTBSystemValue(common.System_RefreshBoxSpeed))
	airStart := linmath.NewVector2(0, 0)
	airEnd := linmath.NewVector2(sf.tb.mapdata.Width, sf.tb.mapdata.Height)
	airlineDist := airEnd.Sub(airStart).Len()
	flyTime := uint32(airlineDist / flyspeed)

	return flyTime
}

func (sf *RefreshItemMgr) resendBoxList(user *RoomUser) {
	msg := &protoMsg.DropBoxList{}
	for _, v := range sf.boxlist {
		msg.Data = append(msg.Data, v)
	}

	user.RPC(iserver.ServerTypeClient, "DropBoxList", msg)
	//user.Debug("@@@@@@@@@@@@@@@@@@@@@ RefreshBoxList")
}

func (sf *RefreshItemMgr) callUpDropBox(pos linmath.Vector3, caller string, color uint32) {
	pos = getXZCanPutHeight(sf.tb, pos)
	droppos := &protoMsg.Vector3{
		X: pos.X,
		Y: pos.Y,
		Z: pos.Z,
	}

	entityid := sf.dropItemByID(1508, pos)
	sf.tb.Debug("生成补给箱", droppos, 1508, len(sf.tb.mapitem))

	sf.tb.TravsalEntity("Player", func(e iserver.IEntity) {
		if e == nil {
			return
		}

		if user, ok := e.(*RoomUser); ok {
			user.RPC(iserver.ServerTypeClient, "RefreshBoxNotify", droppos, entityid, uint32(1508))
			user.RPC(iserver.ServerTypeClient, "CallUpBoxNotify", caller, color)
		}
	})

	boxinfo := &protoMsg.DropBoxInfo{}
	boxinfo.Thisid = entityid
	boxinfo.Fakebox = 1508
	boxinfo.Pos = droppos
	boxinfo.Havepick = false
	sf.boxlist[entityid] = boxinfo

	sf.tb.Info("callUpDropBox: ", caller)
}

func (sf *RefreshItemMgr) callUpDropTank(pos linmath.Vector3, caller string, color, itemID uint32) {
	pos = getXZCanPutHeight(sf.tb, pos)
	droppos := &protoMsg.Vector3{
		X: pos.X,
		Y: pos.Y,
		Z: pos.Z,
	}

	height := float32(common.GetTBSystemValue(common.System_RefreshBoxHeight))
	itemId := itemID

	entityid := GetSrvInst().FetchTempID()
	second := common.ByHeightCalTime(height - pos.Y)
	sf.tb.callBoxItem[entityid] = CallBoxItem{time.Now().Unix() + int64(second+2), pos, itemID}

	sf.tb.TravsalEntity("Player", func(e iserver.IEntity) {
		if e == nil {
			return
		}

		if user, ok := e.(*RoomUser); ok {
			user.RPC(iserver.ServerTypeClient, "RefreshTankNotify", entityid, droppos, itemId, caller, second, color)
		}
	})

	sf.tb.Debug("生成坦克", droppos, itemId, caller, second, entityid)
}

type CallBoxItem struct {
	tim    int64
	pos    linmath.Vector3
	itemID uint32
}

// dropCallBoxItem 生成信号枪召唤的道具
func (sf *RefreshItemMgr) dropCallBoxItem() {
	for k, v := range sf.tb.callBoxItem {
		if time.Now().Unix() >= v.tim {
			var realItemId uint32
			if v.itemID == common.Item_TankBox0 {
				realItemId = uint32(common.GetTBSystemValue(common.System_CallBoxRealItem))
			} else if v.itemID == common.Item_TankBox1 {
				realItemId = uint32(common.GetTBSystemValue(common.System_CallBoxRealItem1))
			}
			sceneitem := &SceneItem{
				id:       k,
				pos:      v.pos,
				itemid:   realItemId,
				item:     NewItem(realItemId, 1),
				haveobjs: GetRefreshItemMgr(sf.tb).dropVehicleWeapons(uint64(realItemId)),
			}

			sf.tb.mapitem[sceneitem.id] = sceneitem
			sf.tb.AddEntity("Vehicle", k, 0, "", false, true)
			sf.tb.Debug("生成坦克", v.pos, k, realItemId, v.tim)

			delete(sf.tb.callBoxItem, k)
		}
	}
}

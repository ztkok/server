package main

import (
	"common"
	"db"
	"excel"
	"math"
	"math/rand"
	"protoMsg"
	"sort"
	"strings"
	"time"
	"zeus/iserver"
	"zeus/linmath"

	log "github.com/cihub/seelog"
)

const (
	//ItemTypeWeaonPistol 手枪
	ItemTypeWeaonPistol = 1
	//ItemTypeShotGun 散弹枪
	ItemTypeShotGun = 4
	//ItemTypeBazooka 火箭筒
	ItemTypeBazooka = 5
	//ItemTypeMeleeWeapon 近战武器
	ItemTypeMeleeWeapon = 9
	//ItemTypeWeapon 武器
	ItemTypeWeapon = 10
	//ItemTypeConsume 消耗品
	ItemTypeConsume = 11
	//ItemTypeArmor 防具
	ItemTypeArmor = 12
	//ItemTypeWeaponReform 武器改造
	ItemTypeWeaponReform = 14
	//ItemTypeBox 补给箱
	ItemTypeBox = 15
	//ItemTypePack 背包
	ItemTypePack = 16
	//ItemTypeCar 载具
	ItemTypeCar = 19
	//ItemTypeBullet 子弹
	ItemTypeBullet = 20
	//ItemTypeThrow 投资物
	ItemTypeThrow = 21
	//ItemTypeInitItems 初始道具
	ItemTypeInitItems = 28
	//ItemTypeFastRescue 快速救援
	ItemTypeFastRescue = 29
	//ItemTypeDeapRescue 深度救援
	ItemTypeDeapRescue = 30
	//ItemTypeCoinBonus 金币加成特权
	ItemTypeCoinBonus = 31
	//ItemTypeBraveBonus 勇气值加成特权
	ItemTypeBraveBonus = 32
	//ItemTypeAddCoinLimit 增加金币上限
	ItemTypeAddCoinLimit = 33
	//ItemTypeAddBraveLimit 增加勇气值上限
	ItemTypeAddBraveLimit = 34
)

const (
	//ItemSubtypeSupply 空头箱
	ItemSubtypeSupply = 1
	//ItemSubtypeBody 背心
	ItemSubtypeBody = 1
	//ItemSubtypeHelmet 头盔
	ItemSubtypeHelmet = 2
	//ItemSubtypeSight 倍镜
	ItemSubtypeSight = 3
	//ItemSubtypeAdditem 弹夹
	ItemSubtypeAdditem = 7
)

const (
	//WeaponMaxNum 武器最大数量
	WeaponMaxNum = 2
)

//Item 道具数据
type Item struct {
	thisid        uint64
	count         uint32
	base          excel.ItemData
	gunreform     []uint32
	bullet        uint32
	reducedam     uint32
	fakebullet    uint32
	relatedItemID uint32 //关联的道具ID
	relatedUserID uint64 //关联的玩家ID
}

//NewCopyItem 创建道具
func NewCopyItem(other *Item) *Item {
	item := &Item{
		thisid:        0,
		count:         other.count,
		base:          other.base,
		gunreform:     make([]uint32, 0),
		bullet:        other.bullet,
		reducedam:     other.reducedam,
		fakebullet:    other.bullet + 5,
		relatedItemID: other.relatedItemID,
		relatedUserID: other.relatedUserID,
	}

	for _, v := range other.gunreform {
		item.gunreform = append(item.gunreform, v)
	}

	return item
}

//NewArmorItem 创建防具
func NewArmorItem(baseid uint32, reduce uint32) *Item {
	itembase, ok := excel.GetItem(uint64(baseid))
	if !ok {
		log.Warn("Item id not exist, baseid：", baseid)
		return nil
	}

	item := &Item{
		thisid:    0,
		count:     1,
		base:      itembase,
		gunreform: make([]uint32, 0),
		bullet:    0,
		reducedam: reduce,
	}

	return item
}

//NewItem 创建道具
func NewItem(baseid uint32, num uint32) *Item {
	itembase, ok := excel.GetItem(uint64(baseid))
	if !ok {
		log.Warn("Item id not exist, baseid：", baseid)
		return nil
	}

	var clip uint32
	gunbase, gunok := excel.GetGun(uint64(baseid))
	if gunok {
		clip = uint32(gunbase.Clipcap)
	}
	// randclip := uint32((float32(rand.Intn(71)+30) / 100.0) * float32(clip))
	// randclip += (5 / 2)
	// coeff := randclip / 5
	// beiclip := coeff * 5
	// if beiclip != 0 && beiclip < clip {
	// 	clip = beiclip
	// }

	item := &Item{
		thisid:     0,
		count:      num,
		base:       itembase,
		gunreform:  make([]uint32, 0),
		bullet:     clip,
		reducedam:  uint32(itembase.Reducedam),
		fakebullet: clip + 5,
	}

	return item
}

//NewItemBullet 创建道具
func NewItemBullet(baseid uint32, num uint32) *Item {
	itembase, ok := excel.GetItem(uint64(baseid))
	if !ok {
		log.Warn("Item id not exist, baseid：", baseid)
		return nil
	}

	var clip uint32
	gunbase, gunok := excel.GetGun(uint64(baseid))
	if gunok {
		clip = uint32(gunbase.Clipcap)
	}

	item := &Item{
		thisid:     0,
		count:      num,
		base:       itembase,
		gunreform:  make([]uint32, 0),
		bullet:     clip,
		reducedam:  uint32(itembase.Reducedam),
		fakebullet: clip + 5,
	}

	return item
}

//MainPack 包裹
type MainPack struct {
	user IRoomChracter

	items map[uint64]*Item //背包里的道具

	equips    map[uint64]*Item //武器装备
	useweapon uint64           //主武器thisid

	tmpequips    map[uint64]*Item //武器装备缓存
	tmpuseweapon uint64           //主武器thisid缓存

	shotgunbullet map[uint64]int
	havethrownum  map[uint32]uint32
	autoreform    bool

	initCells uint32 //初始背包格子数
}

//GetBaseID 获取id
func (sf *Item) GetBaseID() uint32 {
	return uint32(sf.base.Id)
}

//GetMaxNum 最大叠加上限
func (sf *Item) GetMaxNum() uint32 {
	return uint32(sf.base.Addlimit)
}

func (sf *Item) fillInfo() *protoMsg.T_Object {
	info := &protoMsg.T_Object{}
	info.Baseid = sf.GetBaseID()
	info.Count = sf.count
	info.Thisid = sf.thisid
	info.Gunreform = sf.gunreform
	info.Bullet = sf.bullet
	info.Reducedam = sf.reducedam

	return info
}

//RefreshToMe 更新数据
func (sf *Item) RefreshToMe(user IRoomChracter) {
	updateproto := &protoMsg.RefreshObjectNotify{}
	updateproto.Obj = sf.fillInfo()

	if !user.isAI() {
		user.RPC(iserver.ServerTypeClient, "RefreshObjectNotify", updateproto)
		/*
			user.CastMsgToMe(updateproto)
		*/
	}
}

func (sf *Item) checkBulletDrop(user IRoomChracter, drop bool) {
	gunbase, ok := excel.GetGun(uint64(sf.GetBaseID()))
	if !ok {
		return
	}

	clipcap := uint32(gunbase.Clipcap)
	deltas := common.StringToMapUint32(gunbase.Magazinedelta, "|", ";")

	for _, reform := range sf.gunreform {
		clipcap += deltas[reform]
	}

	if sf.bullet > clipcap {
		if drop {
			space := user.GetSpace().(*Scene)
			item := NewItem(uint32(gunbase.Consumebullet), sf.bullet-clipcap)
			if space != nil && item != nil {
				GetRefreshItemMgr(space).dropItem(user.GetID(), item, linmath.RandXZ(user.GetPos(), 0.5))
			}
		} else {
			user.add_item(uint32(gunbase.Consumebullet), sf.bullet-clipcap)
		}

		sf.bullet = clipcap
		sf.fakebullet = clipcap + 5
	}
}

func (sf *Item) checkBullet(user IRoomChracter) {
	sf.checkBulletDrop(user, false)
}

func (sf *Item) clearReform(space *Scene, user IRoomChracter) {
	gunbase, ok := excel.GetGun(uint64(sf.GetBaseID()))
	if !ok {
		return
	}

	//丢掉配件
	for _, reform := range sf.gunreform {
		item := NewItem(reform, 1)
		if item != nil {
			GetRefreshItemMgr(space).dropItem(user.GetID(), item, linmath.RandXZ(user.GetPos(), 0.5))
		}
	}
	sf.gunreform = make([]uint32, 0)

	//丢掉子弹
	if sf.bullet > uint32(gunbase.Clipcap) && gunbase.Consumebullet != 0 {
		item := NewItem(uint32(gunbase.Consumebullet), sf.bullet-uint32(gunbase.Clipcap))
		if item != nil {
			sf.bullet = uint32(gunbase.Clipcap)
			GetRefreshItemMgr(space).dropItem(user.GetID(), item, linmath.RandXZ(user.GetPos(), 0.5))
		}
	}
}

func (sf *Item) isAddBullet(user IRoomChracter, baseid uint32) bool {
	gunbase, ok := excel.GetGun(uint64(sf.GetBaseID()))
	if !ok {
		return false
	}

	clipcap := uint32(gunbase.Clipcap)
	deltas := common.StringToMapUint32(gunbase.Magazinedelta, "|", ";")

	for _, reform := range sf.gunreform {
		if reform == baseid && deltas[reform] != 0 {
			if sf.bullet > clipcap {
				return true
			}
		}
	}

	return false
}

//InitMainPack 初始化
func InitMainPack(u IRoomChracter) *MainPack {
	mainpack := &MainPack{useweapon: 0}
	mainpack.user = u
	mainpack.items = make(map[uint64]*Item)
	mainpack.equips = make(map[uint64]*Item)
	mainpack.tmpequips = make(map[uint64]*Item)
	mainpack.shotgunbullet = make(map[uint64]int)
	mainpack.havethrownum = make(map[uint32]uint32)
	mainpack.autoreform = true

	mainpack.initPackCells()

	return mainpack
}

// initPackCells 初始化玩家的背包格子数
func (sf *MainPack) initPackCells() {
	system, ok := excel.GetSystem(uint64(common.System_InitCell))
	if ok {
		sf.initCells += uint32(system.Value)
	}

	var itemid uint32
	util := db.PlayerGoodsUtil(sf.user.GetDBID())

	if util.IsOwnGoods(common.Item_PackExtendHigh) {
		itemid = common.Item_PackExtendHigh
	} else if util.IsOwnGoods(common.Item_PackExtendMiddle) {
		itemid = common.Item_PackExtendMiddle
	} else if util.IsOwnGoods(common.Item_PackExtendLow) {
		itemid = common.Item_PackExtendLow
	}

	base, ok := excel.GetItem(uint64(itemid))
	if ok {
		sf.initCells += uint32(base.Addcell)
	}
}

// initPackItems 初始化玩家的背包道具
func (sf *MainPack) initPackItems() {
	space := sf.user.GetSpace().(*Scene)
	if space.GetMatchMode() != common.MatchModeNormal {
		return
	}

	if len(sf.items) != 0 {
		return
	}

	util := db.PlayerGoodsUtil(sf.user.GetDBID())
	for _, v := range util.GetAllGoodsInfo() {
		base, ok := excel.GetItem(uint64(v.Id))
		if !ok || base.Type != ItemTypeInitItems {
			continue
		}

		items := common.StringToMapUint32(base.AddValue, "|", ";")
		for k, v := range items {
			sf.AddItem(NewItem(k, v))
		}
	}
}

func (sf *MainPack) add_item(baseid uint32, num uint32) uint32 {
	defer func() {
		var ownDrinks20 bool
		if baseid == common.Item_EnergyDrink && sf.GetItemNum(baseid) >= 20 {
			ownDrinks20 = true
		}

		// log.Debug("增加道具", baseid, num)
		space := sf.user.GetSpace().(*Scene)
		if space != nil {
			user, ok := space.GetEntityByDBID("Player", sf.user.GetDBID()).(*RoomUser)
			if ok {
				user.sumData.isOwnDrink20 = ownDrinks20
				user.tlogBattleFlow(behavetype_pickup, uint64(baseid), 0, 0, 0, uint32(len(user.items)))
			}
		}
	}()

	var ret uint32
	for _, v := range sf.items {
		if v.GetBaseID() != baseid {
			continue
		}

		if v.count+num <= v.GetMaxNum() {
			v.count += num
			ret = num
		} else {
			ret = v.GetMaxNum() - v.count
			v.count = v.GetMaxNum()
		}
		v.RefreshToMe(sf.user)

		num -= ret
		if num == 0 {
			return ret
		}
	}

	for num > 0 {
		base, ok := excel.GetItem(uint64(baseid))
		if !ok {
			return ret
		}
		var nn uint32
		if num > uint32(base.Addlimit) {
			nn = uint32(base.Addlimit)
		} else {
			nn = num
		}

		thisid := GetSrvInst().FetchTempID()
		item := &Item{}
		item.thisid = thisid
		item.count = nn
		item.base = base
		sf.items[item.thisid] = item
		item.RefreshToMe(sf.user)

		num -= nn
		ret += nn
	}

	return ret
}

// getWeaponEquipId 获取玩家为武器装备的指定类型配件的baseid
func (sf *MainPack) getWeaponEquipId(baseid, typ uint32) uint32 {
	user, _ := sf.user.GetRealPtr().(*RoomUser)
	if user == nil {
		return 0
	}

	weapons := user.GetWeaponEquipInGame()
	if weapons == nil {
		return 0
	}

	for _, weapon := range weapons.GetWeapons() {
		if weapon == nil {
			continue
		}

		if weapon.GetWeaponId() != baseid {
			continue
		}

		for _, additionId := range weapon.GetAdditions() {
			base, ok := excel.GetItem(uint64(additionId))
			if !ok {
				continue
			}

			if base.Type == uint64(typ) {
				return additionId
			}
		}
	}

	return 0
}

// itemReplace 捡到原始道具后，替换为幻化道具
func (sf *MainPack) itemReplace(item *Item) {
	if item.relatedItemID != 0 {
		return
	}

	baseid := sf.getWeaponEquipId(uint32(item.base.Id), uint32(item.base.Type))
	if baseid == 0 {
		return
	}

	base, ok := excel.GetItem(uint64(baseid))
	if !ok {
		return
	}

	item.relatedItemID = uint32(item.base.Id)
	item.relatedUserID = sf.user.GetID()
	item.base = base
}

// itemRestore 丢掉幻化道具，还原成原始道具
func itemRestore(id uint64, item *Item) {
	if item == nil || item.relatedItemID == 0 {
		return
	}

	if item.relatedUserID != id {
		return
	}

	base, ok := excel.GetItem(uint64(item.relatedItemID))
	if !ok {
		return
	}

	item.relatedItemID = 0
	item.relatedUserID = 0
	item.base = base
}

// equipTankShell 玩家上坦克装备坦克炮
func (sf *MainPack) equipTankShell() {
	car := sf.user.getRidingCar()
	if car == nil {
		return
	}

	var weapon *Item
	for _, v := range car.haveobjs {
		if v.base.Type == ItemTypeBazooka {
			weapon = v
		}
	}

	if weapon == nil {
		return
	}

	sf.tmpequips = sf.equips
	sf.tmpuseweapon = sf.useweapon

	sf.equips = make(map[uint64]*Item)
	sf.equips[weapon.thisid] = weapon
	sf.useweapon = weapon.thisid

	sf.user.SetChracterMapDataInfo(sf.user.fillMapData())
	sf.user.SetChracterMapDataInfoDirty()
}

// unequipTankShell 下坦克卸载坦克炮
func (sf *MainPack) unequipTankShell() {
	sf.equips = sf.tmpequips
	sf.useweapon = sf.tmpuseweapon

	sf.tmpequips = make(map[uint64]*Item)
	sf.tmpuseweapon = 0
	sf.user.SetChracterMapDataInfo(sf.user.fillMapData())
	sf.user.SetChracterMapDataInfoDirty()
}

func (sf *MainPack) add_gun(sceneitem *Item) {
	space := sf.user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	sf.itemReplace(sceneitem) //替换为幻化枪支

	base := sceneitem.base
	thisid := GetSrvInst().FetchTempID()

	item := &Item{}
	item.thisid = thisid
	item.count = 1
	item.base = base
	item.bullet = sceneitem.bullet
	item.fakebullet = sceneitem.bullet + 5
	item.relatedItemID = sceneitem.relatedItemID
	item.relatedUserID = sceneitem.relatedUserID

	for _, v := range sceneitem.gunreform {
		item.gunreform = append(item.gunreform, v)
	}
	//log.Debug("捡枪", base.Id, item.bullet, item.gunreform, len(item.gunreform))

	dropgun := false
	var dropitem *Item

	replace := false
	if base.Type == ItemTypeWeaonPistol {
		for _, v := range sf.equips {
			if v.base.Type == ItemTypeWeaonPistol {

				dropgun = true
				dropitem = v

				item.thisid = v.thisid
				sf.equips[item.thisid] = item
				replace = true
				break
			}
		}
	}

	if base.Type == ItemTypeMeleeWeapon {
		for _, v := range sf.equips {
			if v.base.Type == ItemTypeMeleeWeapon {
				space := sf.user.GetSpace().(*Scene)
				if space != nil {
					//log.Debug("丢掉手枪", v.GetBaseID())
					GetRefreshItemMgr(space).dropItem(sf.user.GetID(), v, linmath.RandXZ(sf.user.GetPos(), 0.5))
				}
				item.thisid = v.thisid
				sf.equips[item.thisid] = item
				replace = true
				break
			}
		}
	}

	if !replace {
		if len(sf.equips) == 0 {
			sf.equips[item.thisid] = item
			sf.useweapon = item.thisid
		} else if len(sf.equips) == 1 {
			sf.equips[item.thisid] = item
			if item.base.Type == 5 { //玩家捡起火箭筒 做额外通知
				sf.user.CastRpcToAllClient("UpdateRPG", sf.user.GetID())
			}
		} else {

			dropgun = true
			dropitem = sf.equips[sf.useweapon]

			delete(sf.equips, sf.useweapon)
			sf.equips[item.thisid] = item
			sf.useweapon = item.thisid
		}
	}

	if dropgun && dropitem != nil {
		if sf.autoreform {
			sf.usePackDropAutoReform(dropitem, item)

			for _, v := range sf.equips {
				if v.thisid == item.thisid {
					continue
				}

				gunreform := make([]uint32, 0)
				for _, reform := range dropitem.gunreform {
					if sf.canAutoReform(v, reform) {
						sf.equipGunReform(v, reform)
					} else {
						gunreform = append(gunreform, reform)
					}
				}
				dropitem.gunreform = gunreform
			}

			if len(dropitem.gunreform) != 0 {
				dropitem.gunreform = sf.sortByLight(dropitem.gunreform)
			}

			gunreform := make([]uint32, 0)
			for _, reform := range dropitem.gunreform {
				if sf.LeftCell() > 0 {
					sf.add_item(reform, 1)
				} else {
					gunreform = append(gunreform, reform)
				}
			}
			dropitem.gunreform = gunreform
		}

		dropitem.checkBullet(sf.user)
		dropitem.clearReform(space, sf.user)

		GetRefreshItemMgr(space).dropItem(sf.user.GetID(), dropitem, linmath.RandXZ(sf.user.GetPos(), 0.5))
	}

	//log.Info("当前使用武器", sf.equips[sf.useweapon].GetBaseID())
	if sf.autoreform {
		sf.usePackAutoReform(item)
	}

	if !sf.user.isAI() {
		sf.RefreshGunNotifyAll(item)
	}

	sf.user.SetChracterMapDataInfo(sf.user.fillMapData())
	sf.user.SetChracterMapDataInfoDirty()

	user, _ := sf.user.GetRealPtr().(*RoomUser)
	if user != nil {
		user.tlogBattleFlow(behavetype_pickup, uint64(sceneitem.GetBaseID()), 0, 0, 0, uint32(len(user.items)))
	}
}

func (sf *MainPack) addBody(baseid uint32, reducedam uint32, maxdam uint32) {
	prop := sf.user.GetBodyProp()
	if prop == nil {
		return
	}

	space := sf.user.GetSpace().(*Scene)
	if prop.Baseid != 0 && prop.Reducedam != 0 {
		GetRefreshItemMgr(space).dropArmor(prop.Baseid, prop.Reducedam, linmath.RandXZ(sf.user.GetPos(), 0.5))
	}

	prop.Baseid = baseid
	prop.Reducedam = reducedam
	prop.Maxreduce = maxdam
	sf.user.SetBodyProp(prop)
	sf.user.SetBodyPropDirty()
	// log.Info("增加防弹衣", sf.user.GetID(), baseid)

	var level uint32 = 0
	if baseid == 1201 {
		level = 1
	} else if baseid == 1203 {
		level = 2
	} else if baseid == 1205 {
		level = 3
	}
	if space != nil {
		user, ok := space.GetEntityByDBID("Player", sf.user.GetDBID()).(*RoomUser)
		if ok {
			user.tlogBattleFlow(behavetype_pickup, uint64(baseid), 0, level, 0, uint32(len(user.items)))
		}
	}
}

func (sf *MainPack) addHead(baseid uint32, reducedam uint32, maxdam uint32) {
	prop := sf.user.GetHeadProp()
	if prop == nil {
		return
	}

	space := sf.user.GetSpace().(*Scene)
	if prop.Itemid != 0 && prop.Reducedam != 0 {
		GetRefreshItemMgr(space).dropArmor(prop.Itemid, prop.Reducedam, linmath.RandXZ(sf.user.GetPos(), 0.5))
	}

	// 更新幻化id
	if u, ok := sf.user.(*RoomUser); ok {
		packProp := u.GetHeadWearInGame()
		if baseid == 1202 {
			prop.Baseid = packProp.GetFirst()
		} else if baseid == 1204 {
			prop.Baseid = packProp.GetSecond()
		} else if baseid == 1206 {
			prop.Baseid = packProp.GetThird()
		} else {
			prop.Baseid = baseid
		}
	}
	if prop.Baseid == 0 {
		prop.Baseid = baseid
	}
	prop.Itemid = baseid
	prop.Reducedam = reducedam
	prop.Maxreduce = maxdam
	sf.user.SetHeadProp(prop)
	sf.user.SetHeadPropDirty()

	var level uint32 = 0
	if baseid == 1202 {
		level = 1
	} else if baseid == 1204 {
		level = 2
	} else if baseid == 1206 {
		level = 3
	}
	if space != nil {
		user, ok := space.GetEntityByDBID("Player", sf.user.GetDBID()).(*RoomUser)
		if ok {
			user.tlogBattleFlow(behavetype_pickup, uint64(baseid), 0, level, 0, uint32(len(user.items)))
		}
	}
}

//AddItem 增加道具
func (sf *MainPack) AddItem(sceneitem *Item) uint32 {
	if sceneitem == nil {
		return 0
	}

	if 0 == sceneitem.count {
		return 0
	}
	space := sf.user.GetSpace().(*Scene)
	if !space.ISceneData.onPickItem(sf.user.GetDBID(), sceneitem) {
		return 0
	}
	base := sceneitem.base
	baseid := sceneitem.GetBaseID()
	if base.Type <= ItemTypeWeapon {
		sf.add_gun(sceneitem)
		return 1
	} else if base.Type == ItemTypePack {
		sf.addBag(baseid)
		if d, ok := excel.GetItem(uint64(sf.user.GetBodyProp().Baseid)); ok {
			if d.AddValue == "1" {
				sf.user.CastRpcToAllClient("UpdateArmor", sf.user.GetID(), uint32(3), false)
			}
		}
		return 1
	} else if base.Type == ItemTypeArmor && base.Subtype == ItemSubtypeBody {
		sf.addBody(baseid, sceneitem.reducedam, uint32(base.Reducedam))
		if base.AddValue == "1" { //穿上吉利服
			sf.user.SetIsWearingGilley(1)
			sf.user.CastRpcToAllClient("UpdateArmor", sf.user.GetID(), uint32(2), true)
		} else {
			sf.user.SetIsWearingGilley(0)
		}
		return 1
	} else if base.Type == ItemTypeArmor && base.Subtype == ItemSubtypeHelmet {
		sf.addHead(baseid, sceneitem.reducedam, uint32(base.Reducedam))
		if d, ok := excel.GetItem(uint64(sf.user.GetBodyProp().Baseid)); ok {
			if d.AddValue == "1" {
				sf.user.CastRpcToAllClient("UpdateArmor", sf.user.GetID(), uint32(1), false)
			}
		}
		return 1
	} else if base.Type == ItemTypeConsume && base.Subtype == ItemSubtypeAdditem {
		addlist := strings.Split(base.Additem, ",")
		if len(addlist) != 2 {
			return 0
		}

		rand.Seed(time.Now().UnixNano())
		randlist := strings.Split(addlist[1], ":")
		if len(randlist) != 2 {
			return 0
		}
		min := common.StringToInt(randlist[0])
		max := common.StringToInt(randlist[1])
		randnum := min
		if max > min {
			randnum += rand.Intn(max-min) + 1
		}

		return sf.add_item(common.StringToUint32(addlist[0]), uint32(randnum))
	} else if base.Type == ItemTypeCar {
		//log.Warn("不能拾取车")
		return 0
	}

	return sf.add_item(baseid, sceneitem.count)
}

// ExchangeGun 换枪
func (sf *MainPack) ExchangeGun(useweapon uint64) {
	_, ok := sf.equips[useweapon]
	if !ok {
		return
	}

	downWeapon := sf.equips[sf.useweapon]
	sf.useweapon = useweapon
	sf.user.CastRPCToMe("ExchangeGunRet")

	sf.user.SetChracterMapDataInfo(sf.user.fillMapData())
	sf.user.SetChracterMapDataInfoDirty()

	sf.spareGunSightNotify(nil)
	sf.shotgunbullet = make(map[uint64]int)
	if downWeapon.base.Type == 5 { // 玩家换下火箭筒 做额外通知
		sf.user.CastRpcToAllClient("UpdateRPG", sf.user.GetID())
	}
}

//UseObject 使用道具
func (sf *MainPack) UseObject(thisid uint64) {
	// log.Info("使用道具", thisid)
	for _, v := range sf.items {
		if v.thisid == thisid {
			if v.base.Type == ItemTypeConsume {
				sf.user.AddEffect(v.GetBaseID(), uint32(v.base.Effect))
				return
			}

			log.Warn("Base type is not ok, type: ", v.base.Type)
			return
		}
	}
}

//GetItemByThisid 根据thisid获取背包中的道具
func (sf *MainPack) GetItemByThisid(thisid uint64) *Item {
	for _, v := range sf.items {
		if v.thisid == thisid {
			return v
		}
	}

	return nil
}

//GetItemNum get own item num
func (sf *MainPack) GetItemNum(baseid uint32) uint32 {
	var ret uint32
	for _, v := range sf.items {
		if v.GetBaseID() == baseid {
			ret += v.count
		}
	}

	return ret
}

func (sf *MainPack) removeItem(baseid uint32, num uint32) {
	for _, v := range sf.items {
		if v.GetBaseID() == baseid {
			if num == 0 {
				return
			}

			if v.count > num {
				v.count -= num
				v.RefreshToMe(sf.user)
				return
			}

			num -= v.count
			v.count = 0
			delete(sf.items, v.thisid)
			sf.RemoveObjectNotifyMe(v.thisid)
		}
	}
}

// canReform 配件能否装备在武器上
func canReform(baseid uint32, reformid uint32) bool {
	base, ok := excel.GetGun(uint64(baseid))
	if !ok {
		return false
	}

	reformlist := strings.Split(base.Reformitems, ",")
	for _, str := range reformlist {
		if reformid == common.StringToUint32(str) {
			return true
		}
	}

	return false
}

func (sf *MainPack) exchangeReformGun(gunid uint64, reformid uint32, othergun uint64) {
	_, ok := sf.equips[gunid]
	_, otherok := sf.equips[othergun]
	if !ok || !otherok {
		return
	}

	var find bool
	for _, reform := range sf.equips[gunid].gunreform {
		if reform == reformid {
			find = true
			break
		}
	}
	if !find {
		return
	}

	if !canReform(sf.equips[othergun].GetBaseID(), reformid) {
		return
	}

	var havebaseid uint32
	itemreform, ok := excel.GetItem(uint64(reformid))
	if !ok {
		return
	}
	for _, gun := range sf.equips {
		if gun.thisid == othergun {
			indexs := getGunReformIndexs(gun, itemreform.Subtype)
			if len(indexs) >= 1 {
				havebaseid = gun.gunreform[indexs[0]]
			}
			break
		}
	}

	// if havebaseid != 0 && !canReform(sf.equips[gunid].GetBaseID(), havebaseid) {
	// 	return
	// }

	for _, gun := range sf.equips {
		if gun.thisid == othergun {

			for index, baseid := range gun.gunreform {
				if baseid == havebaseid {
					gun.gunreform[index] = reformid
					break
				}
			}

			if havebaseid == 0 {
				gun.gunreform = append(gun.gunreform, reformid)
			}

			sf.RefreshGunNotifyAll(gun)
			break
		}
	}

	for _, gun := range sf.equips {
		if gun.thisid == gunid {
			for index, baseid := range gun.gunreform {
				if baseid == reformid {

					if havebaseid == 0 {
						gun.gunreform = append(gun.gunreform[:index], gun.gunreform[index+1:]...)
					} else {
						if canReform(gun.GetBaseID(), havebaseid) {
							gun.gunreform[index] = havebaseid
						} else {
							gun.gunreform = append(gun.gunreform[:index], gun.gunreform[index+1:]...)
							sf.add_item(havebaseid, 1)
						}
					}

					//校验子弹数量
					gun.checkBullet(sf.user)

					sf.RefreshGunNotifyAll(gun)
					break
				}
			}
		}
	}
}

// canAutoReform 武器能否自动装备配件
func (sf *MainPack) canAutoReform(gun *Item, baseid uint32) bool {
	if gun == nil {
		return false
	}

	base, ok := excel.GetItem(uint64(baseid))
	if !ok {
		return false
	}

	if canReform(gun.GetBaseID(), baseid) {
		indexs := getGunReformIndexs(gun, base.Subtype)

		if base.Subtype == ItemSubtypeSight && sf.canReformSpareGunSight(gun) {
			if len(indexs) < 2 {
				return true
			}
		} else {
			if len(indexs) < 1 {
				return true
			}
		}
	}

	return false
}

// canAutoReformAll 背包里的武器能否自动装备配件
func (sf *MainPack) canAutoReformAll(baseid uint32) bool {
	for _, equip := range sf.equips {
		if sf.canAutoReform(equip, baseid) {
			return true
		}
	}

	return false
}

// autoReform 自动装备武器配件
func (sf *MainPack) autoReform(baseid uint32) {
	base, ok := excel.GetItem(uint64(baseid))
	if !ok {
		return
	}

	main, sec := sf.getMainAndSecWeapon()

	if main != nil && sec != nil {
		if sf.canAutoReform(main, baseid) && sf.canAutoReform(sec, baseid) {
			if base.Subtype == ItemSubtypeSight {
				if len(getGunReformIndexs(main, ItemSubtypeSight)) == 1 && len(getGunReformIndexs(sec, ItemSubtypeSight)) == 0 {
					sf.equipGunReform(sec, baseid)
					return
				}
			}
		}
	}

	if main != nil {
		if sf.canAutoReform(main, baseid) {
			sf.equipGunReform(main, baseid)
			return
		}
	}

	if sec != nil {
		if sf.canAutoReform(sec, baseid) {
			sf.equipGunReform(sec, baseid)
			return
		}
	}
}

// equipGunReform 装备武器配件
func (sf *MainPack) equipGunReform(gun *Item, baseid uint32) {
	base, ok := excel.GetItem(uint64(baseid))
	if !ok {
		return
	}

	gun.gunreform = append(gun.gunreform, baseid)

	if base.Subtype == ItemSubtypeSight {
		exchangeGunSights(gun)
	}

	sf.RefreshGunNotifyAll(gun)
	sf.user.RPC(iserver.ServerTypeClient, "ReformSuccessNofity", gun.GetBaseID(), baseid)

	user, _ := sf.user.GetRealPtr().(*RoomUser)
	if user != nil {
		user.tlogBattleFlow(behavetype_install, uint64(baseid), 0, 0, 0, uint32(len(user.items)))
	}
}

func (sf *MainPack) usePackDropAutoReform(dropitem, equip *Item) {
	packreform := make([]*SortLightItem, 0)

	for _, item := range sf.items {
		baseid := uint32(item.base.Id)
		if item.base.Type != ItemTypeWeaponReform {
			continue
		}

		if canReform(equip.GetBaseID(), baseid) {
			tmp := &SortLightItem{}
			tmp.baseid = baseid
			tmp.light = uint32(item.base.Light)
			tmp.thisid = item.thisid
			packreform = append(packreform, tmp)
		}
	}

	sort.Sort(SliSortLightItem(packreform))

	for _, v := range packreform {
		base, ok := excel.GetItem(uint64(v.baseid))
		if !ok {
			continue
		}

		usedropreform := false
		for index, reformid := range dropitem.gunreform {
			reformbase, ok := excel.GetItem(uint64(reformid))
			if !ok {
				continue
			}

			if reformbase.Subtype == base.Subtype && reformbase.Light > base.Light && sf.canAutoReform(equip, reformid) {
				usedropreform = true
				dropitem.gunreform = append(dropitem.gunreform[:index], dropitem.gunreform[index+1:]...)
				sf.equipGunReform(equip, reformid)
				break
			}
		}

		if !usedropreform && sf.canAutoReform(equip, v.baseid) {
			//删除背包道具
			delete(sf.items, v.thisid)
			sf.RemoveObjectNotifyMe(v.thisid)

			sf.equipGunReform(equip, v.baseid)
		}
	}

	gunreform := make([]uint32, 0)
	for _, reform := range dropitem.gunreform {
		if sf.canAutoReform(equip, reform) {
			sf.equipGunReform(equip, reform)
		} else {
			gunreform = append(gunreform, reform)
		}
	}
	dropitem.gunreform = gunreform
}

func (sf *MainPack) usePackAutoReform(equip *Item) {
	packreform := make([]*SortLightItem, 0)

	for _, item := range sf.items {
		baseid := uint32(item.base.Id)
		if item.base.Type != ItemTypeWeaponReform {
			continue
		}

		if canReform(equip.GetBaseID(), baseid) {
			tmp := &SortLightItem{}
			tmp.baseid = baseid
			tmp.light = uint32(item.base.Light)
			tmp.thisid = item.thisid
			packreform = append(packreform, tmp)
		}
	}

	sort.Sort(SliSortLightItem(packreform))

	for _, v := range packreform {
		if sf.canAutoReform(equip, v.baseid) {
			//删除背包道具
			delete(sf.items, v.thisid)
			sf.RemoveObjectNotifyMe(v.thisid)
			sf.equipGunReform(equip, v.baseid)
		}
	}
}

// reformGun 装备枪支配件
// 装备倍镜时，pos表示倍镜位置(1：主倍镜 2：备用倍镜)
func (sf *MainPack) reformGun(gunthisid, reformthisid uint64, pos uint8) {
	for _, v := range sf.items {
		if v.thisid == reformthisid {
			for _, gun := range sf.equips {
				if gun.thisid == gunthisid {
					if v.base.Type != ItemTypeWeaponReform {
						log.Warn("Base type is not available, type: ", v.base.Type)
						return
					}

					if !canReform(gun.GetBaseID(), v.GetBaseID()) {
						log.Warn("Reform not match gun, gunbaseid: ", gun.GetBaseID(), " reformbaseid: ", v.GetBaseID())
						return
					}

					newbaseid := v.GetBaseID()
					oldbaseid := uint32(0)

					indexs := getGunReformIndexs(gun, v.base.Subtype)
					i := int(-1)

					if v.base.Subtype == ItemSubtypeSight {
						if pos == 1 { //装备主倍镜
							if len(indexs) == 1 {
								i = indexs[0]
							} else if len(indexs) >= 2 {
								i = indexs[len(indexs)-1]
							}
						} else if pos == 2 { //装备备用倍镜
							if !sf.canReformSpareGunSight(gun) {
								return
							}

							if len(indexs) >= 2 {
								i = indexs[0]
							}
						}
					} else {
						if len(indexs) >= 1 {
							i = indexs[0]
						}
					}

					if i != -1 { //替换现有配件
						oldbaseid = gun.gunreform[i]

						base, ok := excel.GetItem(uint64(oldbaseid))
						if !ok {
							return
						}

						v.base = base
						v.RefreshToMe(sf.user)

						gun.gunreform[i] = newbaseid
					} else { //装备新配件
						//删除背包道具
						delete(sf.items, reformthisid)
						sf.RemoveObjectNotifyMe(reformthisid)

						gun.gunreform = append(gun.gunreform, newbaseid)

						if pos == 2 {
							exchangeGunSights(gun)
						}
					}

					sf.RefreshGunNotifyAll(gun)
					sf.user.RPC(iserver.ServerTypeClient, "ReformSuccessNofity", gun.GetBaseID(), newbaseid)

					user, _ := sf.user.GetRealPtr().(*RoomUser)
					if user != nil {
						user.tlogBattleFlow(behavetype_install, uint64(oldbaseid), 0, 0, 0, uint32(len(user.items)))
					}

					return
				}
			}
		}
	}
}

func (sf *MainPack) unequipReform(baseid uint32, gunid uint64, drop bool) {
	for _, v := range sf.equips {
		if v.thisid == gunid {
			for index, reform := range v.gunreform {
				if reform == baseid {

					//修改改造属性
					v.gunreform = append(v.gunreform[:index], v.gunreform[index+1:]...)

					if drop {
						space := sf.user.GetSpace().(*Scene)
						GetRefreshItemMgr(space).dropItemByID(baseid, linmath.RandXZ(sf.user.GetPos(), 0.5))
						v.checkBulletDrop(sf.user, true)
					} else {
						//背包增加道具
						sf.add_item(baseid, 1)

						//校验子弹数量
						v.checkBullet(sf.user)
					}

					sf.RefreshGunNotifyAll(v)
					return
				}
			}

			return
		}
	}
}

// getMainAndSecWeapon 获取主武器和副武器
func (sf *MainPack) getMainAndSecWeapon() (*Item, *Item) {
	var main, sec *Item

	info := sf.user.fillMapData()
	if info == nil {
		return main, sec
	}

	weapon := info.GetWeapon()
	if weapon != nil {
		main = sf.equips[weapon.Thisid]
	}

	secWeapon := info.GetSecweapon()
	if secWeapon != nil {
		sec = sf.equips[secWeapon.Thisid]
	}

	return main, sec
}

// canReformSpareGunSight 是否可以装备备用倍镜
func (sf *MainPack) canReformSpareGunSight(gun *Item) bool {
	if gun == nil {
		return false
	}

	user, ok := sf.user.GetRealPtr().(*RoomUser)
	if !ok {
		return false
	}

	gunbase, ok := excel.GetGun(uint64(gun.GetBaseID()))
	if !ok {
		return false
	}

	if user.GetLevel() < uint32(gunbase.Backupscope) {
		return false
	}

	util := db.PlayerGoodsUtil(user.GetDBID())
	return util.IsOwnGoods(uint32(gunbase.DoubleSightID))
}

// getGunReformIndexs 获取指定类型的配件的下标
func getGunReformIndexs(gun *Item, subTyp uint64) []int {
	indexs := []int{}

	for index, baseid := range gun.gunreform {
		reformbase, ok := excel.GetItem(uint64(baseid))
		if !ok {
			continue
		}

		if reformbase.Subtype == subTyp {
			indexs = append(indexs, index)
		}
	}

	return indexs
}

// exchangeGunSights 交换reform列表中的主倍镜和备用倍镜
func exchangeGunSights(gun *Item) {
	indexs := getGunReformIndexs(gun, ItemSubtypeSight)
	if len(indexs) < 2 {
		return
	}

	tmp := gun.gunreform[indexs[0]]
	gun.gunreform[indexs[0]] = gun.gunreform[indexs[1]]
	gun.gunreform[indexs[1]] = tmp
}

// getSpareGunSight 获取备用倍镜
func (sf *MainPack) getSpareGunSight(gun *Item) (uint8, uint32) {
	if gun == nil {
		return 0, 0
	}

	spare := uint8(0)
	if sf.canReformSpareGunSight(gun) {
		spare = 1
	}

	indexs := getGunReformIndexs(gun, ItemSubtypeSight)
	if len(indexs) >= 2 {
		return spare, gun.gunreform[indexs[0]]
	}

	return spare, 0
}

// switchGunSight 一键切换倍镜
// gunthisid为0时表示在主武器和副武器之间切换倍镜
// gunthisid不为0时表示切换同一把枪的主倍镜和备用倍镜
func (sf *MainPack) switchGunSight(gunthisid uint64, pos1, pos2 int, para uint8) {
	log.Info("switchGunSight, gunthisid: ", gunthisid, " pos1: ", pos1, " pos2: ", pos2)

	var ret uint8 = 1
	defer func() {
		sf.user.RPC(iserver.ServerTypeClient, "SwitchGunSightRet", ret, para)
	}()

	if gunthisid != 0 {
		gun := sf.equips[gunthisid]
		if gun == nil {
			return
		}

		exchangeGunSights(gun)
		sf.RefreshGunNotifyAll(gun)

		ret = 0
		return
	}

	if pos1 == 0 || pos2 == 0 {
		return
	}

	main, sec := sf.getMainAndSecWeapon()
	if main == nil || sec == nil {
		return
	}

	indexs1 := getGunReformIndexs(main, ItemSubtypeSight)
	indexs2 := getGunReformIndexs(sec, ItemSubtypeSight)

	replace1 := true
	replace2 := true

	if len(indexs1) == 0 {
		replace1 = false
	} else if len(indexs1) == 1 {
		if pos1 == 1 {
			pos1 = indexs1[0]
		} else if pos1 == 2 {
			replace1 = false
		}
	} else if len(indexs1) == 2 {
		pos1 = indexs1[2-pos1]
	}

	if len(indexs2) == 0 {
		replace2 = false
	} else if len(indexs2) == 1 {
		if pos2 == 1 {
			pos2 = indexs2[0]
		} else if pos2 == 2 {
			replace2 = false
		}
	} else if len(indexs2) == 2 {
		pos2 = indexs2[2-pos2]
	}

	if !replace1 && !replace2 {
		return
	}

	if !replace1 {
		reformid := sec.gunreform[pos2]
		if !canReform(main.GetBaseID(), reformid) {
			return
		}

		main.gunreform = append(main.gunreform, reformid)
		sec.gunreform = append(sec.gunreform[:pos2], sec.gunreform[pos2+1:]...)

		if len(indexs1) == 1 {
			exchangeGunSights(main)
		}
	} else if !replace2 {
		reformid := main.gunreform[pos1]
		if !canReform(sec.GetBaseID(), reformid) {
			return
		}

		sec.gunreform = append(sec.gunreform, reformid)
		main.gunreform = append(main.gunreform[:pos1], main.gunreform[pos1+1:]...)

		if len(indexs2) == 1 {
			exchangeGunSights(sec)
		}
	} else {
		reformid1 := main.gunreform[pos1]
		reformid2 := sec.gunreform[pos2]

		if !canReform(main.GetBaseID(), reformid2) || !canReform(sec.GetBaseID(), reformid1) {
			return
		}

		main.gunreform[pos1] = reformid2
		sec.gunreform[pos2] = reformid1
	}

	sf.RefreshGunNotifyAll(main)
	sf.RefreshGunNotifyAll(sec)

	ret = 0
}

// canSetGunSightState 当前状态下能否开启或关闭倍镜
func (sf *MainPack) canSetGunSightState() bool {
	for _, v := range sf.equips {
		if v.thisid == sf.useweapon {
			item, ok := excel.GetGun(v.base.Id)
			if !ok {
				return false
			}
			if item.Defaultscope != 0 {
				return true
			}

			if len(getGunReformIndexs(v, ItemSubtypeSight)) > 0 {
				return true
			}
		}
	}

	return false
}

func (sf *MainPack) changeBullet(full bool) {
	for _, v := range sf.equips {
		if v.thisid == sf.useweapon {
			var bullet uint32
			gunbase, ok := excel.GetGun(uint64(v.GetBaseID()))
			if !ok {
				return
			}
			for _, item := range sf.items {
				if item.base.Type == ItemTypeBullet && item.GetBaseID() == uint32(gunbase.Consumebullet) {
					bullet += item.count
				}
			}

			clipcap := uint32(gunbase.Clipcap)
			deltas := common.StringToMapUint32(gunbase.Magazinedelta, "|", ";")

			for _, reform := range v.gunreform {
				clipcap += deltas[reform]
			}

			if v.bullet >= clipcap {
				log.Warn("Bullet is full, bullet: ", v.bullet, " clipcap: ", clipcap)

				space := sf.user.GetSpace().(*Scene)
				if space != nil {
					user, ok := space.GetEntityByDBID("Player", sf.user.GetDBID()).(*RoomUser)
					if ok {
						user.AdviceNotify(common.NotifyCommon, 19)
					}
				}
				return
			}

			if bullet == 0 {
				sf.user.SendChat("无可用子弹")
				return
			}

			var needbullet uint32
			if clipcap >= v.bullet {
				needbullet = clipcap - v.bullet
			}
			if !full {
				needbullet = 1
			}

			var consume uint32

			if bullet >= needbullet {
				v.bullet += needbullet
				consume = needbullet
			} else {
				v.bullet += bullet
				consume = bullet
			}

			v.fakebullet = v.bullet + 5

			if consume == 0 {
				return
			}

			del := []uint64{}
			for _, item := range sf.items {
				if item.base.Type == ItemTypeBullet && item.GetBaseID() == uint32(gunbase.Consumebullet) {
					if consume == 0 {
						break
					}

					if item.count >= consume {
						item.count -= consume
						consume = 0
					} else {
						consume -= item.count
						item.count = 0
					}

					if item.count == 0 {
						del = append(del, item.thisid)
					} else {
						item.RefreshToMe(sf.user)
					}

					// log.Warn("剩余子弹数量", v.thisid, v.bullet, needbullet)
				}
			}

			for _, v := range del {
				delete(sf.items, v)
				sf.RemoveObjectNotifyMe(v)
			}

			sf.RefreshGunNotifyAll(v)
			userR, okR := sf.user.(*RoomUser)
			if okR && userR != nil {
				userR.stateMgr.BreakRescue()
			}

			// log.Warn("换弹", v.bullet)
			return
		}
	}
}

func (sf *MainPack) dropObj(thisid uint64, dropSum uint32) {
	for _, v := range sf.items {
		if v.thisid == thisid {

			if dropSum > v.count {
				return
			}

			surplusSum := v.count - dropSum

			if surplusSum == 0 {
				sf.RemoveObjectNotifyMe(thisid)
				delete(sf.items, thisid)

				//玩家将正在加油的燃料扔掉，则打断加油
				user := sf.user.(*RoomUser)
				if user.addFuelThisid == thisid {
					user.BreakAddFuel(5)
				}
			} else {
				v.count = surplusSum
				v.RefreshToMe(sf.user)
			}

			dropItem := NewCopyItem(v)
			dropItem.count = dropSum

			space := sf.user.GetSpace().(*Scene)
			if space != nil {
				GetRefreshItemMgr(space).dropItem(sf.user.GetID(), dropItem, linmath.RandXZ(sf.user.GetPos(), 0.5))

				user, ok := space.GetEntityByDBID("Player", sf.user.GetDBID()).(*RoomUser)
				if ok {
					user.tlogBattleFlow(behavetype_throw, 0, uint64(v.GetBaseID()), 0, 0, uint32(len(user.items)))
				}
			}
			log.Warn("Drop weapon and add equip, thisid:", thisid, " baseid:", v.GetBaseID())
			return
		}
	}
}

func (sf *MainPack) dropArmor(baseid uint64) {
	base, ok := excel.GetItem(baseid)
	if !ok {
		return
	}

	space := sf.user.GetSpace().(*Scene)
	if space == nil {
		return
	}
	pos := linmath.RandXZ(sf.user.GetPos(), 0.5)

	if base.Type == ItemTypeArmor && base.Subtype == ItemSubtypeHelmet {
		prop := sf.user.GetHeadProp()
		if prop != nil {
			GetRefreshItemMgr(space).dropArmor(prop.Itemid, prop.Reducedam, pos)
			prop.Baseid = 0
			prop.Itemid = 0
			sf.user.SetHeadProp(prop)
			sf.user.SetHeadPropDirty()
		}
	} else if base.Type == ItemTypeArmor && base.Subtype == ItemSubtypeBody {
		prop := sf.user.GetBodyProp()
		if prop != nil {
			GetRefreshItemMgr(space).dropArmor(prop.Baseid, prop.Reducedam, pos)
			prop.Baseid = 0
			sf.user.SetBodyProp(prop)
			sf.user.SetBodyPropDirty()
			if base.AddValue == "1" { //穿上吉利服
				sf.user.SetIsWearingGilley(0)
				sf.user.CastRpcToAllClient("UpdateArmor", sf.user.GetID(), uint32(2), false)
			} else {
				sf.user.SetIsWearingGilley(0)
			}
		}
	} else {
		log.Warn("Armor is not droppable, baseid: ", baseid)
	}
}

func (sf *MainPack) dropGun(thisid uint64) {
	v, ok := sf.equips[thisid]
	if !ok {
		return
	}

	space := sf.user.GetSpace().(*Scene)
	if !space.canDrop(sf.user.GetDBID(), v) {
		return
	}
	if sf.autoreform {
		for _, other := range sf.equips {
			if other.thisid == thisid {
				continue
			}

			gunreform := make([]uint32, 0)
			for _, reform := range v.gunreform {
				if sf.canAutoReform(other, reform) {
					sf.equipGunReform(other, reform)
				} else {
					gunreform = append(gunreform, reform)
				}
			}
			v.gunreform = gunreform
		}

		if len(v.gunreform) != 0 {
			v.gunreform = sf.sortByLight(v.gunreform)
		}

		gunreform := make([]uint32, 0)
		for _, reform := range v.gunreform {
			if sf.LeftCell() > 0 {
				sf.add_item(reform, 1)
			} else {
				gunreform = append(gunreform, reform)
			}
		}
		v.gunreform = gunreform
	}

	v.checkBullet(sf.user)

	v.clearReform(space, sf.user)
	if space != nil {
		GetRefreshItemMgr(space).dropItem(sf.user.GetID(), v, linmath.RandXZ(sf.user.GetPos(), 0.5))

		user, ok := space.GetEntityByDBID("Player", sf.user.GetDBID()).(*RoomUser)
		if ok {
			user.tlogBattleFlow(behavetype_throw, 0, uint64(v.GetBaseID()), 0, 0, uint32(len(user.items))) // BattleFlow 战场流水表
		}
	}
	delete(sf.equips, thisid)

	proto := &protoMsg.DropGunNotify{}
	proto.Uid = sf.user.GetID()
	var weapon uint8

	if thisid == sf.useweapon {
		sf.useweapon = 0
		for _, weapon := range sf.equips {
			sf.useweapon = weapon.thisid
			proto.Use = weapon.fillInfo()
		}
	} else {
		proto.Use = sf.equips[sf.useweapon].fillInfo()
		weapon = 1
	}

	sf.user.RPC(iserver.ServerTypeClient, "DropGunSightNotify", weapon, thisid)
	sf.user.CastRPCToAllClient("DropGunNotify", proto)

	sf.user.SetChracterMapDataInfo(sf.user.fillMapData())
	sf.user.SetChracterMapDataInfoDirty()
}

// OnDeath 角色死亡的时候生成箱子
func (sf *MainPack) OnDeath() (uint64, linmath.Vector3) {
	tb := sf.user.GetSpace().(*Scene)
	droppos := sf.user.GetPos()
	droppos = getCanPutHeight(tb, droppos)

	//itemID := server.GetEntityTempID()
	itemID := tb.GetEntityTempID()
	sceneitem := &SceneItem{
		id:       itemID,
		pos:      droppos,
		itemid:   1502,
		item:     NewItem(1502, 1),
		haveobjs: make(map[uint32]*Item),
	}

	for _, v := range sf.items {
		tb.mapindex++

		item := NewCopyItem(v)
		if item != nil {
			sceneitem.haveobjs[tb.mapindex] = item
		}
	}

	rand.Seed(time.Now().UnixNano())
	for _, v := range sf.equips {

		//ai死亡掉落子弹随机
		if sf.user.isAI() {
			num := v.bullet / 3
			if num > 1 {
				percent := rand.Intn(int(num)) + 1
				v.bullet = uint32(percent)
			}
		} else {
			for _, reform := range v.gunreform {
				item := NewItem(reform, 1)
				if item != nil {
					tb.mapindex++
					sceneitem.haveobjs[tb.mapindex] = item
				}
			}

			v.gunreform = make([]uint32, 0)
		}

		tb.mapindex++
		item := NewCopyItem(v)
		if item != nil {
			sceneitem.haveobjs[tb.mapindex] = item
		}
	}

	headprop := sf.user.GetHeadProp()
	if headprop != nil && headprop.Itemid != 0 {
		tb.mapindex++
		item := NewArmorItem(headprop.Itemid, headprop.Reducedam)
		if item != nil {
			sceneitem.haveobjs[tb.mapindex] = item
		}
	}

	bodyprop := sf.user.GetBodyProp()
	if bodyprop != nil && bodyprop.Baseid != 0 {
		tb.mapindex++
		item := NewArmorItem(bodyprop.Baseid, bodyprop.Reducedam)
		if item != nil {
			sceneitem.haveobjs[tb.mapindex] = item
		}
	}

	packprop := sf.user.GetBackPackProp()
	if packprop != nil && packprop.Itemid != 0 {
		tb.mapindex++
		item := NewArmorItem(packprop.Itemid, 0)
		if item != nil {
			sceneitem.haveobjs[tb.mapindex] = item
		}
	}

	if len(sceneitem.haveobjs) == 0 {
		return 0, linmath.Vector3_Zero()
	}

	tb.mapitem[sceneitem.id] = sceneitem
	tb.AddTinyEntity("Item", itemID, "")

	tb.BroadDeathDropBox(sf.user.GetID(), itemID)

	return itemID, droppos
}

//UpdateCell 更新格子
func (sf *MainPack) UpdateCell() {
	proto := &protoMsg.RefreshPackCellNotify{
		Num: sf.GetMaxCell(),
	}
	sf.user.RPC(iserver.ServerTypeClient, "RefreshPackCellNotify", proto)
}

// GetMaxCell 获取最大格子数
func (sf *MainPack) GetMaxCell() uint32 {
	res := sf.initCells

	prop := sf.user.GetBackPackProp()
	if prop == nil {
		return res
	}

	if base, ok := excel.GetItem(uint64(prop.GetItemid())); ok {
		if base.Type == ItemTypePack {
			res += uint32(base.Addcell)
		}
	}

	return res
}

//LeftCell 剩余格子
func (sf *MainPack) LeftCell() uint32 {
	ret := sf.GetMaxCell()

	if ret >= uint32(len(sf.items)) {
		return ret - uint32(len(sf.items))
	}

	return 0
}

func (sf *MainPack) isHeadShot(defender iDefender) bool {
	var headrate uint32
	var shotdistance float32
	for _, v := range sf.equips {
		if v.thisid == sf.useweapon {
			gunbase, ok := excel.GetGun(uint64(v.GetBaseID()))
			if !ok {
				return false
			}

			if sf.isMeleeWeaponUse() {
				return false
			}

			headrate += uint32(gunbase.Headshotrate)
			shotdistance = gunbase.Distance

			for _, reform := range v.gunreform {
				itembase, ok := excel.GetItem(uint64(reform))
				if !ok {
					return false
				}
				headrate += uint32(itembase.Headratedelta)
			}
		}
	}

	distance := common.Distance(sf.user.GetPos(), defender.GetPos())
	if shotdistance != 0 {
		headrate = uint32(float64(headrate) * math.Pow(float64(distance/shotdistance), -0.5))
		//log.Debug("爆头率计算", headrate)
	}
	if headrate > 100 {
		headrate = 100
	}

	percent := rand.Intn(100) + 1
	if uint32(percent) <= headrate {
		return true
	}

	return false
}

func (sf *MainPack) addBag(baseid uint32) {
	prop := sf.user.GetBackPackProp()
	if prop == nil {
		return
	}

	space := sf.user.GetSpace().(*Scene)
	if prop.Itemid != 0 {
		GetRefreshItemMgr(space).dropArmor(prop.Itemid, 0, linmath.RandXZ(sf.user.GetPos(), 0.5))
	}
	// 更新幻化id
	if u, ok := sf.user.(*RoomUser); ok {
		packProp := u.GetPackWearInGame()
		if baseid == 1601 {
			prop.Baseid = packProp.GetSecond()
		} else if baseid == 1602 {
			prop.Baseid = packProp.GetThird()
		} else if baseid == 1603 {
			prop.Baseid = packProp.GetFirst()
		} else {
			prop.Baseid = baseid
		}
	}
	if prop.Baseid == 0 {
		prop.Baseid = baseid
	}
	prop.Itemid = baseid
	sf.user.SetBackPackProp(prop)
	sf.user.SetBackPackPropDirty()
	if sf.user.GetType() == "Player" {
		user, _ := sf.user.GetRealPtr().(*RoomUser)
		if user != nil {
			sf.UpdateCell()
		}
	}

	var level uint32 = 0
	if baseid == 1601 {
		level = 1
	} else if baseid == 1602 {
		level = 2
	} else if baseid == 1603 {
		level = 3
	}
	if space != nil {
		user, ok := space.GetEntityByDBID("Player", sf.user.GetDBID()).(*RoomUser)
		if ok {
			user.tlogBattleFlow(behavetype_pickup, uint64(baseid), 0, level, 0, uint32(len(user.items)))
		}
	}
}

//GetInUseWeapon get use weapon
func (sf *MainPack) GetInUseWeapon() uint32 {
	for _, v := range sf.equips {
		if v.thisid == sf.useweapon {
			return v.GetBaseID()
		}
	}

	return 0
}

//GetWeaponDistance get weapon shot distance
func (sf *MainPack) GetWeaponDistance() float32 {
	for _, v := range sf.equips {
		if v.thisid == sf.useweapon {
			base, ok := excel.GetGun(uint64(v.GetBaseID()))
			if !ok {
				return 1
			}

			return base.Distance + 1
		}
	}

	return 1
}

//gunID, sightID, silenceID,magazineID
func (sf *MainPack) getUseGunInfo() (gunID uint64, sightID uint32, silenceID uint32, magazineID uint32, stockID uint32, handleID uint32) {
	for _, v := range sf.equips {
		if v.thisid == sf.useweapon {
			if !sf.isMeleeWeaponUse() {
				gunID = v.base.Id

				for _, reformID := range v.gunreform {
					reformInfo, ok := excel.GetItem(uint64(reformID))
					if !ok {
						continue
					}

					if reformInfo.Subtype == 3 {
						sightID = reformID
					} else if reformInfo.Subtype == 5 {
						silenceID = reformID
					} else if reformInfo.Subtype == 6 {
						magazineID = reformID
					} else if reformInfo.Subtype == 4 {
						stockID = reformID
					} else if reformInfo.Subtype == 7 {
						handleID = reformID
					}
				}
			}
			break
		}
	}
	return gunID, sightID, silenceID, magazineID, stockID, handleID
}

// isBazookaWeaponUse 是否正在使用火箭筒武器
func (sf *MainPack) isBazookaWeaponUse() bool {
	if sf.useweapon == 0 {
		return false
	}

	for _, v := range sf.equips {
		if v.thisid == sf.useweapon {
			if v.base.Type == ItemTypeBazooka {
				return true
			}
		}
	}

	return false
}

func (sf *MainPack) isMeleeWeaponUse() bool {
	if sf.useweapon == 0 {
		return true
	}

	for _, v := range sf.equips {
		if v.thisid == sf.useweapon {
			if v.base.Type == ItemTypeMeleeWeapon {
				return true
			}
		}
	}

	return false
}

//GetMeleeWeaponUse if is melle weapon use
func (sf *MainPack) GetMeleeWeaponUse() uint32 {
	if sf.useweapon == 0 {
		return 1009
	}

	for _, v := range sf.equips {
		if v.thisid == sf.useweapon {
			if v.base.Type == ItemTypeMeleeWeapon {
				return v.GetBaseID()
			}
		}
	}

	return 0
}

//CanAddFullPack 是否能叠加包裹内道具
func (sf *MainPack) CanAddFullPack(base excel.ItemData, num uint32) bool {

	baseid := uint32(base.Id)
	addmax := uint32(1)
	if base.Additem != "" {
		// log.Error("mmmmmmmmm: ", base.Additem)
		addlist := strings.Split(base.Additem, ",")
		if len(addlist) != 2 {
			return false
		}

		randlist := strings.Split(addlist[1], ":")
		if len(randlist) != 2 {
			return false
		}

		baseid = common.StringToUint32(addlist[0])
		addmax = common.StringToUint32(randlist[1])
	} else {
		if base.Addlimit <= 1 {
			// log.Error("########## ", base.Addlimit, base.Id)
			return false
		}

		addmax = num
	}

	var left uint32
	for _, v := range sf.items {
		if baseid == v.GetBaseID() && v.count < v.GetMaxNum() {
			left += v.GetMaxNum() - v.count
			// log.Error("@@@@@ ", v.GetMaxNum(), v.count)
		}
	}

	return left >= addmax
}

//RefreshGunNotifyAll 广播刷新枪
func (sf *MainPack) RefreshGunNotifyAll(v *Item) {
	if v == nil {
		return
	}

	updateproto := &protoMsg.RefreshGunNotify{}
	updateproto.Objs = v.fillInfo()
	updateproto.Useweapon = sf.useweapon
	updateproto.Uid = sf.user.GetID()

	sf.user.CastRPCToAllClient("RefreshGunNotify", updateproto)
	sf.spareGunSightNotify(v)

	sf.user.SetChracterMapDataInfo(sf.user.fillMapData())
	sf.user.SetChracterMapDataInfoDirty()
}

// reconnectGunNotify 断线重连后同步枪支状态
func (sf *MainPack) reconnectGunNotify() {
	main, sec := sf.getMainAndSecWeapon()
	if main != nil {
		sf.RefreshGunNotifyAll(main)
	}

	if sec != nil {
		sf.spareGunSightNotify(sec)
	}
}

// spareGunSightNotify 通知客户端备用倍镜
// gun为nil时同时通知主武器和副武器的备用倍镜
// gun不为nil时只通知指定武器的备用倍镜
func (sf *MainPack) spareGunSightNotify(gun *Item) {
	if gun != nil {
		var weapon uint8
		if gun.thisid != sf.useweapon {
			weapon = 1
		}

		spare, baseid := sf.getSpareGunSight(gun)
		sf.user.RPC(iserver.ServerTypeClient, "SpareGunSightNotify", weapon, gun.thisid, spare, baseid)

		return
	}

	main, sec := sf.getMainAndSecWeapon()

	if main != nil {
		spare, baseid := sf.getSpareGunSight(main)
		sf.user.RPC(iserver.ServerTypeClient, "SpareGunSightNotify", uint8(0), main.thisid, spare, baseid)
	}

	if sec != nil {
		spare, baseid := sf.getSpareGunSight(sec)
		sf.user.RPC(iserver.ServerTypeClient, "SpareGunSightNotify", uint8(1), sec.thisid, spare, baseid)
	}
}

//RemoveObjectNotifyMe 通知删除道具
func (sf *MainPack) RemoveObjectNotifyMe(thisid uint64) {
	proto := &protoMsg.RemoveObjectNotify{
		Thisid: thisid,
	}

	sf.user.RPC(iserver.ServerTypeClient, "RemoveObjectNotify", proto)
}

type SliSortLightItem []*SortLightItem
type SortLightItem struct {
	baseid uint32
	light  uint32
	thisid uint64
}

func (a SliSortLightItem) Len() int      { return len(a) }
func (a SliSortLightItem) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a SliSortLightItem) Less(i, j int) bool {
	if a[i].light > a[j].light {
		return true
	} else if a[i].light == a[j].light {
		return a[i].baseid > a[j].baseid
	}

	return false
}

func (sf *MainPack) sortByLight(gunreform []uint32) []uint32 {

	sortreform := make([]*SortLightItem, 0)
	for _, v := range gunreform {
		base, ok := excel.GetItem(uint64(v))
		if ok {
			tmp := &SortLightItem{}
			tmp.baseid = v
			tmp.light = uint32(base.Light)
			sortreform = append(sortreform, tmp)
		}
	}

	sort.Sort(SliSortLightItem(sortreform))

	ret := make([]uint32, 0)
	for _, v := range sortreform {
		ret = append(ret, v.baseid)
	}

	return ret
}

func (sf *MainPack) SendAllObj() {
	msg := &protoMsg.RefreshObjectListNotify{}
	for _, v := range sf.items {
		msg.Obj = append(msg.Obj, v.fillInfo())
	}

	sf.user.RPC(iserver.ServerTypeClient, "RefreshObjectListNotify", msg)
}

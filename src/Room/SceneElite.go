package main

import (
	"common"
	"excel"
	"fmt"
	"math/rand"
	"protoMsg"
	"time"
	"zeus/iserver"
)

type EliteScene struct {
	SceneData
	boxlist        *protoMsg.DropBoxPosList
	eliteAttack    bool     // 是否精英友军伤害
	normalAttack   bool     // 是否普通玩家友军伤害
	initElite      bool     // 是否初始完精英完成
	eliteEquip     []uint32 // 精英玩家的初始装备
	notifyNum      int64    // 初始化预期通知次数
	eliteMember    map[uint64]bool
	equipPosNotify int64
}

func NewEliteScene(sc *Scene) *EliteScene {
	scenedata := &EliteScene{}
	d, ok := excel.GetSystem(common.System_EliteAttackN)
	if ok {
		scenedata.normalAttack = d.Value == 1
	}
	d, ok = excel.GetSystem(common.System_EliteAttackE)
	if ok {
		scenedata.eliteAttack = d.Value == 1
	}
	scenedata.eliteEquip = scenedata.getEliteEquip()
	scenedata.eliteMember = make(map[uint64]bool)
	scenedata.scene = sc
	scenedata.boxlist = &protoMsg.DropBoxPosList{}
	return scenedata
}

func (scenedata *EliteScene) canAttack(user IRoomChracter, defendid uint64) bool {
	defender, ok := scenedata.scene.GetEntity(defendid).(IRoomChracter)
	if !ok {
		return true
	}
	uElite := user.GetGamerType() == RoomUserTypeElite
	dElite := defender.GetGamerType() == RoomUserTypeElite
	if uElite && dElite {
		return scenedata.eliteAttack
	}
	if !uElite && !dElite {
		return scenedata.normalAttack
	}
	return true
}

// getEliteEquip 初始化精英装备
func (scenedata *EliteScene) getEliteEquip() []uint32 {
	var r []uint32
	l, ok := excel.GetSystem(common.System_EliteHead)
	if ok {
		r = append(r, uint32(l.Value))
	}
	l, ok = excel.GetSystem(common.System_EliteBody)
	if ok {
		r = append(r, uint32(l.Value))
	}
	l, ok = excel.GetSystem(common.System_ElitePack)
	if ok {
		r = append(r, uint32(l.Value))
	}
	return r
}

//doLoop
func (scenedata *EliteScene) doLoop() {
	if scenedata.scene.parachuteTime == 0 {
		return
	}

	now := time.Now().Unix()
	//刷新精英
	scenedata.eliteFresh(now)
	//通知刷新精英
	scenedata.eliteFreshNotify(now)
	//通知精英装备位置
	if now >= scenedata.equipPosNotify {
		scenedata.NotifyEliteEquip()
		l, ok := excel.GetSystem(common.System_ElitePosFreshTime)
		if ok {
			scenedata.equipPosNotify = now + int64(l.Value)
		} else {
			scenedata.equipPosNotify = now + 60
		}
	}
	// 精英友军伤害
	scenedata.freshAttackState(now)
}

func (scenedata *EliteScene) refreshSpecialBox(refreshcount uint64) {
	tb := scenedata.scene
	base, ok := excel.GetMaprule(refreshcount)
	if !ok {
		return
	}

	for _, v := range scenedata.boxlist.Items {
		tb.RemoveTinyEntity(v.Id)
	}

	msg := &protoMsg.DropBoxPosList{}
	for i := uint64(0); i < base.Elitebox && i < 10; i++ {
		tmp := &protoMsg.DropBoxPos{}
		droppos := tb.GetCircleRamdomIndexPos(i, GetRefreshZoneMgr(tb).nextsafecenter, GetRefreshZoneMgr(tb).nextsaferadius)
		droppos = getXZCanPutHeight(tb, droppos)
		boxid := uint32(common.GetTBSystemValue(common.System_RefreshSpecialBox))
		entityid := GetRefreshItemMgr(tb).dropItemByID(boxid, droppos)

		tmp.Id = entityid
		tmp.Pos = &protoMsg.Vector3{
			X: droppos.X,
			Y: droppos.Y,
			Z: droppos.Z,
		}
		msg.Items = append(msg.Items, tmp)
		tb.Debug("生成特殊空投", " 坐标", droppos)
	}

	scenedata.boxlist = msg
	tb.TravsalEntity("Player", func(e iserver.IEntity) {
		if e == nil {
			return
		}

		if user, ok := e.(*RoomUser); ok {
			user.RPC(iserver.ServerTypeClient, "DropBoxPosList", msg)
		}
	})
}

func (scenedata *EliteScene) resendData(user *RoomUser) {
	scenedata.SceneData.resendData(user)

	user.RPC(iserver.ServerTypeClient, "DropBoxPosList", scenedata.boxlist)
}

//刷新玩家之间的攻击状态
func (scenedata *EliteScene) freshAttackState(now int64) {
	if scenedata.eliteAttack {
		return
	}
	d, ok := excel.GetSystem(common.System_EliteAttackOpen)
	if !ok {
		return
	}
	if now >= scenedata.scene.parachuteTime+int64(d.Value) {
		scenedata.eliteAttack = true
	}
	if !scenedata.eliteAttack && len(scenedata.eliteMember) == scenedata.scene.getMemSum() {
		scenedata.eliteAttack = true
	}
	if scenedata.eliteAttack {
		scenedata.scene.chatNotify(1, "精英学员之间伤害开启，请小心！")
		scenedata.clearElitePos()
	}
}

//eliteFreshNotify 通知精英刷新的时间
func (scenedata *EliteScene) eliteFreshNotify(now int64) {
	if scenedata.initElite {
		return
	}
	d, ok := excel.GetSystem(common.System_EliteFreshTime)
	if !ok {
		return
	}
	l, e := excel.GetSystem(common.System_EliteNum)
	if !e {
		return
	}
	freshTime := scenedata.scene.parachuteTime + int64(d.Value) - 10
	if now >= freshTime+scenedata.notifyNum {
		s := fmt.Sprintf("%d秒后将随机产生%d名精英学员，获得强力装备！", 10-scenedata.notifyNum, l.Value)
		scenedata.scene.chatNotify(1, s)
		scenedata.notifyNum++
	}
}

//eliteFresh 刷新精英
func (scenedata *EliteScene) eliteFresh(now int64) {
	if scenedata.initElite {
		return
	}
	d, ok := excel.GetSystem(common.System_EliteFreshTime)
	if !ok {
		return
	}

	if now >= scenedata.scene.parachuteTime+int64(d.Value) {
		var eliteWeapon []uint32
		for _, item := range excel.GetItemMap() {
			if item.IsGreatWarrior == 1 {
				eliteWeapon = append(eliteWeapon, uint32(item.Id))
			}
		}
		l, e := excel.GetSystem(common.System_EliteNum)
		if !e {
			return
		}
		num := int(l.Value)
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		scenedata.scene.TravsalEntity("Player", func(entity iserver.IEntity) {
			user, ok := entity.(*RoomUser)
			if !ok {
				return
			}
			if len(scenedata.eliteMember) >= num {
				return
			}
			//随机武器
			i := r.Intn(len(eliteWeapon))
			weaponItem := NewItem(eliteWeapon[i], 1)
			user.AddItem(weaponItem)
			if gun, ok := excel.GetGun(weaponItem.base.Id); ok {
				if bullet, ok := excel.GetItem(gun.Consumebullet); ok {
					d, ok := excel.GetSystem(common.System_EliteWeaponBullet)
					if ok {
						bNum := uint32(bullet.Addlimit * d.Value)
						user.AddItem(NewItemBullet(uint32(bullet.Id), bNum))
					}
				}
			}

			//添加装备
			for _, v := range scenedata.eliteEquip {
				if v == 0 {
					continue
				}
				user.AddItem(NewItem(v, 1))
			}
			if user.GetGamerType() != RoomUserTypeElite{
				scenedata.broadEliteState(user)
			}
			user.SetGamerType(RoomUserTypeElite)
			scenedata.eliteMember[user.GetID()] = true
		})

		scenedata.scene.chatNotify(1, "精英学员已产生，将其击败或者拾取空投可成为精英学员")
		if !scenedata.eliteAttack {
			scenedata.scene.chatNotify(1, "精英学员之间伤害免疫")
		}
		scenedata.initElite = true
		if len(scenedata.eliteMember) < num {
			scenedata.setEliteAi(num - len(scenedata.eliteMember))
		}
	}
}

// SetEliteAi 补充精英AI
func (scenedata *EliteScene) setEliteAi(num int) {
	var eliteWeapon uint32
	for _, item := range excel.GetItemMap() {
		if item.IsGreatWarrior == 1 {
			eliteWeapon = uint32(item.Id)
			break
		}
	}
	i := 0
	scenedata.scene.TravsalEntity("AI", func(entity iserver.IEntity) {
		if i >= num {
			return
		}
		Ai, ok := entity.(*RoomAI)
		if !ok {
			return
		}
		Ai.AddItem(NewItem(eliteWeapon, 1))
		for _, v := range scenedata.eliteEquip {
			if v == 0 {
				continue
			}
			Ai.AddItem(NewItem(v, 1))
		}
		if Ai.GetGamerType() != RoomUserTypeElite{
			scenedata.broadEliteState(Ai)
		}
		Ai.SetGamerType(RoomUserTypeElite)
		scenedata.eliteMember[Ai.GetID()] = true
		i++
	})
}

//NotifyEliteEquip 通知精英装备位置
func (scenedata *EliteScene) NotifyEliteEquip() {
	if !scenedata.showElitePos() {
		return
	}
	notify := &protoMsg.EliteWeaponPosList{}
	for _, item := range scenedata.scene.mapitem {
		if scenedata.scene.GetTinyEntity(item.id) == nil {
			continue
		}
		isAdd := false
		if item.item.GetBaseID() == 1502 {
			for _, i := range item.haveobjs {
				if i.base.IsGreatWarrior == 1 {
					isAdd = true
					break
				}
			}
		}
		if item.item.base.IsGreatWarrior == 1 {
			isAdd = true
		}

		if isAdd {
			notify.Items = append(notify.Items, &protoMsg.EliteWeaponPos{
				Id: item.id,
				Pos: &protoMsg.Vector3{
					X: item.pos.X,
					Y: item.pos.Y,
					Z: item.pos.Z,
				},
			})
		}
	}
	for uid := range scenedata.eliteMember {
		user, ok := scenedata.scene.GetEntity(uid).(IRoomChracter)
		if !ok {
			continue
		}

		if user.GetGamerType() != RoomUserTypeElite {
			continue
		}
		var eliteWeapon uint64
		for _, eq := range user.GetEquips() {
			if eq.base.IsGreatWarrior == 1 {
				eliteWeapon = eq.thisid
				break
			}
		}
		pos := &protoMsg.EliteWeaponPos{
			Id: eliteWeapon,
			Pos: &protoMsg.Vector3{
				X: user.GetPos().X,
				Y: user.GetPos().Y,
				Z: user.GetPos().Z,
			},
		}
		notify.Items = append(notify.Items, pos)
	}
	if len(notify.Items) == 0 {
		return
	}
	scenedata.scene.BroadCastMsg(notify, "SyncEliteWeaponPos")
}

//onPickItem 捡道具 常规捡道具的额外处理
func (scenedata *EliteScene) onPickItem(uid uint64, item *Item) bool {
	if item.base.Type > ItemTypeWeapon { // 不是捡枪 通过
		return true
	}
	user, ok := scenedata.scene.GetEntityByDBID("Player", uid).(*RoomUser)
	if !ok {
		return true
	}

	if item.base.IsGreatWarrior != 1 { // 捡枪非精英武器通过
		if len(user.equips) >= 2 {
			used := user.equips[user.useweapon]
			if used != nil && used.base.IsGreatWarrior == 1 { // 当前使用精英武器 无法捡枪
				user.AdviceNotify(common.NotifyCommon, 62)
				return false
			}
		}
		return true
	}

	// 捡起精英武器 设置变化
	scenedata.eliteMember[user.GetID()] = true
	if user.GetGamerType() != RoomUserTypeElite{
		scenedata.broadEliteState(user)
	}
	user.SetGamerType(RoomUserTypeElite)
	scenedata.clearElitePos()
	return true
}

// onDeath 死亡回调
func (scenedata *EliteScene) onDeath(user IRoomChracter) {
	if user.GetGamerType() == RoomUserTypeElite {
		delete(scenedata.eliteMember, user.GetID())
	}
	scenedata.clearElitePos()
}

// canDrop 丢弃武器回调
func (scenedata *EliteScene) canDrop(uid uint64, item *Item) bool {
	user, ok := scenedata.scene.GetEntityByDBID("Player", uid).(*RoomUser)
	if !ok {
		return false
	}
	if item.base.IsGreatWarrior == 1 {
		user.AdviceNotify(common.NotifyCommon, 62)
		return false
	}
	return true
}

// showElitePos 是否显示精英位置
func (scenedata *EliteScene) showElitePos() bool {
	return !scenedata.eliteAttack
}

// clearElitePos 通知客户端清理精英位置
func (scenedata *EliteScene) clearElitePos() {
	if !scenedata.showElitePos() {
		scenedata.scene.TravsalEntity("Player", func(e iserver.IEntity) {
			if e == nil {
				return
			}

			if user, ok := e.(*RoomUser); ok {
				user.RPC(iserver.ServerTypeClient, "ClearEliteWeaponPos")
			}
		})
	}
}

func (scenedata *EliteScene) broadEliteState(user iserver.ICoordEntity) {
	scenedata.scene.TravsalAOI(user, func(ia iserver.ICoordEntity) {
		if ise, ok := ia.(iserver.IEntityStateGetter); ok {
			if ise.GetEntityState() != iserver.Entity_State_Loop {
				return
			}
			if ie, ok := ia.(iserver.IEntity); ok {
				ie.RPC(iserver.ServerTypeClient, "PlayerChangeGamerType", user.GetID(), uint32(RoomUserTypeElite))
			}
		}
	})
}

package main

import (
	"common"
	"excel"
	"reflect"
	"time"
	"zeus/iserver"

	log "github.com/cihub/seelog"
)

type iEffect interface {
	Update(u *RoomUser)
	Init(baseid uint32, u *RoomUser)
	GetState() uint32
	Close(u *RoomUser, isbreak bool)
}

type effect struct {
	iEffect
	state     uint32
	baseid    uint32
	begintime int64
}

func (sf *effect) Close(u *RoomUser, isbreak bool) {
	sf.state = 1

	if isbreak {
		u.RPC(iserver.ServerTypeClient, "BreakPotion")
	} else {
		u.RPC(iserver.ServerTypeClient, "FinishPotion")
	}

	log.Info("Close effect, name: ", u.GetName(), " uid: ", u.GetDBID(), " isbreak: ", isbreak)
}

func (sf *effect) GetState() uint32 {
	return sf.state
}

type effect1101 struct {
	effect
}

func (sf *effect1101) Init(baseid uint32, u *RoomUser) {
	sf.state = 0
	sf.baseid = baseid

	base, ok := excel.GetItem(uint64(sf.baseid))
	if !ok {
		return
	}
	processtime := base.Processtime
	sf.begintime = time.Now().UnixNano()/1000000 + int64(processtime)*1000 + 200

	state := u.GetPlayerState()
	if state.ActionState == ActionUseMedicine {
		u.RPC(iserver.ServerTypeClient, "ReuseObject", baseid)
	} else {
		state.ActionState = ActionUseMedicine
		state.SetParam2Uint32(baseid)
		state.SetModify(true)
	}
}

func (sf *effect1101) Update(u *RoomUser) {
	base, ok := excel.GetItem(uint64(sf.baseid))
	if !ok {
		return
	}

	if u.GetItemNum(sf.baseid) < 1 {
		return
	}

	if time.Now().UnixNano()/1000000 >= sf.begintime {
		sf.Close(u, false)

		user, _ := u.GetRealPtr().(*RoomUser)
		hplimit, _ := excel.GetSystem(common.System_AddHpLimit)
		if user != nil && user.GetHP() < uint32(hplimit.Value) {
			addhp := uint32(base.Effectparam)
			if user.GetHP()+addhp > uint32(hplimit.Value) {
				addhp = uint32(hplimit.Value) - user.GetHP()
			}
			user.DisposeAddHp(addhp)
			user.removeItem(sf.baseid, 1)
			log.Info("After first-aid, addhp: ", addhp, " hp: ", user.GetHP())

			user.sumData.IncrRecvitemUseNum()
			user.sumData.IncrPainkillerNum()
			log.Info("First-aid packet num add one")

		}
	}
}

type effect1112 struct {
	effect
}

func (sf *effect1112) Init(baseid uint32, u *RoomUser) {
	sf.state = 0
	sf.baseid = baseid

	base, ok := excel.GetItem(uint64(sf.baseid))
	if !ok {
		return
	}
	processtime := base.Processtime
	sf.begintime = time.Now().UnixNano()/1000000 + int64(processtime)*1000 + 200

	state := u.GetPlayerState()
	if state.ActionState == ActionUseMedicine {
		u.RPC(iserver.ServerTypeClient, "ReuseObject", baseid)
	} else {
		state.ActionState = ActionUseMedicine
		state.SetParam2Uint32(baseid)
		state.SetModify(true)
	}

}

func (sf *effect1112) Update(u *RoomUser) {
	base, ok := excel.GetItem(uint64(sf.baseid))
	if !ok {
		return
	}

	if u.GetItemNum(sf.baseid) < 1 {
		return
	}

	if time.Now().UnixNano()/1000000 >= sf.begintime {
		sf.Close(u, false)

		user, _ := u.GetRealPtr().(*RoomUser)
		hplimit, _ := excel.GetSystem(common.System_AddHpLimit)
		if user != nil && user.GetHP() < uint32(hplimit.Value) {
			addhp := uint32(base.Effectparam)
			if user.GetHP()+addhp > uint32(hplimit.Value) {
				addhp = uint32(hplimit.Value) - user.GetHP()
			}
			user.DisposeAddHp(addhp)
			user.removeItem(sf.baseid, 1)
			log.Info("Bandage effect: ", base.Effectparam)

			user.sumData.IncrRecvitemUseNum()
			user.sumData.IncrBandageNum()
			log.Info("Bandage num add one")

		}
	}
}

type effect1120 struct {
	effect
}

func (sf *effect1120) Init(baseid uint32, u *RoomUser) {
	sf.state = 0
	sf.baseid = baseid

	base, ok := excel.GetItem(uint64(sf.baseid))
	if !ok {
		return
	}
	processtime := base.Processtime
	sf.begintime = time.Now().UnixNano()/1000000 + int64(processtime)*1000 + 200

	state := u.GetPlayerState()
	if state.ActionState == ActionUseMedicine {
		u.RPC(iserver.ServerTypeClient, "ReuseObject", baseid)
	} else {
		state.ActionState = ActionUseMedicine
		state.SetParam2Uint32(baseid)
		state.SetModify(true)
	}
}

func (sf *effect1120) Update(u *RoomUser) {
	base, ok := excel.GetItem(uint64(sf.baseid))
	if !ok {
		return
	}

	if u.GetItemNum(sf.baseid) < 1 {
		return
	}

	if time.Now().UnixNano()/1000000 >= sf.begintime {
		sf.Close(u, false)

		user, _ := u.GetRealPtr().(*RoomUser)
		if user != nil {
			addhp := uint32(base.Effectparam)
			user.DisposeAddHp(addhp)
			user.removeItem(sf.baseid, 1)
			log.Info("Medical box effect: ", base.Effectparam)

			user.sumData.IncrRecvitemUseNum()
			user.sumData.IncrMedicalBoxNum()
			log.Info("Medical box num add one")
		}
	}
}

type effect1117 struct {
	effect
}

func (sf *effect1117) Init(baseid uint32, u *RoomUser) {
	sf.state = 0
	sf.baseid = baseid

	item, ok := excel.GetItem(uint64(sf.baseid))
	if !ok {
		return
	}
	processtime := item.Processtime
	sf.begintime = time.Now().UnixNano()/1000000 + int64(processtime)*1000 + 200

	state := u.GetPlayerState()
	if state.ActionState == ActionUseMedicine {
		u.RPC(iserver.ServerTypeClient, "ReuseObject", baseid)
	} else {
		state.ActionState = ActionUseMedicine
		state.SetParam2Uint32(baseid)
		state.SetModify(true)
	}

	log.Info("Drug notify, name: ", u.GetName(), " uid: ", u.GetDBID(), " time:", processtime)
}

func (sf *effect1117) Update(u *RoomUser) {
	base, ok := excel.GetItem(uint64(sf.baseid))
	if !ok {
		return
	}

	if u.GetItemNum(sf.baseid) < 1 {
		return
	}

	if time.Now().UnixNano()/1000000 >= sf.begintime {
		sf.Close(u, false)

		user, _ := u.GetRealPtr().(*RoomUser)
		if user != nil {
			user.energy.addEnergy(uint32(base.Effectparam))
			user.removeItem(sf.baseid, 1)
			log.Info("Energy bar effect: ", base.Effectparam)

			user.sumData.IncrSpeedNum() //玩家加速次数自增
		}
	}
}

type effectM struct {
	user *RoomUser

	effects    map[uint32]iEffect
	effecttype map[uint32]reflect.Type
}

func initEffectM(u *RoomUser) *effectM {
	effectm := &effectM{user: u}
	effectm.effects = make(map[uint32]iEffect)

	effectm.effecttype = make(map[uint32]reflect.Type)
	effectm.regType(1101, &effect1101{})
	effectm.regType(1112, &effect1112{})
	effectm.regType(1117, &effect1117{})
	effectm.regType(1120, &effect1120{})

	return effectm
}

func (sf *effectM) regType(id uint32, protoType iEffect) {
	t := reflect.TypeOf(protoType)
	sf.effecttype[id] = t
}

func (sf *effectM) newType(id uint32) interface{} {
	proto, ok := sf.effecttype[id]
	if !ok {
		log.Warn("Effect type not exist, id: ", id)
		return nil
	}

	e := reflect.New(proto.Elem()).Interface()
	return e
}

func (sf *effectM) BreakEffect(isBreak bool) {
	actionState := sf.user.GetActionState()
	if !isBreak && (actionState == ActionNone || actionState == Potion) {
		//移动使用可回血道具 不打断药品使用
		for k, _ := range sf.user.SkillMgr.passiveEffect {
			if k == SE_MoveUseItem {
				return
			}
		}
	}

	for _, v := range sf.effects {
		v.Close(sf.user, true)

		sf.user.RPC(iserver.ServerTypeClient, "BreakUIProgressBar", uint32(1))
	}

	if actionState == Potion {
		sf.user.SetActionState(ActionNone)
	}
}

func (u *RoomUser) actionCanUseObject() bool {
	state := u.GetActionState()
	if state == ActionNone || state == Potion || state == ShootPrepare || state == MeleePrepare {
		return true
	}

	return false
}

func (sf *effectM) CanUseObject(baseid uint32) bool {
	if _, ok := sf.effects[baseid]; ok {
		log.Warn("The same kind equip is using")
		return false
	}

	if !sf.user.StateM.CanUseObject() {
		log.Warn("State is not available")
		return false
	}

	if !sf.user.actionCanUseObject() {
		log.Warn("Action state is not available")
		return false
	}

	base, ok := excel.GetItem(uint64(baseid))
	if !ok {
		return false
	}

	if base.Subtype == 1 {
		hplimit, ok := excel.GetSystem(common.System_AddHpLimit)
		if !ok {
			return false
		}

		if sf.user.GetHP() >= uint32(hplimit.Value) {
			log.Warn("Reach hp limit, id: ", sf.user.GetID(), " hp: ", sf.user.GetHP())
			if baseid == 1112 {
				sf.user.AdviceNotify(common.NotifyCommon, 26) //通知客户端(生命值高于75，不能使用此物品)
				log.Info("Hp is higher than 75, can not use bandage, id: ", sf.user.GetID(), " hp: ", sf.user.GetHP())
			} else if baseid == 1101 {
				sf.user.AdviceNotify(common.NotifyCommon, 12) //通知客户端(生命值高于75，不能使用此物品)
				log.Info("Hp is higher than 75, can not use first-aid packet, id: ", sf.user.GetID(), " hp: ", sf.user.GetHP())
			}
			return false
		}
	}

	return true
}

func (sf *effectM) AddEffect(baseid uint32, id uint32) {
	if !sf.CanUseObject(baseid) {
		log.Info("AddEffect failed, baseid is not available ", baseid)
		//sf.user.SetActionState(0)
		return
	}

	e := sf.newType(id)
	if e == nil {
		return
	}

	effect := e.(iEffect)
	if effect == nil {
		return
	}

	//强制删除其它效果
	sf.BreakEffect(true)

	effect.Init(baseid, sf.user)
	sf.effects[baseid] = effect

	// 打断救援
	sf.user.stateMgr.BreakRescue()
}

func (sf *effectM) Update() {
	del := make([]uint32, 0)
	for k, v := range sf.effects {
		if v.GetState() == 1 {
			del = append(del, k)
		}
	}

	for _, v := range del {
		delete(sf.effects, v)
		//log.Info("删除效果", v)
	}

	for _, v := range sf.effects {
		v.Update(sf.user)
	}
}

func (sf *effectM) forceClear(isBreak bool) {
	for k, v := range sf.effects {
		if isBreak {
			v.Close(sf.user, false)
		}

		delete(sf.effects, k)
		//log.Debug("强制删除效果  ", k)
	}
}

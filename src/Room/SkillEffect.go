package main

import (
	"common"
	"excel"
	"protoMsg"
	"strings"
	"time"
	"zeus/iserver"
	"zeus/linmath"

	log "github.com/cihub/seelog"
)

// SkillEffectValue 效果值
type SkillEffectValue struct {
	effectID uint32 //效果ID
	value    uint64

	persistTime uint32 //持续时间
	interval    uint32 //间隔时间

	putTime     int64
	nextPutTime int64

	entityID uint64 // 效果生成的实体ID
}

// ISkillEffect 技能效果
type ISkillEffect interface {
	IsStart(param *SkillEffectValue) bool
	Start(user *RoomUser, param *SkillEffectValue)
	IsEnd(param *SkillEffectValue) bool
	End(user *RoomUser, param *SkillEffectValue)
}

type SkillEffect struct {
}

func (skillEffect *SkillEffect) IsStart(param *SkillEffectValue) bool {
	if time.Now().Unix() == param.nextPutTime && param.interval != 0 { //处理持续时间内每间隔一定时间执行一次的效果
		return true
	}

	return false
}

func (skilleffect *SkillEffect) Start(user *RoomUser, param *SkillEffectValue) {
}

func (skillEffect *SkillEffect) IsEnd(param *SkillEffectValue) bool {
	if time.Now().Unix() == param.putTime+int64(param.persistTime) { //处理持续时间过期的效果
		return true
	}

	return false
}

func (skilleffect *SkillEffect) End(user *RoomUser, param *SkillEffectValue) {
}

// SkillEffect_Recoil 枪械后坐力降低百分比
type SkillEffect_Recoil struct {
	SkillEffect
}

func (s *SkillEffect_Recoil) Start(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	param.nextPutTime += int64(param.interval)
	user.SkillData.skillEffectDam[SE_Recoil] += uint32(param.value)
	user.Debug("枪械后坐力降低百分比:", param.value)
}

func (s *SkillEffect_Recoil) End(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	if user.SkillData.skillEffectDam[SE_Recoil] > uint32(param.value) {
		user.SkillData.skillEffectDam[SE_Recoil] -= uint32(param.value)
	} else {
		user.SkillData.skillEffectDam[SE_Recoil] = 0
	}
	user.Debug("枪械后坐力恢复百分比:", param.value)
}

// SkillEffect_ShrinkDam 毒圈伤害降低百分比
type SkillEffect_ShrinkDam struct {
	SkillEffect
}

func (s *SkillEffect_ShrinkDam) Start(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	param.nextPutTime += int64(param.interval)
	user.SkillData.skillEffectDam[SE_ShrinkDam] += uint32(param.value)
	user.Debug("毒圈伤害降低百分比:", param.value)
}

func (s *SkillEffect_ShrinkDam) End(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	if user.SkillData.skillEffectDam[SE_ShrinkDam] > uint32(param.value) {
		user.SkillData.skillEffectDam[SE_ShrinkDam] -= uint32(param.value)
	} else {
		user.SkillData.skillEffectDam[SE_ShrinkDam] = 0
	}
	user.Debug("毒圈伤害恢复百分比:", param.value, " dam:", user.SkillData.skillEffectDam[SE_ShrinkDam])
}

// SkillEffect_TurnHp 按固定值扣血
type SkillEffect_TurnHp struct {
	SkillEffect
}

func (s *SkillEffect_TurnHp) Start(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	param.nextPutTime += int64(param.interval)
	user.DisposeSubHp(InjuredInfo{num: uint32(param.value), injuredType: mephitis, isHeadshot: false})
	user.Debug("按固定值扣血:", param.value)
}

// SkillEffect_WeaponDam 武器伤害增加百分比
type SkillEffect_WeaponDam struct {
	SkillEffect
}

func (s *SkillEffect_WeaponDam) Start(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	param.nextPutTime += int64(param.interval)
	user.SkillData.skillEffectDam[SE_WeaponDam] += uint32(param.value)
	user.Debug("武器攻击力增加百分比", param.value)
}

func (s *SkillEffect_WeaponDam) End(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	if user.SkillData.skillEffectDam[SE_WeaponDam] > uint32(param.value) {
		user.SkillData.skillEffectDam[SE_WeaponDam] -= uint32(param.value)
	} else {
		user.SkillData.skillEffectDam[SE_WeaponDam] = 0
	}
	user.Debug("武器攻击力减少百分比", param.value)
}

// SkillEffect_FireSpeed 射击速度增加百分比
type SkillEffect_FireSpeed struct {
	SkillEffect
}

func (s *SkillEffect_FireSpeed) Start(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	param.nextPutTime += int64(param.interval)
	user.SkillData.skillEffectDam[SE_FireSpeed] += uint32(param.value)
	user.Debug("射击速度增加百分比:", param.value)
}

func (s *SkillEffect_FireSpeed) End(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	if user.SkillData.skillEffectDam[SE_FireSpeed] > uint32(param.value) {
		user.SkillData.skillEffectDam[SE_FireSpeed] -= uint32(param.value)
	} else {
		user.SkillData.skillEffectDam[SE_FireSpeed] = 0
	}
	user.Debug("射击速度减少百分比:", param.value)
}

// SkillEffect_Shield 生成护盾吸收固定值伤害
type SkillEffect_Shield struct {
	SkillEffect
}

func (s *SkillEffect_Shield) Start(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	param.nextPutTime += int64(param.interval)
	user.SkillData.skillEffectDam[SE_Shield] += uint32(param.value)
	user.Debug("生成护盾吸收固定值伤害:", param.value)
}

func (s *SkillEffect_Shield) End(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	if user.SkillData.skillEffectDam[SE_Shield] > uint32(param.value) {
		user.SkillData.skillEffectDam[SE_Shield] -= uint32(param.value)
	} else {
		user.SkillData.skillEffectDam[SE_Shield] = 0

		delete(user.SkillMgr.initiveEffect, SE_Shield)
		user.SkillMgr.refreshSkillEffect()
	}
	user.Debug("减少护盾吸收固定值伤害:", param.value, " skillEffectDam:", user.SkillData.skillEffectDam[SE_Shield])
}

// SkillEffect_MapTag 地图标记一次一定范围内敌人和空投位置
type SkillEffect_MapTag struct {
	SkillEffect
}

func (s *SkillEffect_MapTag) Start(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	param.nextPutTime += int64(param.interval)

	skillSystemData, ok := excel.GetSkillSystem(common.SkillSystem_UAVArea)
	if !ok {
		return
	}

	space := user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	playerPosList := &protoMsg.DropBoxPosList{}
	dropBoxPosList := &protoMsg.DropBoxPosList{}

	for id, ok := range space.members {
		if !ok {
			continue
		}

		if id == user.GetID() {
			continue
		}

		player, _ := space.GetEntity(id).(IRoomChracter)
		if player == nil || player.IsDead() {
			continue
		}

		if space.teamMgr.isTeam {
			if team, ok := space.teamMgr.teams[user.teamid]; ok {
				isTeam := false
				for _, memid := range team {
					if memid == player.GetDBID() {
						isTeam = true
						break
					}
				}

				if isTeam {
					continue
				}
			}
		}

		distance := common.Distance(user.GetPos(), player.GetPos())
		if distance > float32(skillSystemData.Value1) {
			continue
		}

		pos := &protoMsg.DropBoxPos{
			Id: id,
			Pos: &protoMsg.Vector3{
				X: player.GetPos().X,
				Y: player.GetPos().Y,
				Z: player.GetPos().Z,
			},
		}
		playerPosList.Items = append(playerPosList.Items, pos)
	}

	for k, v := range space.refreshitem.boxlist {
		if v == nil {
			continue
		}

		boxPos := linmath.Vector3{v.Pos.X, v.Pos.Y, v.Pos.Z}
		distance := common.Distance(user.GetPos(), boxPos)
		if distance > float32(skillSystemData.Value1) {
			continue
		}

		pos := &protoMsg.DropBoxPos{
			Id:  k,
			Pos: v.Pos,
		}
		dropBoxPosList.Items = append(dropBoxPosList.Items, pos)
	}

	user.RPC(iserver.ServerTypeClient, "SyncNearPos", playerPosList, dropBoxPosList)
	user.Debug("地图标记一次一定范围内敌人和空投位置! playerPosList:", playerPosList, " dropBoxPosList:", dropBoxPosList)
}

func (s *SkillEffect_MapTag) End(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	user.RPC(iserver.ServerTypeClient, "ClearNearPos")
	user.Debug("取消地图标记一次一定范围内敌人和空投位置！")
}

// SkillEffect_RandomItem 随机生成一个道具
type SkillEffect_RandomItem struct {
	SkillEffect
}

func (s *SkillEffect_RandomItem) Start(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	param.nextPutTime += int64(param.interval)

	skillEffectData, ok := excel.GetSkillEffect(SE_RandomItem)
	if !ok {
		return
	}

	var weight uint32
	var weightList, itemIDList, itemNumList []uint32
	params := strings.Split(skillEffectData.Param, ";")
	for _, v := range params {
		param := strings.Split(v, "-")
		if len(param) != 3 {
			user.Debug("SkillEffect excel Param err id:", SE_RandomItem, " params:", params)
			continue
		}

		itemID := common.StringToUint32(param[0])
		itemNum := common.StringToUint32(param[1])
		itemWeight := common.StringToUint32(param[2])

		weight += itemWeight
		weightList = append(weightList, itemWeight)
		itemIDList = append(itemIDList, itemID)
		itemNumList = append(itemNumList, itemNum)
	}

	index := common.WeightRandom(weight, weightList)
	if index == 0 {
		user.Error("WeightRandom err index:", index)
		return
	}
	index--
	if index >= uint32(len(itemIDList)) || index >= uint32(len(itemNumList)) {
		return
	}

	space := user.GetSpace().(*Scene)
	if space != nil {
		for i := uint32(0); i < itemNumList[index]; i++ {
			space.refreshitem.dropItemByID(itemIDList[index], linmath.RandXZ(user.GetPos(), 0.5))
		}
	}

	user.Debug("随机生成一个道具 id:", itemIDList[index], " num:", itemNumList[index], " index:", index, " itemIDList:", itemIDList, " weight:", weight, " weightList:", weightList)
}

// SkillEffect_AddHpCap 血量上限增加百分比
type SkillEffect_AddHpCap struct {
	SkillEffect
}

func (s *SkillEffect_AddHpCap) Start(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	param.nextPutTime += int64(param.interval)

	maxHP := user.GetRoleBaseMaxHP()
	hpValue := uint32(float32(param.value) / 100.0 * float32(maxHP))
	maxHP += hpValue
	user.SetMaxHP(maxHP)
	user.SetHP(user.GetHP() + hpValue)

	user.Debug("血量上限增加百分比", param.value)
}

func (s *SkillEffect_AddHpCap) End(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	maxHP := user.GetRoleBaseMaxHP()
	hpValue := uint32(float32(param.value) / 100.0 * float32(maxHP))
	user.SetMaxHP(maxHP)
	user.DisposeSubHp(InjuredInfo{num: -hpValue, injuredType: losthp, isHeadshot: false})

	user.Debug("血量上限减少百分比", param.value)
}

// SkillEffect_MoveUseItem 移动使用可回血道具
type SkillEffect_MoveUseItem struct {
	SkillEffect
}

func (s *SkillEffect_MoveUseItem) Start(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	param.nextPutTime += int64(param.interval)
	user.SkillData.skillEffectDam[SE_MoveUseItem] = 1
	user.Debug("移动使用可回血道具:", param.value)
}

func (s *SkillEffect_MoveUseItem) End(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	delete(user.SkillData.skillEffectDam, SE_MoveUseItem)
	user.Debug("移动不使用可回血道具!")
}

// SkillEffect_RescueTime 救援时间减少百分比
type SkillEffect_RescueTime struct {
	SkillEffect
}

func (s *SkillEffect_RescueTime) Start(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	param.nextPutTime += int64(param.interval)
	user.SkillData.skillEffectDam[SE_RescueTime] += uint32(param.value)
	user.Debug("救援时间减少百分比:", param.value)
}

func (s *SkillEffect_RescueTime) End(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	if user.SkillData.skillEffectDam[SE_RescueTime] > uint32(param.value) {
		user.SkillData.skillEffectDam[SE_RescueTime] -= uint32(param.value)
	} else {
		user.SkillData.skillEffectDam[SE_RescueTime] = 0
	}
	user.Debug("救援时间恢复百分比:", param.value)
}

// SkillEffect_HpRecover 救援血量恢复至总值百分比
type SkillEffect_HpRecover struct {
	SkillEffect
}

func (s *SkillEffect_HpRecover) Start(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	param.nextPutTime += int64(param.interval)
	user.SkillData.skillEffectDam[SE_HpRecover] += uint32(param.value)
	user.Debug("救援血量恢复至总值百分比:", param.value)
}

func (s *SkillEffect_HpRecover) End(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	if user.SkillData.skillEffectDam[SE_HpRecover] > uint32(param.value) {
		user.SkillData.skillEffectDam[SE_HpRecover] -= uint32(param.value)
	} else {
		user.SkillData.skillEffectDam[SE_HpRecover] = 0
	}
	user.Debug("救援血量恢复至总值百分比:", param.value)
}

// SkillEffect_DamReduce 自身伤害削减百分比
type SkillEffect_DamReduce struct {
	SkillEffect
}

func (s *SkillEffect_DamReduce) Start(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	param.nextPutTime += int64(param.interval)
	user.SkillData.skillEffectDam[SE_DamReduce] += uint32(param.value)
	user.Debug("自身伤害削减百分比:", param.value)
}

func (s *SkillEffect_DamReduce) End(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	if user.SkillData.skillEffectDam[SE_DamReduce] > uint32(param.value) {
		user.SkillData.skillEffectDam[SE_DamReduce] -= uint32(param.value)
	} else {
		user.SkillData.skillEffectDam[SE_DamReduce] = 0
	}
	user.Debug("自身伤害恢复百分比:", param.value)
}

// SkillEffect_VehicleNoDam 免疫载具碰撞
type SkillEffect_VehicleNoDam struct {
	SkillEffect
}

func (s *SkillEffect_VehicleNoDam) Start(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	param.nextPutTime += int64(param.interval)
	user.SkillData.skillEffectDam[SE_VehicleNoDam] += uint32(param.value)
	user.Debug("免疫载具碰撞:", param.value)
}

func (s *SkillEffect_VehicleNoDam) End(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	if user.SkillData.skillEffectDam[SE_VehicleNoDam] > uint32(param.value) {
		user.SkillData.skillEffectDam[SE_VehicleNoDam] -= uint32(param.value)
	} else {
		user.SkillData.skillEffectDam[SE_VehicleNoDam] = 0
	}
	user.Debug("不免疫载具碰撞:", param.value)
}

// SkillEffect_OilLoss 开车油耗减少百分比
type SkillEffect_OilLoss struct {
	SkillEffect
}

func (s *SkillEffect_OilLoss) Start(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	param.nextPutTime += int64(param.interval)
	user.SkillData.skillEffectDam[SE_OilLoss] += uint32(param.value)
	user.Debug("开车油耗减少百分比:", param.value)
}

func (s *SkillEffect_OilLoss) End(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	if user.SkillData.skillEffectDam[SE_OilLoss] > uint32(param.value) {
		user.SkillData.skillEffectDam[SE_OilLoss] -= uint32(param.value)
	} else {
		user.SkillData.skillEffectDam[SE_OilLoss] = 0
	}
	user.Debug("开车油耗恢复百分比:", param.value)
}

// SkillEffect_ReleaseFog 原地瞬时释放烟雾
type SkillEffect_ReleaseFog struct {
	SkillEffect
}

func (s *SkillEffect_ReleaseFog) Start(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	space := user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	param.entityID = space.GetEntityTempID()
	dropitem := &SceneItem{
		id:       param.entityID,
		pos:      user.GetPos(),
		itemid:   2105,
		item:     NewItem(2105, 1),
		haveobjs: make(map[uint32]*Item),
	}

	user.RPC(iserver.ServerTypeClient, "NotPlayerIgnoreFog", param.entityID) //主角自己不看到释放的烟雾技能特效

	space.mapitem[dropitem.id] = dropitem
	space.AddTinyEntity("Item", param.entityID, "")

	user.Debug("原地瞬时释放烟雾:", param.entityID)
}

func (s *SkillEffect_ReleaseFog) End(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	space := user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	if err := space.RemoveTinyEntity(param.entityID); err != nil {
		user.Error("RemoveTinyEntity failed! err: ", err, " param.entityID:", param.entityID)
		return
	}

	user.Debug("原地瞬时取消烟雾:", param.entityID)
}

// SkillEffect_CallEffect 每杀一人触发一个效果
type SkillEffect_CallEffect struct {
	SkillEffect
}

func (s *SkillEffect_CallEffect) Start(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	param.nextPutTime += int64(param.interval)

	skillData, ok := excel.GetSkill(param.value)
	if !ok {
		return
	}
	if skillData.SkillType != common.Skill_Call {
		return
	}

	effects := strings.Split(skillData.SkillEffect, ";")
	for _, j := range effects {
		effect := strings.Split(j, "-")
		if len(effect) != 4 {
			user.Error("skill excel PassiveEffect err! skillId:", param.value, " PassiveEffect:", j)
			continue
		}

		skillEffect := &SkillEffectValue{}
		skillEffect.effectID = common.StringToUint32(effect[0])
		skillEffect.value = common.StringToUint64(effect[1])
		skillEffect.persistTime = common.StringToUint32(effect[2])
		skillEffect.interval = common.StringToUint32(effect[3])

		switch skillData.Active {
		case common.Skill_Passive:
			user.addPassiveEffect(skillEffect)
		case common.Skill_Initive:
			user.addInitiveEffect(skillEffect)
		default:
			user.Error("skill excel Active err! Active=:", skillData.Active)
		}
	}
	user.Debug("每杀一人触发一个效果!")
}

// SkillEffect_VehicleDamReduce 载具内玩家伤害降低百分比
type SkillEffect_VehicleDamReduce struct {
	SkillEffect
}

func (s *SkillEffect_VehicleDamReduce) Start(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	param.nextPutTime += int64(param.interval)
	user.SkillData.skillEffectDam[SE_VehicleDamReduce] += uint32(param.value)
	user.Debug("载具内玩家伤害降低百分比:", param.value)
}

func (s *SkillEffect_VehicleDamReduce) End(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	if user.SkillData.skillEffectDam[SE_VehicleDamReduce] > uint32(param.value) {
		user.SkillData.skillEffectDam[SE_VehicleDamReduce] -= uint32(param.value)
	} else {
		user.SkillData.skillEffectDam[SE_VehicleDamReduce] = 0
	}
	user.Debug("载具内玩家伤害恢复百分比:", param.value)
}

// SkillEffect_RandomItem2 随机生成一个道具2
type SkillEffect_RandomItem2 struct {
	SkillEffect
}

func (s *SkillEffect_RandomItem2) Start(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	param.nextPutTime += int64(param.interval)

	skillEffectData, ok := excel.GetSkillEffect(SE_RandomItem2)
	if !ok {
		return
	}

	var weight uint32
	var weightList, itemIDList, itemNumList []uint32
	params := strings.Split(skillEffectData.Param, ";")
	for _, v := range params {
		param := strings.Split(v, "-")
		if len(param) != 3 {
			user.Debug("SkillEffect excel Param err id:", SE_RandomItem2, " params:", params)
			continue
		}

		itemID := common.StringToUint32(param[0])
		itemNum := common.StringToUint32(param[1])
		itemWeight := common.StringToUint32(param[2])

		weight += itemWeight
		weightList = append(weightList, itemWeight)
		itemIDList = append(itemIDList, itemID)
		itemNumList = append(itemNumList, itemNum)
	}

	index := common.WeightRandom(weight, weightList)
	if index == 0 {
		user.Error("WeightRandom err index:", index)
		return
	}
	index--
	if index >= uint32(len(itemIDList)) || index >= uint32(len(itemNumList)) {
		return
	}

	space := user.GetSpace().(*Scene)
	if space != nil {
		for i := uint32(0); i < itemNumList[index]; i++ {
			space.refreshitem.dropItemByID(itemIDList[index], linmath.RandXZ(user.GetPos(), 0.5))
		}
	}

	user.Debug("随机生成一个道具 id:", itemIDList[index], " num:", itemNumList[index], " index:", index, " itemIDList:", itemIDList, " weight:", weight, " weightList:", weightList)
}

// SkillEffect_RandomItem3 随机生成一个道具3
type SkillEffect_RandomItem3 struct {
	SkillEffect
}

func (s *SkillEffect_RandomItem3) Start(user *RoomUser, param *SkillEffectValue) {
	if user == nil || param == nil {
		log.Error("Start user or param nil!")
		return
	}

	param.nextPutTime += int64(param.interval)

	skillEffectData, ok := excel.GetSkillEffect(SE_RandomItem3)
	if !ok {
		return
	}

	var weight uint32
	var weightList, itemIDList, itemNumList []uint32
	params := strings.Split(skillEffectData.Param, ";")
	for _, v := range params {
		param := strings.Split(v, "-")
		if len(param) != 3 {
			user.Debug("SkillEffect excel Param err id:", SE_RandomItem3, " params:", params)
			continue
		}

		itemID := common.StringToUint32(param[0])
		itemNum := common.StringToUint32(param[1])
		itemWeight := common.StringToUint32(param[2])

		weight += itemWeight
		weightList = append(weightList, itemWeight)
		itemIDList = append(itemIDList, itemID)
		itemNumList = append(itemNumList, itemNum)
	}

	index := common.WeightRandom(weight, weightList)
	if index == 0 {
		user.Error("WeightRandom err index:", index)
		return
	}
	index--
	if index >= uint32(len(itemIDList)) || index >= uint32(len(itemNumList)) {
		return
	}

	space := user.GetSpace().(*Scene)
	if space != nil {
		for i := uint32(0); i < itemNumList[index]; i++ {
			space.refreshitem.dropItemByID(itemIDList[index], linmath.RandXZ(user.GetPos(), 0.5))
		}
	}

	user.Debug("随机生成一个道具 id:", itemIDList[index], " num:", itemNumList[index], " index:", index, " itemIDList:", itemIDList, " weight:", weight, " weightList:", weightList)
}

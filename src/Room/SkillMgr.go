package main

import (
	"common"
	"db"
	"excel"
	"protoMsg"
	"strings"
	"time"
	"zeus/iserver"
)

type SkillMgr struct {
	id     uint64
	modeID uint32
	user   *RoomUser

	initiveSkillData map[uint32]excel.SkillData   //主动技能 key是技能ID
	passiveSkillData map[uint32]excel.SkillData   //被动技能 key是技能ID
	initiveEffect    map[uint32]*SkillEffectValue //主动效果 key是效果ID
	passiveEffect    map[uint32]*SkillEffectValue //被动效果 key是效果ID

	preUseTime map[uint32]int64 //上次释放技能时间

	*SkillData
}

func initSkillMgr(user *RoomUser) *SkillMgr {
	skillMgr := new(SkillMgr)
	skillMgr.id = uint64(user.GetRoleModel())
	space := user.GetSpace().(*Scene)
	if space != nil {
		skillMgr.modeID = space.GetMatchMode()
	}
	skillMgr.user = user

	skillMgr.initiveSkillData = make(map[uint32]excel.SkillData)
	skillMgr.passiveSkillData = make(map[uint32]excel.SkillData)
	skillMgr.initiveEffect = make(map[uint32]*SkillEffectValue)
	skillMgr.passiveEffect = make(map[uint32]*SkillEffectValue)

	skillMgr.preUseTime = make(map[uint32]int64)

	skillMgr.SkillData = initSkillData(user)
	skillMgr.init()

	return skillMgr
}

func (s *SkillMgr) init() {
	if !s.judgeSkillOpenMode() {
		return
	}

	info := &db.RoleSkillInfo{}
	info.InitiveSkill = make(map[uint32]map[uint32]uint32)
	info.PassiveSkill = make(map[uint32]map[uint32]uint32)

	skillUtil := db.PlayerSkillUtil(s.user.GetDBID(), s.id)
	if err := skillUtil.GetRoleSkill(info); err != nil {
		s.user.Error("GetRoleSkill fail! err:", err)
		return
	}

	for _, v := range info.InitiveSkill[s.modeID] {
		base, ok := excel.GetSkill(uint64(v))
		if !ok {
			s.user.Error("GetSkill fail! id:", v)
			continue
		}

		s.initiveSkillData[v] = base
	}

	for _, v := range info.PassiveSkill[s.modeID] {
		base, ok := excel.GetSkill(uint64(v))
		if !ok {
			s.user.Error("GetSkill fail! id:", v)
			continue
		}

		s.passiveSkillData[v] = base
	}

	s.user.Debug("id:", s.id, " init info:", *info, " initiveSkillData:", s.initiveSkillData, " passiveSkillData:", s.passiveSkillData)
}

// checkCold 是否技能冷却中
func (s *SkillMgr) checkCold(skillID uint32) bool {
	data, ok := s.initiveSkillData[skillID]
	if !ok {
		return false
	}

	return time.Now().Unix() < s.preUseTime[skillID]+int64(data.Cold)
}

// setPreUseTime 设置上次释放技能时间
func (s *SkillMgr) setPreUseTime(skillID uint32) {
	s.preUseTime[skillID] = time.Now().Unix()

	if data, ok := s.initiveSkillData[skillID]; ok {
		s.user.RPC(iserver.ServerTypeClient, "SkillPreUseTime", uint32(data.Cold))
	}
}

// judgeSkillOpenMode 判断此模式技能是否开放
func (s *SkillMgr) judgeSkillOpenMode() bool {
	skillSystemData, ok := excel.GetSkillSystem(common.SkillSystem_OpenModeID)
	if !ok {
		return false
	}

	openModeID := strings.Split(skillSystemData.Value2, ";")
	for _, v := range openModeID {
		if s.modeID == common.StringToUint32(v) {
			return true
		}
	}

	return false
}

// doInitiveEffect 主动技能
func (s *SkillMgr) doInitiveEffect(skillID uint32) {
	data, ok := s.initiveSkillData[skillID]
	if !ok {
		return
	}

	if data.Active != common.Skill_Initive {
		return
	}

	effects := strings.Split(data.SkillEffect, ";")
	for _, v := range effects {
		effect := strings.Split(v, "-")
		if len(effect) != 4 {
			s.user.Error("skill excel InitiveEffect err id:", data.Id, " InitiveEffect:", v)
			continue
		}

		skillEffect := &SkillEffectValue{}
		skillEffect.effectID = common.StringToUint32(effect[0])
		skillEffect.value = common.StringToUint64(effect[1])
		skillEffect.persistTime = common.StringToUint32(effect[2])
		skillEffect.interval = common.StringToUint32(effect[3])

		s.addInitiveEffect(skillEffect)
	}

	s.refreshSkillEffect()
}

// addInitiveEffect 添加主动技能的效果
func (s *SkillMgr) addInitiveEffect(effectValue *SkillEffectValue) {
	skilleffect := GetSkillEffect(uint64(effectValue.effectID))
	if skilleffect == nil {
		s.user.Error("不能获取技能效果", effectValue.effectID)
		return
	}

	skilleffect.Start(s.user, effectValue)

	//保存永久技能效果
	if effectValue.persistTime != 0 {
		tmp := &SkillEffectValue{}
		tmp.effectID = effectValue.effectID
		tmp.value = effectValue.value

		tmp.persistTime = effectValue.persistTime
		tmp.interval = effectValue.interval

		tmp.putTime = time.Now().Unix()
		tmp.nextPutTime = time.Now().Unix() + int64(effectValue.interval)

		tmp.entityID = effectValue.entityID
		s.initiveEffect[effectValue.effectID] = tmp
	}

	s.user.Debug("addInitiveEffect effectValue:", effectValue)
}

// doPassiveEffect 被动技能
func (s *SkillMgr) doPassiveEffect(skillID uint32) {
	data, ok := s.passiveSkillData[skillID]
	if !ok {
		return
	}

	if data.Active != common.Skill_Passive {
		return
	}

	effects := strings.Split(data.SkillEffect, ";")
	for _, j := range effects {
		effect := strings.Split(j, "-")
		if len(effect) != 4 {
			s.user.Debug("skill excel PassiveEffect err id:", data.Id, " PassiveEffect:", j)
			continue
		}

		skillEffect := &SkillEffectValue{}
		skillEffect.effectID = common.StringToUint32(effect[0])
		skillEffect.value = common.StringToUint64(effect[1])
		skillEffect.persistTime = common.StringToUint32(effect[2])
		skillEffect.interval = common.StringToUint32(effect[3])

		s.addPassiveEffect(skillEffect)
	}

	s.refreshSkillEffect()
}

// addPassiveEffect 添加被动技能的效果
func (s *SkillMgr) addPassiveEffect(effectValue *SkillEffectValue) {
	skilleffect := GetSkillEffect(uint64(effectValue.effectID))
	if skilleffect == nil {
		s.user.Error("不能获取技能效果", effectValue.effectID)
		return
	}

	// 初始化直接效果的技能
	if effectValue.effectID != SE_CallKillEffect {
		skilleffect.Start(s.user, effectValue)
	}

	//保存被动技能效果
	tmp := &SkillEffectValue{}
	tmp.effectID = effectValue.effectID
	tmp.value = effectValue.value

	tmp.persistTime = effectValue.persistTime
	tmp.interval = effectValue.interval

	tmp.putTime = time.Now().Unix()
	tmp.nextPutTime = time.Now().Unix() + int64(effectValue.interval)

	tmp.entityID = effectValue.entityID
	s.passiveEffect[effectValue.effectID] = tmp

	s.user.Debug("addPassiveEffect effectValue:", effectValue)
}

// update
func (s *SkillMgr) update() {
	if !s.judgeSkillOpenMode() {
		return
	}

	update := false

	// 主动技能
	for k, v := range s.initiveEffect {
		if v == nil {
			continue
		}

		skilleffect := GetSkillEffect(uint64(v.effectID))
		if skilleffect == nil {
			s.user.Debug("不能获取技能效果", v.effectID)
			continue
		}

		if skilleffect.IsEnd(v) { //处理持续时间过期的效果
			skilleffect.End(s.user, v)
			delete(s.initiveEffect, k)
			update = true
		} else if skilleffect.IsStart(v) { //处理持续时间内每间隔一定时间执行一次的效果
			skilleffect.Start(s.user, v)
		}
	}

	// 被动技能
	for _, v := range s.passiveEffect {
		if v == nil {
			continue
		}

		skilleffect := GetSkillEffect(uint64(v.effectID))
		if skilleffect == nil {
			s.user.Debug("不能获取技能效果", v.effectID)
			continue
		}

		if skilleffect.IsEnd(v) {
			skilleffect.End(s.user, v)
			update = true
		} else if skilleffect.IsStart(v) { //处理持续时间内每间隔一定时间执行一次的效果
			skilleffect.Start(s.user, v)
		}
	}

	if update {
		s.refreshSkillEffect()
	}
}

// refreshSkillEffect 刷新技能效果属性
func (s *SkillMgr) refreshSkillEffect() {
	msg := &protoMsg.SkillEffectList{}
	for _, v := range s.initiveEffect {
		if v == nil {
			continue
		}

		tmp := &protoMsg.SkillEffect{}
		tmp.Id = v.effectID
		tmp.Value = s.SkillData.skillEffectDam[tmp.Id]

		msg.List = append(msg.List, tmp)
	}

	for _, v := range s.passiveEffect {
		if v == nil {
			continue
		}

		tmp := &protoMsg.SkillEffect{}
		tmp.Id = v.effectID
		tmp.Value = s.SkillData.skillEffectDam[tmp.Id]

		msg.List = append(msg.List, tmp)
	}

	s.user.Debug("更新客户端技能效果:", msg)
	s.user.SetSkillEffectList(msg)
	s.user.SetSkillEffectListDirty()
}

// putInitiveSkill 释放主动技能
func (s *SkillMgr) putInitiveSkill(skillID uint32) {
	if !s.judgeSkillOpenMode() {
		return
	}

	if s.checkCold(skillID) {
		s.user.Debug("技能冷却中!!!")
		return
	}

	if skillID == 5 && s.user.getRidingCar() != nil {
		s.user.Debug("驾驶载具不能释放技能5：防爆护盾!")
		return
	}

	// s.user.CastRPCToAllClient("PutSkill", skillID)
	s.setPreUseTime(skillID)

	s.skillTargetRange(common.Skill_Initive, skillID)
}

// putPassiveSkill 释放被动技能(触发试的被动技能)
func (s *SkillMgr) putPassiveSkill() {
	if !s.judgeSkillOpenMode() {
		return
	}

	for k, _ := range s.passiveSkillData {
		s.skillTargetRange(common.Skill_Passive, k)
	}
}

// syncCurRoleSkillInfo 同步开启的主动技能
func (s *SkillMgr) syncCurRoleSkillInfo() {
	if !s.judgeSkillOpenMode() {
		s.user.Debug("syncCurRoleSkillInfo !s.judgeSkillOpenMode()")
		return
	}

	var initSkillID uint32
	for k, _ := range s.initiveSkillData {
		initSkillID = k
	}

	var passSkillID uint32
	for k, _ := range s.passiveSkillData {
		passSkillID = k
	}

	s.user.Debug("SyncCurRoleSkillInfo initSkillID:", initSkillID, " passSkillID:", passSkillID)
	s.user.RPC(iserver.ServerTypeClient, "SyncCurRoleSkillInfo", initSkillID, passSkillID)
}

// skillTargetRange 技能目标范围判断
func (s *SkillMgr) skillTargetRange(skillType, skillID uint32) {

	if skillType == common.Skill_Passive {
		data, ok := s.passiveSkillData[skillID]
		if !ok {
			return
		}

		switch data.SkillTarget {
		case SkillTarget_Own:
			s.doPassiveEffect(skillID)
		case SkillTarget_Friend:
			s.skillTargetFriend(skillType, skillID)
		case SkillTarget_OwnAndFriend:
			s.doPassiveEffect(skillID)
			s.skillTargetFriend(skillType, skillID)
		default:
			s.user.Error("skill target err! target:", data.SkillTarget)
		}
	} else if skillType == common.Skill_Initive {
		data, ok := s.initiveSkillData[skillID]
		if !ok {
			return
		}

		switch data.SkillTarget {
		case SkillTarget_Own:
			s.doInitiveEffect(skillID)
		case SkillTarget_Friend:
			s.skillTargetFriend(skillType, skillID)
		case SkillTarget_OwnAndFriend:
			s.doInitiveEffect(skillID)
			s.skillTargetFriend(skillType, skillID)
		default:
			s.user.Error("skill target err! target:", data.SkillTarget)
		}
	}
}

// skillTargetFriend 技能目标针对友方
func (s *SkillMgr) skillTargetFriend(skillType, skillID uint32) {
	space := s.user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	if !space.teamMgr.isTeam {
		return
	}

	team, ok := space.teamMgr.teams[s.user.teamid]
	if !ok {
		return
	}

	skillSystemData1, ok := excel.GetSkillSystem(common.SkillSystem_WarSongArea)
	if !ok {
		return
	}
	skillSystemData2, ok := excel.GetSkillSystem(common.SkillSystem_WarSongPeople)
	if !ok {
		return
	}

	num := 1
	for _, memid := range team {
		if memid == s.user.GetDBID() {
			continue
		}

		player, ok := space.GetEntityByDBID("Player", memid).(*RoomUser)
		if !ok || player == nil {
			continue
		}

		if num >= int(skillSystemData2.Value1) {
			continue
		}

		distance := common.Distance(s.user.GetPos(), player.GetPos())
		if distance > float32(skillSystemData1.Value1) {
			continue
		}

		if skillType == common.Skill_Passive {
			player.SkillMgr.doPassiveEffect(skillID)
			num++
		} else if skillType == common.Skill_Initive {
			player.SkillMgr.doInitiveEffect(skillID)
			num++
		}
	}
}

// syncCurRoleSkillTime 同步技能释放时间
func (s *SkillMgr) syncCurRoleSkillTime() {
	if !s.judgeSkillOpenMode() {
		s.user.Debug("syncCurRoleSkillTime !s.judgeSkillOpenMode()")
		return
	}

	var initSkillID uint32
	for k, _ := range s.initiveSkillData {
		initSkillID = k
	}

	var tm, coldTm uint32
	if data, ok := s.initiveSkillData[initSkillID]; ok {
		if s.preUseTime[initSkillID] != 0 && s.preUseTime[initSkillID]+int64(data.Cold) > time.Now().Unix() {
			tm = uint32(s.preUseTime[initSkillID] + int64(data.Cold) - time.Now().Unix())
		}

		coldTm = uint32(data.Cold)
	}

	s.user.Debug("syncCurRoleSkillTime tm:", tm, " coldTm:", coldTm)
	s.user.RPC(iserver.ServerTypeClient, "SyncCurRoleSkillInfoTimeRsp", tm, coldTm)
}

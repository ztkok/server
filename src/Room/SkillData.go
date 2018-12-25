package main

type SkillData struct {
	user *RoomUser

	skillEffectDam map[uint32]uint32 //伤害变化百分比  key:效果ID
}

func initSkillData(user *RoomUser) *SkillData {
	skilldata := new(SkillData)
	skilldata.user = user

	skilldata.skillEffectDam = make(map[uint32]uint32)

	return skilldata
}

// getSkillEffectDam 获取伤害变化百分比
func (s *SkillData) getSkillEffectDam(effectID uint32) float32 {
	return float32(s.skillEffectDam[effectID]) / 100.0
}

// getReduceTypeDam 效果降低的百分比
func (s *SkillData) getReduceTypeDam(id uint32) float32 {
	per := float32(100-s.skillEffectDam[id]) / 100.0
	if per < 0 {
		per = 0
	}

	return per
}

// getAddTypeDam 效果增加的百分比
func (s *SkillData) getAddTypeDam(id uint32) float32 {
	per := float32(100+s.skillEffectDam[id]) / 100.0

	return per
}

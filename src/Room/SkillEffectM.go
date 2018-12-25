package main

var gskilleffects map[uint64]ISkillEffect

//InitSkillEffect 初始化全局技能效果
func InitSkillEffect() {
	gskilleffects = make(map[uint64]ISkillEffect)

	gskilleffects[SE_Recoil] = &SkillEffect_Recoil{}
	gskilleffects[SE_ShrinkDam] = &SkillEffect_ShrinkDam{}
	gskilleffects[SE_TurnHp] = &SkillEffect_TurnHp{}
	gskilleffects[SE_WeaponDam] = &SkillEffect_WeaponDam{}
	gskilleffects[SE_FireSpeed] = &SkillEffect_FireSpeed{}
	gskilleffects[SE_Shield] = &SkillEffect_Shield{}
	gskilleffects[SE_MapTag] = &SkillEffect_MapTag{}
	gskilleffects[SE_RandomItem] = &SkillEffect_RandomItem{}
	gskilleffects[SE_AddHpCap] = &SkillEffect_AddHpCap{}
	gskilleffects[SE_MoveUseItem] = &SkillEffect_MoveUseItem{}
	gskilleffects[SE_RescueTime] = &SkillEffect_RescueTime{}
	gskilleffects[SE_HpRecover] = &SkillEffect_HpRecover{}
	gskilleffects[SE_DamReduce] = &SkillEffect_DamReduce{}
	gskilleffects[SE_VehicleNoDam] = &SkillEffect_VehicleNoDam{}
	gskilleffects[SE_OilLoss] = &SkillEffect_OilLoss{}
	gskilleffects[SE_ReleaseFog] = &SkillEffect_ReleaseFog{}
	gskilleffects[SE_CallKillEffect] = &SkillEffect_CallEffect{}
	gskilleffects[SE_VehicleDamReduce] = &SkillEffect_VehicleDamReduce{}
	gskilleffects[SE_RandomItem2] = &SkillEffect_RandomItem2{}
	gskilleffects[SE_RandomItem3] = &SkillEffect_RandomItem3{}
}

//GetSkillEffect 获取全局技能效果
func GetSkillEffect(id uint64) ISkillEffect {
	if effect, ok := gskilleffects[id]; ok {
		return effect
	}

	return nil
}

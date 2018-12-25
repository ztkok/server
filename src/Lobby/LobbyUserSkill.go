package main

import (
	"common"
	"db"
	"excel"
	"protoMsg"
	"strings"
	"zeus/iserver"
)

// setRoleSkillInfo 设置角色技能
func (user *LobbyUser) setRoleSkillInfo(roleID, modeID, skillType, skillID, position uint32) bool {
	if !user.judgeEquipSkill(roleID, skillID) {
		user.Error("技能模式未开放!")
		return false
	}

	info := &db.RoleSkillInfo{}
	info.InitiveSkill = make(map[uint32]map[uint32]uint32)
	info.PassiveSkill = make(map[uint32]map[uint32]uint32)

	skillUtil := db.PlayerSkillUtil(user.GetDBID(), uint64(roleID))
	if err := skillUtil.GetRoleSkill(info); err != nil {
		user.Error("GetRoleSkill fail! err:", err)
		return false
	}

	if skillType == common.Skill_Initive { //主动技能
		tmp := make(map[uint32]uint32)
		for k, v := range info.InitiveSkill[modeID] {
			if k != position {
				tmp[k] = v
			} else {
				tmp[position] = skillID
			}
		}
		if _, ok := tmp[position]; !ok {
			tmp[position] = skillID
		}
		info.InitiveSkill[modeID] = tmp
	} else if skillType == common.Skill_Passive { //被动技能
		tmp := make(map[uint32]uint32)
		for k, v := range info.InitiveSkill[modeID] {
			if k != position {
				tmp[k] = v
			} else {
				tmp[position] = skillID
			}
		}
		if _, ok := tmp[position]; !ok {
			tmp[position] = skillID
		}
		info.PassiveSkill[modeID] = tmp
	}

	user.Debug("SetRoleSkill info:", info)
	if err := skillUtil.SetRoleSkill(info); err != nil {
		user.Error("SetRoleSkill fail! err:", err)
		return false
	}

	return true
}

// getOneRoleSkillInfo 获取角色技能信息
func (user *LobbyUser) getOneRoleSkillInfo(roleID, modeID uint32) *protoMsg.RoleSkillInfo {
	msg := &protoMsg.RoleSkillInfo{}
	msg.RoleID = roleID

	info := &db.RoleSkillInfo{}
	info.InitiveSkill = make(map[uint32]map[uint32]uint32)
	info.PassiveSkill = make(map[uint32]map[uint32]uint32)

	skillUtil := db.PlayerSkillUtil(user.GetDBID(), uint64(roleID))
	if err := skillUtil.GetRoleSkill(info); err != nil {
		user.Error("GetRoleSkill fail! err:", err)
		return msg
	}

	modeSkillInfo := &protoMsg.ModeSkillInfo{}
	modeSkillInfo.ModeID = modeID
	for i, j := range info.InitiveSkill[modeID] {
		if j == 0 {
			continue
		}
		skillPosition := &protoMsg.SkillPosition{}
		skillPosition.Position = i
		skillPosition.SkillID = j

		modeSkillInfo.InitiveSkill = append(modeSkillInfo.InitiveSkill, skillPosition)
	}
	for i, j := range info.PassiveSkill[modeID] {
		if j == 0 {
			continue
		}
		skillPosition := &protoMsg.SkillPosition{}
		skillPosition.Position = i
		skillPosition.SkillID = j

		modeSkillInfo.PassiveSkill = append(modeSkillInfo.PassiveSkill, skillPosition)
	}

	msg.ModeList = append(msg.ModeList, modeSkillInfo)

	return msg
}

// getAllRoleSkillInfo 获取全部角色技能信息
func (user *LobbyUser) getAllRoleSkillInfo() *protoMsg.AllSkillList {

	msg := &protoMsg.AllSkillList{}

	for k, _ := range excel.GetRoleMap() {
		info := &db.RoleSkillInfo{}
		info.InitiveSkill = make(map[uint32]map[uint32]uint32)
		info.PassiveSkill = make(map[uint32]map[uint32]uint32)

		skillUtil := db.PlayerSkillUtil(user.GetDBID(), k)
		if err := skillUtil.GetRoleSkill(info); err != nil {
			user.Error("GetRoleSkill fail! err:", err)
			continue
		}

		if len(info.InitiveSkill) == 0 && len(info.PassiveSkill) == 0 {
			continue
		}

		roleSkillInfo := &protoMsg.RoleSkillInfo{}
		roleSkillInfo.RoleID = uint32(k)

		// 遍历开放的模式
		for _, v := range user.getSkillOpenMode() {
			msg := user.getOneRoleSkillInfo(uint32(k), v)
			if msg == nil {
				user.Error("getOneRoleSkillInfo fail! msg == nil")
				continue
			}

			for _, j := range msg.ModeList {
				roleSkillInfo.ModeList = append(roleSkillInfo.ModeList, j)
			}
		}

		msg.List = append(msg.List, roleSkillInfo)
	}
	return msg
}

// judgeEquipSkill 判断技能是否可以装备成功
func (user *LobbyUser) judgeEquipSkill(roleID, skillID uint32) bool {
	base, ok := excel.GetSkill(uint64(skillID))
	if !ok {
		user.Error("GetSkill fail! id:", skillID)
		return false
	}

	// 是否已购买该技能
	if base.SkillShop == 1 {
		if !db.PlayerGoodsUtil(user.GetDBID()).IsOwnGoods(skillID + 6000) {
			user.Error("No Own This Skill! SkillID:", skillID)
			return false
		}
	}

	if base.SkillRole == "0" {
		return true
	}

	var tmp bool
	skillRoles := strings.Split(base.SkillRole, ";")
	for _, v := range skillRoles {
		if roleID == common.StringToUint32(v) {
			tmp = true
			break
		}
	}

	return tmp
}

// getSkillOpenMode 获取技能开放模式
func (user *LobbyUser) getSkillOpenMode() []uint32 {
	var ret []uint32

	skillSystemData, ok := excel.GetSkillSystem(common.SkillSystem_OpenModeID)
	if !ok {
		return ret
	}

	openModeID := strings.Split(skillSystemData.Value2, ";")
	for _, v := range openModeID {
		ret = append(ret, common.StringToUint32(v))
	}

	return ret
}

// skillExpire 技能过期
func (user *LobbyUser) skillExpire(skillID uint32) {
	storeData, ok := excel.GetStore(uint64(skillID))
	if !ok {
		return
	}

	skillData, ok := excel.GetSkill(storeData.RelationID)
	if !ok {
		return
	}

	for _, roleData := range excel.GetRoleMap() {
		info := &db.RoleSkillInfo{}
		info.InitiveSkill = make(map[uint32]map[uint32]uint32)
		info.PassiveSkill = make(map[uint32]map[uint32]uint32)

		skillUtil := db.PlayerSkillUtil(user.GetDBID(), roleData.Id)
		if err := skillUtil.GetRoleSkill(info); err != nil {
			user.Error("GetRoleSkill fail! err:", err)
			continue
		}

		initSkills := strings.Split(roleData.Skill, ";")
		if len(initSkills) != 2 {
			user.Error("len(initSkills) != 2")
			continue
		}

		if skillData.Active == common.Skill_Initive {
			for k, v := range info.InitiveSkill {
				for i, j := range v {
					if j == uint32(skillData.Id) {
						user.setRoleSkillInfo(uint32(roleData.Id), k, common.Skill_Initive, common.StringToUint32(initSkills[0]), i)
					}
				}
			}
		} else if skillData.Active == common.Skill_Passive {
			for k, v := range info.PassiveSkill {

				for i, j := range v {
					if j == uint32(skillData.Id) {
						user.setRoleSkillInfo(uint32(roleData.Id), k, common.Skill_Passive, common.StringToUint32(initSkills[1]), i)
					}
				}
			}
		}
	}

	msg := user.getAllRoleSkillInfo()

	user.Debug("skillExpire SyncRoleSkillRsq skillID:", skillID, " msg:", msg)
	if err := user.RPC(iserver.ServerTypeClient, "SyncRoleSkillRsq", msg); err != nil {
		user.Error(err)
	}
}

// initEquipSkill 给每个未装备技能的角色装备初始技能
func (user *LobbyUser) initEquipSkill() {
	modes := user.getSkillOpenMode()
	if len(modes) == 0 {
		return
	}

	for _, v := range excel.GetRoleMap() {
		initSkills := strings.Split(v.Skill, ";")
		if len(initSkills) != 2 {
			user.Error("len(initSkills) != 2")
			continue
		}

		info := &db.RoleSkillInfo{}
		info.InitiveSkill = make(map[uint32]map[uint32]uint32)
		info.PassiveSkill = make(map[uint32]map[uint32]uint32)

		skillUtil := db.PlayerSkillUtil(user.GetDBID(), v.Id)
		if err := skillUtil.GetRoleSkill(info); err != nil {
			user.Error("GetRoleSkill fail! err:", err)
			continue
		}

		//主动技能
		if len(info.InitiveSkill) == 0 || (len(info.InitiveSkill) != 0 && !user.judgeEquipSkill(uint32(v.Id), info.InitiveSkill[modes[0]][1])) {
			skillID := common.StringToUint32(initSkills[0])
			if !user.judgeEquipSkill(uint32(v.Id), skillID) {
				user.Error("技能未开放! roleId:", v.Id, " skillID:", skillID)
				continue
			}

			tmp := make(map[uint32]uint32)
			tmp[1] = skillID
			info.InitiveSkill[modes[0]] = tmp
		}

		//被动技能
		if len(info.PassiveSkill) == 0 || (len(info.PassiveSkill) != 0 && !user.judgeEquipSkill(uint32(v.Id), info.PassiveSkill[modes[0]][1])) {
			skillID := common.StringToUint32(initSkills[1])
			if !user.judgeEquipSkill(uint32(v.Id), skillID) {
				user.Error("技能未开放! roleId:", v.Id, " skillID:", skillID)
				continue
			}

			tmp := make(map[uint32]uint32)
			tmp[1] = skillID
			info.PassiveSkill[modes[0]] = tmp
		}

		// user.Debug("SetRoleSkill info:", info)
		if err := skillUtil.SetRoleSkill(info); err != nil {
			user.Error("SetRoleSkill fail! err:", err)
			continue
		}
	}
}

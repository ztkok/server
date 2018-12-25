package db

import (
	"encoding/json"
	"fmt"
)

const (
	playerSkillPrefix = "PlayerSkill"
)

//========================================
//PlayerSkillUtil 管理玩家技能信息的工具
//========================================

type playerSkillUtil struct {
	uid    uint64
	roleID uint64
}

//PlayerSkillUtil 生成用于管理玩家技能信息的工具
func PlayerSkillUtil(uid, roleID uint64) *playerSkillUtil {
	return &playerSkillUtil{
		uid:    uid,
		roleID: roleID,
	}
}

func (u *playerSkillUtil) key() string {
	return fmt.Sprintf("%s:%d:%d", playerSkillPrefix, u.uid, u.roleID)
}

/*=========================角色技能信息设置==================*/

// 角色技能信息
type RoleSkillInfo struct {
	InitiveSkill map[uint32]map[uint32]uint32 //主动技能 模式ID做map的key
	PassiveSkill map[uint32]map[uint32]uint32 //被动技能
}

// SetRoleSkill 设置角色技能
func (u *playerSkillUtil) SetRoleSkill(info *RoleSkillInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	hSet(u.key(), "RoleSkill", string(data))
	return nil
}

// GetRoleSkill 获得角色技能
func (u *playerSkillUtil) GetRoleSkill(info *RoleSkillInfo) error {
	data := hGet(u.key(), "RoleSkill")
	if len(data) == 0 {
		return nil
	}

	err := json.Unmarshal([]byte(data), info)
	if err != nil {
		return err
	}

	return nil
}

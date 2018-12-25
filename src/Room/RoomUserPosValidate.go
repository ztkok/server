package main

import (
	"common"
)

//检查是否低于地面或水面
func (user *RoomUser) checkUnderGround(old, new *RoomPlayerState) bool {
	//oldpos := old.GetPos()
	newpos := new.GetPos()

	waterLevel := user.GetSpace().(*Scene).mapdata.Water_height
	isWater, err := user.GetSpace().IsWater(newpos.X, newpos.Z, waterLevel)
	if err != nil {
		return false
	}

	ground, err := user.GetSpace().GetHeight(newpos.X, newpos.Z)
	if err != nil {
		return false
	}

	height := ground
	if isWater {
		height = waterLevel
	}

	underheight := float32(common.GetTBSystemValue(2002))
	if newpos.Y < height-underheight {
		//seelog.Debug(user.GetID(), " 角色高度太低", oldpos, newpos, new.BaseState, height, isWater)
		return true
	}

	return false
}

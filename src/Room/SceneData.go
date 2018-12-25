package main

import (
	"common"
	"zeus/iserver"
)

type ISceneData interface {
	canAttack(user IRoomChracter, defendid uint64) bool
	doLoop()
	refreshSpecialBox(refreshcount uint64)
	resendData(user *RoomUser)
	onDeath(user IRoomChracter)
	onPickItem(uid uint64, item *Item) bool
	canDrop(uint64, *Item) bool
	updateTotalNum(user *RoomUser)
	updateAliveNum(user *RoomUser)
	clearSuperBox(id uint64)
	doResult()
}

type SceneData struct {
	scene *Scene
}

func NewSceneData(sc *Scene, matchmode uint32) ISceneData {
	if matchmode == common.MatchModeNormal {
		scenedata := NewNormalScene(sc)
		return scenedata
	} else if matchmode == common.MatchModeScuffle {
		scenedata := &SceneData{}
		scenedata.scene = sc
		return scenedata
	} else if matchmode == common.MatchModeEndless {
		scenedata := NewNormalScene(sc)
		scenedata.scene = sc
		return scenedata
	} else if matchmode == common.MatchModeElite {
		scenedata := NewEliteScene(sc)
		return scenedata
	} else if matchmode == common.MatchModeVersus {
		scenedata := NewVersusScene(sc)
		return scenedata
	} else if matchmode == common.MatchModeArcade {
		scenedata := NewArcadeScene(sc)
		return scenedata
	} else if matchmode == common.MatchModeTankWar {
		scenedata := NewTankWarScene(sc)
		return scenedata
	}

	scenedata := &SceneData{}
	scenedata.scene = sc
	return scenedata
}

func (scenedata *SceneData) canAttack(user IRoomChracter, defendid uint64) bool {
	return true
}

func (scenedata *SceneData) doLoop() {
}

func (scenedata *SceneData) refreshSpecialBox(refreshcount uint64) {
}

func (scenedata *SceneData) resendData(user *RoomUser) {
}

func (scenedata *SceneData) onDeath(user IRoomChracter) {

}

func (scenedata *SceneData) onPickItem(uid uint64, item *Item) bool {
	return true
}

func (scenedata *SceneData) canDrop(uid uint64, item *Item) bool {
	return true
}

func (scenedata *SceneData) updateTotalNum(user *RoomUser) {
	user.RPC(iserver.ServerTypeClient, "UpdateTotalNum", scenedata.scene.maxUserSum)
}

func (scenedata *SceneData) updateAliveNum(user *RoomUser) {
	user.RPC(iserver.ServerTypeClient, "UpdateAliveNum", uint32(len(scenedata.scene.members)))
}

func (scenedata *SceneData) clearSuperBox(id uint64) {

}

func (scenedata *SceneData) doResult() {
	scenedata.scene.DoResult()
}

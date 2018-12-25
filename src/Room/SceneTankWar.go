package main

// 坦克大战scene
type TankWarScene struct {
	SceneData
}

func NewTankWarScene(sc *Scene) *TankWarScene {
	scenedata := &TankWarScene{}
	scenedata.scene = sc
	return scenedata
}

func (scenedata *TankWarScene) canAttack(user IRoomChracter, defendid uint64) bool {
	if scenedata.scene.teamMgr.IsInOneTeamByID(user.GetID(), defendid) {
		return false
	}

	return true
}

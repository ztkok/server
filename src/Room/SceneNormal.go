package main

type NormalScene struct {
	SceneData
}

func NewNormalScene(sc *Scene) *NormalScene {
	scenedata := &NormalScene{}
	scenedata.scene = sc
	return scenedata
}

func (scenedata *NormalScene) canAttack(user IRoomChracter, defendid uint64) bool {
	if scenedata.scene.teamMgr.IsInOneTeamByID(user.GetID(), defendid) {
		return false
	}

	return true
}

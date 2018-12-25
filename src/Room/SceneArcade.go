package main

// 娱乐模式scene
type ArcadeScene struct {
	SceneData
}

func NewArcadeScene(sc *Scene) *ArcadeScene {
	scenedata := &ArcadeScene{}
	scenedata.scene = sc
	return scenedata
}

func (scenedata *ArcadeScene) canAttack(user IRoomChracter, defendid uint64) bool {
	if scenedata.scene.teamMgr.IsInOneTeamByID(user.GetID(), defendid) {
		return false
	}

	return true
}

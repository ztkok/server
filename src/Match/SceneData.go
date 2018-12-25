package main

import (
	"common"
	"protoMsg"

	log "github.com/cihub/seelog"
)

type ISceneData interface {
	onRoomSceneInited()
}

type SceneData struct {
	scene *Scene
}

func NewSceneData(sc *Scene, matchmode uint32) ISceneData {
	if matchmode == common.MatchModeNormal {
		scenedata := &NormalScene{}
		scenedata.scene = sc
		return scenedata
	} else if matchmode == common.MatchModeScuffle {
		scenedata := &SceneData{}
		scenedata.scene = sc
		return scenedata
	} else if matchmode == common.MatchModeEndless {
		scenedata := &SceneData{}
		scenedata.scene = sc
		return scenedata
	}

	scenedata := &SceneData{}
	scenedata.scene = sc
	return scenedata
}

func (scenedata *SceneData) onRoomSceneInited() {
	scene := scenedata.scene
	for _, obj := range scene.objs {
		obj.onRoomSceneInited(scene.GetID())
	}

	//if scene.isSingleMatch {
	//total := 0
	//for _, obj := range scene.objs {
	//	total += obj.GetNums()
	//}

	//roomusermin := int(common.GetTBSystemValue(common.System_RoomUserMin))
	//roomusermax := int(common.GetTBSystemValue(common.System_RoomUserLimit))
	//if roomusermax > roomusermin {
	//	roomusermin += rand.Intn(roomusermax-roomusermin) + 1
	//}
	//roomuserlimit := roomusermin
	//ainum := 0
	//if roomuserlimit > total {
	//	ainum = roomuserlimit - total
	//	scene.haveai = ainum
	//}
	//} else {
	// 添加机器人
	//teamlen := uint(len(scene.objs))
	//if teamlen < uint(getMinTeamSum(scene.teamType)) {
	//	addTeamSum := int(uint(getMinTeamSum(scene.teamType)) - teamlen)
	//
	//	rand.Seed(time.Now().Unix())
	//	if getMaxTeamSum(scene.teamType) > getMinTeamSum(scene.teamType) {
	//		max := int(getMaxTeamSum(scene.teamType))
	//		min := int(getMinTeamSum(scene.teamType))
	//		addTeamSum += rand.Intn(max - min)
	//	}
	//
	//	scene.haveai = int(addTeamSum * int(getTeamPlayerSum(scene.teamType)))
	//}
	//}

	proto := &protoMsg.SummonAINotify{}
	proto.Ainum = scene.haveai
	for _, obj := range scene.objs {
		proto.Users = append(proto.Users, obj.GetIDs()...)
	}
	proto.Skybox = scene.skybox
	if scene.isSingleMatch {
		proto.SceneType = 2
	} else {
		proto.SceneType = 1
	}

	if err := scene.Post(common.ServerTypeRoom, proto); err != nil {
		log.Error(err, scene)
	}
	//场景创建成功 队伍可以回退
	for _, objs := range scene.objs {
		var team *MatchTeam
		v, ok := GetTeamMgr().teams.Load(objs.GetID())
		if ok {
			team = v.(*MatchTeam)
			team.teamStatus = TeamNotMatch
			team.automatch = true
			team.DisposeTeam(false)
		}
	}
}

package main

import (
	"common"
	"fmt"
	"math"
	"time"
	"zeus/entity"
	"zeus/iserver"

	log "github.com/cihub/seelog"
)

// 轮询 table
func (srv *Server) doTravsal() {
	now := time.Now().Unix()
	srv.TravsalEntity("Space", func(e iserver.IEntity) {
		if e == nil {
			return
		}
		table := e.(*Scene)
		if table.status == common.SpaceStatusClose || (table.createTime != 0 && now >= table.createTime+60) {
			srv.DestroyEntity(table.GetID())
		}
	})
}

// Scene 场景
type Scene struct {
	entity.Entity
	ISceneData

	avgMMR uint32

	objs map[uint64]IMatcher

	mapid         uint32
	skybox        uint32
	haveai        uint32
	isSingleMatch bool
	teamType      uint8
	matchMode     uint32 //匹配模式
	uniqueId      uint32 //matchMode表中对应匹配模式的唯一id
	status        int
	createTime    int64
}

// Init 初始化
func (scene *Scene) Init(initParam interface{}) {
	scene.RegMsgProc(&SceneMsgProc{scene: scene})

	scene.status = common.SpaceStatusInit
	scene.createTime = time.Now().Unix()

	var (
		mapid     int
		uniqueId  uint32
		mgrName   string
		matchmode uint32
	)

	if _, err := fmt.Sscanf(initParam.(string), "%d:%d:%d:%s", &mapid, &uniqueId, &matchmode, &mgrName); err != nil {
		log.Error("初始化参数错误 ", initParam, scene)
		return
	}

	scene.ISceneData = NewSceneData(scene, matchmode)
	match := entity.GetEntityProxy(mgrName)
	if match == nil {
		log.Error("获取匹配管理器失败 ", initParam, scene)
		return
	}

	e := GetSrvInst().GetEntity(match.EntityID)
	if e == nil {
		log.Error("获取匹配管理器为空 ", match)
		return
	}

	mgr := e.(*MatchMgr)
	mgr.pushScene(scene)
}

// Destroy 析构时调用
func (scene *Scene) Destroy() {
	scene.objs = nil
}

// 获取最小队伍数
func getMinTeamSum(teamType uint8) uint32 {
	minSum := common.GetTBSystemValue(8)
	if minSum == 0 {
		return 1
	}

	var ret float64
	if teamType == TwoTeamType {
		ret = float64(minSum) / 2
	} else if teamType == FourTeamType {
		ret = float64(minSum) / 4
	}

	return uint32(math.Ceil(ret))
}

// 获取最大队伍数
func getMaxTeamSum(teamType uint8) uint32 {
	maxSum := common.GetTBSystemValue(6)
	if maxSum == 0 {
		return 1
	}

	var ret float64
	if teamType == TwoTeamType {
		ret = float64(maxSum) / 2
	} else if teamType == FourTeamType {
		ret = float64(maxSum) / 4
	}

	return uint32(math.Floor(ret))
}

// getTeamPlayerSum 获取每个队伍玩家数
func getTeamPlayerSum(teamType uint8) uint32 {

	if teamType == TwoTeamType {
		return 2
	} else if teamType == FourTeamType {
		return 4
	}

	return 0
}

func (scene *Scene) onRoomSceneInited() {
	scene.ISceneData.onRoomSceneInited()
}

package main

import (
	"common"
	"io/ioutil"
	"protoMsg"
	"zeus/iserver"
	"zeus/linmath"
)

const (
	// 关门
	closeDoor = 0
	// 内开门
	innerOpenDoor = 1
	// 外开门
	outerOpenDoor = 2
)

// DoorInfo 门信息
type Door struct {
	id    uint64
	state uint32
	pos   linmath.Vector3
}

// SpaceDoorMgr 场景门管理器
type SpaceDoorMgr struct {
	space   *Scene
	allDoor map[uint64]Door
}

// NewSpaceDoorMgr 初始化场景门管理器
func NewSpaceDoorMgr(scene *Scene) *SpaceDoorMgr {
	doorMgr := &SpaceDoorMgr{
		space:   scene,
		allDoor: make(map[uint64]Door),
	}
	doorMgr.init()

	return doorMgr
}

// 初始化门的信息
func (doorMgr *SpaceDoorMgr) init() {

	path := "../res/space/" + common.Uint64ToString(doorMgr.space.mapdata.Map_pxs_ID) + "/doorConfig"

	data, err := ioutil.ReadFile(path)
	if err != nil {
		doorMgr.space.Error("doorMgr init failed, ReadFile err: ", err, " path: ", path)
		return
	}

	var msg protoMsg.DoorList = protoMsg.DoorList{}
	msg.Unmarshal(data)

	door := Door{}
	for _, item := range msg.DoorList {
		door.id = item.Id
		door.pos.X = item.Pos.X
		door.pos.Y = item.Pos.Y
		door.pos.Z = item.Pos.Z
		door.state = item.State

		doorMgr.allDoor[door.id] = door

	}

	doorMgr.space.Info("doorMgr init success, len: ", len(doorMgr.allDoor))

}

// 请求设置门的状态
func (doorMgr *SpaceDoorMgr) SetDoorState(user *RoomUser, doorID uint64, newState uint32) {
	if user == nil {
		return
	}

	door, ok := doorMgr.allDoor[doorID]
	if ok == false {
		doorMgr.space.Error("SetDoorState failed, can't get door, doorID: ", doorID)
		return
	}

	// 判断设置状态是否合法
	if newState != outerOpenDoor && newState != closeDoor && newState != innerOpenDoor {
		doorMgr.space.Error("SetDoorState failed, state is not ok, state: ", newState)
		return
	}

	if newState == door.state || (newState != closeDoor && door.state != closeDoor) {
		doorMgr.space.Error("SetDoorState failed, state is not ok, state: ", door.state, newState)
		return
	}

	// 验证玩家距离是否合法
	distance := user.GetPos().Sub(door.pos).Len()
	systemValue := float32(common.GetTBSystemValue(50)) / 100
	if distance > systemValue {
		doorMgr.space.Info("SetDoorState failed, distance is so far, distance: ", distance)
		//return
	}

	doorMgr.space.Info("Set Door State success, oldstate: ", door.state, " newstate:", newState)
	door.state = newState
	doorMgr.allDoor[doorID] = door

	// 通知场景所有玩家门的状态
	doorMgr.space.rpcSpaceDoorState(doorID, newState)

}

// SendAllDoorStateToUser 发送所有门的状态给玩家
func (doorMgr *SpaceDoorMgr) SendAllDoorStateToUser(user *RoomUser) {

	for id, door := range doorMgr.allDoor {
		if door.state != closeDoor {
			user.RPC(iserver.ServerTypeClient, "SpaceDoorState", id, door.state)
		}
	}
}

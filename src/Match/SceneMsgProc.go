package main

import "common"

// SceneMsgProc 消息处理函数
type SceneMsgProc struct {
	scene *Scene
}

// RPC_RoomSpaceInited 场景服务器中的Space实体创建成功
func (proc *SceneMsgProc) RPC_RoomSpaceInited() {
	proc.scene.onRoomSceneInited()
	proc.scene.status = common.SpaceStatusClose
}

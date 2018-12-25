package main

import (
	"db"
	"protoMsg"
	"zeus/msgdef"
)

// SpaceMsgProc RoomUser的消息处理函数
type SpaceMsgProc struct {
	space *Scene
}

//RPC_InitSpaceTeam 初始化场景组队信息 c-->s
func (proc *SpaceMsgProc) RPC_InitSpaceTeam() {
	//log.Debug("初始化场景组队信息")
	//proc.space.teamMgr.init(proc.space)
}

func (proc *SpaceMsgProc) RPC_TeamMemberTotalEnter() {
	//log.Debug("组队场景中的玩家全部进入")
	//proc.space.teamMgr.InitRoomTeamInfo()
}

func (proc *SpaceMsgProc) MsgProc_SummonAINotify(content msgdef.IMsg) {
	msg := content.(*protoMsg.SummonAINotify)
	proc.space.aiNum = msg.Ainum
	proc.space.onAirAiNum = msg.Ainum
	proc.space.aiNumSurplus = msg.Ainum

	proc.space.skybox = msg.Skybox
	if msg.SceneType == 2 {
		proc.space.teamMgr.isTeam = false
	} else {
		proc.space.teamMgr.isTeam = true
	}

	proc.space.maxUserSum = proc.space.aiNum + uint32(len(msg.Users))
	proc.space.SummonAI(uint32(msg.Ainum))

	// 保存场景信息, mapid, skybox
	db.SceneTempUtil(proc.space.GetID()).SetInfo(uint32(proc.space.mapdata.Id), proc.space.skybox, proc.space.GetMatchMode(), 0)
}

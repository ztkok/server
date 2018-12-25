package main

import (
	"common"
	"db"
	"excel"
	"protoMsg"
	"zeus/iserver"
)

// RPC_GetAchievementInfo 获取某个玩家的成就信息
func (proc *LobbyUserMsgProc) RPC_GetAchievementInfo(uid uint64) {
	util := db.PlayerAchievementUtil(uid)
	msg := &protoMsg.AchievmentInfo{}
	msg.Level, msg.Exp = util.GetLevelInfo()
	for _, data := range util.GetAchieveInfo() {
		l := &protoMsg.Achievement{}
		l.Id = uint32(data.Id)
		l.Stamp = data.Stamp
		l.Flag = data.Flag
		msg.List = append(msg.List, l)
	}
	condition := util.GetAllAchievementData()
	for k, v := range condition {
		if k == common.AchievementRunDistance {
			v = v / 1000
		}
		msg.Process = append(msg.Process, &protoMsg.AchievementProcess{
			Id:  uint32(k),
			Num: float32(v),
		})
	}
	msg.Reward = util.GetReward()
	msg.Used = util.GetShow()
	proc.user.RPC(iserver.ServerTypeClient, "GetAchievementInfoRet", uid, msg)
}

// RPC_GetAchievementReward 领取成就奖励
func (proc *LobbyUserMsgProc) RPC_GetAchievementReward(level uint32) {
	levelData, ok := excel.GetAchievementLevel(uint64(level))
	if !ok {
		return
	}
	if len(levelData.Reward1) == 0 || len(levelData.Reward1) != len(levelData.Reward2) {
		return
	}

	util := db.PlayerAchievementUtil(proc.user.GetDBID())
	if util.IsGetReward(level) {
		return
	}
	nowLevel, _ := util.GetLevelInfo()
	if nowLevel == 0 || nowLevel < level {
		return
	}
	for k, v := range levelData.Reward1 {
		proc.user.storeMgr.GetGoods(v, levelData.Reward2[k], common.RS_Achievement, common.MT_NO, 0)
	}
	util.AddReward(level)
	proc.user.RPC(iserver.ServerTypeClient, "GetAchievementRewardRet", level)
}

// RPC_AchievementUse 设置成就的展示
func (proc *LobbyUserMsgProc) RPC_AchievementUse(pos uint32, id uint32) {
	if pos == 0 || pos > 3 {
		return
	}
	util := db.PlayerAchievementUtil(proc.user.GetDBID())
	if !util.IsGetAchieve(id) {
		return
	}
	used := util.GetShow()
	for _, v := range used {
		if v == id {
			return
		}
	}
	used[pos-1] = id
	util.SetShow(used)
	proc.user.RPC(iserver.ServerTypeClient, "AchievementUseRet", used[0], used[1], used[2])
}

// RPC_InsigniaInfo 获取某个玩家的勋章信息
func (proc *LobbyUserMsgProc) RPC_InsigniaInfo(uid uint64) {
	proc.user.InsigniaInfo(uid)
}

// RPC_InsigniaUse 使用勋章
func (proc *LobbyUserMsgProc) RPC_InsigniaUse(insigniaId uint64) {
	if !db.PlayerInsigniaUtil(proc.user.GetDBID()).IsExistsInsignia(insigniaId) {
		return
	}
	db.PlayerInfoUtil(proc.user.GetDBID()).SetInsigniaUsed(insigniaId)
	proc.user.RPC(iserver.ServerTypeClient, "InsigniaUseRet", insigniaId)

	if proc.user.teamMgrProxy != nil && proc.user.GetTeamID() != 0 {
		if err := proc.user.teamMgrProxy.RPC(common.ServerTypeMatch, "ChangeTeamMemberInfo",
			proc.user.GetID(), proc.user.GetTeamID()); err != nil {
			proc.user.Error("RPC ChangeTeamMemberInfo err: ", err)
			return
		}
	}
	proc.user.InsigniaFlow(uint32(insigniaId))
}

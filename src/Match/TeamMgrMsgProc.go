package main

import (
	"common"
	"container/list"
	"db"
	"protoMsg"
	"time"
	"zeus/entity"
	"zeus/iserver"

	log "github.com/cihub/seelog"
)

// TeamMgrMsgProc TeamMgr消息处理函数
type TeamMgrMsgProc struct {
	mgr *TeamMgr
}

// RPC_EnterTeam 进入队伍
func (proc *TeamMgrMsgProc) RPC_EnterTeam(srvID, entityID, oldTeam, teamID uint64, mapID, mmr, tworank, fourrank uint32, total uint32, name string, role uint32, dbid uint64, teamtype, automatch, matchMode, isVeteran, color, weapon uint32) {
	var team *MatchTeam

	if oldTeam != 0 {
		v, ok := proc.mgr.teams.Load(oldTeam)
		if ok {
			team = v.(*MatchTeam)
			if team.IsExisted(entityID) && team.GetMatchMode() == matchMode {
				//刷新队伍中该玩家的名字
				m, ok := team.members.Load(entityID)
				if ok {
					mem := m.(*MatchMember)
					mem.name = name
				}
				team.BroadInfo()
				return
			}
		}
	}

	v, ok := proc.mgr.teams.Load(teamID)
	if !ok {
		team = NewMatchTeam(teamID, mapID, total, teamtype, automatch, matchMode)
		proc.mgr.teams.Store(teamID, team)
	} else {
		team = v.(*MatchTeam)
	}

	if team.IsExisted(entityID) {
		log.Warnf("已经在队伍里了 TeamID:%d EntityID:%d", teamID, entityID)
		return
	}

	member := NewMatchMember(srvID, entityID, mmr, tworank, fourrank, mapID, name, role, dbid, matchMode, isVeteran, color, weapon)
	team.Add(member)
	team.BroadInfo()
	if err := member.RPC(common.ServerTypeLobby, "EnterTeamRet", uint32(0), teamID); err != nil {
		log.Error(err)
	}
}

func (proc *TeamMgrMsgProc) RPC_ChangeTeamType(srvID, entityID, teamID uint64, teamtype uint32) {
	var team *MatchTeam
	v, ok := proc.mgr.teams.Load(teamID)
	if !ok {
		log.Warnf("队伍都不存在 TeamID:%d EntityID:%d", teamID, entityID)
		return
	}
	team = v.(*MatchTeam)

	if team.leaderid != entityID {
		return
	}

	if teamtype == uint32(TwoTeamType) && team.GetNums() > 2 {
		log.Warn("队伍人数错误, 双人队伍成员数大于2")
		return
	}

	team.teamType = uint8(teamtype)
	team.calcAvgMMR()
	team.BroadInfo()
}

func (proc *TeamMgrMsgProc) RPC_ChangeAutoMatch(srvID, entityID, teamID uint64, automatch uint32) {
	var team *MatchTeam
	v, ok := proc.mgr.teams.Load(teamID)
	if !ok {
		log.Warnf("队伍都不存在 TeamID:%d EntityID:%d", teamID, entityID)
		return
	}
	team = v.(*MatchTeam)

	if team.leaderid != entityID {
		log.Warnf("非队长不允许换队伍类型, 队长 %d, 请求 %d", team.leaderid, entityID)
		return
	}

	team.automatch = (automatch != 0)
}

func (proc *TeamMgrMsgProc) RPC_ChangeMap(srvID, entityID, teamID uint64, mapid uint32) {
	var team *MatchTeam
	v, ok := proc.mgr.teams.Load(teamID)
	if !ok {
		log.Warnf("队伍都不存在 TeamID:%d EntityID:%d", teamID, entityID)
		return
	}
	team = v.(*MatchTeam)

	if team.leaderid != entityID {
		log.Warnf("非队长不允许换地图, 队长 %d, 请求 %d", team.leaderid, entityID)
		return
	}

	team.mapid = mapid
	team.BroadInfo()
}

func (proc *TeamMgrMsgProc) RPC_checkQuitTeam(srvID, dbid, teamID uint64) {
	var team *MatchTeam
	v, ok := proc.mgr.teams.Load(teamID)
	if !ok {
		log.Warnf("队伍都不存在 TeamID:%d dbid:%d", teamID, dbid)
		return
	}

	team = v.(*MatchTeam)
	inmatch := (team.teamStatus == TeamWatingScene)
	entityID := team.getEntityID(dbid)

	if entityID != 0 {
		//log.Debug("上线退出队伍", entityID)
		team.Remove(entityID)
		if team.IsEmpty() {
			proc.mgr.RemoveTeam(teamID)
			db.PlayerTeamUtil(teamID).Remove()
			db.DelTeamVoiceInfo(teamID)
		} else {
			team.BroadInfo()
			team.RemoveTeamMemberVoiceInfo(entityID)
		}

		if inmatch {
			proc.mgr.delMatch(team)
		}
	}
}

// RPC_LeaveTeam 离开队伍
func (proc *TeamMgrMsgProc) RPC_LeaveTeam(srvID, entityID, teamID uint64, notify bool) {
	proc.LeaveTeam(srvID, entityID, teamID, notify)
}

func (proc *TeamMgrMsgProc) RPC_InviteLeaveTeam(srvID, entityID, teamID uint64) {
	proc.LeaveTeam(srvID, entityID, teamID, false)
}

// RPC_LeaveTeam 离开队伍
func (proc *TeamMgrMsgProc) LeaveTeam(srvID, entityID, teamID uint64, isNotifyLobby bool) {
	var team *MatchTeam
	v, ok := proc.mgr.teams.Load(teamID)
	if !ok {
		log.Warnf("队伍都不存在 TeamID:%d EntityID:%d", teamID, entityID)
		return
	}

	team = v.(*MatchTeam)
	inmatch := (team.teamStatus == TeamWatingScene)

	team.Remove(entityID)
	if team.IsEmpty() {
		proc.mgr.RemoveTeam(teamID)
		db.PlayerTeamUtil(teamID).Remove()
		db.DelTeamVoiceInfo(teamID)
	} else {
		team.BroadInfo()
		team.RemoveTeamMemberVoiceInfo(entityID)
	}

	if inmatch {
		proc.mgr.delMatch(team)
	}

	if isNotifyLobby {
		member := entity.NewEntityProxy(srvID, 0, entityID)
		if err := member.RPC(common.ServerTypeLobby, "QuitTeamRet", uint32(0)); err != nil {
			log.Error(err)
		}
	}
}

// RPC_TeamInfo 更新队伍信息
func (proc *TeamMgrMsgProc) RPC_TeamInfo(teamID uint64, mapid uint32, typ uint8, nums uint8) {
	if typ != 0 && typ != 1 {
		log.Warn("队伍类型错误 Type:", typ)
		return
	}
	if nums > 4 {
		log.Warn("人数错误 ", nums)
		return
	}
	if mapid != 1 && mapid != 2 {
		log.Warn("地图ID错误 ", mapid)
		return
	}

	v, ok := proc.mgr.teams.Load(teamID)
	if !ok {
		log.Warn("队伍不存在 TeamID:", teamID)
		return
	}
	team := v.(*MatchTeam)
	team.typ = typ
	team.nums = int(nums)
}

func (proc *TeamMgrMsgProc) RPC_ConfirmTeamMatch(srvID, entityID, teamID uint64) {
	var team *MatchTeam
	v, ok := proc.mgr.teams.Load(teamID)
	if !ok {
		log.Warnf("队伍都不存在 TeamID:%d EntityID:%d", teamID, entityID)
		return
	}

	team = v.(*MatchTeam)

	if team.teamStatus == TeamMatching || team.teamStatus == TeamWatingScene {
		log.Warn("已在匹配中!")
		return
	}

	isMatch := true
	// for _, mem := range team.members {
	// 	if mem.GetID() == entityID {
	// 		mem.state = MemberReadying

	// 	} else {
	// 		if mem.state == MemberNotReady {
	// 			isMatch = false
	// 		}
	// 	}
	// }
	team.members.Range(
		func(k, v interface{}) bool {
			mem := v.(*MatchMember)
			if mem.GetID() == entityID {
				mem.state = MemberReadying

			} else {
				if mem.state == MemberNotReady {
					isMatch = false
				}
			}
			mem.matchTime = time.Now().Unix()
			return true
		})

	if isMatch {
		team.teamStatus = TeamMatching
		team.matchtingime = time.Now().Unix()
		proc.mgr.boradExpectTime(team)
		team.BroadInfo()

		if !team.automatch || team.IsFull() {
			team.teamStatus = TeamWatingScene
			proc.mgr.addMatch(team)
		} else {
			cpList, ok := proc.mgr.cpLists[team.GetMatchMode()]
			if !ok {
				cpList = list.New()
				proc.mgr.cpLists[team.GetMatchMode()] = cpList
			}
			cpList.PushBack(team)
		}
	} else {
		team.BroadInfo()
	}
}

func (proc *TeamMgrMsgProc) RPC_InviteRsp(srvID, entityID, teamID uint64, name string, role uint32, dbid uint64, curTeamID uint64, tworank, fourrank, isVeteran, color, weapon uint32) {

	proxy := entity.NewEntityProxy(srvID, 0, entityID)

	var team *MatchTeam
	v, ok := proc.mgr.teams.Load(teamID)
	if !ok {
		proxy.RPC(common.ServerTypeLobby, "InviteRspRet", uint32(1), teamID, curTeamID)
		return
	}

	team = v.(*MatchTeam)

	if team.teamStatus != TeamNotMatch {
		proxy.RPC(common.ServerTypeLobby, "InviteRspRet", uint32(1), teamID, curTeamID)
		return
	}
	if team.GetNums() >= 2 && !common.IsMatchModeOk(team.GetMatchMode(), 4) {
		log.Warn("Match mode is not open, matchMode: ", team.GetMatchMode(), " matchTyp: ", 4)
		proxy.RPC(common.ServerTypeLobby, "InviteRspRet", uint32(2), teamID, curTeamID)
		return
	}

	if team.GetNums() >= 4 {
		proxy.RPC(common.ServerTypeLobby, "InviteRspRet", uint32(2), teamID, curTeamID)
		return
	}

	if team.IsExisted(entityID) || team.getEntityID(dbid) != 0 {
		return
	}

	member := NewMatchMember(srvID, entityID, 1000, tworank, fourrank, team.mapid, name, role, dbid, team.GetMatchMode(), isVeteran, color, weapon)
	member.stateEnterTeam = 2 //队友邀请进入队伍状态
	team.members.Range(
		func(k, v interface{}) bool {
			mm := v.(*MatchMember)
			if mm.stateEnterTeam == 0 {
				mm.stateEnterTeam = 2 //好友开黑进入队列
			}
			return true
		})
	team.Add(member)
	team.BroadInfo()
	proxy.RPC(common.ServerTypeLobby, "InviteRspRet", uint32(0), teamID, curTeamID)
	proxy.RPC(common.ServerTypeLobby, "ConfirmTeamMatch")
	log.Debug("邀请加入队伍")
}

// 队员更换角色，通知全队玩家
func (proc *TeamMgrMsgProc) RPC_SetRoleModel(entityID, teamID uint64, role uint32) {
	var team *MatchTeam
	v, ok := proc.mgr.teams.Load(teamID)
	if !ok {
		log.Warnf("队伍都不存在 TeamID:%d EntityID:%d", teamID, entityID)
		return
	}

	team = v.(*MatchTeam)

	team.members.Range(
		func(k, v interface{}) bool {
			mem := v.(*MatchMember)
			if mem.GetID() == entityID {
				mem.role = role
				team.BroadInfo()
				return true
			}
			return true
		})
}

// 队员更换武器，通知全队玩家
func (proc *TeamMgrMsgProc) RPC_SetOutsideWeapon(entityID, teamID uint64, id uint32) {
	var team *MatchTeam
	v, ok := proc.mgr.teams.Load(teamID)
	if !ok {
		log.Warnf("队伍都不存在 TeamID:%d EntityID:%d", teamID, entityID)
		return
	}

	team = v.(*MatchTeam)

	team.members.Range(
		func(k, v interface{}) bool {
			mem := v.(*MatchMember)
			if mem.GetID() == entityID {
				mem.outsideWeapon = id
				team.BroadInfo()
				return true
			}
			return true
		})
}

func (proc *TeamMgrMsgProc) RPC_AutoEnterTeam(srvID, entityID, teamID uint64, name string, role uint32, dbid uint64, curTeamID uint64, tworank, fourrank, isVeteran, color, weapon uint32) {
	// log.Debug("auto enter team:", teamID)
	if curTeamID != 0 { // 如果在队伍中,从队伍中删除
		proc.LeaveTeam(srvID, entityID, curTeamID, false)
	}

	proxy := entity.NewEntityProxy(srvID, 0, entityID)

	var team *MatchTeam
	v, ok := proc.mgr.teams.Load(teamID)
	if !ok {
		proxy.RPC(common.ServerTypeLobby, "AutoEnterTeamRet", uint32(1))
		return
	}

	team = v.(*MatchTeam)

	if team.GetNums() >= 2 && !common.IsMatchModeOk(team.GetMatchMode(), 4) {
		log.Warn("Match mode is not open, matchMode: ", team.GetMatchMode(), " matchTyp: ", 4)
		proxy.RPC(common.ServerTypeLobby, "AutoEnterTeamRet", uint32(1))
		return
	}

	if team.teamStatus != TeamNotMatch {
		proxy.RPC(common.ServerTypeLobby, "AutoEnterTeamRet", uint32(1))
		return
	}

	if team.GetNums() >= 4 {
		proxy.RPC(common.ServerTypeLobby, "AutoEnterTeamRet", uint32(1))
		return
	}

	if team.IsExisted(entityID) || team.getEntityID(dbid) != 0 {
		proxy.RPC(common.ServerTypeLobby, "AutoEnterTeamRet", uint32(1))
		return
	}

	member := NewMatchMember(srvID, entityID, 1000, tworank, fourrank, team.mapid, name, role, dbid, team.GetMatchMode(), isVeteran, color, weapon)
	team.Add(member)
	team.BroadInfo()
	proxy.RPC(common.ServerTypeLobby, "AutoEnterTeamRet", uint32(0))
	// log.Debug("加入队伍", teamID)
}

func (proc *TeamMgrMsgProc) RPC_SyncVoiceInfo(svrId, teamid uint64, memberId int32, entityId uint64) {
	// log.Info("RPC_SyncVoiceInfo ", teamid, " ", entityId, " ", memberId)

	redismemlist := db.GetTeamVoiceInfo(teamid)
	replaced := false
	for _, v := range redismemlist {
		if v.EntityId == entityId {
			v.MemberId = memberId
			replaced = true
			break
		}
	}
	if !replaced {
		redismemlist = append(redismemlist, &db.MemVoiceInfo{
			EntityId:   entityId,
			MemberId:   memberId,
			LobbySrvId: svrId,
		})
	}

	db.SetTeamVoiceInfo(teamid, redismemlist)
	msg := &protoMsg.TeamVoiceInfo{}
	for _, v := range redismemlist {
		msg.MemberInfos = append(msg.MemberInfos, &protoMsg.MemVoiceInfo{
			MemberId: v.MemberId,
			Uid:      v.EntityId,
		})
	}
	for _, v := range redismemlist {
		proxy := entity.NewEntityProxy(v.LobbySrvId, 0, v.EntityId)
		// log.Info("SyncTeamVoiceInfo ", v, " ", msg.MemberInfos)
		proxy.RPC(iserver.ServerTypeClient, "SyncTeamVoiceInfo", msg)
	}
}

// 队员更换名字，通知全队玩家
func (proc *TeamMgrMsgProc) RPC_ChangeTeamMemberName(entityID, teamID uint64, name string) {
	var team *MatchTeam
	v, ok := proc.mgr.teams.Load(teamID)
	if !ok {
		log.Warnf("队伍都不存在 TeamID:%d EntityID:%d", teamID, entityID)
		return
	}

	team = v.(*MatchTeam)

	team.members.Range(
		func(k, v interface{}) bool {
			mem := v.(*MatchMember)
			if mem.GetID() == entityID {
				mem.name = name
				team.BroadInfo()
				return true
			}
			return true
		})
}

// 队员更换名字，通知全队玩家
func (proc *TeamMgrMsgProc) RPC_ChangeTeamMemberInfo(entityID, teamID uint64) {
	var team *MatchTeam
	v, ok := proc.mgr.teams.Load(teamID)
	if !ok {
		log.Warnf("队伍都不存在 TeamID:%d EntityID:%d", teamID, entityID)
		return
	}

	team = v.(*MatchTeam)
	team.BroadInfo()
}

func (proc *TeamMgrMsgProc) RPC_GetTeamInfo(entityID, teamID uint64) {
	v, ok := proc.mgr.teams.Load(teamID)
	if !ok {
		log.Warnf("队伍都不存在 TeamID:%d EntityID:%d", teamID, entityID)
		return
	}

	team := v.(*MatchTeam)
	team.NotifyToMember(entityID)
}

func (proc *TeamMgrMsgProc) RPC_CancelTeamReady(entityID, teamID uint64) {
	v, ok := proc.mgr.teams.Load(teamID)
	if !ok {
		log.Warnf("队伍都不存在 TeamID:%d EntityID:%d", teamID, entityID)
		return
	}

	team := v.(*MatchTeam)
	if entityID == team.leaderid {
		return
	}
	team.members.Range(func(key, value interface{}) bool {
		mem := value.(*MatchMember)
		if mem.GetID() != entityID {
			return true
		}
		mem.state = MemberNotReady
		team.BroadInfo()
		return true
	})
}
func (proc *TeamMgrMsgProc) RPC_SyncGameRecord(entityID, teamID uint64, tworank, fourrank uint32, detail *protoMsg.GameRecordDetail) {

	v, ok := proc.mgr.teams.Load(teamID)
	if !ok {
		log.Warnf("队伍都不存在 TeamID:%d EntityID:%d", teamID, entityID)
		return
	}

	team := v.(*MatchTeam)
	team.members.Range(func(key, value interface{}) bool {
		mem := value.(*MatchMember)
		if mem.GetID() == entityID {
			mem.rank = tworank
			mem.fourrank = fourrank
			mem.gameRecord = detail
		}
		return true
	})
}

func (proc *TeamMgrMsgProc) RPC_CancelQueue(teamID uint64) {
	v, ok := proc.mgr.teams.Load(teamID)
	if !ok {
		log.Warnf("队伍都不存在 TeamID:%d EntityID:%d", teamID)
		return
	}

	team := v.(*MatchTeam)
	if team.teamStatus == TeamMatching {
		team.teamStatus = TeamNotMatch
		team.members.Range(func(key, value interface{}) bool {
			mem := value.(*MatchMember)
			if mem.EntityID == team.leaderid {
				mem.state = MemberNotReady
			}
			return true
		})
		team.BroadInfo()
	}
	if team.teamStatus == TeamWatingScene {
		proc.mgr.delMatch(team)
	}
}

func (proc *TeamMgrMsgProc) RPC_GetExpectTime(entityID, teamID uint64) {
	v, ok := proc.mgr.teams.Load(teamID)
	if !ok {
		log.Warnf("队伍都不存在 TeamID:%d EntityID:%d", teamID, entityID)
		return
	}

	team := v.(*MatchTeam)
	if proc.mgr.expectTime == 0 {
		proc.mgr.expectTime = 1
	}

	team.members.Range(
		func(k, v interface{}) bool {
			pUser := v.(*MatchMember)
			if pUser.GetID() == entityID {
				pUser.RPC(iserver.ServerTypeClient, "ExpectTime", uint64(proc.mgr.expectTime))
			}
			return true
		})
}

// RPC_KickTeamMember 踢出队员
func (proc *TeamMgrMsgProc) RPC_KickTeamMember(entityID, teamID, memID uint64) {
	v, ok := proc.mgr.teams.Load(teamID)
	if !ok {
		log.Error("KickTeamMember failed, team not exist, teamID: ", teamID)
		return
	}

	team := v.(*MatchTeam)
	if team == nil {
		return
	}

	if !team.IsTeamLeader(entityID) {
		log.Warn("Only team leader can kick other member, entityID: ", entityID)
		return
	}

	team.RPC(common.ServerTypeLobby, "KickTeamMember", entityID, memID)

	team.Remove(memID)
	team.RemoveTeamMemberVoiceInfo(memID)
	team.BroadInfo()
}

// RPC_TransferTeamLeader 转让队长
func (proc *TeamMgrMsgProc) RPC_TransferTeamLeader(entityID, teamID, memID uint64) {
	v, ok := proc.mgr.teams.Load(teamID)
	if !ok {
		log.Error("TransferTeamLeader failed, team not exist, teamID: ", teamID)
		return
	}

	team := v.(*MatchTeam)
	if team == nil {
		return
	}

	if !team.IsTeamLeader(entityID) {
		log.Warn("Only team leader can transfer leader, entityID: ", entityID)
		return
	}

	team.SetTeamLeader(memID)

	team.RPC(common.ServerTypeLobby, "TransferTeamLeader", memID)
	team.BroadInfo()
}

// RPC_QuickDialog 快捷对话
func (proc *TeamMgrMsgProc) RPC_QuickDialog(entityID, teamID uint64, id uint32) {
	v, ok := proc.mgr.teams.Load(teamID)
	if !ok {
		log.Error("QuickDialog failed, team not exist, teamID: ", teamID)
		return
	}

	team := v.(*MatchTeam)
	if team == nil {
		return
	}

	team.RPC(common.ServerTypeLobby, "QuickDialog", entityID, id)
}

// RPC_ShowRecord 战绩展示
func (proc *TeamMgrMsgProc) RPC_ShowRecord(entityID, teamID uint64) {
	v, ok := proc.mgr.teams.Load(teamID)
	if !ok {
		log.Error("ShowRecord failed, team not exist, teamID: ", teamID)
		return
	}

	team := v.(*MatchTeam)
	if team == nil {
		return
	}

	mm, ok := team.members.Load(entityID)
	if ok {
		member := mm.(*MatchMember)
		team.RPC(common.ServerTypeLobby, "ShowRecord", member.gameRecord)
	}
}

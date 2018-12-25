package main

import (
	"common"
	"zeus/entity"

	log "github.com/cihub/seelog"
)

// MatchMgrMsgProc MatchMgr消息处理函数
type MatchMgrMsgProc struct {
	mgr *MatchMgr
}

// RPC_EnterSoloQueue 进入匹配队列
func (proc *MatchMgrMsgProc) RPC_EnterSoloQueue(srvID, entityID uint64, mapid, mmr, rank uint32, name string, role uint32, dbid uint64, matchMode, isVeteran, color, weapon uint32) {
	member := NewMatchMember(srvID, entityID, mmr, rank, rank, mapid, name, role, dbid, matchMode, isVeteran, color, weapon)
	l, ok := proc.mgr.wsList[mapid]
	if ok {
		for e := l.Front(); e != nil; e = e.Next() {
			ws := e.Value.(IWaitingScene)
			if m := ws.Get(entityID); m != nil && m.GetMatchMode() == matchMode {
				if err := member.RPC(common.ServerTypeLobby, "EnterSoloQueueRet", uint32(1), uint64(0)); err != nil {
					log.Error(err)
				}
				return
			}
		}
	}

	ws := proc.mgr.getMatchScene(member)
	ws.Add(member)
	if err := member.RPC(common.ServerTypeLobby, "EnterSoloQueueRet", uint32(0), proc.mgr.calExpectTime()); err != nil {
		log.Error(err)
	}
}

func (proc *MatchMgrMsgProc) RPC_CancelSoloQueue(srvID, entityID uint64) {
	proxy := entity.NewEntityProxy(srvID, 0, entityID)
	for _, l := range proc.mgr.wsList {
		for e := l.Front(); e != nil; e = e.Next() {
			ws := e.Value.(IWaitingScene)
			if m := ws.Get(entityID); m != nil {
				ws.Remove(m)
				if err := proxy.RPC(common.ServerTypeLobby, "CancelSoloQueueRet", uint32(0)); err != nil {
					log.Error(err, srvID, entityID)
				}
				return
			}
		}
	}

	if err := proxy.RPC(common.ServerTypeLobby, "CancelSoloQueueRet", uint32(1)); err != nil {
		log.Error(err, srvID, entityID)
	}
}

func (proc *MatchMgrMsgProc) RPC_EnterDuoQueue(teamid uint64) {
	var team *MatchTeam
	v, ok := GetTeamMgr().teams.Load(teamid)
	if !ok {
		log.Warn("找不到队伍 ", teamid)
		return
	}

	team = v.(*MatchTeam)
	if team.IsReady() {
		ws := proc.mgr.getMatchScene(team)
		ws.Add(team)
	}
}

func (proc *MatchMgrMsgProc) RPC_CancelQueue(teamID uint64) {
	for _, l := range proc.mgr.wsList {
		for e := l.Front(); e != nil; e = e.Next() {
			ws := e.Value.(IWaitingScene)
			if m := ws.Get(teamID); m != nil {
				if team, ok := m.(*MatchTeam); ok {
					team.teamStatus = TeamNotMatch
					team.DisposeTeam(true)
				}
				ws.Remove(m)
				return
			}
		}
	}
}

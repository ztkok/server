package main

import (
	"common"
	"zeus/linmath"

	//"math/rand"
	"db"
	"math"
	"protoMsg"
	"time"
	//"strings"

	"zeus/dbservice"
	"zeus/iserver"

	"github.com/cihub/seelog"
	"github.com/garyburd/redigo/redis"
)

const (
	//TeamMemberNotEnter 成员未进入
	TeamMemberNotEnter = 1
	//TeamMemberFighting 成员战斗中
	TeamMemberFighting = 2
	//TeamMemberDeath 成员死亡
	TeamMemberDeath = 3
	//TeamMemberLeave 成员离开
	TeamMemberLeave = 4
)

/*
type TeamMem struct {
	userid uint64
	state  uint32
}

type TeamItem struct {
	mems   map[uint64]*TeamMem
}
*/

// SpaceMemberInfo 场景成员信息
type SpaceMemberInfo struct {
	memberid    uint64
	username    string
	url         string
	killnum     uint32
	headshotnum uint32
	damagehp    uint32
	distance    float32
	gametime    int64
	rating      float32
	veteran     uint32
	namecolor   uint32
}

// 玩家离开场景后，仍保留它的基本信息（刷新队友信息时会使用）
type MemberBaseInfo struct {
	ID      uint64
	Name    string
	SignPos linmath.Vector2
	DiePos  linmath.Vector3
}

//SpaceTeam 场景组队管理
type SpaceTeam struct {
	space           *Scene
	isTeam          bool
	teamType        uint32
	teams           map[uint64][]uint64
	memberInfo      map[uint64]*SpaceMemberInfo
	membersBaseInfo map[uint64]*MemberBaseInfo

	followInvites   map[uint64][]uint64 //跳伞跟随邀请
	followRelations map[uint64][]uint64 //跳伞跟随关系
}

//NewSpaceTeam 创建场景队伍
func NewSpaceTeam(sp *Scene) *SpaceTeam {
	spTeam := &SpaceTeam{
		space:           sp,
		isTeam:          false,
		teams:           make(map[uint64][]uint64),
		memberInfo:      make(map[uint64]*SpaceMemberInfo),
		membersBaseInfo: make(map[uint64]*MemberBaseInfo),
		followInvites:   make(map[uint64][]uint64),
		followRelations: make(map[uint64][]uint64),
	}

	return spTeam
}

// 初始化组队数据
func (st *SpaceTeam) initTeamData(space *Scene) {

	if space == nil {
		return
	}
	// Redis数据库中场景的组队数据读入到内存
	spaceTmInfo := db.SpaceTeamUtil(space.GetID()).GetSpaceTeamInfo()
	if spaceTmInfo == nil {
		st.isTeam = false
		return
	}

	st.space = space
	st.isTeam = true
	st.teamType = spaceTmInfo.Teamtype

	for _, spaceteam := range spaceTmInfo.Teams {
		tm := make([]uint64, 0)
		for _, memid := range spaceteam.MemList {
			tm = append(tm, memid)
		}
		st.teams[spaceteam.Id] = tm
	}

	st.PrintSpaceTeamInfo()
	st.delSpaceTeam()
}

// 删除组队数据
func (st *SpaceTeam) delSpaceTeam() {
	if st.space == nil {
		return
	}

	db.SpaceTeamUtil(st.space.GetID()).DelSpaceTeamInfo()
}

//PrintSpaceTeamInfo 打印场景中队伍信息
func (st *SpaceTeam) PrintSpaceTeamInfo() {
	if st.space == nil {
		return
	}

	st.space.Debug("Space team info, space: ", st.space.GetID(), " type: ", st.isTeam)
	for teamid, team := range st.teams {
		st.space.Debug("teamid: ", teamid)

		for _, memid := range team {
			st.space.Debug("memid: ", memid)
		}
	}
}

//  判断副本中是否还存在同一队伍中的玩家用于判断救援
func (st *SpaceTeam) isExistTeammate(user *RoomUser) bool {
	// log.Info("判断玩家是否为可救援状态", user.GetDBID())
	if st.isTeam == false {
		//st.space.Error("isExistTeammate failed, not a team")
		return false
	}

	teamID := user.GetUserTeamID()
	if teamID == 0 {
		return false
	}

	team, ok := st.teams[teamID]
	if ok == false {
		st.space.Error("isExistTeammate failed, team not exist, teamid: ", teamID)
		return false
	}

	for _, memid := range team {

		tmpUser, ok := st.space.GetEntityByDBID("Player", memid).(*RoomUser)
		if !ok || tmpUser == user || tmpUser.userType == RoomUserTypeWatcher || tmpUser.stateMgr == nil {
			continue
		}
		//log.Info("判断玩家是否为可救援状态队友信息", memid, tmpUser.GetDBID(), tmpUser.stateMgr.GetState())
		state := tmpUser.stateMgr.GetState()
		if state != RoomPlayerBaseState_Dead && state != RoomPlayerBaseState_WillDie /*&& state != BeRescue*/ && state != RoomPlayerBaseState_Watch {
			return true
		}
	}

	return false
}

// isExistTeam 判断队伍中是否还有活着的成员
func (st *SpaceTeam) isExistTeam(user *RoomUser) bool {
	if !st.isTeam {
		return false
	}

	teamID := user.GetUserTeamID()
	if teamID == 0 {
		return false
	}

	team, ok := st.teams[teamID]
	if !ok {
		st.space.Error("isExistTeammate failed, team not exist, teamid: ", teamID)
		return false
	}

	for _, memid := range team {
		tmpUser, ok := st.space.GetEntityByDBID("Player", memid).(*RoomUser)
		if !ok || tmpUser.userType == RoomUserTypeWatcher || tmpUser.stateMgr == nil {
			continue
		}

		state := tmpUser.stateMgr.GetState()
		if state != RoomPlayerBaseState_Dead && state != RoomPlayerBaseState_WillDie && state != RoomPlayerBaseState_Watch {
			return true
		}
	}

	return false
}

// getWatchTargetInTeam 获取队内观战目标
func (st *SpaceTeam) getWatchTargetInTeam(user *RoomUser) uint64 {
	if user == nil || !st.isTeam {
		return 0
	}

	teamid := user.GetUserTeamID()
	if teamid == 0 {
		return 0
	}

	team, ok := st.teams[teamid]
	if !ok {
		st.space.Warn("getWatchTargetInTeam failed, team not exist, teamid: ", teamid)
		return 0
	}

	order := user.curWatchOrder
	memCnt := uint32(len(team))

	for idx := uint32(0); idx < memCnt; idx++ {
		order++
		order %= memCnt
		memid := team[order]

		tmpUser, ok := st.space.GetEntityByDBID("Player", memid).(*RoomUser)
		if !ok || tmpUser == user || tmpUser.userType == RoomUserTypeWatcher {
			continue
		}

		state := tmpUser.stateMgr.GetState()
		if state != RoomPlayerBaseState_Dead && state != RoomPlayerBaseState_WillDie && state != RoomPlayerBaseState_Watch {
			user.curWatchOrder = order
			return tmpUser.GetID()
		}
	}

	return 0
}

// getAliveTeammate 获取一个活着的队友
func (st *SpaceTeam) getAliveTeammate(user *RoomUser) uint64 {
	if user == nil || !st.isTeam {
		return 0
	}

	teamid := user.GetUserTeamID()
	if teamid == 0 {
		return 0
	}

	team, ok := st.teams[teamid]
	if !ok {
		st.space.Info("getAliveTeammate failed, team not exist, teamid: ", teamid)
		return 0
	}

	for _, memid := range team {
		tmpUser, ok := st.space.GetEntityByDBID("Player", memid).(*RoomUser)
		if !ok || tmpUser == user || tmpUser.userType == RoomUserTypeWatcher {
			continue
		}

		state := tmpUser.stateMgr.GetState()
		if state != RoomPlayerBaseState_Dead && state != RoomPlayerBaseState_WillDie && state != RoomPlayerBaseState_Watch {
			return tmpUser.GetID()
		}
	}

	return 0
}

//InitRoomTeamInfo 玩家进入副本初始化组队信息
func (st *SpaceTeam) InitRoomTeamInfo(tb *Scene) {
	//space := proc.user.GetSpace().(*Scene)

	st.initTeamData(tb)

	if st.isTeam == false {
		//st.space.Error("InitRoomTeamInfo failed, not a team")
		return
	}

	//log.Infof("玩家进入副本场景初始化组队信息 InitRoomTeamInfo  st.teams(%d)", len(st.teams))

	for teamid, team := range st.teams {

		retMsg := &protoMsg.InitRoomTeamInfoRet{}
		teammateList := make([]*RoomUser, 0)
		for index, memid := range team {

			user, ok := st.space.GetEntityByDBID("Player", memid).(*RoomUser)

			if !ok {
				//log.Errorf("组队场景中的未找到玩家 memid(%d)", memid)
				continue
			}
			teammateList = append(teammateList, user)

			item := protoMsg.InitRoomTeamPlayerItem{
				Id:    user.GetID(),
				Name_: user.GetName(),
			}
			item.Pos = &protoMsg.Vector3{}
			item.Hp = uint32(user.GetHP())
			item.State = uint32(user.GetState())
			item.Maxhp = uint32(user.GetMaxHP())

			pos := user.GetPos()
			item.Pos.X = pos.X
			item.Pos.Y = pos.Y
			item.Pos.Z = pos.Z

			item.Color = uint32(index)
			item.NameColor = common.GetPlayerNameColor(user.GetDBID())

			retMsg.Item = append(retMsg.Item, &item)

			st.membersBaseInfo[user.GetDBID()] = &MemberBaseInfo{
				ID:   user.GetID(),
				Name: user.GetName(),
			}
			//log.Infof("初始化组队数据  id(%d), hp(%d), maxhp(%d), userName(%s)", item.GetId(), item.GetHp(), item.GetMaxhp(), item.GetName_())
		}

		for _, memid := range team {

			user, ok := st.space.GetEntityByDBID("Player", memid).(*RoomUser)

			if !ok {
				continue
			}
			user.teamid = teamid
			user.RPC(iserver.ServerTypeClient, "SyncUserTeamid", user.GetUserTeamID())
			user.RPC(iserver.ServerTypeClient, "InitRoomTeamInfoRet", retMsg)

			for _, j := range teammateList {
				if user == j {
					continue
				}
				//log.Debug("AddExtWatchEntity ", user.GetName(), " ", j.GetName())
				user.AddExtWatchEntity(j)
			}

		}

	}

	return
}

func (st *SpaceTeam) GetInitRoomTeamInfo(teamID uint64) *protoMsg.InitRoomTeamInfoRet {
	team := st.teams[teamID]
	if team == nil {
		return nil
	}

	retMsg := &protoMsg.InitRoomTeamInfoRet{}
	for index, memid := range team {
		teamUser, ok := st.space.GetEntityByDBID("Player", memid).(*RoomUser)
		item := protoMsg.InitRoomTeamPlayerItem{}
		if !ok || teamUser.userType == RoomUserTypeWatcher {
			baseInfo := st.membersBaseInfo[memid]
			if baseInfo == nil {
				seelog.Errorf("组队场景中的未找到玩家 memid(%d)", memid)
				continue
			} else {
				item.Id = baseInfo.ID
				item.Name_ = baseInfo.Name
				item.State = RoomPlayerBaseState_LeaveMap
				item.Color = uint32(index)
			}
		} else {

			item.Id = teamUser.GetID()
			item.Name_ = teamUser.GetName()
			item.Pos = &protoMsg.Vector3{}
			item.Hp = uint32(teamUser.GetHP())
			item.State = uint32(teamUser.GetState())
			item.Maxhp = uint32(teamUser.GetMaxHP())

			pos := teamUser.GetPos()
			if teamUser.GetState() == RoomPlayerBaseState_Watch {
				if baseInfo := st.membersBaseInfo[memid]; baseInfo != nil {
					pos = baseInfo.DiePos
				}
			}

			item.Pos.X = pos.X
			item.Pos.Y = pos.Y
			item.Pos.Z = pos.Z

			item.Color = uint32(index)
			item.NameColor = common.GetPlayerNameColor(teamUser.GetDBID())
		}

		retMsg.Item = append(retMsg.Item, &item)
		seelog.Infof("重发组队数据  id(%d), hp(%d), maxhp(%d), userName(%s)", item.GetId(), item.GetHp(), item.GetMaxhp(), item.GetName_())
	}
	return retMsg
}

// 发送自己的队伍给玩家
func (st *SpaceTeam) SendMyTeamInfoToMe(user *RoomUser) {

	if st.isTeam == false {
		return
	}

	teamID := user.GetUserTeamID()
	if teamID == 0 {
		return
	}
	team := st.teams[teamID]
	if team == nil {
		user.Error("invalid teamId:", teamID)
		return
	}

	retMsg := st.GetInitRoomTeamInfo(teamID)

	user.RPC(iserver.ServerTypeClient, "SyncUserTeamid", teamID)
	user.RPC(iserver.ServerTypeClient, "InitRoomTeamInfoRet", retMsg)

	for _, memid := range team {
		baseInfo := st.membersBaseInfo[memid]
		if baseInfo != nil {
			signPos := baseInfo.SignPos
			if math.Abs(float64(signPos.X)) > 0.1 || math.Abs(float64(signPos.Y)) > 0.1 {
				user.RPC(iserver.ServerTypeClient, "SyncMapSign", baseInfo.ID, float64(baseInfo.SignPos.X), float64(baseInfo.SignPos.Y))
			}
		}
	}
}

//SyncAllTeamInfo 同步所有队伍信息
func (st *SpaceTeam) SyncAllTeamInfo() {
	st.SyncRoomTeamInfo(true, 0)
}

//SyncRoomTeamInfo 同步组队信息
func (st *SpaceTeam) SyncRoomTeamInfo(allTeam bool, targetTeamID uint64) {
	//log.Infof("同步组队信息 teamssum(%d) ", len(st.teams))
	if st.isTeam == false {
		//st.space.Error("SyncRoomTeamInfo failed, not a team")
		return
	}

	for teamid, team := range st.teams {

		if !allTeam && teamid != targetTeamID {
			continue
		}

		retMsg := &protoMsg.SymcRoomTeamInfoRet{}
		for _, memid := range team {
			user, ok := st.space.GetEntityByDBID("Player", memid).(*RoomUser)

			if !ok || user.userType == RoomUserTypeWatcher {
				continue
			}

			item := protoMsg.SyncRoomTeamPlayerItem{
				Id: user.GetID(),
			}

			item.Hp = uint32(user.GetHP())

			item.Pos = &protoMsg.Vector3{}
			pos := user.GetPos()
			item.Pos.X = pos.X
			item.Pos.Y = pos.Y
			item.Pos.Z = pos.Z
			item.State = uint32(user.GetState())

			item.Rota = &protoMsg.Vector3{}
			rota := user.GetRota()
			item.Rota.X = rota.X
			item.Rota.Y = rota.Y
			item.Rota.Z = rota.Z

			//log.Infof("同步组队数据  id(%d), hp(%d), posx(%d), posy(%d), posz(%d)", item.GetId(), item.GetHp(), item.Pos.GetX, item.Pos.GetY, item.Pos.GetZ)

			retMsg.Item = append(retMsg.Item, &item)
		}

		for _, memid := range team {

			user, ok := st.space.GetEntityByDBID("Player", memid).(*RoomUser)

			if !ok {
				continue
			}

			state := user.stateMgr.GetState()
			if state == RoomPlayerBaseState_Dead || state == RoomPlayerBaseState_Watch {
				//continue
			}

			user.RPC(iserver.ServerTypeClient, "SymcRoomTeamInfoRet", retMsg)
			/*
				if err := user.Post(iserver.ServerTypeClient, retMsg); err != nil {
					log.Error(err)
				}
			*/
		}

	}

	return
}

//GetTeammatedbid 获取队友id
func (st *SpaceTeam) GetTeammatedbid(playerdbid uint64) uint64 {

	user, ok := st.space.GetEntityByDBID("Player", playerdbid).(*RoomUser)

	if !ok {
		st.space.Error("GetTeammatedbid failed, player id not exist, id: ", playerdbid)
		return 0
	}

	team, ok := st.teams[user.GetUserTeamID()]
	if ok == false {
		st.space.Error("GetTeammatedbid failed, team id not exist, id: ", user.GetUserTeamID())
		return 0
	}

	for _, memid := range team {
		if memid != playerdbid {
			return memid
		}
	}

	return 0
}

//SyncMapSign 同步总览地图标记
func (st *SpaceTeam) SyncMapSign(user *RoomUser, x float64, y float64) {
	if user == nil {
		return
	}

	team, ok := st.teams[user.GetUserTeamID()]
	if ok == false {
		st.space.Error("SyncMapSign failed, team id not exist, id: ", user.GetUserTeamID())
		return
	}

	for _, memid := range team {
		targetUser, ok := st.space.GetEntityByDBID("Player", memid).(*RoomUser)
		if !ok || user == targetUser {
			continue
		}
		//log.Infof("同步总览地图标记  id(%d), posx(%g), posy(%g)", user.GetID(), x, y)
		targetUser.RPC(iserver.ServerTypeClient, "SyncMapSign", user.GetID(), x, y)
		targetUser.BroadcastToWatchers(RoomUserTypeWatcher, "SyncMapSign", user.GetID(), x, y)
	}
	user.BroadcastToWatchers(RoomUserTypeWatcher, "SyncMapSign", user.GetID(), x, y)
	// 保存标记位置
	baseInfo := st.membersBaseInfo[user.GetDBID()]
	if baseInfo != nil {
		baseInfo.SignPos.X = float32(x)
		baseInfo.SignPos.Y = float32(y)
	}
}

//CancelSyncMapSign 取消地图标记
func (st *SpaceTeam) CancelSyncMapSign(user *RoomUser) {
	if user == nil {
		return
	}

	team, ok := st.teams[user.GetUserTeamID()]
	if ok == false {
		st.space.Error("CancelSyncMapSign failed, team id not exist, id: ", user.GetUserTeamID())
		return
	}

	for _, memid := range team {
		targetUser, ok := st.space.GetEntityByDBID("Player", memid).(*RoomUser)
		if !ok || user == targetUser {
			continue
		}
		//log.Info("取消地图标记, teamid(%d), id(%d), targetid(%d)", user.GetUserTeamID(), user.GetID(), memid)
		targetUser.RPC(iserver.ServerTypeClient, "CancelSyncMapSign", user.GetID())
		targetUser.BroadcastToWatchers(RoomUserTypeWatcher, "CancelSyncMapSign", user.GetID())
	}
	user.BroadcastToWatchers(RoomUserTypeWatcher, "CancelSyncMapSign", user.GetID())
	// 取消标记位置
	baseInfo := st.membersBaseInfo[user.GetDBID()]
	if baseInfo != nil {
		baseInfo.SignPos.X = float32(0)
		baseInfo.SignPos.Y = float32(0)
	}
}

//SyncDownEndTime 同步队员被击倒后的剩余生存时间
func (st *SpaceTeam) SyncDownEndTime(user *RoomUser, endTimeStamp uint64, surplusTime uint64) {
	if user == nil || st.isTeam == false {
		return
	}

	team, ok := st.teams[user.GetUserTeamID()]
	if ok == false {
		st.space.Error("SyncDownEndTime failed, team id not exist, id: ", user.GetUserTeamID())
		return
	}

	var isPause uint8 = 0
	if user.stateMgr.isBerescue {
		isPause = 1
	}

	for _, memid := range team {
		targetUser, ok := st.space.GetEntityByDBID("Player", memid).(*RoomUser)

		if !ok {
			continue
		}

		//log.Infof("同步队员被击倒后的剩余生存时间 targetid(%d), id(%d), endTimeStamp(%d), surplusTime(%d), isPause(%d)", targetUser.GetDBID(), user.GetDBID(), endTimeStamp, surplusTime, isPause)
		targetUser.RPC(iserver.ServerTypeClient, "SyncDownEndTime", user.GetID(), uint32(endTimeStamp), uint32(surplusTime), isPause)
		targetUser.BroadcastToWatchers(RoomUserTypeWatcher, "SyncDownEndTime", user.GetID(), uint32(endTimeStamp), uint32(surplusTime), isPause)
	}
}

//SurplusTeamSum 场景剩余队伍数
func (st *SpaceTeam) SurplusTeamSum() uint32 {

	if st.isTeam == false {
		//st.space.Error("SurplusTeamSum failed, not a team")
		return 0
	}

	teaminfo := make(map[uint64]int)

	memberSum := 0
	for _, team := range st.teams {
		for _, memid := range team {
			tmpUser, ok := st.space.GetEntityByDBID("Player", memid).(*RoomUser)

			if !ok || tmpUser.userType == RoomUserTypeWatcher {
				continue
			}

			if tmpUser.stateMgr.GetState() == RoomPlayerBaseState_Dead || tmpUser.stateMgr.GetState() == RoomPlayerBaseState_Watch {
				continue
			}
			memberSum++

			teaminfo[tmpUser.GetUserTeamID()]++
		}
	}

	var aisum, aiTeamsum uint32 = 0, 0
	if len(st.space.members) > memberSum {
		aisum = uint32(len(st.space.members) - memberSum)
	}
	if aisum != 0 {
		playerSum := st.getTeamPlayerSum()
		if playerSum == 0 {
			return 0
		}
		aiTeamsum = uint32(aisum-1)/uint32(playerSum) + 1
	}
	// log.Infof("aisum (%d) memberSum(%d) st.space.members(%d)", aisum, memberSum, len(st.space.members), len(teaminfo))

	return uint32(len(teaminfo)) + uint32(aiTeamsum)
}

//TeamsSum 场景队伍数
func (st *SpaceTeam) TeamsSum() uint32 {

	if st.isTeam == false {
		//st.space.Error("TeamsSum failed, not a team")
		return 0
	}

	playerSum := st.getTeamPlayerSum()
	if playerSum == 0 {
		return 0
	}
	return uint32(uint(len(st.teams)) + uint(math.Ceil(float64(st.space.aiNum)/float64(playerSum))))
}

// ShowDieFlag 显示死亡标记
func (st *SpaceTeam) ShowDieFlag(user *RoomUser) {
	if st.isTeam == false {
		//st.space.Error("ShowDieFlag failed, not a team")
		return
	}

	team, ok := st.teams[user.GetUserTeamID()]
	if ok == false {
		st.space.Error("ShowDieFlag failed, team id not exist, id: ", user.GetUserTeamID())
		return
	}
	//记录玩家死亡的位置
	baseInfo := st.membersBaseInfo[user.GetDBID()]
	if baseInfo != nil {
		baseInfo.DiePos = user.GetPos()
	}
	for _, memid := range team {
		tmpUser, ok := st.space.GetEntityByDBID("Player", memid).(*RoomUser)

		if !ok || user.GetDBID() == memid || tmpUser.userType == RoomUserTypeWatcher {
			continue
		}
		tmpUser.RPC(iserver.ServerTypeClient, "ShowDieFlag", user.GetID(), float64(user.GetPos().X), float64(user.GetPos().Z))
		tmpUser.BroadcastToWatchers(RoomUserTypeWatcher, "ShowDieFlag", user.GetID(), float64(user.GetPos().X), float64(user.GetPos().Z))
		//log.Debug("显示死亡标记", user.GetID(), float64(user.GetPos().X), float64(user.GetPos().Z))
	}
}

// TeamDirectBroad 在队伍内广播指挥指令
func (st *SpaceTeam) TeamDirectBroad(user *RoomUser, id uint8, dir uint16) {
	if st.isTeam == false {
		//st.space.Error("TeamDirect failed, not a team")
		return
	}

	team, ok := st.teams[user.GetUserTeamID()]
	if ok == false {
		st.space.Error("TeamDirect failed, team id not exist, id: ", user.GetUserTeamID())
		return
	}

	for _, memid := range team {
		tmpUser, ok := st.space.GetEntityByDBID("Player", memid).(*RoomUser)
		if !ok || tmpUser.userType == RoomUserTypeWatcher {
			continue
		}

		tmpUser.RPC(iserver.ServerTypeClient, "TeamDirectBroad", user.GetID(), id, dir)
	}
}

// InviteTeamFollow 邀请队友跟随跳伞
func (st *SpaceTeam) InviteTeamFollow(user *RoomUser) bool {
	if user.GetBaseState() != RoomPlayerBaseState_Inplane {
		st.space.Warn("InviteTeamFollow failed, inviter is not inplane, BaseState: ", user.GetBaseState())
		return false
	}

	if !st.isTeam {
		return false
	}

	team, ok := st.teams[user.GetUserTeamID()]
	if !ok {
		st.space.Error("InviteTeamFollow failed, team id not exist, id: ", user.GetUserTeamID())
		return false
	}

	reqId := user.GetID()
	target := st.GetFollowTarget(user.GetUserTeamID(), reqId)
	followers := st.followRelations[target]

	invites := []uint64{}

	for _, memid := range team {
		tmpUser, ok := st.space.GetEntityByDBID("Player", memid).(*RoomUser)
		if !ok {
			continue
		}

		if tmpUser.GetBaseState() != RoomPlayerBaseState_Inplane {
			continue
		}

		tmpId := tmpUser.GetID()

		if tmpId == target {
			continue
		}

		var exist bool

		for _, follower := range followers {
			if tmpId == follower {
				exist = true
			}
		}

		if exist {
			continue
		}

		tmpUser.RPC(iserver.ServerTypeClient, "InviteFollowNotify", reqId)
		invites = append(invites, tmpId)
	}

	if len(invites) > 0 {
		st.followInvites[reqId] = invites
		return true
	}

	return false
}

// InviteFollowResp 响应跟随请求
func (st *SpaceTeam) InviteFollowResp(user *RoomUser, inviter uint64) bool {
	reqId := user.GetID()

	for _, v := range st.followInvites[inviter] {
		if v == reqId {
			return st.FollowParachute(user, inviter)
		}
	}

	return false
}

// FollowParachute 跟随队友跳伞
func (st *SpaceTeam) FollowParachute(user *RoomUser, target uint64) bool {
	reqId := user.GetID()
	teamId := user.GetUserTeamID()

	if !st.IsInOneTeamByID(reqId, target) {
		st.space.Warn("FollowParachute failed, not in the same team")
		return false
	}

	if user.GetBaseState() != RoomPlayerBaseState_Inplane {
		st.space.Warn("FollowParachute failed, follower is not inplane, BaseState: ", user.GetBaseState())
		return false
	}

	target = st.GetFollowTarget(teamId, target)
	if target == reqId {
		return false
	}

	followers := st.followRelations[target]
	for _, follower := range followers {
		if follower == reqId {
			return false
		}
	}

	targetUser, ok := st.space.GetEntity(target).(*RoomUser)
	if !ok {
		st.space.Warn("FollowParachute failed, target is not a player, Id: ", target)
		return false
	}

	if targetUser.GetBaseState() != RoomPlayerBaseState_Inplane {
		st.space.Warn("FollowParachute failed, target is not inplane, BaseState: ", targetUser.GetBaseState())
		return false
	}

	oldTarget := st.GetFollowTarget(teamId, reqId)
	if oldTarget != reqId && oldTarget != target {
		st.CancelFollowParachute(user)
	}

	followers = append(followers, reqId)
	user.AdviceNotify(common.NotifyCommon, 81)

	reqFollowers := st.followRelations[reqId]
	if len(reqFollowers) > 0 {
		followers = append(followers, reqFollowers...)
		delete(st.followRelations, reqId)
	}

	st.followRelations[target] = followers
	st.TeamFollowBroad(user.GetUserTeamID())

	targetUser.parachuteType = 1
	user.parachuteType = 2

	return true
}

// CancelFollowParachute 取消跟随关系
func (st *SpaceTeam) CancelFollowParachute(user *RoomUser) {
	reqId := user.GetID()
	teamId := user.GetUserTeamID()

	if user.IsFollowedParachute() { //被跟随

		if user.GetBaseState() == RoomPlayerBaseState_Inplane {
			user.parachuteType = 0
		}

		followers := st.followRelations[reqId]
		delete(st.followRelations, reqId)
		st.TeamFollowBroad(teamId)

		for _, v := range followers {
			followerUser := st.space.GetEntity(v).(*RoomUser)
			followerUser.AdviceNotify(common.NotifyCommon, 80)

			if followerUser.GetBaseState() == RoomPlayerBaseState_Inplane {
				followerUser.parachuteType = 0
			} else {
				followerUser.parachuteType = 3
			}
		}

	} else if user.IsFollowingParachute() { //跟随

		target := st.GetFollowTarget(teamId, reqId)
		followers := st.followRelations[target]

		for i, v := range followers {
			if v == reqId {
				followers = append(followers[:i], followers[i+1:]...)

				if len(followers) > 0 {
					st.followRelations[target] = followers
				} else {
					delete(st.followRelations, target)
				}

				st.TeamFollowBroad(teamId)

				if user.GetBaseState() == RoomPlayerBaseState_Inplane {
					user.parachuteType = 0

					if len(followers) == 0 {
						targetUser := st.space.GetEntity(target).(*RoomUser)
						targetUser.parachuteType = 0
					}
				} else {
					user.parachuteType = 3
				}

				break
			}
		}
	}
}

// GetFollowTarget 获取跟随目标
func (st *SpaceTeam) GetFollowTarget(teamid, id uint64) uint64 {
	if !st.isTeam {
		return id
	}

	_, ok := st.followRelations[id]
	if ok {
		return id
	}

	team, ok := st.teams[teamid]
	if !ok {
		st.space.Error("GetFollowTarget failed, team id not exist, id: ", teamid)
		return id
	}

	for _, memid := range team {
		tmpUser, ok := st.space.GetEntityByDBID("Player", memid).(*RoomUser)
		if !ok {
			continue
		}

		tmpId := tmpUser.GetID()
		for _, follower := range st.followRelations[tmpId] {
			if follower == id {
				return tmpId
			}
		}
	}

	return id
}

// GetTeamFollowDetail 获取队伍内的跟随关系
func (st *SpaceTeam) GetTeamFollowDetail(teamId uint64) *protoMsg.FollowDetail {
	notify := &protoMsg.FollowDetail{}

	if !st.isTeam {
		return notify
	}

	team, ok := st.teams[teamId]
	if !ok {
		st.space.Error("TeamFollowBroad failed, team id not exist, id: ", teamId)
		return notify
	}

	for _, memid := range team {
		tmpUser, ok := st.space.GetEntityByDBID("Player", memid).(*RoomUser)
		if !ok {
			continue
		}

		followers, ok := st.followRelations[tmpUser.GetID()]
		if !ok {
			continue
		}

		notify.Infos = append(notify.Infos, &protoMsg.FollowInfo{
			Target:    tmpUser.GetID(),
			Followers: followers,
		})
	}

	return notify
}

// TeamFollowBroad 在队伍内广播跟随关系
func (st *SpaceTeam) TeamFollowBroad(teamId uint64) {
	if !st.isTeam {
		return
	}

	team, ok := st.teams[teamId]
	if !ok {
		st.space.Error("TeamFollowBroad failed, team id not exist, id: ", teamId)
		return
	}

	notify := st.GetTeamFollowDetail(teamId)

	for _, memid := range team {
		tmpUser, ok := st.space.GetEntityByDBID("Player", memid).(*RoomUser)
		if !ok {
			continue
		}

		tmpUser.RPC(iserver.ServerTypeClient, "TeamFollowNotify", notify)
	}

	st.space.Debugf("TeamFollowNotify: %+v\n", notify)
}

// TeamFollowSync 向玩家同步队伍内的跟随关系
func (st *SpaceTeam) TeamFollowSync(user *RoomUser) {
	if !st.isTeam {
		return
	}

	notify := st.GetTeamFollowDetail(user.GetUserTeamID())

	if notify != nil {
		user.RPC(iserver.ServerTypeClient, "TeamFollowNotify", notify)
		st.space.Debugf("TeamFollowNotify: %+v\n", notify)
	}
}

// TeamFollowFinish 跟随目标着地，通知跟随相关玩家
func (st *SpaceTeam) TeamFollowFinish(target uint64) {
	st.FollowFinish(target)

	followers := st.followRelations[target]
	for _, follower := range followers {
		st.FollowFinish(follower)
	}
}

// FollowFinish 跟随结束，通知客户端
func (st *SpaceTeam) FollowFinish(id uint64) {
	user, ok := st.space.GetEntity(id).(*RoomUser)
	if !ok {
		return
	}

	user.RPC(iserver.ServerTypeClient, "FollowFinishNotify")
}

// TeamFollowParachuteCtrl 场景内队伍跟随跳伞控制
func (st *SpaceTeam) TeamFollowParachuteCtrl() {
	for target, followers := range st.followRelations {
		targetUser, ok := st.space.GetEntity(target).(*RoomUser)
		if !ok {
			continue
		}

		if targetUser.GetBaseState() == RoomPlayerBaseState_Inplane {
			continue
		}

		if targetUser.GetBaseState() != RoomPlayerBaseState_Glide &&
			targetUser.GetBaseState() != RoomPlayerBaseState_Parachute {
			st.TeamFollowFinish(target)
			delete(st.followRelations, target)
			st.TeamFollowBroad(targetUser.GetUserTeamID())
			continue
		}

		if targetUser.parachuteCtrl.glideStartTime.Add(100 * time.Millisecond).After(time.Now()) {
			targetUser.Debug("Time Not Enought  ")
			continue
		}

		targetState := targetUser.GetPlayerState()
		cancelers := []uint64{}

		offsetX := float32(0)
		offsetY := float32(0)
		offsetZ := float32(0)

		for _, follower := range followers {
			followerUser, ok := st.space.GetEntity(follower).(*RoomUser)
			if !ok {
				continue
			}

			oldBaseState := followerUser.GetBaseState()
			if oldBaseState != RoomPlayerBaseState_Inplane &&
				oldBaseState != RoomPlayerBaseState_Glide &&
				oldBaseState != RoomPlayerBaseState_Parachute {
				cancelers = append(cancelers, follower)
				continue
			}

			followerState := followerUser.GetPlayerState()
			targetState.CopyTo(followerState)

			offsetX += st.space.followParachuteOffsetX
			offsetY += st.space.followParachuteOffsetY
			offsetZ += st.space.followParachuteOffsetZ

			pos := targetState.GetPos()
			pos.X -= offsetX
			pos.Y += offsetY
			pos.Z -= offsetZ

			followerUser.SetPos(pos)
			if oldBaseState == RoomPlayerBaseState_Inplane {
				if followerUser.parachuteCtrl == nil {
					followerUser.parachuteCtrl = NewParachuteCtrl(followerUser)
					followerUser.parachuteCtrl.StartGlide()
				}
				st.space.BroadAirLeft()
			}

			if oldBaseState == RoomPlayerBaseState_Glide && followerUser.GetBaseState() == RoomPlayerBaseState_Parachute {
				followerUser.parachuteCtrl.StartParachute()
			}
		}

		for _, canceler := range cancelers {
			followerUser, ok := st.space.GetEntity(canceler).(*RoomUser)
			if !ok {
				continue
			}

			st.CancelFollowParachute(followerUser)
			st.FollowFinish(canceler)
		}
	}
}

// NotifyUserOffline通知队友玩家掉线
func (st *SpaceTeam) NotifyUserOffline(user *RoomUser) {
	if st.isTeam == false {
		return
	}

	// team, ok := st.teams[user.GetUserTeamID()]
	// if ok == false {
	// 	return
	// }

	// for _, memid := range team {
	// 	tmpUser, ok := st.space.GetEntityByDBID("Player", memid).(*RoomUser)

	// 	if !ok || user.GetDBID() == memid {
	// 		continue
	// 	}
	// 	tmpUser.RPC(iserver.ServerTypeClient, "ShowDieFlag", user.GetID(), float64(user.GetPos().X), float64(user.GetPos().Z))
	// 	log.Debug("通知掉线", user.GetID(), float64(user.GetPos().X), float64(user.GetPos().Z))
	// }
}

// NotifyUserOnline通知队友玩家上线
func (st *SpaceTeam) NotifyUserOnline(user *RoomUser) {
	if st.isTeam == false {
		return
	}

	// team, ok := st.teams[user.GetUserTeamID()]
	// if ok == false {
	// 	return
	// }

	// for _, memid := range team {
	// 	tmpUser, ok := st.space.GetEntityByDBID("Player", memid).(*RoomUser)

	// 	if !ok || user.GetDBID() == memid {
	// 		continue
	// 	}
	// 	tmpUser.RPC(iserver.ServerTypeClient, "ShowDieFlag", user.GetID(), float64(user.GetPos().X), float64(user.GetPos().Z))
	// 	log.Debug("通知上线", user.GetID(), float64(user.GetPos().X), float64(user.GetPos().Z))
	// }
}

// getTeamPlayerSum 获取每个队伍玩家数
func (st *SpaceTeam) getTeamPlayerSum() uint32 {
	if st.isTeam == false {
		//st.space.Error("getTeamPlayerSum failed, not a team")
		return 1
	}

	if st.teamType == 0 {
		return 2
	} else if st.teamType == 1 {
		return 4
	}

	return 1
}

// DisposeTeamSettle 处理队伍结算
func (st *SpaceTeam) DisposeTeamSettle(teamid uint64, bVictory bool) {

	if st.isTeam == false {
		//st.space.Error("DisposeTeamSettle failed, not a team")
		return
	}

	team, ok := st.teams[teamid]
	if ok == false {
		st.space.Error("DisposeTeamSettle failed, team not exist, teamid: ", teamid)
		return
	}

	if bVictory == false {
		for _, memid := range team {
			user, ok := st.space.GetEntityByDBID("Player", memid).(*RoomUser)
			if !ok || user.userType == RoomUserTypeWatcher {
				continue
			}

			curState := user.stateMgr.GetState()

			if curState != RoomPlayerBaseState_Dead && curState != RoomPlayerBaseState_Watch && curState != RoomPlayerBaseState_WillDie {
				st.space.Error("DisposeTeamSettle failed, user state is not ok, name: ", user.GetName(), " uid: ", user.GetDBID(), " state: ", curState)
			}

			if curState == RoomPlayerBaseState_WillDie {
				// 终止玩家濒死状态
				user.DisposeDeath(user.stateMgr.injuredtype, user.stateMgr.attactID, false)
				attacker, ok := st.space.GetEntity(user.stateMgr.attactID).(*RoomUser)
				if ok {
					if !(st.IsInOneTeam(user.GetDBID(), attacker.GetDBID()) || user.GetID() == user.stateMgr.attactID) {
						attacker.DisposeIncrKillNum()
						if user.stateMgr.isHeadShot && user.stateMgr.attactID == user.stateMgr.downAttacker {
							attacker.IncrHeadShotNum()
						}
					}
				}
				st.space.BroadDieNotify(user.stateMgr.attactID, user.GetID(), false, InjuredInfo{injuredType: user.stateMgr.injuredtype, isHeadshot: false})
			}

		}
	}

	teamSettleInfo := st.SpaceTeamSettleInfo(teamid) // 队伍个人信息

	// 通知结算
	for _, memid := range team {
		user, ok := st.space.GetEntityByDBID("Player", memid).(*RoomUser)
		if !ok {
			continue
		}

		if bVictory {
			user.SetVictory()
		}

		user.NotifySettle()

		// 通知组队结算信息
		if teamSettleInfo != nil {
			user.RPC(iserver.ServerTypeClient, "TeamSettleInfo", teamSettleInfo)
			user.BroadcastToWatchers(RoomUserTypeWatcher, "TeamSettleInfo", teamSettleInfo)
		}
	}
}

// SpaceTeamSettleInfo 队伍信息获取
func (st *SpaceTeam) SpaceTeamSettleInfo(teamid uint64) *protoMsg.SettleInfo {
	//log.Debug("SpaceTeamSettleInfo")
	if st.isTeam == false {
		//st.space.Error("DisposeTeamSettle failed, not a team")
		return nil
	}

	team, ok := st.teams[teamid]
	if ok == false {
		st.space.Error("SpaceTeamSettleInfo failed, team not exist, teamid: ", teamid)
		return nil
	}

	// 统计组队结算信息
	for _, memid := range team {
		user, ok := st.space.GetEntityByDBID("Player", memid).(*RoomUser)
		if !ok {
			continue
		}

		st.AddSpaceMemberInfo(user)
	}

	// 队伍个人信息
	teamSettleInfo := &protoMsg.SettleInfo{}
	for _, memid := range team {
		info := st.GetSpaceMemberInfo(memid)
		if info == nil {
			continue
		}
		memberinfo := &protoMsg.SettleMemInfo{}
		memberinfo.Uid = info.memberid
		memberinfo.Name_ = info.username
		memberinfo.Url = info.url
		memberinfo.Killnum = info.killnum
		memberinfo.Headshotnum = info.headshotnum
		memberinfo.Damagehp = info.damagehp
		memberinfo.Distance = info.distance
		memberinfo.Gametime = info.gametime
		memberinfo.Veteran = info.veteran
		memberinfo.NameColor = info.namecolor

		teamSettleInfo.List = append(teamSettleInfo.List, memberinfo)
	}
	return teamSettleInfo
}

// IsInOneTeam 是否在同一队伍
func (st *SpaceTeam) IsInOneTeam(playerid1, playerid2 uint64) bool {
	if st.isTeam == false {
		//st.space.Error("SpaceTeamSettleInfo failed, not a team ")
		return false
	}

	for _, team := range st.teams {
		count := 0
		for _, memid := range team {
			if memid == playerid1 || memid == playerid2 {
				count++
			}

			if count == 2 {
				return true
			}
		}
	}

	return false
}

//IsInOneTeamByID 是否在同一队伍
func (st *SpaceTeam) IsInOneTeamByID(playerid1, playerid2 uint64) bool {
	if st.isTeam == false {
		return false
	}

	player1, _ := st.space.GetEntity(playerid1).(*RoomUser)
	player2, _ := st.space.GetEntity(playerid2).(*RoomUser)
	if player1 == nil || player2 == nil {
		return false
	}

	teamid1 := player1.GetUserTeamID()
	teamid2 := player2.GetUserTeamID()

	return teamid1 == teamid2
}

// TeammateLeaveSpace 队友离开场景
func (st *SpaceTeam) TeammateLeaveSpace(user *RoomUser, teamid uint64) {
	if user == nil {
		return
	}

	team, ok := st.teams[teamid]
	if ok == false {
		st.space.Error("TeammateLeaveSpace failed, team not exist, teamid: ", teamid)
		return
	}

	for _, memid := range team {
		targetUser, ok := st.space.GetEntityByDBID("Player", memid).(*RoomUser)
		if !ok || user == targetUser || targetUser.userType == RoomUserTypeWatcher {
			continue
		}
		targetUser.RPC(iserver.ServerTypeClient, "TeammateLeaveSpace", user.GetID())

		targetUser.RemoveExtWatchEntity(user)
		targetUser.NotifyWatcherTeamMateLeave(user)
		//log.Debug("RemoveExtWatchEntity ", user.GetName(), " ", targetUser.GetName())
	}
}

// GetTeamIDByPlayerID 根据dbid获取队伍id
func (st *SpaceTeam) GetTeamIDByPlayerID(playerid uint64) uint64 {

	for i, j := range st.teams {
		for _, memid := range j {
			if playerid == memid {
				return i
			}
		}
	}

	return 0
}

// AddSpaceMemberInfo 添加场景成员信息
func (st *SpaceTeam) AddSpaceMemberInfo(user *RoomUser) {

	// 不重复填充数据
	_, ok := st.memberInfo[user.GetDBID()]
	if ok {
		st.space.Error("AddSpaceMemberInfo failed, member id not exist, id: ", user.GetDBID())
		return
	}

	info := &SpaceMemberInfo{}
	info.memberid = user.GetDBID()
	info.username = user.GetName()

	args := []interface{}{
		"Picture",
		"Veteran",
	}
	values, valueErr := dbservice.EntityUtil("Player", info.memberid).GetValues(args)
	if valueErr != nil || len(values) != len(args) {
		st.space.Error("AddSpaceMemberInfo failed, can't get url, memberid: ", info.memberid)
		return
	}
	tmpUrl, urlErr := redis.String(values[0], nil)
	if urlErr == nil {
		info.url = tmpUrl
	}
	info.killnum = user.kill
	info.headshotnum = user.headshotnum
	info.damagehp = user.effectharm
	info.distance = user.sumData.runDistance / 1000
	if user.sumData.endgametime == user.sumData.begingametime {
		info.gametime = time.Now().Unix() - user.sumData.begingametime
	} else {
		info.gametime = user.sumData.endgametime - user.sumData.begingametime
	}
	info.rating = float32(st.space.scores[user.GetDBID()])
	veteran, veteranErr := redis.Uint64(values[1], nil)
	if veteranErr == nil {
		info.veteran = uint32(veteran)
	}
	info.namecolor = common.GetPlayerNameColor(info.memberid)

	st.memberInfo[user.GetDBID()] = info
}

// GetSpaceMemberInfo 获取空间成员信息
func (st *SpaceTeam) GetSpaceMemberInfo(memberid uint64) *SpaceMemberInfo {

	info, ok := st.memberInfo[memberid]
	if !ok {
		st.space.Error("GetSpaceMemberInfo failed, member id not exist, id: ", memberid)
		return nil
	}

	return info
}

// DelRedisVoiceInfo 游戏结束时删除语音信息
func (st *SpaceTeam) DelRedisVoiceInfo() {
	for teamId, _ := range st.teams {
		//log.Info("DelTeamVoiceInfo ", teamId)
		db.DelTeamVoiceInfo(teamId)
	}
}

// DelTeamInfo 游戏结束时删除相关记录数据
func (st *SpaceTeam) DelTeamInfo() {
	for teamId, _ := range st.teams {
		db.PlayerTeamUtil(teamId).Remove()
	}
}

// isComradeGame 玩家是否与战友组队出战
func (st *SpaceTeam) isComradeGame(user *RoomUser) bool {
	if !st.isTeam {
		return false
	}

	team, ok := st.teams[user.GetUserTeamID()]
	if !ok {
		return false
	}

	util := db.PlayerInfoUtil(user.GetDBID())
	for _, memid := range team {
		if util.IsTeacherPupil(memid) {
			return true
		}
	}

	return false
}

// isComrade 是否是同一个队伍内的战友
func (st *SpaceTeam) isComrade(uid1, uid2 uint64) bool {
	if !st.isTeam {
		return false
	}

	if !st.IsInOneTeam(uid1, uid2) {
		return false
	}

	return db.PlayerInfoUtil(uid1).IsTeacherPupil(uid2)
}

package main

import (
	"common"
	"db"
	"fmt"
	"protoMsg"
	"sync"
	"time"
	"zeus/entity"
	"zeus/iserver"
	"zeus/msgdef"

	log "github.com/cihub/seelog"
)

// IMatcher 匹配对象
type IMatcher interface {
	iserver.IEntityProxy

	GetID() uint64
	GetIDs() []uint64
	GetDBIDs() []uint64
	GetMMR() uint32
	GetRank() uint32
	GetFourRank() uint32
	GetMapID() uint32
	GetNums() int
	IsReady() bool

	onMatchSuccess(uint64, uint32, uint32, uint32, uint32)
	onRoomSceneInited(uint64)
	GetMatchTime() int64
	IsSingleMatch() bool
	GetMatchMode() uint32
	GetTeamType() uint8
	getScoreStr() string
}

// MatchMember 单个匹配成员
type MatchMember struct {
	*entity.EntityProxy

	mmr       uint32
	rank      uint32 //单双排用
	fourrank  uint32 //四排用
	mapid     uint32
	isReady   bool
	matchTime int64
	isVeteran uint32

	state         uint8
	name          string
	role          uint32
	dbid          uint64
	nameColor     uint32
	outsideWeapon uint32

	matchMode      uint32 //匹配模式，具体定义在common/Const.go
	stateEnterTeam uint32 //进入队伍方式 0单人参加组队模式， 1系统分配， 2与好友组队参赛
	originalTeam   uint64 //如果发生队伍合并 记录原有的队伍

	gameRecord *protoMsg.GameRecordDetail //比赛数据记录
}

// NewMatchMember 新建一个匹配成员
func NewMatchMember(srvID, entityID uint64, mmr, tworank, fourrank, mapid uint32, name string, role uint32, dbid uint64, matchMode, isVeteran, color, weapon uint32) *MatchMember {
	mm := &MatchMember{}
	mm.EntityProxy = entity.NewEntityProxy(srvID, 0, entityID)
	mm.mmr = mmr
	mm.rank = tworank
	mm.fourrank = fourrank
	mm.mapid = mapid
	mm.matchTime = time.Now().Unix()
	mm.isVeteran = isVeteran
	mm.state = MemberNotReady
	mm.name = name
	mm.role = role
	mm.dbid = dbid
	mm.stateEnterTeam = 0
	mm.matchMode = matchMode
	mm.gameRecord = &protoMsg.GameRecordDetail{}
	mm.nameColor = color
	mm.outsideWeapon = weapon
	return mm
}

func (mm *MatchMember) String() string {
	return fmt.Sprintf("{SrvID:%d SpaceID:%d EntityID:%d MMR:%d Rank:%d}", mm.SrvID, mm.SpaceID, mm.EntityID, mm.mmr, mm.rank)
}

// GetID 获取ID
func (mm *MatchMember) GetID() uint64 {
	return mm.EntityID
}

// GetIDs 获取成员ID列表
func (mm *MatchMember) GetIDs() []uint64 {
	ids := make([]uint64, 0, 1)
	ids = append(ids, mm.EntityID)
	return ids
}

// GetDBIDs 获取成员DBID
func (mm *MatchMember) GetDBIDs() []uint64 {
	ids := make([]uint64, 0, 1)
	ids = append(ids, mm.dbid)
	return ids
}

// GetMMR 获取MMR
func (mm *MatchMember) GetMMR() uint32 {
	return mm.mmr
}

// GetRank 获取等级
func (mm *MatchMember) GetRank() uint32 {
	return mm.rank
}

func (mm *MatchMember) GetFourRank() uint32 {
	return mm.fourrank
}

func (mm *MatchMember) getScoreStr() string {
	return common.Uint64ToString(uint64(mm.GetMatchMode())) + "|" + common.Uint64ToString(uint64(mm.rank))
}

// GetNums 获取玩家数
func (mm *MatchMember) GetNums() int {
	return 1
}

// GetMapID 获取mapid
func (mm *MatchMember) GetMapID() uint32 {
	return mm.mapid
}

// IsReady 是否已经准备好
func (mm *MatchMember) IsReady() bool {
	return mm.isReady
}

// SetReady 设置是否已经准备好
func (mm *MatchMember) SetReady(ready bool) {
	mm.isReady = ready
}

// 通知匹配成功
func (mm *MatchMember) onMatchSuccess(spaceID uint64, mapid uint32, skybox uint32, mode, realNum uint32) {
	matchTime := time.Now().Unix() - mm.matchTime
	if err := mm.RPC(common.ServerTypeLobby, "MatchSuccess", spaceID, mapid, skybox, mode, realNum, matchTime, mm.stateEnterTeam); err != nil {
		log.Error(err, mm.EntityProxy)
	}
}

// 房间创建成功, 通知玩家进入场景
func (mm *MatchMember) onRoomSceneInited(spaceid uint64) {
	if err := mm.RPC(common.ServerTypeLobby, "EnterScene", spaceid); err != nil {
		log.Error(err, mm.EntityProxy)
	}
}

// GetMatchTime 获取匹配时间
func (mm *MatchMember) GetMatchTime() int64 {
	return mm.matchTime
}

// IsSingleMatch 是否是单人
func (mm *MatchMember) IsSingleMatch() bool {
	return true
}

// GetMatchMode 获取匹配模式
func (mm *MatchMember) GetMatchMode() uint32 {
	return mm.matchMode
}

// GetTeamType 获取队伍类型
func (mm *MatchMember) GetTeamType() uint8 {
	return 100
}

///////////////////////////////////////////////////////////////////////////////

// MatchTeam 组队匹配对象
type MatchTeam struct {
	// members map[uint64]*MatchMember
	members sync.Map

	teamid    uint64
	typ       uint8
	mapid     uint32
	nums      int
	automatch bool   // 自动匹配队友
	matchMode uint32 // 匹配模式，具体定义在common/Const.go

	avgMMR  uint32
	avrRank uint32

	leaderid     uint64
	teamType     uint8
	teamStatus   uint32   // 组队准备状态 0、正在组队状态 1、匹配状态
	matchtingime int64    // 进入匹配时间戳
	orders       []uint64 // 记录队员加入该team的顺序
}

// NewMatchTeam 新建一个队伍匹配对象
func NewMatchTeam(teamid uint64, mapid uint32, totalNum uint32, teamtype, automatch uint32, matchMode uint32) *MatchTeam {
	mt := &MatchTeam{}
	// mt.members = make(map[uint64]*MatchMember)
	mt.teamid = teamid
	mt.mapid = mapid
	mt.nums = int(totalNum)

	mt.teamType = uint8(teamtype)
	/*
		if !viper.GetBool("Lobby.Duo") {
			mt.teamType = FourTeamType
		}
	*/

	mt.automatch = (automatch != 0)
	mt.teamStatus = TeamNotMatch
	mt.orders = []uint64{}
	mt.matchMode = matchMode

	return mt
}

// GetID 获取ID
func (mt *MatchTeam) GetID() uint64 {
	return mt.teamid
}

// GetIDs 获取成员ID列表
func (mt *MatchTeam) GetIDs() []uint64 {
	ids := make([]uint64, 0, 1)
	// for _, mm := range mt.members {
	// 	ids = append(ids, mm.GetID())
	// }
	mt.members.Range(
		func(k, v interface{}) bool {
			mm := v.(IMatcher)
			ids = append(ids, mm.GetID())
			return true
		})
	return ids
}

// GetDBIDs 获取成员DBID列表
func (mt *MatchTeam) GetDBIDs() []uint64 {
	ids := make([]uint64, 0, 1)
	// for _, mm := range mt.members {
	// 	ids = append(ids, mm.dbid)
	// }
	mt.members.Range(
		func(k, v interface{}) bool {
			mm := v.(*MatchMember)
			ids = append(ids, mm.dbid)
			return true
		})
	return ids
}

// GetMMR 获取MMR
func (mt *MatchTeam) GetMMR() uint32 {
	return mt.avgMMR
}

// GetRank 获取等级
func (mt *MatchTeam) GetRank() uint32 {
	if mt.matchMode == common.MatchModeBrave {
		return mt.GetTopRank()
	}
	return mt.avrRank
}

// GetFourRank 获取等级
func (mt *MatchTeam) GetFourRank() uint32 {
	return mt.avrRank
}

// GetTopRank 获取team中最高的rank
func (mt *MatchTeam) GetTopRank() uint32 {
	var max uint32

	mt.members.Range(
		func(k, v interface{}) bool {
			member := v.(*MatchMember)
			if mt.teamType == TwoTeamType {
				if member.GetRank() > max {
					max = member.GetRank()
				}
			} else if mt.teamType == FourTeamType {
				if member.GetFourRank() > max {
					max = member.GetFourRank()
				}
			}
			return true
		})

	return max
}

func (mt *MatchTeam) getScoreStr() string {
	var ret string
	mt.members.Range(
		func(k, v interface{}) bool {
			mm := v.(*MatchMember)
			ret += common.Uint64ToString(uint64(mm.GetMatchMode())) + "|"
			if mt.teamType == TwoTeamType {
				ret += common.Uint64ToString(uint64(mm.rank))
			} else {
				ret += common.Uint64ToString(uint64(mm.fourrank))
			}
			ret += ","
			return true
		})

	return ret
}

// GetMapID 获取地图ID
func (mt *MatchTeam) GetMapID() uint32 {
	return mt.mapid
}

// GetNums 获取玩家数
func (mt *MatchTeam) GetNums() int {
	// return len(mt.members)
	var ret int
	mt.members.Range(
		func(k, v interface{}) bool {
			ret++
			return true
		})
	return ret
}

func (mt *MatchTeam) onMatchSuccess(spaceID uint64, mapid uint32, skybox uint32, mode, realNum uint32) {
	// for _, m := range mt.members {
	// 	m.onMatchSuccess(spaceID, mapid, skybox, mode, realNum)
	// }
	mt.members.Range(
		func(k, v interface{}) bool {
			m := v.(IMatcher)
			m.onMatchSuccess(spaceID, mapid, skybox, mode, realNum)
			return true
		})
}

func (mt *MatchTeam) onRoomSceneInited(spaceid uint64) {
	// for _, m := range mt.members {
	// 	m.onRoomSceneInited(spaceid)
	// }
	mt.members.Range(
		func(k, v interface{}) bool {
			m := v.(IMatcher)
			m.onRoomSceneInited(spaceid)
			return true
		})
}

// Add 加入队伍成员
func (mt *MatchTeam) Add(m *MatchMember) {
	// _, ok := mt.members[m.GetID()]
	// if ok {
	// 	log.Warn("已在队伍中 ", m.GetID())
	// 	return
	// }

	// mt.members[m.GetID()] = m
	// if len(mt.members) == 1 {
	// 	mt.leaderid = m.GetID()
	// }

	// if len(mt.members) > 2 {
	// 	mt.teamType = FourTeamType
	// }

	_, ok := mt.members.Load(m.GetID())
	if ok {
		log.Warn("已在队伍中 ", m.GetID())
		return
	}

	if mt.GetMatchMode() != m.GetMatchMode() {
		log.Error("team's matchMode is ", mt.GetMatchMode(), " while new comer's is ", m.GetMatchMode())
		return
	}

	mt.members.Store(m.GetID(), m)
	if mt.GetNums() == 1 {
		mt.leaderid = m.GetID()
	}

	if mt.GetNums() > 2 {
		mt.teamType = FourTeamType
		db.PlayerTeamUtil(mt.GetID()).SetTeamType(mt.teamType)

		//普通模式之外的其他模式使用的地图均从赛事模式配置表中读取
		if mt.GetMatchMode() != common.MatchModeNormal {
			info := common.GetOpenModeInfo(mt.GetMatchMode(), 4)
			if info == nil {
				return
			}

			mt.mapid = info.MapId
		}
	}

	mt.orders = append(mt.orders, m.GetID())
	mt.calcAvgMMR()

}

// remove 移除队伍成员
func (mt *MatchTeam) Remove(id uint64) {
	// _, ok := mt.members[id]
	// if !ok {
	// 	log.Warn("不在队伍中 ", id)
	// 	return
	// }

	// delete(mt.members, id)
	// mt.calcAvgMMR()

	// for _, mem := range mt.members {
	// 	mem.state = MemberNotReady
	// }

	// if mt.leaderid == id {
	// 	var intotime int64
	// 	for _, mem := range mt.members {
	// 		if intotime == 0 || mem.matchTime < intotime {
	// 			intotime = mem.matchTime
	// 			mt.leaderid = mem.GetID()
	// 		}
	// 	}
	// }

	_, ok := mt.members.Load(id)
	if !ok {
		log.Warn("不在队伍中 ", id)
		return
	}

	for i := 0; i < len(mt.orders); i++ {
		if mt.orders[i] == id {
			mt.orders = append(mt.orders[:i], mt.orders[i+1:]...)
		}
	}

	mt.members.Delete(id)

	mt.members.Range(
		func(k, v interface{}) bool {
			mem := v.(*MatchMember)
			if mem.EntityID == mt.leaderid {
				mem.state = MemberNotReady
			}
			return true
		})

	if mt.leaderid == id {
		var intotime int64
		var tmpMem *MatchMember
		mt.members.Range(
			func(k, v interface{}) bool {
				mem := v.(*MatchMember)
				if intotime == 0 || mem.matchTime < intotime {
					intotime = mem.matchTime
					mt.leaderid = mem.GetID()
					mem.state = MemberNotReady
					tmpMem = mem
				}
				return true
			})

		if tmpMem != nil {
			tmpMem.stateEnterTeam = 0
		}
	}

	mt.calcAvgMMR()
	mt.teamStatus = TeamNotMatch
}

// GetMember 获取队伍成员
func (mt *MatchTeam) GetMember(entityID uint64) *MatchMember {
	// member, ok := mt.members[entityID]
	v, ok := mt.members.Load(entityID)
	if !ok {
		log.Errorf("获取队伍成员失败, 玩家不在队伍中 TeamID:%d EntityID:%d", mt.teamid, entityID)
		return nil
	}

	member := v.(*MatchMember)
	return member
}

// IsReady 是否所有人都准备就绪
func (mt *MatchTeam) IsReady() bool {
	// for _, mm := range mt.members {
	// 	if !mm.IsReady() {
	// 		return false
	// 	}
	// }
	// return true
	return mt.teamStatus == TeamWatingScene
}

// IsEmpty 是否为空队伍
func (mt *MatchTeam) IsEmpty() bool {
	// return len(mt.members) == 0
	return mt.GetNums() == 0
}

// IsFull 是否满
func (mt *MatchTeam) IsFull() bool {
	if mt.teamType == TwoTeamType {
		// return len(mt.members) >= 2
		return mt.GetNums() >= 2
	} else if mt.teamType == FourTeamType {
		// return len(mt.members) >= 4
		return mt.GetNums() >= 4
	}

	return false
}

func (mt *MatchTeam) getEntityID(dbid uint64) uint64 {
	var ret uint64

	mt.members.Range(
		func(k, v interface{}) bool {
			mm := v.(*MatchMember)
			if mm.dbid == dbid {
				ret = mm.EntityID
			}
			return true
		})
	return ret
}

// IsExisted 玩家是否在队伍中
func (mt *MatchTeam) IsExisted(entityID uint64) bool {
	// _, ok := mt.members[entityID]
	_, ok := mt.members.Load(entityID)
	return ok
}

// Post 给队伍中所有成员发送消息
func (mt *MatchTeam) Post(msg msgdef.IMsg) error {
	// for _, mm := range mt.members {
	// 	if err := mm.Post(msg); err != nil {
	// 		log.Error(err, mm)
	// 	}
	// }

	mt.members.Range(
		func(k, v interface{}) bool {
			mm := v.(*MatchMember)
			if err := mm.Post(msg); err != nil {
				log.Error(err, mm)
			}
			return true
		})
	return nil
}

// RPC 调用所有成员的RPC方法
func (mt *MatchTeam) RPC(srvType uint8, methodName string, args ...interface{}) error {
	// for _, mm := range mt.members {
	// 	if err := mm.RPC(srvType, methodName, args...); err != nil {
	// 		log.Error(err, mm)
	// 	}
	// }

	mt.members.Range(
		func(k, v interface{}) bool {
			mm := v.(*MatchMember)
			if err := mm.RPC(srvType, methodName, args...); err != nil {
				log.Error(err, mm)
			}
			return true
		})
	return nil
}

func (mt *MatchTeam) calcAvgMMR() {
	var totalMMR uint32
	var totalRank uint32

	var duoNum, squadNum uint32
	duoTwo := common.GetTBMatchValue(common.Match_DuoTwo)
	squadTwo := common.GetTBMatchValue(common.Match_SquadTwo)
	squadThree := common.GetTBMatchValue(common.Match_SquadThree)
	squadFour := common.GetTBMatchValue(common.Match_SquadFour)

	// for _, member := range mt.members {
	// 	totalMMR += member.GetMMR()
	// 	totalRank += member.GetRank()
	// }
	var length uint32
	mt.members.Range(
		func(k, v interface{}) bool {
			member := v.(*MatchMember)
			totalMMR += member.GetMMR()

			if mt.teamType == TwoTeamType {
				totalRank += member.GetRank()

				if member.stateEnterTeam == 2 { //2 好友拉入
					duoNum++
				}
			} else {
				totalRank += member.GetFourRank()

				if member.stateEnterTeam == 2 { //2 好友拉入
					squadNum++
				}
			}

			length++
			return true
		})

	if length != 0 {
		mt.avgMMR = totalMMR / length

		if mt.teamType == TwoTeamType {
			if duoNum == 2 {
				mt.avrRank = uint32(duoTwo * float32(totalRank/length)) //双排两人
			} else {
				mt.avrRank = totalRank / length
			}
		} else {
			if squadNum == 2 {
				mt.avrRank = uint32(squadTwo * float32(totalRank/length)) //四排两人
			} else if squadNum == 3 {
				mt.avrRank = uint32(squadThree * float32(totalRank/length)) //四排三人
			} else if squadNum == 4 {
				mt.avrRank = uint32(squadFour * float32(totalRank/length)) //四排四人
			} else {
				mt.avrRank = totalRank / length
			}
		}
		log.Debug("teamType:", mt.teamType, " duoNum:", duoNum, " squadNum:", squadNum, " avrRank:", mt.avrRank)
	}
}

// GetMatchTime 获取匹配时间
func (mt *MatchTeam) GetMatchTime() int64 {
	return mt.matchtingime
}

// BroadInfo 广播队伍信息
func (mt *MatchTeam) BroadInfo() {
	retMsg := mt.TeamInfo()
	mt.members.Range(
		func(k, v interface{}) bool {
			pUser := v.(*MatchMember)
			if err := pUser.RPC(common.ServerTypeLobby, "SyncTeamInfoRet", retMsg); err != nil {
				log.Error(err)
			}
			return true
		})

	// for _, pUser := range mt.members {

	// 	if pUser == nil {
	// 		continue
	// 	}

	// 	if err := pUser.RPC(common.ServerTypeLobby, "SyncTeamInfoRet", retMsg); err != nil {
	// 		log.Error(err)
	// 	}
	// }

	mt.printTeamInfo()
}

// printTeamInfo 打印队伍信息
func (mt *MatchTeam) printTeamInfo() {

	// log.Info("\n")

	// log.Info("队伍信息", tmModelPtr.id, tmModelPtr.teamType, tmModelPtr.teamStatus, len(tmModelPtr.mems))

	// for i, mem := range tmModelPtr.mems {
	// 	if mem.member == nil {
	// 		continue
	// 	}

	// 	log.Info("队伍中成员信息", i, mem.member.GetID(), mem.memberState)
	// }

	// log.Info("\n")
}

// IsSingleMatch 是否单人
func (mt *MatchTeam) IsSingleMatch() bool {
	return false
}

// GetMatchMode 获取匹配模式
func (mt *MatchTeam) GetMatchMode() uint32 {
	return mt.matchMode
}

// GetTeamType 获取队伍类型
func (mt *MatchTeam) GetTeamType() uint8 {
	return mt.teamType
}

// IsTeamLeader 是否队长
func (mt *MatchTeam) IsTeamLeader(entityID uint64) bool {
	return mt.leaderid == entityID
}

// SetTeamLeader 设置队长
func (mt *MatchTeam) SetTeamLeader(entityID uint64) {
	if mt.leaderid == entityID {
		return
	}
	if oldLeader, ok := mt.members.Load(mt.leaderid); ok {
		oldLeader.(*MatchMember).state = MemberReadying
	}
	mt.leaderid = entityID
	if newLeader, ok := mt.members.Load(entityID); ok {
		newLeader.(*MatchMember).state = MemberNotReady
	}
}

func (mt *MatchTeam) NotifyToMember(entityID uint64) {
	retMsg := mt.TeamInfo()
	mt.members.Range(
		func(k, v interface{}) bool {
			pUser := v.(*MatchMember)
			if pUser.GetID() != entityID {
				return true
			}
			if err := pUser.RPC(common.ServerTypeLobby, "SyncTeamInfoRet", retMsg); err != nil {
				log.Error(err)
			}
			return true
		})
}

func (mt *MatchTeam) TeamInfo() *protoMsg.SyncTeamInfoRet {
	retMsg := &protoMsg.SyncTeamInfoRet{}
	retMsg.Id = mt.teamid

	//客户端只有匹配和未匹配两个状态
	retMsg.TeamState = mt.teamStatus
	if mt.teamStatus == TeamWatingScene {
		retMsg.TeamState = TeamMatching
	}

	retMsg.Leaderid = mt.leaderid
	retMsg.Teamtype = uint32(mt.teamType)
	retMsg.Mapid = mt.mapid
	if mt.automatch {
		retMsg.Automatch = 1
	} else {
		retMsg.Automatch = 0
	}

	// for k, pUser := range mt.members {

	// 	if pUser == nil {
	// 		continue
	// 	}

	// 	memInfo := protoMsg.TeamMemberInfo{
	// 		Uid:      k,
	// 		Name_:    pUser.name,
	// 		MemState: uint32(pUser.state),
	// 		Modelid:  uint64(pUser.role),
	// 	}
	// 	retMsg.Memberinfo = append(retMsg.Memberinfo, &memInfo)
	// }
	for _, id := range mt.orders {
		m, ok := mt.members.Load(id)
		if ok {
			mem := m.(*MatchMember)
			retMsg.Memberinfo = append(retMsg.Memberinfo, &protoMsg.TeamMemberInfo{
				Uid:           mem.GetID(),
				Name_:         mem.name,
				MemState:      uint32(mem.state),
				Modelid:       uint64(mem.role),
				Location:      db.PlayerTempUtil(mem.dbid).GetPlayerLocation(),
				Dbid:          mem.dbid,
				Insignia:      common.GetInsigniaIcon(mem.dbid),
				Veteran:       mem.isVeteran,
				NameColor:     mem.nameColor,
				OutsideWeapon: mem.outsideWeapon,
			})
		}
	}
	return retMsg
}
func (team *MatchTeam) DisposeTeam(iscancel bool) {
	oldTeams := map[uint64]struct{}{}
	team.members.Range(func(key, value interface{}) bool {
		mem := value.(*MatchMember)
		if iscancel {
			if mem.EntityID == team.leaderid {
				mem.state = MemberNotReady
			}
		} else {
			mem.state = MemberNotReady
		}
		if mem.originalTeam != 0 {
			mem.RPC(common.ServerTypeLobby, "SyncTeamID", mem.originalTeam, mem.state)
			oldTeams[mem.originalTeam] = struct{}{}
			mem.originalTeam = 0
			team.Remove(mem.GetID())
		}
		return true
	})
	for k := range oldTeams {
		if t, ok := GetTeamMgr().teams.Load(k); ok {
			if tt, ok := t.(*MatchTeam); ok {
				tt.automatch = true
				if m, o := tt.members.Load(tt.leaderid); o {
					m.(*MatchMember).state = MemberNotReady
				}
				tt.BroadInfo()
			}
		}
	}
	team.BroadInfo()
}

func (team *MatchTeam) RemoveTeamMemberVoiceInfo(entityId uint64) {
	// log.Info("RemoveTeamMemberVoiceInfo ", team.GetID(), " ", uid)
	redismemlist := db.GetTeamVoiceInfo(team.GetID())
	for i, v := range redismemlist {
		if v.EntityId == entityId {
			redismemlist = append(redismemlist[0:i], redismemlist[i+1:]...)
			break
		}
	}

	db.SetTeamVoiceInfo(team.GetID(), redismemlist)

	msg := &protoMsg.TeamVoiceInfo{}
	for _, v := range redismemlist {
		msg.MemberInfos = append(msg.MemberInfos, &protoMsg.MemVoiceInfo{
			MemberId: v.MemberId,
			Uid:      v.EntityId,
		})
	}

	team.members.Range(func(k, v interface{}) bool {
		pUser := v.(*MatchMember)
		if err := pUser.RPC(iserver.ServerTypeClient, "SyncTeamVoiceInfo", msg); err != nil {
			log.Error(err)
		}
		// log.Info("client ", pUser.GetID(), " members ", msg.MemberInfos)
		return true
	})
}

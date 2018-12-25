package main

import (
	"common"
	"db"
	"errors"
	"excel"
	"fmt"
	"math"
	"math/rand"
	"sort"
	"time"
	"zeus/iserver"
	"zeus/linmath"
	"zeus/space"

	log "github.com/cihub/seelog"
	"github.com/garyburd/redigo/redis"
)

const (
	//TeamSpace 组队场景
	TeamSpace = 1

	//NotTeamSpace 非组队场景
	NotTeamSpace = 2
)

// 轮询 table
func (srv *RoomSrv) doTravsal() {
	srv.TravsalEntity("Space", func(e iserver.IEntity) {
		if e == nil {
			return
		}
		table := e.(*Scene)
		if table.status == common.SpaceStatusClose {
			srv.DestroyEntity(table.GetID())
		}
	})
}

// SceneItem 场景中的道具
type SceneItem struct {
	id         uint64
	pos        linmath.Vector3
	dir        linmath.Vector3
	itemid     uint32
	item       *Item
	haveobjs   map[uint32]*Item
	cangettime int64
}

// Scene 游戏局结构
type Scene struct {
	space.Space
	ISceneData

	status      int            //当前状态，
	createStamp int64          //创建时间
	mapdata     excel.MapsData //地图数据

	mapitem  map[uint64]*SceneItem
	mapindex uint32
	cars     map[uint64]*Car

	safecenter      linmath.Vector3
	saferadius      float32
	maxUserSum      uint32
	entityTempID    uint64
	summonai        int64
	haveai          uint32
	aiNum           uint32
	aiNumSurplus    uint32 // 剩余Ai数量
	onAirAiNum      uint32
	nextsubaitime   int64
	skybox          uint32
	aiNameMap       map[string]bool
	aiNameUsedIndex uint32
	aiNameLib       []int
	members         map[uint64]bool
	closetime       int64

	teamMgr  *SpaceTeam // 组队信息
	uniqueId uint32     //matchMode表中对应匹配模式的唯一id

	airlineStart       linmath.Vector3 // 航线起点
	airlineEnd         linmath.Vector3 // 航线终点
	allowParachuteTime time.Time       // 开启跳伞时间
	allowParachute     bool            //是否允许跳伞
	parachuteTime      int64

	minLoadTime time.Duration
	maxLoadTime time.Duration

	maploadSuccess bool

	refreshItemInst *MapItemRate
	//spaceresult     *SpaceResultMgr
	refreshzone *RefreshZoneMgr
	refreshitem *RefreshItemMgr

	doorMgr          *SpaceDoorMgr
	aiPosALib        []linmath.Vector3
	aiPosBLib        []linmath.Vector3
	aiInRoomASum     uint32
	aiInRoomBSum     uint32
	GameID           uint64
	loadFulPlayerNum uint32             //每局初始化加载成功的玩家数量(统计卡loading的玩家数量)
	averageRating    float64            //本局所有玩家的平均分
	scores           map[uint64]float64 //本局所有玩家的rating

	memberModeScore []*MemberModeScore
	r               *rand.Rand //  随机数

	firstDie  uint64 //本局比赛第一个挂
	firstKill uint64 //本局比赛首杀

	randomAiPos map[uint64]bool //待固定Ai位置map
	randomNum   uint32          //随机Ai位置比例参数

	followParachuteOffsetX float32 //跟随跳伞X方向偏移量
	followParachuteOffsetY float32 //跟随跳伞Y方向偏移量
	followParachuteOffsetZ float32 //跟随跳伞Z方向偏移量

	callBoxItem  map[uint64]CallBoxItem
	secondTicker *time.Ticker

	watchers map[uint64]bool
}

// Init 初始化调用
func (tb *Scene) Init(initParam interface{}) {
	tb.r = rand.New(rand.NewSource(time.Now().UnixNano()))

	tb.RegMsgProc(&SpaceMsgProc{space: tb})

	var (
		mapid     int
		uniqueId  uint32
		mgrName   string
		matchmode uint32
	)

	if _, err := fmt.Sscanf(initParam.(string), "%d:%d:%d:%s", &mapid, &uniqueId, &matchmode, &mgrName); err != nil {
		tb.Error("Invaild init param: ", initParam.(string))
		return
	}

	mapdata, ok := excel.GetMaps(uint64(mapid))
	if !ok {
		return
	}

	tb.ISceneData = NewSceneData(tb, matchmode)
	tb.mapdata = mapdata
	tb.SetMap(common.Uint64ToString(mapdata.Map_pxs_ID))

	tb.uniqueId = uniqueId
	tb.teamMgr = NewSpaceTeam(tb)

	tb.status = common.SpaceStatusInit

	tb.mapitem = make(map[uint64]*SceneItem)
	tb.mapindex = 0
	tb.cars = make(map[uint64]*Car)

	tb.summonai = 0
	tb.skybox = 0
	tb.aiNameMap = make(map[string]bool)
	tb.createStamp = time.Now().Unix()
	tb.members = make(map[uint64]bool)
	tb.closetime = 0
	// tb.entityTempID = tb.GetID() << 32
	tb.entityTempID = 0

	tb.parachuteTime = 0

	minLoad, ok := excel.GetSystem(common.System_MinLoadingTime)
	if ok {
		tb.minLoadTime = time.Duration(minLoad.Value)
	}

	maxLoad, ok := excel.GetSystem(common.System_MaxLoadingTime)
	if ok {
		tb.maxLoadTime = time.Duration(maxLoad.Value)
	}
	if tb.maxLoadTime == 0 {
		tb.maxLoadTime = 30
	}

	GetMapItemRate(tb)
	GetRefreshItemMgr(tb)
	GetRefreshZoneMgr(tb)
	//GetSpaceResultMgr(tb)

	tb.doorMgr = NewSpaceDoorMgr(tb)
	tb.aiPosALib = make([]linmath.Vector3, 0)
	tb.aiPosBLib = make([]linmath.Vector3, 0)
	tb.aiInRoomASum = 0
	tb.aiInRoomBSum = 0
	tb.GameID = db.FetchGameID()
	tb.loadFulPlayerNum = 0
	tb.averageRating = 0
	tb.scores = make(map[uint64]float64)

	//tb.RegTimerByObj(tb, tb.teamMgr.SyncAllTeamInfo, 1*time.Second)

	// 生成初始安全区
	GetRefreshZoneMgr(tb).genSafeZonePoint()
	// 生成航线
	tb.genAirLine()
	// 初始化ai名字库
	tb.aiNameUsedIndex = 0
	tb.aiNameLib = make([]int, 0)

	tb.randomAiPos = make(map[uint64]bool)
	tb.randomNum = 0

	tb.followParachuteOffsetX = float32(common.GetTBSystemValue(common.System_FollowParachuteOffsetX))
	tb.followParachuteOffsetY = float32(common.GetTBSystemValue(common.System_FollowParachuteOffsetY))
	tb.followParachuteOffsetZ = float32(common.GetTBSystemValue(common.System_FollowParachuteOffsetZ))

	tb.callBoxItem = make(map[uint64]CallBoxItem)
	tb.secondTicker = time.NewTicker(1 * time.Second)

	tb.watchers = make(map[uint64]bool)
	tb.Info("Scene inited, map name: ", tb.mapdata.Name, " scene num: ", GetSrvInst().EntityCount())
}

//  initAiNameLib 初始化ai名字库
func (tb *Scene) initAiNameLib(ainum uint32) {
	tb.aiNameLib = tb.r.Perm(excel.GetNameMapLen())

	if (ainum + 1) < uint32(len(tb.aiNameLib)) {
		tb.aiNameLib = append(tb.aiNameLib[0 : ainum+1])
	}

	for i, j := range tb.aiNameLib {
		if j == 0 {
			tb.aiNameLib = append(tb.aiNameLib[:i], tb.aiNameLib[(i+1):]...)
			break
		}
	}

	//tb.Info("初始化ai名字库", len(tb.aiNameLib), tb.aiNameUsedIndex, excel.GetNameMapLen(), ainum)
}

// getAiRandName 获取ai随机初始化名字
func (tb *Scene) getAiRandName() string {
	if tb.aiNameUsedIndex >= (uint32(len(tb.aiNameLib))) {
		tb.Error("getAiRandName failed, index is invalid, index: ", tb.aiNameUsedIndex, " len: ", len(tb.aiNameLib))
		return ""
	}
	index := tb.aiNameLib[tb.aiNameUsedIndex]
	nameData, success := excel.GetName(uint64(index))
	if success == false {
		tb.Error("getAiRandName failed, index not exist, index: ", tb.aiNameUsedIndex)
		return ""
	}

	tb.aiNameUsedIndex++
	//tb.Info("getAiRandName : ", nameData.Name)
	return nameData.Name
}

// initAiPosLib 初始化ai位置库
func (tb *Scene) initAiPosLib() {
	mapranges := tb.GetRanges()
	if mapranges == nil {
		return
	}

	// 获取A类型刷新点
	Atyperanges, Aerr := mapranges.GetRangeList(7)
	if Aerr != nil {
		tb.Error("initAiPosLib failed, GetRangeList err: ", Aerr)
		return
	}

	for _, Atyperange := range Atyperanges {
		droppos := Atyperange.CenterPos
		tb.aiPosALib = append(tb.aiPosALib, droppos)
	}

	// 获取B类型刷新点
	Btyperanges, Berr := mapranges.GetRangeList(5)
	if Berr != nil {
		tb.Error("initAiPosLib failed, GetRangeList err: ", Berr)
		return
	}

	for _, Btyperange := range Btyperanges {
		droppos := Btyperange.CenterPos
		tb.aiPosBLib = append(tb.aiPosBLib, droppos)
	}

	// tb.Info("getaiposAlib ", len(tb.aiPosALib))
	// tb.Info("getaiposBlib ", len(tb.aiPosBLib))
}

// getAiPos 获取ai坐标
func (tb *Scene) getAiPos(userPos linmath.Vector3) (linmath.Vector3, error) {
	sysAValue := common.GetTBSystemValue(51)
	// sysBValue := common.GetTBSystemValue(52)

	// 刷新到A类点
	if tb.aiInRoomASum < uint32(sysAValue) && len(tb.aiPosALib) != 0 {
		for i := 0; i < len(tb.aiPosALib); i++ {
			if ok := tb.randomRightAiPos(userPos, tb.aiPosALib[i]); ok {
				tb.aiInRoomASum++
				pos := tb.aiPosALib[i]

				tb.aiPosALib = append(tb.aiPosALib[:i], tb.aiPosALib[i+1:]...)
				return pos, nil
			}
		}
	}
	// // 刷新到B类点
	// if tb.aiInRoomBSum < uint32(sysBValue) && len(tb.aiPosBLib) != 0 {
	// 	for i := 0; i < len(tb.aiPosBLib); i++ {
	// 		if ok := tb.randomRightAiPos(userPos, tb.aiPosBLib[i]); ok {
	// 			tb.aiInRoomBSum++
	// 			pos := tb.aiPosBLib[i]

	// 			tb.aiPosBLib = append(tb.aiPosBLib[:i], tb.aiPosBLib[i+1:]...)
	// 			return pos, nil
	// 		}
	// 	}
	// }

	// 刷新到A类点
	if tb.aiInRoomASum < uint32(sysAValue) && len(tb.aiPosALib) != 0 {

		tb.aiInRoomASum++
		index := tb.r.Intn(len(tb.aiPosALib))
		pos := tb.aiPosALib[index]
		tb.aiPosALib = append(tb.aiPosALib[:index], tb.aiPosALib[index+1:]...)

		// tb.Info("刷新到A类点 ", pos)
		return pos, nil
	}

	// if tb.aiInRoomBSum < uint32(sysBValue) && len(tb.aiPosBLib) != 0 {

	// 	tb.aiInRoomBSum++
	// 	index := tb.r.Intn(len(tb.aiPosBLib))
	// 	pos := tb.aiPosBLib[index]
	// 	tb.aiPosBLib = append(tb.aiPosBLib[:index], tb.aiPosBLib[index+1:]...)

	// 	// tb.Info("刷新到B类点 ", pos)
	// 	return pos, nil
	// }

	// tb.Info("刷新到房间外 ")
	return linmath.Vector3{0, 0, 0}, errors.New("ai pos lib is nil")

}

// Destroy 删除时调用
func (tb *Scene) Destroy() {
	// tb.mapitem = nil

	// tb.teamMgr.DelRedisVoiceInfo()

	// tb.teamMgr.DelTeamInfo()
	db.SceneTempUtil(tb.GetID()).Remove()

	tb.UnregTimerByObj(tb)
	tb.Info("Scene destroyed, left scene num: ", GetSrvInst().EntityCount())
}

func (tb *Scene) destroyCallback(i uint8, e error) {
	tb.Info("destroyCallback, i: ", i)
}

//OnMapLoadSucceed 地图加载成功, 框架层回调
func (tb *Scene) OnMapLoadSucceed() {
	if tb.minLoadTime != 0 {
		tb.AddDelayCall(tb.checkAllowParachuteState, tb.minLoadTime*time.Second)
	}

	tb.maploadSuccess = true

	tb.initAiPosLib()
	//tb.Info("Init scene item start")
	//start := time.Now()
	GetRefreshItemMgr(tb).InitSceneItem()
	//tb.Info("Init scene item done, cost ", time.Now().Sub(start).String())

	// 通知Match, Space创建完成, 新的匹配流程
	if err := tb.RPC(common.ServerTypeMatch, "RoomSpaceInited"); err != nil {
		tb.Error("RPC RoomSpaceInited err: ", err)
	}

	// 场景创建完成15秒之后强制开启跳伞
	if tb.maxLoadTime != 0 {
		tb.AddDelayCall(tb.startParachute, tb.maxLoadTime*time.Second)
	}

	tb.Info("Map load success, mapid: ", tb.mapdata.Id)
}

//OnMapLoadFailed 地图加载失败
func (tb *Scene) OnMapLoadFailed() {
	tb.Error("Map load failed")
}

//Loop 定时执行
func (tb *Scene) Loop() {
	now := time.Now().Unix()

	if now-tb.createStamp > 2100 {
		tb.status = common.SpaceStatusClose
		tb.BroadBattleOver(0)
		return
	}

	tb.SubAirAI()
	var systemID uint64
	if tb.GetMatchMode() == common.MatchModeArcade {
		systemID = uint64(common.System_ArcadeRefreshSafeArea)
	} else if tb.GetMatchMode() == common.MatchModeTankWar {
		systemID = uint64(common.System_TankWarRefreshSafeArea)
	} else {
		systemID = uint64(common.System_RefreshSafeArea)
	}

	if tb.GetMatchMode() == common.MatchModeNormal && tb.mapdata.Id == 2 {
		systemID = 4005
	} else if tb.GetMatchMode() == common.MatchModeEndless && tb.mapdata.Id == 7 {
		systemID = 4011
	}

	system, ok := excel.GetSystem(systemID)
	if !ok {
		return
	}
	initzonetime := int64(system.Value)
	if tb.status == common.SpaceStatusInit {
		if tb.parachuteTime != 0 && now >= tb.parachuteTime+15 {
			tb.doResult()
		}
		if tb.parachuteTime != 0 && now >= tb.parachuteTime+initzonetime {
			tb.status = common.SpaceStatusBegin

			GetRefreshItemMgr(tb).SetRefreshBox(now)
			GetRefreshZoneMgr(tb).InitZone(now)
		}
	} else if tb.status == common.SpaceStatusBegin {
		tb.doResult()
		GetRefreshItemMgr(tb).RefreshBox(now)
		GetRefreshItemMgr(tb).dropBox()
		GetRefreshItemMgr(tb).RefreshHostage()
		GetRefreshZoneMgr(tb).Update()
	} else if tb.status == common.SpaceStatusBalanceDone {
		if time.Now().Unix() > tb.closetime {
			tb.status = common.SpaceStatusClose
		}
	}

	tb.ISceneData.doLoop()

	tb.teamMgr.TeamFollowParachuteCtrl()

	select {
	case <-tb.secondTicker.C:
		GetRefreshItemMgr(tb).dropCallBoxItem()
	default:
	}
}

// 获取本房间的匹配模式
func (tb *Scene) GetMatchMode() uint32 {
	info, ok := excel.GetMatchmode(uint64(tb.uniqueId))
	if !ok {
		return 0
	}

	return uint32(info.Modeid)
}

func (tb *Scene) isEliteScene() bool {
	return tb.GetMatchMode() == common.MatchModeElite
}

func (tb *Scene) isVersusScene() bool {
	return tb.GetMatchMode() == common.MatchModeVersus
}

func (tb *Scene) isEnd() bool {
	return tb.status == common.SpaceStatusBalanceDone || tb.status == common.SpaceStatusClose
}

// SummonAI 添加AI
func (tb *Scene) SummonAI(ainum uint32) {
	if tb.summonai != 0 {
		return
	}
	tb.summonai = time.Now().Unix()

	tb.initAiNameLib(ainum)

	for i := 0; i < int(ainum); i++ {
		entityID := tb.GetEntityTempID()
		tb.randomAiPos[entityID] = true

		tb.AddEntity("AI", entityID, 0, "", false, true)
	}

	//tb.Error("summon ai alive num: ", len(tb.members))
	tb.BroadAliveNum()
	tb.BroadAirLeft()
}

// GetSpaceType 获取类型
func (tb *Scene) GetSpaceType() uint64 {
	if tb.teamMgr.isTeam {
		return TeamSpace
	}

	return NotTeamSpace
}

func (tb *Scene) getMemSum() int {
	return len(tb.members)
}

// GetCircleRamdomPos 获取随机点
func (tb *Scene) GetCircleRamdomPos(center linmath.Vector3, radius float32) linmath.Vector3 {
	ret := linmath.RandXZ(center, radius)
	waterLevel := tb.mapdata.Water_height
	err := errors.New("GetCircleRamdomPos fail")
	for i := 0; i < 24; {
		ret.Y, err = tb.GetHeight(ret.X, ret.Z)
		if err == nil && ret.Y > waterLevel {
			break
		} else {
			ret.Y = waterLevel
		}

		ret = linmath.RandXZ(center, radius)
		i++
	}

	return ret
}

func (tb *Scene) GetCircleRamdomIndexPos(index uint64, center linmath.Vector3, radius float32) linmath.Vector3 {
	ret := linmath.RandXZ(center, radius)
	waterLevel := tb.mapdata.Water_height
	err := errors.New("GetCircleRamdomPos fail")

	angle := 2 * math.Pi / float64(index)
	for j := 1; j <= 3; j++ {
		tarR := float64(radius) * float64(j) / 3.0

		pos := linmath.Vector3_Zero()
		pos.X = float32(math.Cos(angle) * tarR)
		pos.Z = float32(math.Sin(angle) * tarR)
		ret = center.Add(pos)
		ret.Y, err = tb.GetHeight(ret.X, ret.Z)
		if err == nil && ret.Y > waterLevel {
			//tb.Debug("获取随机点", i, j, ret.Y)
			//seelog.Debug("获取随机点", ret, i, j)
			return ret
		}
	}

	for i := 0; i < 2; {
		ret.Y, err = tb.GetHeight(ret.X, ret.Z)
		if err == nil && ret.Y > waterLevel {
			//seelog.Error("刷新随机点:", i, ret)
			return ret
		}

		ret.Y = waterLevel
		ret = linmath.RandXZ(center, radius)
		i++
	}

	//seelog.Error("@@@@@@@@@@", ret, " center:", center, radius)
	return ret
}

// GetEntityTempID 获取场景内临时EntityID
func (tb *Scene) GetEntityTempID() uint64 {
	tb.entityTempID++
	return tb.entityTempID
}

// IsInOneTeam 是否在同一队伍
func (tb *Scene) IsInOneTeam(playerid1, playerid2 uint64) bool {
	if !tb.teamMgr.isTeam {
		return false
	}
	if playerid1 == playerid2 {
		return false
	}

	player1, ok := tb.GetEntity(playerid1).(*RoomUser)
	if !ok {
		return false
	}

	player2, ok := tb.GetEntity(playerid2).(*RoomUser)
	if !ok {
		return false
	}

	playerteamid1 := player1.GetUserTeamID()
	playerteamid2 := player2.GetUserTeamID()

	if playerteamid1 == 0 || playerteamid2 == 0 {
		return false
	}

	return playerteamid1 == playerteamid2
}

// getWatchTarget 获取观战目标
func (tb *Scene) getWatchTarget(user IRoomChracter, bSwitch bool) uint64 {
	if !user.IsDead() {
		return 0
	}

	target := user.GetWatchingTarget()
	targetUser, _ := tb.GetEntity(target).(IRoomChracter)

	if targetUser != nil && !targetUser.IsDead() && !bSwitch {
		tb.Debug("getWatchTarget, old target: ", target)
		return target
	}

	target = tb.getWatchTargetInCamp(user)
	if target != 0 {
		tb.Debug("getWatchTarget, camp target: ", target)
		return target
	}

	target = user.GetKiller()
	if target != 0 {
		targetUser, _ := tb.GetEntity(target).(IRoomChracter)
		if targetUser != nil && !targetUser.IsDead() {
			tb.Debug("getWatchTarget, killer: ", target)
			return target
		}
	}

	target = tb.getMaxKiller(user.GetRivalGamerType())
	if target != 0 {
		tb.Debug("getWatchTarget, max Killer: ", target)
		return target
	}

	target = tb.getRandomCharacter(user.GetRivalGamerType())
	if target != 0 {
		tb.Debug("getWatchTarget, random target: ", target)
		return target
	}

	return 0
}

// getWatchTargetInCamp 在同一阵营内为玩家选取观战目标
func (tb *Scene) getWatchTargetInCamp(user IRoomChracter) uint64 {

	switch tb.GetMatchMode() {
	//精英模式，红蓝对决
	case common.MatchModeElite, common.MatchModeVersus:
		target := tb.getNearestCharacter(user)
		if target != 0 {
			return target
		}

		target = tb.getMaxKiller(user.GetGamerType())
		if target != 0 {
			return target
		}

		target = tb.getRandomCharacter(user.GetGamerType())
		if target != 0 {
			return target
		}

	//普通模式等
	default:
		return user.GetWatchTargetInTeam()
	}

	return 0
}

// getNearestCharacter 获取剩余角色中距离最近者
func (tb *Scene) getNearestCharacter(user IRoomChracter) uint64 {
	var (
		min  float32
		role uint64
	)

	cood, _ := tb.GetEntity(user.GetID()).(iserver.ICoordEntity)
	if cood == nil {
		return 0
	}

	tb.TravsalAOI(cood, func(o iserver.ICoordEntity) {
		targetUser, _ := tb.GetEntity(o.GetID()).(IRoomChracter)
		if targetUser == nil || targetUser.IsDead() || targetUser.isAI() || targetUser.GetGamerType() != user.GetGamerType() {
			return
		}

		dis := common.Distance(cood.GetPos(), o.GetPos())
		if min == 0 || dis < min {
			min = dis
			role = o.GetID()
		}
	})

	if role != 0 {
		return role
	}

	tb.TravsalAOI(cood, func(o iserver.ICoordEntity) {
		if o.GetPos().IsEqual(linmath.Vector3{0, 0, 0}) {
			return
		}

		targetUser, _ := tb.GetEntity(o.GetID()).(IRoomChracter)
		if targetUser == nil || targetUser.IsDead() || targetUser.GetGamerType() != user.GetGamerType() {
			return
		}

		dis := common.Distance(cood.GetPos(), o.GetPos())
		if min == 0 || dis < min {
			min = dis
			role = o.GetID()
		}
	})

	return role
}

// getMaxKiller 获取剩余角色中击杀人数最多者
func (tb *Scene) getMaxKiller(gamerType uint32) uint64 {
	var (
		max    uint32
		killer uint64
	)

	for id := range tb.members {
		user, _ := tb.GetEntity(id).(IRoomChracter)
		if user == nil || user.IsDead() || user.isAI() || user.GetGamerType() != gamerType {
			continue
		}

		if user.GetKillNum() > max {
			max = user.GetKillNum()
			killer = user.GetID()
		}
	}

	if max > 0 {
		return killer
	}

	for id := range tb.members {
		user, _ := tb.GetEntity(id).(IRoomChracter)
		if user == nil || user.GetPos().IsEqual(linmath.Vector3{0, 0, 0}) || user.IsDead() || user.GetGamerType() != gamerType {
			continue
		}

		if user.GetKillNum() > max {
			max = user.GetKillNum()
			killer = user.GetID()
		}
	}

	return killer
}

// getRandomCharacter 从剩余角色中随机选出一个，真实玩家优先
func (tb *Scene) getRandomCharacter(gamerType uint32) uint64 {

	for id := range tb.members {
		user, _ := tb.GetEntity(id).(IRoomChracter)
		if user == nil || user.IsDead() || user.isAI() || user.GetGamerType() != gamerType {
			continue
		}

		return id
	}

	for id := range tb.members {
		user, _ := tb.GetEntity(id).(IRoomChracter)
		if user == nil || user.GetPos().IsEqual(linmath.Vector3{0, 0, 0}) || user.IsDead() || user.GetGamerType() != gamerType {
			continue
		}

		return id
	}

	return 0
}

// LoadMemberRating 玩家进去后，载入玩家在本模式的评分
func (tb *Scene) LoadMemberRating() {
	var key string
	if !tb.teamMgr.isTeam {
		key = "SoloRating"
	} else {
		if tb.teamMgr.teamType == 0 {
			key = "DuoRating"
		}
		if tb.teamMgr.teamType == 1 {
			key = "SquadRating"
		}
	}

	var totalRating, i float64 = 0, 0
	tb.TravsalEntity("Player", func(e iserver.IEntity) {
		user := e.(*RoomUser)
		dbUtil := db.PlayerCareerDataUtil(common.PlayerCareerTotalData, user.GetDBID(), common.GetSeason())
		rating, err := redis.Float64(dbUtil.GetValue(key))
		if err != nil && err != redis.ErrNil {
			return
		}

		ratingMap := excel.GetRatingMap()
		ratingFloor, ratingUpper := ratingMap[4027].Value, ratingMap[4028].Value
		if rating < float64(ratingFloor) || rating > float64(ratingUpper) || math.IsNaN(rating) {
			ratingData, ok := excel.GetRating(4001)
			if ok {
				rating = float64(ratingData.Value)
			}
		}

		tb.scores[user.GetDBID()] = rating
		totalRating += rating
		i++
	})
	if i != 0 {
		tb.averageRating = totalRating / i
	}
	tb.memberModeScore = sortMapByValue(tb.scores)
}

type MemberModeScore struct {
	uid   uint64
	score float64
}
type PairList []*MemberModeScore

func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].score < p[j].score }

func sortMapByValue(m map[uint64]float64) PairList {
	p := make(PairList, len(m))
	i := 0
	for k, v := range m {
		p[i] = &MemberModeScore{k, v}
		i++
	}
	sort.Sort(p)
	return p
}

func (tb *Scene) GetMemberScoreRoundMe(uid uint64, num int) []uint64 {
	var list []uint64
	index := -1
	var myScore float64
	for k, v := range tb.memberModeScore {
		if v.uid == uid {
			index = k
			myScore = v.score
			break
		}
	}
	if index == -1 {
		return list
	}
	l := len(tb.memberModeScore)
	s := index - num
	e := index + num
	if s <= 0 {
		s = 0
	}
	if e >= l-1 {
		e = l - 1
	}
	tmp := map[uint64]float64{}
	for i := s; i <= e; i++ {
		if i == index {
			continue
		}
		tmp[tb.memberModeScore[i].uid] = math.Abs(float64(tb.memberModeScore[i].score) - float64(myScore))
	}
	if len(tmp) == 0 {
		return list
	}
	tmpSort := sortMapByValue(tmp)
	n := len(tmpSort)
	if n > num {
		n = num
	}

	for i := 0; i < n; i++ {
		list = append(list, tmpSort[i].uid)
	}

	return list
}

// GetSpaceTeamType 获取当前游戏的匹配类型 1 单排 2 双排 4 四排
func (tb *Scene) GetSpaceTeamType() uint8 {
	if !tb.teamMgr.isTeam {
		return 1
	} else {
		if tb.teamMgr.teamType == 0 {
			return 2
		}
		if tb.teamMgr.teamType == 1 {
			return 4
		}
	}
	return 0
}

// randomRightAiPos 判断随机ai位置是否符合配置要求
func (tb *Scene) randomRightAiPos(userPos, aiPos linmath.Vector3) bool {
	floorDis, ok1 := excel.GetAisys(common.AiSys_RandomAiPosFloor)
	if !ok1 {
		return false
	}
	ceilDis, ok2 := excel.GetAisys(common.AiSys_RandomAiPosCeil)
	if !ok2 {
		return false
	}

	curDis := uint64(common.Distance(userPos, aiPos))
	if curDis > floorDis.Valueuint && curDis < ceilDis.Valueuint {
		log.Debug("randomRightAiPos curDis:", curDis, " userPos:", userPos, " aiPos:", aiPos)
		return true
	}

	return false
}

// canAttackSpecial 用来处理canAttack验证的特殊情况，适用于各种匹配模式
func (tb *Scene) canAttackSpecial(user IRoomChracter) bool {
	if user.isInTank() || user.isBazookaWeaponUse() {
		return true
	}

	return false
}

func (tb *Scene) canAttack(user IRoomChracter, defendid uint64) bool {
	if tb.canAttackSpecial(user) {
		return true
	}

	return tb.ISceneData.canAttack(user, defendid)
}

func (tb *Scene) refreshSpecialBox(refreshcount uint64) {
	tb.ISceneData.refreshSpecialBox(refreshcount)
}

func (tb *Scene) resendData(user *RoomUser) {
	tb.ISceneData.resendData(user)
}

func (tb *Scene) updateTotalNum(user *RoomUser) {
	tb.ISceneData.updateTotalNum(user)
}

func (tb *Scene) updateAliveNum(user *RoomUser) {
	tb.ISceneData.updateAliveNum(user)
}

//DoResult 结算
func (tb *Scene) DoResult() {
	now := time.Now().Unix()
	var end bool

	teamid := uint64(0)
	if tb.teamMgr.isTeam {
		membernum := tb.teamMgr.getTeamPlayerSum()
		if len(tb.members) <= int(membernum) {

			if tb.teamMgr.SurplusTeamSum() > 1 {
				return
			}

			end = true
			//var teamid uint64
			tb.TravsalEntity("Player", func(e iserver.IEntity) {
				if e == nil {
					return
				}

				if user, ok := e.(*RoomUser); ok {
					if user.stateMgr.GetState() == RoomPlayerBaseState_Dead || user.stateMgr.GetState() == RoomPlayerBaseState_Watch {
						return
					}
					if teamid == 0 {
						teamid = user.GetUserTeamID()
					} else {
						if teamid != user.GetUserTeamID() {
							end = false
						}
					}
				}
			})
		}
	} else {
		if len(tb.members) <= 1 {
			end = true
		}
	}

	// if len(sf.users) == 0 && now >= tb.createStamp+10 {
	// 	end = true
	// }

	if end && tb.maploadSuccess && tb.allowParachute && now >= tb.summonai+10 {

		tb.closetime = now + int64(common.GetTBSystemValue(common.System_LeaveSceneTime)) + 10 + 10
		tb.status = common.SpaceStatusBalanceDone

		var winner uint64

		if tb.teamMgr.isTeam {
			tb.teamMgr.DisposeTeamSettle(teamid, true)
			winner = teamid
		} else {
			for id := range tb.members {
				user, _ := tb.GetEntity(id).(IRoomChracter)
				if user == nil || user.IsDead() {
					continue
				}

				user.SetVictory()

				if !user.isAI() {
					realUser, _ := tb.GetEntity(id).(*RoomUser)
					if realUser != nil {
						realUser.DisposeSettle()
						winner = id
					}
				}
			}
		}

		tb.BroadBattleOver(winner)
		tb.Info("Do result success, winner: ", winner)

	}
}

// UpdateWatchNum 更新某个玩家的观战人数，并通知
func (tb *Scene) UpdateWatchNum(target uint64, change int, notify bool) {
	targetUser, ok := tb.GetEntity(target).(*RoomUser)
	if !ok {
		return
	}

	if tb.teamMgr.isTeam {
		db.PlayerTeamUtil(targetUser.GetUserTeamID()).IncryTeamWatchNum(change)
	} else {
		db.SceneTempUtil(tb.GetID()).IncrbySingleUserWatchNum(target, change)
	}

	if notify {
		tb.NotifyWatchNum(target)
	}
}

// getWatchNum 获取某个玩家的观战人数 ，如果是组队中 则是队伍的观战人数
func (tb *Scene) getWatchNum(uid uint64) uint32 {
	var num uint32

	if tb.teamMgr.isTeam {
		teamID := tb.teamMgr.GetTeamIDByPlayerID(uid)
		num = db.PlayerTeamUtil(teamID).GetTeamWatchNum()
	} else {
		num = db.SceneTempUtil(tb.GetID()).GetSingleUserWatchNum(uid)
	}

	return num
}

// NotifyWatchNum 通知观战人数
func (tb *Scene) NotifyWatchNum(target uint64) {
	targetUser, ok := tb.GetEntity(target).(*RoomUser)
	if !ok {
		return
	}

	num := tb.getWatchNum(target)
	for _, v := range targetUser.GetTeamMembers() {
		mem, ok := tb.GetEntity(v).(*RoomUser)
		if !ok {
			continue
		}

		mem.RPC(iserver.ServerTypeClient, "UpdateWatchNum", num)
		mem.BroadcastToWatchers(RoomUserTypeWatcher, "UpdateWatchNum", num)
	}
}

package main

import (
	"common"
	"datadef"
	"db"
	"entitydef"
	"fmt"
	"ipip"
	"math"
	"msdk"
	"protoMsg"
	"strconv"
	"strings"
	"time"
	"zeus/dbservice"
	"zeus/entity"
	"zeus/iserver"
	"zeus/msgdef"
	"zeus/tlog"
	"zeus/tsssdk"

	"github.com/garyburd/redigo/redis"
	"github.com/robfig/cron"

	"excel"

	"github.com/spf13/viper"
)

// LobbyUser 大厅玩家
type LobbyUser struct {
	entitydef.PlayerDef
	entity.Entity
	spaceID   uint64 //房间唯一编号
	macthTime int64  //unix时间戳
	//teamID  uint64 // 组队id

	gameState uint64

	// 压测数据相关
	enterRoomStamp time.Time

	friendMgr      *FriendMgr
	gm             *GmMgr
	storeMgr       *StoreMgr
	activityMgr    *ActivityMgr
	treasureBoxMgr *TreasureBoxMgr
	taskMgr        *TaskMgr

	loginMsg  *protoMsg.PlayerLogin
	loginTime int64
	isReg     bool

	matchMode uint32 //匹配模式，具体定义在common/Const.go
	isQemu    uint32 //是否是模拟器:0不是模拟器， 1是模拟器

	matchMgrSoloProxy iserver.IEntityProxy
	teamMgrProxy      iserver.IEntityProxy

	teamInfo *protoMsg.SyncTeamInfoRet //所在队伍信息

	lastTeamCustomTime  time.Time
	lastChatPushTime    time.Time
	lastPrivateChatTime time.Time

	onlineReportTicker *time.Ticker
	secondTicker       *time.Ticker
	lastBatchTime      int64
	token              string
	mapId              uint32
	skyBox             uint32
	version            string
	platID             uint32

	// 观战相关
	watchTarget     uint64
	watchStart      int64
	clearWatchTimer *time.Timer
	isInTeamReady   bool

	cronTask *cron.Cron //定时检查玩家的时限道具是否过期
}

// Init 初始化调用
func (user *LobbyUser) Init(initParam interface{}) {
	user.RegMsgProc(&LobbyUserMsgProc{user: user})

	user.friendMgr = NewFriendMgr(user)
	user.storeMgr = NewStoreMgr(user)
	user.gm = NewGmMgr(user)
	user.activityMgr = NewActivityMgr(user)
	user.treasureBoxMgr = NewTreasureBoxMgr(user)
	user.taskMgr = NewTaskMgr(user)

	user.matchMode = common.MatchModeNormal
	user.isQemu = 0
	user.createInit()

	user.loginTime = time.Now().Unix()
	user.SetLoginTime(user.loginTime)

	var err error
	user.isReg, err = db.PlayerInfoUtil(user.GetDBID()).SetRegisterTime(user.loginTime)
	if err != nil {
		user.Error("SetRegisterTime err: ", err)
	}

	tmpToken, errToken := dbservice.SessionUtil(user.GetDBID()).GetToken()
	if errToken != nil {
		user.Warn("GetToken err: ", errToken)
	} else {
		user.token = tmpToken
	}

	user.checkMatchOpen()
	user.notifyServerMode()
	user.notifyPlayerMode()
	user.checkQuitTeam()
	user.setLocation()
	user.checkModeLightFlag()
	user.onlineReportTicker = time.NewTicker(5 * time.Second)
	user.secondTicker = time.NewTicker(1 * time.Second)

	user.RPC(iserver.ServerTypeClient, "SrvTime", time.Now().UnixNano())

	user.Info("LobbyUser inited")
}

//通知客户端普通模式的开放信息
func (user *LobbyUser) checkMatchOpen() {
	single := viper.GetBool("Lobby.Solo")
	two := viper.GetBool("Lobby.Duo")
	four := viper.GetBool("Lobby.Squad")
	user.RPC(iserver.ServerTypeClient, "OnlineCheckMatchOpen", single, two, four)
}

//新玩家的数据初始化
func (user *LobbyUser) createInit() {
	gender, _ := redis.String(dbservice.EntityUtil("Player", user.GetDBID()).GetValue("Gender"))
	user.SetGender(gender)

	// 玩家的初始等级为1
	if user.GetLevel() == 0 {
		user.SetLevel(1)
	}

	// 初始化角色模型
	if user.GetGoodsRoleModel() == 0 || user.GetRoleModel() == 0 {
		user.SetRoleModel(uint32(common.GetTBSystemValue(32)))
		user.SetGoodsRoleModel(uint32(common.GetTBSystemValue(32)))
	}

	// 初始化伞包id
	if user.GetGoodsParachuteID() == 0 || user.GetParachuteID() == 0 {
		user.SetParachuteID(uint32(common.GetTBSystemValue(common.System_InitParachuteID)))
		user.SetGoodsParachuteID(uint32(common.GetTBSystemValue(common.System_InitParachuteID)))
	}

	//获取用于标识新手类型的数值
	util := db.PlayerInfoUtil(user.GetDBID())
	noviceType, _ := util.GetNoviceType()

	//在新手功能上线前开过游戏的老号，不再提示新手引导
	if noviceType == 0 {
		if db.PlayerCareerDataUtil(common.PlayerCareerTotalData, user.GetDBID(), 1).IsKeyExist() ||
			db.PlayerCareerDataUtil(common.PlayerCareerTotalData, user.GetDBID(), 2).IsKeyExist() {
			noviceType = 4
			util.SetNoviceType(4)
		}
	}

	user.RPC(iserver.ServerTypeClient, "NoviceDataNotify", noviceType)

	//保存新账号绑定战友的结束时间
	if len(user.GetName()) == 0 {
		days := common.GetTBSystemValue(common.System_TakeTeacherValidDays)
		deadTime := time.Now().Unix() + int64(days*24*60*60)
		util.SetTakeTeacherDeadTime(deadTime)
	}

	//在第一赛季的排行榜上，将玩家在功能上线前的历史数据插入排行榜
	user.InsertFirstSeasonOldRank()
	user.initEquipSkill() // 给每个未装备技能的角色装备初始技能

	user.SetPlayerGameState(common.StateFree)
	user.RPC(iserver.ServerTypeClient, "JumpAir", db.PlayerTempUtil(user.GetDBID()).GetPlayerJumpAir())

	// 同步好友列表和好友申请列表
	user.friendMgr.syncFriendList()
	user.friendMgr.syncApplyList()

	user.checkGlobalMail()
	checkMail(user.GetDBID())
	user.MailNotify()

	user.storeMgr.initPropInfo()    //初始化购买物品信息
	user.StartCrondForExpireCheck() //开启crond服务，定时检查玩家的时限道具是否到期

	user.InitAnnounceData()
	user.AnnounceDataCenterAddr() //通知客户端DataCenter服地址
	user.ChangeNameInfo()
	user.AchievementInit()
	user.InsigniaInfo(user.GetDBID())         //同步勋章信息
	user.CareerCurSeason()                    // 通知客户端赛季id
	user.treasureBoxMgr.syncTreasureBoxInfo() //同步宝箱开启状态

	user.ExpAdjust()

	user.dayTaskInfoNotify()
	user.comradeTaskInfoNotify()
	user.taskMgr.syncTaskDetailAll()

	user.updateDayTaskItems(common.TaskItem_Login, 1)
	user.taskMgr.updateTaskItemsAll(common.TaskItem_Login, 1)

	user.festivalDataFlush(common.Act_Login)
	user.checkVeteranRecallList() // 检测老兵召回好友列表

	user.syncWeaponEquipment()   // 同步武器装备信息
	user.initPreferenceInfo()    // 初始化偏好信息
	user.syncUnreadPrivateChat() // 同步私聊信息

	user.syncFirstPay() // 同步首充
	user.syncPayLevel() // 同步首次充送

	user.syncRedDotOnce() // 通知玩家一次行点击类红点
}

// Destroy 析构时调用
func (user *LobbyUser) Destroy() {
	if user.spaceID != 0 {
		user.clear()
	}

	user.delMatch()
	user.MemberQuitTeam()
	user.clearTmpBookingFriends()
	user.SetPayOS("")

	if user.cronTask != nil {
		user.cronTask.Stop()
	}

	user.tlogOnPlayerLogout()
	user.tlogLiveFlow()
	user.tlogGoodRecordFlow(logout)
	user.tsssdkLogout()

	util := db.PlayerTempUtil(user.GetDBID())

	// 检查是否被顶号
	if tmpToken, err := dbservice.SessionUtil(user.GetDBID()).GetToken(); err == nil {
		if tmpToken == user.token {
			user.SetPlayerGameState(common.StateOffline)
			GetSrvInst().DestroyEntityAll(user.GetID())
			util.DelKey()
		} else {
			if err := user.Post(iserver.ServerTypeClient, &msgdef.UserDuplicateLoginNotify{}); err != nil {
				user.Error("Post UserDuplicateLoginNotify err: ", err)
			}
		}
	} else if err == redis.ErrNil {
		user.SetPlayerGameState(common.StateOffline)
		GetSrvInst().DestroyEntityAll(user.GetID())
		util.DelKey()
	}

	user.onlineReportTicker.Stop()
	user.secondTicker.Stop()

	user.Info("Lobby user destroyed")
}

/*
	说明：
	赛季排行功能原始需求是显示(单排、双排、四排)*(获胜数、击杀数、rating积分)共九个榜单，
	为了追溯该功能上线前玩家的历史数据，第一赛季只显示总获胜数和总击杀数两个榜单。因此，
	该功能相关的部分代码只在第一赛季有效，后续赛季可以删除该部分代码。
*/
func (user *LobbyUser) InsertFirstSeasonOldRank() {
	rankSeason := common.GetRankSeason()
	if rankSeason != 1 {
		return
	}

	season := common.RankSeasonToSeason(rankSeason)
	data, _ := common.GetPlayerCareerData(user.GetDBID(), season, 0)
	if data == nil || data.TotalBattleNum < 3 {
		return
	}

	rankTypStrs := map[uint8]string{
		1: common.SeasonRankTypStrWins,
		2: common.SeasonRankTypStrKills,
	}

	rankTypData := map[uint8]uint32{
		1: data.FirstNum,
		2: data.TotalKillNum,
	}

	for rankTyp := range rankTypStrs {
		util := db.PlayerRankUtil(common.TotalRank+rankTypStrs[rankTyp], season)
		rank, _ := util.GetPlayerRank(user.GetDBID())
		if rank != 0 {
			continue
		}

		f := 1 - float64(time.Now().UnixNano()/1000%10000000000000)/float64(10000000000000)

		if _, err := util.PipeRanksUint64(float64(rankTypData[rankTyp])+f, user.GetDBID()); err != nil {
			user.Error("PipeRanks err: ", err)
		}
	}

	user.Info("InsertFirstSeasonOldRank, FirstNum: ", data.FirstNum, " TotalKillNum: ", data.TotalKillNum)
}

//一局游戏结束 清理上局的数据
func (user *LobbyUser) clear() {
	user.spaceID = 0
}

//setOnline 设置在线状态
func (user *LobbyUser) setOnline(s uint64) {
}

// getLocation 获取玩家所在地理位置
func (user *LobbyUser) getLocation() string {
	return db.PlayerTempUtil(user.GetDBID()).GetPlayerLocation()
}

// setLocation 设置玩家所在地理位置
func (user *LobbyUser) setLocation() {
	//获取客户端的ip地址
	ip, err := dbservice.SessionUtil(user.GetDBID()).GetIP()
	if err != nil {
		user.Error("GetIP err: ", err)
		return
	}

	strs := strings.Split(ip, ":")
	ip = strs[0]

	//根据ip地址查询地理位置
	loc, err := ipip.Find(ip)
	if err != nil {
		fmt.Println("Find err: ", err)
		return
	}

	var location string

	//选择能查询到的最精确位置作为玩家客户端显示的地理位置
	if loc.City != "未知" && loc.City != "局域网" {
		location = loc.City
	} else if loc.Province != "未知" && loc.Province != "局域网" {
		location = loc.Province
	} else if loc.Country != "未知" && loc.Country != "局域网" {
		location = loc.Country
	}

	db.PlayerTempUtil(user.GetDBID()).SetPlayerLocation(location)
}

//开启crond服务，定时检查玩家的时限道具是否过期
func (user *LobbyUser) StartCrondForExpireCheck() {
	//先强制检查一次
	user.perMinuteCheck()

	if user.cronTask == nil {
		user.cronTask = cron.New()
		if err := user.cronTask.AddFunc("0 0/1 * * * ?", user.perMinuteCheck); err != nil {
			user.Error("AddFunc err: ", err)
			return
		}
	}

	user.cronTask.Start()
}

// perMinuteCheck 每分钟检测一次 (玩家在游戏中不检测)
func (user *LobbyUser) perMinuteCheck() {
	if user.GetPlayerGameState() != common.StateGame {
		user.storeMgr.itemRegularCheck()
		user.monthCardRegularCheck()
		user.payLevelRegularCheck()
	}
}

//更新玩家的军衔等级
func (user *LobbyUser) UpdateMilitaryRank(incr uint32) {
	if incr == 0 {
		return
	}

	user.ExpAdjust()

	level, exp := common.CalMilitaryRank(user.GetLevel(), user.GetExp(), incr)

	//领取军衔升级奖励
	for i := user.GetLevel() + 1; i <= level; i++ {
		awards := common.GetAwardsByLevel(i)
		user.storeMgr.GetAwards(awards, common.RS_MilitaryRank, false, false)
	}

	user.SetLevel(level)
	user.SetExp(exp)

	max := common.GetMaxExpByLevel(user.GetLevel() + 1)
	r := common.NoRoundFloat(float64(user.GetExp())/float64(max), 6)
	db.PlayerInfoUtil(user.GetDBID()).SetExpRatio(r)

	user.MilitaryRankFlow()
	user.Info("UpdateMilitaryRank, level: ", level, " exp: ", exp, " incr: ", incr)
}

//检测军衔经验值相关的配置是否发生变化，若发生变化，则根据数据库记录的经验值比例调整玩家当前的经验值
func (user *LobbyUser) ExpAdjust() {
	oldr := db.PlayerInfoUtil(user.GetDBID()).GetExpRatio()
	max := common.GetMaxExpByLevel(user.GetLevel() + 1)
	newr := common.NoRoundFloat(float64(user.GetExp())/float64(max), 6)

	if math.Abs(newr-oldr) < 0.000001 {
		return
	}

	user.SetExp(uint32(oldr * float64(max)))
}

// GetTeamID 获取玩家组队id
func (user *LobbyUser) GetTeamID() uint64 {
	return db.PlayerTempUtil(user.GetDBID()).GetPlayerTeamID()
}

// SetTeamID 设置玩家组队id
func (user *LobbyUser) SetTeamID(id uint64) {
	db.PlayerTempUtil(user.GetDBID()).SetPlayerTeamID(id)
}

// GetMatchTyp 获取玩家的匹配类型 1单排 2双排 4四排
func (user *LobbyUser) GetMatchTyp() uint8 {
	teamID := user.GetTeamID()
	if teamID == 0 {
		return 1
	}

	teamTyp, err := db.PlayerTeamUtil(teamID).GetTeamType()
	if err != nil {
		return 1
	}

	if teamTyp == 0 {
		return 2
	} else if teamTyp == 1 {
		return 4
	}

	return 1
}

// AdviceNotify 提示信息通知
func (user *LobbyUser) AdviceNotify(notifyType uint32, id uint64) {
	if err := user.RPC(iserver.ServerTypeClient, "AdviceNotify", notifyType, id); err != nil {
		user.Error("RPC AdviceNotify err: ", err)
	}
}

// MemberQuitTeam 离开队伍
func (user *LobbyUser) MemberQuitTeam() {
	if user.GetTeamID() == 0 {
		return
	}

	if user.teamMgrProxy == nil {
		return
	}

	if err := user.teamMgrProxy.RPC(common.ServerTypeMatch, "LeaveTeam",
		GetSrvInst().GetSrvID(), user.GetID(), user.GetTeamID(), false); err != nil {
		user.Error("RPC LeaveTeam err: ", err)
	}
	user.SetTeamID(0)
	user.teamInfo = nil
	user.SetPlayerGameState(common.StateFree)
}

// checkQuitTeam 上线检查队伍
func (user *LobbyUser) checkQuitTeam() {
	if user.GetTeamID() == 0 {
		return
	}

	srvID, err := db.PlayerTeamUtil(user.GetTeamID()).GetMatchSrvID()
	if err == nil {
		name := fmt.Sprintf("%s:%d", common.TeamMgr, srvID)
		proxy := entity.GetEntityProxy(name)
		if proxy != nil {
			if err := proxy.RPC(common.ServerTypeMatch, "checkQuitTeam",
				GetSrvInst().GetSrvID(), user.GetDBID(), user.GetTeamID()); err != nil {
				user.Error("RPC checkQuitTeam err: ", err)
			}
		}
	}

	// 设置玩家所在队伍id
	user.SetTeamID(0)
	user.teamInfo = nil
	user.SetPlayerGameState(common.StateFree)
}

func (user *LobbyUser) delMatch() {
	if user.GetPlayerGameState() != common.StateMatching || user.GetTeamID() != 0 {
		return
	}

	if user.matchMgrSoloProxy == nil {
		user.Warn("delMatch failed, matchMgrSoloProxy is nil")
		return
	}

	if err := user.matchMgrSoloProxy.RPC(common.ServerTypeMatch, "CancelSoloQueue",
		GetSrvInst().GetSrvID(), user.GetID()); err != nil {
		user.Error("RPC CancelSoloQueue err: ", err)
		return
	}
}

func (user *LobbyUser) tsssdkLogin() {
	if user.loginMsg == nil {
		user.Warn("tsssdkLogin failed, loginMsg is nil")
		return
	}

	ver, err := strconv.Atoi(strings.Replace(user.loginMsg.ClientVersion, ".", "", -1))
	if err != nil {
		user.Error("Atoi err: ", err)
		ver = 0
	}

	ip, err := dbservice.SessionUtil(user.GetDBID()).GetIP()
	if err != nil {
		user.Error("GetIP err: ", err)
		ip = "0.0.0.0"
	}

	err = tsssdk.OnPlayerLogin(user.GetID(), user.loginMsg.VOpenID, uint8(user.loginMsg.PlatID), user.loginMsg.IZoneAreaID,
		user.GetDBID(), uint32(ver), ip, user.GetName())
	if err != nil {
		user.Error("OnPlayerLogin err: ", err)
	}
}

func (user *LobbyUser) tsssdkLogout() {
	if user.loginMsg == nil {
		user.Warn("tsssdkLogout failed, loginMsg is nil")
		return
	}

	if err := tsssdk.OnPlayerLogout(user.GetID(), user.loginMsg.VOpenID, uint8(user.loginMsg.PlatID),
		user.loginMsg.IZoneAreaID, user.GetDBID()); err != nil {
		user.Error("OnPlayerLogout err: ", err)
	}
}

func (user *LobbyUser) tsssdkRecvData(data []byte) {
	if user.loginMsg == nil {
		user.Warn("tsssdkRecvData failed, loginMsg is nil")
		return
	}

	if err := tsssdk.OnRecvAntiData(user.GetID(), user.loginMsg.VOpenID, uint8(user.loginMsg.PlatID),
		user.loginMsg.IZoneAreaID, user.GetDBID(), data); err != nil {
		user.Error("OnRecvAntiData err: ", err)
	}
}

func (user *LobbyUser) tlogOnPlayerLogout() {
	var err error

	// 在线时长计算
	tNow := time.Now()
	now := tNow.Unix()
	user.SetLogoutTime(now)
	onlineTime := user.GetOnlineTime()
	onlineTime += (now - user.loginTime)
	user.SetOnlineTime(onlineTime)

	todayOnlineTime := user.CalTodayOnlineTime(user.GetTodayOnlineTime(), user.loginTime, tNow)
	user.SetTodayOnlineTime(todayOnlineTime)

	if user.loginMsg != nil {
		//qqscorebatch 累计在线时间
		msdk.QQScoreBatch(common.QQAppIDStr, common.MSDKKey, user.loginMsg.VOpenID, user.GetAccessToken(), user.loginMsg.PlatID,
			user.GetName(), user.GetDBID(), 6000, 1, fmt.Sprintf("%v", todayOnlineTime), "不过期")

		// 在线人数计数
		if err = GetSrvInst().logoutCnt(user.loginMsg.VGameAppID, int(user.loginMsg.PlatID)); err != nil {
			user.Error("logoutCnt err: ", err)
		}

		// 登出日志
		msg := &protoMsg.PlayerLogout{}
		msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
		msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
		msg.VGameAppID = user.loginMsg.VGameAppID
		msg.PlatID = user.loginMsg.PlatID
		msg.IZoneAreaID = 0
		msg.VOpenID = user.loginMsg.VOpenID
		if msg.VOpenID == "NoOpenID" || msg.VOpenID == "" {
			msg.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
		}
		msg.OnlineTime = uint32(now - user.loginTime)
		msg.Level = user.GetLevel()
		msg.PlayerFriendsNum = user.GetFriendsNum()
		msg.ClientVersion = user.loginMsg.ClientVersion
		msg.SystemHardware = user.loginMsg.SystemHardware
		msg.TelecomOper = user.loginMsg.TelecomOper
		msg.Network = user.loginMsg.Network
		msg.Network = strings.Replace(msg.Network, "|", "-", -1)
		tlog.Format(msg)
	} else {
		user.Warn("tlogOnPlayerLogout failed, loginMsg is nil")
	}
}

// GetLoginChannel 1: 微信  2: QQ 3: 游客
func (user *LobbyUser) GetLoginChannel() uint32 {
	return user.GetPlayerLogin().LoginChannel
}

// 玩家注册初始化数据
func (user *LobbyUser) onUserRegister() {

}

func (user *LobbyUser) InitAnnounceData() {
	// 初始玩家公告数据
	data := db.GetAllAnnuoncingData()
	if data == nil {
		return
	}

	if err := user.RPC(iserver.ServerTypeClient, "InitAnnounceData", data); err != nil {
		user.Error("RPC InitAnnounceData err: ", err)
	}
}

// SetPlayerGameState 设置玩家游戏状态
func (user *LobbyUser) SetPlayerGameState(state uint64) {
	user.friendMgr.syncFriendState(state)
	db.PlayerTempUtil(user.GetDBID()).SetGameState(state)
	user.gameState = state
	if state == common.StateOffline {
		if user.teamMgrProxy != nil && user.GetTeamID() != 0 {
			user.teamMgrProxy.RPC(common.ServerTypeMatch, "CancelTeamReady", user.GetID(), user.GetTeamID())
		}
	}
}

// GetPlayerGameState 获取玩家游戏状态
func (user *LobbyUser) GetPlayerGameState() uint64 {
	return user.gameState
	// return db.PlayerTempUtil(user.GetDBID()).GetGameState()
}

func (user *LobbyUser) Loop() {
	select {
	case <-user.onlineReportTicker.C:
		if user.loginMsg == nil || user.loginMsg.LoginChannel != 2 { //不是手Q不上报
			return
		}

		now := time.Now()
		nowUnix := now.Unix()
		if nowUnix < user.lastBatchTime+300 { //5分钟上报一次
			zeroUnixTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
			if nowUnix < zeroUnixTime+24*3600-5 { //上报的时间还没到，但是下次触发还是在今天就返回，不然今天就再触发一次
				return
			}
		}
		user.lastBatchTime = nowUnix

		todayOnlineTime := user.CalTodayOnlineTime(user.GetTodayOnlineTime(), user.loginTime, now)
		//qqscorebatch 累计在线时间
		msdk.QQScoreBatch(common.QQAppIDStr, common.MSDKKey, user.loginMsg.VOpenID, user.GetAccessToken(), user.loginMsg.PlatID,
			user.GetName(), user.GetDBID(), 6000, 1, fmt.Sprintf("%v", todayOnlineTime), "不过期")

	case <-user.secondTicker.C:
		weekTime := int64(common.GetTBSystemValue(common.System_BallStarFreshWeek))
		hourTime := int64(common.GetTBSystemValue(common.System_BallStarFreshHour))
		freshTime := common.GetThisWeekBeginStamp() + (weekTime-1)*86400 + hourTime*3600
		if time.Now().Unix() == freshTime {
			user.activityMgr.syncBallStarInfo()
		}

	default:
	}
}

// CalTodayOnlineTime 计算当天在线时间
func (user *LobbyUser) CalTodayOnlineTime(todayOnlineTime, loginTime int64, now time.Time) int64 {
	zeroUnixTime := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	if zeroUnixTime > loginTime {
		return now.Unix() - zeroUnixTime
	}
	return todayOnlineTime + (now.Unix() - loginTime)
}

// AnnounceDataCenterAddr 通知客户端DataCenter服地址
func (user *LobbyUser) AnnounceDataCenterAddr() {
	dataCenterOuterAddr, err := db.GetDataCenterAddr("DataCenterOuterAddr")
	if err != nil {
		user.Error("GetDataCenterAddr err: ", err)
		return
	}

	//user.Debug("dataCenterOuterAddr:", dataCenterOuterAddr)
	if err := user.RPC(iserver.ServerTypeClient, "InitNotifyMysqlDbAddr", "http://"+dataCenterOuterAddr); err != nil {
		user.Error("RPC InitNotifyMysqlDbAddr err: ", err)
	}
}

// ChangeNameInfo 修改名字 的相关信息
func (user *LobbyUser) ChangeNameInfo() {
	util := db.PlayerInfoUtil(user.GetDBID())
	num, e := util.GetChangeNameNum()
	if e != nil && e != redis.ErrNil {
		return
	} else if num == 0 {
		if user.GetName() != "" {
			util.SetChangeNameNum(1)
			num = 1
		}
	}
	stamp, e := util.GetChangeNameTime()
	if e != nil && e != redis.ErrNil {
		return
	}
	user.RPC(iserver.ServerTypeClient, "ChangeNameInfo", num, stamp)
}

// InsigniaInfo 返回某个玩家的勋章信息
func (user *LobbyUser) InsigniaInfo(uid uint64) {
	util := db.PlayerInsigniaUtil(uid)
	msg := &protoMsg.InsigniaInfo{}
	flag := util.GetInsigniaFlag() //勋章标记

	for id, num := range util.GetInsignia() {
		msg.Info = append(msg.Info, &protoMsg.Insignia{
			Id:    id,
			Count: num,
			Flag:  flag[id],
		})
	}
	msg.Use, _ = db.PlayerInfoUtil(uid).GetInsigniaUsed()
	user.RPC(iserver.ServerTypeClient, "InsigniaInfoRet", uid, msg)
	user.Debug("InsigniaInfoRet uid:", uid, " msg:", msg)
}

//AddAchievementData 成就数据变更 并刷新成就进度
func (user *LobbyUser) AddAchievementData(id uint64, value uint32) {
	util := db.PlayerAchievementUtil(user.GetDBID())
	r := util.AddAchievementData(uint32(id), value)
	msg := user.FreshAchievement(user.GetDBID(), map[uint64]float64{
		id: r,
	})
	if msg == nil {
		return
	}
	user.RPC(iserver.ServerTypeClient, "AchievementAddNotify", msg)
	user.Debug("AddAchievementData:", msg)
}

func (user *LobbyUser) AchievementInit() {
	if db.PlayerAchievementUtil(user.GetDBID()).IsInit() {
		return
	}
	freshData := map[uint64]float64{}
	util := db.PlayerGoodsUtil(user.GetDBID())
	for _, v := range util.GetAllGoodsInfo() {
		itemData, ok := excel.GetStore(uint64(v.Id))
		if !ok || itemData.Timelimit != "" {
			continue
		}

		if itemData.Type == 1 {
			freshData[common.AchievementRole]++
		}
		if itemData.Type == 2 {
			freshData[common.AchievementParachute]++
		}
	}
	freshData[common.AchievementCoin] = float64(user.GetCoin())
	for k, v := range freshData {
		db.PlayerAchievementUtil(user.GetDBID()).AddAchievementData(uint32(k), v)
	}
	msg := user.FreshAchievement(user.GetDBID(), freshData)
	if msg != nil {
		//user.RPC(iserver.ServerTypeClient, "AchievementAddNotify", msg)
	}
	db.PlayerAchievementUtil(user.GetDBID()).SetInit()
}

//FreshAchievement 刷新成就进度
func (user *LobbyUser) FreshAchievement(uid uint64, freshData map[uint64]float64) *protoMsg.AchievmentInfo {
	if len(freshData) == 0 {
		return nil
	}
	util := db.PlayerAchievementUtil(uid)
	achieveProcess := util.GetAchieveInfo()
	var needSave []*db.AchieveInfo
	excelData := excel.GetAccomplishmentMap()
	level, exp := util.GetLevelInfo()
	for _, data := range excelData {
		newValue, ok := freshData[data.Condition1]
		if !ok {
			continue
		}

		oneAchieve := achieveProcess[data.Id]
		if oneAchieve == nil {
			oneAchieve = &db.AchieveInfo{
				Id: data.Id,
			}
		}
		oneProcess := len(oneAchieve.Stamp)
		var fresh bool
		for k, num := range data.Amount {
			if oneProcess >= k+1 {
				continue
			}
			if newValue >= float64(num) {
				oneAchieve.Stamp = append(oneAchieve.Stamp, uint64(time.Now().Unix()))
				oneAchieve.Flag = 1 // 1:代表是新获得的成就, 0:旧成就
				fresh = true
				exp += data.Experience[k]
				levelData, ok := excel.GetAchievementLevel(uint64(level))
				if ok {
					for exp >= uint32(levelData.Experience) {
						needExp := levelData.Experience
						levelData, ok = excel.GetAchievementLevel(uint64(level + 1))
						if !ok {
							break
						}
						exp -= uint32(needExp)
						level++
					}
				}
				user.AchievementFlow(uint32(oneAchieve.Id), level, uint32(exp))
			}
		}
		if fresh {
			needSave = append(needSave, oneAchieve)
		}
	}
	if len(needSave) > 0 {
		util.AddAchieve(needSave)
		util.SetExp(exp)
		util.SetLevel(level)

		// 成就公告
		for _, v := range needSave {
			user.announcementBroad(1, uint32(v.Id), 0, 0)
		}
	}

	msg := &protoMsg.AchievmentInfo{}
	msg.Level = level
	msg.Exp = exp
	for _, v := range needSave {
		msg.List = append(msg.List, &protoMsg.Achievement{
			Id:    uint32(v.Id),
			Stamp: v.Stamp,
			Flag:  v.Flag,
		})
	}
	for k, v := range freshData {
		msg.Process = append(msg.Process, &protoMsg.AchievementProcess{
			Id:  uint32(k),
			Num: float32(v),
		})
	}

	return msg
}

// CareerCurSeason 向客户端推送赛季id
func (user *LobbyUser) CareerCurSeason() {
	id := common.GetSeason()

	user.Debug("CareerCurSeason:", id)
	if err := user.RPC(iserver.ServerTypeClient, "CareerCurSeason", int32(id)); err != nil {
		user.Error("err:", err)
	}
}

// getGameRecordDetail 获取战绩详情
func (user *LobbyUser) getGameRecordDetail() *protoMsg.GameRecordDetail {
	detail := &protoMsg.GameRecordDetail{
		Id:    user.GetID(),
		Level: user.GetLevel(),
	}

	//当前赛季数据
	curData := &datadef.CareerBase{}
	if err := db.PlayerCareerDataUtil(common.PlayerCareerTotalData, user.GetDBID(), common.GetSeason()).GetRoundData(curData); err != nil {
		user.Error("GetRoundData err: ", err)
		return detail
	}

	detail.SoloRating = uint32(math.Ceil(float64(curData.SoloRating)))
	detail.DuoRating = uint32(math.Ceil(float64(curData.DuoRating)))
	detail.SquadRating = uint32(math.Ceil(float64(curData.SquadRating)))

	//所有赛季总数据
	totalData := &datadef.CareerBase{}
	if err := db.PlayerCareerDataUtil(common.PlayerCareerTotalData, user.GetDBID(), 0).GetRoundData(totalData); err != nil {
		user.Error("GetRoundData err: ", err)
		return detail
	}

	detail.Battles = totalData.TotalBattleNum
	detail.Wins = totalData.FirstNum
	detail.TopTens = totalData.TopTenNum

	dies := totalData.TotalBattleNum - totalData.FirstNum
	if dies == 0 {
		dies = 1
	}

	detail.Kda = float32(totalData.TotalKillNum) / float32(dies)

	return detail
}

// syncDataToMatch 向Match同步需要的数据
func (user *LobbyUser) syncDataToMatch() {
	teamID := user.GetTeamID()
	if teamID == 0 {
		return
	}

	if user.teamMgrProxy == nil {
		return
	}

	//匹配积分
	ranks := user.GetRanksByMode(user.matchMode, 2, 4)
	if ranks == nil {
		return
	}

	detail := user.getGameRecordDetail()
	user.teamMgrProxy.RPC(common.ServerTypeMatch, "SyncGameRecord", user.GetID(), teamID, ranks[2], ranks[4], detail)
}

// checkModeLightFlag 提示客户端是否打开各种赛事高亮箭头提示信息
func (user *LobbyUser) checkModeLightFlag() {
	systemData := excel.GetSystemMap()
	flagStartTime := int64(systemData[201].Value)
	flagEndTime := int64(systemData[202].Value)
	flagLoopTime := 24 * 60 * 60 * int64(systemData[203].Value)
	var num int64 = 0
	if flagLoopTime != 0 {
		num = (time.Now().Unix() - flagStartTime) / flagLoopTime
	}

	if time.Now().Unix() < flagStartTime {
		return
	}

	if time.Now().Unix() > flagEndTime+flagLoopTime*num {
		return
	}

	user.Debug("CheckModeLightFlag num:", num, " flagLoopTime:", flagLoopTime)
	user.RPC(iserver.ServerTypeClient, "CheckModeLightFlag")
}

// checkVeteranRecallList 检测老兵召回列表是否有好友
func (user *LobbyUser) checkVeteranRecallList() {
	callbackData, ok := excel.GetCallback(1)
	if !ok {
		user.Error("GetCallback fail!")
		return
	}

	if time.Now().Unix()-user.GetLogoutTime() < int64(callbackData.Exittime)*86400 {
		return
	}

	veteranRecallList := &db.VeteranRecallList{
		List: make(map[uint64]uint32),
	}
	playerInfoUtil := db.PlayerInfoUtil(user.GetDBID())
	if err := playerInfoUtil.GetVeteranRecallList(veteranRecallList); err != nil {
		user.Error("GetVeteranRecallList err:", err)
		return
	}

	for k, v := range veteranRecallList.List {
		if v != 1 || k == 0 {
			continue
		}
		user.activityMgr.recallSuccessNotify(k)
		veteranRecallList.List[k] = 0
	}

	user.Debug("SetVeteranRecallList list:", veteranRecallList.List)
	if err := playerInfoUtil.SetVeteranRecallList(veteranRecallList); err != nil {
		user.Error("SetVeteranRecallList err:", err)
	}
}

// weaponEquip 装备武器 action为1表示装备，action为0表示卸载
func (user *LobbyUser) weaponEquip(weaponId, additionId uint32, action uint8) {
	user.Info("weaponEquip, weaponId: ", weaponId, " additionId: ", additionId, " action: ", action)

	addItem, ok := excel.GetItem(uint64(additionId))
	if !ok {
		return
	}

	if weaponId != additionId {
		util := db.PlayerGoodsUtil(user.GetDBID())

		info, err := util.GetGoodsInfo(additionId)
		if err != nil || info == nil {
			user.Error("GetGoodsInfo err: ", err)
			return
		}

		info.Used = uint32(action)
		if !util.AddGoodsInfo(info) {
			return
		}

		// 通知客户端
		user.RPC(iserver.ServerTypeClient, "UpdateGoods", info.ToProto())
	}

	weapons := user.GetWeaponEquipInGame()
	if weapons == nil {
		weapons = &protoMsg.WeaponInGame{}
	}

	var weapon *protoMsg.WeaponInfo
	var index int = -1

	for _, v := range weapons.GetWeapons() {
		if v == nil {
			continue
		}

		if v.GetWeaponId() == weaponId {
			weapon = v
			break
		}
	}

	for i, v := range weapon.GetAdditions() {
		item, ok := excel.GetItem(uint64(v))
		if !ok {
			continue
		}

		if item.Type == addItem.Type {
			index = i
			break
		}
	}

	if action == 1 {
		if index == -1 { //未装备过同种类型配件
			if weapon != nil {
				weapon.Additions = append(weapon.Additions, additionId)
			} else {
				weapons.Weapons = append(weapons.Weapons, &protoMsg.WeaponInfo{
					WeaponId:  weaponId,
					Additions: []uint32{additionId},
				})
			}
		} else { //已装备过同种类型配件
			if weapon.Additions[index] != additionId {
				weapon.Additions = append(weapon.Additions[:index], weapon.Additions[index+1:]...)
				if additionId != weaponId {
					weapon.Additions = append(weapon.Additions, additionId)
				}
			}
		}
	} else {
		if index != -1 { //卸载配件
			weapon.Additions = append(weapon.Additions[:index], weapon.Additions[index+1:]...)
		}
	}

	user.SetWeaponEquipInGame(weapons)
	user.SetWeaponEquipInGameDirty()
	user.syncWeaponEquipment()
}

// syncWeaponEquipment 向客户端同步武器装备信息
func (user *LobbyUser) syncWeaponEquipment() {
	weapons := user.GetWeaponEquipInGame()
	user.RPC(iserver.ServerTypeClient, "WeaponEquipInGameSync", weapons)
	user.Infof("WeaponInGame: %+v\n", weapons)
}

// festivalDataFlush 节日活动数据刷新
func (user *LobbyUser) festivalDataFlush(kind uint32) {
	data := make(map[uint64]int32)

	switch kind {
	case common.Act_Login:
		data[uint64(common.Act_Login)] = 1

		if time.Now().Hour() >= 12 && time.Now().Hour() <= 13 {
			data[uint64(common.Act_LoginNoon)] = 1
		}
		if time.Now().Hour() >= 18 && time.Now().Hour() <= 20 {
			data[uint64(common.Act_LoginNight)] = 1
		}
	case common.Act_ShareNum:
		data[uint64(common.Act_ShareNum)] = 1
	case common.Act_OpenBox:
		data[uint64(common.Act_OpenBox)] = 1
	default:
		return
	}

	user.activityMgr.updateActivityInfo(data)
}

// isTmpBookingFriend 是否向我发送过预约请求的好友
func (user *LobbyUser) isTmpBookingFriend(uid uint64) bool {
	list := db.PlayerTempUtil(user.GetDBID()).GetTmpBookingFriends()
	for _, v := range list {
		if v == uid {
			return true
		}
	}
	return false
}

// isTmpBookedFriend 是否已经发送过预约请求的好友
func (user *LobbyUser) isTmpBookedFriend(uid uint64) bool {
	list := db.PlayerTempUtil(user.GetDBID()).GetTmpBookedFriends()
	for _, v := range list {
		if v == uid {
			return true
		}
	}
	return false
}

// clearTmpBookingFriends 玩家在未响应预约邀请的情况下断线或者退出游戏，清除对应的邀请记录
func (user *LobbyUser) clearTmpBookingFriends() {
	list := db.PlayerTempUtil(user.GetDBID()).GetTmpBookingFriends()
	proc := &LobbyUserMsgProc{user: user}

	for _, v := range list {
		proc.RPC_BattleBookingRespReq(uint8(4), v)
	}
}

// cancelBattleBooking 取消预约关系
func (user *LobbyUser) cancelBattleBooking(uid uint64) {
	if uid == 0 {
		return
	}

	util1 := db.PlayerTempUtil(user.GetDBID())
	util2 := db.PlayerTempUtil(uid)

	if util1.GetBookingFriend() == uid {
		util1.SetBookingFriend(0)
		util2.DelBookedFriend(user.GetDBID())
	} else if user.isBookedFriend(uid) {
		util1.DelBookedFriend(uid)
		util2.SetBookingFriend(0)
	}
}

// isBookedFriend 是否已经预约的好友
func (user *LobbyUser) isBookedFriend(uid uint64) bool {
	list := db.PlayerTempUtil(user.GetDBID()).GetBookedFriends()
	for _, v := range list {
		if v == uid {
			return true
		}
	}
	return false
}

// pullBookedFriendsToTeam 将回到大厅的预约对象拉进队伍
func (user *LobbyUser) pullBookedFriendsToTeam() {
	teamID := user.GetTeamID()
	if teamID == 0 {
		return
	}

	util1 := db.PlayerTempUtil(user.GetDBID())
	for _, v := range util1.GetReadyBookedFriends() {

		util2 := db.PlayerTempUtil(v)
		if util2.GetGameState() != common.StateFree && util2.GetGameState() != common.StateMatchWaiting {
			continue
		}

		if util2.GetPlayerTeamID() == teamID {
			continue
		}

		util1.DelReadyBookedFriend(v)
		user.friendMgr.sendMsgToFriend(v, common.ServerTypeLobby, "InviteRsp", teamID)
	}
}

// syncBattleBooking 预约关系发生变化时或断线重连时向客户端同步有效的预约关系
func (user *LobbyUser) syncBattleBooking() {
	info1 := &protoMsg.BattleBookingInfo{
		Typ: 1,
	}
	info2 := &protoMsg.BattleBookingInfo{
		Typ: 2,
	}

	util := db.PlayerTempUtil(user.GetDBID())
	bookedFriends := util.GetBookedFriends()
	tmpBookedFriends := util.GetTmpBookedFriends()

	if len(bookedFriends) > 0 || len(tmpBookedFriends) > 0 {
		info1.List = bookedFriends
		info1.Tmplist = tmpBookedFriends
	}

	bookingFriend := util.GetBookingFriend()
	tmpBookingFriends := util.GetTmpBookingFriends()

	if bookingFriend != 0 || len(tmpBookingFriends) > 0 {
		info2.List = []uint64{bookingFriend}
		info2.Tmplist = tmpBookingFriends
	}

	user.RPC(iserver.ServerTypeClient, "SyncBattleBooking", info1)
	user.RPC(iserver.ServerTypeClient, "SyncBattleBooking", info2)
}

// isTeamFull 队伍是否满员
func isTeamFull(teamID uint64) bool {
	if teamID == 0 {
		return false
	}

	util := db.PlayerTeamUtil(teamID)

	num, err := util.GetTeamNum()
	if err != nil {
		return false
	}

	if num >= 4 {
		return true
	}

	matchMode, err := util.GetMatchMode()
	if err != nil {
		return false
	}

	if !common.IsMatchModeOk(matchMode, 4) {
		return num >= 2
	}

	return false
}

// getChaterInfo 聊天者个人信息
func (user *LobbyUser) getChaterInfo() *protoMsg.ChaterInfo {
	msg := &protoMsg.ChaterInfo{
		Uid:        user.GetDBID(),
		Name_:      user.GetName(),
		Url:        user.GetPicture(),
		Level:      user.GetLevel(),
		BattleTeam: "",
		NameColor:  common.GetPlayerNameColor(user.GetDBID()),
	}

	if user.GetGender() == "男" {
		msg.Gender = 1
	} else if user.GetGender() == "女" {
		msg.Gender = 2
	}

	return msg
}

// doTeamCustomResp 响应当队长或找队伍
// 返回码：0表示成功，1表示发布已失效 2表示队伍不存在 3队伍不匹配 4队伍已满 5表示已经在同一个队伍中
func (user *LobbyUser) doTeamCustomResp(uid, id uint64) uint8 {
	info := getValidTeamCustom(uid, id)
	if info == nil {
		return 1
	}

	util1 := db.PlayerTempUtil(user.GetDBID())
	util2 := db.PlayerTempUtil(uid)

	switch info.Typ {
	case 0:
		{
			teamID := util2.GetPlayerTeamID()
			if teamID == 0 {
				return 2
			}

			if user.GetTeamID() == teamID {
				return 5
			}

			if !isCustomTeamMatch(teamID, info) {
				return 1
			}

			if isCustomTeamFull(teamID, info) {
				return 4
			}

			if util2.GetGameState() == common.StateMatching {
				util2.SetToInviteUser(user.GetDBID())
				user.friendMgr.sendMsgToFriend(uid, common.ServerTypeLobby, "CancelEnterRoom")
			} else {
				user.pullUserToTeam(user.GetDBID(), teamID)
			}
		}
	case 1:
		{
			teamID := user.GetTeamID()
			if teamID == 0 {
				return 2
			}

			if util2.GetPlayerTeamID() == teamID {
				return 5
			}

			if !isCustomTeamMatch(teamID, info) {
				return 3
			}

			if isCustomTeamFull(teamID, info) {
				return 4
			}

			if user.GetPlayerGameState() == common.StateMatching {
				util1.SetToInviteUser(uid)
				proc := &LobbyUserMsgProc{user: user}
				proc.RPC_CancelEnterRoom()
			} else {
				user.pullUserToTeam(uid, teamID)
			}

			util2.ClearTeamCustoms()
		}
	}

	return 0
}

// pullUserToTeam 将玩家拉入队伍
func (user *LobbyUser) pullUserToTeam(uid, teamID uint64) {
	if teamID == 0 {
		return
	}

	state := user.friendMgr.getFriendState(uid)

	if state == common.StateFree || state == common.StateMatchWaiting {
		user.friendMgr.sendMsgToFriend(uid, common.ServerTypeLobby, "InviteRsp", teamID)
	} else if state == common.StateMatching {
		db.PlayerTempUtil(uid).SetToJoinTeam(teamID)
		user.friendMgr.sendMsgToFriend(uid, common.ServerTypeLobby, "CancelEnterRoom")
	}
}

// isCustomTeamFull 判断玩家所在队伍人数是否已经达到定制要求
func isCustomTeamFull(teamID uint64, info *protoMsg.TeamCustom) bool {
	if teamID == 0 || info == nil {
		return false
	}

	num, err := db.PlayerTeamUtil(teamID).GetTeamNum()
	if err != nil {
		return false
	}

	switch info.GetMatchTyp() {
	case 0:
		return num >= 2
	case 1:
		return num >= 4
	}

	return false
}

// isCustomTeamMatch 判断玩家所在队伍是否与定制队伍信息匹配
func isCustomTeamMatch(teamID uint64, info *protoMsg.TeamCustom) bool {
	util := db.PlayerTeamUtil(teamID)

	matchMode, err := util.GetMatchMode()
	if err != nil {
		return false
	}

	if matchMode != info.MatchMode {
		return false
	}

	mapid, err := util.GetTeamMap()
	if err != nil {
		return false
	}

	if mapid != info.Mapid {
		return false
	}

	teamTyp, err := util.GetTeamType()
	if err != nil {
		return false
	}

	if teamTyp != uint8(info.MatchTyp) {
		return false
	}

	return true
}

// getValidTeamCustom 根据id获取有效的队伍定制信息
func getValidTeamCustom(uid, id uint64) *protoMsg.TeamCustom {
	var (
		info  *protoMsg.TeamCustom
		index int
	)

	teamCustoms := db.PlayerTempUtil(uid).GetTeamCustoms()
	for i, v := range teamCustoms {
		if v.Id == id {
			info = v
			index = i
			break
		}
	}

	if info == nil {
		return nil
	}

	for _, v := range teamCustoms[index+1:] {
		if v.Typ != info.Typ || v.MatchMode != info.MatchMode ||
			v.MatchTyp != info.MatchTyp || v.Mapid != info.Mapid {
			return nil
		}
	}

	return info
}

// announcementBroad 公告广播
// typ为1时表示成就公告，typ为2表示物品公告
func (user *LobbyUser) announcementBroad(typ uint8, id, days, num uint32) {
	detail := &protoMsg.AnnouncementDetail{
		Uid:       user.GetDBID(),
		Name_:     user.GetName(),
		Typ:       uint32(typ),
		Id:        id,
		Days:      days,
		Num:       num,
		NameColor: common.GetPlayerNameColor(user.GetDBID()),
	}

	GetSrvInst().FireEvent(iserver.RPCChannel, "AnnouncementNotify", detail)
}

// syncUnreadPrivateChat 通知玩家未读的私聊信息
func (user *LobbyUser) syncUnreadPrivateChat() {
	msg := db.GetFriendUtil(user.GetDBID()).GetChatDetail()
	user.RPC(iserver.ServerTypeClient, "UnreadPrivateChatNotify", msg)
}

// syncRedDotOnce 通知玩家一次行点击类红点
func (user *LobbyUser) syncRedDotOnce() {
	util := db.PlayerInfoUtil(user.GetDBID())
	info := &db.RedDotOnce{}
	info.RedDot = make(map[uint32]uint32)

	if err := util.GetRedDotOnce(info); err != nil {
		user.Error("RPC_SetRedDotOnce err:", err)
		return
	}

	msg := &protoMsg.RedDotList{}
	for k, v := range info.RedDot {
		redDot := &protoMsg.RedDotOnce{}
		redDot.ID = k
		redDot.Dot = v

		msg.List = append(msg.List, redDot)
	}

	user.RPC(iserver.ServerTypeClient, "syncRedDotOnce", msg)
}

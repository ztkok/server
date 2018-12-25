package main

import (
	"Lobby/online"
	"common"
	"db"
	"errors"
	"fmt"
	"ipip"
	"protoMsg"
	"reflect"
	"sync"
	"time"
	"zeus/iserver"
	"zeus/msgdef"
	"zeus/server"
	"zeus/tlog"
	"zeus/tsssdk"

	log "github.com/cihub/seelog"
	"github.com/robfig/cron"
	"github.com/spf13/viper"
)

var srvInst *LobbySrv

// GetSrvInst 获取服务器全局实例
func GetSrvInst() *LobbySrv {
	if srvInst == nil {
		common.InitMsg()

		srvInst = &LobbySrv{}

		srvID := uint64(viper.GetInt("Lobby.FlagId"))
		pmin := viper.GetInt("Lobby.PortMin")
		pmax := viper.GetInt("Lobby.PortMax")

		srvInst.sqlUser = viper.GetString("Lobby.MySQLUser")
		srvInst.sqlPwd = viper.GetString("Lobby.MySQLPwd")
		srvInst.sqlAddr = viper.GetString("Lobby.MySQLAddr")
		srvInst.sqlDB = viper.GetString("Lobby.MySQLDB")
		srvInst.sqlTB = viper.GetString("Lobby.MySQLTable")
		srvInst.msdkAddr = viper.GetString("Config.MSDKAddr")

		innerPort := server.GetValidSrvPort(pmin, pmax)
		innerAddr := viper.GetString("Lobby.InnerAddr")
		fps := viper.GetInt("Lobby.FPS")
		srvInst.IServer = server.NewServer(common.ServerTypeLobby, srvID, innerAddr+":"+innerPort, "", fps, srvInst)

		srvInst.onlineCnter = make(map[string]*online.Cnter)

		tlogAddr := viper.GetString("Config.TLogAddr")
		if tlogAddr != "" {
			if err := tlog.ConfigRemoteAddr(tlogAddr); err != nil {
				log.Error(err)
			}
		}

		log.Info("Lobby init")
		log.Info("ServerID: ", srvID)
		log.Info("InnerAddr: ", innerAddr+":"+innerPort)
	}

	return srvInst
}

// LobbySrv 大厅服务器
type LobbySrv struct {
	iserver.IServer

	sqlUser  string
	sqlPwd   string
	sqlAddr  string
	sqlDB    string
	sqlTB    string
	msdkAddr string

	// tlog 相关
	onlineCnter map[string]*online.Cnter
	stateLogger *tlog.StateLogger

	cronTask   *cron.Cron                //定时检查
	allowModes *protoMsg.MatchModeNotify //发送给客户端的模式开放信息

	openTasks        sync.Map //开放的任务 k:string; v:map[uint8][]uint32
	taskAwards       sync.Map //任务相关奖励 k:string; v:[]uint32
	taskGroups       sync.Map //任务组数 k:string; v:uint32
	groupToUniqueId  sync.Map //任务组id与配置项唯一id的映射 k:string; v:map[uint8]uint32
	groupEnableItems sync.Map //任务组激活道具 k:string; v:[]uint32

	chats sync.Map //聊天集合，key为每一条聊天对应的UUID，value为回调函数

	activityMgr *ActivityManger
}

// Init 初始化
func (srv *LobbySrv) Init() error {
	srv.RegProtoType("Player", &LobbyUser{})

	//注册timer
	// srv.RegTimer(GetMatchMgr().doCheck, 1*time.Second) //间隔1s调用GetMatchMgr().doCheck
	// srv.RegTimer(GetTeamMgr().Loop, 1*time.Second)
	// srv.RegTimer(GetTeamMgr().boradcastTeamSum, 500*time.Millisecond)
	// srv.RegTimer(srv.doTravsal, 1*time.Second)

	srv.stateLogger = tlog.NewStateLogger(srv.GetSrvAddr(), 0, 5*time.Minute)
	srv.stateLogger.Start()

	tsssdk.Init(GetSrvInst().GetSrvID())
	tsssdk.ChatProc = srv.ChatProcCallBack
	srv.regBroadcaster()

	srv.resetComradeTasks()
	srv.resetChallengeTasks()
	srv.resetSpecialTasks()

	srv.InitActivityManager()

	//开启crond服务
	srv.startCrondService()

	//启动地理位置服务
	err := ipip.Init("../res/17monipdb.dat")
	if err != nil {
		panic(err)
	}

	log.Info("Lobby start")

	return nil
}

// MainLoop 逻辑帧每一帧都会调用
func (srv *LobbySrv) MainLoop() {
}

// Destroy 退出时调用
func (srv *LobbySrv) Destroy() {
	srv.stopOnlineUpdate()
	srv.stateLogger.Stop()
	tsssdk.Destroy()

	srv.cronTask.Stop()

	log.Info("Lobby destroyed")
}

func (srv *LobbySrv) stopOnlineUpdate() {
	for _, cnt := range srv.onlineCnter {
		if err := cnt.Stop(); err != nil {
			log.Error(err)
		}
	}
}

var errGameAppIDEmpty = errors.New("GameAppID empty")

//login更新在线人数统计Cnter
func (srv *LobbySrv) loginCnt(gameApp string, platID int) error {
	if platID != 0 && platID != 1 {
		return fmt.Errorf("Login PlatID Error %d", platID)
	}
	if gameApp == "" {
		return errGameAppIDEmpty
	}

	cnt, ok := srv.onlineCnter[gameApp]
	if !ok {
		var err error
		srv.onlineCnter[gameApp], err = online.NewCnter(srv.sqlUser, srv.sqlPwd, srv.sqlAddr,
			srv.sqlDB, srv.sqlTB, gameApp, 0, srv.GetSrvID())
		if err != nil {
			return err
		}

		cnt = srv.onlineCnter[gameApp]
		cnt.Start()
	}

	cnt.ReportOnline(platID, 1)
	return nil
}

var errLogoutWithoutLogin = errors.New("Logout without login")

//logout更新在线人数统计Cnter
func (srv *LobbySrv) logoutCnt(gameApp string, platID int) error {
	if platID != 0 && platID != 1 {
		return fmt.Errorf("Logout PlatID Error %d", platID)
	}
	if gameApp == "" {
		return errGameAppIDEmpty
	}

	cnt, ok := srv.onlineCnter[gameApp]
	if !ok {
		return errLogoutWithoutLogin
	}

	cnt.ReportOnline(platID, -1)
	return nil
}

func (srv *LobbySrv) regBroadcaster() {
	if 0 == srv.AddListener(iserver.BroadcastChannel, srv, "BroadcastMsg") {
		panic("Register broadcast channel failed")
	}
	if 0 == srv.AddListener(iserver.RPCChannel, srv, "RPCClients") {
		panic("Register rpc channel failed")
	}
}

// BroadcastMsg 广播消息到所有客户端
func (srv *LobbySrv) BroadcastMsg(msg msgdef.IMsg) {
	//log.Info("BroadcastMsg ", msg)

	srv.TravsalEntity("Player", func(o iserver.IEntity) {
		o.Post(iserver.ServerTypeClient, msg)
	})
}

// RPCClients 调用所有客户端的RPC消息
func (srv *LobbySrv) RPCClients(method string, args ...interface{}) {
	//log.Info("RPCClients ", method, args)

	srv.TravsalEntity("Player", func(o iserver.IEntity) {
		o.RPC(iserver.ServerTypeClient, method, args...)
	})
}

// InvalidHandle 服务器不可用时处理回调
func (srv *LobbySrv) InvalidHandle(entityID uint64) {
	e := srv.GetEntity(entityID)
	if e.GetType() == "Player" {

		log.Warn("Lobby server is not available for user ", e.GetDBID())

		srv.DestroyEntityAll(entityID)
	}
}

// PushActivityInfo 推送活动信息
func (srv *LobbySrv) PushActivityInfo() {
	log.Debug("PushActivityInfo")
	srv.resetComradeTasks()
	srv.resetChallengeTasks()

	srv.TravsalEntity("Player", func(e iserver.IEntity) {
		if e == nil {
			return
		}

		if user, ok := e.(*LobbyUser); ok {
			user.activityMgr.syncActiveState() //同步活动信息是否开放
			user.activityMgr.syncActiveInfo()  //同步活动领取状态信息

			user.treasureBoxMgr.syncTreasureBoxInfo() // 同步宝箱信息(0点刷新)

			user.dayTaskInfoNotify()                         //向所有在线玩家推送清空后的每日任务数据
			user.comradeTaskInfoNotify()                     //向所有在线玩家推送清空后的战友任务数据
			user.taskMgr.getChallengeTask().syncTaskDetail() //新赛季开始时向所有在线玩家推送清空后的挑战任务数据
		}
	})
}

//开启crond服务定时检查
func (srv *LobbySrv) startCrondService() {
	srv.cronTask = cron.New()

	//定时推送活动信息
	if err := srv.cronTask.AddFunc("0 0 0 1/1 * ? ", srv.PushActivityInfo); err != nil {
		log.Error("AddFunc err: ", err)
	}

	if err := srv.cronTask.AddFunc("0/1 * * * * ? ", srv.activityMgr.Loop); err != nil {
		log.Error("AddFunc err: ", err)
	}

	//检查服务器开放的匹配模式是否发生变动
	if err := srv.cronTask.AddFunc("* * * * * ?", srv.MatchModeCheck); err != nil {
		log.Error("AddFunc err: ", err)
	}

	//每分钟检测功能
	if err := srv.cronTask.AddFunc("0 * * * * ?", srv.MinuteCheck); err != nil {
		log.Error("AddFunc err: ", err)
	}

	//特训任务清零
	if err := srv.cronTask.AddFunc(common.GetSpecialTaskCronExpr(), srv.clearSpecialTask); err != nil {
		log.Error("AddFunc err: ", err)
	}

	srv.cronTask.Start()
}

//获取服务器当前开放的匹配模式信息，在匹配模式发生变动时，通知所有在线客户端
func (srv *LobbySrv) MatchModeCheck() {
	var (
		msg = &protoMsg.MatchModeNotify{}
	)

	msg.Infos = append(msg.Infos, &protoMsg.ModeInfo{
		Modeid: common.MatchModeNormal,
		Solo:   viper.GetBool("Lobby.Solo"),
		Duo:    viper.GetBool("Lobby.Duo"),
		Squad:  viper.GetBool("Lobby.Squad"),
	})

	for _, v := range common.GetModeInfosGlobalCopy() {
		var modeInfo *protoMsg.ModeInfo

		for _, info := range msg.Infos {
			if info.Modeid == v.ModeId {
				modeInfo = info
			}
		}

		if modeInfo == nil {
			modeInfo = &protoMsg.ModeInfo{
				Modeid: v.ModeId,
				Price:  v.ItemPrice,
			}
			msg.Infos = append(msg.Infos, modeInfo)
		}

		if v.SeasonStart != 0 && v.SeasonEnd != 0 {
			modeInfo.SeasonStart = time.Unix(v.SeasonStart, 0).Format("2006|01|02|15|04|05")
			modeInfo.SeasonEnd = time.Unix(v.SeasonEnd, 0).Format("2006|01|02|15|04|05")
		}

		switch v.OpenTyp {
		case 1:
			modeInfo.Solo = true
		case 2:
			modeInfo.Duo = true
		case 4:
			modeInfo.Squad = true
		default:
			log.Error("Unknow match type: ", v.OpenTyp)
		}
	}

	if !reflect.DeepEqual(srv.allowModes, msg) {
		log.Debug("--------------------------------")
		log.Debugf("allowModes: %+v\n", srv.allowModes)
		log.Debugf("msg       : %+v\n", msg)

		srv.allowModes = msg
		srv.notifyServerModeToAll()

		log.Debug("Opened match modes have changed, notify to all online clients")
		log.Debug("--------------------------------")
	}
}

//通知所有在线客户端各种比赛模式的开放信息
func (srv *LobbySrv) notifyServerModeToAll() {
	srv.TravsalEntity("Player", func(o iserver.IEntity) {
		o.RPC(iserver.ServerTypeClient, "OnlineCheckMatchModeOpen", srv.allowModes)
		if user, ok := o.(*LobbyUser); ok {
			user.notifyPlayerMode()
		}
	})
}

// resetComradeTasks 重置开放的战友任务
func (srv *LobbySrv) resetComradeTasks() {
	tasks := common.GetOpenComradeTasks()
	srv.openTasks.Store(common.TaskName_Comrade, tasks)

	log.Info("Today: ", time.Now().Format("2006-01-02"))
	log.Infof("ComradeTasks: %+v\n", tasks)
}

// resetChallengeTasks 重置开放的挑战任务
func (srv *LobbySrv) resetChallengeTasks() {
	tasks, uniqueM := common.GetOpenChallengeTasks()

	srv.openTasks.Store(common.TaskName_Challenge, tasks)
	srv.groupToUniqueId.Store(common.TaskName_Challenge, uniqueM)
	srv.taskGroups.Store(common.TaskName_Challenge, uint32(len(tasks)/2))

	awards := common.GetSeasonGradeAwards()
	srv.taskAwards.Store(common.TaskName_Challenge, awards)

	items := common.GetChallengeEnableItems()
	srv.groupEnableItems.Store(common.TaskName_Challenge, items)

	log.Infof("ChallengeTasks: %+v\n", tasks)
	log.Infof("ChallengeUniqueM: %+v\n", uniqueM)
	log.Infof("ChallengeAwards: %+v\n", awards)
	log.Infof("ChallengeItems: %+v\n", items)
}

// resetSpecialTasks 重置开放的特训任务
// 一周内服务器热加载或重启后，后续任务日开放的特训任务可能发生变化
func (srv *LobbySrv) resetSpecialTasks() {
	week := common.GetSpecialTaskWeek()
	tasks1 := db.GetOpenSpecialTasks(week)

	tasks2 := common.GetOpenSpecialTasks()
	whatDay := common.WhatDayOfSpecialWeek(common.GetSpecialTaskDay())
	var change bool

	if len(tasks1) == 0 {
		tasks1 = tasks2
		change = true
	} else {
		for i := whatDay; i <= 7; i++ {
			k := 2*i - 1
			if len(tasks1[k]) != len(tasks2[k]) {
				tasks1[k] = tasks2[k]
				tasks1[k+1] = tasks2[k+1]
				change = true
			}
		}
	}

	if change {
		db.SetOpenSpecialTasks(week, tasks1)
	}

	srv.openTasks.Store(common.TaskName_Special, tasks1)
	srv.taskGroups.Store(common.TaskName_Special, uint32(7))

	awards := common.GetSpecialLevelAwards()
	srv.taskAwards.Store(common.TaskName_Special, awards)

	log.Infof("SpecialTasks: %+v\n", tasks1)
	log.Infof("SpecialAwards: %+v\n", awards)
}

// clearSpecialTask 清零特训任务相关数据
func (srv *LobbySrv) clearSpecialTask() {
	srv.resetSpecialTasks()

	srv.TravsalEntity("Player", func(e iserver.IEntity) {
		if e == nil {
			return
		}

		if user, ok := e.(*LobbyUser); ok {
			user.taskMgr.getSpecialTask().syncTaskDetail()
		}
	})
}

// MinuteCheck 每分钟检测
func (srv *LobbySrv) MinuteCheck() {
	srv.TravsalEntity("Player", func(e iserver.IEntity) {
		if e == nil {
			return
		}

		if user, ok := e.(*LobbyUser); ok {
			user.festivalDataFlush(common.Act_Login)
		}
	})
}

// ChatProcCallBack 聊天内容检测处理后回调
func (srv *LobbySrv) ChatProcCallBack(uuid, result string) {
	v, ok := srv.chats.Load(uuid)
	if !ok {
		return
	}

	f := v.(func(string))
	f(result)

	srv.chats.Delete(uuid)
}

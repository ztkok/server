package main

import (
	"common"
	"time"
	"zeus/iserver"
	"zeus/server"
	"zeus/space"
	"zeus/tlog"

	log "github.com/cihub/seelog"
	"github.com/spf13/viper"
)

var srvInst *RoomSrv

// GetSrvInst 获取服务器全局实例
func GetSrvInst() *RoomSrv {
	if srvInst == nil {
		common.InitMsg()

		srvInst = &RoomSrv{}

		srvID := uint64(viper.GetInt("Room.FlagId"))
		pmin := viper.GetInt("Room.PortMin")
		pmax := viper.GetInt("Room.PortMax")

		innerPort := server.GetValidSrvPort(pmin, pmax)
		innerAddr := viper.GetString("Room.InnerAddr")
		fps := viper.GetInt("Room.FPS")

		opmin := viper.GetInt("Room.OuterPortMin")
		opmax := viper.GetInt("Room.OuterPortMax")

		/*
			配置中, OuterAddr+OuterPort写入到redis中, 作为服务器的对外地址
			OuterListen+随机出来的端口作为实际对外监听端口
			当配置中OuterPort为0时, OuterPort就是随机端口
		*/
		outerListen := viper.GetString("Room.OuterListen")
		listenPort := server.GetValidSrvPort(opmin, opmax)
		outerAddr := viper.GetString("Room.OuterAddr")
		outerPort := listenPort
		if viper.GetString("Room.OuterPort") != "0" {
			outerPort = viper.GetString("Room.OuterPort")
		}

		srvInst.IServer = server.NewServer(common.ServerTypeRoom, srvID, innerAddr+":"+innerPort, outerAddr+":"+outerPort, fps, srvInst)

		protocal := viper.GetString("Room.SpaceProtocal")
		maxConns := viper.GetInt("Room.MaxConns")
		srvInst.SpaceSesses = space.NewSpaceSesses(protocal, outerListen+":"+listenPort, maxConns)

		tlogAddr := viper.GetString("Config.TLogAddr")
		if tlogAddr != "" {
			if err := tlog.ConfigRemoteAddr(tlogAddr); err != nil {
				log.Error(err)
			}
		}

		encryptEnabled := viper.GetBool("Config.EncryptEnabled")
		if encryptEnabled {
			srvInst.SpaceSesses.SetEncryptEnabled()
		}

		log.Info("Room Init")
		log.Info("ServerID: ", srvID)
		log.Info("InnerAddr: ", innerAddr+":"+innerPort)
	}

	return srvInst
}

// RoomSrv 房间服务器
type RoomSrv struct {
	iserver.IServer
	*space.SpaceSesses
	stateLogger *tlog.StateLogger
}

// Init 初始化
func (srv *RoomSrv) Init() error {
	srv.RegProtoType("Player", &RoomUser{})
	srv.RegProtoType("AI", &RoomAI{})
	srv.RegProtoType("Space", &Scene{})
	srv.RegProtoType("Item", &SpaceItem{})
	srv.RegProtoType("Vehicle", &SpaceVehicle{})
	srv.RegTimer(srv.doTravsal, 1*time.Second)

	srv.stateLogger = tlog.NewStateLogger(srv.GetSrvAddr(), 0, 5*time.Minute)
	srv.stateLogger.Start()

	if err := srv.SpaceSesses.Init(); err != nil {
		panic("Init space sesses error")
	}

	InitSkillEffect()
	log.Info("Room Start")
	return nil
}

// MainLoop 逻辑帧每一帧都会调用
func (srv *RoomSrv) MainLoop() {
	srv.SpaceSesses.MainLoop()
}

// Destroy 退出时调用
func (srv *RoomSrv) Destroy() {
	srv.stateLogger.Stop()
	srv.SpaceSesses.Destroy()
	log.Info("Room Shutdown")
}

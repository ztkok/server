package main

import (
	"common"
	"time"
	"zeus/iserver"
	"zeus/server"

	log "github.com/cihub/seelog"
	"github.com/spf13/viper"
)

var srvInst *Server

// Server 匹配服务器
type Server struct {
	iserver.IServer
}

// GetSrvInst 获取服务器全局实例
func GetSrvInst() *Server {
	if srvInst == nil {
		common.InitMsg()

		srvInst = &Server{}
		srvID := uint64(viper.GetInt("Match.FlagId"))
		pmin := viper.GetInt("Match.PortMin")
		pmax := viper.GetInt("Match.PortMax")
		innerPort := server.GetValidSrvPort(pmin, pmax)
		innerAddr := viper.GetString("Match.InnerAddr")
		fps := viper.GetInt("Center.FPS")
		srvInst.IServer = server.NewServer(common.ServerTypeMatch, srvID, innerAddr+":"+innerPort, "", fps, srvInst)

		log.Info("Match Init")
		log.Info("ServerID:", srvID)
		log.Info("InnerAddr:", innerAddr+":"+innerPort)
	}

	return srvInst
}

// Init 初始化
func (srv *Server) Init() error {
	srv.RegProtoType("MatchMgr", &MatchMgr{})
	srv.RegProtoType("TeamMgr", &TeamMgr{})
	srv.RegProtoType("Space", &Scene{})
	srv.RegTimer(srv.doTravsal, 1*time.Second)

	_, err := srv.CreateEntity("MatchMgr", srv.FetchTempID(), 0, 0, common.MatchMgrSolo, false, false)
	if err != nil {
		panic(err)
	}
	_, err = srv.CreateEntity("MatchMgr", srv.FetchTempID(), 0, 0, common.MatchMgrDuo, false, false)
	if err != nil {
		panic(err)
	}
	_, err = srv.CreateEntity("MatchMgr", srv.FetchTempID(), 0, 0, common.MatchMgrSquad, false, false)
	if err != nil {
		panic(err)
	}
	_, err = srv.CreateEntity("TeamMgr", srv.FetchTempID(), 0, 0, common.TeamMgr, false, false)
	if err != nil {
		panic(err)
	}

	log.Info("Match start")
	return nil
}

// MainLoop 逻辑帧每一帧都会调用
func (srv *Server) MainLoop() {

}

// Destroy 退出时调用
func (srv *Server) Destroy() {

}

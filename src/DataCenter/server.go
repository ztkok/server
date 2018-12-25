package main

import (
	"DataCenter/data"
	"db"
	"net/http"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	log "github.com/cihub/seelog"
	_ "github.com/go-sql-driver/mysql" //mysql连接库
	"github.com/spf13/viper"
)

// Server 服务器
type Server struct {
	innerAddr       string //内网地址
	innerPort       string //内网端口
	innerListen     string //内网实际监听地址
	innerListenPort string //内网实际监听端口
	outerAddr       string //外网地址
	outerPort       string //外网端口
	outerListen     string //外网实际监听地址
	outerListenPort string //外网实际监听端口

	dbRoundTableName string
	dbDayTableName   string
	dbSearchNum      string

	careerSrv *data.CareerService
}

var srv *Server

// GetSrvInst 获取服务器实例
func GetSrvInst() *Server {
	if srv == nil {
		srv = &Server{}

		srv.innerAddr = viper.GetString("DataCenter.InnerAddr")
		srv.innerPort = viper.GetString("DataCenter.InnerPort")
		srv.innerListen = viper.GetString("DataCenter.InnerListen")
		srv.innerListenPort = viper.GetString("DataCenter.InnerListenPort")
		srv.outerAddr = viper.GetString("DataCenter.OuterAddr")
		srv.outerPort = viper.GetString("DataCenter.OuterPort")
		srv.outerListen = viper.GetString("DataCenter.OuterListen")
		srv.outerListenPort = viper.GetString("DataCenter.OuterListenPort")

		srv.dbRoundTableName = viper.GetString("DataCenter.MySQLRoundTable")
		srv.dbDayTableName = viper.GetString("DataCenter.MySQLDayTable")
		srv.dbSearchNum = viper.GetString("DataCenter.MySQLSearchNumTable")

		sqlUser := viper.GetString("DataCenter.MySQLUser")
		sqlPwd := viper.GetString("DataCenter.MySQLPwd")
		sqlAddr := viper.GetString("DataCenter.MySQLAddr")
		sqlDB := viper.GetString("DataCenter.MySQLDB")
		sqlMaxConns := viper.GetInt("DataCenter.MySQLMaxOpenConns")
		sqlMaxIdleConns := viper.GetInt("DataCenter.MySQLMaxIdleConns")
		sqlMaxLifetime := time.Duration(viper.GetInt("DataCenter.MySQLMaxLifetime"))

		srv.careerSrv = data.NewCareerService(sqlUser, sqlPwd, sqlAddr, sqlDB, sqlMaxConns, sqlMaxIdleConns, sqlMaxLifetime)

		err := db.SetDataCenterAddr("DataCenterInnerAddr", srv.innerAddr+":"+srv.innerPort, "DataCenterOuterAddr", srv.outerAddr+":"+srv.outerPort)
		if err != nil {
			log.Error(err)
			log.Info("DataCenterAddr注册失败")
			return nil
		}

		log.Info("DataCenterAddr注册成功")
	}

	return srv
}

// Run 运行
func (srv *Server) Run() {
	srv.startInnerService()
	srv.startOuterService()

	log.Info("DataCenter Running!")
	return
}

// Stop 停止服务器
func (srv *Server) Stop() {
	log.Info("DataCenter Stoped")

	err := db.DelDataCenterAddr("DataCenterInnerAddr", "DataCenterOuterAddr")
	if err != nil {
		log.Error(err)
	}
	log.Info("DataCenterAddr Del!")
}

func (srv *Server) startInnerService() {
	log.Info("启动处理Room服 post过来每局的数据监听")
	api := rest.NewApi()
	api.Use(rest.DefaultCommonStack...)
	router, err := rest.MakeRouter(
		rest.Post("/dataCenter", srv.postHandler),
	)
	if err != nil {
		panic(err)
	}
	api.SetApp(router)

	go func() {
		err = http.ListenAndServe(srv.innerListen+":"+srv.innerListenPort, api.MakeHandler())
		if err != nil {
			panic(err)
		}
	}()
}

func (srv *Server) startOuterService() {
	log.Info("启动处理客户端请求Get的数据监听")
	api := rest.NewApi()
	api.Use(rest.DefaultCommonStack...)
	router, err := rest.MakeRouter(
		rest.Get("/career/#season/#uid", srv.careerHandler),                                  //生涯数据处理
		rest.Get("/friendrank/#season/#uid", srv.friendrankHandler),                          //好友排行处理
		rest.Get("/matchrecord/#season/#uid", srv.matchrecordHandler),                        //进多少天比赛记录获取
		rest.Get("/daydata/#season/#uid/#dayid", srv.daydataHandler),                         //具体哪天的数据获取
		rest.Get("/searchfriend/#season/#username", srv.searchfriendHandler),                 //搜索玩家信息
		rest.Get("/rank/#season/#model/#rating", srv.rankHandler),                            //全服各种排行信息获取(单人、双人、四人、总场次排行等)
		rest.Get("/searchtopn", srv.searchtopnHandler),                                       //搜索量前n的玩家
		rest.Get("/ranktrend/#season/#uid/#model/#timeStart/#timeEnd", srv.ranktrendHandler), //排行趋势
		rest.Get("/modeldetail/#season/#uid", srv.modeldetailHandler),                        //模式详情
		rest.Get("/braverank/#season", srv.braveRankHandler),                                 //勇者战场排行处理
		rest.Get("/braveuidrank/#season/#uid", srv.braveUserRankHandler),                     //获取某个玩家的勇者战场排名
		rest.Get("/bravefriendrank/#season/#uid", srv.braveFriendRankHandler),                //获取某个玩家的勇者战场排名
		rest.Get("/seasonwinsrank/#uid/#matchtyp/#beg/#end", srv.seasonWinsRankHandler),      //获取赛季获胜数排行榜
		rest.Get("/seasonkillsrank/#uid/#matchtyp/#beg/#end", srv.seasonKillsRankHandler),    //获取赛季击杀数排行榜                  //获取赛季击杀数排行榜
		rest.Get("/seasonratingrank/#uid/#matchtyp/#beg/#end", srv.seasonRatingRankHandler),  //获取赛季积分排行榜
	)
	if err != nil {
		panic(err)
	}
	api.SetApp(router)

	go func() {
		err = http.ListenAndServe(srv.outerListen+":"+srv.outerListenPort, api.MakeHandler())
		if err != nil {
			panic(err)
		}
	}()
}

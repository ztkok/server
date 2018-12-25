package main

import (
	"datadef"
	"strconv"

	"github.com/ant0ine/go-json-rest/rest"
	log "github.com/cihub/seelog"
)

// modeldetailHandler 模式详情查询
func (srv *Server) modeldetailHandler(w rest.ResponseWriter, r *rest.Request) {
	rankTrend := &datadef.ModelDetail{}
	defer func() {
		w.WriteJson(rankTrend)
	}()

	seasonstr := r.PathParam("season")
	season, err := strconv.Atoi(seasonstr)
	if err != nil {
		log.Error(err)
		return
	}
	uidStr := r.PathParam("uid")
	uid, err := strconv.ParseUint(uidStr, 10, 64)
	if err != nil {
		log.Error(err)
		return
	}
	log.Info("season:", season, " uid:", uid)

	solonum := srv.getModelDetailData(season, uid, 0)
	duonum := srv.getModelDetailData(season, uid, 1)
	squadnum := srv.getModelDetailData(season, uid, 2)

	careerdata := &datadef.CareerData{}
	careerdata, err = srv.careerSrv.QueryCareerData(srv.dbRoundTableName, season, uid)
	if err != nil {
		log.Error(err)
		return
	}

	rankTrend.Uid = careerdata.Uid
	rankTrend.SoloRating = careerdata.SoloRating
	rankTrend.SoloRank = careerdata.SoloRank
	rankTrend.SoloFirstNum = solonum[0]
	rankTrend.SoloTopTenNum = solonum[1]
	rankTrend.SoloBattleNum = solonum[2]
	rankTrend.SoloKillNum = solonum[3]
	rankTrend.DuoRating = careerdata.DuoRating
	rankTrend.DuoRank = careerdata.DuoRank
	rankTrend.DuoFirstNum = duonum[0]
	rankTrend.DuoTopTenNum = duonum[1]
	rankTrend.DuoBattleNum = duonum[2]
	rankTrend.DuoKillNum = duonum[3]
	rankTrend.SquadRating = careerdata.SquadRating
	rankTrend.SquadRank = careerdata.SquadRank
	rankTrend.SquadFirstNum = squadnum[0]
	rankTrend.SquadTopTenNum = squadnum[1]
	rankTrend.SquadBattleNum = squadnum[2]
	rankTrend.SquadKillNum = squadnum[3]
}

// getModelDetailData 获取固定模式下的吃鸡数、前十数、总场次和击杀数
func (srv *Server) getModelDetailData(season int, uid uint64, model uint32) [4]uint32 {
	log.Info("season:", season, " uid:", uid, " model:", model)

	var num [4]uint32
	ok, err := srv.careerSrv.IsModelDetail(srv.dbDayTableName, season, uid, model)
	if err != nil {
		log.Error(err)
		return [4]uint32{0}
	} else if !ok {
		return [4]uint32{0}
	} else {
		num, err = srv.careerSrv.QueryModelDetail(srv.dbDayTableName, season, uid, model)
		if err != nil {
			log.Error(err)
			return [4]uint32{0}
		}
	}

	return num
}

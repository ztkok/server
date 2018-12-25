package main

import (
	"datadef"
	"strconv"

	"github.com/ant0ine/go-json-rest/rest"
	log "github.com/cihub/seelog"
)

// ranktrendHandler 排行趋势查询
func (srv *Server) ranktrendHandler(w rest.ResponseWriter, r *rest.Request) {
	rankTrend := &datadef.RankTrend{}
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
	modelStr := r.PathParam("model")
	model, err := strconv.ParseUint(modelStr, 10, 32)
	if err != nil {
		log.Error(err)
		return
	}
	timeStart := r.PathParam("timeStart")
	timeEnd := r.PathParam("timeEnd")

	log.Info("season:", season, " uid:", uid, " model:", model, " timeStart:", timeStart, " timeEnd:", timeEnd)

	rankTrend, err = srv.careerSrv.QueryRankTrend(srv.dbDayTableName, season, uid, uint32(model), timeStart, timeEnd)
	if err != nil {
		log.Error(err)
		return
	}
}

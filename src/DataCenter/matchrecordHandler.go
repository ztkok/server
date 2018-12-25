package main

import (
	"datadef"
	"strconv"

	"github.com/ant0ine/go-json-rest/rest"
	log "github.com/cihub/seelog"
)

// matchrecordHandler 获取历史每天比赛记录
func (srv *Server) matchrecordHandler(w rest.ResponseWriter, r *rest.Request) {
	matchrecord := &datadef.MatchRecord{}
	defer func() {
		w.WriteJson(matchrecord)
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

	num := 10
	matchrecord, err = srv.careerSrv.QueryMatchRecord(srv.dbDayTableName, season, uid, num)
	if err != nil {
		log.Error(err)
	}
}

// daydataHandler 获取具体那天记录数据
func (srv *Server) daydataHandler(w rest.ResponseWriter, r *rest.Request) {
	setday := &datadef.SettleDayData{}
	defer func() {
		w.WriteJson(setday)
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
	dayIdStr := r.PathParam("dayid")
	dayId, err := strconv.ParseUint(dayIdStr, 10, 64)
	if err != nil {
		log.Error(err)
		return
	}
	log.Info("season:", season, " uid:", uid, " dayId:", dayId)

	dayData := &datadef.OneDayData{}
	dayData, err = srv.careerSrv.QueryOneDayData(srv.dbDayTableName, season, uid, dayId)
	if err != nil {
		log.Error(err)
		return
	}

	setday.DayId = dayData.DayId
	setday.NowTime = dayData.NowTime
	setday.DayFirstNum = dayData.DayFirstNum
	setday.DayTopTenNum = dayData.DayTopTenNum
	setday.Rating = dayData.Rating
	setday.WinRating = dayData.WinRating
	setday.KillRating = dayData.KillRating
	setday.DayEffectHarm = dayData.DayEffectHarm
	setday.DayShotNum = dayData.DayShotNum
	setday.DaySurviveTime = dayData.DaySurviveTime
	setday.DayDistance = dayData.DayDistance
	setday.DayAttackNum = dayData.DayAttackNum
	setday.DayRecoverNum = dayData.DayRecoverNum
	setday.DayRevivenum = dayData.DayRevivenum
	setday.DayHeadShotNum = dayData.DayHeadShotNum
	setday.DayBattleNum = dayData.DayBattleNum
	setday.TotalRank = dayData.TotalRank
	log.Info("OneDayData:", dayData)
}

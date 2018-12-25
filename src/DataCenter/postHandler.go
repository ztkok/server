package main

import (
	"datadef"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/ant0ine/go-json-rest/rest"
	log "github.com/cihub/seelog"
)

// postHandler 处理Room服post过来的一局数据
func (srv *Server) postHandler(w rest.ResponseWriter, r *rest.Request) {
	defer func() {
		w.WriteJson("Insert finish!")
	}()

	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		log.Error(err)
		return
	}

	req := &datadef.OneRoundData{}
	err = json.Unmarshal(body, req)
	if err != nil {
		log.Error(err)
		return
	}

	log.Info("dbRoundTableName:", srv.dbRoundTableName, " RoundData-", req)

	err = srv.careerSrv.InsertRoundDatas(srv.dbRoundTableName, req)
	if err != nil {
		log.Error(err)
	}

	srv.updateDayData(req) //更新每天统计表(playerdaydata)中的数据
}

// updateDayData 更新当天的数据统计
func (srv *Server) updateDayData(req *datadef.OneRoundData) {
	daydata := &datadef.OneDayData{}
	dayStr := time.Now().Format("20060102")
	fmt.Sscanf(dayStr, "%d", &daydata.DayId)
	daydata.Season = req.Season
	daydata.UID = req.UID
	daydata.NowTime = req.StartUnix
	daydata.StartTime = time.Unix(req.StartUnix, 0).Format("2006-01-02")

	daydata.Model = req.BattleType
	if req.Rank == 1 {
		daydata.DayFirstNum = 1
	}
	if req.Rank <= 10 {
		daydata.DayTopTenNum = 1
	}
	if req.BattleType == 0 {
		daydata.Rating = req.SoloRating
		daydata.WinRating = req.SoloWinRating
		daydata.KillRating = req.SoloKillRating
		daydata.Rank = req.SoloRank
	} else if req.BattleType == 1 {
		daydata.Rating = req.DuoRating
		daydata.WinRating = req.DuoWinRating
		daydata.KillRating = req.DuoKillRating
		daydata.Rank = req.DuoRank
	} else if req.BattleType == 2 {
		daydata.Rating = req.SquadRating
		daydata.WinRating = req.SquadWinRating
		daydata.KillRating = req.SquadKillRating
		daydata.Rank = req.SquadRank
	}
	daydata.DayEffectHarm = req.EffectHarm
	daydata.DayShotNum = req.ShotNum
	daydata.DaySurviveTime = req.EndUnix - req.StartUnix
	daydata.DayDistance = req.RunDistance
	daydata.DayAttackNum = req.AttackNum
	daydata.DayRecoverNum = req.RecoverUseNum
	daydata.DayRevivenum = req.ReviveNum
	daydata.DayHeadShotNum = req.HeadShotNum
	daydata.DayBattleNum = 1
	daydata.DayKillNum = req.KillNum
	daydata.TotalRank = req.TotalRank
	daydata.UserName = req.UserName
	log.Info("dbDayTableName:", srv.dbDayTableName, " OneDayData:", daydata)

	ok, err := srv.careerSrv.IsDayDatas(srv.dbDayTableName, daydata.DayId, daydata.UID, daydata.Model)
	if err != nil {
		log.Error(err)
		return
	} else if ok {
		err = srv.careerSrv.UpdateDayDatas(srv.dbDayTableName, daydata.DayId, daydata.UID, daydata.Model, daydata)
		if err != nil {
			log.Error(err)
			return
		}
	} else {
		err = srv.careerSrv.InsertDayDatas(srv.dbDayTableName, daydata)
		if err != nil {
			log.Error(err)
			return
		}
	}
}

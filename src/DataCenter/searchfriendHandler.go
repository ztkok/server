package main

import (
	"common"
	"datadef"
	"strconv"
	"zeus/dbservice"

	"github.com/ant0ine/go-json-rest/rest"
	log "github.com/cihub/seelog"
	"github.com/garyburd/redigo/redis"
)

// searchfriendHandler 搜索好友信息
func (srv *Server) searchfriendHandler(w rest.ResponseWriter, r *rest.Request) {
	item := &datadef.FriendRankInfo{}
	defer func() {
		w.WriteJson(item)
	}()

	seasonstr := r.PathParam("season")
	season, err := strconv.Atoi(seasonstr)
	if err != nil {
		log.Error(err)
		return
	}
	usernameStr := r.PathParam("username")
	uid, err := dbservice.GetUID(usernameStr)
	if err != nil {
		w.WriteJson("玩家不存在!")
		log.Error(err)
		return
	}
	log.Info("season:", season, " uid:", uid)

	item.Uid = uid
	item.Name = usernameStr

	req := &datadef.CareerData{}
	req, err = srv.careerSrv.QueryCareerData(srv.dbRoundTableName, season, uid)
	if err != nil {
		log.Error(err)
		return
	}

	item.SoloRating = req.SoloRating
	item.DuoRating = req.DuoRating
	item.SquadRating = req.SquadRating

	args := []interface{}{
		"Picture",
		"QQVIP",
		"GameEnter",
	}
	values, valueErr := dbservice.EntityUtil("Player", uid).GetValues(args)
	if valueErr != nil || len(values) != 3 {
		log.Warnf("搜索玩家失败")
		return
	}
	tmpUrl, urlErr := redis.String(values[0], nil)
	if urlErr == nil {
		item.Url = tmpUrl
	}
	tmpQQvip, qqvipErr := redis.Int64(values[1], nil)
	if qqvipErr == nil {
		item.QqVip = uint32(tmpQQvip)
	}
	tmpEnterplat, platformErr := redis.String(values[2], nil)
	if platformErr == nil {
		item.GameEenter = tmpEnterplat
	}

	item.NameColor = common.GetPlayerNameColor(uid)

	srv.searchNumInr(usernameStr, uid, req.TotalRating) //该玩家搜索次数自增
}

// searchNumInr 玩家搜索次数自增
func (srv *Server) searchNumInr(username string, uid uint64, totalrating float32) {
	searchNumData := &datadef.SearchNumData{}
	searchNumData.UID = uid
	searchNumData.UserName = username
	searchNumData.SearchNum = 1
	searchNumData.TotalRating = totalrating

	ok, err := srv.careerSrv.IsSearchNumDatas(srv.dbSearchNum, uid)
	if err != nil {
		log.Error(err)
	} else if ok {
		err = srv.careerSrv.UpdateSearchNumDatas(srv.dbSearchNum, uid, searchNumData)
		if err != nil {
			log.Error(err)
		}
	} else {
		err = srv.careerSrv.InsertSearchNumDatas(srv.dbSearchNum, searchNumData)
		if err != nil {
			log.Error(err)
		}
	}
}

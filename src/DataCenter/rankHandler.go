package main

import (
	"datadef"
	"db"
	"strconv"
	"zeus/dbservice"

	"github.com/ant0ine/go-json-rest/rest"
	log "github.com/cihub/seelog"
	"github.com/garyburd/redigo/redis"
)

// rankHandler 获取全服前Num人排行信息
func (srv *Server) rankHandler(w rest.ResponseWriter, r *rest.Request) {
	friendrank := &datadef.SyncFriendRankList{}
	defer func() {
		w.WriteJson(friendrank)
	}()

	seasonstr := r.PathParam("season")
	season, err := strconv.Atoi(seasonstr)
	if err != nil {
		log.Error(err)
		return
	}
	modelStr := r.PathParam("model")
	ratingStr := r.PathParam("rating")
	str := "Player" + modelStr + ratingStr

	log.Info("season:", season, " str:", str)

	friendrank.Item = make([]*datadef.FriendRankInfo, 0)

	num := 10
	playerRankUtil := db.PlayerRankUtil(str, season)
	topNum, err := playerRankUtil.GetTopNumScore(num)
	if err != nil {
		log.Error(err)
		return
	}

	if len(topNum) != num*2 {
		log.Warnf("全服没有%d人, 只有%d人!", num, len(topNum)/2)
		num = len(topNum) / 2
	}

	for i := 0; i < num*2; i += 2 {
		uid, err := redis.Uint64(topNum[i], nil)
		if err != nil {
			log.Error(err)
			return
		}
		rating, err := redis.Float64(topNum[i+1], nil)
		if err != nil {
			log.Error(err)
			return
		}

		item := &datadef.FriendRankInfo{}
		item.Uid = uid
		item.Name, err = dbservice.Account(uid).GetUsername()
		if err != nil {
			log.Error(err)
			return
		}

		item.SoloRating = float32(rating)
		item.DuoRating = float32(rating)
		item.SquadRating = float32(rating)

		args := []interface{}{
			"Picture",
			"QQVIP",
			"GameEnter",
		}
		values, valueErr := dbservice.EntityUtil("Player", uid).GetValues(args)
		if valueErr != nil || len(values) != 3 {
			continue
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

		friendrank.Item = append(friendrank.Item, item)
	}
}

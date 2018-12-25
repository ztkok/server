package main

import (
	"datadef"
	"db"
	"strconv"
	"zeus/dbservice"

	"common"

	"github.com/ant0ine/go-json-rest/rest"
	log "github.com/cihub/seelog"
	"github.com/garyburd/redigo/redis"
)

// braveRankHandler 获取全服前Num人排行信息
func (srv *Server) braveRankHandler(w rest.ResponseWriter, r *rest.Request) {
	braveRank := &datadef.SyncBraveRankList{}
	defer func() {
		w.WriteJson(braveRank)
	}()

	seasonstr := r.PathParam("season")
	season, err := strconv.Atoi(seasonstr)
	if err != nil {
		log.Error(err)
		return
	}
	braveRank.Item = make([]*datadef.BraveBattleRankInfo, 0)
	num := 1000
	playerRankUtil := db.PlayerRankUtil(common.BraveBattleRank, season)
	topNum, err := playerRankUtil.GetTopNumScore(num)
	if err != nil {
		log.Error(err)
		return
	}
	has := len(topNum) / 2
	if has != num {
		log.Warnf("全服没有%d人, 只有%d人!", num, has)
		num = has
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

		item := &datadef.BraveBattleRankInfo{}
		item.Uid = uid
		item.Brave = uint32(rating)
		item.Rank = uint32(i/2) + 1
		args := []interface{}{
			"Picture",
			"QQVIP",
			"GameEnter",
			"Name",
		}
		values, valueErr := dbservice.EntityUtil("Player", uid).GetValues(args)
		if valueErr != nil || len(values) != 4 {
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
		tmpUserName, userNameErr := redis.String(values[3], nil)
		if userNameErr == nil {
			item.Name = tmpUserName
		}

		item.NameColor = common.GetPlayerNameColor(uid)

		braveRank.Item = append(braveRank.Item, item)
	}
}

// braveUserRankHandler 获取本人的排行
func (srv *Server) braveUserRankHandler(w rest.ResponseWriter, r *rest.Request) {
	userRank := map[string]uint32{
		"Rank":  0,
		"Score": 0,
	}
	defer func() {
		w.WriteJson(userRank)
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
		w.WriteJson("玩家不存在!")
		log.Error(err)
		return
	}
	playerRankUtil := db.PlayerRankUtil(common.BraveBattleRank, season)
	userRank["Rank"], err = playerRankUtil.GetPlayerRank(uid)
	if err != nil {
		log.Error(err)
		return
	}
	if userRank["Rank"] > 1000 {
		userRank["Rank"] = 0
	}
	userRank["Score"], _ = playerRankUtil.GetPlayerScore(uid)
}

//braveFriendRankHandler 好友排行
func (srv *Server) braveFriendRankHandler(w rest.ResponseWriter, r *rest.Request) {
	braveRank := &datadef.SyncBraveRankList{}

	defer func() {
		w.WriteJson(braveRank)
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
		w.WriteJson("玩家不存在!")
		log.Error(err)
		return
	}
	braveRank.Item = make([]*datadef.BraveBattleRankInfo, 0)
	list := make([]*db.FriendInfo, 0)
	list = db.GetFriendUtil(uid).GetFriendList()

	// 游戏内好友
	mapList := make(map[uint64]*db.FriendInfo)
	for _, j := range list {
		if j == nil {
			continue
		}

		mapList[j.ID] = j
	}

	//  平台好友信息
	platInfo := db.GetFriendUtil(uid).GetPlatFrientInfo()
	for _, platid := range platInfo {
		platInfo := &db.FriendInfo{}
		platInfo.ID = platid
		platInfo.Name = "Name"
		mapList[platid] = platInfo
	}

	// 玩家自己信息
	myInfo := &db.FriendInfo{}
	myInfo.ID = uid
	myInfo.Name = "Name"
	mapList[uid] = myInfo

	playerRankUtil := db.PlayerRankUtil(common.BraveBattleRank, season)
	for _, v := range mapList {
		score, err := playerRankUtil.GetPlayerScore(v.ID)
		if err != nil {
			log.Error(err)
			continue
		}
		if score == 0 {
			continue
		}
		//rank, err := playerRankUtil.GetPlayerRank(v.ID)
		//if err != nil {
		//	log.Error(err)
		//	continue
		//}
		item := &datadef.BraveBattleRankInfo{}
		item.Uid = v.ID
		item.Brave = score
		//item.Rank = rank
		args := []interface{}{
			"Picture",
			"QQVIP",
			"GameEnter",
			"Name",
		}
		values, valueErr := dbservice.EntityUtil("Player", v.ID).GetValues(args)
		if valueErr != nil || len(values) != 4 {
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

		tmpUserName, userNameErr := redis.String(values[3], nil)
		if userNameErr == nil {
			item.Name = tmpUserName
		}
		braveRank.Item = append(braveRank.Item, item)
	}
}

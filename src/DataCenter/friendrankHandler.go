package main

import (
	"common"
	"datadef"
	"db"
	"strconv"
	"zeus/dbservice"

	"github.com/ant0ine/go-json-rest/rest"
	log "github.com/cihub/seelog"
	"github.com/garyburd/redigo/redis"
)

// friendrankHandler 获取好友排行信息
func (srv *Server) friendrankHandler(w rest.ResponseWriter, r *rest.Request) {
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
	uidStr := r.PathParam("uid")
	uid, err := strconv.ParseUint(uidStr, 10, 64)
	if err != nil {
		log.Error(err)
		return
	}
	log.Info("dbRoundTableName:", srv.dbRoundTableName, " season:", season, " uid:", uid)

	friendrank.Item = make([]*datadef.FriendRankInfo, 0)

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

	for _, info := range mapList {
		item := &datadef.FriendRankInfo{}
		item.Uid = info.ID
		item.Name = info.Name

		curSeason := common.GetSeason()
		curData := &datadef.CareerBase{}
		curSeasonUtil := db.PlayerCareerDataUtil(common.PlayerCareerTotalData, info.ID, curSeason)
		if err := curSeasonUtil.GetRoundData(curData); err != nil {
			log.Error("GetRoundData err: ", err)
		}

		item.SoloRating = curData.SoloRating
		item.DuoRating = curData.DuoRating
		item.SquadRating = curData.SquadRating

		args := []interface{}{
			"Picture",
			"QQVIP",
			"GameEnter",
			"Name",
			"Level",
		}
		values, valueErr := dbservice.EntityUtil("Player", info.ID).GetValues(args)
		if valueErr != nil || len(values) != len(args) {
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
		tmpUserLevel, levelErr := redis.Uint64(values[4], nil)
		if levelErr == nil {
			item.Level = uint32(tmpUserLevel)
		}

		item.NameColor = common.GetPlayerNameColor(uid)

		friendrank.Item = append(friendrank.Item, item)
	}
}

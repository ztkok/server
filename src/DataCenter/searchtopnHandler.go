package main

import (
	"common"
	"datadef"
	"zeus/dbservice"

	"github.com/ant0ine/go-json-rest/rest"
	log "github.com/cihub/seelog"
	"github.com/garyburd/redigo/redis"
)

// searchtopnHandler 搜索量前n的玩家
func (srv *Server) searchtopnHandler(w rest.ResponseWriter, r *rest.Request) {
	searchrank := &datadef.SyncFriendRankList{}
	defer func() {
		w.WriteJson(searchrank)
	}()

	num := 10
	searchNumDatas := make([]*datadef.SearchNumData, 0)
	searchNumDatas, err := srv.careerSrv.QuerySearchNumData(srv.dbSearchNum, num)
	if err != nil {
		log.Error(err)
		return
	}
	if num != len(searchNumDatas) {
		log.Warnf("搜索人数没有达到%d人, 只有%d人!", num, len(searchNumDatas))
		num = len(searchNumDatas)
	}

	searchrank.Item = make([]*datadef.FriendRankInfo, 0)

	for i := 0; i < num; i++ {
		item := &datadef.FriendRankInfo{}
		item.Uid = searchNumDatas[i].UID
		item.Name = searchNumDatas[i].UserName

		item.SoloRating = searchNumDatas[i].TotalRating
		item.DuoRating = searchNumDatas[i].TotalRating
		item.SquadRating = searchNumDatas[i].TotalRating

		args := []interface{}{
			"Picture",
			"QQVIP",
			"GameEnter",
		}
		values, valueErr := dbservice.EntityUtil("Player", item.Uid).GetValues(args)
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

		item.NameColor = common.GetPlayerNameColor(item.Uid)

		searchrank.Item = append(searchrank.Item, item)
	}
}

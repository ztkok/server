package main

import (
	"datadef"
	"db"
	"zeus/dbservice"

	"common"

	"github.com/ant0ine/go-json-rest/rest"
	log "github.com/cihub/seelog"
	"github.com/garyburd/redigo/redis"
)

const (
	SeasonRankTopNum = 1000
)

// seasonWinsRankHandler 获取全服赛季获胜数排行榜
func (srv *Server) seasonWinsRankHandler(w rest.ResponseWriter, r *rest.Request) {
	winsRank := &datadef.SeasonWinsRank{}
	rankSeason := common.GetRankSeason()

	if rankSeason == 0 {
		rankSeason = common.GetLastRankSeason()
		log.Info("GetLastRankSeason: ", rankSeason)
		if rankSeason == 0 {
			return
		}
	}

	defer func() {
		winsRank.Season = rankSeason
		w.WriteJson(winsRank)
	}()

	reqUid := common.StringToUint64(r.PathParam("uid"))
	if reqUid == 0 {
		log.Error("Invalid uid")
		return
	}

	matchTyp := common.StringToUint8(r.PathParam("matchtyp"))
	matchTypStrs := map[uint8]string{
		0: common.TotalRank,
		1: common.SoloRank,
		2: common.DuoRank,
		4: common.SquadRank,
	}

	matchTypStr, ok := matchTypStrs[matchTyp]
	if !ok {
		log.Error("Invalid match typ: ", matchTyp)
		return
	}

	beg := common.StringToUint64(r.PathParam("beg"))
	end := common.StringToUint64(r.PathParam("end"))
	if beg == 0 || end == 0 {
		log.Error("Invalid beg/end")
		return
	}

	season := common.RankSeasonToSeason(rankSeason)
	util := db.PlayerRankUtil(matchTypStr+common.SeasonRankTypStrWins, season)

	scores, err := util.GetScoreByRankRange(beg, end)
	if err != nil {
		log.Error("GetScoreByRankRange err: ", err)
		return
	}

	for i := 0; i < len(scores); i += 2 {
		uid, err := redis.Uint64(scores[i], nil)
		if err != nil {
			log.Error("redis Uint64 err: ", err)
			return
		}

		wins, err := redis.Float64(scores[i+1], nil)
		if err != nil {
			log.Error("redis Float64 err: ", err)
			return
		}

		winsRank.Infos = append(winsRank.Infos, ToPlayerWinsRankInfo(season, matchTyp, uid, uint32(wins), uint32(beg)+uint32(i/2)))
	}

	rank, err := util.GetPlayerRank(reqUid)
	if err != nil {
		log.Error("GetPlayerRank err: ", err)
		return
	}

	wins, err := util.GetPlayerScore(reqUid)
	if err != nil {
		log.Error("GetPlayerScore err: ", err)
		return
	}

	winsRank.ReqInfo = ToPlayerWinsRankInfo(season, matchTyp, reqUid, uint32(wins), rank)

	db.PlayerInfoUtil(reqUid).SetSeasonBeginEnter(rankSeason)
	log.Info("Season wins rank, rankSeason: ", rankSeason, " uid: ", reqUid, " matchTyp: ", matchTyp, " beg: ", beg, " end: ", end, " len(winsRank.Infos): ", len(winsRank.Infos))
}

func ToPlayerWinsRankInfo(season int, matchTyp uint8, uid uint64, wins uint32, rank uint32) *datadef.PlayerWinsRankInfo {
	info := &datadef.PlayerWinsRankInfo{
		Uid:  uid,
		Wins: wins,
		Rank: rank,
	}

	args := []interface{}{
		"Picture",
		"QQVIP",
		"GameEnter",
		"Name",
	}

	values, valueErr := dbservice.EntityUtil("Player", uid).GetValues(args)
	if valueErr != nil || len(values) != 4 {
		return nil
	}

	tmpUrl, urlErr := redis.String(values[0], nil)
	if urlErr == nil {
		info.Url = tmpUrl
	}

	tmpQQvip, qqvipErr := redis.Int64(values[1], nil)
	if qqvipErr == nil {
		info.QqVip = uint32(tmpQQvip)
	}

	tmpEnterplat, platformErr := redis.String(values[2], nil)
	if platformErr == nil {
		info.GameEenter = tmpEnterplat
	}

	tmpUserName, userNameErr := redis.String(values[3], nil)
	if userNameErr == nil {
		info.Name = tmpUserName
	}

	info.NameColor = common.GetPlayerNameColor(uid)

	data, err := common.GetPlayerCareerData(uid, season, matchTyp)
	if err != nil {
		log.Error("GetPlayerCareerData err: ", err, " matchTyp: ", matchTyp)
		return nil
	}

	info.TotalBattleNum = data.TotalBattleNum
	info.TopTenNum = data.TopTenNum

	if rank == 0 && info.TotalBattleNum >= 3 {
		util := db.PlayerRankUtil(common.TotalRank+common.SeasonRankTypStrWins, season)

		if _, err := util.PipeRanksUint64(float64(data.FirstNum), uid); err != nil {
			log.Error("PipeRanks err: ", err)
		}

		info.Rank, _ = util.GetPlayerRank(uid)
		info.Wins = data.FirstNum
		util.RemRankByUid(uid)
	}

	return info
}

// seasonKillsRankHandler 获取全服赛季击败数排行榜
func (srv *Server) seasonKillsRankHandler(w rest.ResponseWriter, r *rest.Request) {
	killsRank := &datadef.SeasonKillsRank{}
	rankSeason := common.GetRankSeason()

	if rankSeason == 0 {
		rankSeason = common.GetLastRankSeason()
		log.Info("GetLastRankSeason: ", rankSeason)
		if rankSeason == 0 {
			return
		}
	}

	defer func() {
		killsRank.Season = rankSeason
		w.WriteJson(killsRank)
	}()

	reqUid := common.StringToUint64(r.PathParam("uid"))
	if reqUid == 0 {
		log.Error("Invalid uid")
		return
	}

	matchTyp := common.StringToUint8(r.PathParam("matchtyp"))
	matchTypStrs := map[uint8]string{
		0: common.TotalRank,
		1: common.SoloRank,
		2: common.DuoRank,
		4: common.SquadRank,
	}

	matchTypStr, ok := matchTypStrs[matchTyp]
	if !ok {
		log.Error("Invalid match typ: ", matchTyp)
		return
	}

	beg := common.StringToUint64(r.PathParam("beg"))
	end := common.StringToUint64(r.PathParam("end"))
	if beg == 0 || end == 0 {
		log.Error("Invalid beg/end")
		return
	}

	season := common.RankSeasonToSeason(rankSeason)
	util := db.PlayerRankUtil(matchTypStr+common.SeasonRankTypStrKills, season)

	scores, err := util.GetScoreByRankRange(beg, end)
	if err != nil {
		log.Error("GetScoreByRankRange err: ", err)
		return
	}

	for i := 0; i < len(scores); i += 2 {
		uid, err := redis.Uint64(scores[i], nil)
		if err != nil {
			log.Error("redis Uint64 err: ", err)
			return
		}

		kills, err := redis.Float64(scores[i+1], nil)
		if err != nil {
			log.Error("redis Float64 err: ", err)
			return
		}

		killsRank.Infos = append(killsRank.Infos, ToPlayerKillsRankInfo(season, matchTyp, uid, uint32(kills), uint32(beg)+uint32(i/2)))
	}

	rank, err := util.GetPlayerRank(reqUid)
	if err != nil {
		log.Error("GetPlayerRank err: ", err)
		return
	}

	kills, err := util.GetPlayerScore(reqUid)
	if err != nil {
		log.Error("GetPlayerScore err: ", err)
		return
	}

	killsRank.ReqInfo = ToPlayerKillsRankInfo(season, matchTyp, reqUid, uint32(kills), rank)

	db.PlayerInfoUtil(reqUid).SetSeasonBeginEnter(rankSeason)
	log.Info("Season kills rank, rankSeason: ", rankSeason, " uid: ", reqUid, " matchTyp: ", matchTyp, " beg: ", beg, " end: ", end, " len(killsRank.Infos): ", len(killsRank.Infos))
}

func ToPlayerKillsRankInfo(season int, matchTyp uint8, uid uint64, kills uint32, rank uint32) *datadef.PlayerKillsRankInfo {
	info := &datadef.PlayerKillsRankInfo{
		Uid:   uid,
		Kills: kills,
		Rank:  rank,
	}

	args := []interface{}{
		"Picture",
		"QQVIP",
		"GameEnter",
		"Name",
	}

	values, valueErr := dbservice.EntityUtil("Player", uid).GetValues(args)
	if valueErr != nil || len(values) != 4 {
		return nil
	}

	tmpUrl, urlErr := redis.String(values[0], nil)
	if urlErr == nil {
		info.Url = tmpUrl
	}

	tmpQQvip, qqvipErr := redis.Int64(values[1], nil)
	if qqvipErr == nil {
		info.QqVip = uint32(tmpQQvip)
	}

	tmpEnterplat, platformErr := redis.String(values[2], nil)
	if platformErr == nil {
		info.GameEenter = tmpEnterplat
	}

	tmpUserName, userNameErr := redis.String(values[3], nil)
	if userNameErr == nil {
		info.Name = tmpUserName
	}

	info.NameColor = common.GetPlayerNameColor(uid)

	data, err := common.GetPlayerCareerData(uid, season, matchTyp)
	if err != nil {
		log.Error("GetPlayerCareerData err: ", err, " matchTyp: ", matchTyp)
		return nil
	}

	info.TotalBattleNum = data.TotalBattleNum
	info.MaxKills = data.SingleMaxKill

	if rank == 0 && info.TotalBattleNum >= 3 {
		util := db.PlayerRankUtil(common.TotalRank+common.SeasonRankTypStrKills, season)

		if _, err := util.PipeRanksUint64(float64(data.TotalKillNum), uid); err != nil {
			log.Error("PipeRanks err: ", err)
		}

		info.Rank, _ = util.GetPlayerRank(uid)
		info.Kills = data.TotalKillNum
		util.RemRankByUid(uid)
	}

	return info
}

// seasonRatingRankHandler 获取全服赛季积分排行榜
func (srv *Server) seasonRatingRankHandler(w rest.ResponseWriter, r *rest.Request) {
	ratingRank := &datadef.SeasonRatingRank{}
	rankSeason := common.GetRankSeason()

	if rankSeason == 0 {
		rankSeason = common.GetLastRankSeason()
		log.Info("GetLastRankSeason: ", rankSeason)
		if rankSeason == 0 {
			return
		}
	}

	defer func() {
		ratingRank.Season = rankSeason
		w.WriteJson(ratingRank)
	}()

	reqUid := common.StringToUint64(r.PathParam("uid"))
	if reqUid == 0 {
		log.Error("Invalid uid")
		return
	}

	matchTyp := common.StringToUint8(r.PathParam("matchtyp"))
	matchTypStrs := map[uint8]string{
		0: common.TotalRank,
		1: common.SoloRank,
		2: common.DuoRank,
		4: common.SquadRank,
	}

	matchTypStr, ok := matchTypStrs[matchTyp]
	if !ok {
		log.Error("Invalid match typ: ", matchTyp)
		return
	}

	beg := common.StringToUint64(r.PathParam("beg"))
	end := common.StringToUint64(r.PathParam("end"))
	if beg == 0 || end == 0 {
		log.Error("Invalid beg/end")
		return
	}

	season := common.RankSeasonToSeason(rankSeason)
	util := db.PlayerRankUtil(matchTypStr+common.SeasonRankTypStrRating, season)

	scores, err := util.GetScoreByRankRange(beg, end)
	if err != nil {
		log.Error("GetScoreByRankRange err: ", err)
		return
	}

	for i := 0; i < len(scores); i += 2 {
		uid, err := redis.Uint64(scores[i], nil)
		if err != nil {
			log.Error("redis Uint64 err: ", err)
			return
		}

		rating, err := redis.Float64(scores[i+1], nil)
		if err != nil {
			log.Error("redis Float64 err: ", err)
			return
		}

		ratingRank.Infos = append(ratingRank.Infos, ToPlayerRatingRankInfo(season, matchTyp, uid, uint32(rating), uint32(beg)+uint32(i/2)))
	}

	rank, err := util.GetPlayerRank(reqUid)
	if err != nil {
		log.Error("GetPlayerRank err: ", err)
		return
	}

	rating, err := util.GetPlayerScore(reqUid)
	if err != nil {
		log.Error("GetPlayerScore err: ", err)
		return
	}

	ratingRank.ReqInfo = ToPlayerRatingRankInfo(season, matchTyp, reqUid, uint32(rating), rank)

	db.PlayerInfoUtil(reqUid).SetSeasonBeginEnter(rankSeason)
	log.Info("Season rating rank, rankSeason: ", rankSeason, " uid: ", reqUid, " matchTyp: ", matchTyp, " beg: ", beg, " end: ", end, " len(ratingRank.Infos): ", len(ratingRank.Infos))
}

func ToPlayerRatingRankInfo(season int, matchTyp uint8, uid uint64, rating uint32, rank uint32) *datadef.PlayerRatingRankInfo {
	info := &datadef.PlayerRatingRankInfo{
		Uid:    uid,
		Rating: rating,
		Rank:   rank,
	}

	args := []interface{}{
		"Picture",
		"QQVIP",
		"GameEnter",
		"Name",
	}

	values, valueErr := dbservice.EntityUtil("Player", uid).GetValues(args)
	if valueErr != nil || len(values) != 4 {
		return nil
	}

	tmpUrl, urlErr := redis.String(values[0], nil)
	if urlErr == nil {
		info.Url = tmpUrl
	}

	tmpQQvip, qqvipErr := redis.Int64(values[1], nil)
	if qqvipErr == nil {
		info.QqVip = uint32(tmpQQvip)
	}

	tmpEnterplat, platformErr := redis.String(values[2], nil)
	if platformErr == nil {
		info.GameEenter = tmpEnterplat
	}

	tmpUserName, userNameErr := redis.String(values[3], nil)
	if userNameErr == nil {
		info.Name = tmpUserName
	}

	info.NameColor = common.GetPlayerNameColor(uid)

	data, err := common.GetPlayerCareerData(uid, season, matchTyp)
	if err != nil {
		log.Error("GetPlayerCareerData err: ", err, " matchTyp: ", matchTyp)
		return nil
	}

	info.TotalBattleNum = data.TotalBattleNum

	die := data.TotalBattleNum - data.FirstNum
	if die == 0 {
		die = 1
	}
	info.KDA = float32(data.TotalKillNum) / float32(die)

	return info
}

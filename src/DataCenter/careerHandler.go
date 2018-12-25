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

// careerHandler 获取生涯数据
func (srv *Server) careerHandler(w rest.ResponseWriter, r *rest.Request) {
	careerdata := &datadef.CareerData{}
	defer func() {
		w.WriteJson(careerdata)
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

	//判断是否已经存入所有赛季的总数据Season=0的PlayerCareerData表
	seasonTotalUtil := db.PlayerCareerDataUtil(common.PlayerCareerTotalData, uid, 0)
	if ok := seasonTotalUtil.IsKeyExist(); !ok {
		seasonTwoData := &datadef.CareerBase{}
		seasonTwoUtil := db.PlayerCareerDataUtil(common.PlayerCareerTotalData, uid, 2)
		if err := seasonTwoUtil.GetRoundData(seasonTwoData); err != nil {
			log.Error("GetRoundData err: ", err)
		}

		if err := seasonTotalUtil.SetRoundData(seasonTwoData); err != nil {
			log.Error("SetRoundData err: ", err)
		}
	}

	//临时解决赛季切换小黑盒数据重新开始记录问题
	season = 0

	data := &datadef.CareerBase{}
	seasonUtil := db.PlayerCareerDataUtil(common.PlayerCareerTotalData, uid, season)
	if err := seasonUtil.GetRoundData(data); err != nil {
		log.Error("GetRoundData err: ", err)
	}
	careerdata.Uid = uid
	careerdata.TotalBattleNum = data.TotalBattleNum
	careerdata.TotalFirstNum = data.FirstNum
	careerdata.TotalTopTenNum = data.TopTenNum
	careerdata.TotalKillNum = data.TotalKillNum
	careerdata.TotalHeadShot = data.TotalHeadShot
	careerdata.TotalShotNum = data.Totalshotnum
	careerdata.TotalEffectHarm = data.TotalEffectHarm
	careerdata.TotalSurviveTime = data.SurviveTime
	careerdata.TotalDistance = data.TotalDistance

	curData := &datadef.CareerBase{}
	curSeason := common.GetSeason()
	curSeasonUtil := db.PlayerCareerDataUtil(common.PlayerCareerTotalData, uid, curSeason)
	if err := curSeasonUtil.GetRoundData(curData); err != nil {
		log.Error("GetRoundData err: ", err)
	}
	careerdata.SoloRating = curData.SoloRating
	careerdata.DuoRating = curData.DuoRating
	careerdata.SquadRating = curData.SquadRating
	careerdata.TotalRating = curData.TotalRating
	playerSoloRankUtil := db.PlayerRankUtil(common.SoloRank, curSeason)
	careerdata.SoloRank, err = playerSoloRankUtil.GetPlayerRank(uid)
	if err != nil {
		log.Error(err)
	}
	playerDuoRankUtil := db.PlayerRankUtil(common.DuoRank, curSeason)
	careerdata.DuoRank, err = playerDuoRankUtil.GetPlayerRank(uid)
	if err != nil {
		log.Error(err)
	}
	playerSquadRankUtil := db.PlayerRankUtil(common.SquadRank, curSeason)
	careerdata.SquadRank, err = playerSquadRankUtil.GetPlayerRank(uid)
	if err != nil {
		log.Error(err)
	}
	playerTotalRankUtil := db.PlayerRankUtil(common.TotalRank, curSeason)
	careerdata.TotalRank, err = playerTotalRankUtil.GetPlayerRank(uid)
	if err != nil {
		log.Error(err)
	}
	careerdata.TopRating = common.GetTopRating(curSeason)

	args := []interface{}{
		"Picture",
		"QQVIP",
		"GameEnter",
		"Name",
		"Gender",
		"Level",
		"Exp",
	}
	values, valueErr := dbservice.EntityUtil("Player", uid).GetValues(args)
	if valueErr != nil || len(values) != 7 {
		log.Error("获取url、qqvip错误")
		return
	}
	tmpUrl, urlErr := redis.String(values[0], nil)
	if urlErr == nil {
		careerdata.Url = tmpUrl
	}
	tmpQQvip, qqvipErr := redis.Int64(values[1], nil)
	if qqvipErr == nil {
		careerdata.QqVip = uint32(tmpQQvip)
	}
	tmpEnterplat, platformErr := redis.String(values[2], nil)
	if platformErr == nil {
		careerdata.GameEenter = tmpEnterplat
	}
	tmpUserName, userNameErr := redis.String(values[3], nil)
	if userNameErr == nil {
		careerdata.UserName = tmpUserName
	}
	gender, genderErr := redis.String(values[4], nil)
	if genderErr == nil {
		if gender == "男" {
			careerdata.Gender = 1
		} else if gender == "女" {
			careerdata.Gender = 2
		}
	}
	level, levelErr := redis.Uint64(values[5], nil)
	if levelErr == nil {
		careerdata.Level = uint32(level)
	}
	exp, expErr := redis.Uint64(values[6], nil)
	if expErr == nil {
		careerdata.Exp = uint32(exp)
	}
	careerdata.MaxExp = common.GetMaxExpByLevel(uint32(level + 1))

	careerdata.NameColor = common.GetPlayerNameColor(uid)

	log.Info("careerdata:", careerdata)
}

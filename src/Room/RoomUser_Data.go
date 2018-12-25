package main

import (
	"common"
	"datadef"
	"db"
	"encoding/json"
	"excel"
	"fmt"
	"io/ioutil"
	"math"
	"msdk"
	"net/http"
	"protoMsg"
	"strconv"
	"strings"
	"time"
	"zeus/iserver"
)

// getMaxItemBonus 获取同种类型道具的最大加成
func (user *RoomUser) getMaxItemBonus(typ uint32) uint32 {
	var max uint32
	util := db.PlayerGoodsUtil(user.GetDBID())

	for _, v := range util.GetAllGoodsInfo() {
		base, ok := excel.GetItem(uint64(v.Id))
		if !ok || base.Type != uint64(typ) {
			continue
		}

		add := common.StringToUint32(base.AddValue)
		if add > max {
			max = add
		}
	}

	return max
}

// limitAddCoin 限制每日金币上限
func (user *RoomUser) limitAddCoin(coin uint32) uint32 {
	limit := uint64(common.GetTBSystemValue(common.System_DayCoinLimit))
	limit += uint64(user.getMaxItemBonus(ItemTypeAddCoinLimit))

	now := time.Now()
	lastget := time.Unix(int64(user.GetDayCoinGetTime()), int64(0))
	dayget := user.GetDayCoinGet()
	user.Debug("limit: ", limit, "lastget: ", lastget, " dayget: ", dayget)

	if now.Year() == lastget.Year() && now.YearDay() == lastget.YearDay() {
		if dayget >= limit {
			user.Info("Reach day limit, limit: ", limit, " get: ", dayget)
			return 0
		}

		if dayget+uint64(coin) >= limit {
			user.SetDayCoinGet(limit)
			return uint32(limit - dayget)
		}

		user.SetDayCoinGet(dayget + uint64(coin))
		return coin
	}

	user.SetDayCoinGetTime(uint64(now.Unix()))
	if uint64(coin) >= limit {
		user.SetDayCoinGet(limit)
		return uint32(limit)
	}

	user.SetDayCoinGet(uint64(coin))
	user.Info("Today get coins: ", coin)
	return coin
}

// limitAddBraveCoin 限制每日勇气值上限
func (user *RoomUser) limitAddBraveCoin(brave uint32) uint32 {
	limit := uint64(common.GetTBSystemValue(common.System_DayBraveLimit))
	limit += uint64(user.getMaxItemBonus(ItemTypeAddBraveLimit))

	now := time.Now()
	lastget := time.Unix(int64(user.GetDayBraveGetTime()), int64(0))
	dayget := user.GetDayBraveGet()
	user.Debug("limit: ", limit, "lastget: ", lastget, " dayget: ", dayget)

	if now.Year() == lastget.Year() && now.YearDay() == lastget.YearDay() {
		if dayget >= limit {
			user.Info("Reach day limit, limit: ", limit, " get: ", dayget)
			return 0
		}

		if dayget+uint64(brave) >= limit {
			user.SetDayBraveGet(limit)
			return uint32(limit - dayget)
		}

		user.SetDayBraveGet(dayget + uint64(brave))
		return brave
	}

	user.SetDayBraveGetTime(uint64(now.Unix()))
	if uint64(brave) >= limit {
		user.SetDayBraveGet(limit)
		return uint32(limit)
	}

	user.SetDayBraveGet(uint64(brave))
	user.Info("Today get braves: ", brave)
	return brave
}

// Settle 单局结束玩家结算数据
func (user *RoomUser) Settle() {
	scene := user.GetSpace().(*Scene)
	if scene == nil {
		return
	}

	if user.stateMgr.isSettle {
		return
	}

	if user.sumData.isSettble {
		user.writeDataReds()
		user.sumData.isSettble = false
	}

	ret := &protoMsg.MapCharacterResultNotify{}

	ret.Kill = user.kill
	ret.Timeinseconds = uint32(user.sumData.endgametime - user.sumData.begingametime)
	ret.Rank = user.sumData.rank
	ret.Round = 1
	ret.Headshotnum = user.headshotnum
	ret.Shotnum = user.sumData.shotnum
	ret.Effectharm = user.effectharm
	ret.Recovernum = user.sumData.recvitemusenum
	ret.Revivenum = user.sumData.revivenum
	ret.Killdistance = user.sumData.killdistance
	ret.Killstmnum = user.sumData.killstmnum
	ret.Attacknum = user.sumData.attacknum
	ret.Speednum = user.sumData.speednum
	ret.Rundistance = user.sumData.runDistance / 1000
	ret.Killscore = user.sumData.killRating
	ret.Rankscore = user.sumData.winRating
	ret.Totalscore = user.sumData.careerdata.TotalRating
	ret.Coin = user.limitAddCoin(user.sumData.coin)
	ret.BraveCoin = fmt.Sprintf("%d", user.limitAddBraveCoin(user.sumData.braveCoin))
	ret.Matchmode = scene.GetMatchMode()

	if scene.teamMgr.isTeam {
		ret.Amount = scene.teamMgr.TeamsSum()
	} else {
		ret.Amount = scene.maxUserSum
	}

	if ret.Rank > ret.Amount {
		ret.Rank = ret.Amount
	}

	if scene.GetMatchMode() == common.MatchModeBrave {
		user.BraveSettle()
	}

	user.Info("Balance info, ret:", ret)

	user.RPC(iserver.ServerTypeClient, "MapCharacterResultNotify", ret)
	user.BroadcastToWatchers(RoomUserTypeWatcher, "MapCharacterResultNotify", ret)
	/*
		user.CastMsgToMe(ret)
	*/
	user.stateMgr.isSettle = true

	if ret.Coin > 0 {
		user.RPC(common.ServerTypeLobby, "AddCoin", ret.Coin)
	}

	if user.sumData.braveCoin > 0 {
		user.RPC(common.ServerTypeLobby, "AddBraveCoin", common.StringToUint32(ret.BraveCoin))
	}

	user.updateMilitaryRank() //发放玩家在本局比赛中获得的经验值，更新玩家的军衔
	// user.Info("endgatetime:", user.sumData.endgametime, " survivetime:", ret.Timeinseconds, " killrating:", ret.Killscore,
	// " winrating:", ret.Rankscore, " totalscore:", ret.Totalscore, " cardestoryNum:", user.sumData.carDestoryNum)
}

// BraveSettle 勇者模式结算
func (user *RoomUser) BraveSettle() {
	scene := user.GetSpace().(*Scene)
	if scene == nil {
		return
	}

	brave, ok := excel.GetMatchmode(uint64(scene.uniqueId))
	if ok && user.sumData.braveCoin > 0 {
		d := strings.Replace(brave.Seasonopentime, "|", "", -1)
		if len(d) >= 8 {
			season, _ := strconv.Atoi(d[:8])
			score, _ := db.PlayerRankUtil(common.BraveBattleRank, season).GetPlayerScore(user.GetDBID())
			t := float64(user.sumData.braveCoin+score) + 1 - float64(time.Now().UnixNano()/1000%10000000000000)/float64(10000000000000)
			db.PlayerRankUtil(common.BraveBattleRank, season).PipeRanksUint64(t, user.GetDBID())
		}
	}
}

// updateBraveRecord 更新勇者战场使用同一张入场券的连胜场次
func (user *RoomUser) updateBraveRecord() {
	scene := user.GetSpace().(*Scene)
	if scene == nil {
		return
	}

	info, err := db.PlayerInfoUtil(user.GetDBID()).GetBraveRecord()
	if err != nil || info == nil {
		return
	}

	brave, ok := excel.GetMatchmode(uint64(info.UniqueId))
	if !ok {
		return
	}

	var win bool

	if !scene.teamMgr.isTeam {
		if user.sumData.rank <= uint32(brave.Solorank) {
			win = true
		}
	} else {
		if scene.teamMgr.teamType == 0 {
			if user.sumData.rank <= uint32(brave.Duorank) {
				win = true
			}
		} else if scene.teamMgr.teamType == 1 {
			if user.sumData.rank <= uint32(brave.Squadrank) {
				win = true
			}
		}
	}

	if win {
		db.PlayerInfoUtil(user.GetDBID()).AddBraveWinStreak(1)
	} else {
		db.PlayerInfoUtil(user.GetDBID()).ResetBraveWinStreak()
		db.PlayerGoodsUtil(user.GetDBID()).DelGoods(common.Item_BraveTicket)
	}
}

// updateCareerData 生涯数据统计录入redis
func (user *RoomUser) updateCareerData(season int) {
	user.DoCareerData(season, 0)

	switch user.sumData.battletype {
	case 0:
		user.DoCareerData(season, 1)
	case 1:
		user.DoCareerData(season, 2)
	case 2:
		user.DoCareerData(season, 4)
	default:
		user.Error("Unkonwn battle typ: ", user.sumData.battletype)
	}
}

func (user *RoomUser) DoCareerData(season int, typ uint8) {
	var (
		playerCareerDataTypStrs = map[uint8]string{
			0: common.PlayerCareerTotalData,
			1: common.PlayerCareerSoloData,
			2: common.PlayerCareerDuoData,
			4: common.PlayerCareerSquadData,
		}
	)

	if _, ok := playerCareerDataTypStrs[typ]; !ok {
		user.Error("Invalid playerCareerData typ: ", typ)
		return
	}

	data := &datadef.CareerBase{}
	util := db.PlayerCareerDataUtil(playerCareerDataTypStrs[typ], user.GetDBID(), season)
	if err := util.GetRoundData(data); err != nil {
		user.Error("GetRoundData err: ", err, " typ: ", typ)
		return
	}

	user.UpdateCareerDataAndRank(data, season, typ)

	if err := util.SetRoundData(data); err != nil {
		user.Error("SetRoundData err: ", err, " typ: ", typ)
		return
	}

	if typ == 0 && season != 0 {
		user.sumData.careerdata = data
	}
}

// 更新玩家的生涯数据
func (user *RoomUser) UpdateCareerDataAndRank(data *datadef.CareerBase, season int, typ uint8) {
	var (
		win, kill bool
	)

	data.UID = user.GetDBID()
	data.UserName = user.GetName()
	data.TotalBattleNum++

	if user.sumData.rank == 1 {
		data.FirstNum++
		data.FirstStamp = time.Now().UnixNano()
		win = true
	}

	if user.sumData.rank <= 10 {
		data.TopTenNum++
	}

	if user.kill > 0 {
		data.TotalKillNum += user.kill
		data.KillStamp = time.Now().UnixNano()
		kill = true
	}

	data.TotalHeadShot += user.headshotnum
	data.Totalshotnum += user.sumData.shotnum
	data.TotalEffectHarm += user.effectharm
	data.SurviveTime += user.sumData.endgametime - user.sumData.begingametime
	data.TotalDistance += user.sumData.runDistance / 1000

	if user.kill > data.SingleMaxKill {
		data.SingleMaxKill = user.kill
	}

	if user.headshotnum > data.SingleMaxHeadShot {
		data.SingleMaxHeadShot = user.headshotnum
	}

	data.RecvItemUseNum += user.sumData.recvitemusenum
	data.CarUseNum += user.sumData.carusernum
	data.CarDestroyNum += user.sumData.carDestoryNum
	data.Coin += user.sumData.coin
	data.TotalCarDistance += user.sumData.carDistance / 1000

	var (
		matchTypStrs = map[uint8]string{
			0: common.TotalRank,
			1: common.SoloRank,
			2: common.DuoRank,
			4: common.SquadRank,
		}

		rankTypStrs = map[uint8]string{
			1: common.SeasonRankTypStrWins,
			2: common.SeasonRankTypStrKills,
			3: common.SeasonRankTypStrRating,
		}
	)

	switch typ {
	case 0:
		data.SoloRating = user.sumData.careerdata.SoloRating
		data.DuoRating = user.sumData.careerdata.DuoRating
		data.SquadRating = user.sumData.careerdata.SquadRating
		data.TotalRating = user.sumData.totalScore(data.SoloRating, data.DuoRating, data.SquadRating)
		data.TotalRank = user.setPipeRank(common.TotalRank, season, math.Ceil(float64(data.TotalRating)), data.UID)
		data.TopRating = common.GetTopRating(common.GetSeason())

		if data.TotalTopRating < data.TotalRating {
			data.TotalTopRating = data.TotalRating
		}

		if data.TotalTopRank == 0 || data.TotalTopRank > data.TotalRank {
			data.TotalTopRank = data.TotalRank
		}

		if data.TotalTopSoloRating < data.SoloRating {
			data.TotalTopSoloRating = data.SoloRating
		}

		if data.TotalTopSoloRank == 0 || data.TotalTopSoloRank > data.SoloRank {
			data.TotalTopSoloRank = data.SoloRank
		}

		if data.TotalTopDuoRating < data.DuoRating {
			data.TotalTopDuoRating = data.DuoRating
		}

		if data.TotalTopDuoRank == 0 || data.TotalTopDuoRank > data.DuoRank {
			data.TotalTopDuoRank = data.DuoRank
		}

		if data.TotalTopSquadRaing < data.SquadRating {
			data.TotalTopSquadRaing = data.SquadRating
		}

		if data.TotalTopSquadRank == 0 || data.TotalTopSquadRank > data.SquadRank {
			data.TotalTopSquadRank = data.SquadRank
		}
	case 1, 2, 4:
		if data.TotalBattleNum >= 3 {
			totalData := &datadef.CareerBase{}
			util := db.PlayerCareerDataUtil(common.PlayerCareerTotalData, user.GetDBID(), season)

			if err := util.GetRoundData(totalData); err != nil {
				user.Error("GetRoundData err: ", err, " typ: ", typ)
				return
			}

			f := 1 - float64(time.Now().UnixNano()/1000%10000000000000)/float64(10000000000000)
			f /= float64(10000)

			if typ == 1 {
				totalData.SoloRank = user.setPipeRank(common.SoloRank, season, math.Ceil(float64(totalData.SoloRating))+f, data.UID)
			} else if typ == 2 {
				totalData.DuoRank = user.setPipeRank(common.DuoRank, season, math.Ceil(float64(totalData.DuoRating))+f, data.UID)
			} else if typ == 4 {
				totalData.SquadRank = user.setPipeRank(common.SquadRank, season, math.Ceil(float64(totalData.SquadRating))+f, data.UID)
			}

			if err := util.SetRoundData(totalData); err != nil {
				user.Error("SetRoundData err: ", err, " typ: ", typ)
				return
			}
		}
	}

	if data.TotalBattleNum == 3 || (data.TotalBattleNum > 3 && win) {
		f := float64(0)
		if data.FirstStamp != 0 {
			f = 1 - float64(data.FirstStamp/1000%10000000000000)/float64(10000000000000)
		}

		if _, err := db.PlayerRankUtil(matchTypStrs[typ]+rankTypStrs[1], season).PipeRanksUint64(float64(data.FirstNum)+f, user.GetDBID()); err != nil {
			user.Error("PipeRanks err: ", err, " typ: ", typ)
			return
		}
	}

	if data.TotalBattleNum == 3 || (data.TotalBattleNum > 3 && kill) {
		f := float64(0)
		if data.KillStamp != 0 {
			f = 1 - float64(data.KillStamp/1000%10000000000000)/float64(10000000000000)
		}

		if _, err := db.PlayerRankUtil(matchTypStrs[typ]+rankTypStrs[2], season).PipeRanksUint64(float64(data.TotalKillNum)+f, user.GetDBID()); err != nil {
			user.Error("PipeRanks err: ", err, " typ: ", typ)
			return
		}
	}
}

// DayData 每天数据写入redis
func (user *RoomUser) DayData() {
	// user.Info("DayData 一天数据录入redis")
	season := common.GetSeason()
	daydata := &datadef.DayData{}

	playerDayDataUtil := db.PlayerDayDataUtil(user.GetDBID(), season)
	if err := playerDayDataUtil.GetDayData(daydata); err != nil {
		user.Error("GetDayData err: ", err)
		return
	}
	if daydata.StartTime != time.Now().Format("2006-01-02") {
		daydata.UID = user.GetDBID()
		daydata.UserName = user.GetName()
		daydata.Season = season
		daydata.NowTime = time.Now().Unix()
		daydata.StartTime = time.Now().Format("2006-01-02")
		if user.sumData.rank == 1 {
			daydata.DayFirstNum = 1
		} else {
			daydata.DayFirstNum = 0
		}
		if user.sumData.rank <= 10 {
			daydata.DayTopTenNum = 1
		} else {
			daydata.DayTopTenNum = 0
		}
		daydata.DayEffectHarm = user.effectharm
		daydata.DayShotNum = user.sumData.shotnum
		daydata.DaySurviveTime = user.sumData.endgametime - user.sumData.begingametime
		daydata.DayDistance = user.sumData.runDistance / 1000
		daydata.DayCarDistance = user.sumData.carDistance / 1000
		daydata.DayAttackNum = user.sumData.attacknum
		daydata.DayRecoverNum = user.sumData.recvitemusenum
		daydata.DayRevivenum = user.sumData.revivenum
		daydata.DayHeadShotNum = user.headshotnum
		daydata.DayBattleNum = 1
		daydata.DayKillNum = user.kill
		daydata.DayCoin = user.sumData.coin
		daydata.TotalRating = user.sumData.careerdata.TotalRating
		daydata.TotalRank = user.sumData.careerdata.TotalRank
	} else {
		daydata.UID = user.GetDBID()
		daydata.UserName = user.GetName()
		daydata.Season = season
		daydata.NowTime = time.Now().Unix()
		daydata.StartTime = time.Now().Format("2006-01-02")
		if user.sumData.rank == 1 {
			daydata.DayFirstNum++
		}
		if user.sumData.rank <= 10 {
			daydata.DayTopTenNum++
		}
		daydata.DayEffectHarm += user.effectharm
		daydata.DayShotNum += user.sumData.shotnum
		daydata.DaySurviveTime += user.sumData.endgametime - user.sumData.begingametime
		daydata.DayDistance += user.sumData.runDistance / 1000
		daydata.DayCarDistance += user.sumData.carDistance / 1000
		daydata.DayAttackNum += user.sumData.attacknum
		daydata.DayRecoverNum += user.sumData.recvitemusenum
		daydata.DayRevivenum += user.sumData.revivenum
		daydata.DayHeadShotNum += user.headshotnum
		daydata.DayBattleNum++
		daydata.DayKillNum += user.kill
		daydata.DayCoin += user.sumData.coin
		daydata.TotalRating = user.sumData.careerdata.TotalRating
		daydata.TotalRank = user.sumData.careerdata.TotalRank
	}

	err := playerDayDataUtil.SetDayData(daydata)
	if err != nil {
		user.Error("SetDayData err: ", err)
	}

	// 当天首次吃鸡
	if daydata.DayFirstNum == 1 && user.sumData.rank == 1 {
		user.RPC(common.ServerTypeLobby, "PickFirstWin")
	}

	loginMsg := user.GetPlayerLogin()
	if loginMsg != nil && loginMsg.LoginChannel == 2 {
		//qqscorebatch
		// 当天参与比赛总场次
		// 当天获得比赛第一名（吃鸡）场次
		// 当天获得比赛前10名场次
		// 个人综合评分
		// 个人综合评分排名
		// 累计吃鸡数量
		// 累计参与比赛场次
		lst := []*msdk.Param{
			&msdk.Param{
				Tp:      1004,
				BCover:  1,
				Data:    fmt.Sprintf("%v", user.sumData.careerdata.FirstNum),
				Expires: "不过期",
			},
			&msdk.Param{
				Tp:      1005,
				BCover:  1,
				Data:    fmt.Sprintf("%v", user.sumData.careerdata.TotalBattleNum),
				Expires: "不过期",
			},
			&msdk.Param{
				Tp:      1006,
				BCover:  1,
				Data:    fmt.Sprintf("%v", daydata.DayBattleNum),
				Expires: "不过期",
			},
			&msdk.Param{
				Tp:      1007,
				BCover:  1,
				Data:    fmt.Sprintf("%v", daydata.DayFirstNum),
				Expires: "不过期",
			},
			&msdk.Param{
				Tp:      1008,
				BCover:  1,
				Data:    fmt.Sprintf("%v", daydata.DayTopTenNum),
				Expires: "不过期",
			},
			&msdk.Param{
				Tp:      1009,
				BCover:  1,
				Data:    fmt.Sprintf("%v", user.sumData.careerdata.TotalRating),
				Expires: "不过期",
			},
			&msdk.Param{
				Tp:      1010,
				BCover:  1,
				Data:    fmt.Sprintf("%v", user.sumData.careerdata.TotalRank),
				Expires: "不过期",
			},
		}

		msdk.QQScoreBatchList(common.QQAppIDStr, common.MSDKKey, loginMsg.VOpenID, user.GetAccessToken(), loginMsg.PlatID, user.GetName(), user.GetDBID(), lst)
	}
}

// writeDataReds
func (user *RoomUser) writeDataReds() {

	scene := user.GetSpace().(*Scene)
	if scene == nil {
		user.Error("writeDataReds failed, can't get scene")
		return
	}

	//计算排名
	user.sumData.rankCal()

	user.sumData.RunDistance()
	user.sumData.CarDistance()
	user.sumData.SwimDistance()

	//穿吉利服吃鸡
	if user.sumData.rank == 1 && user.GetIsWearingGilley() == 1 {
		user.sumData.isGilleyWin = true
	}

	//吃鸡未调用死亡结算时
	if user.sumData.endgametime == user.sumData.begingametime {
		user.sumData.endgametime = time.Now().Unix()
		if user.sumData.deadType == 0 && user.sumData.rank == 1 {
			user.sumData.deadType = 9
		}
	}

	//统计观战时长
	if user.sumData.watchStartTime != 0 {
		user.sumData.watchEndTime = time.Now().Unix()
	}

	user.sumData.coin = user.sumData.GetCoin(user.sumData.rank, user.kill)
	user.sumData.braveCoin = user.sumData.GetBraveCoin(user.sumData.rank, user.kill)

	user.SendNewYearActivityInfo() //统计新年活动数据

	user.updateTaskItems() //更新玩家任务项的进度

	//勇者战场记录使用同一张入场券的连胜场次
	if scene.GetMatchMode() == common.MatchModeBrave {
		user.updateBraveRecord()
		return
	}

	if scene.GetMatchMode() == common.MatchModeNormal {
		user.updateAchievement() //更新成就任务
		user.sumData.AchievementData()
		user.sumData.HonorInfo()
	}

	//娱乐模式下不记录比赛数据，不计算rating分
	if scene.GetMatchMode() >= 1 && scene.GetMatchMode() <= 10 {
		return
	}

	user.sumData.singleKillWinRating() //本局killrating，winrating 分计算
	user.sumData.totalKillWinRating()  //总rating分总计

	//判断所有赛季的数据是否在redis中有记录，如果无将第二赛季的数据导入总赛季中，标志是Season=0代表总赛季
	util := db.PlayerCareerDataUtil(common.PlayerCareerTotalData, user.GetDBID(), 0)
	if !util.IsKeyExist() {
		user.initTotalCareerData()
	}

	user.updateCareerData(0)                  //总赛季生涯数据统计录入redis(所有赛季)
	user.updateCareerData(common.GetSeason()) //当前赛季生涯数据统计录入redis
	user.DayData()                            //一天数据录入redis

	go user.postOneRoundData()
}

// initTotalCareerData 将第二赛季的数据导入总赛季中，标志是Season=0代表总赛季
func (user *RoomUser) initTotalCareerData() {
	var (
		playerCareerDataTypStrs = map[uint8]string{
			0: common.PlayerCareerTotalData,
			1: common.PlayerCareerSoloData,
			2: common.PlayerCareerDuoData,
			4: common.PlayerCareerSquadData,
		}
	)

	for _, v := range playerCareerDataTypStrs {
		seasonTwoData := &datadef.CareerBase{}
		seasonTwoUtil := db.PlayerCareerDataUtil(v, user.GetDBID(), 2)
		if err := seasonTwoUtil.GetRoundData(seasonTwoData); err != nil {
			user.Error("GetRoundData err: ", err)
		}

		util := db.PlayerCareerDataUtil(v, user.GetDBID(), 0)
		if err := util.SetRoundData(seasonTwoData); err != nil {
			user.Error("SetRoundData err: ", err)
		}
	}
}

// httpPost post本局数据到DataCenter服
func (user *RoomUser) postOneRoundData() {
	scene := user.GetSpace().(*Scene)
	if scene == nil {
		user.Error("postOneRoundData failed, can not get scene")
		return
	}
	oneRoundData := &datadef.OneRoundData{}
	oneRoundData.GameID = scene.GameID
	oneRoundData.UID = user.GetDBID()
	oneRoundData.Season = common.GetSeason()
	oneRoundData.Rank = user.sumData.rank
	oneRoundData.KillNum = user.kill
	oneRoundData.HeadShotNum = user.headshotnum
	oneRoundData.EffectHarm = user.effectharm
	oneRoundData.RecoverUseNum = user.sumData.recvitemusenum
	oneRoundData.ShotNum = user.sumData.shotnum
	oneRoundData.ReviveNum = user.sumData.revivenum
	oneRoundData.KillDistance = user.sumData.killdistance
	oneRoundData.KillStmNum = user.sumData.killstmnum
	oneRoundData.FinallHp = user.GetHP()
	oneRoundData.RecoverHp = user.sumData.recoverHp
	oneRoundData.RunDistance = user.sumData.runDistance / 1000
	oneRoundData.CarUseNum = user.sumData.carusernum
	oneRoundData.CarDestoryNum = user.sumData.carDestoryNum
	oneRoundData.AttackNum = user.sumData.attacknum
	oneRoundData.SpeedNum = user.sumData.speednum
	oneRoundData.Coin = user.sumData.coin
	oneRoundData.StartUnix = user.sumData.begingametime
	oneRoundData.EndUnix = user.sumData.endgametime
	oneRoundData.StartTime = time.Unix(user.sumData.begingametime, 0).Format("2006-01-02 15:04:05")
	oneRoundData.EndTime = time.Unix(user.sumData.endgametime, 0).Format("2006-01-02 15:04:05")
	oneRoundData.BattleType = user.sumData.battletype
	oneRoundData.KillRating = user.sumData.killRating
	oneRoundData.WinRatig = user.sumData.winRating
	oneRoundData.SoloKillRating = user.sumData.careerdata.SoloKillRating
	oneRoundData.SoloWinRating = user.sumData.careerdata.SoloWinRating
	oneRoundData.SoloRating = user.sumData.careerdata.SoloRating
	oneRoundData.SoloRank = user.sumData.careerdata.SoloRank
	oneRoundData.DuoKillRating = user.sumData.careerdata.DuoKillRating
	oneRoundData.DuoWinRating = user.sumData.careerdata.DuoWinRating
	oneRoundData.DuoRating = user.sumData.careerdata.DuoRating
	oneRoundData.DuoRank = user.sumData.careerdata.DuoRank
	oneRoundData.SquadKillRating = user.sumData.careerdata.SquadKillRating
	oneRoundData.SquadWinRating = user.sumData.careerdata.SquadWinRating
	oneRoundData.SquadRating = user.sumData.careerdata.SquadRating
	oneRoundData.SquadRank = user.sumData.careerdata.SquadRank
	oneRoundData.TotalRating = user.sumData.careerdata.TotalRating
	oneRoundData.TotalRank = user.sumData.careerdata.TotalRank
	oneRoundData.TopRating = user.sumData.careerdata.TopRating
	oneRoundData.TotalBattleNum = user.sumData.careerdata.TotalBattleNum
	oneRoundData.TotalFirstNum = user.sumData.careerdata.FirstNum
	oneRoundData.TotalTopTenNum = user.sumData.careerdata.TopTenNum
	oneRoundData.TotalHeadShot = user.sumData.careerdata.TotalHeadShot
	oneRoundData.TotalKillNum = user.sumData.careerdata.TotalKillNum
	oneRoundData.TotalShotNum = user.sumData.careerdata.Totalshotnum
	oneRoundData.TotalEffectHarm = user.sumData.careerdata.TotalEffectHarm
	oneRoundData.TotalSurviveTime = user.sumData.careerdata.SurviveTime
	oneRoundData.TotalDistance = user.sumData.careerdata.TotalDistance
	oneRoundData.TotalRecvItemUseNum = user.sumData.careerdata.RecvItemUseNum
	oneRoundData.SingleMaxKill = user.sumData.careerdata.SingleMaxKill
	oneRoundData.SingleMaxHeadShot = user.sumData.careerdata.SingleMaxHeadShot
	oneRoundData.DeadType = user.sumData.deadType
	oneRoundData.UserName = user.GetName()
	oneRoundData.SkyBox = scene.skybox
	loginMsg := user.GetPlayerLogin()
	oneRoundData.PlatID = loginMsg.PlatID
	oneRoundData.LoginChannel = loginMsg.LoginChannel

	// user.Info("username:", user.GetName(), " oneRoundData.UserName:", oneRoundData.UserName, " skybox:", scene.skybox, " oneRoundData.SkyBox:", oneRoundData.SkyBox)
	// user.Info("oneRoundData ", oneRoundData)

	data, err := json.Marshal(oneRoundData)
	if err != nil {
		user.Error("postOneRoundData failed, Marshal err: ", err)
		return
	}

	dataCenterInnerAddr, err := db.GetDataCenterAddr("DataCenterInnerAddr")
	if err != nil {
		user.Error("postOneRoundData failed, GetDataCenterAddr err: ", err)
		return
	}

	resp, err := http.Post("http://"+dataCenterInnerAddr+"/dataCenter", "application/json", strings.NewReader(string(data)))
	if err != nil {
		user.Error("postOneRoundData failed, Post err: ", err)
		return
	}
	//user.Info("dataCenterInnerAddr:", dataCenterInnerAddr, " post data 成功")

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		user.Error("postOneRoundData failed, ReadAll err: ", err)
		return
	}

	user.Info("postOneRoundData success, body: ", string(body))
}

// setPipeRank 先设置再获取排名
func (user *RoomUser) setPipeRank(rankType string, season int, score float64, uid uint64) uint32 {
	//log.Info("rankType:", rankType, " season:", season, " score:", score, " uid:", uid)
	if score == 0 {
		user.Warn("score is zero, rankType: ", rankType)
		return 0
	}
	playerRankUtil := db.PlayerRankUtil(rankType, season)
	rank, err := playerRankUtil.PipeRanksUint64(score, uid)
	if err != nil {
		user.Error("setPipeRank failed, PipeRanks err: ", err)
		return 0
	}

	return rank
}

// SendNewYearActivityInfo 发送本局个别数据到lobby
func (user *RoomUser) SendNewYearActivityInfo() {
	scene := user.GetSpace().(*Scene)
	if scene == nil {
		user.Error("SendNewYearActivityInfo failed, can't get scene")
		return
	}

	newYearInfo := &protoMsg.NewYearInfo{}
	if user.sumData.rescueNum != 0 {
		newYearGood := &protoMsg.NewYearGood{}
		newYearGood.Key = common.Act_RescueNum
		newYearGood.Num = int32(user.sumData.rescueNum)
		newYearInfo.GoodID = append(newYearInfo.GoodID, newYearGood) // 累积救援次数
	}

	if scene.GetMatchMode() == common.MatchModeScuffle {
		newYearGood := &protoMsg.NewYearGood{}
		newYearGood.Key = common.Act_ScuffleNum
		newYearGood.Num = 1
		newYearInfo.GoodID = append(newYearInfo.GoodID, newYearGood) //累积参加快速模式次数
	}

	if user.sumData.runDistance != 0 {
		newYearGood := &protoMsg.NewYearGood{}
		newYearGood.Key = common.Act_RunDistance
		newYearGood.Num = int32(user.sumData.runDistance)
		newYearInfo.GoodID = append(newYearInfo.GoodID, newYearGood) //累积统计跑动距离
	}

	if user.kill != 0 {
		newYearGood := &protoMsg.NewYearGood{}
		newYearGood.Key = common.Act_KillNum
		newYearGood.Num = int32(user.kill)
		newYearInfo.GoodID = append(newYearInfo.GoodID, newYearGood) //累积击杀次数
	}

	if user.headshotnum != 0 {
		newYearGood := &protoMsg.NewYearGood{}
		newYearGood.Key = common.Act_Headshotnum
		newYearGood.Num = int32(user.headshotnum)
		newYearInfo.GoodID = append(newYearInfo.GoodID, newYearGood) //累积爆头次数
	}

	if user.sumData.rank <= 10 {
		newYearGood := &protoMsg.NewYearGood{}
		newYearGood.Key = common.Act_RankTen
		newYearGood.Num = 1
		newYearInfo.GoodID = append(newYearInfo.GoodID, newYearGood) //累积前十名次数
	}

	if user.sumData.endgametime-user.sumData.begingametime > 0 {
		newYearGood := &protoMsg.NewYearGood{}
		newYearGood.Key = common.Act_SurviveTime
		newYearGood.Num = int32(math.Floor(float64(user.sumData.endgametime-user.sumData.begingametime) / 60))
		newYearInfo.GoodID = append(newYearInfo.GoodID, newYearGood) //累积生存时间
	}

	if user.sumData.rank == 1 {
		newYearGood := &protoMsg.NewYearGood{}
		newYearGood.Key = common.Act_WinNum
		newYearGood.Num = 1
		newYearInfo.GoodID = append(newYearInfo.GoodID, newYearGood) //累积吃鸡次数
	}

	if user.effectharm > 0 {
		newYearGood := &protoMsg.NewYearGood{}
		newYearGood.Key = common.Act_Effectharm
		newYearGood.Num = int32(user.effectharm)
		newYearInfo.GoodID = append(newYearInfo.GoodID, newYearGood) //累积有效伤害量次数
	}

	newYearGood := &protoMsg.NewYearGood{}
	newYearGood.Key = common.Act_GameNum
	newYearGood.Num = 1
	newYearInfo.GoodID = append(newYearInfo.GoodID, newYearGood) //累积参加演戏次数

	if scene.teamMgr.isTeam {
		newYearGood := &protoMsg.NewYearGood{}
		newYearGood.Key = common.Act_GameTeamNum
		newYearGood.Num = 1
		newYearInfo.GoodID = append(newYearInfo.GoodID, newYearGood) //累积组队参加演戏次数
	}

	if user.sumData.recoverHp > 0 {
		newYearGood := &protoMsg.NewYearGood{}
		newYearGood.Key = common.Act_RecoverHp
		newYearGood.Num = int32(user.sumData.recoverHp)
		newYearInfo.GoodID = append(newYearInfo.GoodID, newYearGood) //累积恢复血量
	}

	if scene.teamMgr.isTeam && user.sumData.rank <= 10 {
		newYearGood := &protoMsg.NewYearGood{}
		newYearGood.Key = common.Act_RankTeamTen
		newYearGood.Num = 1
		newYearInfo.GoodID = append(newYearInfo.GoodID, newYearGood) //累积组队进入前十
	}

	if user.sumData.doubleKillNum > 0 {
		newYearGood := &protoMsg.NewYearGood{}
		newYearGood.Key = common.Act_MultiKillnum
		newYearGood.Num = int32(user.sumData.doubleKillNum)
		newYearInfo.GoodID = append(newYearInfo.GoodID, newYearGood) //累计打出二连击次数
	}

	user.RPC(common.ServerTypeLobby, "SetNewYearActivityInfo", newYearInfo)
}

// updateAchievement 更新成就系统
func (user *RoomUser) updateAchievement() {
	totalData := &datadef.CareerBase{}
	util := db.PlayerCareerDataUtil(common.PlayerCareerTotalData, user.GetDBID(), 0)

	if err := util.GetRoundData(totalData); err != nil {
		user.Error("GetRoundData err: ", err)
		return
	}

	goodsUtil := db.PlayerGoodsUtil(user.GetDBID())

	isHaveRole1, isHaveRole7 := false, false
	mails := db.MailUtil(user.GetDBID()).GetMails()
	for _, v := range mails {
		for _, j := range v.Objs {
			if j.Id == 1 {
				isHaveRole1 = true
			}

			if j.Id == 7 {
				isHaveRole7 = true
			}
		}
	}

	// 完成成就1 玩家首次单局十杀
	if totalData.SingleMaxKill < 10 && user.kill >= 10 {
		user.HaveAchieved(1)
	} else if !goodsUtil.IsOwnGoods(7) && user.kill >= 10 && !isHaveRole7 {
		user.HaveAchieved(1)
	}

	// 完成成就2 玩家首次获得第一名
	if totalData.FirstNum == 0 && user.sumData.rank == 1 {
		user.HaveAchieved(2)
	} else if !goodsUtil.IsOwnGoods(1) && user.sumData.rank == 1 && !isHaveRole1 {
		user.HaveAchieved(2)
	}
}

// updateMilitaryRank 更新玩家的军衔等级
func (user *RoomUser) updateMilitaryRank() {
	scene := user.GetSpace().(*Scene)
	if scene == nil {
		user.Error("updateMilitaryRank failed, can't get scene")
		return
	}

	veteranNum := user.calVeteranNum()
	total, _ := db.PlayerCareerDataUtil(common.PlayerCareerTotalData, user.GetDBID(), 0).GetValueField("TotalBattleNum")
	comrade := scene.teamMgr.isComradeGame(user)
	baseExp, totalExp, levelBonus, olderBonus, cardBonus, comradeBonus, veteranBonus, actualExp := common.GetExpAfterGame(user.GetDBID(), total, user.GetLevel(), user.sumData.rank, user.kill, scene.GetMatchMode(), uint32(scene.mapdata.Map_pxs_ID), user.GetMatchTyp(), veteranNum, comrade)

	if actualExp > 0 {
		user.RPC(common.ServerTypeLobby, "AddExp", actualExp)
	}

	user.RPC(iserver.ServerTypeClient, "BonusExpNotify", baseExp, totalExp, levelBonus, olderBonus, cardBonus, comradeBonus, veteranBonus, actualExp)
	user.Info("updateMilitaryRank, baseExp: ", baseExp, " totalExp: ", totalExp, " levelBonus: ", levelBonus, " olderBonus: ", olderBonus, " cardBonus: ", cardBonus, " comradeBonus: ", comradeBonus, " veteranBonus:", veteranBonus, " actualExp:", actualExp)
}

// updateTaskItems 更新玩家的任务完成进度
func (user *RoomUser) updateTaskItems() {
	scene := user.GetSpace().(*Scene)
	if scene == nil {
		return
	}

	items := []uint32{}
	items = append(items, common.TaskItem_Game, 1)

	live := uint32(user.sumData.endgametime - user.sumData.begingametime)
	if live > 0 {
		items = append(items, common.TaskItem_Live, live)
	}

	if live >= 800 {
		items = append(items, common.TaskItem_Live800, 1)
	}

	if live >= 1500 {
		items = append(items, common.TaskItem_Live1500, 1)
	}

	if user.sumData.rank <= 1 {
		items = append(items, common.TaskItem_Win, 1)
	}

	if user.sumData.rank <= 10 {
		items = append(items, common.TaskItem_Topten, 1)
	}

	if user.sumData.rank <= 10 && user.sumData.battletype != 0 {
		items = append(items, common.TaskItem_TeamTopten, 1)
	}

	if user.kill > 0 {
		items = append(items, common.TaskItem_KillEnemy, user.kill)
	}

	if user.kill >= 3 {
		items = append(items, common.TaskItem_KillEnemy3, 1)
	}

	if user.sumData.recoverHp > 0 {
		items = append(items, common.TaskItem_RecoverHp, user.sumData.recoverHp)
	}

	if user.sumData.recoverHp >= 300 {
		items = append(items, common.TaskItem_RecoverHp300, 1)
	}

	if user.sumData.recoverHp > 1000 {
		items = append(items, common.TaskItem_RecoverHp1000, 1)
	}

	if user.effectharm > 0 {
		items = append(items, common.TaskItem_DamageEnemy, user.effectharm)
	}

	move := uint32(math.Abs(float64(user.sumData.runDistance)))
	if move > 0 {
		items = append(items, common.TaskItem_Move, move)
	}

	if move >= 20000 {
		items = append(items, common.TaskItem_Move20000, 1)
	}

	swim := uint32(math.Abs(float64(user.sumData.swimDistance)))
	if swim > 0 {
		items = append(items, common.TaskItem_Swim, swim)
	}

	if user.sumData.rescueNum > 0 {
		items = append(items, common.TaskItem_RescueTeammate, user.sumData.rescueNum)
	}

	if user.sumData.rescueNum >= 3 {
		items = append(items, common.TaskItem_RescueTeammate3, 1)
	}

	if user.sumData.rescueNum >= 5 {
		items = append(items, common.TaskItem_RescueTeammate5, 1)
	}

	if user.sumData.doubleKillNum > 0 {
		items = append(items, common.TaskItem_DoubleKill, user.sumData.doubleKillNum)
	}

	if user.sumData.tripleKillNum > 0 {
		items = append(items, common.TaskItem_TripleKill, user.sumData.tripleKillNum)
	}

	if user.sumData.fistKillNum >= 5 {
		items = append(items, common.TaskItem_FistKill5, 1)
	}

	if user.sumData.tankKillNum > 0 {
		items = append(items, common.TaskItem_TankKill, user.sumData.tankKillNum)
	}

	if user.sumData.carKillNum >= 8 {
		items = append(items, common.TaskItem_CarKill8, 1)
	}

	if user.sumData.sniperGunKillNum >= 8 {
		items = append(items, common.TaskItem_AWMKill8, 1)
	}

	if user.sumData.rpgKillNum > 0 {
		items = append(items, common.TaskItem_RPGKill, user.sumData.rpgKillNum)
	}

	if user.sumData.pistolKillNum >= 5 {
		items = append(items, common.TaskItem_PistolKill5, 1)
	}

	if user.headshotnum > 0 {
		items = append(items, common.TaskItem_Crit, user.headshotnum)
	}

	dropBoxs := uint32(len(user.sumData.getDropBoxs))
	if dropBoxs > 0 {
		items = append(items, common.TaskItem_GetDropBox, dropBoxs)
	}

	if dropBoxs >= 5 {
		items = append(items, common.TaskItem_GetDropBox5, 1)
	}

	if user.sumData.signalGunUseNum > 0 {
		items = append(items, common.TaskItem_SignalGun, user.sumData.signalGunUseNum)
	}

	if user.sumData.signalGunUseNum >= 3 {
		items = append(items, common.TaskItem_SignalGun3, 1)
	}

	if user.sumData.shotnum >= 500 {
		items = append(items, common.TaskItem_ShootBullet500, 1)
	}

	if user.sumData.isParachuteDie {
		items = append(items, common.TaskItem_ParachuteDie, 1)
	}

	if user.sumData.isLandDie {
		items = append(items, common.TaskItem_LandDie, 1)
	}

	if user.sumData.fallDamage > 0 {
		items = append(items, common.TaskItem_FallDamage, user.sumData.fallDamage)
	}

	if user.sumData.fallDamage >= 500 {
		items = append(items, common.TaskItem_FallDamage500, 1)
	}

	if user.sumData.isGilleyWin {
		items = append(items, common.TaskItem_GilleyWin, 1)
	}

	if user.sumData.isOwnDrink20 {
		items = append(items, common.TaskItem_EnergyDrink20, 1)
	}

	comrade := scene.teamMgr.isComradeGame(user)
	user.RPC(common.ServerTypeLobby, "UpdateTaskItems", common.Uint32sToBytes(items), comrade, user.sumData.rescueComradeNum)
}

// calVeteranNum 统计本局队伍中老兵的数量
func (user *RoomUser) calVeteranNum() uint8 {
	scene := user.GetSpace().(*Scene)
	if scene == nil {
		return 0
	}

	if !scene.teamMgr.isTeam {
		return 0
	}

	team, ok := scene.teamMgr.teams[user.teamid]
	if !ok || len(team) <= 1 {
		return 0
	}

	var veteranNum uint8
	for _, v := range team {
		info := scene.teamMgr.GetSpaceMemberInfo(v)
		if info == nil {
			continue
		}

		if info.veteran == 1 {
			veteranNum++
		}
	}

	return veteranNum
}

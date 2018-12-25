package main

import (
	"common"
	"datadef"
	"db"
	"excel"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"zeus/iserver"

	"github.com/garyburd/redigo/redis"
)

// GmMgr Gm命令
type GmMgr struct {
	user *LobbyUser

	// cmds 命令集合
	cmds map[string](func(map[string]string))

	gmSkyType      uint32 // gm设置天空盒类型
	isUseGmSkyType uint32 // 0 不使用 1 使用
}

// NewGmMgr 获取Gm管理器
func NewGmMgr(user *LobbyUser) *GmMgr {
	gm := &GmMgr{
		user:           user,
		cmds:           make(map[string](func(map[string]string))),
		gmSkyType:      0,
		isUseGmSkyType: 0,
	}
	gm.init()

	return gm
}

// init 初始化管理器
func (gm *GmMgr) init() {
	gm.cmds["SetSkyType"] = gm.SetSkyType               // 设置天空盒类型
	gm.cmds["SetScore"] = gm.SetScore                   // 设置rating积分
	gm.cmds["SetWins"] = gm.SetWins                     // 设置获胜数
	gm.cmds["SetKills"] = gm.SetKills                   // 设置击杀数
	gm.cmds["ClearAward"] = gm.ClearAward               // 清除领奖记录
	gm.cmds["ServerTime"] = gm.ServerTime               // 获取服务器当前时间
	gm.cmds["AddExp"] = gm.AddExp                       // 增加经验值
	gm.cmds["SetActiveness"] = gm.SetActiveness         // 设置活跃度
	gm.cmds["OpenTreasureBox"] = gm.OpenTreasureBox     // 开宝箱
	gm.cmds["SetTotalBattleNum"] = gm.SetTotalBattleNum // 设置总场次
	gm.cmds["ClickBallStar"] = gm.ClickBallStar         // 抽一球成名转盘
	gm.cmds["ClearWorldCup"] = gm.ClearWorldCup         // 清理世界杯活动数据
	gm.cmds["AddGood"] = gm.AddGood
	gm.cmds["AddComradePoints"] = gm.AddComradePoints //添加战友积分
}

// exec 执行命令
func (gm *GmMgr) exec(paras string) {
	pairSet := strings.Split(paras, " ")

	pairMap := make(map[string]string)
	for _, pair := range pairSet {
		paraSet := strings.Split(pair, "=")
		if len(paraSet) != 2 {
			continue
		}

		pairMap[paraSet[0]] = paraSet[1]
	}

	if cmdStr, ok := pairMap["lcmd"]; ok {
		if cmd, ok := gm.cmds[cmdStr]; ok {
			cmd(pairMap)
		} else {
			gm.user.Warn("cmds don't contain ", cmdStr)
		}
	} else {
		gm.user.Warn("client params contains no cmds")
	}

	gm.user.Info("exec gm cmds, params: ", paras)
}

// SetSkyType 设置天空盒类型
func (gm *GmMgr) SetSkyType(paras map[string]string) {
	var skyType, useGm int
	if strSkyType, typeErr := paras["skyType"]; typeErr {
		var valueErr error
		skyType, valueErr = strconv.Atoi(strSkyType)
		if valueErr != nil {
			gm.user.Error("skyType param err: ", valueErr)
			return
		}
	} else {
		gm.user.Error("skyType param not exist")
		return
	}

	if strIsUseType, isUseErr := paras["isUseGm"]; isUseErr {
		var valueErr error
		useGm, valueErr = strconv.Atoi(strIsUseType)
		if valueErr != nil {
			gm.user.Error("isUseGm param err: ", valueErr)
			return
		}
	} else {
		gm.user.Error("isUseGm param is not provided")
		return
	}

	gm.gmSkyType = uint32(skyType)
	gm.isUseGmSkyType = uint32(useGm)

	gm.user.Info("Set sky type, skyType: ", skyType, " isUseGm: ", useGm)
}

// SetScore 设置分数
func (gm *GmMgr) SetScore(paras map[string]string) {
	//该指令目前仅支持普通模式
	if gm.user.matchMode != common.MatchModeNormal {
		return
	}

	var matchTyp uint64
	var score float64
	var rank uint64
	var err error

	if typeValue, ok := paras["type"]; ok {
		matchTyp, err = strconv.ParseUint(typeValue, 10, 64)
		if err != nil {
			gm.user.Error("SetScore failed, ParseUint err: ", err)
			return
		}
	}

	if scoreValue, ok := paras["score"]; ok {
		score, err = strconv.ParseFloat(scoreValue, 64)
		if err != nil {
			gm.user.Error("SetScore failed, ParseFloat err: ", err)
			return
		}
	}

	if rankValue, ok := paras["rank"]; ok {
		rank, err = strconv.ParseUint(rankValue, 10, 64)
		if err != nil {
			gm.user.Error("SetScore failed, ParseUint err: ", err)
			return
		}
	}

	playerCareerDataTypStrs := map[uint64]string{
		0: common.PlayerCareerTotalData,
		1: common.PlayerCareerSoloData,
		2: common.PlayerCareerDuoData,
		4: common.PlayerCareerSquadData,
	}

	matchTypStrs := map[uint64]string{
		0: common.TotalRank,
		1: common.SoloRank,
		2: common.DuoRank,
		4: common.SquadRank,
	}

	data := &datadef.CareerBase{}
	careerUtil := db.PlayerCareerDataUtil(playerCareerDataTypStrs[matchTyp], gm.user.GetDBID(), common.GetSeason())
	rankUtil := db.PlayerRankUtil(matchTypStrs[matchTyp]+common.SeasonRankTypStrRating, common.GetSeason())

	if careerUtil.IsKeyExist() {
		if err := careerUtil.GetRoundData(data); err != nil {
			gm.user.Error("GetRoundData err: ", err)
			return
		}
	} else {
		data = &datadef.CareerBase{
			UID: gm.user.GetDBID(),
		}
	}

	if rank > 0 {
		myrank, _ := rankUtil.GetPlayerRank(gm.user.GetDBID())
		if myrank == uint32(rank) {
			return
		}

		rank1 := rank - 1
		rank2 := rank

		if myrank > 0 && myrank < uint32(rank) {
			rank1 = rank
			rank2 = rank + 1
		}

		info1, _ := rankUtil.GetScoreByRank(rank1)
		info2, _ := rankUtil.GetScoreByRank(rank2)

		if len(info1) == 2 && len(info2) == 2 {
			score1, _ := redis.Float64(info1[1], nil)
			score2, _ := redis.Float64(info2[1], nil)
			score = (score1 + score2) / 2.0
		} else if len(info1) == 2 {
			score1, _ := redis.Float64(info1[1], nil)
			score = score1 - 1
			if score < 0 {
				score = 0
			}
		} else if len(info2) == 2 {
			score2, _ := redis.Float64(info2[1], nil)
			score = score2 + 1
		}
	}

	// randValue := uint32(rand.Intn(int(20)))
	// data.FirstNum = randValue + 1
	// data.TotalKillNum = uint32(rand.Intn(int(30))) + randValue

	// total := uint32(rand.Intn(int(40))) + randValue
	// if total > data.TotalBattleNum {
	// 	data.TotalBattleNum = total
	// }

	f := 1 - float64(time.Now().UnixNano()/1000%10000000000000)/float64(10000000000000)
	f /= 10000
	score += f

	if err := careerUtil.SetRoundData(data); err != nil {
		gm.user.Error("SetRoundData err: ", err)
		return
	}

	if _, err := rankUtil.PipeRanksUint64(score, gm.user.GetDBID()); err != nil {
		gm.user.Error("PipeRanks err: ", err)
		return
	}

	data = &datadef.CareerBase{}
	careerUtil = db.PlayerCareerDataUtil(common.PlayerCareerTotalData, gm.user.GetDBID(), common.GetSeason())

	if careerUtil.IsKeyExist() {
		if err := careerUtil.GetRoundData(data); err != nil {
			gm.user.Error("GetRoundData err: ", err)
			return
		}
	} else {
		data = &datadef.CareerBase{
			UID: gm.user.GetDBID(),
		}
	}

	switch matchTyp {
	case 0:
		data.TotalRating = float32(score)
	case 1:
		data.SoloRating = float32(score)
	case 2:
		data.DuoRating = float32(score)
	case 4:
		data.SquadRating = float32(score)
	}

	if err := careerUtil.SetRoundData(data); err != nil {
		gm.user.Error("SetRoundData err: ", err)
		return
	}

	gm.user.Debug("设置rating积分:", score)
}

// SetWins 设置获胜数
func (gm *GmMgr) SetWins(paras map[string]string) {
	//该指令目前仅支持普通模式
	if gm.user.matchMode != common.MatchModeNormal {
		return
	}

	var matchTyp uint64
	var score float64
	var rank uint64
	var err error

	if typeValue, ok := paras["type"]; ok {
		matchTyp, err = strconv.ParseUint(typeValue, 10, 64)
		if err != nil {
			gm.user.Error("SetWins failed, ParseUint err: ", err)
			return
		}
	}

	if scoreValue, ok := paras["wins"]; ok {
		score, err = strconv.ParseFloat(scoreValue, 64)
		if err != nil {
			gm.user.Error("SetWins failed, ParseFloat err: ", err)
			return
		}
	}

	if rankValue, ok := paras["rank"]; ok {
		rank, err = strconv.ParseUint(rankValue, 10, 64)
		if err != nil {
			gm.user.Error("SetWins failed, ParseUint err: ", err)
			return
		}
	}

	playerCareerDataTypStrs := map[uint64]string{
		0: common.PlayerCareerTotalData,
		1: common.PlayerCareerSoloData,
		2: common.PlayerCareerDuoData,
		4: common.PlayerCareerSquadData,
	}

	matchTypStrs := map[uint64]string{
		0: common.TotalRank,
		1: common.SoloRank,
		2: common.DuoRank,
		4: common.SquadRank,
	}

	data := &datadef.CareerBase{}
	careerUtil := db.PlayerCareerDataUtil(playerCareerDataTypStrs[matchTyp], gm.user.GetDBID(), common.GetSeason())
	rankUtil := db.PlayerRankUtil(matchTypStrs[matchTyp]+common.SeasonRankTypStrWins, common.GetSeason())

	if careerUtil.IsKeyExist() {
		if err := careerUtil.GetRoundData(data); err != nil {
			gm.user.Error("GetRoundData err: ", err)
			return
		}
	} else {
		data = &datadef.CareerBase{
			UID: gm.user.GetDBID(),
		}
	}

	if rank > 0 {
		myrank, _ := rankUtil.GetPlayerRank(gm.user.GetDBID())
		if myrank == uint32(rank) {
			return
		}

		rank1 := rank - 1
		rank2 := rank

		if myrank > 0 && myrank < uint32(rank) {
			rank1 = rank
			rank2 = rank + 1
		}

		info1, _ := rankUtil.GetScoreByRank(rank1)
		info2, _ := rankUtil.GetScoreByRank(rank2)

		if len(info1) == 2 && len(info2) == 2 {
			score1, _ := redis.Float64(info1[1], nil)
			score2, _ := redis.Float64(info2[1], nil)
			score = (score1 + score2) / 2.0
		} else if len(info1) == 2 {
			score1, _ := redis.Float64(info1[1], nil)
			score = score1 - 1
			if score < 0 {
				score = 0
			}
		} else if len(info2) == 2 {
			score2, _ := redis.Float64(info2[1], nil)
			score = score2 + 1
		}
	}

	randValue := uint32(rand.Intn(2 + int(score)))
	data.FirstNum = uint32(score)
	data.TopTenNum = uint32(score) + randValue + 2

	total := uint32(score) + 3*randValue + 5
	if total > data.TotalBattleNum {
		data.TotalBattleNum = total
	}

	f := 1 - float64(time.Now().UnixNano()/1000%10000000000000)/float64(10000000000000)
	score += f

	if err := careerUtil.SetRoundData(data); err != nil {
		gm.user.Error("SetRoundData err: ", err)
		return
	}

	if _, err := rankUtil.PipeRanksUint64(score, gm.user.GetDBID()); err != nil {
		gm.user.Error("PipeRanks err: ", err)
		return
	}

	gm.user.Debug("设置获胜数: ", score)
}

// SetKills 设置击杀数
func (gm *GmMgr) SetKills(paras map[string]string) {
	//该指令目前仅支持普通模式
	if gm.user.matchMode != common.MatchModeNormal {
		return
	}

	var matchTyp uint64
	var score float64
	var rank uint64
	var err error

	if typeValue, ok := paras["type"]; ok {
		matchTyp, err = strconv.ParseUint(typeValue, 10, 64)
		if err != nil {
			gm.user.Error("SetKills failed, ParseUint err: ", err)
			return
		}
	}

	if scoreValue, ok := paras["kills"]; ok {
		score, err = strconv.ParseFloat(scoreValue, 64)
		if err != nil {
			gm.user.Error("SetKills failed, ParseFloat err: ", err)
			return
		}
	}

	if rankValue, ok := paras["rank"]; ok {
		rank, err = strconv.ParseUint(rankValue, 10, 64)
		if err != nil {
			gm.user.Error("SetKills failed, ParseUint err: ", err)
			return
		}
	}

	playerCareerDataTypStrs := map[uint64]string{
		0: common.PlayerCareerTotalData,
		1: common.PlayerCareerSoloData,
		2: common.PlayerCareerDuoData,
		4: common.PlayerCareerSquadData,
	}

	matchTypStrs := map[uint64]string{
		0: common.TotalRank,
		1: common.SoloRank,
		2: common.DuoRank,
		4: common.SquadRank,
	}

	data := &datadef.CareerBase{}
	careerUtil := db.PlayerCareerDataUtil(playerCareerDataTypStrs[matchTyp], gm.user.GetDBID(), common.GetSeason())
	rankUtil := db.PlayerRankUtil(matchTypStrs[matchTyp]+common.SeasonRankTypStrKills, common.GetSeason())

	if careerUtil.IsKeyExist() {
		if err := careerUtil.GetRoundData(data); err != nil {
			gm.user.Error("GetRoundData err: ", err)
			return
		}
	} else {
		data = &datadef.CareerBase{
			UID: gm.user.GetDBID(),
		}
	}

	if rank > 0 {
		myrank, _ := rankUtil.GetPlayerRank(gm.user.GetDBID())
		if myrank == uint32(rank) {
			return
		}

		rank1 := rank - 1
		rank2 := rank

		if myrank > 0 && myrank < uint32(rank) {
			rank1 = rank
			rank2 = rank + 1
		}

		info1, _ := rankUtil.GetScoreByRank(rank1)
		info2, _ := rankUtil.GetScoreByRank(rank2)

		if len(info1) == 2 && len(info2) == 2 {
			score1, _ := redis.Float64(info1[1], nil)
			score2, _ := redis.Float64(info2[1], nil)
			score = (score1 + score2) / 2.0
		} else if len(info1) == 2 {
			score1, _ := redis.Float64(info1[1], nil)
			score = score1 - 1
			if score < 0 {
				score = 0
			}
		} else if len(info2) == 2 {
			score2, _ := redis.Float64(info2[1], nil)
			score = score2 + 1
		}
	}

	randValue := uint32(rand.Intn(2 + int(score)/2))
	data.TotalKillNum = uint32(score)

	data.SingleMaxKill = randValue + 1
	if data.SingleMaxKill > 50 {
		data.SingleMaxKill = 50
	}
	if data.SingleMaxKill > data.TotalKillNum {
		data.SingleMaxKill = data.TotalKillNum
	}

	total := uint32(score) + randValue + 10
	if total > data.TotalBattleNum {
		data.TotalBattleNum = total
	}

	f := 1 - float64(time.Now().UnixNano()/1000%10000000000000)/float64(10000000000000)
	score += f

	if err := careerUtil.SetRoundData(data); err != nil {
		gm.user.Error("SetRoundData err: ", err)
		return
	}

	if _, err := rankUtil.PipeRanksUint64(score, gm.user.GetDBID()); err != nil {
		gm.user.Error("PipeRanks err: ", err)
		return
	}

	gm.user.Debug("设置击杀数: ", score)
}

// ClearAward 清除领奖记录，实现重复领奖
func (gm *GmMgr) ClearAward(paras map[string]string) {
	rankSeason := common.GetLastRankSeason()
	if rankSeason == 0 {
		return
	}
	db.PlayerInfoUtil(gm.user.GetDBID()).ClearSeasonAwardRecord(rankSeason)
}

// AddExp 增加经验值
func (gm *GmMgr) AddExp(paras map[string]string) {
	var (
		incr uint64
		err  error
	)

	if incrValue, ok := paras["incr"]; ok {
		incr, err = strconv.ParseUint(incrValue, 10, 64)
		if err != nil {
			gm.user.Error("AddExp failed, ParseUint err: ", err)
			return
		}
	}

	gm.user.UpdateMilitaryRank(uint32(incr))
}

// ServerTime 获取服务器当前时间
func (gm *GmMgr) ServerTime(paras map[string]string) {
	gm.user.RPC(iserver.ServerTypeClient, "GmServerTime", uint32(time.Now().Unix()))
}

// OpenTreasureBox 开宝箱
func (gm *GmMgr) OpenTreasureBox(paras map[string]string) {
	var boxId, num uint64
	var err error
	if boxIdValue, ok := paras["BoxId"]; ok {
		boxId, err = strconv.ParseUint(boxIdValue, 10, 64)
		if err != nil {
			gm.user.Error("OpenTreasureBox failed, ParseUint err: ", err, " boxId:", boxId)
			return
		}
	}

	if NumValue, ok := paras["Num"]; ok {
		num, err = strconv.ParseUint(NumValue, 10, 32)
		if err != nil {
			gm.user.Error("OpenTreasureBox failed, ParseUint err: ", err, " num:", num)
			return
		}
	}

	itemMap := make(map[uint32]uint32)
	poolMap := make(map[uint32]uint32)
	for i := 0; i < int(num); i++ {
		itemId, poolId := gm.user.treasureBoxMgr.randomItem(boxId, uint32(num))
		itemMap[itemId]++
		poolMap[poolId]++
	}

	gm.user.Debug("boxId:", boxId, " num:", num, " itemMap:", itemMap, " poolMap:", poolMap)
}

// SetActiveness 设置活跃度
func (gm *GmMgr) SetActiveness(paras map[string]string) {
	var (
		incr uint64
		err  error
	)

	if incrValue, ok := paras["incr"]; ok {
		incr, err = strconv.ParseUint(incrValue, 10, 64)
		if err != nil {
			gm.user.Error("SetActiveness failed, ParseUint err: ", err)
			return
		}
	}

	gm.user.updateActivenessProgress(1, uint32(incr))
	gm.user.updateActivenessProgress(2, uint32(incr))

	today := common.GetTodayBeginStamp()
	week := common.GetThisWeekBeginStamp()
	dayActiveness := db.PlayerInfoUtil(gm.user.GetDBID()).GetActiveness(1, today)
	weekActiveness := db.PlayerInfoUtil(gm.user.GetDBID()).GetActiveness(2, week)

	gm.user.RPC(iserver.ServerTypeClient, "SetActivenessRet", dayActiveness, weekActiveness)
}

// SetTotalBattleNum 设置总场次
func (gm *GmMgr) SetTotalBattleNum(paras map[string]string) {
	var (
		num uint64
		err error
	)

	if numValue, ok := paras["num"]; ok {
		num, err = strconv.ParseUint(numValue, 10, 64)
		if err != nil {
			gm.user.Error("SetTotalBattleNum failed, ParseUint err: ", err)
			return
		}
	}

	data := &datadef.CareerBase{}
	careerUtil := db.PlayerCareerDataUtil(common.PlayerCareerTotalData, gm.user.GetDBID(), 0)

	if careerUtil.IsKeyExist() {
		if err := careerUtil.GetRoundData(data); err != nil {
			gm.user.Error("GetRoundData err: ", err)
			return
		}
	} else {
		data = &datadef.CareerBase{
			UID: gm.user.GetDBID(),
		}
	}

	data.TotalBattleNum = uint32(num)

	if err := careerUtil.SetRoundData(data); err != nil {
		gm.user.Error("SetRoundData err: ", err)
		return
	}
}

// ClickBallStar 抽一球成名
func (gm *GmMgr) ClickBallStar(paras map[string]string) {
	var num uint64
	var err error

	if NumValue, ok := paras["Num"]; ok {
		num, err = strconv.ParseUint(NumValue, 10, 32)
		if err != nil {
			gm.user.Error("ClickBallStar failed, ParseUint err: ", err, " num:", num)
			return
		}
	}

	var position uint32 = 1
	roundMap := make(map[uint32]uint32)
	for i := 0; i < int(num); i++ {
		randomValue := gm.user.activityMgr.sweepstake(position, uint32(i))
		if randomValue == 0 {
			gm.user.Error("sweepstake Excel Err!")
			return
		}

		grid := uint32(common.GetTBSystemValue(common.System_BallStarGridNum))
		if position+randomValue > grid {
			position += randomValue - grid
		} else {
			position += randomValue
		}

		roundMap[position]++
	}

	gm.user.Debug("num:", num, " roundMap:", roundMap)
}

func (gm *GmMgr) ClearWorldCup(paras map[string]string) {
	db.NewPlayerWorldCupUtil(gm.user.GetDBID()).ClearAll()
}

func (gm *GmMgr) AddGood(paras map[string]string) {
	id := common.StringToUint32(paras["id"])
	num := common.StringToUint32(paras["num"])
	times := common.StringToUint32(paras["time"]) * 60 * num

	if num == 0 {
		return
	}

	// 商品信息
	goodsConfig, ok := excel.GetStore(uint64(id))
	if !ok {
		gm.user.Error("GetGoods failed, goods doesn't exist, id: ", id)
		return
	}

	if goodsConfig.Type == GoodsGiftPack && goodsConfig.UseType == 2 {
		gm.user.storeMgr.getGiftPack(id, num, 100, common.MT_NO, 0)
		return
	}

	if isMoneyGoods(uint32(goodsConfig.Type)) {
		return
	}

	isNew := false
	util := db.PlayerGoodsUtil(gm.user.GetDBID())

	dbid := uint32(0)
	if goodsConfig.Type == GoodsGiftPack {
		dbid = id
	} else {
		dbid = uint32(goodsConfig.RelationID)
	}

	info, err := util.GetGoodsInfo(dbid)

	if err != nil || info == nil {
		info = &db.GoodsInfo{
			Id:    dbid,
			State: 2, //State:0、1状态都代表旧道具(已经点击过),2代表新道具(未进行点击)
		}
	}
	switch info.Id {
	case common.Item_SeasonHeroMedal:
		{
			_, end := common.GetRankSeasonTimeStamp()
			info.EndTime = end
			isNew = true
		}
	default:
		if info.Time == 0 { //首次获得该道具
			info.Time = time.Now().Unix()
			if times > 0 {
				info.EndTime = time.Now().Unix() + int64(times)
			} else {
				isNew = true
			}
		} else { //已拥有该道具
			if times > 0 {
				if info.EndTime > 0 {
					info.EndTime += int64(times) //限时道具时效叠加
				}
			} else {
				if info.EndTime > 0 {
					info.EndTime = 0
					isNew = true
				}
			}
		}
	}

	info.Sum += num
	info.State = 2
	util.AddGoodsInfo(info)
	gm.user.storeMgr.rareItem(id, num)
	gm.user.activityMgr.updateExchangeOwnNum(dbid)

	// 通知客户端
	gm.user.RPC(iserver.ServerTypeClient, "AddGoods", info.ToProto())
	if goodsConfig.Timelimit == "" && isNew {
		if goodsConfig.Type == GoodsRoleType {
			gm.user.AddAchievementData(common.AchievementRole, 1)
		}
		if goodsConfig.Type == GoodsParachuteType {
			gm.user.AddAchievementData(common.AchievementParachute, 1)
		}
	}

	// 特殊道具处理
	gm.user.storeMgr.specialGoodsProc(dbid)

	if info.EndTime != 0 {
		gm.user.StartCrondForExpireCheck()
	}

	var leftTime uint32
	if info.EndTime > 0 {
		leftTime = uint32(info.EndTime - time.Now().Unix())
	}
	gm.user.tlogItemFlow(id, num, 100, 0, common.MT_NO, ADD, 0, leftTime) //tlog道具流水表
	gm.user.Debug("GetGoods success, id:", id, " num:", num, " times:", times, "s")
}

// AddComradePoints 添加战友积分
func (gm *GmMgr) AddComradePoints(paras map[string]string) {
	num := common.StringToUint32(paras["num"])
	if num == 0 {
		return
	}

	gm.user.SetComradePoints(gm.user.GetComradePoints() + uint64(num))
	gm.user.activityMgr.updateExchangeOwnNum(2206)
}

package main

import (
	"common"
	"db"
	"math"
	"zeus/iserver"
)

// notifyServerMode 将服务器当前开放的匹配模式信息通知给客户端
func (user *LobbyUser) notifyServerMode() {
	user.RPC(iserver.ServerTypeClient, "OnlineCheckMatchModeOpen", GetSrvInst().allowModes)
}

// notifyPlayerMode 将服务器记录的玩家所在的匹配模式通知给客户端
func (user *LobbyUser) notifyPlayerMode() {
	typ := user.GetMatchTyp()
	if !common.IsMatchModeOk(user.matchMode, typ) {
		user.matchMode = common.MatchModeNormal
	}

	user.RPC(iserver.ServerTypeClient, "PlayerModeNotify", user.matchMode)
	user.Info("Player mode notify, matchmode: ", user.matchMode)
}

// GetRanksByMode 获取对应匹配模式指定匹配类型的rank分
func (user *LobbyUser) GetRanksByMode(matchMode uint32, matchTyps ...uint8) map[uint8]uint32 {
	var (
		ranks = make(map[uint8]uint32)
		rank  uint32
	)

	switch matchMode {
	case common.MatchModeNormal:
		{
			data, _ := common.GetPlayerCareerData(user.GetDBID(), common.GetSeason(), 0)
			for _, typ := range matchTyps {
				switch typ {
				case 1:
					if data == nil || data.SoloRating == 0 {
						rank = uint32(common.GetTBMatchValue(common.Match_InitRank))
					} else {
						rank = uint32(math.Ceil(float64(data.SoloRating)))
					}
				case 2:
					if data == nil || data.DuoRating == 0 {
						rank = uint32(common.GetTBMatchValue(common.Match_InitRank))
					} else {
						rank = uint32(math.Ceil(float64(data.DuoRating)))
					}
				case 4:
					if data == nil || data.SquadRating == 0 {
						rank = uint32(common.GetTBMatchValue(common.Match_InitRank))
					} else {
						rank = uint32(math.Ceil(float64(data.SquadRating)))
					}
				default:
					user.Error("Unkonwn match typ: ", typ)
					return nil
				}

				ranks[typ] = rank
			}
		}
	case common.MatchModeScuffle, common.MatchModeEndless:
	case common.MatchModeBrave, common.MatchModeElite, common.MatchModeVersus, common.MatchModeArcade, common.MatchModeTankWar:
		{
			rank, _ := db.PlayerInfoUtil(user.GetDBID()).GetBraveWinStreak()
			for _, typ := range matchTyps {
				ranks[typ] = rank
			}
		}
	default:
		user.Error("Unkonwn match mode: ", matchMode)
		return nil
	}

	user.Debug("ranks:", ranks)
	if user.isQemu == 1 {
		value := uint32(common.GetTBMatchValue(common.Match_Simulator))
		for k, _ := range ranks {
			ranks[k] += value
		}
		user.Debug("模拟器加分:", value, " ranks:", ranks)
	}

	return ranks
}

// CanMatchBraveGame 能否匹配勇者战场比赛
func (user *LobbyUser) CanMatchBraveGame(matchTyp uint8) bool {
	info := common.GetOpenModeInfo(user.matchMode, matchTyp)
	if info == nil {
		return false
	}

	//每一开放时间段只允许玩家进行有限场次的比赛
	record, _ := db.PlayerInfoUtil(user.GetDBID()).GetBraveRecord()
	if record != nil {
		if info.UniqueId == record.UniqueId && info.DayStart == record.DayStart {
			if info.MatchNum != 0 && record.MatchNum >= info.MatchNum {
				user.AdviceNotify(common.NotifyCommon, 61)
				user.Warnf("Only can play %d round in a match time\n", info.MatchNum)
				return false
			}
		}
	}

	//需要购买入场券才能进入
	if !db.PlayerGoodsUtil(user.GetDBID()).IsGoodsEnough(common.Item_BraveTicket, info.ItemNum) {
		user.AdviceNotify(common.NotifyCommon, 60)
		user.Warnf("At least %d ticket is needed\n", info.ItemNum)
		return false
	}

	return true
}

// UpdateBraveRecord 记录玩家在勇者模式的同一开放时间段参加的比赛记录
func (user *LobbyUser) UpdateBraveRecord(mode uint32) {
	var info *common.ModeInfo
	if mode == 0 {
		info = common.GetOpenModeInfo(user.matchMode, 1)
	} else {
		info = common.GetOpenModeInfo(user.matchMode, uint8(mode*2))
	}

	if info == nil {
		return
	}

	util := db.PlayerInfoUtil(user.GetDBID())
	record, _ := util.GetBraveRecord()

	if record != nil {
		if info.UniqueId == record.UniqueId && info.DayStart == record.DayStart {
			record.MatchNum++
			util.SetBraveRecord(record)
			return
		}
	}

	util.SetBraveRecord(&db.BraveRecord{
		UniqueId: info.UniqueId,
		DayStart: info.DayStart,
		MatchNum: 1,
	})
}

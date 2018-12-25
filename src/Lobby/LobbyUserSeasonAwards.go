package main

import (
	"common"
	"db"
	"excel"
	"protoMsg"
	"zeus/iserver"
)

const (
	SeasonRankTopNum = 1000
)

// syncSeasonRankInfo 同步赛季排行标记信息
func (user *LobbyUser) syncSeasonRankInfo() {
	util := db.PlayerInfoUtil(user.GetDBID())
	rankSeason := 0 //客户端当前显示的排行榜对应的排行赛季id

	//当前赛季是否查看过排行榜
	enter := true
	cur := common.GetRankSeason()
	if cur != 0 {
		enter = util.IsSeasonBeginEnter(cur)
		rankSeason = cur
	}

	//上一赛季的奖励是否已领取
	award := true
	last := common.GetLastRankSeason()
	if last != 0 {
		if common.CanDrawSeasonAwards(user.GetDBID(), common.RankSeasonToSeason(last)) {
			award = util.IsSeasonEndAward(last)
		}
		if rankSeason == 0 {
			rankSeason = last
		}
	}

	if err := user.RPC(iserver.ServerTypeClient, "SeasonRankInfoRet", int32(rankSeason), enter, award); err != nil {
		user.Error("RPC SeasonRankInfoRet err: ", err)
		return
	}

	user.Info("SeasonRankInfoRet, rankSeason:  ", rankSeason, " enter: ", enter, " award: ", award)
}

// DrawSeasonAwards 赛季结束后，玩家领取奖励
func (user *LobbyUser) DrawSeasonAwards() {
	util := db.PlayerInfoUtil(user.GetDBID())
	rankSeason := common.GetLastRankSeason()
	awards := &protoMsg.SeasonAwards{
		Season: int32(rankSeason),
	}

	if rankSeason != 0 {
		if !util.IsSeasonEndAward(rankSeason) {
			//设置领奖记录
			util.SetSeasonEndAward(rankSeason)

			//领取本赛季所有排行榜的奖励
			if rankSeason == 1 {
				user.GetSeasonAwardsTotal(rankSeason, awards)
			} else {
				user.GetSeasonAwards(rankSeason, awards)
			}

			//将玩家领取到的赛季奖励保存到redis里
			for _, item := range awards.GetItems() {
				user.storeMgr.GetGoods(item.GetId(), item.GetNum(), common.RS_SeasonAward, common.MT_NO, 0)
			}
		}
	}

	if err := user.RPC(iserver.ServerTypeClient, "DrawSeasonAwardsRet", awards); err != nil {
		user.Error("RPC DrawSeasonAwardsRet err: ", err)
	}

	user.Infof("Draw season awards success, awards: %+v\n", awards)
}

// GetSeasonAwardsTotal 统计玩家获得的赛季排行奖励，仅对第一赛季有效
func (user *LobbyUser) GetSeasonAwardsTotal(rankSeason int, awards *protoMsg.SeasonAwards) {
	var (
		rankTypStrs = map[uint8]string{
			1: common.SeasonRankTypStrWins,
			2: common.SeasonRankTypStrKills,
		}

		season = common.RankSeasonToSeason(rankSeason)
	)

	for _, rankTyp := range []uint8{1, 2} {
		rank, err := db.PlayerRankUtil(common.TotalRank+rankTypStrs[rankTyp], season).GetPlayerRank(user.GetDBID())
		if err != nil {
			continue
		}

		if rank > SeasonRankTopNum {
			rank = 0
		}

		CombineAwards(awards, common.GetSeasonAwardsByRank(rankSeason, 1, rankTyp, rank))
	}
}

// GetSeasonAwards 统计玩家获得的赛季排行奖励
func (user *LobbyUser) GetSeasonAwards(rankSeason int, awards *protoMsg.SeasonAwards) {
	var (
		matchTypStrs = map[uint8]string{
			1: common.SoloRank,
			2: common.DuoRank,
			4: common.SquadRank,
		}

		rankTypStrs = map[uint8]string{
			1: common.SeasonRankTypStrWins,
			2: common.SeasonRankTypStrKills,
			3: common.SeasonRankTypStrRating,
		}

		season = common.RankSeasonToSeason(rankSeason)
	)

	for _, matchTyp := range []uint8{1, 2, 4} {
		for _, rankTyp := range []uint8{1, 2, 3} {
			rank, err := db.PlayerRankUtil(matchTypStrs[matchTyp]+rankTypStrs[rankTyp], season).GetPlayerRank(user.GetDBID())
			if err != nil || rank == 0 {
				continue
			}

			if rank > SeasonRankTopNum {
				rank = 0
			}

			CombineAwards(awards, common.GetSeasonAwardsByRank(rankSeason, matchTyp, rankTyp, rank))
		}
	}
}

// CombineAwards 合并玩家在不同排行榜中获得的奖励，如果获得了重复的奖励（限时道具除外），则换成等量的金币
func CombineAwards(oldAwards, newAwards *protoMsg.SeasonAwards) {
	var coinItem *protoMsg.AwardItem

	for _, oldItem := range oldAwards.Items {
		if oldItem.Id == common.Item_Coin {
			coinItem = oldItem
			break
		}
	}

	for _, newItem := range newAwards.Items {
		var combined bool

		for _, oldItem := range oldAwards.Items {
			if oldItem.Id == newItem.Id {
				if getGoodsTotalTime(newItem.Id) > 0 {
					continue
				}

				var coinNum uint32

				if newItem.Id == common.Item_Coin {
					coinNum = newItem.Num
				} else {
					goodsConfig, ok := excel.GetStore(uint64(newItem.Id))
					if !ok {
						continue
					}
					coinNum = uint32(goodsConfig.Price)
				}

				if coinItem == nil {
					coinItem = &protoMsg.AwardItem{
						Id:  common.Item_Coin,
						Num: 0,
					}
					oldAwards.Items = append(oldAwards.Items, coinItem)
				}

				coinItem.Num += coinNum
				combined = true
			}
		}

		if !combined {
			oldAwards.Items = append(oldAwards.Items, newItem)
		}
	}
}

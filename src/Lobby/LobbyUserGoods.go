package main

import (
	"common"
	"db"
	"excel"
	"math"
	"protoMsg"
	"strings"
	"time"
	"zeus/iserver"
)

// StoreMgr 个人商店管理器
type StoreMgr struct {
	user *LobbyUser
}

func NewStoreMgr(user *LobbyUser) *StoreMgr {
	store := &StoreMgr{
		user: user,
	}
	return store
}

func (mgr *StoreMgr) initPropInfo() {
	// 转化老数据
	mgr.oldDataTransfer()

	data := db.PlayerGoodsUtil(mgr.user.GetDBID()).GetAllGoodsInfo()

	retMsg := &protoMsg.OwnGoodsInfo{
		List: make([]*protoMsg.OwnGoodsItem, 0),
	}

	for _, j := range data {
		item := j.ToProto()
		retMsg.List = append(retMsg.List, item)
	}

	// 初始化玩家拥有的所有购买商品
	mgr.user.RPC(iserver.ServerTypeClient, "InitOwnGoodsInfo", retMsg)
	mgr.user.Infof("Goods list: %+v\n", retMsg)
}

// oldDataTransfer 老数据转化
// 线上数据库里存在以商城表id记录的限时道具，将其转化为以道具表id来记录，同时记录道具到期时间(endTime)
func (mgr *StoreMgr) oldDataTransfer() {
	util := db.PlayerGoodsUtil(mgr.user.GetDBID())
	data := util.GetAllGoodsInfo()

	for _, info := range data {
		if info.EndTime > 0 {
			continue
		}

		limitTime := getGoodsTotalTime(info.Id)
		if limitTime == 0 {
			continue
		}

		goodsConfig, ok := excel.GetStore(uint64(info.Id))
		if !ok {
			continue
		}

		info.Id = uint32(goodsConfig.RelationID)
		info.EndTime = info.Time + limitTime

		util.DelGoods(uint32(goodsConfig.Id))
		util.AddGoodsInfo(info)

		if mgr.user.GetParachuteID() == uint32(goodsConfig.RelationID) {
			mgr.user.SetGoodsParachuteID(uint32(goodsConfig.RelationID))
		}

		if mgr.user.GetRoleModel() == uint32(goodsConfig.RelationID) {
			mgr.user.SetGoodsRoleModel(uint32(goodsConfig.RelationID))
		}
	}
}

// 商品类型
const (
	// 角色
	GoodsRoleType = 1
	// 伞包
	GoodsParachuteType = 2
	// 金币
	GoodsGoldCoin = 3
	// 幻化背包
	GoodsGameEquipPack = 5
	// 幻化头盔
	GoodsGameEquipHead = 6
	// 活跃度
	GoodsActiveness = 7
	// 经验值
	GoodsExp = 8
	// 钻石
	GoodsDiamond = 9
	// 战友积分
	GoodsComradePoints = 11
	// 枪支幻化
	GoodsGameEquipGunSkin = 12
	// 勇气值
	GoodsBraveCoin = 16
	// 特训经验勋章
	GoodsSpecailExpMedal = 17
	// 赛季英雄勋章
	GoodsSeasonHeroMedal = 19
	// 赛季经验勋章
	GoodsSeasonExpMedal = 20
	// 礼包
	GoodsGiftPack = 21
	// 金砖
	GoodsGoldBrick = 9999
)

// 商品出售状态
const (
	// 出售中
	Onselling = 0
	// 不显示不出售
	NotShowNotSell = 1
	// 显示不出售
	ShowNotSell = 2
	// 免费
	GoodsFree = 3
)

// 购买结果
const (
	// 购买成功
	BuyGoodsSuccess = 1
	// 金币不足
	GoldCoinLack = 2
	// 已经购买
	RepeatBuyGoods = 3
	// 商品不存在
	GoodsNotExist = 4
	// 暂不出售
	GoodsNotSell = 5
	// 购买失败
	BuyGoodsFail = 6
)

// isMoneyGoods 是否货币类型的道具
func isMoneyGoods(goodsType uint32) bool {
	moneys := map[uint32]bool{
		GoodsGoldCoin:        true,
		GoodsDiamond:         true,
		GoodsBraveCoin:       true,
		GoodsExp:             true,
		GoodsActiveness:      true,
		GoodsComradePoints:   true,
		GoodsSpecailExpMedal: true,
		GoodsSeasonExpMedal:  true,
		GoodsGoldBrick:       true,
	}

	if moneys[goodsType] {
		return true
	}

	return false
}

// 购买商品
func (mgr *StoreMgr) BuyGoods(id uint32, num uint32, param []byte) uint32 {
	// 商品信息
	goodsConfig, ok := excel.GetStore(uint64(id))
	if !ok {
		mgr.user.Error("BuyGoods failed, goods does't exist, id: ", id)
		return GoodsNotExist
	}

	// 是否可购买
	if goodsConfig.State != Onselling {
		mgr.user.Error("BuyGoods failed, goods is not on selling, id: ", id)
		return GoodsNotSell
	}

	// 判断是否可重复购买
	if !mgr.repeatBuyGoodsCheck(id) {
		return BuyGoodsFail
	}

	moneyType := uint32(goodsConfig.Moneytype)
	amount := uint32(goodsConfig.Price * uint64(num))

	// 货币是否足够
	if !mgr.IsMoneyEnough(moneyType, amount) {
		mgr.user.Error("BuyGoods failed, coins is not enough, id: ", id)
		return GoldCoinLack
	}

	var ret uint32                          //金砖购买返回码
	var billno string                       //金砖购买订单号
	var succGoods = make(map[uint32]uint32) //已成功发放的道具
	var failGoods = make(map[uint32]uint32) //未成功发放的道具

	if moneyType == common.MT_Brick {
		ret, billno = mgr.user.consumeBalance(param, amount, common.RS_BuyGoods)
	} else {
		mgr.reduceMoney(moneyType, common.RS_BuyGoods, uint64(amount))
	}

	if ret != 0 {
		return BuyGoodsFail
	}

	// 添加商品
	if !mgr.GetGoods(id, num, common.RS_BuyGoods, moneyType, amount) {
		failGoods[id] = num
		mgr.BuyRollback(billno, moneyType, amount, succGoods, failGoods, param)
		return BuyGoodsFail
	}

	succGoods[id] = num
	mgr.user.BuyGoodsFlow(billno, moneyType, amount, succGoods, failGoods, 1, 0)
	mgr.user.Info("BuyGoods success, id: ", id)

	return BuyGoodsSuccess
}

// IsMoneyEnough 购买商品扣除货币
func (mgr *StoreMgr) IsMoneyEnough(moneyType, amount uint32) bool {
	var haveRes uint64
	switch moneyType {
	case common.MT_MONEY:
		haveRes = mgr.user.GetCoin()
	case common.MT_DIAMOND:
		haveRes = mgr.user.GetDiam()
	case common.MT_BraveCoin:
		haveRes = mgr.user.GetBraveCoin()
	case common.MT_ComradePoints:
		haveRes = mgr.user.GetComradePoints()
	case common.MT_Brick:
		haveRes = mgr.user.GetBrick()
	}
	return uint32(haveRes) >= amount
}

// NewBuyGoods 购买商品
func (mgr *StoreMgr) NewBuyGoods(channel, tracks, id, priceNum, roundType uint32, param []byte) uint32 {

	var goodID, goodNum uint32
	var moneyType, moneyID, moneyNum, needMoney uint32

	var ret uint32                          //金砖购买返回码
	var billno string                       //金砖购买订单号
	var succGoods = make(map[uint32]uint32) //已成功发放的道具
	var failGoods = make(map[uint32]uint32) //未成功发放的道具

	//购买渠道
	switch channel {
	case GoodChannel_Normal:
		sellGoods, ok := excel.GetSelling(uint64(id))
		if !ok {
			mgr.user.Error("Getselling failed, Getselling doesn't exist, id: ", id)
			return BuyGoodsFail
		}

		goodID, goodNum, moneyID, needMoney = mgr.getSaleGoodInfo(tracks, priceNum, sellGoods.Goods, sellGoods.Discount1, sellGoods.Discount2)
	case GoodChannel_Special:
		if !mgr.judgeSaleGoodRound(id, roundType) {
			mgr.user.Error("judgeSaleGoodRound failed")
			return BuyGoodsFail
		}

		sellGoods, ok := excel.GetGoods(uint64(id))
		if !ok {
			mgr.user.Error("GetGoods failed, GetGoods doesn't exist, id: ", id)
			return BuyGoodsFail
		}

		goodID, goodNum, moneyID, needMoney = mgr.getSaleGoodInfo(tracks, priceNum, sellGoods.Goods, sellGoods.Discount1, sellGoods.Discount2)
	case GoodChannel_SeasonPass:
		sellGoods, ok := excel.GetSeasontik(uint64(id))
		if !ok {
			mgr.user.Error("Getselling failed, Getselling doesn't exist, id: ", id)
			return BuyGoodsFail
		}

		discount := strings.Split(sellGoods.Discount, ";")
		if len(discount) != 2 {
			mgr.user.Error("GetSeasontik sellGoods.Goods err discount:", discount)
			return BuyGoodsFail
		}

		moneyID = common.StringToUint32(discount[0])
		needMoney = common.StringToUint32(discount[1])

		if needMoney == 0 {
			return BuyGoodsFail
		}

		moneyNum, moneyType = mgr.getOwnMoneyNum(moneyID)
		if moneyNum < needMoney {
			mgr.user.Error("NewBuyGoods failed, coins is not enough, moneyNum: ", moneyNum, " needMoney:", needMoney, " moneyID:", moneyID)
			return GoldCoinLack
		}

		// 预扣除
		if moneyType == common.MT_Brick {
			ret, billno = mgr.user.consumeBalance(param, uint32(needMoney), common.RS_BuyGoods)
		} else {
			mgr.reduceMoney(moneyType, common.RS_BuyGoods, uint64(needMoney))
		}

		if ret != 0 {
			return BuyGoodsFail
		}

		goods := strings.Split(sellGoods.Goods, "|")
		for _, v := range goods {
			good := strings.Split(v, ";")
			if len(good) != 2 {
				mgr.user.Error("GetSeasontik sellGoods.Goods err good:", good)
				continue
			}

			goodID = common.StringToUint32(good[0])
			goodNum = common.StringToUint32(good[1])

			if mgr.GetGoods(goodID, goodNum, common.RS_BuyGoods, moneyType, needMoney) {
				succGoods[goodID] = goodNum
			} else {
				failGoods[goodID] = goodNum
			}
		}

		if len(failGoods) > 0 {
			mgr.BuyRollback(billno, moneyType, uint32(needMoney), succGoods, failGoods, param)
			return BuyGoodsFail
		}

		mgr.user.BuyGoodsFlow(billno, moneyType, uint32(needMoney), succGoods, failGoods, 1, 0)
		return BuyGoodsSuccess

	default:
		mgr.user.Error("channel fail! channel:", channel)
		return BuyGoodsFail
	}

	// 商品信息
	goodsConfig, ok := excel.GetStore(uint64(goodID))
	if !ok {
		mgr.user.Error("NewBuyGoods failed, goods does't exist, goodID: ", goodID)
		return GoodsNotExist
	}

	// 判断是否可重复购买
	if !mgr.repeatBuyGoodsCheck(goodID) {
		return BuyGoodsFail
	}

	if needMoney == 0 {
		return BuyGoodsFail
	}

	// 购买赛季战阶商品时，判断赛季战阶是否已打满级，如果达满级购买失败
	if goodsConfig.Type == GoodsSeasonExpMedal {
		if mgr.user.taskMgr.getChallengeTask().checkSeasonMaxGrade() {
			mgr.user.AdviceNotify(common.NotifyCommon, 98)
			return BuyGoodsFail
		}
	}

	// 预扣除
	moneyNum, moneyType = mgr.getOwnMoneyNum(moneyID)
	if moneyNum < needMoney {
		mgr.user.Error("NewBuyGoods failed, coins is not enough, moneyNum: ", moneyNum, " needMoney:", needMoney, " moneyID:", moneyID)
		return GoldCoinLack
	}

	if moneyType == common.MT_Brick {
		ret, billno = mgr.user.consumeBalance(param, uint32(needMoney), common.RS_BuyGoods)
	} else {
		mgr.reduceMoney(moneyType, common.RS_BuyGoods, uint64(needMoney))
	}

	if ret != 0 {
		return BuyGoodsFail
	}

	// 添加商品
	if !mgr.GetGoods(goodID, goodNum, common.RS_BuyGoods, moneyType, needMoney) {
		failGoods[goodID] = goodNum
		mgr.BuyRollback(billno, moneyType, uint32(needMoney), succGoods, failGoods, param)
		return BuyGoodsFail
	}

	succGoods[goodID] = goodNum
	mgr.user.BuyGoodsFlow(billno, moneyType, uint32(needMoney), succGoods, failGoods, 1, 0)

	mgr.user.Info("NewBuyGoods success, goodID: ", goodID)
	return BuyGoodsSuccess
}

// BuyRollback 发货失败时交易回滚
func (mgr *StoreMgr) BuyRollback(billno string, moneyType, amount uint32, succGoods, failGoods map[uint32]uint32, param []byte) {
	if len(failGoods) == 0 {
		return
	}

	var ret uint32
	if moneyType == common.MT_Brick {
		ret = mgr.user.cancelPay(param, billno, amount, common.RS_RollBack)
	} else {
		mgr.addMoney(moneyType, common.RS_RollBack, amount)
	}

	if ret != 0 {
		mgr.user.BuyGoodsFlow(billno, moneyType, amount, succGoods, failGoods, 0, 0)
		return
	}

	mgr.ReduceGoodsAll(succGoods, common.RS_RollBack)
	mgr.user.BuyGoodsFlow(billno, moneyType, amount, succGoods, failGoods, 0, 1)
}

// getOwnMoneyNum 通过商品ID 获取货币拥有的 数量和类型
func (mgr *StoreMgr) getOwnMoneyNum(moneyID uint32) (moneyNum, moneyType uint32) {
	//购买货币
	switch moneyID {
	case common.Item_Coin:
		moneyNum = uint32(mgr.user.GetCoin())
		moneyType = common.MT_MONEY
	case common.Item_BraveCoin:
		moneyNum = uint32(mgr.user.GetBraveCoin())
		moneyType = common.MT_BraveCoin
	case common.Item_Diam:
		moneyNum = uint32(mgr.user.GetDiam())
		moneyType = common.MT_DIAMOND
	case common.Item_ComradePoints:
		moneyNum = uint32(mgr.user.GetComradePoints())
		moneyType = common.MT_ComradePoints
	case common.Item_Exp:
		moneyNum = mgr.user.GetExp()
		moneyType = common.MT_EXP
	case common.Item_Brick:
		moneyNum = uint32(mgr.user.GetBrick())
		moneyType = common.MT_Brick
	default:
		mgr.user.Error("moneyID fail! moneyID:", moneyID)
	}

	return
}

// judgeSaleGoodRound 判断购买的特惠商品是否在此轮中出售
func (mgr *StoreMgr) judgeSaleGoodRound(id, roundType uint32) bool {
	round7, round1, _, _ := mgr.getSaleGoodRound()

	oddsMap := excel.GetOddsMap()
	for _, v := range oddsMap {
		if v.Type != uint64(roundType) {
			continue
		}

		switch roundType {
		case SaleType_Day1:
			if v.Turn != uint64(round1) {
				continue
			}
		case SaleType_Day7:
			if v.Turn != uint64(round7) {
				continue
			}
		default:
			return false
		}

		groups := strings.Split(v.Group, ";")
		for _, j := range groups {
			if id == common.StringToUint32(j) {
				return true
			}
		}
	}

	return false
}

// getSaleGoodInfo 获取购买商品的信息(物品id，价格，数量)
func (mgr *StoreMgr) getSaleGoodInfo(tracks, priceNum uint32, sellGoodsGoods, sellGoodsDiscount1, sellGoodsDiscount2 string) (goodID, goodNum, moneyID, needMoney uint32) {

	goods := strings.Split(sellGoodsGoods, "|")
	if len(goods) != 3 {
		mgr.user.Error("sellGoods.Goods err sellGoodsGoods:", sellGoodsGoods)
		return 0, 0, 0, 0
	}

	// 获取购买物品的价格
	var discount string
	if priceNum == 1 {
		discount = sellGoodsDiscount1
	} else if priceNum == 2 {
		discount = sellGoodsDiscount2
	}
	prices := strings.Split(discount, "|")
	if len(prices) != 3 {
		mgr.user.Error("sellGoods.Goods err Discount:", discount)
		return 0, 0, 0, 0
	}

	// 获取购买物品的id和数量
	switch tracks {
	case GoodType_7Days:
		good := strings.Split(goods[0], ";")
		if len(good) != 2 {
			mgr.user.Error("sellGoods.Goods err good:", good)
			return 0, 0, 0, 0
		}

		goodID = common.StringToUint32(good[0])
		goodNum = common.StringToUint32(good[1])

		price := strings.Split(prices[0], ";")
		if len(price) != 2 {
			mgr.user.Error("sellGoods.Goods err price:", price)
			return 0, 0, 0, 0
		}
		moneyID = common.StringToUint32(price[0])
		needMoney = common.StringToUint32(price[1])

	case GoodType_30Days:
		good := strings.Split(goods[1], ";")
		if len(good) != 2 {
			mgr.user.Error("sellGoods.Goods err good:", good)
			return 0, 0, 0, 0
		}

		goodID = common.StringToUint32(good[0])
		goodNum = common.StringToUint32(good[1])

		price := strings.Split(prices[1], ";")
		if len(price) != 2 {
			mgr.user.Error("sellGoods.Goods err price:", price)
			return 0, 0, 0, 0
		}
		moneyID = common.StringToUint32(price[0])
		needMoney = common.StringToUint32(price[1])

	case GoodType_Forever:
		good := strings.Split(goods[2], ";")
		if len(good) != 2 {
			mgr.user.Error("sellGoods.Goods err good:", good)
			return 0, 0, 0, 0
		}

		goodID = common.StringToUint32(good[0])
		goodNum = common.StringToUint32(good[1])

		price := strings.Split(prices[2], ";")
		if len(price) != 2 {
			mgr.user.Error("sellGoods.Goods err price:", price)
			return 0, 0, 0, 0
		}
		moneyID = common.StringToUint32(price[0])
		needMoney = common.StringToUint32(price[1])

	default:
		mgr.user.Error("tracks fail! tracks:", tracks)
		return 0, 0, 0, 0
	}

	return
}

// getSaleGoodRound 获取出售商品的轮数 (应该周轮数；应该天轮数；实际周轮数；实际天轮数)
func (mgr *StoreMgr) getSaleGoodRound() (uint32, uint32, uint32, uint32) {
	var roundMax7, roundMax1, round7, round1, realRound7, realRound1 uint32

	roundMax7 = uint32(common.GetTBSystemValue(common.System_PrivilegeGoodMaxRound7))
	roundMax1 = uint32(common.GetTBSystemValue(common.System_PrivilegeGoodMaxRound1))

	startTime := float64(common.GetTBSystemValue(common.System_PrivilegeGoodStartTime))
	interTime7 := float64(common.GetTBSystemValue(common.System_PrivilegeGoodInterTime7))
	interTime1 := float64(common.GetTBSystemValue(common.System_PrivilegeGoodInterTime1))

	nowTime := float64(time.Now().Unix())

	if nowTime < startTime {
		return 0, 0, 0, 0
	}

	if nowTime == startTime {
		return 1, 1, 1, 1
	}

	round7 = uint32(math.Ceil((nowTime - startTime) / interTime7))
	round1 = uint32(math.Ceil((nowTime - startTime) / interTime1))

	realRound7 = round7
	realRound1 = round1

	if round7 > roundMax7 {
		round7 = roundMax7
	}
	if round1 > roundMax1 {
		round1 = roundMax1
	}

	return round7, round1, realRound7, realRound1
}

// 获得商品
func (mgr *StoreMgr) GetGoods(id, num, reason, iMoneyType, iMoneyNum uint32) bool {
	if num == 0 {
		return false
	}

	// 商品信息
	goodsConfig, ok := excel.GetStore(uint64(id))
	if !ok {
		mgr.user.Error("GetGoods failed, goods doesn't exist, id: ", id)
		return false
	}

	if goodsConfig.Type == GoodsGiftPack && goodsConfig.UseType == 2 {
		return mgr.getGiftPack(id, num, reason, iMoneyType, iMoneyNum)
	}

	if isMoneyGoods(uint32(goodsConfig.Type)) {
		return mgr.GetMoneyGoods(num, reason, uint32(goodsConfig.Type))
	}

	isNew := false
	util := db.PlayerGoodsUtil(mgr.user.GetDBID())

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
		//if isReplaceWay != 0 {
		//	isReplace = true
		//}
	} else {
		//if isReplaceWay == 2 {
		//	isReplace = true
		//}
	}

	switch info.Id {
	case common.Item_SeasonHeroMedal:
		{
			_, end := common.GetRankSeasonTimeStamp()
			info.EndTime = end
			isNew = true
		}
	default:
		timeLimit := getGoodsTotalTime(id) * int64(num)

		if info.Time == 0 { //首次获得该道具
			info.Time = time.Now().Unix()
			if timeLimit > 0 {
				info.EndTime = time.Now().Unix() + timeLimit
			} else {
				isNew = true
			}
		} else { //已拥有该道具
			if timeLimit > 0 {
				if info.EndTime > 0 {
					info.EndTime += timeLimit //限时道具时效叠加
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
	mgr.rareItem(id, num)
	mgr.user.activityMgr.updateExchangeOwnNum(dbid)

	// 通知客户端
	mgr.user.RPC(iserver.ServerTypeClient, "AddGoods", info.ToProto())
	if goodsConfig.Timelimit == "" && isNew {
		if goodsConfig.Type == GoodsRoleType {
			mgr.user.AddAchievementData(common.AchievementRole, 1)
		}
		if goodsConfig.Type == GoodsParachuteType {
			mgr.user.AddAchievementData(common.AchievementParachute, 1)
		}
	}

	// 特殊道具处理
	mgr.specialGoodsProc(dbid)

	if info.EndTime != 0 {
		mgr.user.StartCrondForExpireCheck()
	}

	if reason != common.RS_BuyGoods {
		iMoneyNum = 0
	}

	var leftTime uint32
	if info.EndTime > 0 {
		leftTime = uint32(info.EndTime - time.Now().Unix())
	}

	mgr.user.tlogItemFlow(id, num, reason, iMoneyNum, iMoneyType, ADD, 0, leftTime) //tlog道具流水表
	mgr.user.Info("GetGoods success, id: ", id, " num: ", num)
	return true
}

// getGiftPack 获得礼包道具
func (mgr *StoreMgr) getGiftPack(id, num, reason, iMoneyType, iMoneyNum uint32) bool {
	if num == 0 {
		return false
	}

	// 商品信息
	goodsConfig, ok := excel.GetStore(uint64(id))
	if !ok {
		mgr.user.Error("getGiftPack failed, goods doesn't exist, id: ", id)
		return false
	}

	for i := uint32(0); i < num; i++ {
		rewards1 := common.GetBoxReward1(uint32(goodsConfig.RelationID))
		for k, v := range rewards1 {
			mgr.GetGoods(k, v, reason, iMoneyType, iMoneyNum)
		}

		rewards2 := common.GetBoxReward2(uint32(goodsConfig.RelationID))
		for k, v := range rewards2 {
			mgr.GetGoods(k, v, reason, iMoneyType, iMoneyNum)
		}

		mgr.user.RPC(iserver.ServerTypeClient, "GetGiftPackNotify", &protoMsg.Uint32Array{
			List: common.MapUint32ToSlice(common.CombineMapUint32(rewards1, rewards2)),
		})
	}

	mgr.user.Info("getGiftPack, id: ", id, " num: ", num, " reason: ", reason, " iMoneyType: ", iMoneyType)
	return true
}

// specialGoodsProc 特殊道具的处理
func (mgr *StoreMgr) specialGoodsProc(id uint32) {
	switch id {
	case common.Item_SeasonHeroMedal:
		mgr.user.taskMgr.getChallengeTask().enableSeasonElite()
		mgr.user.taskMgr.getChallengeTask().syncTaskDetail()
	}

	if id == common.GetNameColorItem() {
		mgr.user.friendMgr.syncFriendName()
	}

	// 检测并激活任务组
	mgr.ItemEnableTaskGroup(id)
}

// ItemEnableTaskGroup 获得道具激活任务组
func (mgr *StoreMgr) ItemEnableTaskGroup(id uint32) {
	GetSrvInst().groupEnableItems.Range(
		func(key, value interface{}) bool {
			name := key.(string)
			items := value.([]uint32)

			var exist bool
			for _, v := range items {
				if v == id {
					exist = true
					break
				}
			}

			if exist {
				mgr.user.taskMgr.enableTaskGroupByType(name)
			}

			return true
		})
}

// GetMoneyGoods 获得货币类型的道具
func (mgr *StoreMgr) GetMoneyGoods(num, reason, goodsType uint32) bool {
	switch goodsType {
	//金币
	case GoodsGoldCoin:
		mgr.addMoney(common.MT_MONEY, reason, num)
	//钻石
	case GoodsDiamond:
		mgr.addMoney(common.MT_DIAMOND, reason, num)
	//活跃度
	case GoodsActiveness:
		day := mgr.user.updateActivenessProgress(1, num)
		week := mgr.user.updateActivenessProgress(2, num)
		mgr.user.ActivenessFlow(day, week)
	//经验值
	case GoodsExp:
		mgr.user.UpdateMilitaryRank(num)
	//战友积分
	case GoodsComradePoints:
		mgr.addMoney(common.MT_ComradePoints, reason, num)
	//勇气值
	case GoodsBraveCoin:
		mgr.addMoney(common.MT_BraveCoin, reason, num)
	//特训经验勋章
	case GoodsSpecailExpMedal:
		mgr.user.taskMgr.getSpecialTask().updateSpecialAwards(num)
	//赛季经验勋章
	case GoodsSeasonExpMedal:
		mgr.user.taskMgr.getChallengeTask().updateSeasonGrade(num)
	}
	return true
}

// limitAddCoin 增加金币，限制每日上限
func (mgr *StoreMgr) limitAddCoin(num, reason uint32) {
	var get uint64
	var temp uint64

	max := uint64(common.GetTBSystemValue(common.System_DayCoinLimit))

	today := time.Now().Format("2006-01-02")
	last := time.Unix(int64(mgr.user.GetDayCoinGetTime()), 0).Format("2006-01-02")

	if today == last {
		temp = mgr.user.GetDayCoinGet()
		if temp == max {
			mgr.user.AdviceNotify(common.NotifyCommon, 65) //提示今日金币达到上限
			return
		}

		if max-temp <= uint64(num) {
			get = max
		} else {
			get = temp + uint64(num)
		}
	} else {
		if uint64(num) >= max {
			get = max
		} else {
			get = uint64(num)
		}
	}

	mgr.user.SetDayCoinGet(get)
	mgr.user.SetDayCoinGetTime(uint64(time.Now().Unix()))

	mgr.addMoney(common.MT_MONEY, reason, uint32(get-temp))
}

// limitAddDiam 增加钻石，限制每日上限
func (mgr *StoreMgr) limitAddDiam(num, reason uint32) {
	var get uint64
	var temp uint64

	max := uint64(common.GetTBSystemValue(common.System_DayDiamLimit))

	today := time.Now().Format("2006-01-02")
	last := time.Unix(int64(mgr.user.GetDayDiamGetTime()), 0).Format("2006-01-02")

	if today == last {
		temp = mgr.user.GetDayDiamGet()
		if temp == max {
			mgr.user.AdviceNotify(common.NotifyCommon, 64) //提示今日金币达到上限
			return
		}

		if max-temp <= uint64(num) {
			get = max
		} else {
			get = temp + uint64(num)
		}
	} else {
		if uint64(num) >= max {
			get = max
		} else {
			get = uint64(num)
		}
	}

	mgr.user.SetDayDiamGet(get)
	mgr.user.SetDayDiamGetTime(uint64(time.Now().Unix()))

	mgr.addMoney(common.MT_DIAMOND, reason, uint32(get-temp))
}

// addMoney 增加货币
func (mgr *StoreMgr) addMoney(kind uint32, method uint32, count uint32) {
	if count == 0 {
		return
	}
	var now uint64
	switch kind {
	case common.MT_MONEY:
		mgr.user.SetCoin(mgr.user.GetCoin() + uint64(count))
		now = mgr.user.GetCoin()

		mgr.user.AddAchievementData(common.AchievementCoin, count)
		mgr.user.activityMgr.updateExchangeOwnNum(common.Item_Coin)
	case common.MT_BraveCoin:
		mgr.user.SetBraveCoin(mgr.user.GetBraveCoin() + uint64(count))
		now = mgr.user.GetBraveCoin()

		mgr.user.activityMgr.updateExchangeOwnNum(common.Item_BraveCoin)
	case common.MT_DIAMOND:
		mgr.user.SetDiam(mgr.user.GetDiam() + uint64(count))
		now = mgr.user.GetDiam()

		mgr.user.activityMgr.updateExchangeOwnNum(common.Item_Diam)
	case common.MT_ComradePoints:
		mgr.user.SetComradePoints(mgr.user.GetComradePoints() + uint64(count))
		now = mgr.user.GetComradePoints()

		mgr.user.activityMgr.updateExchangeOwnNum(common.Item_ComradePoints)
	case common.MT_EXP:
		mgr.user.UpdateMilitaryRank(count)
		now = uint64(mgr.user.GetExp())

		mgr.user.activityMgr.updateExchangeOwnNum(common.Item_Exp)
	default:
		return
	}

	mgr.user.tlogMoneyFlow(count, method, ADD, kind, now) //tlog货币流水表
}

// reduceMoney 消耗玩家货币
func (mgr *StoreMgr) reduceMoney(kind uint32, method uint32, count uint64) {
	var deductCoin uint64
	switch kind {
	case common.MT_MONEY: // 扣除金币
		deductCoin = mgr.user.GetCoin() - count
		mgr.user.SetCoin(deductCoin)
		mgr.user.updateDayTaskItems(common.TaskItem_CostCoin, uint32(count))
		mgr.user.taskMgr.updateTaskItemsAll(common.TaskItem_CostCoin, uint32(count))
	case common.MT_BraveCoin: //勇气值
		deductCoin = mgr.user.GetBraveCoin() - count
		mgr.user.SetBraveCoin(deductCoin)
	case common.MT_DIAMOND: //钻石
		deductCoin = mgr.user.GetDiam() - count
		mgr.user.SetDiam(deductCoin)
		mgr.user.updateDayTaskItems(common.TaskItem_CostDiam, uint32(count))
		mgr.user.taskMgr.updateTaskItemsAll(common.TaskItem_CostDiam, uint32(count))
	case common.MT_ComradePoints: //战友积分
		deductCoin = mgr.user.GetComradePoints() - count
		mgr.user.SetComradePoints(deductCoin)
	case common.MT_EXP: //经验值
		deductCoin = uint64(mgr.user.GetExp()) - count
		mgr.user.SetExp(uint32(deductCoin))
	default:
		return
	}
	mgr.user.tlogMoneyFlow(uint32(count), method, REDUCE, kind, deductCoin) //tlog货币流水表
}

// GetAwards 获取奖励
func (mgr *StoreMgr) GetAwards(awards map[uint32]uint32, reason uint32, limitCoin, limitDiam bool) {
	for id, num := range awards {
		switch id {
		//金币奖励
		case common.Item_Coin:
			if limitCoin {
				mgr.limitAddCoin(num, reason)
			} else {
				mgr.addMoney(common.MT_MONEY, reason, num)
			}
		//钻石奖励
		case common.Item_Diam:
			if limitDiam {
				mgr.limitAddDiam(num, reason)
			} else {
				mgr.addMoney(common.MT_DIAMOND, reason, num)
			}
		//道具奖励
		default:
			mgr.GetGoods(id, num, reason, common.MT_NO, 0)
		}
	}
}

// rareItem 获得稀有道具通知客户端并且数量库累加
func (mgr *StoreMgr) rareItem(id, incr uint32) {
	goodsConfig, ok := excel.GetStore(uint64(id))
	if !ok {
		mgr.user.Error("GetGoods failed, goods doesn't exist, id: ", id)
		return
	}
	systemData, ok := excel.GetSystem(88)
	if !ok {
		return
	}
	if goodsConfig.Raretype <= systemData.Value {
		return
	}

	util := db.PlayerInfoUtil(mgr.user.GetDBID())
	num, err := util.GetRateItemNum()
	if err != nil {
		mgr.user.Error("rareItem GetRateItemNum err:", err)
		return
	}
	num += incr
	util.SetRateItemNum(num)

	if err := mgr.user.RPC(iserver.ServerTypeClient, "RareGoodShare", id, num); err != nil {
		mgr.user.Error(err)
	}

	// 稀有物品公告
	days := uint32(getGoodsTotalTime(id) / (24 * 60 * 60))
	mgr.user.announcementBroad(2, id, days, incr)
}

// IsAllGoodsEnough 判断是否所有商品都足够
func (mgr *StoreMgr) IsAllGoodsEnough(goods map[uint32]uint32) bool {
	util := db.PlayerGoodsUtil(mgr.user.GetDBID())

	for id, num := range goods {
		item, ok := excel.GetStore(uint64(id))
		if !ok {
			return false
		}

		switch item.RelationID {
		//金币
		case common.Item_Coin:
			if mgr.user.GetCoin() < uint64(num) {
				return false
			}
		//钻石
		case common.Item_Diam:
			if mgr.user.GetDiam() < uint64(num) {
				return false
			}
		//战友积分
		case common.Item_ComradePoints:
			if mgr.user.GetComradePoints() < uint64(num) {
				return false
			}
		//勇气值
		case common.Item_BraveCoin:
			if mgr.user.GetBraveCoin() < uint64(num) {
				return false
			}
		//道具
		default:
			if !util.IsGoodsEnough(uint32(item.RelationID), num) {
				return false
			}
		}
	}

	return true
}

// ReduceGoods 扣除物品
func (mgr *StoreMgr) ReduceGoods(id, num, reason uint32) bool {
	util := db.PlayerGoodsUtil(mgr.user.GetDBID())

	info, err := util.GetGoodsInfo(id)
	if err != nil || info == nil {
		mgr.user.Error("GetGoodsInfo err: ", err, " or goods not exist, id: ", id)
		return false
	}
	if info.Sum < num {
		return false
	}
	if info.Sum > num {
		info.Sum -= num
		util.AddGoodsInfo(info)
	} else {
		info.Sum = 0
		util.DelGoods(id)
	}

	// 通知客户端
	mgr.user.RPC(iserver.ServerTypeClient, "UpdateGoods", info.ToProto())

	mgr.user.tlogItemFlow(info.Id, num, 0, 0, 0, REDUCE, reason, 0) //tlog道具流水表
	return true
}

// ReduceGoodsAll 扣除多种物品
func (mgr *StoreMgr) ReduceGoodsAll(goods map[uint32]uint32, reason uint32) {
	for id, num := range goods {
		item, ok := excel.GetStore(uint64(id))
		if !ok {
			continue
		}

		switch item.RelationID {
		//金币
		case common.Item_Coin:
			mgr.reduceMoney(common.MT_MONEY, reason, uint64(num))
		//钻石
		case common.Item_Diam:
			mgr.reduceMoney(common.MT_DIAMOND, reason, uint64(num))
		//战友积分
		case common.Item_ComradePoints:
			mgr.reduceMoney(common.MT_ComradePoints, reason, uint64(num))
		//勇气值
		case common.Item_BraveCoin:
			mgr.reduceMoney(common.MT_BraveCoin, reason, uint64(num))
		//道具
		default:
			mgr.ReduceGoods(uint32(item.RelationID), num, reason)
		}
	}
}

// UpdateGoodsState 更新物品状态
func (mgr *StoreMgr) updateGoodsState(id uint32, state uint32) {

	// 获取数据库信息
	info, err := db.PlayerGoodsUtil(mgr.user.GetDBID()).GetGoodsInfo(id)
	if err != nil || info == nil {
		mgr.user.Error("GetGoodsInfo err: ", err, " or goods not exist, id: ", id)
		return
	}

	// 更改数据库状态
	info.State = state
	result := db.PlayerGoodsUtil(mgr.user.GetDBID()).AddGoodsInfo(info)
	if !result {
		return
	}

	// 通知客户端
	mgr.user.RPC(iserver.ServerTypeClient, "UpdateGoods", info.ToProto())

}

// 根据配置，获取商品的使用时限
func getGoodsTotalTime(id uint32) int64 {
	goodsConfig, ok := excel.GetStore(uint64(id))
	if !ok {
		return 0
	}

	if goodsConfig.Timelimit == "" {
		return 0
	}

	strs := strings.Split(goodsConfig.Timelimit, "|")
	if len(strs) != 4 {
		return 0
	}

	total := common.StringToInt64(strs[0]) * 24 * 60 * 60
	total += common.StringToInt64(strs[1]) * 60 * 60
	total += common.StringToInt64(strs[2]) * 60
	total += common.StringToInt64(strs[3])

	return total
}

// itemRegularCheck 道具定时检测
func (mgr *StoreMgr) itemRegularCheck() {
	var dels []uint32

	for _, info := range db.PlayerGoodsUtil(mgr.user.GetDBID()).GetAllGoodsInfo() {
		if info.EndTime == 0 {
			continue
		}

		if info.EndTime > time.Now().Unix() {
			continue
		}

		dels = append(dels, info.Id)
	}

	if len(dels) > 0 {
		mgr.DelExpiredGoods(dels)
		for _, v := range dels {
			mgr.user.RPC(iserver.ServerTypeClient, "ExpireGoodId", v)
		}
	}
}

//删除过期的商品
func (mgr *StoreMgr) DelExpiredGoods(dels []uint32) {
	util := db.PlayerGoodsUtil(mgr.user.GetDBID())
	for _, id := range dels {
		goodInfo, _ := util.GetGoodsInfo(id)
		if goodInfo == nil {
			continue
		}

		util.DelGoods(id)
		mgr.user.tlogItemFlow(id, 1, 0, 0, 0, REDUCE, 1, 0) //tlog道具流水表
		mgr.delPreferenceItem(id)
		mgr.user.skillExpire(id) //处理过期技能

		if id == common.GetNameColorItem() {
			mgr.user.friendMgr.syncFriendName()
		}

		goodsConfig, ok := excel.GetStore(uint64(id))
		if !ok {
			mgr.user.Error("Goods does't exist, id: ", id)
			continue
		}

		//玩家正在使用的伞过期，替换为默认伞
		if mgr.user.GetParachuteID() == uint32(goodsConfig.RelationID) {
			mgr.user.SetParachuteID(uint32(common.GetTBSystemValue(common.System_InitParachuteID)))
			mgr.user.SetGoodsParachuteID(uint32(common.GetTBSystemValue(common.System_InitParachuteID)))
		}

		//玩家正在使用的角色过期，替换为默认角色，并通知Match服，在team内广播
		if mgr.user.GetRoleModel() == uint32(goodsConfig.RelationID) {
			mgr.user.SetRoleModel(uint32(common.GetTBSystemValue(32)))
			mgr.user.SetGoodsRoleModel(uint32(common.GetTBSystemValue(32)))

			if mgr.user.teamMgrProxy != nil && mgr.user.GetTeamID() != 0 {
				if err := mgr.user.teamMgrProxy.RPC(common.ServerTypeMatch, "SetRoleModel",
					mgr.user.GetID(), mgr.user.GetTeamID(), uint32(common.GetTBSystemValue(32))); err != nil {
					mgr.user.Error("RPC SetRoleModel err: ", err)
				}
			}
		}

		//玩家在大厅展示的枪支到期
		if mgr.user.GetOutsideWeapon() == uint32(goodsConfig.RelationID) {
			mgr.user.SetOutsideWeapon(0)

			if mgr.user.teamMgrProxy != nil && mgr.user.GetTeamID() != 0 {
				if err := mgr.user.teamMgrProxy.RPC(common.ServerTypeMatch, "SetOutsideWeapon",
					mgr.user.GetID(), mgr.user.GetTeamID(), uint32(0)); err != nil {
					mgr.user.Error("RPC SetOutsideWeapon err: ", err)
				}
			}
		}

		if goodInfo.Used != 1 {
			continue
		}

		itemData, ok := excel.GetItem(goodsConfig.RelationID)
		if !ok {
			continue
		}
		if goodsConfig.Type == 15 {
			mgr.user.SetTheme(0)
		}
		//幻化道具过期 清理
		if goodsConfig.Type == GoodsGameEquipHead || goodsConfig.Type == GoodsGameEquipPack {
			var prop *protoMsg.WearInGame

			switch itemData.Type {
			case 12: //头盔
				prop = mgr.user.GetHeadWearInGame()
			case 16: //背包
				prop = mgr.user.GetPackWearInGame()
			}

			if prop == nil {
				continue
			}

			if itemData.Subtype == 1 {
				prop.First = 0
			}

			if itemData.Subtype == 2 {
				prop.Second = 0
			}

			if itemData.Subtype == 3 {
				prop.Third = 0
			}

			mgr.user.WearInGameFlow(uint32(itemData.Id), 0)

			switch itemData.Type {
			case 12: //头盔
				mgr.user.SetHeadWearInGame(prop)
				mgr.user.SetHeadWearInGameDirty()
			case 16: //背包
				mgr.user.SetPackWearInGame(prop)
				mgr.user.SetPackWearInGameDirty()
			}
		}

		//枪支装备道具过期
		if goodsConfig.Type == GoodsGameEquipGunSkin {
			weapons := mgr.user.GetWeaponEquipInGame()
			if weapons == nil {
				continue
			}

			for _, v := range weapons.GetWeapons() {
				if v.GetWeaponId() == uint32(itemData.GunCustom) {
					for i, j := range v.GetAdditions() {
						if j == id {
							v.Additions = append(v.Additions[:i], v.Additions[i+1:]...)
						}
					}
				}
			}

			mgr.user.SetWeaponEquipInGame(weapons)
			mgr.user.SetWeaponEquipInGameDirty()
			mgr.user.syncWeaponEquipment()
		}
	}
}

// UpdateGoodsPreference 更新物品偏好
func (mgr *StoreMgr) updateGoodsPreference(id uint32, preference uint32) {

	// 获取数据库信息
	info, err := db.PlayerGoodsUtil(mgr.user.GetDBID()).GetGoodsInfo(id)
	if err != nil || info == nil {
		mgr.user.Error("GetGoodsInfo err: ", err, " or goods not exist, id: ", id)
		return
	}

	// 更改数据库状态
	info.Preference = preference
	result := db.PlayerGoodsUtil(mgr.user.GetDBID()).AddGoodsInfo(info)
	if !result {
		return
	}

	// 通知客户端
	mgr.user.RPC(iserver.ServerTypeClient, "UpdateGoods", info.ToProto())
}

// delPreferenceItem 删除偏好列表中的偏好物品
func (mgr *StoreMgr) delPreferenceItem(id uint32) {
	info := &db.PreferenceInfo{}
	info.Start = make(map[uint32]map[uint32]bool)
	info.PreType = make(map[uint32]map[uint32]uint32)
	info.MayaItemList = make(map[uint32]map[uint32][]uint32)

	util := db.PlayerInfoUtil(mgr.user.GetDBID())
	if err := util.GetPreferenceList(info); err != nil {
		mgr.user.Error("preItemRandomCreate GetPreferenceList err:", err)
		return
	}

	goodsConfig, ok := excel.GetStore(uint64(id))
	if !ok {
		mgr.user.Error("Goods does't exist, id:", id)
		return
	}

	var itemlist []uint32
	switch goodsConfig.Type {
	case 1, 2: //1:角色，2：伞包
		for _, v := range info.MayaItemList[uint32(goodsConfig.Type)][1] {
			if v == id {
				continue
			}
			itemlist = append(itemlist, v)
		}

		mayaItemList := make(map[uint32][]uint32)
		mayaItemList[1] = itemlist
		info.MayaItemList[uint32(goodsConfig.Type)] = mayaItemList

		if info.Start[uint32(goodsConfig.Type)][1] && !mgr.checkOwnItemType(goodsConfig.Type, 1) {
			start := make(map[uint32]bool)
			start[1] = false
			info.Start[uint32(goodsConfig.Type)] = start

			preType := make(map[uint32]uint32)
			preType[1] = 0
			info.PreType[uint32(goodsConfig.Type)] = preType

			if err := mgr.user.RPC(iserver.ServerTypeClient, "OpenRandomPreferenceRsp", true, false, uint32(goodsConfig.Type), uint32(1)); err != nil {
				mgr.user.Error(err)
			}
			if err := mgr.user.RPC(iserver.ServerTypeClient, "AllOrPartPreSwitchRsp", true, uint32(0), uint32(goodsConfig.Type), uint32(1)); err != nil {
				mgr.user.Error(err)
			}
		} else if info.PreType[uint32(goodsConfig.Type)][1] == 1 && len(itemlist) == 0 {
			preType := make(map[uint32]uint32)
			preType[1] = 0
			info.PreType[uint32(goodsConfig.Type)] = preType

			if err := mgr.user.RPC(iserver.ServerTypeClient, "AllOrPartPreSwitchRsp", true, uint32(0), uint32(goodsConfig.Type), uint32(1)); err != nil {
				mgr.user.Error(err)
			}
		}

		if err := util.SetPreferenceList(info); err != nil {
			mgr.user.Error("RPC_SetItemPreference SetPreferenceList err:", err)
			return
		}
		mgr.user.preItemRandomCreate(uint32(goodsConfig.Type), 1) // 刷新
	case 5, 6: //5：背包，6:头盔
		itemData, ok := excel.GetItem(goodsConfig.RelationID)
		if !ok {
			mgr.user.Error("GetItem does't exist goodsConfig.RelationID:", goodsConfig.RelationID)
			return
		}

		for _, v := range info.MayaItemList[uint32(goodsConfig.Type)][uint32(itemData.Subtype)] {
			if v == id {
				continue
			}
			itemlist = append(itemlist, v)
		}

		mayaItemList := make(map[uint32][]uint32)
		for i := 1; i <= 3; i++ {
			if uint64(i) != itemData.Subtype {
				mayaItemList[uint32(i)] = info.MayaItemList[uint32(goodsConfig.Type)][uint32(i)]
			} else {
				mayaItemList[uint32(itemData.Subtype)] = itemlist
			}
		}
		info.MayaItemList[uint32(goodsConfig.Type)] = mayaItemList

		if info.Start[uint32(goodsConfig.Type)][uint32(itemData.Subtype)] && !mgr.checkOwnItemType(goodsConfig.Type, itemData.Subtype) {
			start := make(map[uint32]bool)
			for i := 1; i <= 3; i++ {
				if uint32(i) != uint32(itemData.Subtype) {
					start[uint32(i)] = info.Start[uint32(goodsConfig.Type)][uint32(i)]
				} else {
					start[uint32(itemData.Subtype)] = false
				}
			}
			info.Start[uint32(goodsConfig.Type)] = start

			preType := make(map[uint32]uint32)
			for i := 1; i <= 3; i++ {
				if uint32(i) != uint32(itemData.Subtype) {
					preType[uint32(i)] = info.PreType[uint32(goodsConfig.Type)][uint32(i)]
				} else {
					preType[uint32(itemData.Subtype)] = 0
				}
			}
			info.PreType[uint32(goodsConfig.Type)] = preType

			mgr.user.Debug("OpenRandomPreferenceRsp AllOrPartPreSwitchRsp")
			if err := mgr.user.RPC(iserver.ServerTypeClient, "OpenRandomPreferenceRsp", true, false, uint32(goodsConfig.Type), uint32(itemData.Subtype)); err != nil {
				mgr.user.Error(err)
			}
			if err := mgr.user.RPC(iserver.ServerTypeClient, "AllOrPartPreSwitchRsp", true, uint32(0), uint32(goodsConfig.Type), uint32(itemData.Subtype)); err != nil {
				mgr.user.Error(err)
			}
		} else if info.PreType[uint32(goodsConfig.Type)][uint32(itemData.Subtype)] == 1 && len(itemlist) == 0 {
			preType := make(map[uint32]uint32)
			for i := 1; i <= 3; i++ {
				if uint32(i) != uint32(itemData.Subtype) {
					preType[uint32(i)] = info.PreType[uint32(goodsConfig.Type)][uint32(i)]
				} else {
					preType[uint32(itemData.Subtype)] = 0
				}
			}
			info.PreType[uint32(goodsConfig.Type)] = preType

			if err := mgr.user.RPC(iserver.ServerTypeClient, "AllOrPartPreSwitchRsp", true, uint32(0), uint32(goodsConfig.Type), uint32(itemData.Subtype)); err != nil {
				mgr.user.Error(err)
			}
		}

		if err := util.SetPreferenceList(info); err != nil {
			mgr.user.Error("RPC_SetItemPreference SetPreferenceList err:", err)
			return
		}
		mgr.user.preItemRandomCreate(uint32(goodsConfig.Type), uint32(itemData.Subtype)) // 刷新
	default:
		mgr.user.Error("Type error: ", goodsConfig.Type)
		return
	}
}

// checkOwnItemType 检测是否拥有该类型商品
func (mgr *StoreMgr) checkOwnItemType(typ, level uint64) bool {
	for _, info := range db.PlayerGoodsUtil(mgr.user.GetDBID()).GetAllGoodsInfo() {
		if info == nil {
			continue
		}

		goodsConfig, ok := excel.GetStore(uint64(info.Id))
		if !ok {
			mgr.user.Error("GetStore does't exist, info.Id: ", info.Id)
			continue
		}
		if goodsConfig.State == 3 { //免费道具
			continue
		}

		switch typ {
		case 1, 2:
			if goodsConfig.Type == typ {
				return true
			}
		case 5, 6:
			itemData, ok := excel.GetItem(goodsConfig.RelationID)
			if !ok {
				mgr.user.Error("GetItem does't exist goodsConfig.RelationID:", goodsConfig.RelationID)
				continue
			}
			if goodsConfig.Type == typ && itemData.Subtype == level {
				return true
			}
		}

	}

	return false
}

// useMayaItem 使用道具 1 装备  0 卸下
func (mgr *StoreMgr) useMayaItem(goodsID uint32, state uint32) {
	// 商品信息
	goodsConfig, ok := excel.GetStore(uint64(goodsID))
	if !ok {
		mgr.user.Error(" goods doesn't exist, id: ", goodsID)
		return
	}

	// 获取数据库信息
	util := db.PlayerGoodsUtil(mgr.user.GetDBID())
	info, err := util.GetGoodsInfo(goodsID)
	if goodsConfig.Type == GoodsParachuteType {
		if info != nil || goodsConfig.State == GoodsFree {
			mgr.user.SetParachuteID(uint32(goodsConfig.RelationID))
			mgr.user.SetGoodsParachuteID(goodsID)
		}
	}
	if info == nil {
		mgr.user.Debug("GetGoodsInfo Error ", err)
		return
	}
	// 更改数据库状态
	info.Used = state
	result := util.AddGoodsInfo(info)
	if !result {
		return
	}

	// 通知客户端
	mgr.user.RPC(iserver.ServerTypeClient, "UpdateGoods", info.ToProto())

	//使用道具后的效果处理
	if goodsConfig.Type == GoodsGameEquipHead || goodsConfig.Type == GoodsGameEquipPack {
		mgr.useWearInGame(goodsConfig.RelationID, state)
	}
	if goodsConfig.Type == 15 {
		mgr.useThemeItem(goodsConfig.RelationID, state)
	}
}

//useThemeItem 使用场景道具
func (mgr *StoreMgr) useThemeItem(relationID uint64, state uint32) {
	if state == 1 {
		mgr.user.SetTheme(uint32(relationID))
	}
	if state == 0 {
		mgr.user.SetTheme(uint32(0))
	}
}

//useWearInGame 使用幻化道具的处理
func (mgr *StoreMgr) useWearInGame(relationID uint64, state uint32) {
	itemData, ok := excel.GetItem(relationID)
	if !ok {
		return
	}
	var prop *protoMsg.WearInGame
	switch itemData.Type {
	case 12: //头盔
		prop = mgr.user.GetHeadWearInGame()
	case 16: //背包
		prop = mgr.user.GetPackWearInGame()
	default:
		return
	}
	if prop == nil {
		prop = &protoMsg.WearInGame{}
	}
	if state == 1 {
		if itemData.Subtype == 1 && prop.First != 0 {
			return
		}
		if itemData.Subtype == 2 && prop.Second != 0 {
			return
		}
		if itemData.Subtype == 3 && prop.Third != 0 {
			return
		}
	}
	id := uint32(itemData.Id)
	if state == 0 { //卸下
		id = 0
	}
	if itemData.Subtype == 1 {
		prop.First = id
	}
	if itemData.Subtype == 2 {
		prop.Second = id
	}
	if itemData.Subtype == 3 {
		prop.Third = id
	}
	mgr.user.WearInGameFlow(uint32(itemData.Id), state)
	switch itemData.Type {
	case 12: //头盔
		mgr.user.SetHeadWearInGame(prop)
		mgr.user.SetHeadWearInGameDirty()
	case 16: //背包
		mgr.user.SetPackWearInGame(prop)
		mgr.user.SetPackWearInGameDirty()
	default:
		return
	}

}

// repeatBuyGoodsCheck 重复购买道具检测
func (mgr *StoreMgr) repeatBuyGoodsCheck(goodID uint32) bool {
	// 商品信息
	goodsConfig, ok := excel.GetStore(uint64(goodID))
	if !ok {
		mgr.user.Error("repeatBuyGoodsCheck failed, goods does't exist, goodID: ", goodID)
		return false
	}

	util := db.PlayerGoodsUtil(mgr.user.GetDBID())
	info, err := util.GetGoodsInfo(uint32(goodsConfig.RelationID))
	if err != nil || info == nil {
		return true
	}

	// 可重复购买判断
	if info.Time != 0 && info.EndTime == 0 {
		if goodsConfig.Piles == 0 {
			mgr.user.Error("repeatBuyGoodsCheck failed, repeat buy!")
			return false
		}
	}

	return true
}

package main

import (
	"common"
	"db"
	"excel"
	"protoMsg"
	"zeus/iserver"

	log "github.com/cihub/seelog"
)

/*------------------------------------------兑换活动-----------------------------------------------*/

// syncExchangeInfo 同步兑换活动信息
func (a *ActivityMgr) syncExchangeInfo() {
	a.user.Debug("syncExchangeInfo")

	msg := &protoMsg.ExchangeTotalList{}
	for k, v := range a.activeInfo {
		// 非兑换活动
		if v.activityType != 1 {
			continue
		}

		state, index := a.checkOpenDate(k)
		if state == ConfigErr {
			continue
		}

		info := &db.ExchangeActivityInfo{}
		info.ExchangeNum = make(map[uint32]uint32)
		info.CallState = make(map[uint32]uint32)

		if ok := a.activityUtil[k].IsActivity(); ok && state == 0 {
			if err := a.activityUtil[k].GetInfo(info); err != nil {
				a.user.Error("GetInfo fail:", err)
				return
			}
			if info.ActStartTm != a.activeInfo[k].tmStart[index] {
				a.activityUtil[k].ClearOwnActivity()

				info.ExchangeNum = make(map[uint32]uint32)
				for _, v1 := range excel.GetExchangestoreMap() {
					if v1.Activityid != k {
						continue
					}

					info.CallState[uint32(v1.ExchangeList)] = uint32(v1.ExchangeTips)
				}
			}
		} else {
			for _, v1 := range excel.GetExchangestoreMap() {
				if v1.Activityid != k {
					continue
				}

				info.CallState[uint32(v1.ExchangeList)] = uint32(v1.ExchangeTips)
			}
		}

		exchangeList := &protoMsg.ExchangeList{}
		exchangeList.Id = uint32(k)
		exchangeList.ActState = state

		for _, v1 := range excel.GetExchangestoreMap() {
			if v1.Activityid != k {
				continue
			}

			exchangeInfo := &protoMsg.ExchangeInfo{}

			exchangeInfo.Id = uint32(v1.ExchangeList)
			exchangeInfo.Num, _, _ = a.getExchangeGoods(v1.ExchangeA)
			exchangeInfo.State = info.CallState[exchangeInfo.Id]
			exchangeInfo.ExchangeNum = info.ExchangeNum[exchangeInfo.Id]

			exchangeList.Info = append(exchangeList.Info, exchangeInfo)
		}

		msg.List = append(msg.List, exchangeList)
	}

	log.Debug("syncExchangeInfo msg:", msg)
	if err := a.user.RPC(iserver.ServerTypeClient, "syncExchangeInfo", msg); err != nil {
		a.user.Error(err)
	}
}

// exchangeGoods 兑换物品
func (a *ActivityMgr) exchangeGoods(actId, exchangeId uint32) uint32 {

	state, index := a.checkOpenDate(uint64(actId))
	if state != 0 {
		return 0
	}

	info := &db.ExchangeActivityInfo{}
	info.ExchangeNum = make(map[uint32]uint32)
	info.CallState = make(map[uint32]uint32)

	if ok := a.activityUtil[uint64(actId)].IsActivity(); ok {
		if err := a.activityUtil[uint64(actId)].GetInfo(info); err != nil {
			a.user.Error("GetInfo fail:", err)
			return 0
		}
	} else {
		for _, v := range excel.GetExchangestoreMap() {
			if v.Activityid != uint64(actId) {
				continue
			}

			info.CallState[uint32(v.ExchangeList)] = uint32(v.ExchangeTips)
		}
	}
	info.Id = actId
	info.ActStartTm = a.activeInfo[uint64(actId)].tmStart[index]

	exchangestoreData := a.getExchangestoreData(actId, exchangeId)
	if uint64(info.ExchangeNum[exchangeId]) >= exchangestoreData.ExchangeMax {
		return 0
	}

	_, aOwnNumMap, aNeedNumMap := a.getExchangeGoods(exchangestoreData.ExchangeA)
	_, _, bNeedNumMap := a.getExchangeGoods(exchangestoreData.ExchangeB)

	for k, v := range aNeedNumMap {
		if aOwnNumMap[k] < v {
			return 0
		}
	}

	if ok1 := a.addExchangeRewardBag(aNeedNumMap, bNeedNumMap); ok1 {
		info.ExchangeNum[exchangeId]++
		if ok2 := a.activityUtil[uint64(actId)].SetInfo(info); !ok2 {
			a.user.Error("SetInfo fail!")
			return 0
		}
	}

	a.user.Debug("exchangeGoods info:", info)
	a.syncExchangeInfo()
	a.user.tlogExchangeFlow(actId, exchangeId, info.ExchangeNum[exchangeId], bNeedNumMap)

	return 1
}

// exchangeCallState 兑换活动提醒状态
func (a *ActivityMgr) exchangeCallState(actId, exchangeId, open uint32) (uint32, uint32) {

	state, index := a.checkOpenDate(uint64(actId))
	if state != 0 {
		return 0, 0
	}

	info := &db.ExchangeActivityInfo{}
	info.ExchangeNum = make(map[uint32]uint32)
	info.CallState = make(map[uint32]uint32)

	if ok := a.activityUtil[uint64(actId)].IsActivity(); ok {
		if err := a.activityUtil[uint64(actId)].GetInfo(info); err != nil {
			a.user.Error("GetInfo fail:", err)
			return 0, 0
		}
	} else {
		for _, v := range excel.GetExchangestoreMap() {
			if v.Activityid != uint64(actId) {
				continue
			}

			info.CallState[uint32(v.ExchangeList)] = uint32(v.ExchangeTips)
		}
	}

	info.Id = actId
	info.ActStartTm = a.activeInfo[uint64(actId)].tmStart[index]
	info.CallState[exchangeId] = open
	if ok2 := a.activityUtil[uint64(actId)].SetInfo(info); !ok2 {
		a.user.Error("SetInfo fail!")
		return 0, 0
	}

	a.user.Debug("exchangeGoods info:", info)
	return 1, info.CallState[exchangeId]
}

// getExchangeGoods 获得兑换物的拥有数量
func (a *ActivityMgr) getExchangeGoods(exchangeA string) ([]uint32, map[uint32]uint32, map[uint32]uint32) {
	var ownNum []uint32
	ownNumMap := make(map[uint32]uint32)
	needNumMap := make(map[uint32]uint32)

	awardsMap, err := common.SplitReward(exchangeA)
	if err != nil {
		a.user.Error("splitReward err:", err)
		return ownNum, ownNumMap, needNumMap
	}

	for k, v := range awardsMap {
		if k == common.Item_Coin { //金币
			num := uint32(a.user.GetCoin())
			ownNum = append(ownNum, num)
			ownNumMap[k] = num
		} else if k == common.Item_Diam { //钻石
			num := uint32(a.user.GetDiam())
			ownNum = append(ownNum, num)
			ownNumMap[k] = num
		} else if k == common.Item_Exp { //经验值
			num := a.user.GetExp()
			ownNum = append(ownNum, num)
			ownNumMap[k] = num
		} else if k == common.Item_ComradePoints { //战友积分
			num := uint32(a.user.GetComradePoints())
			ownNum = append(ownNum, num)
			ownNumMap[k] = num
		} else if k == common.Item_BraveCoin { //勇气值
			num := uint32(a.user.GetBraveCoin())
			ownNum = append(ownNum, num)
			ownNumMap[k] = num
		} else {
			goodsInfo, _ := db.PlayerGoodsUtil(a.user.GetDBID()).GetGoodsInfo(k)
			if goodsInfo == nil {
				ownNum = append(ownNum, 0)
				ownNumMap[k] = 0
			} else {
				ownNum = append(ownNum, goodsInfo.Sum)
				ownNumMap[k] = goodsInfo.Sum
			}
		}

		needNumMap[k] = v
	}

	return ownNum, ownNumMap, needNumMap
}

// getExchangestoreData 根据活动id 兑换id获得所需兑换数据的信息ExchangestoreData
func (a *ActivityMgr) getExchangestoreData(actId, exchangeId uint32) excel.ExchangestoreData {
	data := excel.ExchangestoreData{}
	for _, v := range excel.GetExchangestoreMap() {
		if v.Activityid == uint64(actId) && v.ExchangeList == uint64(exchangeId) {
			data = v
			break
		}
	}

	return data
}

// addExchangeRewardBag 添加领取的物品到包裹
func (a *ActivityMgr) addExchangeRewardBag(aNeedNumMap, bNeedNumMap map[uint32]uint32) bool {
	for k, v := range aNeedNumMap {
		if k == common.Item_Coin { //金币
			a.user.storeMgr.reduceMoney(common.MT_MONEY, common.RS_Exchange, uint64(v))
		} else if k == common.Item_Diam { //钻石
			a.user.storeMgr.reduceMoney(common.MT_DIAMOND, common.RS_Exchange, uint64(v))
		} else if k == common.Item_Exp { //经验值
			a.user.storeMgr.reduceMoney(common.MT_EXP, common.RS_Exchange, uint64(v))
		} else if k == common.Item_ComradePoints { //战友积分
			a.user.storeMgr.reduceMoney(common.MT_ComradePoints, common.RS_Exchange, uint64(v))
		} else if k == common.Item_BraveCoin { //勇气值
			a.user.storeMgr.reduceMoney(common.MT_BraveCoin, common.RS_Exchange, uint64(v))
		} else {
			if ok := a.user.storeMgr.ReduceGoods(k, v, common.RS_Exchange); !ok {
				return false
			}
		}
	}

	for k, v := range bNeedNumMap {
		a.user.storeMgr.GetGoods(k, v, common.RS_Exchange, common.MT_NO, 0)
	}

	return true
}

// updateExchangeOwnNum 更新拥有商品的数量
func (a *ActivityMgr) updateExchangeOwnNum(id uint32) {
	var openActId, closeActId uint64
	for _, v := range excel.GetExchangestoreMap() {
		if v.Activityid == closeActId {
			continue
		}

		if v.Activityid != openActId {
			state, _ := a.checkOpenDate(v.Activityid)
			if state == ConfigErr {
				closeActId = v.Activityid
				continue
			} else {
				openActId = v.Activityid
			}
		}

		_, _, aNeedNumMap := a.getExchangeGoods(v.ExchangeA)
		for i, _ := range aNeedNumMap {
			if i == id {
				a.syncExchangeInfo()
				return
			}
		}
	}
}

package main

import (
	"common"
	"db"
	"excel"
	"protoMsg"
	"strconv"
	"strings"
	"time"
	"zeus/iserver"
)

/*--------------------------------一球成名活动------------------------------------*/

// syncBallStarInfo 同步一球成名活动信息
func (a *ActivityMgr) syncBallStarInfo() {
	a.user.Debug("syncBallStarInfo")

	state, index := a.checkOpenDate(ballStarActivityID)
	if state == ConfigErr {
		return
	}

	info := &db.BallStarInfo{}
	info.PickState = make(map[uint32]uint32)

	info.Id = ballStarActivityID
	info.ActStartTm = a.activeInfo[ballStarActivityID].tmStart[index]
	info.Position = 1
	if ok := a.activityUtil[ballStarActivityID].IsActivity(); ok && state == 0 {
		if err := a.activityUtil[ballStarActivityID].GetInfo(info); err != nil {
			a.user.Error("GetInfo fail:", err)
			return
		}

		changeBool := false
		if info.ActStartTm != a.activeInfo[ballStarActivityID].tmStart[index] {
			a.activityUtil[ballStarActivityID].ClearOwnActivity()
		} else if info.Time != 0 && time.Unix(info.Time, 0).Format("2006-01-02") != time.Now().Format("2006-01-02") {
			info.DayNum = 0
			changeBool = true
		}

		weekTime := int64(common.GetTBSystemValue(common.System_BallStarFreshWeek))
		hourTime := int64(common.GetTBSystemValue(common.System_BallStarFreshHour))
		freshTime := common.GetThisWeekBeginStamp() + (weekTime-1)*86400 + hourTime*3600
		if info.FreshSum != 0 && time.Now().Unix() >= freshTime && info.Time < freshTime {
			for k, _ := range info.PickState {
				info.PickState[k] = 0
			}
			info.FreshSum = 0
			changeBool = true
		}

		if changeBool {
			if ok := a.activityUtil[ballStarActivityID].SetInfo(info); !ok {
				a.user.Error("SetInfo fail!")
				return
			}
		}
	}

	msg := &protoMsg.BallStarInfo{}
	msg.Id = ballStarActivityID
	msg.ActState = state
	msg.Position = info.Position
	msg.Sum = info.FreshSum
	for k, _ := range excel.GetBallStarListMap() {
		ballStarReward := &protoMsg.BallStarReward{}
		ballStarReward.Id = uint32(k)
		ballStarReward.State = info.PickState[uint32(k)]

		msg.RewardState = append(msg.RewardState, ballStarReward)
	}

	a.user.Debug("syncBallStarInfo msg:", msg)
	if err := a.user.RPC(iserver.ServerTypeClient, "syncBallStarInfo", msg); err != nil {
		a.user.Error(err)
	}
}

// clickBallStarReward 抽一球成名活动奖励
func (a *ActivityMgr) clickBallStarReward() (bool, uint32, uint32, uint32) {
	state, index := a.checkOpenDate(ballStarActivityID)
	if state != 0 {
		return false, 0, 0, 0
	}

	info := &db.BallStarInfo{}
	info.PickState = make(map[uint32]uint32)

	info.Id = ballStarActivityID
	info.ActStartTm = a.activeInfo[ballStarActivityID].tmStart[index]
	info.Position = 1

	if ok := a.activityUtil[ballStarActivityID].IsActivity(); ok {
		if err := a.activityUtil[ballStarActivityID].GetInfo(info); err != nil {
			a.user.Error("GetInfo fail:", err)
			return false, 0, 0, 0
		}
	}
	info.Time = time.Now().Unix()

	if info.DayNum >= uint32(common.GetTBSystemValue(common.System_BallStarDayNum)) {
		a.user.AdviceNotify(common.NotifyCommon, 82)
		a.user.Warn("click num upper!")
		return false, 0, 0, 0
	}

	// 判断抽奖消耗的道具id存量是否足够
	itemId := uint32(common.GetTBSystemValue(common.System_BallStarUseItem))
	goodsInfo, _ := db.PlayerGoodsUtil(a.user.GetDBID()).GetGoodsInfo(itemId)
	if goodsInfo == nil {
		a.user.Warn("no own item:", itemId)
		return false, 0, 0, 0
	}
	if goodsInfo.Sum < 1 {
		a.user.Warn("ownnum < 1 goodsInfo.Sum:", goodsInfo.Sum)
		return false, 0, 0, 0
	}

	// 开始抽奖
	randomValue := a.sweepstake(info.Position, info.Sum)
	if randomValue == 0 {
		a.user.Error("BallStarRound Excel Err!")
		return false, 0, 0, 0
	}
	grid := uint32(common.GetTBSystemValue(common.System_BallStarGridNum))
	if info.Position+randomValue > grid {
		info.Position += randomValue - grid
	} else {
		info.Position += randomValue
	}

	// 获取抽到的奖励配置
	ballStarRoundData, ok := excel.GetBallStarRound(uint64(info.Position))
	if !ok {
		a.user.Warn("GetBallStarRound err! position:", info.Position)
		return false, 0, 0, 0
	}

	// 获取抽到的奖励信息string
	var rewards string
	if ballStarRoundData.RewardType == 2 { //代表的是奖品
		rewards = ballStarRoundData.Rewards
	} else if ballStarRoundData.RewardType == 1 { //代表的是宝箱
		rewards = a.randomBoxReward(ballStarRoundData.Rewards)
	}

	var err error
	var rewardType, rewardNum int
	if rewards != "" {
		rewardMsg := strings.Split(rewards, "|")
		if len(rewardMsg) != 2 {
			a.user.Warn("len(rewardMsg) != 2  len(rewardMsg):", len(rewardMsg))
			return false, 0, 0, 0
		}
		rewardType, err = strconv.Atoi(rewardMsg[0])
		if err != nil {
			a.user.Warn("strconv.Atoi(rewardMsg[0]) err:", err, " rewardMsg[0]:", rewardMsg[0])
			return false, 0, 0, 0
		}
		rewardNum, err = strconv.Atoi(rewardMsg[1])
		if err != nil {
			a.user.Warn("strconv.Atoi(rewardMsg[1]) err:", err, " rewardMsg[1]:", rewardMsg[1])
			return false, 0, 0, 0
		}
	}

	// 添加奖励信息到背包
	if ok1 := a.addRewardToBag(rewards, common.RS_BallStar); ok1 {
		info.Sum++
		info.DayNum++
		info.FreshSum++
		if ok2 := a.activityUtil[ballStarActivityID].SetInfo(info); !ok2 {
			a.user.Error("SetInfo fail!")
			return false, 0, 0, 0
		}
	}
	if !a.user.storeMgr.ReduceGoods(itemId, 1, common.RS_BallStar) {
		a.user.Warn("ReduceGoods fail!")
		return false, 0, 0, 0
	}

	a.user.tlogBallStarFlow(info.Position, info.Sum, uint32(rewardType), uint32(rewardNum))

	a.user.Debug("clickBallStarReward info:", info)
	return true, randomValue, uint32(rewardType), uint32(rewardNum)
}

// pickBallStarReward 领取一球成名活动奖励
func (a *ActivityMgr) pickBallStarReward(id uint32) bool {

	state, index := a.checkOpenDate(ballStarActivityID)
	if state != 0 {
		return false
	}

	info := &db.BallStarInfo{}
	info.PickState = make(map[uint32]uint32)

	info.Id = ballStarActivityID
	info.ActStartTm = a.activeInfo[ballStarActivityID].tmStart[index]
	info.Position = 1

	if ok := a.activityUtil[ballStarActivityID].IsActivity(); !ok {
		return false
	}

	if err := a.activityUtil[ballStarActivityID].GetInfo(info); err != nil {
		a.user.Error("GetInfo fail:", err)
		return false
	}

	if info.PickState[id] == 1 {
		return false
	}

	ballStarListData, ok := excel.GetBallStarList(uint64(id))
	if !ok {
		return false
	}
	if uint64(info.FreshSum) < ballStarListData.Num {
		return false
	}

	if ok1 := a.addRewardToBag(ballStarListData.Rewards, common.RS_BallStar); ok1 {
		info.PickState[id] = 1
		if ok2 := a.activityUtil[ballStarActivityID].SetInfo(info); !ok2 {
			a.user.Error("SetInfo fail!")
			return false
		}
	}

	a.user.Debug("pickBallStarReward info:", info)
	return true
}

// sweepstake 转盘抽奖
func (a *ActivityMgr) sweepstake(position, sum uint32) uint32 {

	ballStarMap := excel.GetBallStarRoundMap()

	wheel := uint32(common.GetTBSystemValue(common.System_BallStarWheelNum))
	grid := uint32(common.GetTBSystemValue(common.System_BallStarGridNum))

	var weight uint32
	var weightList []uint32
	for i := position + 1; i <= position+wheel; i++ {
		var index uint32 = i
		if i > grid {
			index = i - grid
		}

		if uint64(sum) < ballStarMap[uint64(index)].MiniNum {
			weightList = append(weightList, 0)
			continue
		}
		weight += uint32(ballStarMap[uint64(index)].Weight)
		weightList = append(weightList, uint32(ballStarMap[uint64(index)].Weight))
	}

	return common.WeightRandom(weight, weightList)
}

// randomBoxReward 随机到宝箱奖励
func (a *ActivityMgr) randomBoxReward(rewards string) string {
	var weight uint32
	var weightList []uint32
	rewardList := make(map[uint32]string)

	reward := strings.Split(rewards, ";")
	for k, v := range reward {
		itemList := strings.Split(v, "-")
		if len(itemList) != 2 {
			return ""
		}

		rewardNum, err := strconv.Atoi(itemList[1])
		if err != nil {
			return ""
		}

		weight += uint32(rewardNum)
		weightList = append(weightList, uint32(rewardNum))
		rewardList[uint32(k)+1] = itemList[0]
	}

	index := common.WeightRandom(weight, weightList)
	return rewardList[index]
}

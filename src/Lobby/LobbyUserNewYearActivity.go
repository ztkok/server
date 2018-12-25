package main

import (
	"common"
	"db"
	"excel"
	"math/rand"
	"protoMsg"
	"strconv"
	"strings"
	"time"
	"zeus/iserver"

	log "github.com/cihub/seelog"
)

/*--------------------------------新年活动------------------------------------*/

// syncNewYearActivity 同步新年活动信息
func (a *ActivityMgr) syncNewYearActivity() {
	a.user.Debug("syncNewYearActivity")
	if a.activeInfo[newYearActivityID] == nil || a.activityUtil[newYearActivityID] == nil {
		return
	}

	if time.Now().Unix() < a.activeInfo[newYearActivityID].tmSartShow || time.Now().Unix() > a.activeInfo[newYearActivityID].tmEndShow {
		a.user.Info("syncNewYearActivity out of date!")
		return
	}

	index := a.getCurActivityTime(a.activeInfo[newYearActivityID])
	if index >= len(a.activeInfo[newYearActivityID].tmStart) || index >= len(a.activeInfo[newYearActivityID].tmEnd) || index >= len(a.activeInfo[newYearActivityID].pickNum) {
		a.user.Warn("tmStart:", len(a.activeInfo[newYearActivityID].tmStart), " tmEnd:", len(a.activeInfo[newYearActivityID].tmEnd), " pickNum:", len(a.activeInfo[newYearActivityID].pickNum), " index:", index+1)
		return
	}

	var recState uint32 = 0 //活动开始状态
	if time.Now().Unix() < a.activeInfo[newYearActivityID].tmStart[index] {
		recState = 1 //未开始
	}
	if time.Now().Unix() > a.activeInfo[newYearActivityID].tmEnd[index] {
		recState = 2 //已过期
	}

	var isClear bool = false
	info := &db.NewYearActivityInfo{}
	if ok := a.activityUtil[newYearActivityID].IsActivity(); ok {
		err := a.activityUtil[newYearActivityID].GetNewYearActivityInfo(info)
		if err != nil {
			a.user.Error("GetNewYearActivityInfo fail:", err)
			return
		}

		if info.ActStartTm != a.activeInfo[newYearActivityID].tmStart[index] {
			a.activityUtil[newYearActivityID].ClearOwnActivity()
			isClear = true
		}
	}
	randomInfo := a.SetRandomInfo(index)
	if randomInfo == nil {
		a.user.Error("SetRandomInfo fail:")
		return
	}

	var i uint32 = 1
	newYearInfo := &protoMsg.NewYearInfo{}
	for ; i <= a.activeInfo[newYearActivityID].pickNum[index]; i++ {
		newYearGood := &protoMsg.NewYearGood{}
		newYearGood.Key = i
		if !isClear {
			newYearGood.Num = info.GoodId[i]
		} else {
			newYearGood.Num = 0
		}
		newYearGood.TargetType = randomInfo.TargetType[i]
		newYearGood.TargetNum = randomInfo.TargetNum[i]

		newYearInfo.GoodID = append(newYearInfo.GoodID, newYearGood)
	}
	newYearInfo.RecState = recState
	newYearInfo.Index = uint32(index + 1)

	log.Info("SyncNewYearInfo-newYearInfo:", newYearInfo)
	// 返回当前活动可领取id
	if err := a.user.RPC(iserver.ServerTypeClient, "SyncNewYearInfo", newYearInfo); err != nil {
		a.user.Error(err)
	}
}

// PickNewYearActivity 领取新年活动物品
func (a *ActivityMgr) PickNewYearActivity(id uint32) {
	a.user.Debug("PickNewYearActivity")
	if a.activeInfo[newYearActivityID] == nil || a.activityUtil[newYearActivityID] == nil {
		return
	}

	if time.Now().Unix() < a.activeInfo[newYearActivityID].tmSartShow || time.Now().Unix() > a.activeInfo[newYearActivityID].tmEndShow {
		a.user.Info("PickNewYearActivity out of date!")
		return
	}

	index := a.getCurActivityTime(a.activeInfo[newYearActivityID])
	if index >= len(a.activeInfo[newYearActivityID].tmStart) || index >= len(a.activeInfo[newYearActivityID].tmEnd) || index >= len(a.activeInfo[newYearActivityID].pickNum) {
		a.user.Warn("tmStart:", len(a.activeInfo[newYearActivityID].tmStart), " tmEnd:", len(a.activeInfo[newYearActivityID].tmEnd), " pickNum:", len(a.activeInfo[newYearActivityID].pickNum), " index:", index+1)
		return
	}

	if time.Now().Unix() < a.activeInfo[newYearActivityID].tmStart[index] {
		a.user.Info("activity no start!")
		return
	}

	randomInfo := &db.NewYearRandomInfo{}
	if is := a.activityUtil[newYearActivityID].IsNewYearRandomActivity(); !is {
		a.user.Error("IsNewYearRandomActivity fail!")
		return
	}
	if err := a.activityUtil[newYearActivityID].GetNewYearRandomInfo(randomInfo); err != nil {
		a.user.Error("GetNewYearRandomInfo fail:", err)
		return
	}

	info := &db.NewYearActivityInfo{}
	info.Id = newYearActivityID
	info.Time = time.Now().Unix()
	info.ActStartTm = a.activeInfo[newYearActivityID].tmStart[index]
	info.TargetStage = 1

	var recState uint32 = 0 //有资格领取（0：领取失败）

	if ok := a.activityUtil[newYearActivityID].IsActivity(); ok {
		if err := a.activityUtil[newYearActivityID].GetNewYearActivityInfo(info); err != nil {
			a.user.Error("GetNewYearActivityInfo fail:", err)
			return
		}

		totalNum := a.GetIndex(index)
		if info.GoodId[id] != -1 && info.GoodId[id] >= int32(randomInfo.TargetNum[id]) {
			if sucss := a.AddNewYearRewardToBag(uint64(totalNum+id), index); sucss {
				info.GoodId[id] = -1
				ok := a.activityUtil[newYearActivityID].SetNewYearActivityInfo(info)
				if ok {
					recState = 1 //领取成功
				}
			}
		}
	}

	log.Info("PickNewYearActRet-GoodId:", info.GoodId, " id:", id, " recState:", recState, " index:", index)
	if err := a.user.RPC(iserver.ServerTypeClient, "PickNewYearActRet", id, recState); err != nil {
		a.user.Error(err)
	}
}

// SetNewYearInfo 设置新年活动领取状态信息
func (a *ActivityMgr) SetNewYearInfo(goodID map[uint64]int32) {
	if a.activeInfo[newYearActivityID] == nil || a.activityUtil[newYearActivityID] == nil {
		return
	}

	index := a.getCurActivityTime(a.activeInfo[newYearActivityID])
	if index >= len(a.activeInfo[newYearActivityID].tmStart) || index >= len(a.activeInfo[newYearActivityID].tmEnd) || index >= len(a.activeInfo[newYearActivityID].pickNum) {
		a.user.Warn("tmStart:", len(a.activeInfo[newYearActivityID].tmStart), " tmEnd:", len(a.activeInfo[newYearActivityID].tmEnd), " pickNum:", len(a.activeInfo[newYearActivityID].pickNum), " index:", index+1)
		return
	}

	if time.Now().Unix() < a.activeInfo[newYearActivityID].tmStart[index] || time.Now().Unix() > a.activeInfo[newYearActivityID].tmEnd[index] {
		a.user.Info("activity no start or expire")
		return
	}

	randomInfo := &db.NewYearRandomInfo{}
	if is := a.activityUtil[newYearActivityID].IsNewYearRandomActivity(); !is {
		a.user.Error("IsNewYearRandomActivity fail!")
		return
	}
	if err := a.activityUtil[newYearActivityID].GetNewYearRandomInfo(randomInfo); err != nil {
		a.user.Error("GetNewYearRandomInfo fail:", err)
		return
	}

	var isStage bool = true // 活动阶数是否改变
	info := &db.NewYearActivityInfo{}
	info.GoodId = make(map[uint32]int32, 0)
	info.TargetStage = 1
	if ok := a.activityUtil[newYearActivityID].IsActivity(); ok {
		if err := a.activityUtil[newYearActivityID].GetNewYearActivityInfo(info); err != nil {
			a.user.Error("GetNewYearActivityInfo fail:", err)
			return
		}
	}
	info.Id = newYearActivityID
	info.Time = time.Now().Unix()
	info.ActStartTm = a.activeInfo[newYearActivityID].tmStart[index]

	var i, reachNum uint32 = 1, 0
	totalNum := a.GetIndex(index)
	targetMap := excel.GetNewyearCheckinMap()
	for ; i <= a.activeInfo[newYearActivityID].pickNum[index]; i++ {
		if targetMap[uint64(totalNum+i)].ActivityNum != uint64(index+1) {
			continue
		}

		if uint64(info.TargetStage) == targetMap[uint64(totalNum+i)].TargetStage && randomInfo.TargetType[i] != 7 && info.GoodId[i] == -1 {
			reachNum++
		}

		if info.GoodId[i] == -1 {
			continue
		}

		if uint64(info.TargetStage) == targetMap[uint64(totalNum+i)].TargetStage {
			info.GoodId[i] += goodID[uint64(randomInfo.TargetType[i])]

			if randomInfo.TargetType[i] != 7 && info.GoodId[i] >= int32(randomInfo.TargetNum[i]) {
				info.GoodId[i] = int32(randomInfo.TargetNum[i])
				reachNum++
			}

			if randomInfo.TargetType[i] == 7 {
				info.GoodId[i] = int32(reachNum) //7表示已达成目标(前几个目标都达成)
			}

			if info.GoodId[i] < int32(randomInfo.TargetNum[i]) {
				isStage = false
			}
		}
	}

	if isStage && info.TargetStage < a.activeInfo[newYearActivityID].pickNum[index] {
		info.TargetStage++
	}

	a.user.Info("SetNewYearActivityInfo:", info)
	if ok := a.activityUtil[newYearActivityID].SetNewYearActivityInfo(info); !ok {
		a.user.Error("SetNewYearActivityInfo fail")
		return
	}

	a.syncNewYearActivity() //同步新年活动物品
}

// SetRandomInfo 设置该轮活动目标id
func (a *ActivityMgr) SetRandomInfo(index int) *db.NewYearRandomInfo {
	info := &db.NewYearRandomInfo{}
	if ok := a.activityUtil[newYearActivityID].IsNewYearRandomActivity(); ok {
		if err := a.activityUtil[newYearActivityID].GetNewYearRandomInfo(info); err != nil {
			a.user.Error("GetNewYearRandomInfo fail:", err)
			return nil
		}

		if info.ActStartTm != a.activeInfo[newYearActivityID].tmStart[index] {
			a.activityUtil[newYearActivityID].ClearNewYearRandomActivity()
		} else {
			return info
		}
	}

	info.Id = newYearActivityID
	info.TargetType = make(map[uint32]uint32)
	info.TargetNum = make(map[uint32]uint32)
	info.ActStartTm = a.activeInfo[newYearActivityID].tmStart[index]

	var i uint32 = 1
	totalNum := a.GetIndex(index)
	targetMap := excel.GetNewyearCheckinMap()
	for ; i <= a.activeInfo[newYearActivityID].pickNum[index]; i++ {
		if targetMap[uint64(totalNum+i)].ActivityNum != uint64(index+1) {
			continue
		}

		if targetMap[uint64(totalNum+i)].TargetType != 0 {
			info.TargetType[i] = uint32(targetMap[uint64(totalNum+i)].TargetType)
			info.TargetNum[i] = uint32(targetMap[uint64(totalNum+i)].TargetNum)
			continue
		}

		if targetMap[uint64(totalNum+i)].RandomType != "" {
			randomType := strings.Split(targetMap[uint64(totalNum+i)].RandomType, ";")
			randomNum := strings.Split(targetMap[uint64(totalNum+i)].RandomNum, ";")
			num := rand.Intn(len(randomType))

			if len(randomType) != len(randomNum) || num >= len(randomType) {
				a.user.Warn("NewyearCheckin table error!")
				continue
			}

			targetType, err := strconv.ParseUint(randomType[num], 10, 32)
			if err != nil {
				a.user.Error("ParseUint err:", err)
				return nil
			}
			info.TargetType[i] = uint32(targetType)

			targetNum, err := strconv.ParseUint(randomNum[num], 10, 32)
			if err != nil {
				a.user.Error("ParseUint err:", err)
				return nil
			}
			info.TargetNum[i] = uint32(targetNum)
		}
	}

	if ok := a.activityUtil[newYearActivityID].SetNewYearRandomInfo(info); !ok {
		a.user.Error("SetNewYearRandomInfo fail!")
	}

	return info
}

// AddFirstWinRewardToBag 添加领取的物品到包裹
func (a *ActivityMgr) AddNewYearRewardToBag(id uint64, index int) bool {
	getTarget, ok := excel.GetNewyearCheckin(id)
	if ok {
		if getTarget.ActivityNum == uint64(index+1) {
			a.user.storeMgr.GetGoods(uint32(getTarget.BonusID), uint32(getTarget.BonusNum), common.RS_NewYearAct, common.MT_NO, 0) //common.RS_NewYearAct
			return true
		}
	}

	return false
}

// GetIndex 获取第几次活动的开始id数
func (a *ActivityMgr) GetIndex(index int) uint32 {
	var totalNum uint32 = 0
	for i := 0; i < index; i++ {
		totalNum += a.activeInfo[newYearActivityID].pickNum[i]
	}

	return totalNum
}

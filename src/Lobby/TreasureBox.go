package main

import (
	"common"
	"db"
	"excel"
	"fmt"
	"math/rand"
	"protoMsg"
	"strconv"
	"strings"
	"time"
	"zeus/iserver"

	//log "github.com/cihub/seelog"
)

const (
	NoFlush   = 0 //不刷新
	DayFlush  = 1 //每天刷新
	WeekFlush = 2 //每周刷新
)

// OpenBoxMoney 开宝箱价格(根据开箱次数变化)
type OpenBoxMoney struct {
	num       uint32 //周期内开箱次数
	moneyType uint32 //货币类型
	money     uint32 //货币数量
}

// MegaPoolInfo 奖池信息
type MegaPoolInfo struct {
	poolId uint32 //奖池id
	weight uint32 //权重
	minNum uint32 //保底次数(总领取次数)
}

// ItemInfo 道具奖品信息
type ItemInfo struct {
	poolId uint32 //奖池id
	itemId uint32 //道具id
	weight uint32 //权重
	grade  uint32 //品质
}

// TreasureBoxInfo 宝箱信息
type TreasureBoxInfo struct {
	id        uint64 //宝箱id
	tmStart   int64
	tmEnd     int64
	flushType uint32 //刷新方式

	openBoxMoney map[uint32]OpenBoxMoney //开箱价格信息
	megaPoolInfo []MegaPoolInfo          //奖池信息
	itemInfo     map[uint32][]ItemInfo   //道具奖品信息
}

func (treasureBoxInfo *TreasureBoxInfo) String() string {
	return fmt.Sprintf("%+v\n", *treasureBoxInfo)
}

// TreasureBoxMgr 宝箱管理器
type TreasureBoxMgr struct {
	user *LobbyUser

	treasureBoxInfo map[uint64]*TreasureBoxInfo    //宝箱信息
	treasureBoxUtil map[uint64]*db.TreasureBoxUtil //宝箱工具
}

// NewTreasureBoxMgr 获取宝箱管理器
func NewTreasureBoxMgr(user *LobbyUser) *TreasureBoxMgr {
	treasureBox := &TreasureBoxMgr{
		user:            user,
		treasureBoxInfo: make(map[uint64]*TreasureBoxInfo),
		treasureBoxUtil: make(map[uint64]*db.TreasureBoxUtil),
	}

	treasureBox.initTreasureBoxInfo()
	return treasureBox
}

// initTreasureBoxInfo 初始化宝箱信息
func (t *TreasureBoxMgr) initTreasureBoxInfo() {
	boxMap := excel.GetBoxMap()
	for _, v := range boxMap {
		treasureBoxInfo := &TreasureBoxInfo{}
		treasureBoxInfo.openBoxMoney = make(map[uint32]OpenBoxMoney)
		treasureBoxInfo.itemInfo = make(map[uint32][]ItemInfo)

		treasureBoxInfo.id = v.Id
		startDate, err := time.ParseInLocation("2006|01|02|15|04|05", v.StartDate, time.Local)
		if err != nil {
			t.user.Error("initTreasureBoxInfo StartDate:", err)
			return
		}
		treasureBoxInfo.tmStart = startDate.Unix()

		endData, err := time.ParseInLocation("2006|01|02|15|04|05", v.EndDate, time.Local)
		if err != nil {
			t.user.Error("initTreasureBoxInfo endData:", err)
			return
		}
		treasureBoxInfo.tmEnd = endData.Unix()
		treasureBoxInfo.flushType = uint32(v.Refreshrate)

		money := strings.Split(v.Money, ";")
		for _, v1 := range money {
			moneyInfo := strings.Split(v1, "-")
			if len(moneyInfo) != 3 {
				t.user.Warn("len(moneyInfo) != 3")
				return
			}
			openBoxMoney := OpenBoxMoney{}
			num, err := strconv.ParseUint(moneyInfo[0], 10, 32)
			if err != nil {
				t.user.Error("ParseUint:", err, " ", num)
				return
			}
			openBoxMoney.num = uint32(num)
			moneyType, err := strconv.ParseUint(moneyInfo[1], 10, 32)
			if err != nil {
				t.user.Error("ParseUint:", err, " ", moneyType)
				return
			}
			openBoxMoney.moneyType = uint32(moneyType)
			money, err := strconv.ParseUint(moneyInfo[2], 10, 32)
			if err != nil {
				t.user.Error("ParseUint:", err, " ", money)
				return
			}
			openBoxMoney.money = uint32(money)

			treasureBoxInfo.openBoxMoney[openBoxMoney.num] = openBoxMoney
		}

		pond := strings.Split(v.Pond, ";")
		for _, v1 := range pond {
			pondInfo := strings.Split(v1, "-")
			if len(pondInfo) != 3 {
				t.user.Error("len(pondInfo) != 3")
				return
			}
			megaPoolInfo := MegaPoolInfo{}
			poolId, err := strconv.ParseUint(pondInfo[0], 10, 32)
			if err != nil {
				t.user.Error("ParseUint:", err, " ", poolId)
				return
			}
			megaPoolInfo.poolId = uint32(poolId)
			weight, err := strconv.ParseUint(pondInfo[1], 10, 32)
			if err != nil {
				t.user.Error("ParseUint:", err, " ", weight)
				return
			}
			megaPoolInfo.weight = uint32(weight)
			minNum, err := strconv.ParseUint(pondInfo[2], 10, 32)
			if err != nil {
				t.user.Error("ParseUint:", err, " ", minNum)
				return
			}
			megaPoolInfo.minNum = uint32(minNum)

			treasureBoxInfo.megaPoolInfo = append(treasureBoxInfo.megaPoolInfo, megaPoolInfo)
		}

		item := strings.Split(v.Item, ";")
		for _, v1 := range item {
			itemInfo := strings.Split(v1, "-")
			if len(itemInfo) != 4 {
				t.user.Error("len(itemInfo) != 4")
				return
			}
			itemInfos := ItemInfo{}
			poolId, err := strconv.ParseUint(itemInfo[0], 10, 32)
			if err != nil {
				t.user.Error("ParseUint:", err, " ", poolId)
				return
			}
			itemInfos.poolId = uint32(poolId)
			itemId, err := strconv.ParseUint(itemInfo[1], 10, 32)
			if err != nil {
				t.user.Error("ParseUint:", err, " ", itemId)
				return
			}
			itemInfos.itemId = uint32(itemId)
			weight, err := strconv.ParseUint(itemInfo[2], 10, 32)
			if err != nil {
				t.user.Error("ParseUint:", err, " ", weight)
				return
			}
			itemInfos.weight = uint32(weight)
			grade, err := strconv.ParseUint(itemInfo[3], 10, 32)
			if err != nil {
				t.user.Error("ParseUint:", err, " ", grade)
				return
			}
			itemInfos.grade = uint32(grade)

			treasureBoxInfo.itemInfo[itemInfos.poolId] = append(treasureBoxInfo.itemInfo[itemInfos.poolId], itemInfos)
		}

		t.treasureBoxInfo[v.Id] = treasureBoxInfo
		t.treasureBoxUtil[v.Id] = db.PlayerTreasureBoxUtil(t.user.GetDBID(), uint32(v.Id))
	}

	//log.Debug("treasureBoxInfo:", t.treasureBoxInfo, " treasureBoxUtil:", t.treasureBoxUtil)
}

// syncTreasureBoxInfo 同步宝箱信息
func (t *TreasureBoxMgr) syncTreasureBoxInfo() {
	treasureBoxList := &protoMsg.TreasureBoxList{}
	for _, v := range t.treasureBoxInfo {
		if time.Now().Unix() < v.tmStart || time.Now().Unix() > v.tmEnd {
			continue
		}

		treasureBoxInfo := &protoMsg.TreasureBoxInfo{}
		treasureBoxInfo.Id = v.id

		info := &db.TreasureBoxInfo{}
		if t.treasureBoxUtil[v.id].IsGet() {
			if err := t.treasureBoxUtil[v.id].GetTreasureBoxInfo(info); err != nil {
				t.user.Warn("GetTreasureBoxInfo err:", err)
				return
			}

			// 本轮宝箱时间是否到期，到期进行清空
			if info.ActStartTm != v.tmStart {
				t.treasureBoxUtil[v.id].Clear()
				info.CycNum = 0
			}

			// 是否进行天或周刷新
			if v.flushType != NoFlush && info.CycNum != 0 {
				timeStr := time.Unix(info.Time, 0).Format("2006-01-02")
				tDay, _ := time.ParseInLocation("2006-01-02", timeStr, time.Local)
				weekDay := time.Unix(info.Time, 0).Weekday()
				if weekDay == 0 {
					weekDay = 7
				}
				tWeek := tDay.Unix() - int64(weekDay-1)*86400 //获取info.Time时间对应的周一0点时间戳

				if v.flushType == DayFlush && time.Now().Unix()-tDay.Unix() >= 86400 {
					info.CycNum = 0
				} else if v.flushType == WeekFlush && time.Now().Unix()-tWeek >= 604800 {
					info.CycNum = 0
				}

				if info.CycNum == 0 {
					t.user.Debug("tDay:", tDay.Unix(), " tWeek:", tWeek, " Now:", time.Now().Unix())

					info.Id = uint32(v.id)
					info.ActStartTm = v.tmStart
					if ok := t.treasureBoxUtil[v.id].SetTreasureBoxInfo(info); !ok {
						t.user.Warn("SetTreasureBoxInfo fail!")
						return
					}
				}
			}
		}

		if info.CycNum == uint32(len(v.openBoxMoney)) {
			treasureBoxInfo.CoinType = v.openBoxMoney[info.CycNum].moneyType
			treasureBoxInfo.CoinNum = v.openBoxMoney[info.CycNum].money
		} else {
			treasureBoxInfo.CoinType = v.openBoxMoney[info.CycNum+1].moneyType
			treasureBoxInfo.CoinNum = v.openBoxMoney[info.CycNum+1].money
		}

		treasureBoxList.Info = append(treasureBoxList.Info, treasureBoxInfo)
	}

	t.user.Debug("syncTreasureBoxInfo-treasureBoxList:", treasureBoxList)
	if err := t.user.RPC(iserver.ServerTypeClient, "RspTreasureBoxInfo", treasureBoxList); err != nil {
		t.user.Error(err)
	}
}

// syncOpenTreasureBox 同步开宝箱结果 (返回值中第一个变量 0:失败 1:领取成功 2:货币不够)
func (t *TreasureBoxMgr) syncOpenTreasureBox(id uint32) (uint32, uint32, uint32, uint32) {
	if t.treasureBoxInfo[uint64(id)] == nil || t.treasureBoxUtil[uint64(id)] == nil {
		t.user.Warn("TreasureBoxInfo nil or TreasureBoxUtil nil! id:", id)
		return 0, 0, 0, 0
	}

	if time.Now().Unix() < t.treasureBoxInfo[uint64(id)].tmStart || time.Now().Unix() > t.treasureBoxInfo[uint64(id)].tmEnd {
		t.user.Warn("time expire!")
		return 0, 0, 0, 0
	}

	var itemId, moneyType, money uint32
	info := &db.TreasureBoxInfo{}
	if t.treasureBoxUtil[uint64(id)].IsGet() {
		if err := t.treasureBoxUtil[uint64(id)].GetTreasureBoxInfo(info); err != nil {
			t.user.Warn("GetTreasureBoxInfo err:", err)
			return 0, 0, 0, 0
		}
	}

	info.Id = id
	info.ActStartTm = t.treasureBoxInfo[uint64(id)].tmStart
	info.Time = time.Now().Unix()
	info.TotalNum++
	info.CycNum++
	if info.CycNum >= uint32(len(t.treasureBoxInfo[uint64(id)].openBoxMoney)) {
		info.CycNum = uint32(len(t.treasureBoxInfo[uint64(id)].openBoxMoney))
		moneyType = t.treasureBoxInfo[uint64(id)].openBoxMoney[info.CycNum].moneyType
		money = t.treasureBoxInfo[uint64(id)].openBoxMoney[info.CycNum].money
	} else {
		moneyType = t.treasureBoxInfo[uint64(id)].openBoxMoney[info.CycNum+1].moneyType
		money = t.treasureBoxInfo[uint64(id)].openBoxMoney[info.CycNum+1].money
	}

	// 抽取宝箱道具
	itemId, poolId := t.randomItem(uint64(id), info.TotalNum)
	if itemId == 0 {
		t.user.Warn("randomItem == 0! poolId:", poolId)
		return 0, 0, 0, 0
	}

	iMoneyType := t.treasureBoxInfo[uint64(id)].openBoxMoney[info.CycNum].moneyType
	iMoney := t.treasureBoxInfo[uint64(id)].openBoxMoney[info.CycNum].money
	// 判断货币是否足够
	if ok := t.judgeReduceMoney(iMoneyType, iMoney); !ok {
		return 2, 0, 0, 0
	}

	t.user.Debug("TreasureBoxInfo:", info)
	if ok := t.treasureBoxUtil[uint64(id)].SetTreasureBoxInfo(info); !ok {
		return 0, 0, 0, 0
	}

	t.user.storeMgr.GetGoods(itemId, 1, common.RS_TreasureBox, iMoneyType, iMoney) //common.RS_TreasureBox开宝箱
	t.user.tlogTreasureBoxFlow(id, info.TotalNum, iMoneyType, iMoney, itemId)      //tlog道具流水表

	return 1, itemId, moneyType, money //(第一个变量 0:失败 1:领取成功 2:货币不够)
}

// judgeReduceMoney 判断货币是否足够进行抽奖，如果货币足够进行扣除
func (t *TreasureBoxMgr) judgeReduceMoney(iMoneyType, iMoney uint32) bool {
	switch iMoneyType {
	case common.MT_MONEY:
		coin := t.user.GetCoin()

		if coin < uint64(iMoney) {
			return false
		} else {
			t.user.storeMgr.reduceMoney(common.MT_MONEY, common.RS_TreasureBox, uint64(iMoney))
		}
	case common.MT_DIAMOND:
		diamond := t.user.GetDiam()

		if diamond < uint64(iMoney) {
			return false
		} else {
			t.user.storeMgr.reduceMoney(common.MT_DIAMOND, common.RS_TreasureBox, uint64(iMoney))
		}
	case common.MT_BraveCoin:
		braveCoin := t.user.GetBraveCoin()

		if braveCoin < uint64(iMoney) {
			return false
		} else {
			t.user.storeMgr.reduceMoney(common.MT_BraveCoin, common.RS_TreasureBox, uint64(iMoney))
		}
	default:
		t.user.Warn("iMoneyType is err! moneyType:", iMoneyType, " iMoney:", iMoney)
		return false
	}

	return true
}

// randomItemID 随机生成一个道具
func (t *TreasureBoxMgr) randomItem(id uint64, totalNum uint32) (uint32, uint32) {
	if t.treasureBoxInfo[uint64(id)] == nil {
		t.user.Warn("TreasureBoxInfo nil! id:", id)
		return 0, 0
	}

	var weightPool, weightItem uint32
	var listPool []MegaPoolInfo
	var listItem []ItemInfo

	for _, v := range t.treasureBoxInfo[id].megaPoolInfo {
		if totalNum >= v.minNum {
			weightPool += v.weight
			listPool = append(listPool, v)
		}
	}
	poolId := t.weightChoice(weightPool, 1, listPool)

	for _, v := range t.treasureBoxInfo[id].itemInfo[poolId] {
		if poolId == v.poolId {
			weightItem += v.weight
			listItem = append(listItem, v)
		}
	}
	itemId := t.weightChoice(weightItem, 2, listItem)

	return itemId, poolId
}

// weightChoice 加权随机算法
func (t *TreasureBoxMgr) weightChoice(weight, typ uint32, list interface{}) uint32 {
	if weight <= 0 {
		t.user.Warn("weight:", weight)
		return 0
	}
	rad := rand.Int31n(int32(weight))

	var curTotal uint32
	switch typ {
	case 1:
		if listPool, ok := list.([]MegaPoolInfo); ok {
			for _, v := range listPool {
				curTotal += v.weight
				if uint32(rad) < curTotal {
					return v.poolId
				}
			}
		}
	case 2:
		if listItem, ok := list.([]ItemInfo); ok {
			for _, v := range listItem {
				curTotal += v.weight
				if uint32(rad) < curTotal {
					return v.itemId
				}
			}
		}
	}
	return 0
}

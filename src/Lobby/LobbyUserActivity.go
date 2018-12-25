package main

import (
	"common"
	"db"
	"excel"
	"fmt"
	"protoMsg"
	"strconv"
	"strings"
	"time"
	"zeus/iserver"

	log "github.com/cihub/seelog"
)

const (
	signActivityID        = 1  // 签到活动id
	updateAcitityID       = 2  // 更新活动id
	firstWinActivityID    = 3  // 每天首胜活动id
	threeDayActivityID    = 4  // 3日签到活动id
	platWelfareID         = 5  // 平台福利
	newYearActivityID     = 6  // 新年活动
	VeteranRecallID       = 7  // 老兵召回活动
	recruitActivityID     = 8  // 新兵招募
	bindComradeActivityID = 9  // 战友绑定
	backBattleID          = 10 // 重返光荣活动
	worldCupChampion      = 11
	worldCupMatch         = 12
	ballStarActivityID    = 13 // 一球成名活动
	festivalActivityID    = 14 // 节日活动
	exchangeActivityID    = 15 // 兑换活动
)

const (
	Finish     = 0 //领取成功
	Failure    = 1 //领取失败
	HasChecked = 2 //已领取
	NoStart    = 3 //活动未开始
	OutDate    = 4 //已过期
	ConfigErr  = 5 //配表出错
)

// Activeinfo 活动信息
type Activeinfo struct {
	id                   uint64
	tmSartShow           int64
	tmEndShow            int64
	tmStart              []int64
	tmEnd                []int64
	pickNum              []uint32 //活动领取次数
	iosrewardVersion     string
	androidrewardVersion string
	activityType         uint32
}

func (activeinfo *Activeinfo) String() string {
	return fmt.Sprintf("%+v\n", *activeinfo)
}

// ActivityMgr 活动管理器
type ActivityMgr struct {
	user *LobbyUser

	activeInfo   map[uint64]*Activeinfo      //活动信息状态
	activityUtil map[uint64]*db.ActivityUtil //活动工具
}

// NewActivityMgr 获取活动管理器
func NewActivityMgr(user *LobbyUser) *ActivityMgr {
	activity := &ActivityMgr{
		user:         user,
		activeInfo:   make(map[uint64]*Activeinfo),
		activityUtil: make(map[uint64]*db.ActivityUtil),
	}

	activity.initActiveinfo()
	return activity
}

// initActiveinfo 初始化管理器
func (a *ActivityMgr) initActiveinfo() {
	promotion := excel.GetPromotionMap()

	for _, v := range promotion {
		activeinfo := &Activeinfo{}

		activeinfo.id = v.Id
		startshowDate, err := time.ParseInLocation("2006|01|02|15|04|05", v.StartshowDate, time.Local)
		if err != nil {
			a.user.Error("SetActiveCheckin StartDate:", err)
			continue
		}
		activeinfo.tmSartShow = startshowDate.Unix()

		endshowDate, err := time.ParseInLocation("2006|01|02|15|04|05", v.EndshowDate, time.Local)
		if err != nil {
			a.user.Error("SetActiveCheckin EndDate:", err)
			continue
		}
		activeinfo.tmEndShow = endshowDate.Unix()

		startDate := strings.Split(v.StartDate, ";")
		endDate := strings.Split(v.EndDate, ";")
		days := strings.Split(v.Days, ";")
		if len(startDate) != len(endDate) || len(startDate) != len(days) || len(endDate) != len(days) {
			a.user.Error("len(startDate) != len(endDate) != len(days) ActivityID = ", v.Id)
			continue
		}

		for i, _ := range startDate {
			tmStart, err := time.ParseInLocation("2006|01|02|15|04|05", startDate[i], time.Local)
			if err != nil {
				a.user.Error("SetActiveCheckin StartDate:", err)
			}

			tmEnd, err := time.ParseInLocation("2006|01|02|15|04|05", endDate[i], time.Local)
			if err != nil {
				a.user.Error("SetActiveCheckin EndDate:", err)
			}

			pickNum, err := strconv.ParseUint(days[i], 10, 32)
			if err != nil {
				a.user.Error("ParseUint:", err, " ", pickNum)
			}

			activeinfo.tmStart = append(activeinfo.tmStart, tmStart.Unix())
			activeinfo.tmEnd = append(activeinfo.tmEnd, tmEnd.Unix())
			activeinfo.pickNum = append(activeinfo.pickNum, uint32(pickNum))
		}
		activeinfo.iosrewardVersion = v.IosrewardVersion
		activeinfo.androidrewardVersion = v.AndroidrewardVersion
		activeinfo.activityType = uint32(v.Activitytype)

		a.activeInfo[v.Id] = activeinfo
		a.activityUtil[v.Id] = db.PlayerActivityUtil(a.user.GetDBID(), uint32(v.Id))
	}

	log.Info("activeInfo:", a.activeInfo, " activityUtil:", a.activityUtil)
}

// syncActiveState 同步活动状态
func (a *ActivityMgr) syncActiveState() {
	a.user.Debug("syncActiveState")
	a.checkVeteran() //初始化老兵状态

	activitysInfo := &protoMsg.ActivitysInfo{}
	for _, v := range a.activeInfo {
		if v == nil {
			continue
		}

		activityState := &protoMsg.ActivityState{
			Id: uint32(v.id),
		}
		switch v.activityType {
		case 0: // 普通活动
			activitysInfo.ActivityState = append(activitysInfo.ActivityState, activityState)
		case 1: // 兑换活动
			activitysInfo.ExchangeState = append(activitysInfo.ExchangeState, activityState)
		}

		isVersion := a.checkVersion(v.iosrewardVersion, v.androidrewardVersion)
		if !isVersion {
			continue
		}

		if time.Now().Unix() < v.tmSartShow || time.Now().Unix() > v.tmEndShow {
			continue
		}

		if v.id == recruitActivityID && !a.user.needDisplayReceivePupil() {
			continue
		}

		if v.id == bindComradeActivityID && !a.user.needDisplayTakeTeacher() {
			continue
		}

		if v.id == backBattleID && a.user.GetVeteran() != 1 {
			continue
		}

		activityState.State = 1
		if a.activityUtil[v.id] != nil {
			activityState.RedDot = a.activityUtil[v.id].GetRedDot(a.user.version)
		}
	}

	a.user.Info("activityState.State close(0), open(1): ", activitysInfo)
	// 活动是否能开启
	if err := a.user.RPC(iserver.ServerTypeClient, "InitCampaignsState", activitysInfo); err != nil {
		a.user.Error(err)
	}
}

func (a *ActivityMgr) syncActiveInfo() {
	a.syncSignActivityID()      //同步签到活动已经领取到的id
	a.syncUpdateActivity()      //同步更新活动信息
	a.syncFirstWinActivity()    //同步首胜活动信息
	a.syncThreeDayActivityID()  //同步3天签到活动信息
	a.syncPlatWelfareActivity() //同步平台福利活动信息
	a.syncNewYearActivity()     //同步新年活动信息
	a.syncVeteranRecall()       //同步老兵召回的平台好友信息列表
	a.syncBackBattleInfo()      //同步已经领取重返光荣任务的id
	a.syncFestivalInfo()        //同步已经领取到任务的id
	a.syncExchangeInfo()        //同步兑换活动信息
	a.syncBallStarInfo()        //同步一球成名活动信息
}

// getCurActivityTime 获取当前活动时间段id
func (a *ActivityMgr) getCurActivityTime(activeinfo *Activeinfo) int {
	if activeinfo == nil {
		return 0
	}

	if len(activeinfo.tmStart) != len(activeinfo.tmEnd) || len(activeinfo.tmStart) != len(activeinfo.pickNum) || len(activeinfo.tmEnd) != len(activeinfo.pickNum) {
		a.user.Warn("getCurActivityTime fail")
		return 0
	}

	var tmp int = 0
	for i, k := range activeinfo.tmEnd {
		if time.Now().Unix() <= k {
			return i
		}
		tmp = i
	}

	return tmp
}

// checkVersion 检测版本号 小于版本号(false), 大于等于版本号(true)
func (a *ActivityMgr) checkVersion(iosrewardVersion, androidrewardVersion string) bool {
	a.user.Info("checkVersion")
	if a.user.version == "" {
		a.user.Error("user version nil!")
		return false
	}

	var clientVersion, rewardVersion []string
	clientVersion = strings.Split(a.user.version, ".")
	if a.user.platID == 0 {
		rewardVersion = strings.Split(iosrewardVersion, ".")
	} else if a.user.platID == 1 {
		rewardVersion = strings.Split(androidrewardVersion, ".")
	}

	a.user.Info("ClientVersion:", clientVersion, " IosrewardVersion:", iosrewardVersion, " AndroidrewardVersion:", androidrewardVersion, " rewardVersion:", rewardVersion)

	if len(clientVersion) != len(rewardVersion) {
		a.user.Error("len(clientVersion):", len(clientVersion), " != len(rewardVersion):", len(rewardVersion))
		return false
	}

	//更新后才可领取
	for i, k := range clientVersion {
		cVersion, err := strconv.Atoi(k)
		if err != nil {
			a.user.Error("err:", err)
			return false
		}
		rVersion, err := strconv.Atoi(rewardVersion[i])
		if err != nil {
			a.user.Error("err:", err)
			return false
		}

		if cVersion < rVersion {
			return false
		} else if cVersion > rVersion {
			return true
		}
	}

	return true
}

// checkShowDate 检测活动页签是否显示
func (a *ActivityMgr) checkShowDate(id uint64) bool {
	if time.Now().Unix() < a.activeInfo[festivalActivityID].tmSartShow {
		return false
	}

	if time.Now().Unix() > a.activeInfo[festivalActivityID].tmEndShow {
		return false
	}

	return true
}

// checkOpenDate 检测活动是否开启(0:开启, NoStart:未开启 , OutDate:已过期, ConfigErr:配表出错)
func (a *ActivityMgr) checkOpenDate(id uint64) (uint32, int) {
	if a.activeInfo[id] == nil || a.activityUtil[id] == nil {
		a.user.Error("activeInfo:", a.activeInfo[id], " activityUtil:", a.activityUtil[id])
		return ConfigErr, 0
	}

	index := a.getCurActivityTime(a.activeInfo[id])
	if index >= len(a.activeInfo[id].tmStart) || index >= len(a.activeInfo[id].tmEnd) || index >= len(a.activeInfo[id].pickNum) {
		a.user.Error("tmStart:", len(a.activeInfo[id].tmStart), " tmEnd:", len(a.activeInfo[id].tmEnd), " pickNum:", len(a.activeInfo[id].pickNum), " index:", index+1)
		return ConfigErr, index
	}

	if time.Now().Unix() < a.activeInfo[id].tmStart[index] {
		return NoStart, index //未开始
	}
	if time.Now().Unix() > a.activeInfo[id].tmEnd[index] {
		return OutDate, index //已过期
	}

	return 0, index
}

// addRewardToBag 添加领取的物品到包裹
func (a *ActivityMgr) addRewardToBag(rewards string, tlogId uint32) bool {

	awardsMap, err := common.SplitReward(rewards)
	if err != nil {
		a.user.Error("splitReward err:", err)
		return false
	}

	for k, v := range awardsMap {
		a.user.storeMgr.GetGoods(k, v, tlogId, common.MT_NO, 0)
	}

	return true
}

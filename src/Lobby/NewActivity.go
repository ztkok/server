package main

import (
	"excel"
	"strconv"
	"strings"
	"time"

	log "github.com/cihub/seelog"
)

func (srv *LobbySrv) InitActivityManager() {
	srv.activityMgr = &ActivityManger{
		activities: make(map[uint64]IActivity),
	}

	srv.activityMgr.Init()
}

type ActivityManger struct {
	activities map[uint64]IActivity
}

func (actMgr *ActivityManger) Init() {
	actMgr.activities[worldCupChampion] = NewWorldCupChampionActivity()
	actMgr.activities[worldCupMatch] = NewWorldCupMatchActivity()
	for _, act := range actMgr.activities {
		act.init()
	}
	log.Info("activeInfo:", actMgr.activities)
}
func (actMgr *ActivityManger) SyncState() {

}
func (actMgr *ActivityManger) GetActivity(id uint64) IActivity {
	return actMgr.activities[id]
}
func (actMgr *ActivityManger) Loop() {
	for _, act := range actMgr.activities {
		t, _ := act.checkOpenDate()
		if t == 0 {
			if time.Now().Minute() == 0 && time.Now().Second() == 0 {
				act.doHour()
			}
			if time.Now().Hour() == 0 && time.Now().Minute() == 0 && time.Now().Second() == 0 {
				act.doDay()
			}
		}
		if t == OutDate {
			act.endReward()
		}
	}

}

type IActivity interface {
	checkOpenDate() (uint32, int)
	checkShowDate() bool
	doHour()
	doDay()
	init()
	endReward()
}

type Activity struct {
	id uint64
	Activeinfo
}

func (act *Activity) init() {
	v, ok := excel.GetPromotion(act.id)
	if !ok {
		return
	}

	act.id = v.Id
	startshowDate, err := time.ParseInLocation("2006|01|02|15|04|05", v.StartshowDate, time.Local)
	if err != nil {
		return
	}
	act.tmSartShow = startshowDate.Unix()

	endshowDate, err := time.ParseInLocation("2006|01|02|15|04|05", v.EndshowDate, time.Local)
	if err != nil {
		return
	}
	act.tmEndShow = endshowDate.Unix()

	startDate := strings.Split(v.StartDate, ";")
	endDate := strings.Split(v.EndDate, ";")
	days := strings.Split(v.Days, ";")
	if len(startDate) != len(endDate) || len(startDate) != len(days) || len(endDate) != len(days) {
		return
	}

	for i := range startDate {
		tmStart, err := time.ParseInLocation("2006|01|02|15|04|05", startDate[i], time.Local)
		if err != nil {
		}

		tmEnd, err := time.ParseInLocation("2006|01|02|15|04|05", endDate[i], time.Local)
		if err != nil {
		}

		pickNum, err := strconv.ParseUint(days[i], 10, 32)
		if err != nil {
		}

		act.tmStart = append(act.tmStart, tmStart.Unix())
		act.tmEnd = append(act.tmEnd, tmEnd.Unix())
		act.pickNum = append(act.pickNum, uint32(pickNum))
	}
	act.iosrewardVersion = v.IosrewardVersion
	act.androidrewardVersion = v.AndroidrewardVersion
	act.activityType = uint32(v.Activitytype)

}
func (act *Activity) checkShowDate() bool {
	if time.Now().Unix() < act.tmSartShow {
		return false
	}

	if time.Now().Unix() > act.tmEndShow {
		return false
	}

	return true
}
func (act *Activity) getCurActivityTime() int {
	if len(act.tmStart) != len(act.tmEnd) || len(act.tmStart) != len(act.pickNum) || len(act.tmEnd) != len(act.pickNum) {
		return 0
	}

	var tmp int
	for i, k := range act.tmEnd {
		if time.Now().Unix() <= k {
			return i
		}
		tmp = i
	}

	return tmp
}
func (act *Activity) checkOpenDate() (uint32, int) {
	index := act.getCurActivityTime() //TODO 目的是什么
	if index >= len(act.tmStart) || index >= len(act.tmEnd) || index >= len(act.pickNum) {
		return ConfigErr, index
	}

	if time.Now().Unix() < act.tmStart[index] {
		return NoStart, index //未开始
	}
	if time.Now().Unix() > act.tmEnd[index] {
		return OutDate, index //已过期
	}

	return 0, index
}

func (act *Activity) doHour() {

}
func (act *Activity) endReward() {

}
func (act *Activity) doDay() {

}

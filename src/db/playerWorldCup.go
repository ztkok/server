package db

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/cihub/seelog"
)

func NewPlayerWorldCupUtil(uid uint64) *playerWorldCupUtil {
	return &playerWorldCupUtil{
		uid: uid,
	}
}

const (
	PlayerWorldCupInfo         = "PlayerWCupInfo"            //玩家的世界杯活动数据
	WCupChampionOdds           = "WCupChampionOdds"          //冠军竞猜活动的队伍赔率
	WCupChampionBonus          = "WCupChampionBonus"         //冠军竞猜的队伍下注总数
	PlayerWorldChampionRecord  = "PlayerWCupChampionRecord"  //冠军竞猜玩家下注记录
	PlayerWorldChampionContest = "PlayerWCupChampionContest" //冠军竞猜玩家下注数据
	PlayerWorldMatchRecord     = "PlayerWCupMatchRecord"     //胜负竞猜玩家下注记录
	PlayerWorldMatchContest    = "PlayerWCupMatchContest"    //胜负竞猜玩家下注数据
	WCupMatchOdds              = "WCupMatchOdds"
	WCupMatchBonus             = "WCupMatchBonus"
	PlayerWorldMatchReward     = "PlayerWCupMatchReward"
)

type playerWorldCupUtil struct {
	uid uint64
}

func (util *playerWorldCupUtil) recordKey() string {
	return fmt.Sprintf("%s:%d", PlayerWorldChampionRecord, util.uid)
}
func (util *playerWorldCupUtil) contestKey() string {
	return fmt.Sprintf("%s:%d", PlayerWorldChampionContest, util.uid)
}
func (util *playerWorldCupUtil) worldCupInfoKey() string {
	return fmt.Sprintf("%s:%d", PlayerWorldCupInfo, util.uid)
}

func (util *playerWorldCupUtil) SetChampionOdds(odds map[uint32]float32, expire int64) {
	s, e := json.Marshal(odds)
	if e == nil {
		if setNx(WCupChampionOdds, string(s)) == 1 {
			expireKey(WCupChampionOdds, expire)
		}
	}
}
func (util *playerWorldCupUtil) GetChampionOdds() map[uint32]float32 {
	odds := get(WCupChampionOdds)
	ret := map[uint32]float32{}
	e := json.Unmarshal([]byte(odds), &ret)
	if e != nil {
		log.Error("unmarshal Error ", odds, e)
	}
	return ret
}

func (util *playerWorldCupUtil) IncryTeamBonus(id uint32, num uint64) {
	hIncrBy(WCupChampionBonus, id, num)
}

func (util *playerWorldCupUtil) GetTeamBonus() map[string]string {
	return hGetAll(WCupChampionBonus)
}

// AddContest
func (util *playerWorldCupUtil) AddContest(teamId uint32) uint64 {
	now := hIncrBy(util.contestKey(), teamId, 1)
	util.AddContestTime(teamId)
	util.IncryTeamBonus(teamId, 1)
	return uint64(now)
}

// GetContest
func (util *playerWorldCupUtil) GetContest() map[string]string {
	return hGetAll(util.contestKey())
}

// AddContestTime
func (util *playerWorldCupUtil) AddContestTime(teamId uint32) {
	hIncrBy(util.worldCupInfoKey(), "championContestTime", 1)
}

// GetContestTime
func (util *playerWorldCupUtil) GetContestTime(teamId uint32) string {
	r := hGet(util.worldCupInfoKey(), "championContestTime")
	day := hGet(util.worldCupInfoKey(), "championContestDay")
	if day == "" || day != time.Now().Format("20060102") {
		hSet(util.worldCupInfoKey(), "championContestDay", time.Now().Format("20060102"))
		hSet(util.worldCupInfoKey(), "championContestTime", 0)
		return ""
	}
	return r
}

type WorldCupContestRecord struct {
	Id    uint32  `json:"id"`
	Stamp int64   `json:"stamp"`
	Odds  float32 `json:"odds"`
}

func (util *playerWorldCupUtil) AddContestRecord(teamId uint32, odds float32) {
	r := &WorldCupContestRecord{
		Id:    teamId,
		Stamp: time.Now().Unix(),
		Odds:  odds,
	}
	s, e := json.Marshal(r)
	if e != nil {

	}
	rPush(util.recordKey(), string(s))
}
func (util *playerWorldCupUtil) GetContestRecord() []*WorldCupContestRecord {
	var ret []*WorldCupContestRecord
	for _, v := range lRangeAll(util.recordKey()) {
		r := &WorldCupContestRecord{}
		e := json.Unmarshal([]byte(v), r)
		if e != nil {
			continue
		}
		ret = append(ret, r)
	}
	return ret
}

// AddContestMatchTime
func (util *playerWorldCupUtil) AddContestMatchTime(matchId uint32) {
	hIncrBy(util.worldCupInfoKey(), "matchContestTime", 1)
}

// GetContestMatchTime
func (util *playerWorldCupUtil) GetContestMatchTime() string {
	r := hGet(util.worldCupInfoKey(), "matchContestTime")
	day := hGet(util.worldCupInfoKey(), "matchContestDay")
	if day == "" || day != time.Now().Format("20060102") {
		hSet(util.worldCupInfoKey(), "matchContestDay", time.Now().Format("20060102"))
		hSet(util.worldCupInfoKey(), "matchContestTime", 0)
		return ""
	}
	return r
}
func (util *playerWorldCupUtil) IsChampionReward() bool {
	return hExists(util.worldCupInfoKey(), "matchChampionDone")
}

func (util *playerWorldCupUtil) ChampionReward() {
	hSet(util.worldCupInfoKey(), "matchChampionDone", 1)
}
func (util *playerWorldCupUtil) matchContestKey() string {
	return fmt.Sprintf("%s:%d", PlayerWorldMatchContest, util.uid)
}
func (util *playerWorldCupUtil) matchRecordKey() string {
	return fmt.Sprintf("%s:%d", PlayerWorldMatchRecord, util.uid)
}
func (util *playerWorldCupUtil) matchRewardKey() string {
	return fmt.Sprintf("%s:%d", PlayerWorldMatchReward, util.uid)
}

func (util *playerWorldCupUtil) GetMatchContest() map[string][]uint64 {
	ret := map[string][]uint64{}
	for k, v := range hGetAll(util.matchContestKey()) {
		var d []uint64
		json.Unmarshal([]byte(v), &d)
		ret[k] = d
	}
	return ret
}
func (util *playerWorldCupUtil) AddMatchContest(matchId, index uint32) []uint64 {
	contestData := hGet(util.matchContestKey(), matchId)
	var info [3]uint64
	json.Unmarshal([]byte(contestData), &info)
	info[index]++
	s, e := json.Marshal(info)
	if e != nil {
		log.Error("addMatchContest Error ", e)
	}
	hSet(util.matchContestKey(), matchId, string(s))
	util.AddContestMatchTime(matchId)
	mId := fmt.Sprintf("%d_%d", matchId, index)
	hIncrBy(WCupMatchBonus, mId, 1)
	return info[:]
}

type WorldCupMatchContestRecord struct {
	MatchId uint32  `json:"match_id"`
	Index   uint32  `json:"index"`
	Stamp   int64   `json:"stamp"`
	Odds    float32 `json:"odds"`
}

func (util *playerWorldCupUtil) AddMatchContestRecord(matchId, index uint32, odds float32) {
	r := &WorldCupMatchContestRecord{
		MatchId: matchId,
		Stamp:   time.Now().Unix(),
		Index:   index,
		Odds:    odds,
	}
	s, e := json.Marshal(r)
	if e != nil {

	}
	rPush(util.matchRecordKey(), string(s))
}

func (util *playerWorldCupUtil) GetMatchContestRecord() []*WorldCupMatchContestRecord {
	var ret []*WorldCupMatchContestRecord
	for _, v := range lRangeAll(util.matchRecordKey()) {
		r := &WorldCupMatchContestRecord{}
		e := json.Unmarshal([]byte(v), r)
		if e != nil {
			continue
		}
		ret = append(ret, r)
	}
	return ret
}

func (util *playerWorldCupUtil) GetMatchBonus(matchId uint32) map[string]string {
	var keys []string
	for i := 0; i < 3; i++ {
		keys = append(keys, fmt.Sprintf("%d_%d", matchId, i))
	}
	return hMGet(WCupMatchBonus, keys)
}

func (util *playerWorldCupUtil) SetMatchOdds(odds []float32, matchId uint32, expire int64) {
	s, e := json.Marshal(odds)
	key := fmt.Sprintf("%s:%d", WCupMatchOdds, matchId)
	if e == nil {
		if setNx(key, string(s)) == 1 && expire > 0 {
			expireKeyAt(key, expire)
		}
	}
}

func (util *playerWorldCupUtil) GetMatchOdds(matchId uint32) []float32 {
	key := fmt.Sprintf("%s:%d", WCupMatchOdds, matchId)
	ss := get(key)
	var ret []float32
	if ss == "" {
		return ret
	}
	e := json.Unmarshal([]byte(ss), &ret)
	if e != nil {
		log.Error("GetMatchOdds Error ", ss, e)
	}
	return ret
}
func (util *playerWorldCupUtil) SetMatchReward(matchId uint32, num float32) {
	hSet(util.matchRewardKey(), matchId, num)
}

func (util *playerWorldCupUtil) IsMatchReward(matchId uint32) bool {
	return hExists(util.matchRewardKey(), matchId)
}

func (util *playerWorldCupUtil) GetMatchRewardAll() map[string]string {
	return hGetAll(util.matchRewardKey())
}

func (util *playerWorldCupUtil) ClearAll() {
	delKey("WCupChampionOdds")
	delKey(WCupChampionBonus)
	delKey(WCupMatchBonus)
	for i := 1; i < 65; i++ {
		delKey(fmt.Sprintf(WCupMatchOdds+":%d", i))
	}
	keys := keys("PlayerWCup*")
	for _, k := range keys {
		delKey(k)
	}
}

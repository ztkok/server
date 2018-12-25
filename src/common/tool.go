package common

import (
	"datadef"
	"db"
	"encoding/binary"
	"errors"
	"excel"
	"fmt"
	"math"
	"math/rand"
	"protoMsg"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"zeus/dbservice"
	"zeus/linmath"
	"zeus/msgdef"

	log "github.com/cihub/seelog"
	"github.com/garyburd/redigo/redis"
	"github.com/spf13/viper"
)

func StringToUint64(i string) uint64 {
	d, e := strconv.ParseUint(i, 10, 64)
	if e != nil {
		log.Info("string convent err ", i, e)
		return 0
	}
	return d
}

func StringToInt64(i string) int64 {
	d, e := strconv.ParseInt(i, 10, 64)
	if e != nil {
		log.Info("string convent err ", i, e)
		return 0
	}
	return d
}

func StringToUint32(i string) uint32 {
	if i == "" {
		return 0
	}
	d, e := strconv.ParseUint(i, 10, 64)
	if e != nil {
		log.Info("string convent err ", i, e)
		return 0
	}
	return uint32(d)
}

func StringToInt(i string) int {
	d, e := strconv.ParseInt(i, 10, 64)
	if e != nil {
		log.Info("string convent err ", i, e)
		return 0
	}
	return int(d)
}

func StringToUint8(i string) uint8 {
	d, e := strconv.ParseUint(i, 10, 64)
	if e != nil {
		log.Info("string convent err ", i, e)
		return 0
	}
	return uint8(d)
}

func StringToFloat64(i string) float64 {
	if i == "" {
		return 0
	}
	d, e := strconv.ParseFloat(i, 64)
	if e != nil {
		log.Info("string convent err ", i, e)
		return 0
	}
	return d
}

func StringToFloat32(i string) float32 {
	d, e := strconv.ParseFloat(i, 64)
	if e != nil {
		log.Info("string convent err ", i, e)
		return 0
	}
	return float32(d)
}

func Uint64ToString(i uint64) string {
	return strconv.FormatUint(i, 10)
}

func UintToString(i uint) string {
	return strconv.FormatUint(uint64(i), 10)
}

func Uint32ToString(i uint32) string {
	return strconv.FormatUint(uint64(i), 10)
}

func Int64ToString(i int64) string {
	return strconv.FormatInt(i, 10)
}

func IntToString(i int) string {
	return strconv.FormatInt(int64(i), 10)
}

//进行四舍五入，保留n位小数
func RoundFloat(f float64, n int) float64 {
	pow10_n := math.Pow10(n)
	return math.Trunc((f+0.5/pow10_n)*pow10_n) / pow10_n
}

//不进行四舍五入，保留n位小数
func NoRoundFloat(f float64, n int) float64 {
	pow10_n := math.Pow10(n)
	return math.Trunc(f*pow10_n) / pow10_n
}

//[]uint32转化为[]byte
func Uint32sToBytes(args []uint32) []byte {
	bytes := make([]byte, 0)
	for _, arg := range args {
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, arg)
		bytes = append(bytes, b...)
	}
	return bytes
}

//[]byte转化为[]uint32
func BytesToUint32s(bytes []byte) []uint32 {
	res := make([]uint32, 0)
	for i := 0; i < len(bytes)/4; i++ {
		b := bytes[4*i : 4*i+4]
		res = append(res, binary.BigEndian.Uint32(b))
	}
	return res
}

func StringToMapInt(str, first, second string) map[interface{}]int {
	data := map[interface{}]int{}
	tmp1 := strings.Split(str, first)
	for _, v := range tmp1 {
		tmp2 := strings.Split(v, second)
		if len(tmp2) != 2 {
			continue
		}
		data[tmp2[0]] = StringToInt(tmp2[1])
	}
	return data
}

func StringToMapUint32(str, sep1, sep2 string) map[uint32]uint32 {
	res := make(map[uint32]uint32)
	ss := strings.Split(str, sep1)

	for _, s := range ss {
		ss2 := strings.Split(s, sep2)
		if len(ss2) != 2 {
			continue
		}

		res[StringToUint32(ss2[0])] = StringToUint32(ss2[1])
	}

	return res
}

func MapUint32ToString(m map[uint32]uint32, sep1, sep2 string) string {
	var res string

	for k, v := range m {
		res += Uint32ToString(k)
		res += sep2
		res += Uint32ToString(v)
		res += sep1
	}

	if len(res) > 0 {
		return res[:len(res)-1]
	}

	return ""
}

func CombineMapUint32(m1, m2 map[uint32]uint32) map[uint32]uint32 {
	res := make(map[uint32]uint32)

	for k, v := range m1 {
		res[k] = v + m2[k]
	}

	for k, v := range m2 {
		if _, ok := res[k]; !ok {
			res[k] = v
		}
	}

	return res
}

func MapUint32ToSlice(m map[uint32]uint32) []uint32 {
	res := make([]uint32, 0)

	for k, v := range m {
		res = append(res, k)
		res = append(res, v)
	}

	return res
}

func InitMsg() {
	def := msgdef.GetMsgDef()
	for k, v := range ProtoMap {
		def.RegMsg(k, v.Name(), v)
	}
}

func Distance(v linmath.Vector3, o linmath.Vector3) float32 {
	dx := v.X - o.X
	dz := v.Z - o.Z

	return float32(math.Sqrt(float64(dx*dx + dz*dz)))
}

func CreateNewMailID() uint64 {
	u, err := dbservice.UIDGenerator().Get("mail")
	if err != nil {
		return 0
	}
	return u
}

func GetDefiniteequinox(x1 float32, x2 float32, rate float32) float32 {
	return (x1 + rate*x2) / (1 + rate)
}

func GetTBSystemValue(id uint64) uint {
	base, ok := excel.GetSystem(id)
	if !ok {
		return 0
	}

	return uint(base.Value)
}

func GetTBTwoSystemValue(id uint64) string {
	base, ok := excel.GetSystem2(id)
	if !ok {
		return ""
	}

	return base.Value
}

func GetPaySystemValue(id uint64) string {
	base, ok := excel.GetPaysystem2(id)
	if !ok {
		return ""
	}

	return base.Value
}

func GetTBMatchValue(id uint64) float32 {
	base, ok := excel.GetMatch(id)
	if !ok {
		return 0
	}

	return base.Value
}

//获取每日定点执行的cron表达式
func GetDayFixedTimeCronExpr(h, m, s int64) string {
	expr := "* * ?"
	hs := strconv.FormatInt(h, 10)
	ms := strconv.FormatInt(m, 10)
	ss := strconv.FormatInt(s, 10)

	return ss + " " + ms + " " + hs + " " + expr
}

//返回今天零点对应的时间戳
func GetTodayBeginStamp() int64 {
	return GetDayBeginStamp(time.Now().Unix())
}

//返回零点对应的时间戳
func GetDayBeginStamp(stamp int64) int64 {
	t := time.Unix(stamp, 0)
	t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
	return t.Unix()
}

//返回本周一零点对应的时间戳
func GetThisWeekBeginStamp() int64 {
	t := time.Now()
	year, week := t.ISOWeek()
	date := time.Date(year, 0, 0, 0, 0, 0, 0, time.Local)
	isoYear, isoWeek := date.ISOWeek()

	for date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, -1)
		isoYear, isoWeek = date.ISOWeek()
	}

	for isoYear < year {
		date = date.AddDate(0, 0, 1)
		isoYear, isoWeek = date.ISOWeek()
	}

	for isoWeek < week {
		date = date.AddDate(0, 0, 1)
		isoYear, isoWeek = date.ISOWeek()
	}

	return date.Unix()
}

//判断一个时间戳是周几(1-7)
func WhatDayOfWeek(stamp int64) int {
	weekDays := map[string]int{
		"monday":    1,
		"tuesday":   2,
		"wednesday": 3,
		"thursday":  4,
		"friday":    5,
		"saturday":  6,
		"sunday":    7,
	}

	tm := time.Unix(stamp, 0)
	day := strings.ToLower(tm.Weekday().String())

	return weekDays[day]
}

//获取本周某一日零点对应的时间戳，whatDay表示周几(1-7)
func GetWeekDayBeginStamp(whatDay uint8) int64 {
	begin := GetThisWeekBeginStamp()
	day := int64(24 * 60 * 60)
	return begin + int64(whatDay-1)*day
}

//判断一个时间戳是否在本周的时间段内
func IsInThisWeek(stamp int64) bool {
	begin := GetThisWeekBeginStamp()
	if stamp < begin {
		return false
	}

	day := int64(24 * 60 * 60)
	if stamp >= begin+7*day {
		return false
	}

	return true
}

//获取下一个整点的时间戳
func GetNextHoursStamp() int64 {
	now := time.Now()
	return now.Unix() + int64((60-now.Minute())*60-now.Second())
}

//获取下一个整点的时间差
func GetNextHoursDur() int64 {
	now := time.Now()
	return int64((60-now.Minute())*60 - now.Second())
}

func TimeStringToTime(ts string) *time.Time {
	t, e := time.ParseInLocation("2006|01|02|15|04|05", ts, time.Local)
	if e != nil {
		log.Error("ParseInLocation err: ", e, ts)
	}
	return &t
}

// GetSeason 获取当前赛季id
func GetSeason() int {
	for _, v := range excel.GetSeasonMap() {
		start, err := time.ParseInLocation("2006|01|02", v.StartTime, time.Local)
		if err != nil {
			log.Error("ParseInLocation err: ", err)
			continue
		}

		end, err := time.ParseInLocation("2006|01|02", v.EndTime, time.Local)
		if err != nil {
			log.Error("ParseInLocation err: ", err)
			continue
		}

		now := time.Now().Unix()
		if now >= start.Unix() && now < end.Unix() {
			return int(v.Id)
		}
	}

	return 0
}

/*--------------------赛季排行---------------------*/
// GetRankSeason 获取当前排行赛季的id
func GetRankSeason() int {
	id := GetSeason()
	if id == 0 {
		return 0
	}

	data, ok := excel.GetSeason(uint64(id))
	if !ok {
		return 0
	}

	return int(data.Season)
}

// GetRankSeasonTimeStamp 获取当前排行赛季的开始和结束时间戳
func GetRankSeasonTimeStamp() (int64, int64) {
	id := GetSeason()
	if id == 0 {
		return 0, 0
	}

	data, ok := excel.GetSeason(uint64(id))
	if !ok {
		return 0, 0
	}

	start, err := time.ParseInLocation("2006|01|02", data.StartTime, time.Local)
	if err != nil {
		log.Error("ParseInLocation err: ", err)
		return 0, 0
	}

	end, err := time.ParseInLocation("2006|01|02", data.EndTime, time.Local)
	if err != nil {
		log.Error("ParseInLocation err: ", err)
		return 0, 0
	}

	return start.Unix(), end.Unix()
}

// GetLastRankSeason 获取刚结束的排行赛季的id
func GetLastRankSeason() int {
	cur := GetRankSeason()
	if cur == 0 {
		return 0
	}
	return cur - 1
}

// RankSeasonToSeason 根据排行赛季的id获取赛季id
func RankSeasonToSeason(rankSeason int) int {
	if rankSeason <= 0 {
		return 0
	}

	for _, v := range excel.GetSeasonMap() {
		if int(v.Season) == rankSeason {
			return int(v.Id)
		}
	}

	return 0
}

// GetSeasonAwardsByRank 根据玩家的赛季排名，获取对应的奖励
func GetSeasonAwardsByRank(season int, matchTyp uint8, rankTyp uint8, rank uint32) *protoMsg.SeasonAwards {
	awards := &protoMsg.SeasonAwards{}

	for _, v := range excel.GetAwardMap() {
		if v.Season != uint64(season) {
			continue
		}

		if v.Matchtyp != uint64(matchTyp) {
			continue
		}

		if v.Ranktyp != uint64(rankTyp) {
			continue
		}

		strs := strings.Split(v.Level, ";")
		if len(strs) != 1 && len(strs) != 2 {
			continue
		}

		if len(strs) == 1 && StringToUint32(strs[0]) != rank {
			continue
		}

		if len(strs) == 2 && (StringToUint32(strs[0]) > rank || StringToUint32(strs[1]) < rank) {
			continue
		}

		awardMap := StringToMapUint32(v.Awards, "|", ";")
		for id, num := range awardMap {
			awards.Items = append(awards.Items, &protoMsg.AwardItem{
				Id:  id,
				Num: num,
			})
		}
	}

	return awards
}

// GetPlayerCareerData 获取玩家不同类型的生涯数据
func GetPlayerCareerData(uid uint64, season int, typ uint8) (*datadef.CareerBase, error) {
	var (
		playerCareerDataTypStrs = map[uint8]string{
			0: PlayerCareerTotalData,
			1: PlayerCareerSoloData,
			2: PlayerCareerDuoData,
			4: PlayerCareerSquadData,
		}
		data = &datadef.CareerBase{}
	)

	if _, ok := playerCareerDataTypStrs[typ]; !ok {
		return nil, fmt.Errorf("Invalid playerCareerDataTyp: %d", typ)
	}

	if err := db.PlayerCareerDataUtil(playerCareerDataTypStrs[typ], uid, season).GetRoundData(data); err != nil {
		return nil, err
	}

	return data, nil
}

// CanDrawSeasonAwards 玩家是否有赛季奖励可以领取
func CanDrawSeasonAwards(uid uint64, season int) bool {
	total, _ := GetPlayerCareerData(uid, season, 0)
	if total != nil && total.TotalBattleNum >= 3 {
		return true
	}

	solo, _ := GetPlayerCareerData(uid, season, 1)
	if solo != nil && solo.TotalBattleNum >= 3 {
		return true
	}

	duo, _ := GetPlayerCareerData(uid, season, 2)
	if duo != nil && duo.TotalBattleNum >= 3 {
		return true
	}

	squad, _ := GetPlayerCareerData(uid, season, 4)
	if squad != nil && squad.TotalBattleNum >= 3 {
		return true
	}

	return false
}

/*--------------------匹配模式相关---------------------*/
var (
	modeInfosGlobal     ModeInfoSlice //全局变量，缓存了服务器当前开放的所有匹配模式信息，刷新时间间隔是1秒
	modeInfosGlobalLock sync.RWMutex  //用于保护modeInfosGlobal的并发读写
	lastTimeStamp       int64         //modeInfosGlobal上一次刷新的时间戳
)

//实现sort.Interface接口
type ModeInfoSlice []*ModeInfo

func (infos ModeInfoSlice) Len() int           { return len(infos) }
func (infos ModeInfoSlice) Swap(i, j int)      { infos[i], infos[j] = infos[j], infos[i] }
func (infos ModeInfoSlice) Less(i, j int) bool { return infos[i].UniqueId < infos[j].UniqueId }

//服务器当前开放的一条匹配模式的信息
type ModeInfo struct {
	ModeId      uint32 //匹配模式id
	UniqueId    uint32 //开放的比赛模式，在matchmode配置表中对应的唯一id
	OpenTyp     uint8  //开放的匹配类型，1表示单排，2表示双排，4表示四排
	SeasonStart int64  //当前赛季的开始时间戳
	SeasonEnd   int64  //当前赛季的结束时间戳
	DayStart    int64  //当前开放时间段的开始时间戳
	DayEnd      int64  //当前开放时间段的结束时间戳
	MatchNum    uint32 //当前赛季每一个开放时间段允许进行的比赛场次
	ItemId      uint32 //入场券道具id
	ItemNum     uint32 //消耗的入场券的数量
	ItemPrice   uint32 //入场券单价
	WaitTime    uint32 //每隔多少秒扩展一次rank分
	MapId       uint32 //地图id
}

func toModeInfo(modeId, uniqueId, matchNum, itemId, itemNum, itemPrice, wateTime uint32, seasonStart, seasonEnd, dayStart, dayEnd int64, openTyp uint8, mapId uint32) *ModeInfo {
	return &ModeInfo{
		ModeId:      modeId,
		UniqueId:    uniqueId,
		OpenTyp:     openTyp,
		SeasonStart: seasonStart,
		SeasonEnd:   seasonEnd,
		DayStart:    dayStart,
		DayEnd:      dayEnd,
		MatchNum:    matchNum,
		ItemId:      itemId,
		ItemNum:     itemNum,
		ItemPrice:   itemPrice,
		WaitTime:    wateTime,
		MapId:       mapId,
	}
}

//获取当前开放的所有匹配模式的信息，刷新modeInfosGlobal
func reflushModeInfosGlobal() {
	modeInfosGlobalLock.Lock()
	defer modeInfosGlobalLock.Unlock()

	var (
		matchM = excel.GetMatchmodeMap()
	)

	modeInfosGlobal = ModeInfoSlice{}

	for _, v := range matchM {
		if v.State == 0 {
			continue
		}

		var (
			seasonStart, seasonEnd, dayStart, dayEnd time.Time
			dayStartStamp, dayEndStamp               int64
			season, open                             bool
			err                                      error
		)

		if v.Seasonopentime != "" && v.Seasonclosetime != "" {
			seasonStart, err = time.ParseInLocation("2006|01|02|15|04|05", v.Seasonopentime, time.Local)
			if err != nil {
				log.Error("ParseInLocation err: ", err)
				continue
			}

			seasonEnd, err = time.ParseInLocation("2006|01|02|15|04|05", v.Seasonclosetime, time.Local)
			if err != nil {
				log.Error("ParseInLocation err: ", err)
				continue
			}

			season = true
		}

		dayStart, err = time.ParseInLocation("2006|01|02|15|04|05", v.Opentime, time.Local)
		if err != nil {
			log.Error("ParseInLocation err: ", err)
			continue
		}

		dayEnd, err = time.ParseInLocation("2006|01|02|15|04|05", v.Closetime, time.Local)
		if err != nil {
			log.Error("ParseInLocation err: ", err)
			continue
		}

		if season {
			open, dayStartStamp, dayEndStamp = isInOpenTime(seasonStart.Unix(), seasonEnd.Unix(), dayStart.Unix(), dayEnd.Unix(), int64(v.Looptime))
		} else {
			open, dayStartStamp, dayEndStamp = isInOpenTime(0, 0, dayStart.Unix(), dayEnd.Unix(), int64(v.Looptime))
		}

		if open {
			if season {
				modeInfosGlobal = append(modeInfosGlobal, toModeInfo(uint32(v.Modeid), uint32(v.Id), uint32(v.Matchnum), uint32(v.Itemid), uint32(v.Itemnum), uint32(v.Itemprice),
					uint32(v.Waittime), seasonStart.Unix(), seasonEnd.Unix(), dayStartStamp, dayEndStamp, uint8(v.Opennum), uint32(v.Mapid)))
			} else {
				modeInfosGlobal = append(modeInfosGlobal, toModeInfo(uint32(v.Modeid), uint32(v.Id), uint32(v.Matchnum), uint32(v.Itemid), uint32(v.Itemnum), uint32(v.Itemprice),
					uint32(v.Waittime), 0, 0, dayStartStamp, dayEndStamp, uint8(v.Opennum), uint32(v.Mapid)))
			}
		}
	}

	sort.Sort(modeInfosGlobal)
	lastTimeStamp = time.Now().Unix()
}

//判断当前时刻服务是否开放，若开放，则返回当前开放时间段开始和结束的时间戳
func isInOpenTime(seasonStart, seasonEnd, dayStart, dayEnd, loopDays int64) (bool, int64, int64) {
	var (
		now      int64 = time.Now().Unix()
		loopTime int64 = 24 * 60 * 60 * loopDays
		season   bool
	)

	if seasonStart != 0 && seasonEnd != 0 {
		season = true
		if now < seasonStart || now >= seasonEnd {
			return false, 0, 0
		}
	}

	for {
		if now < dayStart {
			return false, 0, 0
		}

		if now < dayEnd {
			return true, dayStart, dayEnd
		}

		if loopTime == 0 {
			return false, 0, 0
		}

		dayStart += loopTime
		dayEnd += loopTime

		if !season {
			continue
		}

		if dayStart >= seasonEnd || dayEnd >= seasonEnd {
			return false, 0, 0
		}
	}

	return false, 0, 0
}

//获取modeInfosGlobal的完整拷贝
func GetModeInfosGlobalCopy() ModeInfoSlice {
	checkForReflush()

	modeInfos := make(ModeInfoSlice, len(modeInfosGlobal))
	copy(modeInfos, modeInfosGlobal)

	return modeInfos
}

func checkForReflush() {
	if int64(math.Abs(float64(time.Now().Unix()-lastTimeStamp))) >= 1 {
		reflushModeInfosGlobal()
	}
}

//判断指定匹配模式和类型是否开放，matchMode表示匹配模式的id，详见Const.go
//matchTyp表示匹配的类型，1表示单排，2表示双排，4表示四排
func IsMatchModeOk(matchMode uint32, matchTyp uint8) bool {
	if matchMode == 0 {
		switch matchTyp {
		case 1:
			return viper.GetBool("Lobby.Solo")
		case 2:
			return viper.GetBool("Lobby.Duo")
		case 4:
			return viper.GetBool("Lobby.Squad")
		default:
			log.Error("Unknown match type: ", matchTyp)
			return false
		}
	}

	checkForReflush()

	modeInfosGlobalLock.RLock()
	defer modeInfosGlobalLock.RUnlock()

	for _, info := range modeInfosGlobal {
		if info.ModeId != matchMode {
			continue
		}

		if info.OpenTyp == matchTyp {
			return true
		}
	}

	return false
}

//获取特定匹配模式的当前开放信息，matchMode表示匹配模式的id，详见Const.go
//matchTyp表示匹配的类型，1表示单排，2表示双排，4表示四排
func GetOpenModeInfo(matchMode uint32, matchTyp uint8) *ModeInfo {
	checkForReflush()

	modeInfosGlobalLock.RLock()
	defer modeInfosGlobalLock.RUnlock()

	for _, info := range modeInfosGlobal {
		if info.ModeId != matchMode {
			continue
		}

		if info.OpenTyp == matchTyp {
			return info
		}
	}

	return nil
}

//在快速模式下，根据玩家的名次和匹配类型，获取指定的奖品
func GetFunRewards(rank uint32, typ uint8, goods ...uint32) map[uint32]uint32 {
	base, ok := excel.GetLuandoureward(uint64(rank))
	if !ok {
		return nil
	}

	var modeStr string

	switch typ {
	case 1:
		modeStr = base.Single
	case 2:
		modeStr = base.Two
	case 4:
		modeStr = base.Four
	default:
		return nil
	}

	rewards := make(map[uint32]uint32)

	strs := strings.Split(modeStr, "|")
	for i := 0; i < len(strs); i += 2 {
		for _, good := range goods {
			if StringToUint32(strs[i]) == good {
				rewards[good] = StringToUint32(strs[i+1])
			}
		}
	}

	return rewards
}

//在精英模式下，根据玩家的名次，发放指定的奖品
func GetDragonRewards(rank uint32, goods ...uint32) map[uint32]uint32 {
	base, ok := excel.GetDragonreward(uint64(rank))
	if !ok {
		return nil
	}

	rewards := make(map[uint32]uint32)

	strs := strings.Split(base.Single, "|")
	for i := 0; i < len(strs); i += 2 {
		for _, good := range goods {
			if StringToUint32(strs[i]) == good {
				rewards[good] = StringToUint32(strs[i+1])
			}
		}
	}

	return rewards
}

//在红蓝对决模式下，根据玩家的名次，发放指定的奖品
func GetRedAndBlueRewards(rank uint32, goods ...uint32) map[uint32]uint32 {
	base, ok := excel.GetRedandblue(uint64(rank))
	if !ok {
		return nil
	}

	rewards := make(map[uint32]uint32)

	strs := strings.Split(base.Single, "|")
	for i := 0; i < len(strs); i += 2 {
		for _, good := range goods {
			if StringToUint32(strs[i]) == good {
				rewards[good] = StringToUint32(strs[i+1])
			}
		}
	}

	return rewards
}

//在娱乐模式下，根据玩家的名次和匹配类型，获取指定的奖品
func GetFunmode(rank uint32, typ uint8, goods ...uint32) map[uint32]uint32 {
	base, ok := excel.GetFunmode(uint64(rank))
	if !ok {
		return nil
	}

	var modeStr string

	switch typ {
	case 1:
		modeStr = base.Single
	case 2:
		modeStr = base.Two
	case 4:
		modeStr = base.Four
	default:
		return nil
	}

	rewards := make(map[uint32]uint32)

	strs := strings.Split(modeStr, "|")
	for i := 0; i < len(strs); i += 2 {
		for _, good := range goods {
			if StringToUint32(strs[i]) == good {
				rewards[good] = StringToUint32(strs[i+1])
			}
		}
	}

	return rewards
}

//在娱乐模式下，根据玩家的名次和匹配类型，获取指定的奖品
func GetTankWarmode(rank uint32, typ uint8, goods ...uint32) map[uint32]uint32 {
	base, ok := excel.GetTankmode(uint64(rank))
	if !ok {
		return nil
	}

	var modeStr string

	switch typ {
	case 1:
		modeStr = base.Single
	case 2:
		modeStr = base.Two
	case 4:
		modeStr = base.Four
	default:
		return nil
	}

	rewards := make(map[uint32]uint32)

	strs := strings.Split(modeStr, "|")
	for i := 0; i < len(strs); i += 2 {
		for _, good := range goods {
			if StringToUint32(strs[i]) == good {
				rewards[good] = StringToUint32(strs[i+1])
			}
		}
	}

	return rewards
}

/*--------------------军衔---------------------*/

//获取配置的最大军衔等级
func GetMaxMilitaryLevel() uint32 {
	return uint32(excel.GetMilitaryrankMapLen())
}

//获取指定等级对应的军衔
func GetMilitaryRankByLevel(level uint32) string {
	data, ok := excel.GetMilitaryrank(uint64(level))
	if !ok {
		return ""
	}

	return data.Name
}

//获取指定等级下可以积累的最大经验值
func GetMaxExpByLevel(level uint32) uint32 {
	data, ok := excel.GetMilitaryrank(uint64(level))
	if !ok {
		return 0
	}

	return uint32(data.Value)
}

//获取达到指定等级得到的奖励
func GetAwardsByLevel(level uint32) map[uint32]uint32 {
	awards := make(map[uint32]uint32)

	data, ok := excel.GetMilitaryrank(uint64(level))
	if !ok {
		return awards
	}

	awards = StringToMapUint32(data.Award, "|", ";")
	return awards
}

//获取经验卡的配置参数(倍率，有效时间)
func GetExpCardArgs(cardId uint32) (float32, int64) {
	item, ok := excel.GetItem(uint64(cardId))
	if !ok {
		return 0, 0
	}

	strs := strings.Split(item.AddValue, ";")
	if len(strs) != 2 {
		return 0, 0
	}

	return StringToFloat32(strs[0]), StringToInt64(strs[1]) * 24 * 60 * 60
}

//获取老兵经验加成系数
func GetOlderBonusExpFactor(total, level uint32) float32 {
	var res uint64

	for _, v := range excel.GetExplaobingMap() {
		if total < uint32(v.Roundlimit) {
			continue
		}

		if level > uint32(v.Levellimit) {
			continue
		}

		if res == 0 || res < v.Value {
			res = v.Value
		}
	}

	if res == 0 {
		return 1
	}

	return float32(res) / 100.0
}

//获取经验卡加成系数
func GetCardBonusExpFactor(uid uint64) float32 {
	factor, _ := db.PlayerInfoUtil(uid).GetUsingExpCardInfo()
	if math.Abs(float64(factor)) < 0.1 {
		return 1
	}

	return factor
}

//获取战友组队经验加成
func GetComradeBonusExpFactor(comrade bool) float32 {
	if !comrade {
		return 0
	}

	data, ok := excel.GetAddition(uint64(AdditionBonus_ComradeExp))
	if !ok {
		return 0
	}

	return float32(data.Value) / 100.0
}

//游戏对局结束后获得的经验值详情
func GetExpAfterGame(uid uint64, total, level, rank, kill, matchMode, mapID uint32, matchTyp, veteranNum uint8, comrade bool) (uint32, uint32, uint32, uint32, uint32, uint32, uint32, uint32) {
	levelData, ok1 := excel.GetExplevel(uint64(level))
	rankData, ok2 := excel.GetExprank(uint64(rank))
	modeData, ok3 := excel.GetExpmode(uint64(matchMode))
	sysData, ok4 := excel.GetSystem(System_ExpRateMax)
	if !ok1 || !ok2 || !ok3 || !ok4 {
		return 0, 0, 0, 0, 0, 0, 0, 0
	}

	var modeDataValue float32
	var killData excel.ExpkillData

	if mapID == 2 {
		modeDataValue = float32(modeData.Valuenew)
	} else {
		modeDataValue = float32(modeData.Value)
	}

	if kill > 0 {
		killData, ok1 = excel.GetExpkill(uint64(kill))
		if !ok1 {
			return 0, 0, 0, 0, 0, 0, 0, 0
		}
	}

	var baseExp, totalExp, levelBonus, olderBonus, cardBonus, comradeBonus, veteranBonus, actualBonus uint32

	switch matchTyp {
	case 1:
		baseExp = uint32(float32(rankData.Single+killData.Single) * modeDataValue / 100.0)
	case 2:
		baseExp = uint32(float32(rankData.Two+killData.Two) * modeDataValue / 100.0)
	case 4:
		baseExp = uint32(float32(rankData.Four+killData.Four) * modeDataValue / 100.0)
	}

	olderFactor := GetOlderBonusExpFactor(total, level)
	cardFactor := GetCardBonusExpFactor(uid)
	comradeFactor := GetComradeBonusExpFactor(comrade)
	veteranExpRate := VeteranExpRate(veteranNum)
	expRate := float32(levelData.Value)/100.0*olderFactor*cardFactor + veteranExpRate + comradeFactor

	levelBonus = uint32(float32(baseExp) * float32(levelData.Value-100) / 100.0)
	comradeBonus = uint32(float32(baseExp) * comradeFactor)
	cardBonus = uint32(float32(baseExp) * float32(levelData.Value) / 100.0 * olderFactor * (cardFactor - 1))
	veteranBonus = uint32(float32(baseExp) * veteranExpRate)
	totalExp = uint32(float32(baseExp) * expRate)
	olderBonus = totalExp - baseExp - levelBonus - cardBonus - veteranBonus - comradeBonus

	if expRate > float32(sysData.Value)/100.0 {
		actualBonus = uint32(float32(baseExp) * float32(sysData.Value) / 100.0)
	} else {
		actualBonus = totalExp
	}

	return baseExp, totalExp, levelBonus, olderBonus, cardBonus, comradeBonus, veteranBonus, actualBonus
}

// VeteranExpRate 获取老兵回流经验加成比例
func VeteranExpRate(veteranNum uint8) float32 {
	var additionKey uint64
	if veteranNum == 0 {
		return 0
	}

	switch veteranNum {
	case 1:
		additionKey = 3
	case 2:
		additionKey = 5
	case 3:
		additionKey = 7
	case 4:
		additionKey = 9
	}

	additionData, ok := excel.GetAddition(additionKey)
	if !ok {
		return 0
	}

	return float32(additionData.Value) / 100.0
}

// VeteranCoinRate 获取老兵回流金币加成比例
func VeteranCoinRate(veteranNum uint8) float32 {
	var additionKey uint64
	if veteranNum == 0 {
		return 0
	}

	switch veteranNum {
	case 1:
		additionKey = 4
	case 2:
		additionKey = 6
	case 3:
		additionKey = 8
	case 4:
		additionKey = 10
	}

	additionData, ok := excel.GetAddition(additionKey)
	if !ok {
		return 0
	}

	return float32(additionData.Value) / 100.0
}

//计算得出新的军衔等级和经验进度
func CalMilitaryRank(level, exp, incr uint32) (uint32, uint32) {
	if level >= GetMaxMilitaryLevel() {
		return level, 0
	}

	exp += incr
	max := GetMaxExpByLevel(level + 1)

	for {
		if exp < max {
			break
		}

		level += 1

		if level < GetMaxMilitaryLevel() {
			exp -= max
			max = GetMaxExpByLevel(level + 1)
		} else {
			exp = 0
		}
	}

	return level, exp
}

/*--------------------每日任务---------------------*/

// GetDayTaskAwards 从配置表中读取每日任务相关奖励
func GetDayTaskAwards(typ uint8, id uint32) map[uint32]uint32 {
	awardsStr := ""
	awards := make(map[uint32]uint32)

	switch typ {
	case 1: //日活跃奖励
		data, ok := excel.GetActiveness(uint64(id))
		if !ok {
			return awards
		}
		awardsStr = data.Awards
	case 2: //周活跃奖励
		data, ok := excel.GetActiveness(uint64(id))
		if !ok {
			return awards
		}
		awardsStr = data.Awards
	case 3: //每日任务项奖励
		data, ok := excel.GetTask(uint64(id))
		if !ok {
			return awards
		}
		awardsStr = data.Awards
	default:
		return awards
	}

	awards = StringToMapUint32(awardsStr, "|", ";")
	return awards
}

// GetDayTaskIds 获取指定任务池对应的每日任务id
func GetDayTaskIds(taskPoolId uint32) []uint32 {
	res := []uint32{}
	taskM := excel.GetTaskMap()

	for _, task := range taskM {
		if task.Taskpool == uint64(taskPoolId) {
			res = append(res, uint32(task.Id))
		}
	}

	return res
}

// GetMaxActiveness 从配置文件中读取活跃度上限
func GetMaxActiveness(typ uint8) uint32 {
	if typ == 1 {
		return 300
	} else if typ == 2 {
		return 7 * 300
	}

	return 0
}

/*--------------------战友任务---------------------*/
// GetComradeTaskAwards 从配置表中读取战友任务相关奖励
func GetComradeTaskAwards(taskId uint32) map[uint32]uint32 {
	awards := make(map[uint32]uint32)

	data, ok := excel.GetComradeTask(uint64(taskId))
	if !ok {
		return awards
	}

	awards = StringToMapUint32(data.Awards, "|", ";")
	return awards
}

// GetComradeTaskMaxGroup 获取最大战友任务组
func GetComradeTaskMaxGroup() uint32 {
	taskM := excel.GetComradeTaskMap()
	max := uint64(0)

	for _, v := range taskM {
		if v.Groupid > max {
			max = v.Groupid
		}
	}

	return uint32(max)
}

// GetComradeTaskStartTime 获取战友任务开启时间
func GetComradeTaskStartTime() time.Time {
	taskM := excel.GetComradeTaskMap()
	var start time.Time
	var err error

	for _, v := range taskM {
		if v.Groupid == 1 {
			if v.Starttime == "" {
				continue
			}

			start, err = time.ParseInLocation("2006|01|02", v.Starttime, time.Local)
			if err != nil {
				log.Error("ParseInLocation err: ", err)
			}
			return start
		}
	}

	return start
}

// GetOpenComradeTasks 获取今日开放的战友任务
func GetOpenComradeTasks() map[uint8][]uint32 {
	tasks := make(map[uint8][]uint32)

	start := GetComradeTaskStartTime()
	if start.IsZero() {
		return tasks
	}

	days := uint32(time.Now().Sub(start).Hours() / 24)
	max := GetComradeTaskMaxGroup()
	group := 2*days%max + 1

	for _, v := range excel.GetComradeTaskMap() {
		if v.Groupid == uint64(group) {
			tasks[1] = append(tasks[1], uint32(v.Id))

			var exist bool
			for _, id := range tasks[2] {
				if id == uint32(v.Taskpool) {
					exist = true
				}
			}

			if !exist {
				tasks[2] = append(tasks[2], uint32(v.Taskpool))
			}
		}

		if v.Groupid == uint64(group+1) {
			tasks[3] = append(tasks[3], uint32(v.Id))

			var exist bool
			for _, id := range tasks[4] {
				if id == uint32(v.Taskpool) {
					exist = true
				}
			}

			if !exist {
				tasks[4] = append(tasks[4], uint32(v.Taskpool))
			}
		}
	}

	return tasks
}

/*--------------------特训任务---------------------*/
//GetOpenSpecialTasks 获取今日开放的特训任务
func GetOpenSpecialTasks() map[uint8][]uint32 {
	tasks := make(map[uint8][]uint32)
	tmpM := make(map[uint8][][]uint32)

	for _, v := range excel.GetSpecialTaskMap() {
		day := strings.Replace(v.Starttime, "|", "-", -1)
		whatDay := WhatDayOfSpecialWeek(day)

		if whatDay < 1 || whatDay > 7 {
			continue
		}

		k := 2*whatDay - 1
		exist := false

		for i, s := range tmpM[k] {
			for _, id := range s {
				t, ok := excel.GetSpecialTask(uint64(id))
				if !ok {
					continue
				}

				if t.Groupid == v.Groupid {
					tmpM[k][i] = append(tmpM[k][i], uint32(v.Id))
					exist = true
					break
				}
			}

			if exist {
				break
			}
		}

		if !exist {
			tmpM[k] = append(tmpM[k], []uint32{uint32(v.Id)})
		}

		exist = false

		for _, id := range tasks[k+1] {
			if id == uint32(v.Taskpool) {
				exist = true
			}
		}

		if !exist {
			tasks[k+1] = append(tasks[k+1], uint32(v.Taskpool))
		}
	}

	for k, ss := range tmpM {
		for _, s := range ss {
			num := len(s)
			if num < 1 {
				continue
			}

			i := rand.Intn(num)
			tasks[k] = append(tasks[k], s[i])
		}
	}

	return tasks
}

// GetSpecialLevelAwards 获取特训等级奖励
func GetSpecialLevelAwards() []uint32 {
	awards := []uint32{}

	for _, v := range excel.GetSpecialLevelMap() {
		day := strings.Replace(v.Starttime, "|", "-", -1)
		whatDay := WhatDayOfSpecialWeek(day)

		if whatDay >= 1 && whatDay <= 7 {
			awards = append(awards, uint32(v.Id))
		}
	}

	return awards
}

// getDayClearTime 获取每日清零时间
func getDayClearTime() int64 {
	timeStr := GetTBTwoSystemValue(System2_SpecialTaskClearTime)
	strs := strings.Split(timeStr, "|")
	if len(strs) != 3 {
		return 0
	}

	res := StringToInt64(strs[0]) * 60 * 60
	res += StringToInt64(strs[1]) * 60
	res += StringToInt64(strs[2])

	return res
}

// GetSpecialTaskDay 获取特训任务的任务日
func GetSpecialTaskDay() string {
	day := int64(24 * 60 * 60)
	now := time.Now().Unix()
	begin := GetTodayBeginStamp()

	if now-begin < getDayClearTime() {
		return time.Unix(now-day, 0).Format("2006-01-02")
	}

	return time.Now().Format("2006-01-02")
}

// GetSpecialTaskWeek 获取特训任务的任务周
func GetSpecialTaskWeek() string {
	day := int64(24 * 60 * 60)
	now := time.Now().Unix()
	begin := GetTodayBeginStamp()

	if WhatDayOfWeek(now) == 1 && now-begin < getDayClearTime() {
		last := GetThisWeekBeginStamp() - 7*day
		return time.Unix(last, 0).Format("2006-01-02")
	}

	this := GetThisWeekBeginStamp()
	return time.Unix(this, 0).Format("2006-01-02")
}

// GetSpecialTaskWeekDays 获取特训任务周包含的所有任务日
func GetSpecialTaskWeekDays() map[string]uint8 {
	day := int64(24 * 60 * 60)
	week := GetSpecialTaskWeek()
	res := map[string]uint8{
		week: 1,
	}

	start, err := time.ParseInLocation("2006-01-02", week, time.Local)
	if err != nil {
		log.Error("ParseInLocation err: ", err)
		return res
	}

	stamp := start.Unix()
	for i := uint8(2); i <= 7; i++ {
		res[time.Unix(stamp+day*int64(i-1), 0).Format("2006-01-02")] = i
	}

	return res
}

// WhatDayOfSpecialWeek 任务日是当前任务周的第几个任务日
func WhatDayOfSpecialWeek(day string) uint8 {
	days := GetSpecialTaskWeekDays()
	what, ok := days[day]
	if !ok {
		return 0
	}
	return what
}

// GetSpecialTaskCronExpr 获取特训任务定点清零的cron表达式
func GetSpecialTaskCronExpr() string {
	timeStr := GetTBTwoSystemValue(System2_SpecialTaskClearTime)
	strs := strings.Split(timeStr, "|")
	if len(strs) != 3 {
		return ""
	}

	h := StringToInt64(strs[0])
	m := StringToInt64(strs[1])
	s := StringToInt64(strs[2])

	return GetDayFixedTimeCronExpr(h, m, s)
}

/*--------------------以老带新---------------------*/
// GetOldBringNewAwards 获取指定id的奖励
func GetOldBringNewAwards(id uint32) map[uint32]uint32 {
	awards := make(map[uint32]uint32)

	data, ok := excel.GetBindingAwards(uint64(id))
	if !ok {
		return awards
	}

	awards = StringToMapUint32(data.Awards, "|", ";")
	return awards
}

// GetAvailableAwards 获取可以领取的奖励
func GetAvailableAwards(typ uint8, num uint32) []uint32 {
	awardM := excel.GetBindingAwardsMap()
	res := []uint32{}

	for _, v := range awardM {
		if v.Type != uint64(typ) {
			continue
		}

		if v.Require > uint64(num) {
			continue
		}

		res = append(res, uint32(v.Id))
	}

	return res
}

/*--------------------聊天系统---------------------*/

var (
	teamCustomID uint64
	chatID       uint64
)

// GetNewTeamCustomID 生成唯一队伍定制消息ID
func GetNewTeamCustomID() uint64 {
	return atomic.AddUint64(&teamCustomID, 1)
}

// GetChatUUID 生成聊天消息对应的唯一字符串
func GetChatUUID() string {
	return Uint64ToString(atomic.AddUint64(&chatID, 1))
}

/*--------------------赛季战阶---------------------*/

// GetMaxSeasonGrade 获取最大的赛季战阶等级
func GetMaxSeasonGrade() uint32 {
	return uint32(excel.GetSeasonGradeMapLen())
}

// GetMaxMedalsByGrade 获取指定等级下可以积累的最大勋章数量
func GetMaxMedalsByGrade(grade uint32) uint32 {
	data, ok := excel.GetSeasonGrade(uint64(grade))
	if !ok {
		return 0
	}

	return uint32(data.Value)
}

// CalSeasonGrade 计算得出新的赛季战阶等级和勋章进度
func CalSeasonGrade(grade, medals, incr uint32) (uint32, uint32) {
	if grade >= GetMaxSeasonGrade() {
		return grade, 0
	}

	medals += incr
	max := GetMaxMedalsByGrade(grade + 1)

	for {
		if medals < max {
			break
		}

		grade += 1

		if grade < GetMaxSeasonGrade() {
			medals -= max
			max = GetMaxMedalsByGrade(grade + 1)
		} else {
			medals = 0
		}
	}

	return grade, medals
}

// GetOpenChallengeTasks 获取开放的挑战任务
func GetOpenChallengeTasks() (map[uint8][]uint32, map[uint8]uint32) {
	tasks := make(map[uint8][]uint32)
	uniqueIds := make(map[uint8]uint32)
	season := GetRankSeason()

	for _, v := range excel.GetChallengeMap() {
		if int(v.Season) != season {
			continue
		}

		uniqueIds[uint8(v.GroupId)] = uint32(v.Id)
		k := uint8(2*v.GroupId - 1)

		for _, str := range strings.Split(v.Task, ";") {
			id := StringToUint32(str)
			item, ok := excel.GetChallengeTask(uint64(id))
			if !ok {
				continue
			}

			tasks[k] = append(tasks[k], id)
			tasks[k+1] = append(tasks[k+1], uint32(item.Taskpool))
		}
	}

	return tasks, uniqueIds
}

// GetSeasonGradeAwards 获取赛季战阶奖励
func GetSeasonGradeAwards() []uint32 {
	awards := []uint32{}

	for _, v := range excel.GetSeasonGradeMap() {
		awards = append(awards, uint32(v.Id))
	}

	return awards
}

// GetChallengeEnableItems 获取激活挑战任务组的道具
func GetChallengeEnableItems() []uint32 {
	items := []uint32{}
	season := GetRankSeason()

	for _, v := range excel.GetChallengeMap() {
		if v.Season == uint64(season) && v.Unlock == TaskUnlock_EnableItem {
			items = append(items, uint32(v.UnlockValue))
		}
	}

	return items
}

/*--------------------名字变色道具---------------------*/
// GetNameColorItem 获取名字变色道具id
func GetNameColorItem() uint32 {
	value := GetTBTwoSystemValue(System2_NameColorItem)
	strs := strings.Split(value, "|")
	return StringToUint32(strs[0])
}

// GetPlayerNameColor 获取玩家名字颜色
func GetPlayerNameColor(uid uint64) uint32 {
	if db.PlayerGoodsUtil(uid).IsGoodsEnough(GetNameColorItem(), 1) {
		return 1
	}
	return 0
}

/*--------------------礼包道具---------------------*/
// GetBoxReward1 获取礼包道具第一部分的奖励
func GetBoxReward1(id uint32) map[uint32]uint32 {
	res := make(map[uint32]uint32)

	boxConfig, ok := excel.GetBoxreward(uint64(id))
	if !ok {
		return res
	}

	if boxConfig.Reward1 == "" {
		return res
	}

	rand.Seed(time.Now().UnixNano())

	ss := strings.Split(boxConfig.Reward1, "|")
	for _, s := range ss {
		ss1 := strings.Split(s, ":")
		if len(ss1) != 3 {
			continue
		}

		id := StringToUint32(ss1[0])
		num := StringToUint32(ss1[1])
		wgt := StringToInt(ss1[2])

		tmp := rand.Intn(10000)
		if tmp < wgt {
			res[id] += num
		}
	}

	return res
}

// GetBoxReward2 获取礼包道具第二部分的奖励
func GetBoxReward2(id uint32) map[uint32]uint32 {
	res := make(map[uint32]uint32)

	boxConfig, ok := excel.GetBoxreward(uint64(id))
	if !ok {
		return res
	}

	if boxConfig.Reward2 == "" {
		return res
	}

	ss := strings.Split(boxConfig.Reward2, ";")
	if len(ss) != 2 {
		return res
	}

	ss1 := strings.Split(ss[0], "-")
	if len(ss1) != 2 {
		return res
	}

	rand.Seed(time.Now().UnixNano())

	wgt := StringToInt(ss1[0])
	num := StringToInt(ss1[1])

	tmp := rand.Intn(10000)
	if tmp >= wgt {
		return res
	}

	rewardList := []uint32{}
	weightList := []uint32{}
	weightTotal := uint32(0)
	counter := uint32(1)
	rewards := map[uint32][]uint32{}

	ss2 := strings.Split(ss[1], "|")
	for _, s := range ss2 {
		ss3 := strings.Split(s, ":")
		if len(ss3) != 3 {
			continue
		}

		id := StringToUint32(ss3[0])
		num := StringToUint32(ss3[1])
		wgt := StringToUint32(ss3[2])

		rewards[counter] = []uint32{id, num}
		weightTotal += wgt

		rewardList = append(rewardList, counter)
		weightList = append(weightList, wgt)

		counter++
	}

	if weightTotal == 0 {
		return res
	}

	var i int
	for {
		index := WeightRandom(weightTotal, weightList)
		if index == 0 {
			continue
		}

		id := rewardList[index-1]
		rs := rewards[id]
		if len(rs) != 2 {
			continue
		}

		res[rs[0]] += rs[1]
		i++

		if i >= num {
			break
		}
	}

	return res
}

/*-------------------------营销系统-----------------------------*/
// GetMarketingIDsByType 通过营销类型获取对应的活动id列表
func GetMarketingIDsByType(typ uint32) []uint64 {
	res := []uint64{}

	for _, v := range excel.GetPaysystemMap() {
		if uint32(v.MarketingType) == typ {
			res = append(res, v.Id)
		}
	}

	return res
}

// GetMarketingAwards 获取营销活动的奖励
func GetMarketingAwards(id uint64) map[uint32]uint32 {
	item, ok := excel.GetPaysystem(id)
	if !ok {
		return map[uint32]uint32{}
	}

	return StringToMapUint32(item.Awards, "|", ";")
}

// GetMonthCardPeriod 获取月卡周期
func GetMonthCardPeriod() uint32 {
	for _, v := range excel.GetPayMap() {
		if v.PayType == 4 {
			return StringToUint32(v.Payitem)
		}
	}

	return 30
}

// GetMonthCardPrice 获取月卡价格
func GetMonthCardPrice() uint32 {
	for _, v := range excel.GetPayMap() {
		if v.PayType == 4 {
			return uint32(v.Price)
		}
	}

	return 0
}

// GetMaxPayLevel 获取充值金额的对应的档位
func GetMaxPayLevel(amount uint32) uint32 {
	var max uint32

	for k, v := range excel.GetPaysystemMap() {
		if v.MarketingType != uint64(MarketingTypePayLevel) {
			continue
		}

		if amount < uint32(v.Condition) {
			continue
		}

		if max < uint32(k) {
			max = uint32(k)
		}
	}

	return max
}

/*-------------------------------------------------*/

// GetTopRating 获取全服最高rating分
func GetTopRating(season int) float32 {
	playerRankUtil := db.PlayerRankUtil(TotalRank, season)
	var score float64
	topscore, err := playerRankUtil.GetTopScore()
	if err != nil {
		log.Error(err)
		return 0
	}
	if len(topscore) != 2 {
		log.Error("获取全服最高分fail")
		return 0
	}
	score, err = redis.Float64(topscore[1], nil)
	if err != nil {
		log.Error(err)
		return 0
	}

	return float32(score)
}

// GetInsigniaIcon 获取某个玩家使用中的勋章图标
func GetInsigniaIcon(uid uint64) string {
	id, _ := db.PlayerInfoUtil(uid).GetInsigniaUsed()
	data, ok := excel.GetMedal(uint64(id))
	if !ok {
		return ""
	}
	insignia := data.Lv1_image
	num := db.PlayerInsigniaUtil(uid).GetInsigniaById(id)
	if num >= data.Lv2_amount {
		insignia = data.Lv2_image
	}
	if num >= data.Lv3_amount {
		insignia = data.Lv3_image
	}
	if num >= data.Lv4_amount {
		insignia = data.Lv4_image
	}
	if num >= data.Lv5_amount {
		insignia = data.Lv5_image
	}
	return insignia
}

// SplitReward 奖励分割函数
func SplitReward(awards string) (map[uint32]uint32, error) {
	awardsMap := make(map[uint32]uint32)
	if awards == "" {
		return awardsMap, nil
	}

	reward := strings.Split(awards, ";")
	for _, v := range reward {
		rewardMsg := strings.Split(v, "|")
		if len(rewardMsg) != 2 {
			return awardsMap, errors.New("Excel Config Err!")
		}

		rewardType, err := strconv.Atoi(rewardMsg[0])
		if err != nil {
			return awardsMap, err
		}
		rewardNum, err := strconv.Atoi(rewardMsg[1])
		if err != nil {
			return awardsMap, err
		}

		awardsMap[uint32(rewardType)] = uint32(rewardNum)
	}

	return awardsMap, nil
}

// WeightRandom 加权随机算法
func WeightRandom(weight uint32, list []uint32) uint32 {
	if weight <= 0 {
		log.Error("WeightRandom weight:", weight, " list:", list)
		return 0
	}

	rand.Seed(time.Now().UnixNano())
	index := rand.Int31n(int32(weight))

	var curWeight uint32
	for k, v := range list {
		curWeight += v
		if uint32(index) < curWeight {
			return uint32(k) + 1
		}
	}

	return 0
}

// MapToString 奖励转换为string存储 [[1656,2],[1657,3],[1658,2]]
func MapToString(input map[uint32]uint32) string {
	output := "["
	for k, v := range input {
		output += "[" + strconv.FormatUint(uint64(k), 10) + "," + strconv.FormatUint(uint64(v), 10) + "],"
	}
	output = strings.TrimSuffix(output, ",") + "]"

	return output
}

// ByHeightCalTime 通过高度计算时间(s)
func ByHeightCalTime(h float32) uint32 {
	tmp := 2.0 * h / 9.8
	if tmp < 0 {
		return 0
	}
	second := math.Sqrt(float64(tmp))

	return uint32(5 * second) //特殊处理
}

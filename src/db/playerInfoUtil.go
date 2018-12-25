package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"zeus/dbservice"

	"github.com/garyburd/redigo/redis"
)

const (
	playerInfoPrefix = "PlayerInfo"
)

//========================================
//playerInfoUtil管理玩家的各种信息
//========================================

type playerInfoUtil struct {
	uid uint64
}

//PlayerInfoUtil 生成用于管理玩家信息的工具
func PlayerInfoUtil(uid uint64) *playerInfoUtil {
	return &playerInfoUtil{
		uid: uid,
	}
}

// SetRegisterTime 设置注册时间, 设置成功返回true, 失败返回false
func (u *playerInfoUtil) SetRegisterTime(regTime int64) (bool, error) {
	c := dbservice.Get()
	defer c.Close()

	reply, err := c.Do("HSETNX", u.key(), "registertime", regTime)
	if err != nil {
		return false, err
	}
	v, err := redis.Int(reply, nil)
	if err != nil {
		return false, err
	}
	if v == 1 {
		return true, nil
	}
	return false, err
}

// GetRegisterTime 获取注册时间
func (u *playerInfoUtil) GetRegisterTime() (int64, error) {
	regTime, err := u.getValue("registertime")
	if err != nil {
		return 0, err
	}

	return redis.Int64(regTime, nil)
}

func (u *playerInfoUtil) IsKeyExist() bool {
	c := dbservice.Get()
	defer c.Close()

	b, err := redis.Bool(c.Do("EXISTS", u.key()))
	if err != nil {
		return false
	}

	return b
}

func (u *playerInfoUtil) key() string {
	return fmt.Sprintf("%s:%d", playerInfoPrefix, u.uid)
}

func (u *playerInfoUtil) getValue(field string) (interface{}, error) {
	c := dbservice.Get()
	defer c.Close()
	return c.Do("HGET", u.key(), field)
}

func (u *playerInfoUtil) setValue(field string, value interface{}) error {
	c := dbservice.Get()
	defer c.Close()
	_, err := c.Do("HSET", u.key(), field, value)
	return err
}

func (u *playerInfoUtil) SetLoginAward(str string, value int64) {
	hSet(u.key(), str, value)
}

func (u *playerInfoUtil) GetLoginAward(str string) (int64, error) {
	if !hExists(u.key(), str) {
		return 0, nil
	}

	regTime, err := u.getValue(str)
	if err != nil {
		return 0, err
	}

	return redis.Int64(regTime, nil)
}

// SetNoviceType 设置新手的类型，0:尚未勾选；1:新手；2:菜鸟；3:有经验 4:已完成新手引导
func (u *playerInfoUtil) SetNoviceType(value uint8) {
	hSet(u.key(), "novicetype", value)
}

// GetNoviceType 获取新手类型
func (u *playerInfoUtil) GetNoviceType() (uint8, error) {
	if !hExists(u.key(), "novicetype") {
		return 0, nil
	}

	ntyp, err := redis.Int64(u.getValue("novicetype"))

	return uint8(ntyp), err
}

/*-------------------------勇者模式-----------------------------*/
// 玩家参加勇者比赛的记录
type BraveRecord struct {
	UniqueId uint32 //上一场勇者比赛对应的matchmode配置文件中的唯一id
	DayStart int64  //上一场勇者比赛对应的开放时间段的开始时间戳
	MatchNum uint32 //玩家在DayStart对应的开放时间段进行的游戏的场次
}

// 设置玩家勇者模式比赛的记录
func (u *playerInfoUtil) SetBraveRecord(info *BraveRecord) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	hSet(u.key(), "braverecord", string(data))
	return nil
}

// 获取玩家上一场勇者模式比赛的记录
func (u *playerInfoUtil) GetBraveRecord() (*BraveRecord, error) {
	if !hExists(u.key(), "braverecord") {
		return nil, nil
	}

	data, err := redis.String(u.getValue("braverecord"))
	if err != nil {
		return nil, err
	}

	var info BraveRecord

	err = json.Unmarshal([]byte(data), &info)
	if err != nil {
		return nil, err
	}

	return &info, nil
}

// AddBraveWinStreak 增加勇者模式下玩家使用同一张入场券的连胜场次
func (u *playerInfoUtil) AddBraveWinStreak(value uint32) {
	hIncrBy(u.key(), "bravewin", value)
}

// ResetBraveWinStreak 重置勇者模式下玩家使用同一张入场券的连胜场次
func (u *playerInfoUtil) ResetBraveWinStreak() {
	hSet(u.key(), "bravewin", 0)
}

// GetBraveWinStreak 获取勇者模式下玩家使用同一张入场券的连胜场次
func (u *playerInfoUtil) GetBraveWinStreak() (uint32, error) {
	if !hExists(u.key(), "bravewin") {
		return 0, nil
	}

	win, err := redis.Int64(u.getValue("bravewin"))

	return uint32(win), err
}

/*-------------------------赛季排行榜-----------------------------*/
type SeasonRankFlag struct {
	Season int   //赛季id
	Enter  uint8 //赛季开始后，玩家是否查看过排行榜。0标识未查看过，1标识查看过
	Award  uint8 //赛季结束后，玩家是否领取过奖励。0标识未领取过，1标识领取过
}

//玩家查看赛季排行榜和领取赛季奖励的记录
type SeasonRankRecord struct {
	Seasons []*SeasonRankFlag
}

//赛季开始后，玩家是否查看过排行榜
func (u *playerInfoUtil) IsSeasonBeginEnter(season int) bool {
	record, err := u.getSeasonRankRecord()
	if err != nil || record == nil {
		return false
	}

	for _, info := range record.Seasons {
		if info.Season == season && info.Enter == 1 {
			return true
		}
	}

	return false
}

//赛季开始后，玩家查看了排行榜
func (u *playerInfoUtil) SetSeasonBeginEnter(season int) error {
	if u.IsSeasonBeginEnter(season) {
		return nil
	}

	record, err := u.getSeasonRankRecord()
	if err != nil {
		return err
	}

	if record == nil {
		record = &SeasonRankRecord{}
	}

	var exist bool
	for _, info := range record.Seasons {
		if info.Season == season {
			info.Enter = 1
			exist = true
			break
		}
	}

	if !exist {
		record.Seasons = append(record.Seasons, &SeasonRankFlag{
			Season: season,
			Enter:  1,
		})
	}

	u.setSeasonRankRecord(record)

	return nil
}

//赛季结束后，玩家是否领取过奖励
func (u *playerInfoUtil) IsSeasonEndAward(season int) bool {
	record, err := u.getSeasonRankRecord()
	if err != nil || record == nil {
		return false
	}

	for _, info := range record.Seasons {
		if info.Season == season && info.Award == 1 {
			return true
		}
	}

	return false
}

//赛季结束后，玩家领取了奖励
func (u *playerInfoUtil) SetSeasonEndAward(season int) error {
	if u.IsSeasonEndAward(season) {
		return nil
	}

	record, err := u.getSeasonRankRecord()
	if err != nil {
		return err
	}

	if record == nil {
		record = &SeasonRankRecord{}
	}

	var exist bool
	for _, info := range record.Seasons {
		if info.Season == season {
			info.Award = 1
			exist = true
			break
		}
	}

	if !exist {
		record.Seasons = append(record.Seasons, &SeasonRankFlag{
			Season: season,
			Enter:  1,
			Award:  1,
		})
	}

	u.setSeasonRankRecord(record)

	return nil
}

// 清除玩家的领奖记录
func (u *playerInfoUtil) ClearSeasonAwardRecord(season int) {
	if !u.IsSeasonEndAward(season) {
		return
	}

	record, err := u.getSeasonRankRecord()
	if err != nil || record == nil {
		return
	}

	for _, info := range record.Seasons {
		if info.Season == season {
			info.Award = 0
			break
		}
	}

	u.setSeasonRankRecord(record)
}

// 设置玩家赛季排行相关的记录
func (u *playerInfoUtil) setSeasonRankRecord(record *SeasonRankRecord) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	hSet(u.key(), "seasonrankrecord", string(data))
	return nil
}

// 获取玩家赛季排行相关的记录
func (u *playerInfoUtil) getSeasonRankRecord() (*SeasonRankRecord, error) {
	if !hExists(u.key(), "seasonrankrecord") {
		return nil, nil
	}

	data, err := redis.String(u.getValue("seasonrankrecord"))
	if err != nil {
		return nil, err
	}

	var record SeasonRankRecord

	err = json.Unmarshal([]byte(data), &record)
	if err != nil {
		return nil, err
	}

	return &record, nil
}

/*-------------------------军衔系统-----------------------------*/
// SetExpRatio 记录玩家在当前军衔等级下积累的经验值比例
func (u *playerInfoUtil) SetExpRatio(r float64) {
	hSet(u.key(), "expratio", r)
}

// GetExpRatio 获取玩家在当前军衔等级下积累的经验值比例
func (u *playerInfoUtil) GetExpRatio() float64 {
	if !hExists(u.key(), "expratio") {
		return 0
	}

	r, err := redis.Float64(u.getValue("expratio"))
	if err != nil {
		return 0
	}

	return r
}

//正在使用的经验卡记录
type ExpCardUseRecord struct {
	CardId    uint32  //经验卡的道具id
	Factor    float32 //经验卡倍率
	BeginTime int64   //开始使用的时间
	EndTime   int64   //失效时间
}

// AddExpCardUseRecord 添加经验卡使用记录
func (u *playerInfoUtil) AddExpCardUseRecord(cardId uint32, factor float32, validTime int64) error {
	record, err := u.getExpCardUseRecord()
	if err != nil {
		return err
	}

	if record == nil || record.EndTime <= time.Now().Unix() {
		record = &ExpCardUseRecord{
			CardId:    cardId,
			Factor:    factor,
			BeginTime: time.Now().Unix(),
			EndTime:   time.Now().Unix() + validTime,
		}
	} else {
		record.CardId = cardId
		record.EndTime += validTime
	}

	return u.setExpCardUseRecord(record)
}

// GetUsingExpCardInfo 获取正在使用的经验卡的信息
func (u *playerInfoUtil) GetUsingExpCardInfo() (float32, uint32) {
	record, err := u.getExpCardUseRecord()
	if err != nil || record == nil {
		return 0, 0
	}

	left := record.EndTime - time.Now().Unix()
	if left > 0 {
		return record.Factor, uint32(left)
	}

	return 0, 0
}

// setExpCardUseRecord 设置经验卡使用记录
func (u *playerInfoUtil) setExpCardUseRecord(record *ExpCardUseRecord) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	hSet(u.key(), "expcarduserecord", string(data))
	return nil
}

// getExpCardUseRecord 获取经验卡使用记录
func (u *playerInfoUtil) getExpCardUseRecord() (*ExpCardUseRecord, error) {
	if !hExists(u.key(), "expcarduserecord") {
		return nil, nil
	}

	data, err := redis.String(u.getValue("expcarduserecord"))
	if err != nil {
		return nil, err
	}

	var record ExpCardUseRecord

	err = json.Unmarshal([]byte(data), &record)
	if err != nil {
		return nil, err
	}

	return &record, nil
}

/*-------------------------每日任务-----------------------------*/
//任务项记录
type TaskItemRecord struct {
	TaskId      uint32 //任务项id
	FinishedNum uint32 //已完成数量
	Awards      uint8  //是否已领奖，0表示未领取，1表示已领取
	State       uint8  //奖励状态 0表示不可领取，1表示可领取，2表示已领取，3表示过期后不可领取
}

//活跃度宝箱领取记录
type ActiveAwardsBox struct {
	BoxId  uint32 //宝箱id
	Awards uint8  //是否已领奖，0表示未领取，1表示已领取
}

//每日任务相关记录
type DayTaskRecord struct {
	Day              string             //今日日期
	DayActiveness    uint32             //今日活跃度
	DayActiveAwards  []*ActiveAwardsBox //今日活跃度奖励领取记录
	DayTaskItems     []*TaskItemRecord  //各任务项记录
	DayInviteFriends []uint64           //今日邀请上线的好友
	Week             string             //本周一日期
	WeekActiveness   uint32             //本周活跃度
	WeekActiveAwards []*ActiveAwardsBox //本周活跃度奖励领取记录
}

// AddActiveness 增加活跃度
func (u *playerInfoUtil) AddActiveness(typ uint8, stamp int64, incr uint32) error {
	record, err := u.getDayTaskRecord()
	if err != nil {
		return err
	}

	if stamp == 0 {
		return errors.New("Invalid timestamp")
	}

	timeStr := time.Unix(stamp, 0).Format("2006-01-02")

	switch typ {
	case 1: //日活跃
		if record == nil {
			record = &DayTaskRecord{
				Day: timeStr,
			}
		}

		if record.Day != timeStr {
			record.Day = timeStr
			record.DayActiveness = 0
			record.DayActiveAwards = nil
			record.DayTaskItems = nil
			record.DayInviteFriends = nil
		}

		record.DayActiveness += incr
	case 2: //周活跃
		if record == nil {
			record = &DayTaskRecord{
				Week: timeStr,
			}
		}

		if record.Week != timeStr {
			record.Week = timeStr
			record.WeekActiveness = 0
			record.WeekActiveAwards = nil
		}

		record.WeekActiveness += incr
	}

	return u.setDayTaskRecord(record)
}

// GetActiveness 获取活跃度
func (u *playerInfoUtil) GetActiveness(typ uint8, stamp int64) uint32 {
	record, err := u.getDayTaskRecord()
	if record == nil || err != nil {
		return 0
	}

	if stamp == 0 {
		return 0
	}

	timeStr := time.Unix(stamp, 0).Format("2006-01-02")

	switch typ {
	case 1:
		if record.Day == timeStr {
			return record.DayActiveness
		}
	case 2:
		if record.Week == timeStr {
			return record.WeekActiveness
		}
	}

	return 0
}

// AddInviteUpLineFriend 添加邀请好友上线记录
func (u *playerInfoUtil) AddInviteUpLineFriend(today int64, uid uint64) error {
	record, err := u.getDayTaskRecord()
	if err != nil {
		return err
	}

	if today == 0 {
		return errors.New("Invalid timestamp")
	}

	todayStr := time.Unix(today, 0).Format("2006-01-02")

	if record == nil {
		record = &DayTaskRecord{
			Day: todayStr,
		}
	}

	if record.Day != todayStr {
		record.Day = todayStr
		record.DayActiveness = 0
		record.DayActiveAwards = nil
		record.DayTaskItems = nil
		record.DayInviteFriends = nil
	}

	record.DayInviteFriends = append(record.DayInviteFriends, uid)

	return u.setDayTaskRecord(record)
}

// IsFriendInvited 今日是否已经邀请过该好友
func (u *playerInfoUtil) IsFriendInvited(today int64, uid uint64) bool {
	record, err := u.getDayTaskRecord()
	if err != nil || record == nil {
		return false
	}

	if today == 0 {
		return false
	}

	todayStr := time.Unix(today, 0).Format("2006-01-02")
	if record.Day != todayStr {
		return false
	}

	for _, v := range record.DayInviteFriends {
		if v == uid {
			return true
		}
	}

	return false
}

// SetDayTaskDrawed 保存每日任务相关奖励的领取记录
func (u *playerInfoUtil) SetDayTaskDrawed(typ uint8, stamp int64, id uint32) error {
	record, err := u.getDayTaskRecord()
	if err != nil || record == nil {
		return err
	}

	if stamp == 0 {
		return errors.New("Invalid timestamp")
	}

	timeStr := time.Unix(stamp, 0).Format("2006-01-02")

	switch typ {
	case 1: //领取日活跃奖励
		if record.Day != timeStr {
			return nil
		}

		record.DayActiveAwards = append(record.DayActiveAwards, &ActiveAwardsBox{
			BoxId:  id,
			Awards: 1,
		})
	case 2: //领取周活跃奖励
		if record.Week != timeStr {
			return nil
		}

		record.WeekActiveAwards = append(record.WeekActiveAwards, &ActiveAwardsBox{
			BoxId:  id,
			Awards: 1,
		})
	case 3: //领取任务奖励
		if record.Day != timeStr {
			return nil
		}

		for _, v := range record.DayTaskItems {
			if v.TaskId == id {
				v.Awards = 1
			}
		}
	}

	return u.setDayTaskRecord(record)
}

// IsDayTaskDrawed 奖励是否已经领取
func (u *playerInfoUtil) IsDayTaskDrawed(typ uint8, stamp int64, id uint32) bool {
	record, err := u.getDayTaskRecord()
	if err != nil || record == nil {
		return false
	}

	if stamp == 0 {
		return false
	}

	timeStr := time.Unix(stamp, 0).Format("2006-01-02")

	switch typ {
	case 1:
		if record.Day != timeStr {
			return false
		}

		for _, v := range record.DayActiveAwards {
			if v.BoxId == id {
				return v.Awards == 1
			}
		}
	case 2:
		if record.Week != timeStr {
			return false
		}

		for _, v := range record.WeekActiveAwards {
			if v.BoxId == id {
				return v.Awards == 1
			}
		}
	case 3:
		if record.Day != timeStr {
			return false
		}

		for _, v := range record.DayTaskItems {
			if v.TaskId == id {
				return v.Awards == 1
			}
		}
	}

	return false
}

// SetDayTaskItemProgress 设置每日任务项的完成进度
func (u *playerInfoUtil) SetDayTaskItemProgress(today int64, taskId uint32, num uint32) error {
	record, err := u.getDayTaskRecord()
	if err != nil {
		return err
	}

	if today == 0 {
		return errors.New("Invalid timestamp")
	}

	todayStr := time.Unix(today, 0).Format("2006-01-02")

	if record == nil {
		record = &DayTaskRecord{
			Day: todayStr,
		}
	}

	if record.Day != todayStr {
		record.Day = todayStr
		record.DayActiveness = 0
		record.DayActiveAwards = nil
		record.DayTaskItems = nil
		record.DayInviteFriends = nil
	}

	var exist bool
	for _, v := range record.DayTaskItems {
		if v.TaskId == taskId {
			v.FinishedNum = num
			exist = true
		}
	}

	if !exist {
		record.DayTaskItems = append(record.DayTaskItems, &TaskItemRecord{
			TaskId:      taskId,
			FinishedNum: num,
		})
	}

	return u.setDayTaskRecord(record)
}

// GetDayTaskItemProgress 获取每日任务项的完成进度
func (u *playerInfoUtil) GetDayTaskItemProgress(today int64, taskId uint32) uint32 {
	record, err := u.getDayTaskRecord()
	if err != nil || record == nil {
		return 0
	}

	if today == 0 {
		return 0
	}

	todayStr := time.Unix(today, 0).Format("2006-01-02")
	if record.Day != todayStr {
		return 0
	}

	for _, v := range record.DayTaskItems {
		if v.TaskId == taskId {
			return v.FinishedNum
		}
	}

	return 0
}

// GetDayTaskRecord 获取每日任务详细数据记录
func (u *playerInfoUtil) GetDayTaskRecord(today, week int64) *DayTaskRecord {
	record, err := u.getDayTaskRecord()
	if err != nil || record == nil {
		return nil
	}

	if today == 0 || week == 0 {
		return nil
	}

	todayStr := time.Unix(today, 0).Format("2006-01-02")
	weekStr := time.Unix(week, 0).Format("2006-01-02")

	if record.Day != todayStr {
		record.Day = todayStr
		record.DayActiveness = 0
		record.DayActiveAwards = nil
		record.DayTaskItems = nil
		record.DayInviteFriends = nil
	}

	if record.Week != weekStr {
		record.Week = weekStr
		record.WeekActiveness = 0
		record.WeekActiveAwards = nil
	}

	return record
}

func (u *playerInfoUtil) setDayTaskRecord(record *DayTaskRecord) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	hSet(u.key(), "daytaskrecord", string(data))
	return nil
}

func (u *playerInfoUtil) getDayTaskRecord() (*DayTaskRecord, error) {
	if !hExists(u.key(), "daytaskrecord") {
		return nil, nil
	}

	data, err := redis.String(u.getValue("daytaskrecord"))
	if err != nil {
		return nil, err
	}

	var record DayTaskRecord

	err = json.Unmarshal([]byte(data), &record)
	if err != nil {
		return nil, err
	}

	return &record, nil
}

/*-------------------------战友任务-----------------------------*/

//战友任务相关记录
type ComradeTaskRecord struct {
	Day              string            //今日日期
	TaskRound        uint8             //1表示第一轮任务，2表示第二轮任务
	ComradeTaskItems []*TaskItemRecord //各任务项记录
}

// GetComradeTaskRound 获取玩家的当前战友任务轮次
func (u *playerInfoUtil) GetComradeTaskRound(today int64) uint8 {
	record, err := u.getComradeTaskRecord()
	if err != nil || record == nil || today == 0 {
		return 1
	}

	todayStr := time.Unix(today, 0).Format("2006-01-02")
	if record.Day != todayStr {
		return 1
	}

	return record.TaskRound
}

// IncreaseComradeTaskRound 增加战友任务轮次
func (u *playerInfoUtil) IncreaseComradeTaskRound() error {
	record, err := u.getComradeTaskRecord()
	if err != nil || record == nil {
		return err
	}

	if record.TaskRound <= 1 {
		record.TaskRound++
	}

	return u.setComradeTaskRecord(record)
}

// SetComradeTaskDrawed 保存战友任务相关奖励的领取记录
func (u *playerInfoUtil) SetComradeTaskDrawed(today int64, id uint32) error {
	record, err := u.getComradeTaskRecord()
	if err != nil || record == nil {
		return err
	}

	if today == 0 {
		return errors.New("Invalid timestamp")
	}

	todayStr := time.Unix(today, 0).Format("2006-01-02")
	if record.Day != todayStr {
		return nil
	}

	for _, v := range record.ComradeTaskItems {
		if v.TaskId == id {
			v.Awards = 1
		}
	}

	return u.setComradeTaskRecord(record)
}

// IsComradeTaskDrawed 战友任务奖励是否已经领取
func (u *playerInfoUtil) IsComradeTaskDrawed(today int64, taskId uint32) bool {
	record, err := u.getComradeTaskRecord()
	if err != nil || record == nil {
		return false
	}

	if today == 0 {
		return false
	}

	todayStr := time.Unix(today, 0).Format("2006-01-02")
	if record.Day != todayStr {
		return false
	}

	for _, v := range record.ComradeTaskItems {
		if v.TaskId == taskId {
			return v.Awards == 1
		}
	}

	return false
}

// SetComradeTaskItemProgress 设置战友任务项的完成进度
func (u *playerInfoUtil) SetComradeTaskItemProgress(today int64, taskId uint32, num uint32) error {
	record, err := u.getComradeTaskRecord()
	if err != nil {
		return err
	}

	if today == 0 {
		return errors.New("Invalid timestamp")
	}

	todayStr := time.Unix(today, 0).Format("2006-01-02")

	if record == nil {
		record = &ComradeTaskRecord{
			Day:       todayStr,
			TaskRound: 1,
		}
	}

	if record.Day != todayStr {
		record.Day = todayStr
		record.TaskRound = 1
		record.ComradeTaskItems = nil
	}

	var exist bool
	for _, v := range record.ComradeTaskItems {
		if v.TaskId == taskId {
			v.FinishedNum = num
			exist = true
		}
	}

	if !exist {
		record.ComradeTaskItems = append(record.ComradeTaskItems, &TaskItemRecord{
			TaskId:      taskId,
			FinishedNum: num,
		})
	}

	return u.setComradeTaskRecord(record)
}

// GetComradeTaskItemProgress 获取战友任务项的完成进度
func (u *playerInfoUtil) GetComradeTaskItemProgress(today int64, taskId uint32) uint32 {
	record, err := u.getComradeTaskRecord()
	if err != nil || record == nil {
		return 0
	}

	if today == 0 {
		return 0
	}

	todayStr := time.Unix(today, 0).Format("2006-01-02")
	if record.Day != todayStr {
		return 0
	}

	for _, v := range record.ComradeTaskItems {
		if v.TaskId == taskId {
			return v.FinishedNum
		}
	}

	return 0
}

// GetComradeTaskRecord 获取战友任务详细数据记录
func (u *playerInfoUtil) GetComradeTaskRecord(today int64) *ComradeTaskRecord {
	record, err := u.getComradeTaskRecord()
	if err != nil || record == nil {
		return nil
	}

	if today == 0 {
		return nil
	}

	todayStr := time.Unix(today, 0).Format("2006-01-02")

	if record.Day != todayStr {
		record.Day = todayStr
		record.TaskRound = 1
		record.ComradeTaskItems = nil
	}

	return record
}

func (u *playerInfoUtil) setComradeTaskRecord(record *ComradeTaskRecord) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	hSet(u.key(), "ComradeTaskRecord", string(data))
	return nil
}

func (u *playerInfoUtil) getComradeTaskRecord() (*ComradeTaskRecord, error) {
	if !hExists(u.key(), "ComradeTaskRecord") {
		return nil, nil
	}

	data, err := redis.String(u.getValue("ComradeTaskRecord"))
	if err != nil {
		return nil, err
	}

	var record ComradeTaskRecord

	err = json.Unmarshal([]byte(data), &record)
	if err != nil {
		return nil, err
	}

	return &record, nil
}

/*-------------------------以老带新-----------------------------*/
//奖励信息
type AwardInfo struct {
	Id    uint32 //奖励项id
	State uint8  //当前状态：0表示不可领取，1表示可领取，2表示已领取
}

//以老带新相关记录
type OldBringNewRecord struct {
	Teachers            []uint64     //师父
	TakeTeacherDeadTime int64        //拜师截止时间
	TakeTeacherAwards   []*AwardInfo //拜师奖励
	Pupils              []uint64     //徒弟
	ReceivePupilAwards  []*AwardInfo //收徒弟奖励
}

// SetTakeTeacherDeadTime 设置拜师截止时间
func (u *playerInfoUtil) SetTakeTeacherDeadTime(deadTime int64) error {
	record := &OldBringNewRecord{
		TakeTeacherDeadTime: deadTime,
	}

	return u.setOldBringNewRecord(record)
}

// AddTeacher 添加师父
func (u *playerInfoUtil) AddTeacher(uid uint64) error {
	record, err := u.getOldBringNewRecord()
	if err != nil {
		return err
	}

	if record == nil {
		record = &OldBringNewRecord{}
	}

	record.Teachers = append(record.Teachers, uid)

	return u.setOldBringNewRecord(record)
}

// RemoveTeacher 删除师父
func (u *playerInfoUtil) RemoveTeacher(uid uint64) error {
	record, err := u.getOldBringNewRecord()
	if err != nil || record == nil {
		return err
	}

	for i, teacher := range record.Teachers {
		if teacher == uid {
			record.Teachers = append(record.Teachers[:i], record.Teachers[i+1:]...)
		}
	}

	return u.setOldBringNewRecord(record)
}

// GetTeachers 获取师父数量
func (u *playerInfoUtil) GetTeachers() uint32 {
	record, err := u.getOldBringNewRecord()
	if err != nil || record == nil {
		return 0
	}

	return uint32(len(record.Teachers))
}

// IsTeacher 是否师父
func (u *playerInfoUtil) IsTeacher(uid uint64) bool {
	record, err := u.GetOldBringNewRecord()
	if err != nil || record == nil {
		return false
	}

	for _, teacher := range record.Teachers {
		if teacher == uid {
			return true
		}
	}

	return false
}

// AddPupil 添加徒弟
func (u *playerInfoUtil) AddPupil(uid uint64) error {
	record, err := u.getOldBringNewRecord()
	if err != nil {
		return err
	}

	if record == nil {
		record = &OldBringNewRecord{}
	}

	record.Pupils = append(record.Pupils, uid)

	return u.setOldBringNewRecord(record)
}

// RemovePupil 删除徒弟
func (u *playerInfoUtil) RemovePupil(uid uint64) error {
	record, err := u.getOldBringNewRecord()
	if err != nil || record == nil {
		return err
	}

	for i, pupil := range record.Pupils {
		if pupil == uid {
			record.Pupils = append(record.Pupils[:i], record.Pupils[i+1:]...)
		}
	}

	return u.setOldBringNewRecord(record)
}

// GetPupils 获取徒弟数量
func (u *playerInfoUtil) GetPupils() uint32 {
	record, err := u.getOldBringNewRecord()
	if err != nil || record == nil {
		return 0
	}

	return uint32(len(record.Pupils))
}

// IsPupil 是否徒弟
func (u *playerInfoUtil) IsPupil(uid uint64) bool {
	record, err := u.GetOldBringNewRecord()
	if err != nil || record == nil {
		return false
	}

	for _, pupil := range record.Pupils {
		if pupil == uid {
			return true
		}
	}

	return false
}

// IsTeacherPupil 是否师徒
func (u *playerInfoUtil) IsTeacherPupil(uid uint64) bool {
	record, err := u.GetOldBringNewRecord()
	if err != nil || record == nil {
		return false
	}

	ids := append(record.Teachers, record.Pupils...)
	for _, id := range ids {
		if id == uid {
			return true
		}
	}

	return false
}

// HaveTeacherPupil 是否有师父或徒弟
func (u *playerInfoUtil) HaveTeacherPupil() bool {
	record, err := u.GetOldBringNewRecord()
	if err != nil || record == nil {
		return false
	}

	if len(record.Teachers) > 0 || len(record.Pupils) > 0 {
		return true
	}

	return false
}

// SetOldBringNewDrawable 设置以老带新相关奖励可以领取
func (u *playerInfoUtil) SetOldBringNewDrawable(typ uint8, ids []uint32) error {
	record, err := u.getOldBringNewRecord()
	if err != nil || record == nil {
		return err
	}

	switch typ {
	case 1:
		for _, id := range ids {
			var exist bool

			for _, info := range record.TakeTeacherAwards {
				if info.Id == id {
					exist = true
					if info.State == 0 {
						info.State = 1
					}
				}
			}

			if !exist {
				record.TakeTeacherAwards = append(record.TakeTeacherAwards, &AwardInfo{
					Id:    id,
					State: 1,
				})
			}
		}
	case 2:
		for _, id := range ids {
			var exist bool

			for _, info := range record.ReceivePupilAwards {
				if info.Id == id {
					exist = true
					if info.State == 0 {
						info.State = 1
					}
				}
			}

			if !exist {
				record.ReceivePupilAwards = append(record.ReceivePupilAwards, &AwardInfo{
					Id:    id,
					State: 1,
				})
			}
		}
	}

	return u.setOldBringNewRecord(record)
}

// CanDrawOldBringNew 是否可以领取以老带新相关奖励
func (u *playerInfoUtil) CanDrawOldBringNew(typ uint8, id uint32) bool {
	record, err := u.getOldBringNewRecord()
	if err != nil || record == nil {
		return false
	}

	switch typ {
	case 1:
		for _, info := range record.TakeTeacherAwards {
			if info.Id == id {
				return info.State == 1
			}
		}
	case 2:
		for _, info := range record.ReceivePupilAwards {
			if info.Id == id {
				return info.State == 1
			}
		}
	}

	return false
}

// SetOldBringNewDrawed 记录领取以老带新相关奖励
func (u *playerInfoUtil) SetOldBringNewDrawed(typ uint8, id uint32) error {
	record, err := u.getOldBringNewRecord()
	if err != nil || record == nil {
		return err
	}

	switch typ {
	case 1:
		for _, info := range record.TakeTeacherAwards {
			if info.Id == id {
				info.State = 2
			}
		}
	case 2:
		for _, info := range record.ReceivePupilAwards {
			if info.Id == id {
				info.State = 2
			}
		}
	}

	return u.setOldBringNewRecord(record)
}

// GetOldBringNewRecord 获取以老带新相关记录
func (u *playerInfoUtil) GetOldBringNewRecord() (*OldBringNewRecord, error) {
	return u.getOldBringNewRecord()
}

func (u *playerInfoUtil) setOldBringNewRecord(record *OldBringNewRecord) error {
	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	hSet(u.key(), "OldBringNewRecord", string(data))
	return nil
}

func (u *playerInfoUtil) getOldBringNewRecord() (*OldBringNewRecord, error) {
	if !hExists(u.key(), "OldBringNewRecord") {
		return nil, nil
	}

	data, err := redis.String(u.getValue("OldBringNewRecord"))
	if err != nil {
		return nil, err
	}

	var record OldBringNewRecord

	err = json.Unmarshal([]byte(data), &record)
	if err != nil {
		return nil, err
	}

	return &record, nil
}

//==============修改名字========================

// SetChangeNameNum 设置改名次数
func (u *playerInfoUtil) SetChangeNameNum(num uint64) {
	hSet(u.key(), "ChangeNameNum", num)
	u.SetChangeNameTime()
}

// GetChangeNameNum 获取改名次数
func (u *playerInfoUtil) GetChangeNameNum() (uint64, error) {
	if !hExists(u.key(), "ChangeNameNum") {
		return 0, nil
	}

	num, err := u.getValue("ChangeNameNum")
	if err != nil {
		return 0, err
	}

	return redis.Uint64(num, nil)
}

// SetChangeNameTime 设置改名时间
func (u *playerInfoUtil) SetChangeNameTime() {
	t, e := time.ParseInLocation("2006-01-02", time.Now().Format("2006-01-02"), time.Local)
	if e != nil {
		t = time.Now()
	}
	hSet(u.key(), "ChangeNameTime", t.Unix())
}

// GetChangeNameTime 获取改名时间
func (u *playerInfoUtil) GetChangeNameTime() (uint64, error) {
	if !hExists(u.key(), "ChangeNameTime") {
		return 0, nil
	}

	num, err := u.getValue("ChangeNameTime")
	if err != nil {
		return 0, err
	}

	return redis.Uint64(num, nil)
}

// GetInsigniaUsed 当前展示中的勋章
func (u *playerInfoUtil) GetInsigniaUsed() (uint32, error) {
	id, e := redis.Uint64(u.getValue("insignia"))
	if e != nil {
		return 0, e
	}
	return uint32(id), nil
}

// SetInsigniaUsed 设置展示的勋章
func (u *playerInfoUtil) SetInsigniaUsed(id uint64) {
	u.setValue("insignia", id)
}

/*=========================设置通过老兵回流召回该玩家的列表==================*/

//每日任务项记录
type VeteranRecallList struct {
	List map[uint64]uint32
}

func (u *playerInfoUtil) SetVeteranRecallList(list *VeteranRecallList) error {
	data, err := json.Marshal(list)
	if err != nil {
		return err
	}

	hSet(u.key(), "VeteranRecallList", string(data))
	return nil
}

func (u *playerInfoUtil) GetVeteranRecallList(list *VeteranRecallList) error {
	if !hExists(u.key(), "VeteranRecallList") {
		return nil
	}

	data, err := redis.String(u.getValue("VeteranRecallList"))
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(data), list)
	if err != nil {
		return err
	}

	return nil
}

/*=========================获得稀有道具个数==================*/

// GetRateItemNum 获得稀有道具个数
func (u *playerInfoUtil) GetRateItemNum() (uint32, error) {
	if !hExists(u.key(), "RateItemNum") {
		return 0, nil
	}

	num, e := redis.Uint64(u.getValue("RateItemNum"))
	if e != nil {
		return 0, e
	}
	return uint32(num), nil
}

// SetRateItemNum 设置稀有道具个数
func (u *playerInfoUtil) SetRateItemNum(num uint32) {
	u.setValue("RateItemNum", num)
}

/*=========================设置角色外化偏好信息=======================*/

// 偏好信息设置
type PreferenceInfo struct {
	Start        map[uint32]map[uint32]bool     // 索引含义 1:角色开关，2：伞包开关，5：背包开关，6:头盔开关
	PreType      map[uint32]map[uint32]uint32   // 索引含义(1:角色，2：伞包，5：背包，6:头盔) 值含义(0全体， 1偏好)
	MayaItemList map[uint32]map[uint32][]uint32 // (角色、伞包、背包、头盔)偏好道具信息设置列表
}

func (u *playerInfoUtil) SetPreferenceList(info *PreferenceInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	hSet(u.key(), "PreferenceInfo", string(data))
	return nil
}

func (u *playerInfoUtil) GetPreferenceList(info *PreferenceInfo) error {
	if !hExists(u.key(), "PreferenceInfo") {
		return nil
	}

	data, err := redis.String(u.getValue("PreferenceInfo"))
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(data), info)
	if err != nil {
		return err
	}

	return nil
}

/*=========================获得举报类型记录次数==================*/

// 举报次数信息设置
type CheaterReportInfo struct {
	ReportNum map[uint32]uint32
}

// SetCheaterReportNum 设置举报次数
func (u *playerInfoUtil) SetCheaterReportNum(info *CheaterReportInfo) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	hSet(u.key(), "CheaterReport", string(data))
	return nil
}

// GetCheaterReportNum 获得举报次数
func (u *playerInfoUtil) GetCheaterReportNum(info *CheaterReportInfo) error {
	if !hExists(u.key(), "CheaterReport") {
		return nil
	}

	data, err := redis.String(u.getValue("CheaterReport"))
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(data), info)
	if err != nil {
		return err
	}

	return nil
}

/*=========================红点提示记录==================*/

// 一次行点击类红点
type RedDotOnce struct {
	RedDot map[uint32]uint32
}

// SetRedDotOnce 设置一次行点击类红点
func (u *playerInfoUtil) SetRedDotOnce(info *RedDotOnce) error {
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	hSet(u.key(), "RedDotOnce", string(data))
	return nil
}

// GetRedDotOnce 获得一次行点击类红点
func (u *playerInfoUtil) GetRedDotOnce(info *RedDotOnce) error {
	if !hExists(u.key(), "RedDotOnce") {
		return nil
	}

	data, err := redis.String(u.getValue("RedDotOnce"))
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(data), info)
	if err != nil {
		return err
	}

	return nil
}

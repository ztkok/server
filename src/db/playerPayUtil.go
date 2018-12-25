package db

import (
	"encoding/json"
	"fmt"
	"math"
	"time"
	"zeus/dbservice"

	log "github.com/cihub/seelog"
	"github.com/garyburd/redigo/redis"
)

const (
	payRecordPrefix = "PayRecord"
)

type playerPayUtil struct {
	uid       uint64
	marketing string
}

// PlayerPayUtil 生成用于管理玩家充值的工具
func PlayerPayUtil(uid uint64, marketing string) *playerPayUtil {
	return &playerPayUtil{
		uid:       uid,
		marketing: marketing,
	}
}

func (u *playerPayUtil) getValue(field string) (interface{}, error) {
	c := dbservice.Get()
	defer c.Close()
	return c.Do("HGET", u.key(), field)
}

func (u *playerPayUtil) getUintValue(field string) uint64 {
	if !hExists(u.key(), field) {
		return 0
	}

	data, err := redis.Uint64(u.getValue(field))
	if err != nil {
		log.Error("redis Uint64 err: ", err)
		return 0
	}

	return data
}

func (u *playerPayUtil) getStringValue(field string) string {
	if !hExists(u.key(), field) {
		return ""
	}

	data, err := redis.String(u.getValue(field))
	if err != nil {
		log.Error("redis String err: ", err)
		return ""
	}

	return data
}

func (u *playerPayUtil) getBoolValue(field string) bool {
	if !hExists(u.key(), field) {
		return false
	}

	data, err := redis.Bool(u.getValue(field))
	if err != nil {
		log.Error("redis Bool err: ", err)
		return false
	}

	return data
}

func (u *playerPayUtil) setValue(field string, value interface{}) error {
	c := dbservice.Get()
	defer c.Close()
	_, err := c.Do("HSET", u.key(), field, value)
	return err
}

func (u *playerPayUtil) setComplexValue(field string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return u.setValue(field, string(data))
}

func (u *playerPayUtil) hdel(field string) {
	hDEL(u.key(), field)
}

func (u *playerPayUtil) del() {
	delKey(u.key())
}

func (u *playerPayUtil) key() string {
	return fmt.Sprintf("%s:%s:%d", payRecordPrefix, u.marketing, u.uid)
}

/*-------------------------首充系统-----------------------------*/
// GetFirstPayRecord 获取玩家首次充值记录
func (u *playerPayUtil) GetFirstPayRecord() uint8 {
	return uint8(u.getUintValue("State"))
}

// SetFirstPayRecord 记录玩家首次充值
// 0表示未完成首充，1表示已完成首充，2表示已领取首充奖励
func (u *playerPayUtil) SetFirstPayRecord(state uint8) error {
	return u.setValue("State", state)
}

/*-------------------------月卡系统-----------------------------*/
// 日结充值币领取记录
type DayAwardsRecord struct {
	Day    string //日期
	Awards uint8  //是否领取日结充值币
}

// 即将过期通知记录
type SoonExpireNotify struct {
	Day    string //日期
	Notify uint8  //0表示未通知 1表示已通知
}

// SetMonthCardRecord 添加玩家购买月卡的记录
func (u *playerPayUtil) SetMonthCardRecord(beginTime, endTime int64) error {
	today := time.Now().Format("2006-01-02")
	expire := u.GetMonthCardEndTime()
	expireDay := time.Unix(expire, 0).Format("2006-01-02")

	if expire <= time.Now().Unix() && expireDay != today {
		u.del()
	}

	if err := u.setValue("BeginTime", beginTime); err != nil {
		return err
	}

	if err := u.setValue("EndTime", endTime); err != nil {
		return err
	}

	return nil
}

// SetMonthCardNum 设置月卡数量
func (u *playerPayUtil) SetMonthCardNum(num uint32) error {
	return u.setValue("CardNum", num)
}

// GetMonthCardEndTime 获取玩家月卡的过期时间
func (u *playerPayUtil) GetMonthCardEndTime() int64 {
	return int64(u.getUintValue("EndTime"))
}

// GetMonthCardLeftDays 获取玩家月卡的剩余天数
func (u *playerPayUtil) GetMonthCardLeftDays() uint32 {
	expire := u.GetMonthCardEndTime()
	left := expire - time.Now().Unix() - 60

	if left > 0 {
		return uint32(math.Ceil(float64(left) / float64(24*60*60)))
	}

	return 0
}

// IncrOnceAwardDrawNum 增加即得奖励领取次数
func (u *playerPayUtil) IncrOnceAwardDrawNum() {
	hIncrBy(u.key(), "DrawNum", 1)
}

// GetOnceAwardUndrawedNum 获取可以领取的即得奖励次数
func (u *playerPayUtil) GetOnceAwardUndrawedNum() uint32 {
	if !u.IsOwnValidMonthCard() {
		return 0
	}

	cardNum := uint32(u.getUintValue("CardNum"))
	drawNum := uint32(u.getUintValue("DrawNum"))

	if cardNum > drawNum {
		return cardNum - drawNum
	}

	return 0
}

// IsOwnValidMonthCard 玩家是否拥有一张有效的月卡
func (u *playerPayUtil) IsOwnValidMonthCard() bool {
	return u.GetMonthCardEndTime()-60 > time.Now().Unix()
}

// IsDayAwardsDrawed 玩家是否已经领取当天奖励
func (u *playerPayUtil) IsDayAwardsDrawed() bool {
	today := time.Now().Format("2006-01-02")
	awards, _ := u.getDayAwardsRecord()

	for _, v := range awards {
		if v.Day == today {
			return v.Awards == 1
		}
	}

	return false
}

// SetDayAwardsDrawed 记录玩家领取当天奖励
func (u *playerPayUtil) SetDayAwardsDrawed() error {
	awards, err := u.getDayAwardsRecord()
	if err != nil {
		return err
	}

	today := time.Now().Format("2006-01-02")
	awards = append(awards, &DayAwardsRecord{
		Day:    today,
		Awards: 1,
	})

	return u.setComplexValue("DayAwards", awards)
}

func (u *playerPayUtil) getDayAwardsRecord() ([]*DayAwardsRecord, error) {
	if !hExists(u.key(), "DayAwards") {
		return nil, nil
	}

	data, err := redis.String(u.getValue("DayAwards"))
	if err != nil {
		return nil, err
	}

	var record []*DayAwardsRecord
	err = json.Unmarshal([]byte(data), &record)
	if err != nil {
		return nil, err
	}

	return record, nil
}

// NeedNotifyExpire 是否需要通知玩家过期信息
// typ为1表示即将过期通知，typ为2表示已经过期通知
func (u *playerPayUtil) NeedNotifyExpire(typ uint8) bool {
	endTime := u.GetMonthCardEndTime()
	if endTime == 0 {
		return false
	}

	leftDays := u.GetMonthCardLeftDays()

	switch typ {
	case 1:
		if leftDays >= 1 && leftDays <= 3 {
			today := time.Now().Format("2006-01-02")
			record, _ := u.getSoonExpireNotifyRecord()

			for _, v := range record {
				if v.Day == today {
					return v.Notify != 1
				}
			}

			return true
		}
	case 2:
		if leftDays <= 0 {
			return u.getUintValue("ExpiredNotify") != 1
		}
	}

	return false
}

// SetExpireNotifyed 添加通知玩家过期的记录
func (u *playerPayUtil) SetExpireNotifyed(typ uint8) error {
	switch typ {
	case 1:
		{
			record, err := u.getSoonExpireNotifyRecord()
			if err != nil {
				return err
			}

			today := time.Now().Format("2006-01-02")
			record = append(record, &SoonExpireNotify{
				Day:    today,
				Notify: 1,
			})

			return u.setComplexValue("SoonExpireNotify", record)
		}
	case 2:
		return u.setValue("ExpiredNotify", 1)
	}

	return nil
}

func (u *playerPayUtil) getSoonExpireNotifyRecord() ([]*SoonExpireNotify, error) {
	if !hExists(u.key(), "SoonExpireNotify") {
		return nil, nil
	}

	data, err := redis.String(u.getValue("SoonExpireNotify"))
	if err != nil {
		return nil, err
	}

	var record []*SoonExpireNotify
	err = json.Unmarshal([]byte(data), &record)
	if err != nil {
		return nil, err
	}

	return record, nil
}

/*-------------------------首次充送-----------------------------*/
// GetPayLevelResetRound 获取首次按档位充送活动的重置轮次
func (u *playerPayUtil) GetPayLevelResetRound() uint32 {
	return uint32(u.getUintValue("Resets"))
}

// SetPayLevelResetRound 记录首次按档位充送活动的重置轮次
func (u *playerPayUtil) SetPayLevelResetRound(round uint32) error {
	if round != u.GetPayLevelResetRound() {
		u.hdel("Levels")
		return u.setValue("Resets", round)
	}
	return nil
}

// AddPayLevelRecord 添加首次按档位充送活动的记录
func (u *playerPayUtil) AddPayLevelRecord(id uint32) {
	list := u.GetPayLevelRecord()
	list = append(list, id)
	u.setComplexValue("Levels", list)
}

// GetPayLevelRecord 获取首次按档位充送活动的记录
func (u *playerPayUtil) GetPayLevelRecord() []uint32 {
	if !hExists(u.key(), "Levels") {
		return nil
	}

	data, err := redis.String(u.getValue("Levels"))
	if err != nil {
		return nil
	}

	var list []uint32
	if err := json.Unmarshal([]byte(data), &list); err != nil {
		return nil
	}

	return list
}

// IsPayLevelMatched 玩家是否已经达到过指定档次
func (u *playerPayUtil) IsPayLevelMatched(level uint32) bool {
	list := u.GetPayLevelRecord()
	for _, v := range list {
		if v == level {
			return true
		}
	}
	return false
}

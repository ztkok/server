package db

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	log "github.com/cihub/seelog"
)

const (
	activityPrefix = "ActivityInfoTable"
	activityRedDot = "ActivityRedHot"
)

// ActivityUtil
type ActivityUtil struct {
	uid uint64
	id  uint32 //活动id
}

func PlayerActivityUtil(uid uint64, id uint32) *ActivityUtil {
	return &ActivityUtil{
		uid: uid,
		id:  id,
	}
}

func (a *ActivityUtil) key() string {
	return fmt.Sprintf("%s:%d", activityPrefix, a.id)
}

// ClearOwnActivity 清空已领取过信息
func (a *ActivityUtil) ClearOwnActivity() {
	hDEL(a.key(), a.uid)
}

// IsActivity 是否已领取过
func (a *ActivityUtil) IsActivity() bool {
	return hExists(a.key(), a.uid)
}

/*=========================获得活动首次点击红点提示==================*/

func (a *ActivityUtil) KeyRedDot() string {
	return fmt.Sprintf("%s:%d:%d", activityRedDot, a.id, a.uid)
}

// GetRedDot 获得每次版本更新首次点击红点提示
func (a *ActivityUtil) GetRedDot(version string) uint32 {
	if !hExists(a.KeyRedDot(), version) {
		return 0
	}
	return 1
}

// SetRedDot 设置每次版本更新首次点击红点提示
func (a *ActivityUtil) SetRedDot(version string) {
	if hExists(a.KeyRedDot(), version) {
		return
	}

	hSet(a.KeyRedDot(), version, 1)
}

/*-----------------------------------------签到活动--------------------------------------*/

// SigActivityInfo 签到活动信息
type SigActivityInfo struct {
	Id         uint32 //活动id
	Time       int64  //上次领取时间
	GoodId     uint32 //上次领取到的物品
	Sum        uint32 //领取次数
	ActStartTm int64  //活动开始时间
}

func (sigActivityInfo *SigActivityInfo) String() string {
	return fmt.Sprintf("%+v\n", *sigActivityInfo)
}

// SetActivityInfo 设置活动信息
func (a *ActivityUtil) SetActivityInfo(info *SigActivityInfo) bool {
	if info == nil {
		return false
	}

	d, err := json.Marshal(info)
	if err != nil {
		log.Warn("SetActivityInfo error = ", err)
		return false
	}

	hSet(a.key(), a.uid, string(d))
	return true
}

// GetActivityInfo 获取活动信息
func (a *ActivityUtil) GetActivityInfo(info *SigActivityInfo) error {
	v := hGet(a.key(), a.uid)
	if err := json.Unmarshal([]byte(v), &info); err != nil {
		log.Warn("GetActivityInfo Failed to Unmarshal ", err)
		return errors.New("unmarshal error")
	}

	return nil
}

/*-----------------------------------------更新活动--------------------------------------*/

// UpdateActivityInfo 更新活动信息
type UpdateActivityInfo struct {
	Id         uint32 //活动id
	Time       int64  //上次领取时间
	ActStartTm int64  //活动开始时间
}

func (updateActivityInfo *UpdateActivityInfo) String() string {
	return fmt.Sprintf("%+v\n", *updateActivityInfo)
}

// SetActivityInfo 设置更新活动信息
func (a *ActivityUtil) SetUpdateActivityInfo(info *UpdateActivityInfo) bool {
	if info == nil {
		return false
	}

	d, err := json.Marshal(info)
	if err != nil {
		log.Warn("SetUpdateActivityInfo error = ", err)
		return false
	}

	field := strconv.FormatUint(a.uid, 10) + ":"
	hSet(a.key(), field, string(d))
	return true
}

// GetActivityInfo 获取更新活动信息
func (a *ActivityUtil) GetUpdateActivityInfo(info *UpdateActivityInfo) error {
	field := strconv.FormatUint(a.uid, 10) + ":"
	v := hGet(a.key(), field)
	if err := json.Unmarshal([]byte(v), &info); err != nil {
		log.Warn("GetUpdateActivityInfo Failed to Unmarshal ", err)
		return errors.New("unmarshal error")
	}

	return nil
}

// IsUpdateActivity 是否已领取过
func (a *ActivityUtil) IsUpdateActivity() bool {
	field := strconv.FormatUint(a.uid, 10) + ":"

	return hExists(a.key(), field)
}

// ClearUpdateActivity 清空已领取过信息
func (a *ActivityUtil) ClearUpdateActivity() {
	field := strconv.FormatUint(a.uid, 10) + ":"

	hDEL(a.key(), field)
}

/*-----------------------------------------新年活动--------------------------------------*/

// NewYearActivityInfo 新年活动信息
type NewYearActivityInfo struct {
	Id          uint32           //活动id
	Time        int64            //上次领取时间
	GoodId      map[uint32]int32 //领取物品状态
	ActStartTm  int64            //活动开始时间
	TargetStage uint32           //活动阶数
}

func (newYearActivityInfo *NewYearActivityInfo) String() string {
	return fmt.Sprintf("%+v\n", *newYearActivityInfo)
}

// SetNewYearActivityInfo 设置新年活动信息
func (a *ActivityUtil) SetNewYearActivityInfo(info *NewYearActivityInfo) bool {
	if info == nil {
		return false
	}

	d, err := json.Marshal(info)
	if err != nil {
		log.Warn("SetNewYearActivityInfo error = ", err)
		return false
	}

	hSet(a.key(), a.uid, string(d))
	return true
}

// GetNewYearActivityInfo 获取新年活动信息
func (a *ActivityUtil) GetNewYearActivityInfo(info *NewYearActivityInfo) error {
	v := hGet(a.key(), a.uid)
	if err := json.Unmarshal([]byte(v), &info); err != nil {
		log.Warn("GetNewYearActivityInfo Failed to Unmarshal ", err)
		return errors.New("unmarshal error")
	}

	return nil
}

/*-----------------------------------------新年活动目标id--------------------------------------*/

// NewYearRandomInfo 新年活动随机目标值
type NewYearRandomInfo struct {
	Id         uint32            //活动id
	TargetType map[uint32]uint32 //该轮活动目标类型
	TargetNum  map[uint32]uint32 //该轮活动目标数量
	ActStartTm int64             //活动开始时间
}

// SetNewYearRandomInfo 设置该轮活动目标id
func (a *ActivityUtil) SetNewYearRandomInfo(info *NewYearRandomInfo) bool {
	if info == nil {
		return false
	}

	d, err := json.Marshal(info)
	if err != nil {
		log.Warn("SetNewYearRandomInfo error = ", err)
		return false
	}

	field := "TargetId:" + strconv.FormatUint(a.uid, 10)
	hSet(a.key(), field, string(d))
	return true
}

// GetNewYearRandomInfo 获取该轮活动目标id
func (a *ActivityUtil) GetNewYearRandomInfo(info *NewYearRandomInfo) error {
	field := "TargetId:" + strconv.FormatUint(a.uid, 10)
	v := hGet(a.key(), field)
	if err := json.Unmarshal([]byte(v), &info); err != nil {
		log.Warn("GetNewYearRandomInfo Failed to Unmarshal ", err)
		return errors.New("unmarshal error")
	}

	return nil
}

// IsNewYearRandomActivity 是否已经生成活动目标信息
func (a *ActivityUtil) IsNewYearRandomActivity() bool {
	field := "TargetId:" + strconv.FormatUint(a.uid, 10)

	return hExists(a.key(), field)
}

// ClearNewYearRandomActivity 清空该轮活动目标id
func (a *ActivityUtil) ClearNewYearRandomActivity() {
	field := "TargetId:" + strconv.FormatUint(a.uid, 10)

	hDEL(a.key(), field)
}

/*-----------------------------------------老兵召回活动--------------------------------------*/

// VeteranRecallInfo 老兵召回信息
type VeteranRecallInfo struct {
	Id           uint32            //活动id
	ActStartTm   int64             //活动开始时间
	RecallSusNum uint32            //总共召回成功数量
	Time         int64             //上次召回时间
	RecallDayNum uint32            //每天召回次数
	RecallInfo   map[uint64]int64  //已经召唤的玩家列表
	RecallReward map[uint32]uint32 //成功召回奖励领取信息
}

func (veteranRecallInfo *VeteranRecallInfo) String() string {
	return fmt.Sprintf("%+v\n", *veteranRecallInfo)
}

/*-----------------------------------------重返光荣战场任务--------------------------------------*/

// BackBattleInfo 重返光荣战场任务信息
type BackBattleInfo struct {
	Id          uint32 //活动id
	Time        int64  //上次领取时间
	GoodId      uint32 //上次领取到的物品
	Sum         uint32 //领取次数
	TaskStartTm int64  //任务开始时间
	RoundNum    uint32 //奖励开到第几轮
}

func (backBattleInfo *BackBattleInfo) String() string {
	return fmt.Sprintf("%+v\n", *backBattleInfo)
}

/*-----------------------------------------节日活动--------------------------------------*/

// FestivalActivityInfo 节日活动信息
type FestivalActivityInfo struct {
	Id         uint32            //活动id
	Time       int64             //上次领取时间
	ActStartTm int64             //活动开始时间
	FinishNum  map[uint32]uint32 //任务已完成次数
	PickState  map[uint32]uint32 //领取状态(0未领取，1领取)
	FinishTime map[uint32]int64  //每个任务最近一天完成时间
}

func (festivalActivityInfo *FestivalActivityInfo) String() string {
	return fmt.Sprintf("%+v\n", *festivalActivityInfo)
}

/*-----------------------------------------兑换活动--------------------------------------*/

// ExchangeActivityInfo 兑换活动信息
type ExchangeActivityInfo struct {
	Id          uint32            //活动id
	ActStartTm  int64             //活动开始时间
	ExchangeNum map[uint32]uint32 //已兑换次数
	CallState   map[uint32]uint32 //兑换提醒状态
}

func (exchangeActivityInfo *ExchangeActivityInfo) String() string {
	return fmt.Sprintf("%+v\n", *exchangeActivityInfo)
}

/*-----------------------------------------一球成名活动--------------------------------------*/

// BallStarInfo 一球成名信息
type BallStarInfo struct {
	Id         uint32            //活动id
	ActStartTm int64             //活动开始时间
	Time       int64             //上次抽取时间
	Position   uint32            //进球图标位置
	Sum        uint32            //总抽奖次数
	DayNum     uint32            //一天抽奖次数
	PickState  map[uint32]uint32 //领取状态(0未领取，1领取)
	FreshSum   uint32            //周期内需要刷新的次数统计
}

func (ballStarInfo *BallStarInfo) String() string {
	return fmt.Sprintf("%+v\n", *ballStarInfo)
}

/*-----------------------------------------华丽丽的分割线--------------------------------------*/

// SetInfo 设置活动信息
func (a *ActivityUtil) SetInfo(info interface{}) bool {
	if info == nil {
		return false
	}

	d, err := json.Marshal(info)
	if err != nil {
		log.Error("SetInfo error = ", err)
		return false
	}

	hSet(a.key(), a.uid, string(d))
	return true
}

// GetInfo 获取活动信息
func (a *ActivityUtil) GetInfo(info interface{}) error {

	v := hGet(a.key(), a.uid)
	if err := json.Unmarshal([]byte(v), &info); err != nil {
		log.Error("GetInfo Failed to Unmarshal ", err)
		return err
	}

	return nil
}

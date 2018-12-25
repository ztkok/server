package db

import (
	"encoding/json"
	"fmt"
	"zeus/dbservice"

	log "github.com/cihub/seelog"
	"github.com/garyburd/redigo/redis"
)

/*
	通过数据库共享随机选取的特训任务
*/

// SetOpenSpecialTasks 保存本周开放的特训任务
func SetOpenSpecialTasks(stage string, tasks map[uint8][]uint32) {
	data, err := json.Marshal(tasks)
	if err != nil {
		log.Error("Marshal err: ", err)
		return
	}

	hSet("OpenSpecialTasks", stage, string(data))
}

// GetOpenSpecialTasks 获取本周开放的特训任务
func GetOpenSpecialTasks(stage string) map[uint8][]uint32 {
	c := dbservice.Get()
	defer c.Close()

	tasks := make(map[uint8][]uint32)
	data, err := redis.String(c.Do("HGET", "OpenSpecialTasks", stage))
	if err != nil {
		log.Error("redis String err: ", err)
		return tasks
	}

	err = json.Unmarshal([]byte(data), &tasks)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return tasks
	}

	return tasks
}

const (
	taskRecordPrefix = "TaskRecord"
)

// 奖励状态
const (
	AwardsStateNotDrawable = 0 //未达到领奖条件不可领取
	AwardsStateDrawable    = 1 //达到领奖条件可领取
	AwardsStateDrawed      = 2 //已经领取
	AwardsStateExpired     = 3 //奖励过期不可领取
)

type PlayerTaskUtil struct {
	uid    uint64
	task   string
	groups uint32
}

// GetPlayerTaskUtil 获取管理玩家任务数据的工具
func GetPlayerTaskUtil(uid uint64, task, stage string, groups uint32) *PlayerTaskUtil {
	u := &PlayerTaskUtil{
		uid:    uid,
		task:   task,
		groups: groups,
	}

	if u.getTaskStage() != stage {
		u.del()
		u.setTaskStage(stage)
	}

	u.adjustGroups()

	return u
}

// GetTaskList 获取任务列表
func (u *PlayerTaskUtil) GetTaskList() [][]*TaskItemRecord {
	list, err := u.getTaskList()
	if err != nil {
		log.Error("getTaskList err: ", err)
		return nil
	}

	return list
}

// SetTaskItemProgress 记录任务项进度
func (u *PlayerTaskUtil) SetTaskItemProgress(groupId uint8, taskId uint32, num uint32) {
	list := u.GetTaskList()
	if list == nil {
		list = make([][]*TaskItemRecord, u.groups)
	}

	index := int(groupId - 1)
	var exist bool

	for _, v := range list[index] {
		if v.TaskId == taskId {
			v.FinishedNum = num
			exist = true
			break
		}
	}

	if !exist {
		list[index] = append(list[index], &TaskItemRecord{
			TaskId:      taskId,
			FinishedNum: num,
		})
	}

	u.setTaskList(list)
}

// GetTaskItemProgress 获取任务项进度
func (u *PlayerTaskUtil) GetTaskItemProgress(groupId uint8, taskId uint32) uint32 {
	list := u.GetTaskList()
	if list == nil {
		return 0
	}

	index := int(groupId - 1)
	for _, v := range list[index] {
		if v.TaskId == taskId {
			return v.FinishedNum
		}
	}

	return 0
}

// SetTaskItemAwardsState 设置任务项奖励状态
func (u *PlayerTaskUtil) SetTaskItemAwardsState(groupId uint8, taskId uint32, state uint8) {
	list := u.GetTaskList()
	if list == nil {
		list = make([][]*TaskItemRecord, u.groups)
	}

	index := int(groupId - 1)
	var exist bool

	for _, v := range list[index] {
		if v.TaskId == taskId {
			v.State = state
			exist = true
			break
		}
	}

	if !exist {
		list[index] = append(list[index], &TaskItemRecord{
			TaskId: taskId,
			State:  state,
		})
	}

	u.setTaskList(list)
}

// GetTaskItemAwardsState 判断任务项奖励是否是指定状态
func (u *PlayerTaskUtil) GetTaskItemAwardsState(groupId uint8, taskId uint32) uint8 {
	state := uint8(AwardsStateNotDrawable)

	list := u.GetTaskList()
	if list == nil {
		return state
	}

	index := int(groupId - 1)
	for _, v := range list[index] {
		if v.TaskId == taskId {
			state = v.State
			break
		}
	}

	return state
}

// GetTaskAwards 获取任务奖励记录
func (u *PlayerTaskUtil) GetTaskAwards() []*AwardInfo {
	awards, err := u.getTaskAwards("Awards")
	if err != nil {
		log.Error("getTaskAwards err: ", err)
		return nil
	}

	return awards
}

// SetTaskAwardsState 设置任务奖励状态
func (u *PlayerTaskUtil) SetTaskAwardsState(id uint32, state uint8) {
	u.setTaskAwardsState("Awards", id, state)
}

// GetTaskAwardsState 获取任务奖励状态
func (u *PlayerTaskUtil) GetTaskAwardsState(id uint32) uint8 {
	return u.getTaskAwardsState("Awards", id)
}

// adjustGroups 任务组数增加时调整数据库记录
func (u *PlayerTaskUtil) adjustGroups() {
	list := u.GetTaskList()
	if list != nil && len(list) < int(u.groups) {
		for len(list) < int(u.groups) {
			list = append(list, []*TaskItemRecord(nil))
		}
		u.setTaskList(list)
	}

	enables := u.GetGroupEnables()
	if enables != nil && len(enables) < int(u.groups) {
		for len(enables) < int(u.groups) {
			enables = append(enables, false)
		}
		u.setGroupEnables(enables)
	}
}

func (u *PlayerTaskUtil) setTaskStage(stage string) error {
	return u.setValue("Stage", stage)
}

func (u *PlayerTaskUtil) getTaskStage() string {
	return u.getStringValue("Stage")
}

func (u *PlayerTaskUtil) setTaskList(list [][]*TaskItemRecord) error {
	return u.setComplexValue("TaskList", list)
}

func (u *PlayerTaskUtil) getTaskList() ([][]*TaskItemRecord, error) {
	if !hExists(u.key(), "TaskList") {
		return nil, nil
	}

	data, err := redis.String(u.getValue("TaskList"))
	if err != nil {
		return nil, err
	}

	var list [][]*TaskItemRecord

	err = json.Unmarshal([]byte(data), &list)
	if err != nil {
		return nil, err
	}

	return list, nil
}

// setTaskAwardsState
func (u *PlayerTaskUtil) setTaskAwardsState(field string, id uint32, state uint8) {
	awards, err := u.getTaskAwards(field)
	if err != nil {
		log.Error("getTaskAwards err: ", err)
		return
	}

	var exist bool
	for _, v := range awards {
		if v.Id == id {
			v.State = state
			exist = true
			break
		}
	}

	if !exist {
		awards = append(awards, &AwardInfo{
			Id:    id,
			State: state,
		})
	}

	u.setComplexValue(field, awards)
}

// getTaskAwardsState
func (u *PlayerTaskUtil) getTaskAwardsState(field string, id uint32) uint8 {
	state := uint8(AwardsStateNotDrawable)

	awards, err := u.getTaskAwards(field)
	if err != nil {
		log.Error("getTaskAwards err: ", err)
		return state
	}

	for _, v := range awards {
		if v.Id == id {
			state = v.State
			break
		}
	}

	return state
}

func (u *PlayerTaskUtil) getTaskAwards(field string) ([]*AwardInfo, error) {
	if !hExists(u.key(), field) {
		return nil, nil
	}

	data, err := redis.String(u.getValue(field))
	if err != nil {
		return nil, err
	}

	var awards []*AwardInfo

	err = json.Unmarshal([]byte(data), &awards)
	if err != nil {
		return nil, err
	}

	return awards, nil
}

func (u *PlayerTaskUtil) getValue(field string) (interface{}, error) {
	c := dbservice.Get()
	defer c.Close()
	return c.Do("HGET", u.key(), field)
}

func (u *PlayerTaskUtil) getUintValue(field string) uint64 {
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

func (u *PlayerTaskUtil) getStringValue(field string) string {
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

func (u *PlayerTaskUtil) getBoolValue(field string) bool {
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

func (u *PlayerTaskUtil) setValue(field string, value interface{}) error {
	c := dbservice.Get()
	defer c.Close()
	_, err := c.Do("HSET", u.key(), field, value)
	return err
}

func (u *PlayerTaskUtil) setComplexValue(field string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return u.setValue(field, string(data))
}

func (u *PlayerTaskUtil) del() {
	delKey(u.key())
}

func (u *PlayerTaskUtil) key() string {
	return fmt.Sprintf("%s:%s:%d", taskRecordPrefix, u.task, u.uid)
}

/*******************************特训任务*********************************/

// SetSpecialEnable 激活精英权限
func (u *PlayerTaskUtil) SetSpecialEnable() {
	u.setValue("Enable", true)
}

// IsSpecialEnable 是否激活精英权限
func (u *PlayerTaskUtil) IsSpecialEnable() bool {
	return u.getBoolValue("Enable")
}

// AddSpecialMedals 添加特训勋章
func (u *PlayerTaskUtil) AddSpecialMedals(num uint32) {
	medals := u.GetSpecialMedals()
	medals += num

	u.setValue("Medals", medals)
}

// GetSpecialMedals 获取特训勋章
func (u *PlayerTaskUtil) GetSpecialMedals() uint32 {
	return uint32(u.getUintValue("Medals"))
}

// IncrSpecialReplacements 增加特训奖励补领次数
func (u *PlayerTaskUtil) IncrSpecialReplacements() {
	reps := u.GetSpecialReplacements()
	reps++

	u.setValue("Replacements", reps)
}

// GetSpecialReplacements 获取特训奖励补领次数
func (u *PlayerTaskUtil) GetSpecialReplacements() uint32 {
	return uint32(u.getUintValue("Replacements"))
}

/*******************************挑战任务*********************************/

// SetEliteEnable 激活精英权限
func (u *PlayerTaskUtil) SetEliteEnable() {
	u.setValue("Enable", true)
}

// IsEliteEnable 是否激活精英权限
func (u *PlayerTaskUtil) IsEliteEnable() bool {
	return u.getBoolValue("Enable")
}

// SetSeasonGrade 记录赛季等级
func (u *PlayerTaskUtil) SetSeasonGrade(grade uint32) {
	u.setValue("Grade", grade)
}

// GetGradeAndMedals 获取赛季等级
func (u *PlayerTaskUtil) GetSeasonGrade() uint32 {
	grade := uint32(u.getUintValue("Grade"))
	if grade == 0 {
		u.SetSeasonGrade(1)
		return 1
	}
	return grade
}

// SetExpMedals 记录经验勋章
func (u *PlayerTaskUtil) SetExpMedals(medals uint32) {
	u.setValue("Medals", medals)
}

// GetExpMedals 获取经验勋章
func (u *PlayerTaskUtil) GetExpMedals() uint32 {
	return uint32(u.getUintValue("Medals"))
}

// SetEliteAwardsState 设置精英奖励状态
func (u *PlayerTaskUtil) SetEliteAwardsState(id uint32, state uint8) {
	u.setTaskAwardsState("EliteAwards", id, state)
}

// GetEliteAwardsState 获取精英奖励状态
func (u *PlayerTaskUtil) GetEliteAwardsState(id uint32) uint8 {
	return u.getTaskAwardsState("EliteAwards", id)
}

// GetEliteAwards 获取精英奖励记录
func (u *PlayerTaskUtil) GetEliteAwards() []*AwardInfo {
	awards, err := u.getTaskAwards("EliteAwards")
	if err != nil {
		log.Error("getTaskAwards err: ", err)
		return nil
	}

	return awards
}

// GetGroupEnable 激活任务组
func (u *PlayerTaskUtil) SetGroupEnable(groupId uint8) {
	enables := u.GetGroupEnables()
	if enables == nil {
		enables = make([]bool, u.groups)
	}

	enables[int(groupId-1)] = true

	u.setGroupEnables(enables)
}

// GetGroupEnables 获取任务组的激活状态
func (u *PlayerTaskUtil) GetGroupEnables() []bool {
	enables, err := u.getGroupEnables()
	if err != nil {
		log.Error("getGroupEnables err: ", err)
		return nil
	}
	return enables
}

func (u *PlayerTaskUtil) setGroupEnables(enables []bool) error {
	return u.setComplexValue("GroupEnables", enables)
}

func (u *PlayerTaskUtil) getGroupEnables() ([]bool, error) {
	if !hExists(u.key(), "GroupEnables") {
		return nil, nil
	}

	data, err := redis.String(u.getValue("GroupEnables"))
	if err != nil {
		return nil, err
	}

	var enables []bool

	err = json.Unmarshal([]byte(data), &enables)
	if err != nil {
		return nil, err
	}

	return enables, nil
}

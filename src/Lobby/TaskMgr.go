package main

import (
	"common"
	"db"
	"protoMsg"
	"zeus/iserver"
)

// 任务必须实现的接口
type iTask interface {
	init(user *LobbyUser, name string)
	getTaskStage() string

	syncTaskDetail()
	checkForEnableTaskGroup()

	getTaskType(taskId uint32) uint8
	getRequireNum(taskId uint32) uint32
	getTaskIdsByPool(groupId uint8, taskPoolId uint32) []uint32

	enableTaskItem(taskId uint32)
	isTaskItemEnable(taskId uint32) bool

	isTaskItemFinished(groupId uint8, taskId uint32) bool
	getEnabledGroups() []uint8

	updateTaskItemProgress(groupId uint8, taskId, incr uint32) (uint32, uint8)
	updateTaskPool(groupId uint8, taskPoolId, incr uint32) []*protoMsg.TaskItem
	updateTaskItems(items ...uint32)

	getTaskItemAwards(taskId uint32) map[uint32]uint32
	canDrawTaskItemAwards(groupId uint8, taskId uint32) bool
	drawTaskItemAwards(groupId uint8, taskId uint32)
	replaceTaskItemAwards(groupId uint8, taskId uint32)

	getTaskAwards(id uint32, typ uint8) map[uint32]uint32
	canDrawTaskAwards(id uint32, typ uint8) bool
	drawTaskAwards(id uint32, typ uint8)
}

// 任务管理器
type TaskMgr struct {
	user *LobbyUser

	tasks map[string]iTask
}

// NewTaskMgr 创建任务管理器
func NewTaskMgr(user *LobbyUser) *TaskMgr {
	mgr := &TaskMgr{
		user:  user,
		tasks: make(map[string]iTask),
	}

	mgr.tasks[common.TaskName_Special] = &specialTask{}
	mgr.tasks[common.TaskName_Challenge] = &challengeTask{}

	for name, t := range mgr.tasks {
		t.init(user, name)
	}

	return mgr
}

// getSpecialTask 获取特训任务
func (mgr *TaskMgr) getSpecialTask() *specialTask {
	it, ok := mgr.tasks[common.TaskName_Special]
	if !ok {
		return nil
	}

	t, ok := it.(*specialTask)
	if !ok {
		return nil
	}

	return t
}

// getChallengeTask 获取挑战任务
func (mgr *TaskMgr) getChallengeTask() *challengeTask {
	it, ok := mgr.tasks[common.TaskName_Challenge]
	if !ok {
		return nil
	}

	t, ok := it.(*challengeTask)
	if !ok {
		return nil
	}

	return t
}

// syncTaskDetailAll 向客户端同步所有任务详情
func (mgr *TaskMgr) syncTaskDetailAll() {
	for _, v := range mgr.tasks {
		v.syncTaskDetail()
	}
}

// updateTaskItemsByType 更新某一项任务的任务进度
func (mgr *TaskMgr) updateTaskItemsByType(name string, items ...uint32) {
	it, ok := mgr.tasks[name]
	if !ok {
		return
	}

	it.updateTaskItems(items...)
}

// updateTaskItemsAll 更新所有任务的任务进度
func (mgr *TaskMgr) updateTaskItemsAll(items ...uint32) {
	for _, v := range mgr.tasks {
		v.updateTaskItems(items...)
	}
}

// enableTaskItemByType 激活任务项
func (mgr *TaskMgr) enableTaskItemByType(name string, taskId uint32) {
	it, ok := mgr.tasks[name]
	if !ok {
		return
	}

	it.enableTaskItem(taskId)
}

// enableTaskGroupByType 激活任务组
func (mgr *TaskMgr) enableTaskGroupByType(name string) {
	it, ok := mgr.tasks[name]
	if !ok {
		return
	}

	it.checkForEnableTaskGroup()
}

// replaceTaskItemAwardsByType 补领任务项奖励
func (mgr *TaskMgr) replaceTaskItemAwardsByType(name string, groupId uint8, taskId uint32) {
	it, ok := mgr.tasks[name]
	if !ok {
		return
	}

	it.replaceTaskItemAwards(groupId, taskId)
}

// drawTaskItemAwardsByType 领取任务项奖励
func (mgr *TaskMgr) drawTaskItemAwardsByType(name string, groupId uint8, taskId uint32) {
	it, ok := mgr.tasks[name]
	if !ok {
		return
	}

	it.drawTaskItemAwards(groupId, taskId)
}

// drawTaskAwardsByType 领取任务奖励
func (mgr *TaskMgr) drawTaskAwardsByType(name string, id uint32, typ uint8) {
	it, ok := mgr.tasks[name]
	if !ok {
		return
	}

	it.drawTaskAwards(id, typ)
}

// 任务基类
type task struct {
	user    *LobbyUser
	realPtr interface{}

	name        string //任务名
	awardReason int    //奖励标签
}

// getTaskUtil 获取任务数据管理工具
func (t *task) getTaskUtil() *db.PlayerTaskUtil {
	it, ok := t.realPtr.(iTask)
	if !ok {
		return nil
	}

	stage := it.getTaskStage()
	v, ok := GetSrvInst().taskGroups.Load(t.name)
	if !ok {
		return nil
	}

	groups := v.(uint32)
	return db.GetPlayerTaskUtil(t.user.GetDBID(), t.name, stage, groups)
}

// isTaskItemFinished 任务项是否已经完成
func (t *task) isTaskItemFinished(groupId uint8, taskId uint32) bool {
	it, ok := t.realPtr.(iTask)
	if !ok {
		return false
	}

	util := t.getTaskUtil()
	if util == nil {
		return false
	}

	cur := util.GetTaskItemProgress(groupId, taskId)
	max := it.getRequireNum(taskId)

	return cur >= max
}

// updateTaskItemProgress 更新任务项进度
func (t *task) updateTaskItemProgress(groupId uint8, taskId, incr uint32) (uint32, uint8) {
	state := uint8(db.AwardsStateNotDrawable)
	it, ok := t.realPtr.(iTask)
	if !ok {
		return 0, state
	}

	util := t.getTaskUtil()
	if util == nil {
		return 0, state
	}

	cur := util.GetTaskItemProgress(groupId, taskId)
	max := it.getRequireNum(taskId)
	if cur >= max {
		state = util.GetTaskItemAwardsState(groupId, taskId)
		return cur, state
	}

	cur += incr
	if cur >= max {
		cur = max
		state = db.AwardsStateDrawable
	}

	util.SetTaskItemProgress(groupId, taskId, cur)
	util.SetTaskItemAwardsState(groupId, taskId, state)

	return cur, state
}

// updateTaskPool 更新任务池进度
func (t *task) updateTaskPool(groupId uint8, taskPoolId, incr uint32) []*protoMsg.TaskItem {
	if incr == 0 {
		return nil
	}

	it, ok := t.realPtr.(iTask)
	if !ok {
		return nil
	}

	items := []*protoMsg.TaskItem{}
	maxs := make(map[uint8]uint32)

	for _, taskId := range it.getTaskIdsByPool(groupId, taskPoolId) {
		if t.isTaskItemFinished(groupId, taskId) {
			continue
		}

		cur, state := it.updateTaskItemProgress(groupId, taskId, incr)
		items = append(items, &protoMsg.TaskItem{
			TaskId:      taskId,
			FinishedNum: cur,
			State:       uint32(state),
		})

		typ := it.getTaskType(taskId)
		if cur > maxs[typ] {
			maxs[typ] = cur
		}
	}

	for k, v := range maxs {
		if v > 0 {
			t.user.TaskFlow(t.name, k, taskPoolId, v)
		}
	}

	return items
}

// updateTaskItems 更新任务完成进度
func (t *task) updateTaskItems(items ...uint32) {
	it, ok := t.realPtr.(iTask)
	if !ok {
		return
	}

	v, ok := GetSrvInst().openTasks.Load(t.name)
	if !ok {
		return
	}

	taskM := v.(map[uint8][]uint32)

	for _, groupId := range it.getEnabledGroups() {
		msg := &protoMsg.TaskUpdate{
			TaskName: t.name,
			Groupid:  uint32(groupId),
		}

		for _, taskPoolId := range taskM[2*groupId] {
			for i := 0; i < len(items); i += 2 {
				if items[i] == taskPoolId {
					tmpitems := t.updateTaskPool(groupId, items[i], items[i+1])
					if len(tmpitems) > 0 {
						msg.TaskItems = append(msg.TaskItems, tmpitems...)
					}
					break
				}
			}
		}

		if len(msg.TaskItems) > 0 {
			t.user.RPC(iserver.ServerTypeClient, "TaskUpdateNotify", msg)
			t.user.Infof("TaskUpdateNotify: %+v\n", msg)

			for _, item := range msg.TaskItems {
				if item.State == db.AwardsStateDrawable {
					it.checkForEnableTaskGroup()
					break
				}
			}
		}
	}
}

// canDrawTaskItemAwards 是否可以领取任务奖励
// whatDay表示本周的第几天(1-7)
func (t *task) canDrawTaskItemAwards(groupId uint8, taskId uint32) bool {
	it, ok := t.realPtr.(iTask)
	if !ok {
		return false
	}

	util := t.getTaskUtil()
	if util == nil {
		return false
	}

	if !it.isTaskItemEnable(taskId) {
		return false
	}

	return util.GetTaskItemAwardsState(groupId, taskId) == db.AwardsStateDrawable
}

// drawTaskItemAwards 领取任务项奖励
// whatDay表示本周的第几天(1-7)
func (t *task) drawTaskItemAwards(groupId uint8, taskId uint32) {
	it, ok := t.realPtr.(iTask)
	if !ok {
		return
	}

	util := t.getTaskUtil()
	if util == nil {
		return
	}

	ret := &protoMsg.TaskItem{
		TaskId: taskId,
	}

	if t.canDrawTaskItemAwards(groupId, taskId) {
		util.SetTaskItemAwardsState(groupId, taskId, db.AwardsStateDrawed)

		awards := it.getTaskItemAwards(taskId)
		t.user.storeMgr.GetAwards(awards, uint32(t.awardReason), false, false)
	}

	ret.FinishedNum = util.GetTaskItemProgress(groupId, taskId)
	ret.State = uint32(util.GetTaskItemAwardsState(groupId, taskId))

	t.user.RPC(iserver.ServerTypeClient, "DrawTaskItemAwardsRet", ret, t.name, uint32(groupId))
	t.user.Info("drawTaskItemAwards, ret: ", *ret, " name: ", t.name, " groupId: ", groupId)
}

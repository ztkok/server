package main

import (
	"common"
	"db"
	"excel"
	"protoMsg"
	"zeus/iserver"
)

// 特训任务
type specialTask struct {
	task
}

// init 初始化任务数据
func (t *specialTask) init(user *LobbyUser, name string) {
	t.user = user
	t.realPtr = t
	t.name = name
	t.awardReason = common.RS_SpecialTask
}

// getTaskStage 获取任务阶段标识
func (t *specialTask) getTaskStage() string {
	return common.GetSpecialTaskWeek()
}

// getTaskType 获取任务项的类型 0表示普通任务 1表示精英任务
func (t *specialTask) getTaskType(taskId uint32) uint8 {
	item, ok := excel.GetSpecialTask(uint64(taskId))
	if !ok {
		return 0
	}

	return uint8(item.Tasktype)
}

// getRequireNum 获取任务项需要达到的最大数量
func (t *specialTask) getRequireNum(taskId uint32) uint32 {
	item, ok := excel.GetSpecialTask(uint64(taskId))
	if !ok {
		return 0
	}

	return uint32(item.Requirenum)
}

// getTaskItemAwards 获取任务项奖励数据
func (t *specialTask) getTaskItemAwards(taskId uint32) map[uint32]uint32 {
	item, ok := excel.GetSpecialTask(uint64(taskId))
	if !ok {
		return make(map[uint32]uint32)
	}

	return common.StringToMapUint32(item.Awards, "|", ";")
}

// getTaskAwards 获取任务奖励数据
func (t *specialTask) getTaskAwards(id uint32, typ uint8) map[uint32]uint32 {
	item, ok := excel.GetSpecialLevel(uint64(id))
	if !ok {
		return make(map[uint32]uint32)
	}

	return common.StringToMapUint32(item.Awards, "|", ";")
}

// syncTaskDetail 向客户端同步任务详情
func (t *specialTask) syncTaskDetail() {
	util := t.getTaskUtil()
	list := util.GetTaskList()
	awards := util.GetTaskAwards()

	notify := &protoMsg.SpecialTaskDetail{
		Detail:       &protoMsg.TaskDetail{},
		Enable:       util.IsSpecialEnable(),
		Medals:       util.GetSpecialMedals(),
		Replacements: util.GetSpecialReplacements(),
		Awards:       &protoMsg.AwardList{},
	}

	whatDay := common.WhatDayOfSpecialWeek(common.GetSpecialTaskDay())
	notify.Detail.Name_ = t.name
	notify.Detail.WhatDay = uint32(whatDay)

	for i := 0; i < 7; i++ {
		notify.Detail.Lists = append(notify.Detail.Lists, &protoMsg.TaskList{})
	}

	v, ok := GetSrvInst().openTasks.Load(t.name)
	if !ok {
		return
	}

	taskM := v.(map[uint8][]uint32)
	for k, s := range taskM {
		if k%2 == 0 {
			continue
		}
		index := int(k / 2)

		for _, v := range s {
			item := &protoMsg.TaskItem{
				TaskId: v,
			}

			if list != nil {
				for _, j := range list[index] {
					if j.TaskId == v {
						item.FinishedNum = j.FinishedNum
						item.State = uint32(j.State)
						break
					}
				}
			}

			if index < int(whatDay-1) && item.GetState() != db.AwardsStateDrawed {
				item.State = db.AwardsStateExpired
			}

			notify.Detail.Lists[index].TaskItems = append(notify.Detail.Lists[index].TaskItems, item)
		}
	}

	v, ok = GetSrvInst().taskAwards.Load(t.name)
	if !ok {
		return
	}

	awardM := v.([]uint32)
	for _, v := range awardM {
		info := &protoMsg.AwardInfo{
			Id: v,
		}

		for _, j := range awards {
			if j.Id == v {
				info.State = uint32(j.State)
			}
		}

		notify.Awards.Infos = append(notify.Awards.Infos, info)
	}

	t.user.RPC(iserver.ServerTypeClient, "TaskDetailNotify", notify)
	t.user.Infof("TaskDetailNotify: %+v\n", notify)
}

// getEnabledGroups 获取激活的任务组
func (t *specialTask) getEnabledGroups() []uint8 {
	whatDay := common.WhatDayOfSpecialWeek(common.GetSpecialTaskDay())
	return []uint8{whatDay}
}

// getTaskIdsByPool 根据任务池获取对应的开放任务
func (t *specialTask) getTaskIdsByPool(groupId uint8, taskPoolId uint32) []uint32 {
	res := []uint32{}

	v, ok := GetSrvInst().openTasks.Load(t.name)
	if !ok {
		return res
	}

	taskM := v.(map[uint8][]uint32)
	for _, v := range taskM[2*groupId-1] {
		task, ok := excel.GetSpecialTask(uint64(v))
		if !ok {
			continue
		}

		if task.Taskpool == uint64(taskPoolId) {
			res = append(res, v)
		}
	}

	return res
}

// isTaskItemEnable 任务项是否已经激活
func (t *specialTask) isTaskItemEnable(taskId uint32) bool {
	util := t.getTaskUtil()
	item, ok := excel.GetSpecialTask(uint64(taskId))
	if !ok {
		return false
	}

	if item.Tasktype == 1 && !util.IsSpecialEnable() {
		return false
	}

	return true
}

// enableTaskItem 激活精英任务项
func (t *specialTask) enableTaskItem(taskId uint32) {
	util := t.getTaskUtil()
	if util == nil {
		return
	}

	ret := uint8(1)
	goods := t.getEnableTaskItemGoods(taskId)

	if !util.IsSpecialEnable() && t.user.storeMgr.IsAllGoodsEnough(goods) {
		t.user.storeMgr.ReduceGoodsAll(goods, common.RS_SpecialTask)
		util.SetSpecialEnable()
		ret = 0
	}

	t.user.RPC(iserver.ServerTypeClient, "EnableTaskItemRet", ret, taskId)
}

// replaceTaskItemAwards 补领任务项奖励
func (t *specialTask) replaceTaskItemAwards(groupId uint8, taskId uint32) {
	ret := uint8(1)
	util := t.getTaskUtil()
	goods := t.getEnableTaskReplacementGoods(taskId)
	whatDay := common.WhatDayOfSpecialWeek(common.GetSpecialTaskDay())

	expired := false
	if groupId < whatDay && util.GetTaskItemAwardsState(uint8(groupId), taskId) != db.AwardsStateDrawed {
		expired = true
	}

	if expired && util.GetSpecialReplacements() < t.getMaxReplacements() && t.user.storeMgr.IsAllGoodsEnough(goods) {
		t.user.storeMgr.ReduceGoodsAll(goods, common.RS_SpecialTask)
		util.IncrSpecialReplacements()

		util.SetTaskItemAwardsState(groupId, taskId, db.AwardsStateDrawable)
		t.drawTaskItemAwards(groupId, taskId)

		ret = 0
	}

	t.user.RPC(iserver.ServerTypeClient, "ReplaceTaskItemAwardsRet", ret, util.GetSpecialReplacements())
}

// updateSpecialAwards 更新特训奖励状态
func (t *specialTask) updateSpecialAwards(num uint32) {
	util := t.getTaskUtil()
	num1 := util.GetSpecialMedals()
	util.AddSpecialMedals(num)

	num2 := util.GetSpecialMedals()
	normalAwards := &protoMsg.AwardList{}
	eliteAwards := &protoMsg.AwardList{}

	for _, v := range t.getNewDrawableAwards(num1, num2) {
		util.SetTaskAwardsState(v, db.AwardsStateDrawable)
		normalAwards.Infos = append(normalAwards.Infos, &protoMsg.AwardInfo{
			Id:    v,
			State: db.AwardsStateDrawable,
		})
	}

	t.user.RPC(iserver.ServerTypeClient, "UpdateTaskAwardsState", num2, normalAwards, t.name, uint32(0), eliteAwards)
	t.user.SpecialExpFlow(num2)
}

// getNewDrawableAwards 获取刚达到领取条件的奖励
func (t *specialTask) getNewDrawableAwards(num1, num2 uint32) []uint32 {
	res := []uint32{}

	v, ok := GetSrvInst().taskAwards.Load(t.name)
	if !ok {
		return res
	}

	awardM := v.([]uint32)
	for _, v := range awardM {
		item, ok := excel.GetSpecialLevel(uint64(v))
		if !ok {
			continue
		}

		if item.Requirenum > uint64(num1) && item.Requirenum <= uint64(num2) {
			res = append(res, v)
		}
	}

	return res
}

// getEnableTaskItemGoods 获取用于激活任务项的道具
func (t *specialTask) getEnableTaskItemGoods(taskId uint32) map[uint32]uint32 {
	str := common.GetTBTwoSystemValue(common.System2_SpecialTaskEnableItem)
	return common.StringToMapUint32(str, "|", ";")
}

// getEnableTaskReplacementGoods 获取用于激活补领任务项的道具
func (t *specialTask) getEnableTaskReplacementGoods(taskId uint32) map[uint32]uint32 {
	res := make(map[uint32]uint32)
	util := t.getTaskUtil()

	item, ok := excel.GetSpecialReplace(uint64(util.GetSpecialReplacements() + 1))
	if !ok {
		return res
	}

	return common.StringToMapUint32(item.Items, "|", ";")
}

// getMaxReplacements 获取最大补领次数
func (t *specialTask) getMaxReplacements() uint32 {
	util := t.getTaskUtil()
	normal := common.GetTBSystemValue(common.System_SpecialMaxReplacements)
	add := common.GetTBSystemValue(common.System_SpecialAddReplacements)

	if util.IsSpecialEnable() {
		return uint32(normal + add)
	}

	return uint32(normal)
}

// checkForEnableTaskGroup 检测并激活任务组
func (t *specialTask) checkForEnableTaskGroup() {

}

// canDrawTaskAwards 能否领取任务奖励
func (t *specialTask) canDrawTaskAwards(id uint32, typ uint8) bool {
	if typ != 0 {
		return false
	}

	util := t.getTaskUtil()
	return util.GetTaskAwardsState(id) == db.AwardsStateDrawable
}

// drawTaskAwards 领取任务奖励
func (t *specialTask) drawTaskAwards(id uint32, typ uint8) {
	if typ != 0 {
		return
	}

	util := t.getTaskUtil()
	ret := &protoMsg.AwardInfo{
		Id: id,
	}

	if t.canDrawTaskAwards(id, typ) {
		util.SetTaskAwardsState(id, db.AwardsStateDrawed)

		awards := t.getTaskAwards(id, typ)
		t.user.storeMgr.GetAwards(awards, uint32(t.awardReason), false, false)
	}

	ret.State = uint32(util.GetTaskAwardsState(id))

	t.user.RPC(iserver.ServerTypeClient, "DrawTaskAwardsRet", ret, t.name, typ)
	t.user.Info("drawTaskAwards, ret: ", *ret, " name: ", t.name, " typ: ", typ)
}

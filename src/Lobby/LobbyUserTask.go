package main

import (
	"common"
	"db"
	"excel"
	"protoMsg"
	"sort"
	"zeus/iserver"
)

//实现sort.Interface接口
type TaskItemSlice []*protoMsg.TaskItem

func (items TaskItemSlice) Len() int           { return len(items) }
func (items TaskItemSlice) Swap(i, j int)      { items[i], items[j] = items[j], items[i] }
func (items TaskItemSlice) Less(i, j int) bool { return items[i].TaskId < items[j].TaskId }

/*-------------------------每日任务-----------------------------*/

// isDayTaskItemFinished 每日任务项是否已经完成
func (user *LobbyUser) isDayTaskItemFinished(taskId uint32) bool {
	today := common.GetTodayBeginStamp()
	cur := db.PlayerInfoUtil(user.GetDBID()).GetDayTaskItemProgress(today, taskId)

	item, ok := excel.GetTask(uint64(taskId))
	if !ok {
		return false
	}

	return cur >= uint32(item.Requirenum)
}

// updateDayTaskItemProgress 更新每日任务项的进度
func (user *LobbyUser) updateDayTaskItemProgress(taskId uint32, incr uint32) uint32 {
	util := db.PlayerInfoUtil(user.GetDBID())
	today := common.GetTodayBeginStamp()
	cur := util.GetDayTaskItemProgress(today, taskId)

	if incr == 0 {
		return cur
	}

	item, ok := excel.GetTask(uint64(taskId))
	if !ok {
		return cur
	}

	max := uint32(item.Requirenum)
	if cur >= max {
		return cur
	}

	cur += incr
	if cur >= max {
		cur = max
	}

	util.SetDayTaskItemProgress(today, taskId, cur)

	return cur
}

// updateActivenessProgress 更新活跃度
func (user *LobbyUser) updateActivenessProgress(typ uint8, incr uint32) uint32 {
	var stamp int64
	if typ == 1 {
		stamp = common.GetTodayBeginStamp()
	} else if typ == 2 {
		stamp = common.GetThisWeekBeginStamp()
	}

	util := db.PlayerInfoUtil(user.GetDBID())
	cur := util.GetActiveness(typ, stamp)
	max := common.GetMaxActiveness(typ)

	if incr == 0 || cur >= max {
		return cur
	}

	if cur+incr > max {
		incr = max - cur
	}

	util.AddActiveness(typ, stamp, incr)
	cur += incr

	return cur
}

// canDrawDayTaskAwards 玩家是否可以领取每日任务相关奖励
func (user *LobbyUser) canDrawDayTaskAwards(typ uint8, id uint32) bool {
	util := db.PlayerInfoUtil(user.GetDBID())

	switch typ {
	case 1: //日活跃
		stamp := common.GetTodayBeginStamp()
		if util.IsDayTaskDrawed(typ, stamp, id) {
			return false
		}

		box, ok := excel.GetActiveness(uint64(id))
		if !ok {
			return false
		}

		if util.GetActiveness(typ, stamp) >= uint32(box.Maxactive) {
			return true
		}
	case 2: //周活跃
		stamp := common.GetThisWeekBeginStamp()
		if util.IsDayTaskDrawed(typ, stamp, id) {
			return false
		}

		box, ok := excel.GetActiveness(uint64(id))
		if !ok {
			return false
		}

		if util.GetActiveness(typ, stamp) >= uint32(box.Maxactive) {
			return true
		}
	case 3: //每日任务项
		stamp := common.GetTodayBeginStamp()
		if util.IsDayTaskDrawed(typ, stamp, id) {
			return false
		}

		item, ok := excel.GetTask(uint64(id))
		if !ok {
			return false
		}

		if util.GetDayTaskItemProgress(stamp, id) >= uint32(item.Requirenum) {
			return true
		}
	}

	return false
}

// dayTaskInfoNotify 通知客户端每日任务相关数据
func (user *LobbyUser) dayTaskInfoNotify() {
	notify := &protoMsg.DayTaskDetail{}
	util := db.PlayerInfoUtil(user.GetDBID())
	record := util.GetDayTaskRecord(common.GetTodayBeginStamp(), common.GetThisWeekBeginStamp())

	if record != nil {
		notify.DayActiveness = record.DayActiveness
		notify.WeekActiveness = record.WeekActiveness
	}

	for _, box := range excel.GetActivenessMap() {
		switch box.Resetloop {
		case 1:
			boxItem := &protoMsg.ActiveAwardsBox{
				BoxId: uint32(box.Id),
			}

			if record != nil {
				for _, item := range record.DayActiveAwards {
					if item.BoxId == boxItem.BoxId {
						boxItem.Awards = uint32(item.Awards)
						break
					}
				}
			}

			notify.DayBoxs = append(notify.DayBoxs, boxItem)
		case 7:
			boxItem := &protoMsg.ActiveAwardsBox{
				BoxId: uint32(box.Id),
			}

			if record != nil {
				for _, item := range record.WeekActiveAwards {
					if item.BoxId == boxItem.BoxId {
						boxItem.Awards = uint32(item.Awards)
						break
					}
				}
			}

			notify.WeekBoxs = append(notify.WeekBoxs, boxItem)
		}
	}

	for _, task := range excel.GetTaskMap() {
		taskItem := &protoMsg.TaskItem{
			TaskId: uint32(task.Id),
		}

		if record != nil {
			for _, item := range record.DayTaskItems {
				if item.TaskId == taskItem.TaskId {
					taskItem.FinishedNum = item.FinishedNum
					taskItem.Awards = uint32(item.Awards)
					break
				}
			}
		}

		notify.DayTaskList = append(notify.DayTaskList, taskItem)
	}

	sort.Sort(TaskItemSlice(notify.DayTaskList))

	user.RPC(iserver.ServerTypeClient, "DayTaskInfoNotify", notify)
	user.Infof("DayTaskInfoNotify: %+v\n", notify)
}

// updateDayTaskItem 更新玩家的每日任务项完成进度
func (user *LobbyUser) updateDayTaskItems(items ...uint32) {
	msg := &protoMsg.TaskUpdate{}

	for i := 0; i < len(items); i += 2 {
		tmpitems := user.updateDayTaskPool(items[i], items[i+1])
		if len(tmpitems) > 0 {
			msg.TaskItems = append(msg.TaskItems, tmpitems...)
		}
	}

	if len(msg.TaskItems) > 0 {
		user.RPC(iserver.ServerTypeClient, "DayTaskUpdateNotify", msg)
		user.Infof("DayTaskUpdateNotify: %+v\n", msg)
	}
}

// updateDayTaskPool 更新玩家的每日任务池的任务进度
func (user *LobbyUser) updateDayTaskPool(taskPoolId, incr uint32) []*protoMsg.TaskItem {
	if incr == 0 {
		return nil
	}

	var (
		items []*protoMsg.TaskItem
		max   uint32
	)

	for _, taskId := range common.GetDayTaskIds(taskPoolId) {
		if user.isDayTaskItemFinished(taskId) {
			continue
		}

		cur := user.updateDayTaskItemProgress(taskId, incr)
		items = append(items, &protoMsg.TaskItem{
			TaskId:      taskId,
			FinishedNum: cur,
		})

		if cur > max {
			max = cur
		}
	}

	if len(items) > 0 {
		user.DayTaskFlow(taskPoolId, max)
	}

	return items
}

// drawDayTaskAwards 领取每日任务奖励
func (user *LobbyUser) drawDayTaskAwards(typ uint8, id uint32) {
	today := common.GetTodayBeginStamp()
	week := common.GetThisWeekBeginStamp()

	stamp := int64(0)
	if typ == 1 || typ == 3 {
		stamp = today
	} else if typ == 2 {
		stamp = week
	}

	ret := uint8(0)
	util := db.PlayerInfoUtil(user.GetDBID())

	if user.canDrawDayTaskAwards(typ, id) {
		awards := common.GetDayTaskAwards(typ, id)
		user.storeMgr.GetAwards(awards, common.RS_DayTask, false, true)

		if typ == 3 {
			user.updateDayTaskItems(common.TaskItem_FinishTask, 1)
		}

		util.SetDayTaskDrawed(typ, stamp, id)
		ret = 1
	}

	dayActiveness := util.GetActiveness(1, today)
	weekActiveness := util.GetActiveness(2, week)

	user.RPC(iserver.ServerTypeClient, "DrawDayTaskAwardsRet", ret, typ, id, dayActiveness, weekActiveness)
	user.Info("drawDayTaskAwards, ret: ", ret, " typ: ", typ, " id: ", id, " dayActiveness: ", dayActiveness, " weekActiveness: ", weekActiveness)
}

/*-------------------------战友任务-----------------------------*/
// syncComradeTask 同步战友任务标记信息
func (user *LobbyUser) syncComradeTask() {
	typ := uint8(0)
	minLevel := uint32(common.GetTBSystemValue(common.System_RecruitInitLevel))
	util := db.PlayerInfoUtil(user.GetDBID())

	if util.HaveTeacherPupil() {
		typ = 1
	} else if user.GetLevel() >= minLevel {
		typ = 2
	}

	user.RPC(iserver.ServerTypeClient, "ComradeTaskRet", typ)
	user.Info("ComradeTaskRet, typ: ", typ)
}

// isComradeTask 是否为玩家当前正在进行的战友任务
func (user *LobbyUser) isComradeTask(taskPoolId uint32) bool {
	today := common.GetTodayBeginStamp()
	v, ok := GetSrvInst().openTasks.Load(common.TaskName_Comrade)
	if !ok {
		return false
	}

	tasks := v.(map[uint8][]uint32)
	round := db.PlayerInfoUtil(user.GetDBID()).GetComradeTaskRound(today)

	taskPoolIds, ok := tasks[2*round]
	if !ok {
		return false
	}

	for _, v := range taskPoolIds {
		if v == taskPoolId {
			return true
		}
	}

	return false
}

// getComradeTaskIds 获取指定任务池对应的战友任务id
func (user *LobbyUser) getComradeTaskIds(taskPoolId uint32) []uint32 {
	res := []uint32{}
	today := common.GetTodayBeginStamp()

	v, ok := GetSrvInst().openTasks.Load(common.TaskName_Comrade)
	if !ok {
		return res
	}

	tasks := v.(map[uint8][]uint32)
	round := db.PlayerInfoUtil(user.GetDBID()).GetComradeTaskRound(today)

	taskIds, ok := tasks[2*round-1]
	if !ok {
		return res
	}

	for _, taskId := range taskIds {
		task, ok := excel.GetComradeTask(uint64(taskId))
		if !ok {
			continue
		}

		if task.Taskpool == uint64(taskPoolId) {
			res = append(res, taskId)
		}
	}

	return res
}

// isComradeTaskItemFinished 战友任务项是否已经完成
func (user *LobbyUser) isComradeTaskItemFinished(taskId uint32) bool {
	today := common.GetTodayBeginStamp()
	cur := db.PlayerInfoUtil(user.GetDBID()).GetComradeTaskItemProgress(today, taskId)

	item, ok := excel.GetComradeTask(uint64(taskId))
	if !ok {
		return false
	}

	return cur >= uint32(item.Requirenum)
}

// isComradeTaskAllDrawed 本轮战友任务奖励是否已经全部领取
func (user *LobbyUser) isComradeTaskAllDrawed(round uint8) bool {
	today := common.GetTodayBeginStamp()
	util := db.PlayerInfoUtil(user.GetDBID())

	v, ok := GetSrvInst().openTasks.Load(common.TaskName_Comrade)
	if !ok {
		return false
	}

	tasks := v.(map[uint8][]uint32)
	taskIds, ok := tasks[2*round-1]
	if !ok {
		return false
	}

	for _, taskId := range taskIds {
		if !util.IsComradeTaskDrawed(today, taskId) {
			return false
		}
	}

	return true
}

// canIncreaseComradeTaskRound 是否需要增加战友任务轮次
func (user *LobbyUser) canIncreaseComradeTaskRound() bool {
	today := common.GetTodayBeginStamp()
	round := db.PlayerInfoUtil(user.GetDBID()).GetComradeTaskRound(today)

	if round > 1 {
		return false
	}

	if !user.isComradeTaskAllDrawed(round) {
		return false
	}

	return true
}

// updateComradeTaskItemProgress 更新战友任务项的进度
func (user *LobbyUser) updateComradeTaskItemProgress(taskId uint32, incr uint32) uint32 {
	util := db.PlayerInfoUtil(user.GetDBID())
	today := common.GetTodayBeginStamp()
	cur := util.GetComradeTaskItemProgress(today, taskId)

	if incr == 0 {
		return cur
	}

	item, ok := excel.GetComradeTask(uint64(taskId))
	if !ok {
		return cur
	}

	max := uint32(item.Requirenum)
	if cur >= max {
		return cur
	}

	cur += incr
	if cur >= max {
		cur = max
	}

	util.SetComradeTaskItemProgress(today, taskId, cur)

	return cur
}

// updateComradeTaskItems 更新玩家的战友任务项完成进度
func (user *LobbyUser) updateComradeTaskItems(comrades uint32, items ...uint32) {
	msg := &protoMsg.TaskUpdate{}
	items = append(items, common.TaskItem_FriendGame, 1)

	for i := 0; i < len(items); i += 2 {
		taskPoolId := items[i]
		if !user.isComradeTask(taskPoolId) {
			continue
		}

		var incr uint32

		if taskPoolId == common.TaskItem_RescueTeammate {
			incr = comrades
		} else {
			incr = items[i+1]
		}

		tmpitems := user.updateComradeTaskPool(taskPoolId, incr)
		if len(tmpitems) > 0 {
			msg.TaskItems = append(msg.TaskItems, tmpitems...)
		}
	}

	if len(msg.TaskItems) > 0 {
		user.RPC(iserver.ServerTypeClient, "ComradeTaskUpdateNotify", msg)
		user.Infof("ComradeTaskUpdateNotify: %+v\n", msg)
	}
}

// updateComradeTaskPool 更新玩家的战友任务池的任务进度
func (user *LobbyUser) updateComradeTaskPool(taskPoolId, incr uint32) []*protoMsg.TaskItem {
	if incr == 0 {
		return nil
	}

	var (
		items []*protoMsg.TaskItem
		max   uint32
	)

	for _, taskId := range user.getComradeTaskIds(taskPoolId) {
		if user.isComradeTaskItemFinished(taskId) {
			continue
		}

		cur := user.updateComradeTaskItemProgress(taskId, incr)
		items = append(items, &protoMsg.TaskItem{
			TaskId:      taskId,
			FinishedNum: cur,
		})

		if cur > max {
			max = cur
		}
	}

	if len(items) > 0 {
		user.ComradeTaskFlow(taskPoolId, max)
	}

	return items
}

// comradeTaskInfoNotify 通知客户端战友任务相关数据
func (user *LobbyUser) comradeTaskInfoNotify() {
	util := db.PlayerInfoUtil(user.GetDBID())
	if !util.HaveTeacherPupil() {
		return
	}

	notify := &protoMsg.ComradeTaskDetail{}
	today := common.GetTodayBeginStamp()

	record := util.GetComradeTaskRecord(today)
	round := util.GetComradeTaskRound(today)

	v, ok := GetSrvInst().openTasks.Load(common.TaskName_Comrade)
	if !ok {
		return
	}

	tasks := v.(map[uint8][]uint32)
	taskIds, ok := tasks[2*round-1]
	if !ok {
		return
	}

	for _, taskId := range taskIds {
		task := &protoMsg.TaskItem{
			TaskId: taskId,
		}

		if record != nil {
			for _, item := range record.ComradeTaskItems {
				if item.TaskId == taskId {
					task.FinishedNum = item.FinishedNum
					task.Awards = uint32(item.Awards)
					break
				}
			}
		}

		notify.ComradeTaskList = append(notify.ComradeTaskList, task)
	}

	sort.Sort(TaskItemSlice(notify.ComradeTaskList))

	user.RPC(iserver.ServerTypeClient, "ComradeTaskInfoNotify", notify)
	user.Infof("ComradeTaskInfoNotify: %+v\n", notify)
}

// canDrawComradeTaskAwards 玩家是否可以领取战友任务奖励
func (user *LobbyUser) canDrawComradeTaskAwards(taskId uint32) bool {
	today := common.GetTodayBeginStamp()
	util := db.PlayerInfoUtil(user.GetDBID())

	if util.IsComradeTaskDrawed(today, taskId) {
		return false
	}

	if !user.isComradeTaskItemFinished(taskId) {
		return false
	}

	return true
}

// drawComradeTaskAwards 领取战友任务奖励
func (user *LobbyUser) drawComradeTaskAwards(taskId uint32) {
	ret := uint8(1)
	today := common.GetTodayBeginStamp()
	util := db.PlayerInfoUtil(user.GetDBID())

	if user.canDrawComradeTaskAwards(taskId) {
		awards := common.GetComradeTaskAwards(taskId)
		user.storeMgr.GetAwards(awards, common.RS_ComradeTask, false, false)

		util.SetComradeTaskDrawed(today, taskId)
		ret = 0
	}

	user.RPC(iserver.ServerTypeClient, "DrawComradeTaskAwardsRet", ret, taskId)
	user.Info("drawComradeTaskAwards, ret: ", ret, " taskId: ", taskId)

	if ret == 0 {
		if user.canIncreaseComradeTaskRound() {
			util.IncreaseComradeTaskRound()
			user.comradeTaskInfoNotify()
		}
	}
}

package main

import (
	"common"
	"db"
	"excel"
	"protoMsg"
	"time"
	"zeus/iserver"
)

/*
说明：挑战任务复用了特训任务相关消息，概念上挑战任务的赛季对应于特训任务的任务周（week）；
	 挑战任务的任务组对应于特训任务的任务日（day）
*/

// 挑战任务
type challengeTask struct {
	task
}

// init 初始化任务数据
func (t *challengeTask) init(user *LobbyUser, name string) {
	t.user = user
	t.realPtr = t
	t.name = name
	t.awardReason = common.RS_ChallengeTask
}

// getTaskStage 获取任务阶段标识
func (t *challengeTask) getTaskStage() string {
	return common.IntToString(common.GetRankSeason())
}

// getEnabledGroups 获取激活的任务组
func (t *challengeTask) getEnabledGroups() []uint8 {
	res := []uint8{}
	util := t.getTaskUtil()

	for i, enable := range util.GetGroupEnables() {
		if enable {
			res = append(res, uint8(i+1))
		}
	}

	return res
}

// getTaskType 获取任务项的类型 0表示普通任务 1表示精英任务
func (t *challengeTask) getTaskType(taskId uint32) uint8 {
	return 0
}

// getRequireNum 获取任务项需要达到的最大数量
func (t *challengeTask) getRequireNum(taskId uint32) uint32 {
	item, ok := excel.GetChallengeTask(uint64(taskId))
	if !ok {
		return 0
	}

	return uint32(item.Requirenum)
}

// getTaskItemAwards 获取任务项奖励数据
func (t *challengeTask) getTaskItemAwards(taskId uint32) map[uint32]uint32 {
	item, ok := excel.GetChallengeTask(uint64(taskId))
	if !ok {
		return make(map[uint32]uint32)
	}

	return common.StringToMapUint32(item.Awards, "|", ";")
}

// getTaskAwards 获取任务奖励数据
func (t *challengeTask) getTaskAwards(id uint32, typ uint8) map[uint32]uint32 {
	item, ok := excel.GetSeasonGrade(uint64(id))
	if !ok {
		return make(map[uint32]uint32)
	}

	switch typ {
	case 0:
		return common.StringToMapUint32(item.NormalAwards, "|", ";")
	case 1:
		return common.StringToMapUint32(item.EliteAwards, "|", ";")
	}

	return make(map[uint32]uint32)
}

// syncTaskDetail 向客户端同步任务详情
func (t *challengeTask) syncTaskDetail() {
	v, ok := GetSrvInst().openTasks.Load(t.name)
	if !ok {
		return
	}

	taskM := v.(map[uint8][]uint32)

	v, ok = GetSrvInst().taskAwards.Load(t.name)
	if !ok {
		return
	}

	awardM := v.([]uint32)

	util := t.getTaskUtil()
	grade := util.GetSeasonGrade()
	medals := util.GetExpMedals()

	list := util.GetTaskList()
	groupEnables := util.GetGroupEnables()
	normalAwards := util.GetTaskAwards()
	eliteAwards := util.GetEliteAwards()

	now := time.Now().Unix()
	_, end := common.GetRankSeasonTimeStamp()

	var leftTime uint32
	if end >= now {
		leftTime = uint32(end - now)
	}

	notify := &protoMsg.ChallengeTaskDetail{
		Season:       uint32(common.GetRankSeason()),
		LeftTime:     leftTime,
		Grade:        grade,
		Medals:       medals,
		EliteEnable:  util.IsEliteEnable(),
		NormalAwards: &protoMsg.AwardList{},
		EliteAwards:  &protoMsg.AwardList{},
	}

	for i := 0; i < len(taskM)/2; i++ {
		notify.List = append(notify.List, &protoMsg.CommonTaskList{})
	}

	for k, s := range taskM {
		if k%2 == 0 {
			continue
		}

		index := int(k / 2)
		notify.List[index].Id = uint32((k + 1) / 2)

		if groupEnables != nil {
			notify.List[index].Enable = groupEnables[index]
		}

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

			notify.List[index].TaskItems = append(notify.List[index].TaskItems, item)
		}
	}

	for _, v := range awardM {
		info := &protoMsg.AwardInfo{
			Id: v,
		}

		for _, j := range normalAwards {
			if j.Id == v {
				info.State = uint32(j.State)
			}
		}

		notify.NormalAwards.Infos = append(notify.NormalAwards.Infos, info)
	}

	for _, v := range awardM {
		info := &protoMsg.AwardInfo{
			Id: v,
		}

		for _, j := range eliteAwards {
			if j.Id == v {
				info.State = uint32(j.State)
			}
		}

		notify.EliteAwards.Infos = append(notify.EliteAwards.Infos, info)
	}

	t.user.RPC(iserver.ServerTypeClient, "ChallengeTaskNotify", notify)
	t.user.Infof("ChallengeTaskNotify: %+v\n", notify)

	t.syncInitTaskAwards()
	t.checkForEnableTaskGroup()
}

// getTaskIdsByPool 根据任务池获取对应的开放任务
func (t *challengeTask) getTaskIdsByPool(groupId uint8, taskPoolId uint32) []uint32 {
	res := []uint32{}

	v, ok := GetSrvInst().openTasks.Load(t.name)
	if !ok {
		return res
	}

	taskM := v.(map[uint8][]uint32)
	for _, v := range taskM[2*groupId-1] {
		task, ok := excel.GetChallengeTask(uint64(v))
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
func (t *challengeTask) isTaskItemEnable(taskId uint32) bool {
	return true
}

// enableTaskItem 激活精英任务项
func (t *challengeTask) enableTaskItem(taskId uint32) {

}

// updateSeasonGrade 更新赛季战阶及奖励
func (t *challengeTask) updateSeasonGrade(num uint32) {
	if num == 0 {
		return
	}

	util := t.getTaskUtil()
	grade1 := util.GetSeasonGrade()
	medals1 := util.GetExpMedals()

	grade2, medals2 := common.CalSeasonGrade(grade1, medals1, num)
	util.SetSeasonGrade(grade2)
	util.SetExpMedals(medals2)

	normalAwards := &protoMsg.AwardList{}
	eliteAwards := &protoMsg.AwardList{}

	for g := grade1 + 1; g <= grade2; g++ {
		util.SetTaskAwardsState(g, db.AwardsStateDrawable)
		normalAwards.Infos = append(normalAwards.Infos, &protoMsg.AwardInfo{
			Id:    g,
			State: db.AwardsStateDrawable,
		})

		if util.IsEliteEnable() {
			util.SetEliteAwardsState(g, db.AwardsStateDrawable)
			eliteAwards.Infos = append(eliteAwards.Infos, &protoMsg.AwardInfo{
				Id:    g,
				State: db.AwardsStateDrawable,
			})
		}
	}

	t.user.RPC(iserver.ServerTypeClient, "UpdateTaskAwardsState", medals2, normalAwards, t.name, grade2, eliteAwards)
	t.user.SeasonExpFlow(num)

	if grade2 > grade1 {
		t.checkForEnableTaskGroup()
	}
}

// getEnableTaskItemGoods 获取用于激活任务项的道具
func (t *challengeTask) getEnableTaskItemGoods(taskId uint32) map[uint32]uint32 {
	return nil
}

// replaceTaskItemAwards 补领任务项奖励
func (t *challengeTask) replaceTaskItemAwards(groupId uint8, taskId uint32) {

}

// enableSeasonElite 激活赛季精英资格
func (t *challengeTask) enableSeasonElite() {
	util := t.getTaskUtil()
	util.SetEliteEnable()

	t.syncInitTaskAwards()
	t.syncDrawableEliteAwards()
	t.checkForEnableTaskGroup()

	t.user.RPC(iserver.ServerTypeClient, "EliteEnableNotify", t.name)
}

// syncInitTaskAwards 同步初始任务奖励
func (t *challengeTask) syncInitTaskAwards() {
	util := t.getTaskUtil()
	grade := util.GetSeasonGrade()
	medals := util.GetExpMedals()

	normalAwards := &protoMsg.AwardList{}
	eliteAwards := &protoMsg.AwardList{}

	if len(t.getTaskAwards(1, 0)) > 0 && util.GetTaskAwardsState(1) == db.AwardsStateNotDrawable {
		util.SetTaskAwardsState(1, db.AwardsStateDrawable)
		normalAwards.Infos = append(normalAwards.Infos, &protoMsg.AwardInfo{
			Id:    1,
			State: db.AwardsStateDrawable,
		})
	}

	if util.IsEliteEnable() && len(t.getTaskAwards(1, 1)) > 0 && util.GetEliteAwardsState(1) == db.AwardsStateNotDrawable {
		util.SetEliteAwardsState(1, db.AwardsStateDrawable)
		eliteAwards.Infos = append(eliteAwards.Infos, &protoMsg.AwardInfo{
			Id:    1,
			State: db.AwardsStateDrawable,
		})
	}

	if len(normalAwards.Infos) > 0 || len(eliteAwards.Infos) > 0 {
		t.user.RPC(iserver.ServerTypeClient, "UpdateTaskAwardsState", medals, normalAwards, t.name, grade, eliteAwards)
	}
}

// syncDrawableEliteAwards 激活精英权限时同步可领取的精英奖励
func (t *challengeTask) syncDrawableEliteAwards() {
	util := t.getTaskUtil()
	grade := util.GetSeasonGrade()
	medals := util.GetExpMedals()

	normalAwards := &protoMsg.AwardList{}
	eliteAwards := &protoMsg.AwardList{}

	for _, info := range util.GetTaskAwards() {
		if info.Id == 1 {
			continue
		}

		util.SetEliteAwardsState(info.Id, db.AwardsStateDrawable)
		eliteAwards.Infos = append(eliteAwards.Infos, &protoMsg.AwardInfo{
			Id:    info.Id,
			State: db.AwardsStateDrawable,
		})
	}

	if len(eliteAwards.Infos) > 0 {
		t.user.RPC(iserver.ServerTypeClient, "UpdateTaskAwardsState", medals, normalAwards, t.name, grade, eliteAwards)
	}
}

// isTaskGroupFinished 是否完成任务组中的所有任务
func (t *challengeTask) isTaskGroupFinished(groupId uint8) bool {
	v, ok := GetSrvInst().openTasks.Load(t.name)
	if !ok {
		return false
	}

	taskM := v.(map[uint8][]uint32)
	k := 2*groupId - 1

	for _, v := range taskM[k] {
		if !t.isTaskItemFinished(groupId, v) {
			return false
		}
	}

	return true
}

// checkForEnableTaskGroup 检测并激活任务组
func (t *challengeTask) checkForEnableTaskGroup() {
	util := t.getTaskUtil()
	groups := t.getEnabledGroups()

	v, ok := GetSrvInst().groupToUniqueId.Load(t.name)
	if !ok {
		return
	}

	uniqueM := v.(map[uint8]uint32)
	for groupId, uniqueid := range uniqueM {
		var enable bool
		for _, v := range groups {
			if v == groupId {
				enable = true
				break
			}
		}

		if enable {
			continue
		}

		data, ok := excel.GetChallenge(uint64(uniqueid))
		if !ok {
			continue
		}

		switch data.Unlock {
		case 1:
			{
				grade := util.GetSeasonGrade()
				if grade >= uint32(data.UnlockValue) {
					enable = true
				}
			}
		case 2:
			{
				if t.isTaskGroupFinished(uint8(data.UnlockValue)) {
					enable = true
				}
			}
		case 3:
			{
				if util.IsEliteEnable() {
					enable = true
				}
			}
		case 4:
			{
				day := int64(24 * 60 * 60)
				var pastTime int64

				start, _ := common.GetRankSeasonTimeStamp()
				if start != 0 {
					pastTime = time.Now().Unix() - start
				}

				if pastTime >= int64(data.UnlockValue)*day {
					enable = true
				}
			}
		case 5:
			{
				goodsUtil := db.PlayerGoodsUtil(t.user.GetDBID())
				if goodsUtil.IsGoodsEnough(uint32(data.UnlockValue), 1) {
					enable = true
				}
			}
		}

		if enable {
			util.SetGroupEnable(groupId)
			t.user.RPC(iserver.ServerTypeClient, "UnlockTaskGroupNotify", t.name, uint32(groupId))
		}
	}
}

// canDrawTaskAwards 能否领取任务奖励
func (t *challengeTask) canDrawTaskAwards(id uint32, typ uint8) bool {
	util := t.getTaskUtil()

	switch typ {
	case 0:
		return util.GetTaskAwardsState(id) == db.AwardsStateDrawable
	case 1:
		return util.GetEliteAwardsState(id) == db.AwardsStateDrawable
	}

	return false
}

// drawTaskAwards 领取任务奖励
func (t *challengeTask) drawTaskAwards(id uint32, typ uint8) {
	util := t.getTaskUtil()
	info := &protoMsg.AwardInfo{
		Id: id,
	}

	if t.canDrawTaskAwards(id, typ) {
		switch typ {
		case 0:
			util.SetTaskAwardsState(id, db.AwardsStateDrawed)
		case 1:
			util.SetEliteAwardsState(id, db.AwardsStateDrawed)
		}

		awards := t.getTaskAwards(id, typ)
		t.user.storeMgr.GetAwards(awards, uint32(t.awardReason), false, false)
	}

	switch typ {
	case 0:
		info.State = uint32(util.GetTaskAwardsState(id))
	case 1:
		info.State = uint32(util.GetEliteAwardsState(id))
	}

	t.user.RPC(iserver.ServerTypeClient, "DrawTaskAwardsRet", info, t.name, typ)
	t.user.Info("drawTaskAwards, info: ", *info, " name: ", t.name, " typ: ", typ)
}

// checkSeasonMaxGrade 检测赛季战阶是否达最大
func (t *challengeTask) checkSeasonMaxGrade() bool {
	util := t.getTaskUtil()
	grade := util.GetSeasonGrade()

	gradeMax := common.GetMaxSeasonGrade()
	if grade >= gradeMax {
		return true
	}
	return false
}

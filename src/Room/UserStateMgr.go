package main

import (
	"common"
	"excel"
	"time"
	"zeus/iserver"
	"zeus/linmath"
)

// UserStateMgr 玩家状态管理
type UserStateMgr struct {
	user *RoomUser
	//state uint64 // 玩家状态

	attactID     uint64 // 攻击者id
	downStamp    int64  // 玩家被击倒时间戳
	downEndStamp int64  // 击倒结束时间戳
	injuredtype  uint32 // 被击倒收到伤害类型
	downAttacker uint64 // 被击倒的攻击者
	isHeadShot   bool   // 被击倒是否因为暴击

	isRescue       bool            // 是否救援
	RescuePos      linmath.Vector3 // 救援时坐标
	rescueStamp    int64           // 玩家救援时间
	rescueTarget   *RoomUser       // 救援目标玩家
	preRescueState uint8           // 救援前状态

	isBerescue     bool            // 是否被救援
	BeRescuePos    linmath.Vector3 // 被救援时坐标
	beRescueStamp  int64           // 玩家被救援时间
	beRescueTarget *RoomUser       // 被救援目标玩家

	isSettle bool //  玩家是否结算
}

// NewUserStateMgr 初始化玩家状态管理器
func NewUserStateMgr(user *RoomUser) *UserStateMgr {

	stateMgr := &UserStateMgr{}

	stateMgr.user = user
	stateMgr.SetState(RoomPlayerBaseState_Stand) // 初始化玩家为站立状态
	stateMgr.isSettle = false
	stateMgr.isRescue = false
	stateMgr.isBerescue = false

	return stateMgr
}

// GetState 获得玩家状态
func (stateMgr *UserStateMgr) GetState() uint8 {

	if stateMgr.user == nil {
		stateMgr.user.Info("GetState failed, user is nil")
		return 0
	}

	return stateMgr.user.GetState()
}

// SetState 设置玩家状态
func (stateMgr *UserStateMgr) SetState(state uint8) {
	stateMgr.user.SetBaseState(state)
}

// printValue 打印玩家状态信息
func (stateMgr *UserStateMgr) printValue() {
	stateMgr.user.Infof("UserState userdbid(%d), userid(%d), state(%d)", stateMgr.user.GetDBID(), stateMgr.user.GetID(), stateMgr.GetState())
}

// Loop 定时处理
func (stateMgr *UserStateMgr) Loop() {

	if stateMgr.GetState() == RoomPlayerBaseState_Dead {
		return
	} else if stateMgr.isRescue {
		stateMgr.LoopRescue()
	} else if stateMgr.isBerescue {
		stateMgr.LoopBeRescue()
	} else if stateMgr.GetState() == RoomPlayerBaseState_WillDie {
		stateMgr.LoopDown(false)
	}
}

// 设置玩家濒死状态
func (stateMgr *UserStateMgr) setDownState(attackID uint64, injuredType uint32, isHeadShot bool) bool {
	curState := stateMgr.GetState()
	if curState == RoomPlayerBaseState_WillDie || curState == RoomPlayerBaseState_Dead || curState == RoomPlayerBaseState_Watch {
		return false
	}

	space := stateMgr.user.GetSpace().(*Scene)
	if space == nil || space.teamMgr.isExistTeammate(stateMgr.user) == false {
		return false
	}

	stateMgr.SetState(RoomPlayerBaseState_WillDie)
	stateMgr.user.SetActionState(0)
	stateMgr.user.AdviceNotify(common.NotifyCommon, 20) //通知玩家（您已被击倒，请等待救援）

	stateMgr.attactID = attackID
	stateMgr.injuredtype = injuredType
	stateMgr.downAttacker = attackID
	stateMgr.isHeadShot = isHeadShot

	stateMgr.downStamp = time.Now().Unix()
	downTime, _ := excel.GetSystem(9)

	stateMgr.downEndStamp = stateMgr.downStamp + int64(downTime.Value)

	// 通知玩家濒死进度
	stateMgr.user.SyncProgressBar(USERDOWNTYPE, uint64(stateMgr.downEndStamp), uint64(downTime.Value), uint64(stateMgr.downEndStamp)-uint64(time.Now().Unix()))

	// 通知队友自己处于濒死状态
	space.teamMgr.SyncDownEndTime(stateMgr.user, uint64(stateMgr.downEndStamp), uint64(stateMgr.downEndStamp)-uint64(time.Now().Unix()))

	return true
}

// 修改玩家濒死状态时间
func (stateMgr *UserStateMgr) amendDownTime(attackID uint64, injuredType uint32, amendTime int64) {
	curState := stateMgr.GetState()
	if curState != RoomPlayerBaseState_WillDie {
		return
	}

	space := stateMgr.user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	stateMgr.attactID = attackID
	stateMgr.injuredtype = injuredType

	stateMgr.downEndStamp = stateMgr.downEndStamp + amendTime

	curTime := time.Now().Unix()
	if stateMgr.downEndStamp <= curTime { // 玩家濒死状态死亡

		if stateMgr.beRescueTarget != nil && stateMgr.isBerescue {
			stateMgr.beRescueTarget.stateMgr.clearBeRescueState()
		}

		stateMgr.LoopDown(true)

	} else {

		// 通知玩家濒死进度
		downTime, _ := excel.GetSystem(9)
		stateMgr.user.SyncProgressBar(USERDOWNTYPE, uint64(stateMgr.downEndStamp), uint64(downTime.Value), uint64(stateMgr.downEndStamp)-uint64(time.Now().Unix()))

		// 通知队友自己处于濒死状态
		space.teamMgr.SyncDownEndTime(stateMgr.user, uint64(stateMgr.downEndStamp), uint64(stateMgr.downEndStamp)-uint64(time.Now().Unix()))

	}
}

// LoopDown 定时刷新被击倒状态
func (stateMgr *UserStateMgr) LoopDown(isBeattack bool) {
	if stateMgr.downEndStamp <= time.Now().Unix() {

		space := stateMgr.user.GetSpace().(*Scene)
		if space == nil {
			return
		}
		//stateMgr.user.Info("濒死时间到", stateMgr.GetState(), isBeattack, stateMgr.downEndStamp, time.Now().Unix())
		// 打断进度条
		stateMgr.user.BreakProgressBar()

		if stateMgr.user.SetWatchBattle() == false {
			//stateMgr.user.Info("进入死亡状态")

			stateMgr.user.DisposeDeath(stateMgr.injuredtype, stateMgr.attactID, false)
		} else {

			stateMgr.user.DisposeDeath(stateMgr.injuredtype, stateMgr.attactID, true)

		}

		if !isBeattack {
			attacker, ok := space.GetEntity(stateMgr.attactID).(*RoomUser)
			if ok {
				if !(space.teamMgr.IsInOneTeam(stateMgr.user.GetDBID(), attacker.GetDBID()) || stateMgr.user.GetID() == stateMgr.attactID) {
					attacker.DisposeIncrKillNum()
					if stateMgr.isHeadShot && stateMgr.attactID == stateMgr.downAttacker {
						attacker.IncrHeadShotNum()
					}
				}
			}

			space.BroadDieNotify(stateMgr.attactID, stateMgr.user.GetID(), false, InjuredInfo{injuredType: losthp, isHeadshot: false})
		}

		if stateMgr.user.SetWatchBattle() == false {
			stateMgr.user.DisposeSettle()
		}

	}
}

// setRescue 设置玩家救援状态
func (stateMgr *UserStateMgr) setRescue(targetUser *RoomUser) {

	//stateMgr.user.Infof("设置玩家救援状态开始 %d, %d", stateMgr.user.GetDBID(), stateMgr.GetState())

	if stateMgr.GetState() == RoomPlayerBaseState_WillDie || stateMgr.GetState() == RoomPlayerBaseState_Dead || stateMgr.isRescue || stateMgr.isBerescue {
		stateMgr.user.RPC(iserver.ServerTypeClient, "RescueFail")
		return
	}

	if targetUser == nil || targetUser.stateMgr.isBerescue {
		stateMgr.user.RPC(iserver.ServerTypeClient, "RescueFail")
		return
	}

	if targetUser.stateMgr.GetState() != RoomPlayerBaseState_WillDie {
		stateMgr.user.RPC(iserver.ServerTypeClient, "RescueFail")
		return
	}

	stateMgr.preRescueState = stateMgr.GetState()
	stateMgr.isRescue = true
	stateMgr.RescuePos = stateMgr.user.GetPos()
	stateMgr.rescueStamp = time.Now().Unix()
	stateMgr.rescueTarget = targetUser

	// 设置目标玩家为被救援状态
	stateMgr.rescueTarget.stateMgr.setBeRescue(stateMgr.user)

	rescuetime := stateMgr.user.GetRescueTime()
	if rescuetime == 0 {
		return
	}

	stateMgr.user.SyncProgressBar(USERRESCUETYPE, uint64(stateMgr.rescueStamp+int64(rescuetime)), uint64(rescuetime), uint64(rescuetime))

	//stateMgr.user.Infof("设置玩家救援状态结束 %d", stateMgr.user.GetDBID())
}

// LoopRescue 定时刷新救援状态
func (stateMgr *UserStateMgr) LoopRescue() {
	rescuetime := stateMgr.user.GetRescueTime()
	if rescuetime == 0 {
		return
	}

	if stateMgr.rescueStamp+int64(rescuetime) <= time.Now().Unix() {
		// 处理玩家救援
		//stateMgr.user.Debug("定时刷新救援状态")
		stateMgr.DisposeRescue(true)

	}
}

// DisposeRescue 处理玩家救援(控制被救起玩家的逻辑)
func (stateMgr *UserStateMgr) DisposeRescue(isScuess bool) {

	//stateMgr.user.Infof("处理玩家救援(控制被救起玩家的逻辑) userdbid(%d), isScuess(%d)", stateMgr.user.GetDBID(), isScuess)

	if stateMgr.isRescue == false {
		return
	}

	// 打断进度条
	stateMgr.user.BreakProgressBar()

	stateMgr.isRescue = false
	stateMgr.rescueStamp = 0

	if stateMgr.rescueTarget == nil {
		return
	}

	stateMgr.rescueTarget.stateMgr.DisposeBeRescue(isScuess)
	stateMgr.rescueTarget = nil
	stateMgr.preRescueState = 0

	if isScuess {
		stateMgr.user.MedalNotifyRpc(5)
	}
	//stateMgr.user.Infof("处理玩家救援(控制被救起玩家的逻辑) userdbid(%d), isScuess(%d)", stateMgr.user.GetDBID(), isScuess)
}

// clearBeRescueState 清空救援状态
func (stateMgr *UserStateMgr) clearBeRescueState() {

	if stateMgr.isRescue == false {
		return
	}

	// 打断进度条
	stateMgr.user.BreakProgressBar()

	stateMgr.isRescue = false
	stateMgr.rescueStamp = 0

	stateMgr.rescueTarget = nil
	stateMgr.preRescueState = 0

}

// setBeRescue 设置玩家被救援状态
func (stateMgr *UserStateMgr) setBeRescue(targetUser *RoomUser) {
	//stateMgr.user.Infof("设置玩家被救援状态开始 userdbid(%d)", stateMgr.user.GetDBID())
	//
	//stateMgr.SetState(BeRescue)
	stateMgr.isBerescue = true
	stateMgr.BeRescuePos = stateMgr.user.GetPos()

	stateMgr.beRescueStamp = time.Now().Unix()
	stateMgr.beRescueTarget = targetUser

	rescuetime := targetUser.GetRescueTime()
	if rescuetime == 0 {
		return
	}
	stateMgr.user.SyncProgressBar(USERBERESCUETYPE, uint64(stateMgr.beRescueStamp+int64(rescuetime)), uint64(rescuetime), uint64(rescuetime))

	// 通知队友自己暂停濒死进度
	space := stateMgr.user.GetSpace().(*Scene)
	if space != nil {
		space.teamMgr.SyncDownEndTime(stateMgr.user, uint64(stateMgr.downEndStamp), uint64(stateMgr.downEndStamp)-uint64(time.Now().Unix()))
	}

}

// LoopBeRescue 定时刷新被救援状态
func (stateMgr *UserStateMgr) LoopBeRescue() {
	system, ok := excel.GetSystem(10)
	if !ok {
		return
	}
	if stateMgr.rescueStamp+int64(system.Value) <= time.Now().Unix() {
		//stateMgr.user.Infoln("玩家在规定时间内未被救起 userdbid(%d), userid(%d)", stateMgr.user.GetDBID(), stateMgr.user.GetID())
	}
}

// DisposeBeRescue 处理玩家被救援
func (stateMgr *UserStateMgr) DisposeBeRescue(isScuess bool) {

	//stateMgr.user.Infof("处理玩家被救援======================开始 %d", isScuess)

	if stateMgr.isBerescue == false {
		//if stateMgr.GetState() != BeRescue {
		return
	}

	stateMgr.isBerescue = false

	// 打断进度条
	stateMgr.user.BreakProgressBar()

	if isScuess { // 救援成功

		stateMgr.SetState(RoomPlayerBaseState_Down)

		hpValue := stateMgr.beRescueTarget.GetRescueHP()
		stateMgr.user.SetHP(hpValue)
		stateMgr.user.sumData.IncrReviveNum()           //成功复活次数自增
		stateMgr.beRescueTarget.sumData.IncrRescueNum() //成功救援次数自增

		space := stateMgr.user.GetSpace().(*Scene)
		if space.teamMgr.isComrade(stateMgr.user.GetDBID(), stateMgr.beRescueTarget.GetDBID()) {
			stateMgr.beRescueTarget.sumData.IncrRescueComradeNum() //成功救援战友次数自增
		}

		stateMgr.downStamp = 0
		stateMgr.downEndStamp = 0
		stateMgr.user.shotGunToWillDie = 0 //重置玩家被散弹枪击倒至濒死状态的时间

		//stateMgr.user.Debug("处理玩家被救援======================成功")

	} else { // 救援失败

		pauseTime := time.Now().Unix() - stateMgr.beRescueStamp
		if pauseTime < 0 {
			pauseTime = 0
		}

		stateMgr.amendDownTime(stateMgr.attactID, stateMgr.injuredtype, pauseTime)

		//stateMgr.user.Debug("处理玩家被救援======================失败")
	}

	stateMgr.beRescueStamp = 0

	stateMgr.beRescueTarget = nil

	//stateMgr.user.Infof("处理玩家被救援======================结束 %d", isScuess)

}

// BreakBerescue 打断被救援状态
func (stateMgr *UserStateMgr) BreakBerescue() {

	// stateMgr.user.Infof("打断被救援状态 开始 userdbid(%d)", stateMgr.user.GetDBID())

	if stateMgr.isBerescue == false {
		//if stateMgr.GetState() != BeRescue {
		return
	}

	if stateMgr.beRescueTarget == nil {
		return
	}

	stateMgr.beRescueTarget.stateMgr.DisposeRescue(false)

	// stateMgr.user.Infof("打断被救援状态 结束 userdbid(%d)", stateMgr.user.GetDBID())

}

// BreakRescue 打断救援
func (stateMgr *UserStateMgr) BreakRescue() {
	if stateMgr.isRescue == false {
		return
	}

	stateMgr.DisposeRescue(false)
}

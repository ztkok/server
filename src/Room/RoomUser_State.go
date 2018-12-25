package main

import (
	"time"
	"zeus/linmath"
	"zeus/space"
)

// StateValidate 状态快照验证，当客户端上发状态快照时，服务器在此验证
// 如果验证不通过，则当前快照不会被广播给其它的客户端
// 可以使用的矫正函数，框架层提供 SetPos 与 SetRota
// 其它对快照的矫正函数 ，都在这个文件中实现
func (user *RoomUser) StateValidate(oldEntityState, newEntityState space.IEntityState) bool {

	user.stateSyncTime = user.GetSpace().GetTimeStamp()
	oldState := oldEntityState.(*RoomPlayerState)
	newState := newEntityState.(*RoomPlayerState)

	// log.Error(user.GetID(), "接收到状态", newState.BaseState, "原先状态", oldState.BaseState)

	//死亡不再接收状态快照
	if user.GetState() == RoomPlayerBaseState_Dead {
		//log.Info(user.GetID(), " 死亡不再改变状态快照")
		return true
	}

	// just validate new state
	//do some thing
	if !user.speedValidate(oldState, newState) {
		return false
	}

	if !user.checkStateValidate(oldState, newState) {
		return false
	}

	if !user.checkPosValidate(oldState, newState) {
		return false
	}

	user.checkRotaValidate(newState)
	return true
}

func (user *RoomUser) StateChange(oldEntityState, newEntityState space.IEntityState) {

	oldState := oldEntityState.(*RoomPlayerState)
	newState := newEntityState.(*RoomPlayerState)

	if user.first && (newState.BaseState == RoomPlayerBaseState_Swim || newState.BaseState == RoomPlayerBaseState_Stand) {
		user.sumData.isGround = true //开始统计跑动距离
		user.sumData.landtime = time.Now().Unix()
		user.sumData.tmprunDistance = user.GetPos()

		user.initPackItems()
		user.setAiPos(user.sumData.tmprunDistance) // 设置ai位置

		user.tlogBattleFlow(0, 0, 0, 0, 0, 0) // tlog战场流水表
		user.tlogSecGameStartFlow()           // (安全tlog)游戏开始流水表
		user.Debug("开始统计跑动距离runDistance！")
		user.first = false
	}

	if user.sumData.tmpswimDistance.IsEqual(linmath.Vector3_Invalid()) && newState.BaseState == RoomPlayerBaseState_Swim {
		user.sumData.tmpswimDistance = user.GetPos()
	}

	if !user.sumData.tmpswimDistance.IsEqual(linmath.Vector3_Invalid()) && newState.BaseState != RoomPlayerBaseState_Swim {
		user.sumData.SwimDistance()
		user.sumData.tmpswimDistance = linmath.Vector3_Invalid()
	}

	if oldState.BaseState != newState.BaseState {
		user.onBaseStateChange(oldState, newState)
	}

	if oldState.ActionState != newState.ActionState {
		user.onActionStateChange(oldState, newState)
	}
}

func (user *RoomUser) onActionStateChange(oldState, newState *RoomPlayerState) {
	//log.Debug("状态改变 oldbaseState:", oldState.BaseState, " oldActionState:", oldState.ActionState, " newBaseState:", newState.BaseState, "newActionState: ", newState.ActionState)
	user.BreakEffect(false)
	if user.stateMgr.isRescue {
		if newState.ActionState == Melee {
			user.stateMgr.DisposeRescue(false)
		}
	}
}

func (user *RoomUser) onBaseStateChange(oldState, newState *RoomPlayerState) {
	//log.Info("oldState: ", oldState.BaseState, " newState: ", newState.BaseState)
	if oldState.BaseState != newState.BaseState {
		if oldState.BaseState == RoomPlayerBaseState_Parachute {
			user.UpdateCell()
			user.StateM.hasinit = true
			user.parachuteCtrl = nil // 已落地，移除跳伞控制
		} else if newState.BaseState == RoomPlayerBaseState_Parachute {
			if user.parachuteCtrl != nil {
				user.parachuteCtrl.StartParachute()
			}
		} else if newState.BaseState == RoomPlayerBaseState_Jump {
			user.BreakEffect(true)
		} else if newState.BaseState == RoomPlayerBaseState_Down || oldState.BaseState == RoomPlayerBaseState_Down {
			user.BreakEffect(false)
		}
	}

	if oldState.BaseState == RoomPlayerBaseState_Glide && newState.BaseState != RoomPlayerBaseState_Glide && newState.BaseState != RoomPlayerBaseState_Parachute {
		user.UpdateCell()
		user.StateM.hasinit = true
	}

	if user.stateMgr.isRescue {
		user.stateMgr.DisposeRescue(false)
		//log.Info("打断救援", user.GetDBID())
	}

	if oldState.BaseState != RoomPlayerBaseState_Fall && newState.BaseState == RoomPlayerBaseState_Fall {
		//记录坠落开始坐标
		user.lastfallpos = newState.GetPos()
		//log.Debug("记录坠落开始坐标:", user.lastfallpos)
	}

	if oldState.BaseState == RoomPlayerBaseState_Fall && newState.BaseState != RoomPlayerBaseState_Fall {
		user.doFallDam()
	}
}

// GetState 获取当前的状态快照
func (user *RoomUser) GetPlayerState() *RoomPlayerState {
	return user.GetStates().GetLastState().(*RoomPlayerState)
}

// SetActionState 设置actionState的值
func (user *RoomUser) SetActionState(as byte) {
	state := user.GetPlayerState()
	state.ActionState = as
	state.SetModify(true)
}

// GetActionState 获取ActionState
func (user *RoomUser) GetActionState() byte {
	state := user.GetPlayerState()
	return state.ActionState
}

// SetBaseState 设置 BaseState的值
func (user *RoomUser) SetBaseState(bs byte) {
	state := user.GetPlayerState()
	state.BaseState = bs
	state.SetModify(true)
}

func (user *RoomUser) GetBaseState() byte {
	state := user.GetPlayerState()
	return state.BaseState
}

func (user *RoomUser) SetSpeed(bs linmath.Vector3) {
	state := user.GetPlayerState()
	state.Speed = bs
	state.SetModify(true)
}

func (user *RoomUser) GetSpeed() linmath.Vector3 {
	state := user.GetPlayerState()
	return state.Speed
}

// GetHistoryState 获取历史的状态快照
func (user *RoomUser) GetHistoryUserState(timeStamp uint32) *RoomPlayerState {

	is := user.GetStates().GetHistoryState(timeStamp)
	if is == nil {
		return nil
	}

	return is.(*RoomPlayerState)
}

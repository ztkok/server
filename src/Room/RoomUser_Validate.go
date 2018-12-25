package main

import (
	"common"
	"math"
	"time"
	"zeus/linmath"
)

// speedValidate 速度校验
func (user *RoomUser) speedValidate(old, new *RoomPlayerState) bool {
	if !user.gm.isValidate {
		return true
	}

	var limitSpeed float32
	limitSpeed = math.MaxFloat32

	if old.BaseState != new.BaseState {
		return true
	}

	switch new.BaseState {
	case RoomPlayerBaseState_LeaveMap, RoomPlayerBaseState_LoadingMap, RoomPlayerBaseState_Inplane, RoomPlayerBaseState_Glide, RoomPlayerBaseState_Parachute:
		return true
	case RoomPlayerBaseState_Dead, RoomPlayerBaseState_Watch:
		return true
	case RoomPlayerBaseState_Stand:
		limitSpeed = user.rolebase.Mvspeedlimit
	case RoomPlayerBaseState_Down:
		limitSpeed = user.rolebase.Crawlspeedlimit
	case RoomPlayerBaseState_Ride:
		limitSpeed = user.rolebase.Vehiclespeedlimit
	case RoomPlayerBaseState_Swim:
		limitSpeed = user.rolebase.Swimspeedlimit
	case RoomPlayerBaseState_WillDie:
		limitSpeed = user.rolebase.Willdiespeedlimit
	case RoomPlayerBaseState_Crouch:
		limitSpeed = user.rolebase.Crouchspeedlimit
	case RoomPlayerBaseState_Fall, RoomPlayerBaseState_Jump:
		//坠落,跳起限制水平方向坐标变化
		limitSpeed = user.rolebase.Mvspeedlimit
		//		oldvec2 := linmath.NewVector2(old.GetPos().X, old.GetPos().Z)
		//		newvec2 := linmath.NewVector2(new.GetPos().X, new.GetPos().Z)
		//		changedis := newvec2.Sub(oldvec2).Len()
		//		limitdis := float32(common.GetTBSystemValue(2011))
		//		if changedis > limitdis {
		//			user.Debug("速度非法, 坠落跳起水平位移偏大", old.GetPos(), new.GetPos(), changedis, limitdis)
		//			return false
		//		}
	case RoomPlayerBaseState_Passenger:
		limitSpeed = user.rolebase.Vehiclespeedlimit

	default:
		return true
	}

	dist := new.GetPos().Sub(old.GetPos()).Len()
	if new.BaseState == RoomPlayerBaseState_Fall || new.BaseState == RoomPlayerBaseState_Jump {
		//跳起，坠落取水平距离
		dist = common.Distance(new.GetPos(), old.GetPos())
	}

	dur := float32(new.TimeStamp-old.TimeStamp) * float32(GetSrvInst().GetFrameDeltaTime().Seconds())
	curSpeed := dist / dur

	if curSpeed > limitSpeed {
		//user.Debug("速度非法 ", curSpeed, user, "坐标: ", new.GetPos(), old.GetPos(), "状态", new.BaseState, "原状态", old.BaseState, " 计算：", dist, dur, new.TimeStamp-old.TimeStamp)

		//副驾速度非法，重置坐标
		if new.BaseState == RoomPlayerBaseState_Passenger {
			space := user.GetSpace().(*Scene)
			prop := user.GetVehicleProp()
			if prop != nil && prop.PilotID != 0 {
				pilot, ok := space.GetEntity(prop.PilotID).(*RoomUser)
				if ok {
					user.SetPos(pilot.GetPos())
					user.Debug(user.GetID(), " 速度非法,副驾强制设置坐标", old.GetPos(), new.GetPos(), pilot.GetPos())
					return false
				}
			}
		}

		return false
	}

	return true
}

func (user *RoomUser) checkStateValidate(old, new *RoomPlayerState) bool {
	if old.BaseState == new.BaseState {
		return true
	}

	if !user.CanChangeState() {
		return false
	}

	if old.BaseState != new.BaseState && (old.BaseState == RoomPlayerBaseState_Ride || old.BaseState == RoomPlayerBaseState_Passenger) {
		// user.Error("驾驶状态， 客户端不能通知切换")
		return false
	}

	//TODO:这儿可以加个简单的状态验证，比如 坠落状态不能转换为跳起状态，游泳不能转换为跳起等等
	//验证旧状态是否是新状态所对应的能转换至该新状态的状态列表中的一个

	return true
}

func (user *RoomUser) getCanStandHeight(x, z float32) float32 {
	space := user.GetSpace().(*Scene)
	origin := linmath.Vector3{
		X: x,
		Y: 1000,
		Z: z,
	}
	direction := linmath.Vector3{
		X: 0,
		Y: -1,
		Z: 0,
	}

	_, pos, _, hit, _ := space.Raycast(origin, direction, 2000, unityLayerGround|unityLayerBuilding|unityLayerFurniture)
	if hit {
		return pos.Y
	}

	return 0
}

func (user *RoomUser) getStandHeight() float32 {
	space := user.GetSpace().(*Scene)
	origin := linmath.Vector3{
		X: user.GetPos().X,
		Y: user.GetPos().Y + 1.6,
		Z: user.GetPos().Z,
	}
	direction := linmath.Vector3{
		X: 0,
		Y: -1,
		Z: 0,
	}

	_, pos, _, hit, _ := space.Raycast(origin, direction, 2000, unityLayerGround|unityLayerBuilding|unityLayerFurniture|unityLayerPlantAndRock)
	if hit {
		return pos.Y
	}

	return 0
}

//校验Y坐标改变是否合法
func (user *RoomUser) checkPosValidate(old, new *RoomPlayerState) bool {
	//高度检验控制开关
	if !user.safeCheckHeight {
		return true
	}

	if !user.gm.isValidate {
		return true
	}

	oldpos := old.GetPos()
	newpos := new.GetPos()
	disy := math.Abs(float64(oldpos.Y - newpos.Y))
	oldstate := old.BaseState
	newstate := new.BaseState

	if newstate != RoomPlayerBaseState_Fall {
		user.SetFallEndTime(0)
	}

	if newstate != RoomPlayerBaseState_Jump {
		user.SetJumpEndTime(0)
	}

	if newstate != RoomPlayerBaseState_Ride && newstate != RoomPlayerBaseState_Passenger {
		user.SetRideAirEndTime(0)
	}

	//跳伞准备不校验高度
	if newstate == RoomPlayerBaseState_Inplane {
		return true
	}

	//俯冲
	if newstate == RoomPlayerBaseState_Glide {
		changeheight := float64(common.GetTBSystemValue(2009))

		if disy > changeheight || newpos.Y > oldpos.Y {
			user.Debug(user.GetID(), " 俯冲高度差变化异常", oldpos, newpos, oldstate, changeheight)
			return false
		}

		return true
	}

	//跳伞
	if newstate == RoomPlayerBaseState_Parachute {
		changeheight := float64(common.GetTBSystemValue(2010))

		if disy > changeheight || newpos.Y > oldpos.Y {
			user.Debug(user.GetID(), " 开伞高度差变化异常", oldpos, newpos, oldstate, changeheight)
			return false
		}

		return true
	}

	//死亡，濒死，观战不校验高度
	if newstate == RoomPlayerBaseState_Dead || newstate == RoomPlayerBaseState_Watch {
		return true
	}

	if newstate == RoomPlayerBaseState_LeaveMap || newstate == RoomPlayerBaseState_LoadingMap {
		return true
	}

	//检查低于地面高度
	if user.checkUnderGround(old, new) {
		return false
	}

	if newstate == RoomPlayerBaseState_WillDie {
		return true
	}

	//游泳
	if newstate == RoomPlayerBaseState_Swim {
		waterLevel := user.GetSpace().(*Scene).mapdata.Water_height
		isWater, err := user.GetSpace().IsWater(newpos.X, newpos.Z, waterLevel)
		if err != nil || (!isWater && (newpos.Y > waterLevel+1 || newpos.Y < waterLevel-1)) {
			user.Debug(user.GetID(), " 在非水域游泳异常", oldpos, newpos)
			return false
		}

		if newpos.Y > waterLevel+1 || newpos.Y < waterLevel-1 {
			newpos.Y = waterLevel

			user.Debug(user.GetID(), " 游泳高度异常， 强设置水面高度", oldpos, newpos, waterLevel)
			user.SetPos(newpos)
			return false
		}

		return true
	}

	//站立，蹲，趴
	if newstate == RoomPlayerBaseState_Stand || newstate == RoomPlayerBaseState_Down || newstate == RoomPlayerBaseState_Crouch {
		changeheight := float64(common.GetTBSystemValue(2001))

		if disy > changeheight {
			user.Debug(user.GetID(), " 站立,蹲,趴高度差变化异常", oldpos, newpos, old.BaseState, changeheight)
			return false
		}

		//前后坐标都不在地面上, 改为角色坠落状态
		if !user.isOnNormalPos(oldpos) && !user.isOnNormalPos(newpos) {
			user.SetBaseState(RoomPlayerBaseState_Fall)
			user.SetLastFallTime(time.Now().Unix())

			falllast := user.getFallTime(newpos)
			user.SetFallEndTime(time.Now().Unix() + int64(falllast))
			user.Debug(user.GetID(), " 站立,蹲,趴前后都不在地面上，改为坠落状态", oldpos, newpos, new.BaseState, " 坠落持续:", falllast)
			return true //此处返回true，同意客户端坐标，但强制变为坠落状态
		}

		return true
	}

	//坠落
	if newstate == RoomPlayerBaseState_Fall {

		//其它状态切换至坠落
		if oldstate != newstate {
			changeheight := float64(common.GetTBSystemValue(2004))

			if disy > changeheight {
				user.Debug(user.GetID(), " 其它状态切换至坠落高度差变化异常", oldpos, newpos, oldstate, changeheight)
				return false
			}

			//记录坠落持续时间
			user.SetLastFallTime(time.Now().Unix())
			falllast := user.getFallTime(newpos)
			user.SetFallEndTime(time.Now().Unix() + int64(falllast))
			return true
		}

		if newpos.Y > oldpos.Y {
			user.Debug(user.GetID(), " 持续坠落状态高度升高异常", oldpos, newpos)
			return false
		}

		//坠落时间太长直接变站立并强设坐标
		if user.GetFallEndTime() != 0 && time.Now().Unix() > user.GetFallEndTime() {
			user.SetFallEndTime(0)
			user.SetBaseState(RoomPlayerBaseState_Stand)
			user.resetStandPos()
			user.doFallDam()
			user.Debug(user.GetID(), " 坠落时间太长重置坐标", oldpos, user.GetPos())
			return false
		}

		return true
	}

	//跳起
	if newstate == RoomPlayerBaseState_Jump {

		//其它状态切换至跳起
		if oldstate != newstate {
			changeheight := float64(common.GetTBSystemValue(2005))

			if disy > changeheight {
				user.Debug(user.GetID(), " 其它状态切换至跳起高度差变化异常", oldpos, newpos, oldstate, changeheight)
				return false
			}

			//记录跳起持续时间
			user.SetLastJumpTime(time.Now().Unix())
			jumplast := user.getJumpTime(newpos)
			user.SetJumpEndTime(time.Now().Unix() + int64(jumplast))
			return true
		}

		changeheight := float64(common.GetTBSystemValue(2005))
		if newpos.Y < oldpos.Y || disy > changeheight {
			user.Debug(user.GetID(), " 持续跳起状态高度下降异常或者位移太大", oldpos, newpos)
			return false
		}

		//跳起时间太长直接变站立并强设坐标
		if user.GetJumpEndTime() != 0 && time.Now().Unix() > user.GetJumpEndTime() {
			user.Debug(user.GetID(), " 跳起时间太长重置坐标", oldpos, newpos)
			user.SetJumpEndTime(0)
			user.SetBaseState(RoomPlayerBaseState_Stand)
			user.resetStandPos()
			return false
		}

		return true
	}

	//驾驶，乘客
	if newstate == RoomPlayerBaseState_Ride || newstate == RoomPlayerBaseState_Passenger {
		changeheight := float64(common.GetTBSystemValue(2007))

		if disy > changeheight {
			user.Debug(user.GetID(), " 驾驶状态高度差变化异常", oldpos, newpos, oldstate, changeheight)
			return false
		}

		//不再地面上驾驶
		if !user.isRideOnNormalPos(newpos) {

			//从再地面切入不再地面记录时间
			if user.isRideOnNormalPos(oldpos) {
				lastime := int64(common.GetTBSystemValue(2008))
				user.SetRideAirEndTime(time.Now().Unix() + lastime)
				user.Debug(user.GetID(), " 驾驶从地面切入不再地面记录时间", oldpos, newpos, lastime)
				return true
			}

			if user.GetRideAirEndTime() != 0 && time.Now().Unix() > user.GetRideAirEndTime() {
				user.SetRideAirEndTime(0)

				user.vehicleEngineBroke(true)
				user.Debug(user.GetID(), " 驾驶持续10s不再地面上， 强制车损毁")
				return false
			}

			return true
		}

		user.SetRideAirEndTime(0)
		return true
	}

	return true
}

func (user *RoomUser) addStateLast(old, new *RoomPlayerState) {
	now := uint32(time.Now().Unix())
	if _, ok := user.stateLastTime[new.BaseState]; !ok {
		user.stateLastTime[new.BaseState] = make(map[byte]uint32, 0)
	}

	if new.BaseState == RoomPlayerBaseState_Fall {
		if old.BaseState == RoomPlayerBaseState_Glide || old.BaseState == RoomPlayerBaseState_Parachute {
			user.stateLastTime[new.BaseState][old.BaseState] = now + 3
		}
	}
}

func (user *RoomUser) getValidStateHeight(old, new *RoomPlayerState) float32 {
	now := uint32(time.Now().Unix())
	if _, ok := user.stateLastTime[new.BaseState]; !ok {
		return 3
	}

	if new.BaseState == RoomPlayerBaseState_Fall {
		for oldState, v := range user.stateLastTime[new.BaseState] {
			if oldState == RoomPlayerBaseState_Glide || oldState == RoomPlayerBaseState_Parachute {
				if now <= v {
					return 15.0
				} else {
					delete(user.stateLastTime[new.BaseState], oldState)
					return 3
				}
			}
		}
	}

	return 3
}

// checkRotaValidate 玩家倾斜角度校验
func (user *RoomUser) checkRotaValidate(newState *RoomPlayerState) {
	// if newState.BaseState == RoomPlayerBaseState_Stand && (newState.Rota.X != 0 || newState.Rota.Z != 0) {
	// 	newState.Rota.X = 0
	// 	newState.Rota.Z = 0
	// 	user.SetRota(newState.Rota)
	// }
}

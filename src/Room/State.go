package main

import (
	"common"
)

const (
	ActionNone   = 0
	ActionRescue = 1
	Potion       = 2
	Shoot        = 3
	ShootPrepare = 4
	Aim          = 5
	ThrowPrepare = 6
	Throw        = 7
	Melee        = 8
	MeleePrepare = 9
	Change       = 10
	Reload       = 11
)

//StateM 状态管理
type StateM struct {
	user    IRoomChracter
	hasinit bool
}

//InitStateM 初始化
func InitStateM(user IRoomChracter) *StateM {
	statem := &StateM{}
	statem.user = user
	statem.hasinit = false

	return statem
}

//GetState 获取当前状态
func (sf *StateM) GetState() uint8 {
	return sf.user.GetBaseState()
}

//CanAliveDown 趴下
func (sf *StateM) CanAliveDown() bool {
	state := sf.user.GetBaseState()
	if state == RoomPlayerBaseState_Inplane || state == RoomPlayerBaseState_Swim {
		return false
	}

	return true
}

//CanUseObject 使用道具
func (sf *StateM) CanUseObject() bool {
	state := sf.user.GetBaseState()
	if state == RoomPlayerBaseState_Stand || state == RoomPlayerBaseState_Down || state == RoomPlayerBaseState_Crouch {
		return true
	}

	// 可移动使用回血道具
	if sf.user.GetSkillEffectDam(SE_MoveUseItem) != 0 {
		if state == RoomPlayerBaseState_Ride || state == RoomPlayerBaseState_Passenger || state == RoomPlayerBaseState_Swim {
			return true
		}
	}

	return false
}

//CanAttack 攻击
func (sf *StateM) CanAttack() bool {
	state := sf.user.GetBaseState()
	if state == RoomPlayerBaseState_Inplane || state == RoomPlayerBaseState_Swim {
		// log.Info("不能发起攻击，当前状态", state)
		return false
	}

	if state == RoomPlayerBaseState_WillDie || state == RoomPlayerBaseState_Dead || state == RoomPlayerBaseState_Watch {
		// log.Info("不能发起攻击，当前状态", state)
		return false
	}

	return true
}

//CanDrop 是否能主动丢弃
func (sf *StateM) CanDrop() bool {
	state := sf.user.GetBaseState()
	if state == RoomPlayerBaseState_Watch {
		return false
	}

	return true
}

// CanFall 掉落检查
func (sf *StateM) CanFall() bool {
	state := sf.user.GetBaseState()
	if state == RoomPlayerBaseState_WillDie || state == RoomPlayerBaseState_Dead || state == RoomPlayerBaseState_Watch {
		return false
	}

	return true
}

// CanSwim 游泳检查
func (sf *StateM) CanSwim() bool {
	state := sf.user.GetBaseState()
	if state == RoomPlayerBaseState_WillDie || state == RoomPlayerBaseState_Dead || state == RoomPlayerBaseState_Watch {
		return false
	}

	if state == RoomPlayerBaseState_Ride || state == RoomPlayerBaseState_Passenger {
		return false
	}

	return true
}

// CanChangeState 状态检查
func (sf *StateM) CanChangeState() bool {
	state := sf.user.GetBaseState()
	if state == RoomPlayerBaseState_WillDie || state == RoomPlayerBaseState_Dead || state == RoomPlayerBaseState_Watch {
		return false
	}

	return true
}

//CanUpVehicle 上车
func (sf *StateM) CanUpVehicle() bool {
	state := sf.user.GetBaseState()
	if state == RoomPlayerBaseState_Inplane {
		// log.Info("不能上车，当前状态", state)
		return false
	}

	if state == RoomPlayerBaseState_WillDie || state == RoomPlayerBaseState_Dead || state == RoomPlayerBaseState_Watch {
		// log.Info("不能上车，当前状态", state)
		return false
	}

	if state == RoomPlayerBaseState_Ride || state == RoomPlayerBaseState_Passenger {
		// log.Info("不能上车，当前在车上")
		return false
	}

	if state == RoomPlayerBaseState_Fall || state == RoomPlayerBaseState_Jump {
		// log.Info("不能上车，当前在坠落")
		return false
	}

	return true
}

//IsReadyParachute 是否跳伞准备
func (sf *StateM) IsReadyParachute() bool {
	return sf.user.GetBaseState() == RoomPlayerBaseState_Inplane
}

//IsSwimming 是否游泳
func (sf *StateM) IsSwimming() bool {
	return sf.user.GetBaseState() == RoomPlayerBaseState_Swim
}

//Update 更新
func (sf *StateM) Update() {
	if sf.hasinit {
		pos := sf.user.GetPos()
		waterLevel := sf.user.GetSpace().(*Scene).mapdata.Water_height
		isWater, err := sf.user.IsWater(pos, waterLevel)
		if err != nil {
			return
		}
		if isWater {
			if !sf.IsSwimming() {
				if !sf.user.isOnBridge(sf.user.GetPos()) && (sf.GetState() == RoomPlayerBaseState_Ride || sf.GetState() == RoomPlayerBaseState_Passenger) && !sf.user.IsinBoat() {
					sf.user.vehicleEngineBroke(true)
				}

				if sf.CanSwim() && pos.Y <= waterLevel+0.2 {
					sf.user.SetBaseState(RoomPlayerBaseState_Swim)

					if pos.Y < waterLevel {
						pos.Y = waterLevel
						sf.user.SetPos(pos)
					}
				}
			}
		} else {
			if (sf.GetState() == RoomPlayerBaseState_Ride || sf.GetState() == RoomPlayerBaseState_Passenger) && sf.user.IsinBoat() {
				pos.Y, err = sf.user.GetSpace().GetHeight(pos.X, pos.Z)
				if err == nil && pos.Y > waterLevel+1.6 {
					sf.user.vehicleEngineBroke(false)
				}
			}

			if sf.IsSwimming() {
				sf.user.SetBaseState(RoomPlayerBaseState_Stand)
			}
		}

		pos = sf.user.GetPos()
		space := sf.user.GetSpace().(*Scene)
		clientWaterLevel := space.mapdata.Client_Water_height
		height, err := space.GetHeight(pos.X, pos.Z)

		if err == nil && height < clientWaterLevel+waterLevel && !sf.user.isOnBridge(pos) {
			if sf.GetState() == RoomPlayerBaseState_Crouch || sf.GetState() == RoomPlayerBaseState_Down {
				sf.user.SetBaseState(RoomPlayerBaseState_Stand)
				sf.user.AdviceNotify(common.NotifyCommon, 55)
			}
		}
	}
}

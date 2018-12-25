package main

import (
	"errors"
	"time"
	"zeus/linmath"
)

const (
	maxGlideTime     = 60  // 最大滑翔时间，单位：秒
	maxParachuteTime = 40  // 最大飘伞时间，单位：秒 应根据开伞时高度计算一个时间
	glideSpeed       = 20  // 滑翔速度，单位：米/秒(27)
	parachureSpeed   = 4   // 飘伞速度，单位：米/秒(4.7)
	parachureMinHigh = 120 // 最小开伞高度
)

// ParachuteCtrl 玩家跳伞过程控制
type ParachuteCtrl struct {
	user                  *RoomUser
	glideStartTime        time.Time
	parachuteStartTime    time.Time
	addedParachureSeconds float64 // 提前开伞，需要增加强制落地的时间
}

// NewParachuteCtrl 创建玩家跳伞过程控制器
func NewParachuteCtrl(u *RoomUser) *ParachuteCtrl {

	ctrl := &ParachuteCtrl{
		user: u,
	}

	return ctrl
}

// StartGlide 跳机，开始滑翔
func (ctrl *ParachuteCtrl) StartGlide() {
	ctrl.glideStartTime = time.Now()
	// ctrl.user.Debug("StartGlide 开始滑翔")
}

// StartParachute 滑翔结束，打开降落伞
func (ctrl *ParachuteCtrl) StartParachute() {
	if !ctrl.parachuteStartTime.IsZero() {
		return
	}
	ctrl.parachuteStartTime = time.Now()
	// ctrl.user.Debug("StartParachute 打开降落伞")
	curHight := ctrl.user.GetPos().Y
	if curHight > 750 {
		curHight = 750
	}
	standPos, err := ctrl.getStandPos()
	if err == nil {
		curHight -= standPos.Y
	}
	if curHight > parachureMinHigh {
		ctrl.addedParachureSeconds = float64(curHight/parachureMinHigh-1) * maxParachuteTime
	}
}

// CalcCurPos 服务器接管高度位置计算
func (ctrl *ParachuteCtrl) CalcCurPos() (linmath.Vector3, error) {
	// 实时计算玩家控制位置，并广播
	lastUpdateTime := ctrl.user.GetStates().GetLastState().GetTimeStamp()
	currTime := ctrl.user.GetSpace().GetTimeStamp()

	var minPos linmath.Vector3
	var speed float32
	baseState := ctrl.user.GetBaseState()
	if baseState == RoomPlayerBaseState_Glide {
		speed = glideSpeed
		minPos, _ = ctrl.getParachutePos()
		minPos.Y += 2
	} else if baseState == RoomPlayerBaseState_Parachute {
		speed = parachureSpeed
		minPos, _ = ctrl.getStandPos()
	} else {
		return ctrl.getStandPos()
	}

	dis := float32(currTime-lastUpdateTime) * 0.033333 * speed
	curPos := ctrl.user.GetPos()
	curPos.Y -= dis

	// ctrl.user.Debug("CalcCurPos lastTime:", lastUpdateTime, " curTime:", currTime, " dis:", dis, " minPos:", minPos, " curPos:", curPos)
	if minPos.Y > curPos.Y {
		return minPos, nil
	}

	return curPos, nil
}

func (ctrl *ParachuteCtrl) getParachutePos() (linmath.Vector3, error) {
	pos, err := ctrl.getStandPos()
	if err != nil {
		return pos, err
	}

	pos.Y = pos.Y + parachureMinHigh
	return pos, nil
}

func (ctrl *ParachuteCtrl) getStandPos() (linmath.Vector3, error) {
	pos := ctrl.user.GetPos()

	waterLevel := ctrl.user.GetSpace().(*Scene).mapdata.Water_height
	isWater, err := ctrl.user.GetSpace().IsWater(pos.X, pos.Z, waterLevel)
	if err != nil {
		return linmath.Vector3_Zero(), err
	}

	if isWater {
		pos.Y = waterLevel
	} else {
		standheight := ctrl.user.getCanStandHeight(pos.X, pos.Z)
		if standheight == 0 {
			return linmath.Vector3_Zero(), errors.New("stand height is zero")
		}
		pos.Y = standheight
	}

	return pos, nil
}

// update 定时更新
func (ctrl *ParachuteCtrl) update() {
	baseState := ctrl.user.GetBaseState()
	if baseState == RoomPlayerBaseState_Glide && time.Now().Sub(ctrl.glideStartTime).Seconds() > maxGlideTime {
		// 计算开闪位置，广播开伞状态
		// ctrl.user.Debug("ParachuteCtrl 自动开伞", ctrl.glideStartTime)
		pos, err := ctrl.getParachutePos()
		if err == nil {
			ctrl.user.GetPlayerState().SetTimeStamp(ctrl.user.GetSpace().GetTimeStamp())
			ctrl.user.SetPos(pos)
			ctrl.user.SetBaseState(RoomPlayerBaseState_Parachute)
			ctrl.StartParachute()
		} else {
			ctrl.user.Debug(err.Error())
		}
	} else if baseState == RoomPlayerBaseState_Parachute && !ctrl.parachuteStartTime.IsZero() && time.Now().Sub(ctrl.parachuteStartTime).Seconds() > maxParachuteTime+ctrl.addedParachureSeconds {
		// 计算落地位置，广播落地状态
		// ctrl.user.Debug("ParachuteCtrl 自动落地", ctrl.parachuteStartTime)
		pos, err := ctrl.getStandPos()
		if err == nil {
			ctrl.user.GetPlayerState().SetTimeStamp(ctrl.user.GetSpace().GetTimeStamp())
			ctrl.user.SetPos(pos)
			ctrl.user.SetBaseState(RoomPlayerBaseState_Stand)
			ctrl.user.setAiPos(ctrl.user.GetPos())
			ctrl.user.initPackItems()
		} else {
			ctrl.user.Debug(err.Error())
		}
		// 已落地，移除跳伞控制
		ctrl.user.parachuteCtrl = nil
	}
}

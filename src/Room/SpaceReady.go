package main

import (
	"common"
	//"db"
	"excel"
	"math"
	"math/rand"
	"time"
	"zeus/iserver"
	"zeus/linmath"
)

//生成随机航线
func (tb *Scene) genAirLine() {
	if tb.GetMatchMode() != common.MatchModeArcade && tb.GetMatchMode() != common.MatchModeTankWar {
		stepNum := int(tb.mapdata.Parts_num)
		minStep := int(tb.mapdata.End_point)
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		startPoint := r.Intn(stepNum)
		rs := r.Intn(stepNum/2-minStep) + 1
		endPointFront := (startPoint + minStep + rs) % stepNum
		endPointBack := (stepNum + startPoint - minStep - rs) % stepNum
		tb.Info("startpoint=", startPoint, ",step=", minStep, ",rs=", rs, ",stepnum=", stepNum, "front=", endPointFront, ",back=", endPointBack)
		tb.airlineStart = tb.changePoint(startPoint, tb.mapdata.Width, stepNum)
		switch r.Intn(2) {
		case 0:
			tb.airlineEnd = tb.changePoint(endPointFront, tb.mapdata.Width, stepNum)
		case 1:
			tb.airlineEnd = tb.changePoint(endPointBack, tb.mapdata.Width, stepNum)
		}
	} else {
		var airLineRadius float32
		if tb.GetMatchMode() == common.MatchModeArcade {
			airLineRadius = float32(common.GetTBSystemValue(common.System_ArcadeAirLineRadius))
		} else if tb.GetMatchMode() == common.MatchModeTankWar {
			airLineRadius = float32(common.GetTBSystemValue(common.System_TankWarAirLineRadius))
		}

		rand.Seed(time.Now().UnixNano())
		angle := rand.Float64()
		cosX := float32(math.Cos(2*angle*math.Pi)) * airLineRadius
		sinZ := float32(math.Sin(2*angle*math.Pi)) * airLineRadius
		tb.airlineStart = linmath.Vector3{
			X: GetRefreshZoneMgr(tb).nextsafecenter.X + cosX,
			Y: 0,
			Z: GetRefreshZoneMgr(tb).nextsafecenter.Z + sinZ,
		}

		distance := common.Distance(tb.airlineStart, GetRefreshZoneMgr(tb).tmpsafecenter)
		sTmp := (airLineRadius + distance + tb.mapdata.Saferadius) * (airLineRadius + distance - tb.mapdata.Saferadius) * (airLineRadius + tb.mapdata.Saferadius - distance) * (distance + tb.mapdata.Saferadius - airLineRadius)
		if sTmp < 0 {
			tb.Debug("sTmp:", sTmp)
			sTmp = 0
		}
		airLineHigh := float32(math.Sqrt(float64(sTmp))) / (2 * distance)
		airLineLength := 2 * float32(math.Sqrt(float64(airLineRadius*airLineRadius-airLineHigh*airLineHigh)))
		tmpX := (GetRefreshZoneMgr(tb).tmpsafecenter.X - tb.airlineStart.X) * airLineLength / distance
		tmpZ := (GetRefreshZoneMgr(tb).tmpsafecenter.Z - tb.airlineStart.Z) * airLineLength / distance

		tb.airlineEnd = linmath.Vector3{
			X: tb.airlineStart.X + tmpX,
			Y: 0,
			Z: tb.airlineStart.Z + tmpZ,
		}

		distanceEnd := common.Distance(tb.airlineEnd, GetRefreshZoneMgr(tb).nextsafecenter)

		tb.Debug("angle:", angle, " cosX:", cosX, " sinZ:", sinZ, " airline,start=", tb.airlineStart, " end=", tb.airlineEnd, " airLineRadius:", airLineRadius, " distance:", distance, " distanceEnd:", distanceEnd, " airLineHigh:", airLineHigh, " airLineLength:", airLineLength, " tmpX:", tmpX, " tmpZ:", tmpZ, " sTmp:", sTmp)
	}

	tb.Info("airline,start=", tb.airlineStart, ",end=", tb.airlineEnd)
}
func (tb *Scene) changePoint(point int, width float32, stepNum int) linmath.Vector3 {
	step := int(width * 4 / float32(stepNum))
	pointLength := float32(point * step)
	var ss linmath.Vector3
	switch int(pointLength / width) {
	case 0:
		ss = linmath.NewVector3(pointLength, 0, 0)
	case 1:
		ss = linmath.NewVector3(width, 0, pointLength-width)
	case 2:
		ss = linmath.NewVector3(3*width-pointLength, 0, width)
	case 3:
		ss = linmath.NewVector3(0, 0, 4*width-pointLength)
	default:

	}
	return ss
}

func (tb *Scene) checkAllowParachuteState() {
	allReady := true
	isExistPlayer := false

	tb.TravsalEntity("Player", func(e iserver.IEntity) {
		if e == nil {
			return
		}

		if user, ok := e.(*RoomUser); ok {
			isExistPlayer = true
			if !user.StateM.IsReadyParachute() {
				allReady = false
			}
		}
	})

	curTime := time.Now().Unix()
	if allReady && isExistPlayer && curTime >= tb.createStamp+int64(tb.minLoadTime) {
		tb.startParachute()
	}
}

// 所有人加载完毕, 可以开始跳伞
func (tb *Scene) startParachute() {
	if tb.allowParachute {
		return
	}

	// 初始化组队信息
	tb.teamMgr.InitRoomTeamInfo(tb)
	// 载入玩家在该模式的评分
	tb.LoadMemberRating()

	//tb.Error("客户端加载完毕: ", len(tb.members))
	// 广播场景人数
	tb.BroadAliveNum()
	tb.BroadAirLeft()

	tb.loadFulPlayerNum = uint32(len(tb.members))
	tb.allowParachute = true
	tb.allowParachuteTime = time.Now()
	tb.parachuteTime = time.Now().Unix()

	airlineDist := tb.airlineStart.Sub(tb.airlineEnd).Len()
	flyTime := time.Duration(airlineDist / float32(tb.mapdata.Fly_Speed) * tb.mapdata.Drop_Time_Point)
	tb.AddDelayCallByObj(tb, tb.ForceEject, flyTime*time.Second)

	system, ok := excel.GetSystem(uint64(common.System_ForceEjectNotify))
	if ok {
		if uint64(flyTime) > system.Value {
			tb.AddDelayCallByObj(tb, tb.ForceEjectNotify, (flyTime-time.Duration(system.Value))*time.Second)
		}
	}

	subaitime := time.Duration(airlineDist / float32(tb.mapdata.Fly_Speed) * tb.mapdata.AI_Drop_Time * 1000)
	tb.nextsubaitime = time.Now().UnixNano()/1000000 + int64(subaitime) + int64(tb.mapdata.AI_Drop_Delay*1000)

	tb.TravsalEntity("Player", func(e iserver.IEntity) {
		if e == nil {
			return
		}

		if user, ok := e.(*RoomUser); ok {
			// 可以跳伞之后, 还未处于可跳伞状态的玩家从场景中移除
			// if !user.StateM.IsReadyParachute() {
			// 	// tb.Debug("强制跳伞删除")
			// 	if err := user.RPC(iserver.ServerTypeClient, "AllowParachute"); err != nil {
			// 		tb.Error(err, user)
			// 	}
			// 	// user.RPC(iserver.ServerTypeClient, "ForceLeaveSpace")
			// 	// user.LeaveScene()
			// 	return
			// 	// if err := tb.RemoveEntity(user.GetID()); err != nil {
			// 	// 	tb.Error(err, user)
			// 	// }
			// }
			if user.GetBaseState() == RoomPlayerBaseState_LoadingMap {
				user.SetBaseState(RoomPlayerBaseState_Inplane)
			}

			if err := user.RPC(iserver.ServerTypeClient, "AllowParachute"); err != nil {
				tb.Error("RPC AllowParachute err: ", err)
			}

			//if db.PlayerTempUtil(user.GetDBID()).GetPlayerJumpAir() == 1 {
			if true { //跳过跳伞
				// user.SetState(Stand)
				user.SetBaseState(RoomPlayerBaseState_Stand)
				if tb.mapdata.Id == 1 {
					user.SetPos(linmath.NewVector3(6312, 40.5, 3934))
				} else {
					user.SetPos(linmath.NewVector3(1506, 30, 811))
				}
				//tb.Info("临时设置客户端跳伞位置", user.GetID(), user.GetPos())
				user.UpdateCell()
				user.StateM.hasinit = true
			}
		}
	})

	//if tb.maploadSuccess {
	//	tb.SummonAI(tb.aiNum)
	//}
}

// ForceEject 强制跳伞
func (tb *Scene) ForceEject() {
	tb.TravsalEntity("Player", func(e iserver.IEntity) {
		if user, ok := e.(*RoomUser); ok {
			if user.StateM.IsReadyParachute() {
				// tb.Info("客户端强制跳伞", user)
				user.doParachute(true)
				if err := user.RPC(iserver.ServerTypeClient, "ParachutePos"); err != nil {
					tb.Error("RPC ParachutePos err: ", err)
				}
				user.RPC(iserver.ServerTypeClient, "ForceEject")

				user.tlogBattleFlow(behavetype_forcejump, 0, 0, 0, 0, 0) // tlog战场流水表(behavetype_forcejump代表强制跳伞)
			}
		}
	})
}

// ForceEjectNotify 强制跳伞
func (tb *Scene) ForceEjectNotify() {
	tb.TravsalEntity("Player", func(e iserver.IEntity) {
		if user, ok := e.(*RoomUser); ok {
			if user.StateM.IsReadyParachute() {
				// tb.Info("客户端强制跳伞", user)
				if err := user.RPC(iserver.ServerTypeClient, "ForceEjectNotify"); err != nil {
					tb.Error("RPC ForceEjectNotify err: ", err)
				}
			}
		}
	})
}

func (tb *Scene) SubAirAI() {
	if tb.onAirAiNum == 0 || tb.nextsubaitime == 0 {
		return
	}

	now := time.Now().UnixNano() / 1000000
	if now >= tb.nextsubaitime {
		dropnum := int(tb.mapdata.AI_Drop_Percent * float32(tb.aiNum))
		if dropnum == 0 {
			dropnum = 1
		}

		num := uint32(rand.Intn(dropnum) + 1)
		if num > tb.onAirAiNum {
			tb.onAirAiNum = 0
		} else {
			tb.onAirAiNum -= num
		}

		airlineDist := tb.airlineStart.Sub(tb.airlineEnd).Len()
		subaitime := time.Duration(airlineDist / float32(tb.mapdata.Fly_Speed) * tb.mapdata.AI_Drop_Time * 1000)
		tb.nextsubaitime = now + int64(subaitime)
		tb.BroadAirLeft()
	}
}

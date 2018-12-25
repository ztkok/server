package main

import (
	"common"
	"excel"
	"math"
	"time"
	"zeus/linmath"

	"protoMsg"

	"github.com/spf13/viper"
)

type SafeCheckM struct {
	user *RoomUser

	lastfalltime int64
	fallendtime  int64
	lastfallpos  linmath.Vector3

	lastjumptime int64
	jumpendtime  int64

	rideairendtime int64

	lastattacktime int64
	attackcount    map[uint32]uint32

	lastheadshottime int64
	headshotuser     map[uint64]uint32

	lastFireTime    uint32
	lastFireDir     string
	lastFireOrigion string
}

func InitSafeCheckM(user *RoomUser) *SafeCheckM {
	safecheck := &SafeCheckM{}
	safecheck.user = user
	safecheck.lastfallpos = linmath.Vector3_Zero()
	safecheck.attackcount = make(map[uint32]uint32)
	safecheck.headshotuser = make(map[uint64]uint32)

	return safecheck
}

func (sc *SafeCheckM) SetRideAirEndTime(time int64) {
	sc.rideairendtime = time
}

func (sc *SafeCheckM) GetRideAirEndTime() int64 {
	return sc.rideairendtime
}

func (sc *SafeCheckM) SetLastFallTime(time int64) {
	sc.lastfalltime = time
}

func (sc *SafeCheckM) SetFallEndTime(time int64) {
	sc.fallendtime = time
}

func (sc *SafeCheckM) GetFallEndTime() int64 {
	return sc.fallendtime
}

func (sc *SafeCheckM) SetLastJumpTime(time int64) {
	sc.lastjumptime = time
}

func (sc *SafeCheckM) SetJumpEndTime(time int64) {
	sc.jumpendtime = time
}

func (sc *SafeCheckM) GetJumpEndTime() int64 {
	return sc.jumpendtime
}

//站立，蹲趴贴近地面
func (sc *SafeCheckM) isOnNormalPos(pos linmath.Vector3) bool {
	space := sc.user.GetSpace().(*Scene)
	origin := linmath.Vector3{
		X: pos.X,
		Y: pos.Y + 1.2,
		Z: pos.Z,
	}
	direction := linmath.Vector3{
		X: 0,
		Y: -1,
		Z: 0,
	}
	_, _, _, hit, _ := space.Raycast(origin, direction, 4, unityLayerGround|unityLayerBuilding|unityLayerFurniture|unityLayerPlantAndRock)
	if hit {
		// log.Error("检测到建筑物")
		return true
	}

	return false
}

//驾驶贴近地面或水面
func (sc *SafeCheckM) isRideOnNormalPos(pos linmath.Vector3) bool {
	space := sc.user.GetSpace().(*Scene)
	waterLevel := sc.user.GetSpace().(*Scene).mapdata.Water_height
	isWater, err := sc.user.GetSpace().IsWater(pos.X, pos.Z, waterLevel)
	if err != nil {
		return false
	}

	origin := linmath.Vector3{
		X: pos.X,
		Y: pos.Y + 1.2,
		Z: pos.Z,
	}
	direction := linmath.Vector3{
		X: 0,
		Y: -1,
		Z: 0,
	}

	if !isWater {
		_, _, _, hit, _ := space.Raycast(origin, direction, 3, unityLayerGround|unityLayerBuilding|unityLayerFurniture)
		if hit {
			// log.Error("检测到建筑物")
			return true
		}
	} else {
		_, _, _, hit, _ := space.Raycast(origin, direction, 3, unityLayerGround|unityLayerBuilding|unityLayerFurniture)
		if hit {
			// log.Error("检测到建筑物")
			return true
		}

		if pos.Y > (waterLevel-2) && pos.Y < (waterLevel+2) {
			return true
		}
	}

	return false
}

//获取坠落时间
func (sc *SafeCheckM) getFallTime(pos linmath.Vector3) uint32 {
	falltimeadd := uint32(common.GetTBSystemValue(2003))
	space := sc.user.GetSpace().(*Scene)
	origin := linmath.Vector3{
		X: pos.X,
		Y: pos.Y,
		Z: pos.Z,
	}
	direction := linmath.Vector3{
		X: 0,
		Y: -1,
		Z: 0,
	}
	_, hitpos, _, hit, _ := space.Raycast(origin, direction, 1000, unityLayerGround|unityLayerBuilding|unityLayerFurniture)
	if !hit || pos.Y <= hitpos.Y {
		return falltimeadd
	}

	t := math.Sqrt(2 * float64(pos.Y-hitpos.Y) / 9.78)
	return uint32(t) + falltimeadd
}

func (sc *SafeCheckM) resetStandPos() {
	pos := sc.user.GetPos()
	sc.user.Info("-------------resetStandPos, old: ", pos)

	waterLevel := sc.user.GetSpace().(*Scene).mapdata.Water_height
	isWater, err := sc.user.GetSpace().IsWater(pos.X, pos.Z, waterLevel)
	if err != nil {
		return
	}

	standheight := sc.user.getStandHeight()
	if standheight == 0 {
		return
	}
	if isWater {
		standheight = waterLevel
	}

	pos.Y = standheight
	sc.user.SetPos(pos)
	sc.user.Info("-------------resetStandPos, new: ", pos)
}

//获取跳起时间
func (sc *SafeCheckM) getJumpTime(pos linmath.Vector3) uint32 {
	jumptimeadd := uint32(common.GetTBSystemValue(2006))

	return jumptimeadd
}

//处理坠落伤害
func (sc *SafeCheckM) doFallDam() {
	if !sc.user.StateM.CanChangeState() {
		return
	}

	if sc.user.GetState() == RoomPlayerBaseState_Ride || sc.user.GetState() == RoomPlayerBaseState_Passenger {
		//seelog.Error("驾驶状态不掉落")
		return
	}

	pos := sc.user.GetPos()
	waterLevel := sc.user.GetSpace().(*Scene).mapdata.Water_height
	isWater, err := sc.user.GetSpace().IsWater(pos.X, pos.Z, waterLevel)
	if err != nil {
		return
	}
	if isWater {
		// log.Debug("掉落水域不受伤害", p.user.GetID())
		return
	}

	system, ok := excel.GetSystem(uint64(common.System_FallDamageA))
	if !ok {
		return
	}
	A := float64(system.Value)

	system, ok = excel.GetSystem(uint64(common.System_FallDamageB))
	if !ok {
		return
	}
	B := float64(system.Value)

	system, ok = excel.GetSystem(uint64(common.System_FallDamage))
	if !ok {
		return
	}
	damage := float64(system.Value)

	if sc.user.lastfallpos.IsEqual(linmath.Vector3_Zero()) {
		return
	}

	if sc.user.lastfallpos.Y <= pos.Y {
		return
	}

	height := float64(sc.user.lastfallpos.Y - pos.Y)

	// seelog.Debug("坠落", " 参数A:", A, " 参数B:", B, " speed:", height, sc.user.lastfallpos, pos)
	var subhp uint32
	if B != 0 && height > A {
		subhp = uint32(damage * (height - A) / B)
		// seelog.Debug(sc.user.GetID(), "坠落", " 参数A:", A, " 参数B:", B, " damage:", damage, " 掉血:", subhp)
	}

	sc.user.DisposeSubHp(InjuredInfo{num: subhp, injuredType: falldam, isHeadshot: false})
	sc.user.lastfallpos = linmath.Vector3_Zero()
}

func (sc *SafeCheckM) clearAttackCount() {
	for k, _ := range sc.attackcount {
		sc.attackcount[k] = 0
	}
}

func (sc *SafeCheckM) addAttackCount(weapon uint32) {
	sc.attackcount[weapon]++
}

func (sc *SafeCheckM) getAttackCount(weapon uint32) uint32 {
	return sc.attackcount[weapon]
}

//checkAttackCold 是否攻击冷却中
func (sc *SafeCheckM) checkAttackCold() bool {
	now := time.Now().UnixNano() / (1000 * 1000)

	if sc.lastattacktime == 0 || now > sc.lastattacktime+1000 {
		sc.lastattacktime = now
		sc.clearAttackCount()
		return false
	}

	weapon := sc.user.GetInUseWeapon()
	if weapon == 0 {
		weapon = 1009
	}

	base, ok := excel.GetGun(uint64(weapon))
	if !ok {
		//seelog.Debug(sc.user.GetID(), "未找到攻击武器", weapon)
		return true
	}

	//1秒内最大攻击限制
	var maxlimit uint32 = 1
	if base.Shootinterval1 != 0 {
		maxlimit += uint32(1.0 / base.Shootinterval1)
	}

	itembase, ok := excel.GetItem(uint64(weapon))
	if ok {
		if itembase.Type == ItemTypeShotGun {
			maxlimit *= uint32(base.Bulletnum)
		}
	}

	if sc.getAttackCount(weapon) > maxlimit {
		//seelog.Debug(sc.user.GetID(), "攻击冷却中", sc.attackcount, " limit:", maxlimit, " weapon:", weapon)
		return true
	}

	sc.addAttackCount(weapon)
	//seelog.Debug(sc.user.GetID(), "攻击冷却检测", sc.attackcount, " limit:", maxlimit, " weapon:", weapon)
	return false
}

func (sc *SafeCheckM) clearHeadShot() {
	sc.headshotuser = make(map[uint64]uint32)
}

func (sc *SafeCheckM) addHeadShotUser(defendid uint64) {
	sc.headshotuser[defendid]++
}

//checkAttackCold 是否枪枪爆头
func (sc *SafeCheckM) checkHeadShot(defendid uint64) bool {
	if !viper.GetBool("Room.CheckHeadShot") {
		return false
	}

	now := time.Now().UnixNano() / (1000 * 1000)

	if sc.lastheadshottime == 0 || now > sc.lastheadshottime+1000 {
		sc.lastheadshottime = now
		sc.clearHeadShot()
		return false
	}

	weapon := sc.user.GetInUseWeapon()
	if weapon == 0 {
		weapon = 1009
	}

	base, ok := excel.GetGun(uint64(weapon))
	if !ok {
		//seelog.Debug(sc.user.GetID(), "未找到攻击武器", weapon)
		return true
	}

	//1秒内最大爆头人数限制
	var maxlimit uint32 = 1
	if base.Shotheadlimit != 0 {
		maxlimit = uint32(base.Shotheadlimit)
	}

	if uint32(len(sc.headshotuser)) > maxlimit {
		//seelog.Debug(sc.user.GetID(), "单位时间内爆头数量超过限制", len(sc.headshotuser), " limit:", maxlimit, " weapon:", weapon)
		return true
	}

	sc.addHeadShotUser(defendid)
	// seelog.Debug(sc.user.GetID(), "单位时间内爆头数量", len(sc.headshotuser), " limit:", maxlimit, " weapon:", weapon)
	return false
}

// 检测重放开火数据包（一枪秒杀）
func (sc *SafeCheckM) checkFireReplay(reqMsg *protoMsg.AttackReq) bool {

	// 一枪秒杀开关，和枪枪爆头共享
	if !viper.GetBool("Room.CheckHeadShot") {
		return false
	}

	// 未命中，不检测
	if reqMsg.Defendid == 0 {
		return false
	}

	//近战武器不检测
	if sc.user.isMeleeWeaponUse() {
		return false
	}

	// 重放旧的数据包
	if sc.lastFireTime > reqMsg.Firetime {
		//seelog.Debug(sc.user.GetID(), " 重放旧的攻击数据包", sc.lastFireTime, reqMsg.Firetime)
		return true
	}

	stringDir := reqMsg.Dir.String()
	stringOrg := reqMsg.Origion.String()
	if sc.lastFireTime == reqMsg.Firetime &&
		sc.lastFireDir == stringDir &&
		sc.lastFireOrigion == stringOrg {
		//seelog.Debug(sc.user.GetID(), " 重复相同的攻击数据包", sc.lastFireTime, sc.lastFireDir, sc.lastFireOrigion)
		return true
	}

	sc.lastFireTime = reqMsg.Firetime
	sc.lastFireDir = stringDir
	sc.lastFireOrigion = stringOrg

	return false
}

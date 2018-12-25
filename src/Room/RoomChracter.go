package main

import (
	"common"
	"excel"
	"math"
	"protoMsg"
	"time"
	"zeus/iserver"
	"zeus/linmath"
	"zeus/space"
)

// IRoomChracter 场景中的角色
type IRoomChracter interface {
	iserver.ISpaceEntity

	isAI() bool
	fillMapData() *protoMsg.ChracterMapDataInfo
	AddEffect(baseid uint32, id uint32)
	SendChat(str string)
	GetChracterMapDataInfo() *protoMsg.ChracterMapDataInfo
	SetChracterMapDataInfo(v *protoMsg.ChracterMapDataInfo)
	SetChracterMapDataInfoDirty()
	SetBackPackProp(v *protoMsg.BackPackProp)
	SetBackPackPropDirty()
	SetHeadProp(v *protoMsg.HeadProp)
	SetHeadPropDirty()
	SetBodyProp(v *protoMsg.BodyProp)
	SetBodyPropDirty()
	SetIsWearingGilley(v uint32)
	SetIsWearingGilleyDirty()
	GetHeadProp() *protoMsg.HeadProp
	GetBodyProp() *protoMsg.BodyProp
	GetIsWearingGilley() uint32
	GetBackPackProp() *protoMsg.BackPackProp
	GetVehicleProp() *protoMsg.VehicleProp
	GetBaseState() uint8
	SetBaseState(v uint8)
	computeAttack(gunid uint32, targetPos linmath.Vector3, headshot bool) uint32
	downVehicle(pos, dir linmath.Vector3, prop *protoMsg.VehiclePhysics, reducedam bool)
	compilotDown(pos, dir linmath.Vector3, speed float32)
	getRidingCar() *Car
	IsinBoat() bool
	isInTank() bool
	isBazookaWeaponUse() bool
	vehicleEngineBroke(autodown bool)
	add_item(baseid uint32, num uint32) uint32
	isOnBridge(pos linmath.Vector3) bool
	AdviceNotify(notifyType uint32, id uint64)
	GetGamerType() uint32
	SetGamerType(uint32)
	GetEquips() map[uint64]*Item
	CastRpcToAllClient(method string, args ...interface{})

	GetKillNum() uint32
	IncrKillNum()
	GetKiller() uint64
	SetKiller(killer uint64)
	SetVictory()
	IsDead() bool
	IsOffline() bool
	GetName() string
	GetUserTeamID() uint64
	GetRivalGamerType() uint32
	GetWatchingTarget() uint64
	GetWatchTargetInTeam() uint64
	GetWatchers() []uint64
	SetDisabledWatchTarget(target uint64)
	GetDisabledWatchTarget() uint64
	IsWatching() bool
	IsBeingWatched() bool
	DisposeWatch(target uint64)
	CancelWatch()
	AddWatcher(id uint64)
	RemoveWatcher(id uint64)
	BroadcastToWatchers(userType uint32, method string, args ...interface{})
	TerminateWatch()
	SyncWatchTargetInfo()
	LeaveScene()
	GetUserType() uint32
	GetTeamMembers() []uint64
	GetSkillEffectDam(effectID uint32) float32
	IsWater(origin linmath.Vector3, waterlevel float32) (bool, error)
}

// RoomChracter 场景中的角色
type RoomChracter struct {
	space.Entity

	kill     uint32
	killer   uint64
	rolebase excel.RoleData

	*MainPack
	*StateM

	nosafetime int64
	bombtime   int64

	curWatchOrder       uint32   //当前观战次序
	watchingTarget      uint64   //正在观战的目标
	watchers            []uint64 //观战者列表
	disabledWatchTarget uint64   //失效的观战目标

	bVictory bool //是否获胜
}

// GetKillNum 获取击杀数
func (rc *RoomChracter) GetKillNum() uint32 {
	return rc.kill
}

// IncrKillNum 增加击杀数
func (rc *RoomChracter) IncrKillNum() {
	rc.kill++
}

// GetKiller 获取击杀者
func (rc *RoomChracter) GetKiller() uint64 {
	return rc.killer
}

// SetKiller 设置击杀者
func (rc *RoomChracter) SetKiller(killer uint64) {
	rc.killer = killer
}

// SetVictory 设置获胜
func (rc *RoomChracter) SetVictory() {
	rc.bVictory = true
}

// GetRivalGamerType 获取对手的类型
func (rc *RoomChracter) GetRivalGamerType() uint32 {
	space := rc.GetSpace().(*Scene)
	if space == nil {
		return RoomUserTypeNormal
	}

	user, _ := space.GetEntity(rc.GetID()).(IRoomChracter)
	if user == nil {
		return RoomUserTypeNormal
	}

	switch user.GetGamerType() {
	case RoomUserTypeBlue:
		return RoomUserTypeRed
	case RoomUserTypeRed:
		return RoomUserTypeBlue
	}

	return RoomUserTypeNormal
}

// SetBaseState 设置 BaseState的值
func (rc *RoomChracter) SetBaseState(bs byte) {
	space := rc.GetSpace().(*Scene)
	if space == nil {
		return
	}

	user, _ := space.GetEntity(rc.GetID()).(*RoomUser)
	if user == nil {
		return
	}

	user.SetBaseState(RoomPlayerBaseState_Watch)
}

// SetRoleBase 设置角色
func (rc *RoomChracter) SetRoleBase(roleID uint32) {
	role, ok := excel.GetRole(uint64(roleID))
	if !ok {
		return
	}
	rc.rolebase = role
}

// GetRoleBaseSpeed 获取角色移动速度
func (rc *RoomChracter) GetRoleBaseSpeed() float32 {
	return rc.rolebase.Initmvspeed
}

// GetRoleSwimSpeed 获取角色游泳速度
func (rc *RoomChracter) GetRoleSwimSpeed() float32 {
	return rc.rolebase.Initswimspeed
}

// GetRoleBaseMaxHP 获取角色最大生命值
func (rc *RoomChracter) GetRoleBaseMaxHP() uint32 {
	return uint32(rc.rolebase.Inithp)
}

func (rc *RoomChracter) fillMapData() *protoMsg.ChracterMapDataInfo {
	info := &protoMsg.ChracterMapDataInfo{}
	info.Uid = rc.GetDBID()
	// piu := db.PlayerInfoUtil(rc.GetDBID())
	// info.Name_ = piu.GetName()
	info.Level = 1

	pos := &protoMsg.Vector3{
		X: rc.GetPos().X,
		Y: rc.GetPos().Y,
		Z: rc.GetPos().Z,
	}

	info.Pos = pos

	for _, v := range rc.equips {
		if v.thisid == rc.useweapon {
			info.Weapon = v.fillInfo()
		} else {
			info.Secweapon = v.fillInfo()
		}
	}

	// for _, v := range rc.armors {
	// 	info.Armors = append(info.Armors, v.GetBaseID())
	// }

	return info
}

//Loop 定时执行
func (rc *RoomChracter) Loop() {
	space := rc.GetSpace().(*Scene)
	shrinkdam := GetRefreshZoneMgr(space).shrinkdam
	// log.Info("计算距离", user.GetPos().Distance(space.safecenter), user.GetPos(), space.safecenter, space.saferadius)
	if shrinkdam != 0 && common.Distance(rc.GetPos(), space.safecenter) > space.saferadius {
		if time.Now().Unix() > rc.nosafetime+5 {
			rc.nosafetime = time.Now().Unix()

			defender, ok := rc.GetRealPtr().(iDefender)

			if ok {
				defender.DisposeSubHp(InjuredInfo{num: shrinkdam, injuredType: mephitis, isHeadshot: false})
			}
		}
	}

	if rc.StateM != nil {
		rc.StateM.Update()
	}
}

func (rc *RoomChracter) computeAttack(gunid uint32, targetPos linmath.Vector3, headshot bool) uint32 {
	base, ok := excel.GetGun(uint64(gunid))
	if !ok {
		return 0
	}
	distance := common.Distance(rc.GetPos(), targetPos)
	reducedistance := base.Reducedistance
	hurt := base.Hurt
	if headshot {
		hurt = base.Headhurt
	}

	if distance == 0 || reducedistance == 0 {
		return uint32(hurt)
	}

	if distance >= base.Distance {
		//log.Warn("伤害距离", distance, base.Distance, rc.GetPos(), targetPos)
		return 0
	}

	coeff := distance / reducedistance
	r := float32(math.Pow(float64(coeff), -0.8))
	if r >= 1 {
		return uint32(hurt)
	}

	// log.Debug("伤害计算", distance, reducedistance, r, hurt)
	return uint32(hurt * r)
}

//isOnBridge 是否在桥上
func (rc *RoomChracter) isOnBridge(pos linmath.Vector3) bool {
	space := rc.GetSpace().(*Scene)
	origin := linmath.Vector3{
		X: pos.X,
		Y: pos.Y + 1.6,
		Z: pos.Z,
	}
	direction := linmath.Vector3{
		X: 0,
		Y: -1,
		Z: 0,
	}
	// dist, pos, _, hit, _ := space.Raycast(origin, direction, 100, unityLayerBuilding)
	_, _, _, hit, _ := space.Raycast(origin, direction, 50, unityLayerBuilding)
	if hit {
		// log.Error("检测到建筑物")
		return true
	}

	return false
}

func (rc *RoomChracter) GetEquips() map[uint64]*Item {
	return rc.equips
}

// GetWatchingTarget 获取正在观战的目标
func (rc *RoomChracter) GetWatchingTarget() uint64 {
	return rc.watchingTarget
}

// GetWatchers 获取观战者
func (rc *RoomChracter) GetWatchers() []uint64 {
	return rc.watchers
}

// IsWatching 是否正在观战
func (rc *RoomChracter) IsWatching() bool {
	return rc.watchingTarget != 0
}

// IsBeingWatched 是否正在被观战
func (rc *RoomChracter) IsBeingWatched() bool {
	return len(rc.watchers) > 0
}

// DisposeWatch 处理观战
func (rc *RoomChracter) DisposeWatch(target uint64) {
	space := rc.GetSpace().(*Scene)
	if space == nil {
		return
	}

	targetUser, _ := space.GetEntity(target).(IRoomChracter)
	if targetUser == nil {
		return
	}

	if rc.IsWatching() {
		rc.CancelWatch()
	}

	// 建立观战关系
	rc.watchingTarget = target
	rc.disabledWatchTarget = 0
	targetUser.AddWatcher(rc.GetID())

	rc.SetBaseState(RoomPlayerBaseState_Watch)
	rc.SetPos(targetUser.GetPos())

	// if !space.teamMgr.IsInOneTeamByID(rc.GetID(), target) {
	// 	for _, v := range targetUser.GetTeamMembers() {
	// 		if targetUser.isAI() {
	// 			ai, _ := space.GetEntity(v).(*RoomAI)
	// 			if ai != nil {
	// 				rc.AddExtWatchEntity(ai)
	// 				ai.AddExtWatchEntity(rc)
	// 			}
	// 		} else {
	// 			realUser, _ := space.GetEntity(v).(*RoomUser)
	// 			if realUser != nil {
	// 				rc.AddExtWatchEntity(realUser)
	// 				realUser.AddExtWatchEntity(rc)
	// 			}
	// 		}
	// 	}
	// }

	rc.Info("DisposeWatch success, target: ", target)
}

// CancelWatch 取消观战
func (rc *RoomChracter) CancelWatch() {
	space := rc.GetSpace().(*Scene)
	if space == nil {
		return
	}

	target := rc.watchingTarget
	if target != 0 {
		targetUser, _ := space.GetEntity(target).(IRoomChracter)
		if targetUser != nil {
			targetUser.RemoveWatcher(rc.GetID())
		}
	}

	rc.watchingTarget = 0
}

// AddWatcher 添加观战者
func (rc *RoomChracter) AddWatcher(id uint64) {
	for _, j := range rc.watchers {
		if j == id {
			return
		}
	}

	rc.watchers = append(rc.watchers, id)

	// 通知角色被观战
	if len(rc.watchers) == 1 {
		rc.RPC(iserver.ServerTypeClient, "BeWatched")
	}
}

// RemoveWatcher 删除观战者
func (rc *RoomChracter) RemoveWatcher(id uint64) {
	for i, j := range rc.watchers {
		if j == id {
			rc.watchers = append(rc.watchers[0:i], rc.watchers[i+1:]...)
			break
		}
	}

	// 通知玩家被观战结束
	if len(rc.watchers) == 0 {
		rc.RPC(iserver.ServerTypeClient, "TerminateBeWatched")
	}
}

// BroadcastToWatchers 发送广播消息给观战者
// userType为0时表示通知给所有的观战者
// userType为1时表示仅通知场景内的观战者
// userType为2时表示仅通知场景外的观战者
func (rc *RoomChracter) BroadcastToWatchers(userType uint32, method string, args ...interface{}) {
	space := rc.GetSpace().(*Scene)
	if space == nil {
		return
	}

	for _, id := range rc.watchers {
		user, _ := space.GetEntity(id).(IRoomChracter)
		if user == nil {
			continue
		}

		if userType != 0 && user.GetUserType() != userType {
			continue
		}

		user.RPC(iserver.ServerTypeClient, method, args...)
	}
}

// TerminateWatch 角色获胜或死亡，终止观战
func (rc *RoomChracter) TerminateWatch() {
	space := rc.GetSpace().(*Scene)
	if space == nil {
		return
	}

	tmpWatchers := make([]uint64, len(rc.watchers))
	copy(tmpWatchers, rc.watchers)

	for _, id := range tmpWatchers {
		user, _ := space.GetEntity(id).(IRoomChracter)
		if user == nil {
			continue
		}

		user.CancelWatch()

		user.SetKiller(rc.GetKiller())
		user.SetDisabledWatchTarget(rc.GetID())

		if !rc.bVictory {
			user.RPC(iserver.ServerTypeClient, "WatchTargetDieNotify", rc.GetID())
		}
	}
}

// SyncWatchTargetInfo 向客户端同步观战目标信息
func (rc *RoomChracter) SyncWatchTargetInfo() {
	space := rc.GetSpace().(*Scene)
	if space == nil {
		return
	}

	if rc.disabledWatchTarget != 0 {
		rc.RPC(iserver.ServerTypeClient, "WatchTargetDieNotify", rc.disabledWatchTarget)
	}

	target := rc.watchingTarget
	if target == 0 {
		return
	}

	targetUser, _ := space.GetEntity(target).(IRoomChracter)
	if targetUser == nil {
		return
	}

	color := common.GetPlayerNameColor(targetUser.GetDBID())
	rc.RPC(iserver.ServerTypeClient, "SyncWatchTarget", target, targetUser.GetDBID(), targetUser.GetName(), targetUser.GetUserTeamID(), color)
}

// SetDisabledWatchTarget 记录失效的观战目标
func (rc *RoomChracter) SetDisabledWatchTarget(target uint64) {
	rc.disabledWatchTarget = target
}

// GetDisabledWatchTarget 获取失效的观战目标
func (rc *RoomChracter) GetDisabledWatchTarget() uint64 {
	return rc.disabledWatchTarget
}

// IsWater 判断坐标点是否是水域
func (rc *RoomChracter) IsWater(origin linmath.Vector3, waterlevel float32) (bool, error) {
	space := rc.GetSpace().(*Scene)
	if space == nil {
		return false, nil
	}

	origin.Y += 1.5
	direction := linmath.Vector3{
		X: 0,
		Y: -1,
		Z: 0,
	}

	_, pos, _, hit, err := space.Raycast(origin, direction, 2000, 1<<12|1<<16)
	if !hit {
		return false, err
	}

	return pos.Y <= waterlevel, nil
}

package main

import (
	"common"
	"entitydef"
	"errors"
	"excel"
	"math"
	"math/rand"
	"protoMsg"
	"time"
	"zeus/iserver"
	"zeus/linmath"
	"zeus/space"

	log "github.com/cihub/seelog"
	"github.com/spf13/viper"
)

// RoomUser 大厅玩家
type RoomUser struct {
	entitydef.PlayerDef
	RoomChracter

	stateMgr *UserStateMgr
	*effectM
	energy *Energy

	gm      *GmMgr
	sumData *SumData

	headshotnum uint32
	effectharm  uint32

	stateLastTime map[byte]map[byte]uint32

	notjumppos      linmath.Vector3
	needcheckjump   bool
	safeCheckHeight bool

	haveenter bool
	*SafeCheckM

	teamid           uint64
	stateSyncTime    uint32 //玩家状态同步的帧数
	shotGunToWillDie uint32 //散弹枪导致玩家濒死状态的帧数(来源attackReq)

	parachuteCtrl *ParachuteCtrl // 玩家掉线后自动控制跳伞过程
	parachuteType uint8          //跳伞类型 0:正常跳伞 1:被跟随 2:跟随 3:跟随过程中脱离跟随

	_offline  bool // true:玩家断线
	dieNotify *protoMsg.DieNotifyRet

	multikill    uint32
	multikillend int64
	medalhave    *protoMsg.MedalDataList
	killdrop     map[uint64]*protoMsg.KillDrop

	first    bool   // 开始统计跑动距离开关
	insignia string //当前使用中的勋章图标

	addFuelTimer  int    //定时器句柄，用于给载具加燃料倒计时
	addFuelThisid uint64 //给载具加燃料时，使用的油桶的thisid

	leaveSceneTime  int //结算后离开场景的倒计时时间
	leaveSceneTimer int //定时器句柄，用于结算后离开场景倒计时

	reportCheck map[uint64]bool //举报标记

	*SkillMgr
	secondTicker *time.Ticker

	userType      uint32          //玩家的来源类型 1 游戏玩家 2 观战玩家
	watchVoice    uint64          //单人玩家被观战时的语音房间
	voiceMemberId int32           //观战时语音房间内的编号
	signPos       linmath.Vector2 //单人地图标记位置
	watchLoading  uint32          //离开观战是否加载完成
}

// Init 初始化调用
func (user *RoomUser) Init(initParam interface{}) {
	user.RegMsgProc(&RoomUserMsgProc{user: user}) //注册消息处理对象

	//加载数据
	user.loadData()

	user.SetWatcher(true)

	user.nosafetime = 0
	user.bombtime = 0
	user.stateLastTime = make(map[byte]map[byte]uint32, 0)
	user.leaveSceneTime = int(common.GetTBSystemValue(common.System_LeaveSceneTime))

	if user.GetRoleModel() == 0 {
		user.SetRoleModel(uint32(common.GetTBSystemValue(32)))
		user.SetGoodsRoleModel(uint32(common.GetTBSystemValue(32)))
	}
	user.SetRoleBase(1)
	user.SetSpeedRate(1.0)
	//user.SetSpeed(user.GetRoleBaseSpeed())
	user.SetGunSight(0)

	maxHP := user.GetRoleBaseMaxHP()
	user.SetMaxHP(maxHP)
	user.SetHP(maxHP)
	user.SetChracterMapDataInfo(user.fillMapData())
	user.SetChracterMapDataInfoDirty()

	width := rand.Float32() * user.GetSpace().(*Scene).mapdata.Width
	height := rand.Float32() * user.GetSpace().(*Scene).mapdata.Height

	user.SetPos(linmath.NewVector3(width, 500, height))
	user.SetRota(linmath.Vector3{})
	user.stateMgr = NewUserStateMgr(user)
	user.sumData = NewSumData(user)
	user.GetEntities().RegTimerByObj(user, user.stateMgr.Loop, 500*time.Millisecond)
	user.GetEntities().RegTimerByObj(user, user.sumData.RunDistance, 1*time.Second)  //跑动距离统计
	user.GetEntities().RegTimerByObj(user, user.sumData.CarDistance, 1*time.Second)  //载具距离统计
	user.GetEntities().RegTimerByObj(user, user.sumData.SwimDistance, 1*time.Second) //游泳距离统计

	user.SetVehicleProp(&protoMsg.VehicleProp{})

	packprop := &protoMsg.BackPackProp{}
	system, ok := excel.GetSystem(uint64(common.System_InitBagPack))
	if ok {
		packprop.Baseid = uint32(system.Value)
	}
	user.SetBackPackProp(packprop)

	user.SetHeadProp(&protoMsg.HeadProp{})
	user.SetBodyProp(&protoMsg.BodyProp{})
	user.SetIsWearingGilley(0)
	user.SetSkillEffectList(&protoMsg.SkillEffectList{})

	user.safeCheckHeight = viper.GetBool("Room.SafeCheckHeight")
	user.insignia = common.GetInsigniaIcon(user.GetDBID())

	user.notifyAirlineInfo()
	//user.SetState(UserStateLoadingMap)
	//user.SetActionState(0)
	user.gm = NewGmMgr(user)

	user.SetBaseState(RoomPlayerBaseState_LoadingMap)

	user.headshotnum = 0
	user.effectharm = 0
	user.curWatchOrder = 0
	user.first = true
	user.reportCheck = make(map[uint64]bool) //举报标记

	user.teamid = 0
	user.RPC(common.ServerTypeLobby, "EnterRoomSuccess")

	user.SendAllObj()

	user.SkillMgr = initSkillMgr(user)
	user.secondTicker = time.NewTicker(1 * time.Second)
	user.syncCurRoleSkillInfo() // 通知当前使用的技能
	user.SkillMgr.putPassiveSkill()
}

func (user *RoomUser) loadData() {
	user.MainPack = InitMainPack(user)
	user.effectM = initEffectM(user)
	user.StateM = InitStateM(user)
	user.energy = InitEnergy(user)
	user.SafeCheckM = InitSafeCheckM(user)
	user.medalhave = &protoMsg.MedalDataList{}
	user.killdrop = make(map[uint64]*protoMsg.KillDrop)
}

// Destroy 析构时调用
func (user *RoomUser) Destroy() {
	user.GetEntities().UnregTimerByObj(user)
	user.GetEntities().RemoveDelayCall(user.addFuelTimer)
	user.GetEntities().RemoveDelayCall(user.leaveSceneTimer)

	space := user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	if space.teamMgr.isTeam && user.teamid != 0 {
		teamid := space.teamMgr.GetTeamIDByPlayerID(user.GetDBID())
		space.teamMgr.TeammateLeaveSpace(user, teamid)
	}
	user.secondTicker.Stop()

	user.Info("Room user destroy")
}

// OnEnterSpace 玩家进入地图
func (user *RoomUser) OnEnterSpace() {

}

// RoomUser 创建成功后进入
func (user *RoomUser) enterBattle() {
	space := user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	space.members[user.GetID()] = true
	space.updateTotalNum(user)

	user.haveenter = true
	user.userType = RoomUserTypePlayer
	user.RPC(iserver.ServerTypeClient, "UpdateKillNum", uint32(user.GetKillNum()))

	// space.BroadAliveNum()
	space.BroadAirLeft()
	user.Info("Room user enter space")
}

func (user *RoomUser) watcherLeave() {
	space := user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	delete(space.watchers, user.GetID())
	space.UpdateWatchNum(user.watchingTarget, -1, true)

	targetUser, ok := space.GetEntity(user.watchingTarget).(*RoomUser)
	if ok {
		targetUser.RemoveExtWatchEntity(user)
	}

	user.CancelWatch()
	user.RPC(common.ServerTypeLobby, "WatchFlow", space.GetID(), uint32(time.Now().Unix()-user.sumData.watchStartTime), user.watchLoading, uint32(0))

	user.sumData.watchStartTime = 0
	user.watchLoading = 0

	user.RPC(common.ServerTypeLobby, "LeaveWatchTarget")
	user.Info("Watch end")
}

// OnLeaveSpace 玩家离开地图
func (user *RoomUser) OnLeaveSpace() {
	space := user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	if user.userType == RoomUserTypeWatcher {
		user.watcherLeave()
		return
	}

	if _, ok := space.members[user.GetID()]; ok {
		user.userLeave() //玩家的entity 由底层删除时额外清理数据
		if space.teamMgr.isTeam {
			teammate := space.teamMgr.getAliveTeammate(user)
			if teammate == 0 {
				space.teamMgr.DisposeTeamSettle(user.GetUserTeamID(), false)
			}
		}
	}
	user.tlogRoundFlow()
	user.tlogBattleResult()
	user.tlogSecGameEndFlow()
	user.tlogSkillUseFlow()
	// 通知推荐好友列表
	user.recommendStranger()

	//大厅服角色离开地图
	user.RPC(common.ServerTypeLobby, "leaveSpace")
	user.Info("Room user leave space")
}

//获取玩家本局比赛的匹配类型
func (user *RoomUser) GetMatchTyp() uint8 {
	switch user.sumData.battletype {
	case 0:
		return 1
	case 1:
		return 2
	case 2:
		return 4
	}
	return 0
}

// GetTeamMembers 获取队友信息
func (user *RoomUser) GetTeamMembers() []uint64 {
	res := []uint64{}

	space := user.GetSpace().(*Scene)
	if space == nil {
		return res
	}

	if space.teamMgr.isTeam {
		for _, v := range space.teamMgr.teams[user.GetUserTeamID()] {
			mem, ok := space.GetEntityByDBID("Player", v).(*RoomUser)
			if !ok || mem.userType == RoomUserTypeWatcher {
				continue
			}

			res = append(res, mem.GetID())
		}
	} else {
		res = append(res, user.GetID())
	}

	return res
}

// GetAttack 获取攻击值
func (user *RoomUser) GetAttack(targetPos linmath.Vector3, headshot bool) uint32 {
	if user.isMeleeWeaponUse() {
		useweapon := user.GetMeleeWeaponUse()
		base, ok := excel.GetGun(uint64(useweapon))
		if !ok {
			return 1
		}

		attack := uint32(base.Hurt * user.SkillMgr.getAddTypeDam(SE_WeaponDam))
		user.Debug("GetAttack Hurt:", base.Hurt, " attack:", attack)

		return attack
	}

	for _, v := range user.equips {
		if v.thisid == user.useweapon {
			attack := user.computeAttack(v.GetBaseID(), targetPos, headshot)
			attack = uint32(float32(attack) * user.SkillMgr.getAddTypeDam(SE_WeaponDam))
			user.Debug("GetAttack attack:", attack)

			return attack
		}
	}

	return 1
}

// CanAttackAndConsume 能否攻击
func (user *RoomUser) CanAttackAndConsume(firetime uint32) bool {
	if user.isMeleeWeaponUse() {
		return true
	}

	bulletproto := &protoMsg.RefreshGunBulletNotify{}
	for _, v := range user.equips {
		if v.thisid == user.useweapon {
			if v.base.Type == ItemTypeShotGun {
				gunConfig, ok := excel.GetGun(uint64(v.GetBaseID()))
				if !ok {
					return false
				}

				num, ok := user.shotgunbullet[v.thisid]
				if !ok {
					if v.bullet > 0 {
						v.bullet--
						user.sumData.IncrShotNum() //开枪次数自增

						user.shotgunbullet[v.thisid] = 1
					} else {
						return false
					}
				} else {
					num++
					if num == gunConfig.Bulletnum {
						delete(user.shotgunbullet, v.thisid)
					} else {
						user.shotgunbullet[v.thisid] = num
					}
				}
			} else {
				if v.bullet > 0 {
					v.bullet--
					user.sumData.IncrShotNum() //开枪次数自增
				} else {
					return false
				}
			}

			bulletproto.Thisid = v.thisid
			bulletproto.Bullet = v.bullet
			user.SetChracterMapDataInfo(user.fillMapData())
			user.SetChracterMapDataInfoDirty()
			// user.CastMsgToMe(bulletproto)
			return true
		}
	}

	return false
}

func (user *RoomUser) CanFakeAttackAndConsume() bool {
	if user.isMeleeWeaponUse() {
		return true
	}

	for _, v := range user.equips {
		if v.thisid == user.useweapon {

			if v.fakebullet > 0 {
				v.fakebullet--
			} else {
				return false
			}

			return true
		}
	}

	return false
}

// SetWatchBattle 设置观战
func (user *RoomUser) SetWatchBattle() bool {
	// 判断是否有好友在战斗中
	space := user.GetSpace().(*Scene)
	teammate := space.teamMgr.getAliveTeammate(user)

	if teammate != 0 {
		user.stateMgr.SetState(RoomPlayerBaseState_Watch)
		return true
	}

	return false
}

// Death 玩家死亡
func (user *RoomUser) Death(injuredType uint32, attackid uint64) {

}

// DisposeAddHp 增加血量
func (user *RoomUser) DisposeAddHp(num uint32) {

	curState := user.stateMgr.GetState()
	if curState == RoomPlayerBaseState_WillDie || curState == RoomPlayerBaseState_Watch || curState == RoomPlayerBaseState_Dead {
		user.Warn("DisposeAddHp failed, curState: ", curState)
		return
	}

	user.AddHp(num)
	user.Debug("Dispose add hp success, num: ", num)
}

// AddHp 增加血量
func (user *RoomUser) AddHp(num uint32) {
	curHP := user.GetHP()
	maxHP := user.GetMaxHP()
	if curHP >= maxHP {
		return
	}
	if curHP+num > maxHP {
		num = maxHP - curHP
	}
	user.SetHP(curHP + num)
	user.sumData.AddRecoverHp(num) //恢复血量统计
	user.Debug("Add hp num: ", num)
}

const (
	// SubHpNormal 正常减血
	SubHpNormal = 1

	// SubHpWillDie 减血后濒死
	SubHpWillDie = 2

	// SubHpDie 减血后死亡
	SubHpDie = 3

	// WillDieSubHp 濒死状态下减血
	WillDieSubHp = 4

	// SubHpError 异常减血
	SubHpError = 5

	// WillDieSubHpToWatch 濒死状态下减血观战
	WillDieSubHpToWatch = 6

	// WillDieSubHpToDead 濒死状态下减血死亡
	WillDieSubHpToDead = 7
)

// InjuredInfo 受伤害信息
type InjuredInfo struct {
	num uint32 // 受伤害值

	injuredType uint32 // 受伤害类型
	isHeadshot  bool   // 是否爆头

	attackid   uint64 // 攻击者entityid
	attackdbid uint64 // 攻击者DBid

	killDownInjured bool   // 本次伤害击杀了濒死玩家
	effectHarm      uint32 // 真实伤害
}

// DisposeSubHp 减少血量
func (user *RoomUser) DisposeSubHp(injuredInfo InjuredInfo) {
	if user.userType == RoomUserTypeWatcher {
		return
	}

	//毒气伤害
	if injuredInfo.injuredType == mephitis {
		per := user.SkillMgr.getReduceTypeDam(SE_ShrinkDam)
		injuredInfo.num = uint32(math.Ceil(float64(injuredInfo.num) * float64(per)))
		user.Debug("DisposeSubHp per:", per, " injuredInfo.num:", injuredInfo.num)
	}

	damReducePre := user.SkillMgr.getReduceTypeDam(SE_DamReduce) // 自身伤害削减百分比(效果13)
	vehicleDamReducePre := user.GetVehicleDamReduce()            // 载具内玩家伤害降低百分比(效果18)
	injuredInfo.num = uint32(float32(injuredInfo.num) * damReducePre * vehicleDamReducePre)

	space := user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	if injuredInfo.num == 0 {
		return
	}

	if injuredInfo.injuredType == falldam {
		user.sumData.fallDamage += injuredInfo.num
	}

	result := user.TmpSubHp(injuredInfo.num, injuredInfo.attackid, &injuredInfo)

	if result == SubHpDie || result == WillDieSubHpToWatch || result == WillDieSubHpToDead {
		user.sumData.DisposeAttackNum(injuredInfo.attackid) //助攻次数统计
		space.BroadAliveNum()
	}

	if result == SubHpWillDie || result == SubHpDie || result == WillDieSubHpToWatch || result == WillDieSubHpToDead {
		user.BreakEffect(true)
		user.energy.clearEnergy()
		if result == WillDieSubHpToWatch || result == WillDieSubHpToDead {
			injuredInfo.killDownInjured = true //濒死状态被杀死 死亡通知特殊处理
		}
		space.BroadDieNotify(injuredInfo.attackid, user.GetID(), false, injuredInfo)
	}

	if result == SubHpDie {
		user.DisposeSettle()
	}

	attacker, _ := space.GetEntity(injuredInfo.attackid).(IRoomChracter)
	if attacker != nil {
		if !attacker.isAI() {
			attackUser, _ := space.GetEntity(injuredInfo.attackid).(*RoomUser)
			if attackUser != nil {
				attackUser.DisposeAttackResult(&injuredInfo, result, user.GetID())
			}
		}
	}

	user.Info("Dispose sub hp success, result: ", result, " harmValue: ", injuredInfo.num, " injuredtype: ", injuredInfo.injuredType, " attackid: ", injuredInfo.attackid, " userState: ", user.stateMgr.GetState())
}

// DisposeAttackResult 处理攻击结果
func (user *RoomUser) DisposeAttackResult(injuredInfo *InjuredInfo, result uint32, defendid uint64) {
	space := user.GetSpace().(*Scene)
	if space == nil {
		return
	}
	defender, ok := space.GetEntity(defendid).(*RoomUser)

	// 统计枪支信息
	if injuredInfo.injuredType == gunAttack || injuredInfo.injuredType == headShotAttack {
		var killDistance float32
		if ok {
			killDistance = common.Distance(user.GetPos(), defender.GetPos())
		}
		user.tlogGunFlow(injuredInfo.num, result, injuredInfo.isHeadshot, killDistance)
	}
	if ok {
		if !(space.teamMgr.IsInOneTeam(defender.GetDBID(), user.GetDBID())) {
			if (result == SubHpDie && injuredInfo.isHeadshot) || //直接爆头击杀
				// 爆头击倒的敌人死亡
				((result == WillDieSubHpToWatch || result == WillDieSubHpToDead) && defender.stateMgr.isHeadShot && user.GetDBID() == defender.stateMgr.downAttacker) {
				user.IncrHeadShotNum()
			}
			if user.GetID() != defendid {
				user.AddEffectHarm(injuredInfo.effectHarm) //累积有效伤害
			}
		}
	} else {
		if (result == SubHpDie && injuredInfo.isHeadshot) || //直接爆头击杀
			// 爆头击倒的敌人死亡
			((result == WillDieSubHpToWatch || result == WillDieSubHpToDead) && defender.stateMgr.isHeadShot && user.GetDBID() == defender.stateMgr.downAttacker) {
			user.IncrHeadShotNum()
		}
		if user.GetID() != defendid {
			user.AddEffectHarm(injuredInfo.effectHarm) //累积有效伤害
		}
	}

	// 同步杀人数
	if result == SubHpDie || result == WillDieSubHpToWatch || result == WillDieSubHpToDead {

		if ok {
			if !(space.teamMgr.IsInOneTeam(defender.GetDBID(), user.GetDBID()) || user.GetID() == defendid) {
				user.DisposeIncrKillNum()
				user.sumData.KillInfo(injuredInfo.injuredType) //玩家击杀他人时的数据统计 此处的injuredType只有 gunAttack headShotAttack
			}
		} else {
			user.DisposeIncrKillNum()
			user.sumData.KillInfo(injuredInfo.injuredType) //玩家击杀他人时的数据统计 此处的injuredType只有 gunAttack headShotAttack
		}

		user.sumData.KillDistance(defendid) //最高击杀距离统计
		user.sumData.MaxChainKill()         //统计最大连杀次数
	}

	if result == SubHpWillDie || result == SubHpNormal || result == WillDieSubHp {
		user.sumData.DisposeAttackNumMap(defendid)
	}
}

// TmpSubHp 处理减血逻辑
func (user *RoomUser) TmpSubHp(num uint32, attackid uint64, injuredInfo *InjuredInfo) uint32 {

	curHp := user.GetHP()
	if curHp == 0 {

		if user.stateMgr.GetState() == RoomPlayerBaseState_WillDie {
			downHp := user.stateMgr.downEndStamp - time.Now().Unix()
			injuredInfo.effectHarm = num
			if downHp < int64(num) {
				injuredInfo.effectHarm = uint32(downHp)
			}
			user.SetKiller(attackid)
			user.stateMgr.amendDownTime(attackid, injuredInfo.injuredType, -int64(num))
			user.CastRPCToAllClient("OnDamage", uint64(injuredInfo.injuredType))
			curState := user.stateMgr.GetState()
			if curState == RoomPlayerBaseState_Watch {
				return WillDieSubHpToWatch
			} else if curState == RoomPlayerBaseState_Dead {
				return WillDieSubHpToDead
			}

			return WillDieSubHp
		}

		user.Debug("TmpSubHp failed, hp: ", user.GetHP(), " num: ", num, " attackid: ", attackid, "injuredType: ", injuredInfo.injuredType, " state: ", user.stateMgr.GetState())
		return SubHpError
	}

	if curHp > num {
		user.SetKiller(0)
		user.SetHP(curHp - num)
		user.CastRPCToAllClient("OnDamage", uint64(injuredInfo.injuredType))
		injuredInfo.effectHarm = num
		return SubHpNormal
	}

	user.SetKiller(attackid)
	injuredInfo.effectHarm = curHp
	user.SetHP(0)
	user.CastRPCToAllClient("OnDamage", uint64(injuredInfo.injuredType))

	onvehicle := false
	if user.GetBaseState() == RoomPlayerBaseState_Ride || user.GetBaseState() == RoomPlayerBaseState_Passenger {
		onvehicle = true
	}
	user.deathOnVehicle()

	// 中断救援队友
	user.stateMgr.DisposeRescue(false)

	// 是否可以设置为救援状态
	if user.stateMgr.setDownState(attackid, injuredInfo.injuredType, injuredInfo.isHeadshot) {
		return SubHpWillDie
	}

	//在驾驶状态死亡不重置坐标
	space := user.GetSpace().(*Scene)
	if !onvehicle && space != nil {
		pos := user.GetPos()
		pos = getCanPutHeight(space, pos)
		user.SetPos(pos)
	}

	user.DisposeDeath(injuredInfo.injuredType, attackid, false)

	//	user.Death(injuredType)
	return SubHpDie
}

// UnderAttack 玩家被攻击
func (user *RoomUser) UnderAttack(entityID uint64) {

}

func (user *RoomUser) reduceDam(attack uint32, headshot bool) uint32 {
	var ret uint32
	if headshot {
		prop := user.GetHeadProp()
		if prop != nil {
			if prop.Reducedam > attack {
				prop.Reducedam -= attack
				ret += attack
			} else {
				ret += prop.Reducedam
				prop.Baseid = 0
				prop.Itemid = 0
				prop.Reducedam = 0
			}
		}
		user.SetHeadProp(prop)
		user.SetHeadPropDirty()
	} else {
		prop := user.GetBodyProp()
		if prop != nil {
			if prop.Reducedam > attack {
				prop.Reducedam -= attack
				ret += attack
			} else {
				ret += prop.Reducedam
				prop.Baseid = 0
				prop.Reducedam = 0
			}
		}
		user.SetBodyProp(prop)
		user.SetBodyPropDirty()
	}

	return ret
}

// GetUserTeamID 获取玩家队伍id
func (user *RoomUser) GetUserTeamID() uint64 {
	return user.teamid
}

// GetUserType 获取玩家的类型
func (user *RoomUser) GetUserType() uint32 {
	return user.userType
}

func (user *RoomUser) isAI() bool {
	return false
}

// IsDead 是否已经死亡
func (user *RoomUser) IsDead() bool {
	state := user.GetBaseState()
	return state == RoomPlayerBaseState_Dead || state == RoomPlayerBaseState_Watch
}

// GetWatchTargetInTeam 在队伍内获取观战目标
func (user *RoomUser) GetWatchTargetInTeam() uint64 {
	space := user.GetSpace().(*Scene)
	if space == nil {
		return 0
	}

	return space.teamMgr.getWatchTargetInTeam(user)
}

// NotifyWatcherTeamMateLeave 通知观战者 被观战者队友离开
func (user *RoomUser) NotifyWatcherTeamMateLeave(leaver *RoomUser) {
	space := user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	for _, id := range user.watchers {
		watcher, ok := space.GetEntity(id).(*RoomUser)
		if !ok {
			continue
		}

		watcher.RemoveExtWatchEntity(leaver)
	}
}

// SendChat 发送聊天信息
func (user *RoomUser) SendChat(str string) {
	proto := &protoMsg.ChatNotify{}
	proto.Type = 1
	proto.Content = str

	user.RPC(iserver.ServerTypeClient, "ChatNotify", proto)
	/*
		user.CastMsgToMe(proto)
	*/
}

//SyncProgressBar 同步玩家进度条信息
func (user *RoomUser) SyncProgressBar(barType uint64, endTime uint64, durationTime uint64, surplusTime uint64) {
	user.RPC(iserver.ServerTypeClient, "SyncProgressBar", uint32(barType), endTime, uint32(durationTime), uint32(surplusTime))
	user.BroadcastToWatchers(RoomUserTypeWatcher, "SyncProgressBar", uint32(barType), endTime, uint32(durationTime), uint32(surplusTime))
}

//BreakProgressBar 打断玩家进度条信息
func (user *RoomUser) BreakProgressBar() {
	//user.RPC(iserver.ServerTypeClient, user.GetID(), "onActionStateRet", uint64(0), uint64(0))
	if user.GetActionState() == ActionRescue {
		user.SetActionState(0)
	}
	user.RPC(iserver.ServerTypeClient, "BreakProgressBar")
	user.BroadcastToWatchers(RoomUserTypeWatcher, "BreakProgressBar")
}

// LeaveScene 离开场景
func (user *RoomUser) LeaveScene() {
	space := user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	user.BroadcastToWatchers(RoomUserTypeWatcher, "WatchTargetLeave")
	if user.IsBeingWatched() {
		user.TerminateWatch()
	}

	if user.IsWatching() {
		user.CancelWatch()
	}

	delete(space.members, user.GetID())

	//离开地图
	user.LeaveSpace()
	user.Info("Room user leave scene")
}

// 通知客户端航线信息
func (user *RoomUser) notifyAirlineInfo() {
	user.notifyAirline("AirLine")
}

func (user *RoomUser) notifyAirline(name string) {
	space := user.GetSpace().(*Scene)
	start := space.airlineStart
	end := space.airlineEnd

	//user.Debug("Airline", start, end, user.GetName())
	if err := user.RPC(iserver.ServerTypeClient, name, float64(start.X), float64(start.Z), float64(end.X), float64(end.Z)); err != nil {
		user.Error("RPC AirLine err: ", err)
	}
}

// 根据当前飞行时间, 计算出角色所在的位置
func (user *RoomUser) doParachute(randpos bool) {
	scene := user.GetSpace().(*Scene)

	speed := float32(scene.mapdata.Fly_Speed)
	dur := float32(time.Now().Sub(scene.allowParachuteTime).Seconds())
	distance := speed * dur
	direction := scene.airlineEnd.Sub(scene.airlineStart)
	pos := scene.airlineStart.Add(direction.Mul(distance / direction.Len()))
	pos.Y = scene.mapdata.Parachute_height
	if pos.X < math.SmallestNonzeroFloat32 {
		pos.X = math.SmallestNonzeroFloat32
	}
	if pos.X > scene.mapdata.Width {
		pos.X = scene.mapdata.Width
	}
	if pos.Z < math.SmallestNonzeroFloat32 {
		pos.Z = math.SmallestNonzeroFloat32
	}
	if pos.Z > scene.mapdata.Height {
		pos.Z = scene.mapdata.Height
	}

	if randpos {
		pos = linmath.RandXZ(pos, 50.0)
		// user.Debug("强制跳伞时50米内随机坐标", pos)
	}
	if pos.X > scene.mapdata.Width {
		pos.X = scene.mapdata.Width - 1
	}
	if pos.Z > scene.mapdata.Height {
		pos.Z = scene.mapdata.Height - 1
	}
	user.SetPos(pos)
	user.SetBaseState(RoomPlayerBaseState_Glide)

	user.parachuteCtrl = NewParachuteCtrl(user)
	user.parachuteCtrl.StartGlide()

	//log.Debug(user.GetID(), "跳伞位置", pos, user)
}

// OnPosChange 玩家位置改变
func (user *RoomUser) OnPosChange(curPos, origPos linmath.Vector3) {
	if user.stateMgr == nil {
		return
	}

	//user.Debug("玩家位置改变 originPos:", origPos, " curPos:", curPos)
	user.effectM.BreakEffect(false)

	if user.stateMgr.isRescue {
		systemValue := float32(common.GetTBSystemValue(41)) / 100
		distance := user.GetPos().Sub(user.stateMgr.RescuePos).Len()

		if distance > systemValue {
			user.stateMgr.DisposeRescue(false)
			//user.Info("打断救援", user.GetDBID(), distance, systemValue)
		}

	} else if user.stateMgr.isBerescue {
		systemValue := float32(common.GetTBSystemValue(41)) / 100
		distance := user.GetPos().Sub(user.stateMgr.BeRescuePos).Len()
		if distance > systemValue {
			user.stateMgr.BreakBerescue()
			//user.Info("打断被救援", user.GetDBID(), distance, systemValue)
		}
	}

	if user.IsInAddFuelState() {
		car := user.getRidingCar()
		if car == nil {
			return
		}

		if common.Distance(user.GetPos(), car.lastConsumePos) >= 0.15 {
			user.BreakAddFuel(2)
		}
	}

}

//Loop 定时执行
func (user *RoomUser) Loop() {
	state := user.GetBaseState()
	if state == RoomPlayerBaseState_Dead || state == RoomPlayerBaseState_Watch {
		return
	}

	user.RoomChracter.Loop()

	if user.effectM != nil {
		user.effectM.Update()
	}

	if user.energy != nil {
		user.energy.update()
	}

	if user.parachuteCtrl != nil {
		user.parachuteCtrl.update()
	}

	select {
	case <-user.secondTicker.C:
		if user.SkillMgr != nil {
			user.SkillMgr.update()
		}
	default:
	}
}

// 在point点的radius半径范围内随机获取可以抵达并且不是水面的点
func (user *RoomUser) getRandPoint(point linmath.Vector3, radius float32) (linmath.Vector3, error) {

	tryTime := 0
	var pos linmath.Vector3
	waterLevel := user.GetSpace().(*Scene).mapdata.Water_height
	err := errors.New("getRandPoint fail")
	for {
		if tryTime > 15 {
			break
		}
		tryTime++

		pos = linmath.RandXZ(point, radius)
		pos.Y, err = user.GetSpace().GetHeight(pos.X, pos.Z)
		if err == nil && pos.Y > waterLevel {
			break
		}
	}

	return pos, err
}

// AdviceNotify 提示信息通知
func (user *RoomUser) AdviceNotify(notifyType uint32, id uint64) {
	if err := user.RPC(iserver.ServerTypeClient, "AdviceNotify", notifyType, id); err != nil {
		user.Error(err)
	}
}

// IncrHeadShotNum 当场爆头数自增
func (user *RoomUser) IncrHeadShotNum() {
	user.headshotnum++
}

// AddEffectHarm 计数
func (user *RoomUser) AddEffectHarm(effect uint32) {
	user.effectharm += effect
}

// GetLoginChannel 1: 微信  2: QQ 3: 游客
func (user *RoomUser) GetLoginChannel() uint32 {
	return user.GetPlayerLogin().LoginChannel
}

// CreateNewEntityState 创建一个新的状态快照，由框架层调用
func (user *RoomUser) CreateNewEntityState() space.IEntityState {
	return NewRoomPlayerState()
}

// DisposeIncrKillNum 处理击杀数量
func (user *RoomUser) DisposeIncrKillNum() {
	user.IncrKillNum()
	user.RPC(iserver.ServerTypeClient, "UpdateKillNum", uint32(user.GetKillNum()))

	user.DisposeMultiKill()

	proto := &protoMsg.ChatNotify{}
	proto.Type = 2
	proto.Content = "击败" + common.Uint64ToString(uint64(user.GetKillNum())) + "人"

	user.RPC(iserver.ServerTypeClient, "ChatNotify", proto)
	/*
		user.CastMsgToMe(proto)
	*/

	user.GetCallKillEffect() //每杀一人触发一个效果
}

// DisposeDeath 处理玩家死亡
func (user *RoomUser) DisposeDeath(injuredType uint32, attackid uint64, isWatch bool) {

	space := user.GetSpace().(*Scene)
	if space == nil {
		user.Error("DisposeDeath failed, can't get space")
		return
	}

	dropid, droppos := user.MainPack.OnDeath()
	if dropid != 0 {
		attacker, _ := space.GetEntity(attackid).(*RoomUser)
		if attacker != nil {
			attacker.addKillDrop(dropid, droppos)
		}
	}

	user.sumData.endgametime = time.Now().Unix()
	user.sumData.isGround = false                //停止统计跑动距离
	user.sumData.DeadType(injuredType, attackid) //死亡类型判断

	oldState := user.stateMgr.GetState()
	if oldState == RoomPlayerBaseState_Glide || oldState == RoomPlayerBaseState_Parachute {
		user.sumData.isParachuteDie = true
	}

	if time.Now().Unix()-user.sumData.landtime <= 10 {
		user.sumData.isLandDie = true
	}

	delete(space.members, user.GetID())
	space.BroadAliveNum()

	// 打断所有状态
	user.SetActionState(0)

	// 技能特效清空
	for _, v := range user.SkillMgr.initiveEffect {
		if v == nil {
			continue
		}

		skilleffect := GetSkillEffect(uint64(v.effectID))
		if skilleffect == nil {
			user.Error("不能获取技能效果", v.effectID)
			continue
		}
		skilleffect.End(user, v)
	}
	for _, v := range user.SkillMgr.passiveEffect {
		if v == nil {
			continue
		}

		skilleffect := GetSkillEffect(uint64(v.effectID))
		if skilleffect == nil {
			user.Error("不能获取技能效果", v.effectID)
			continue
		}
		skilleffect.End(user, v)
	}
	user.SetSkillEffectList(&protoMsg.SkillEffectList{})
	user.SetSkillEffectListDirty()

	// 设置玩家状态
	if isWatch {
		user.stateMgr.SetState(RoomPlayerBaseState_Watch)
	} else {
		user.stateMgr.SetState(RoomPlayerBaseState_Dead)
	}

	user.BroadcastToWatchers(RoomUserTypeWatcher, "WatchTargetLeave")
	if user.IsBeingWatched() {
		user.TerminateWatch()
	}

	// 显示死亡标记
	if space.teamMgr.isTeam {
		space.teamMgr.ShowDieFlag(user)
	}

	user.RPC(iserver.ServerTypeClient, "DeadNotify")

	newState := user.stateMgr.GetState()
	user.Info("Room user die,  oldState: ", oldState, " newState: ", newState)
}

// DisposeSettle 处理玩家结算
func (user *RoomUser) DisposeSettle() {
	if user.userType != RoomUserTypePlayer {
		return
	}

	space := user.GetSpace().(*Scene)
	if space == nil {
		user.Error("DisposeSettle failed, can't get space")
		return
	}

	teamid := user.GetUserTeamID()
	if space.teamMgr.isTeam {
		space.teamMgr.DisposeTeamSettle(teamid, user.bVictory)
	} else {
		user.NotifySettle()
	}

	user.Info("Dispose settle success: teamid: ", teamid, " bVictory: ", user.bVictory)
}

// NotifySettle 通知结算
func (user *RoomUser) NotifySettle() {
	if user.userType != RoomUserTypePlayer {
		return
	}

	space := user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	if !user.sumData.isSettble {
		return
	}

	// 通知客户端结算信息
	user.Settle()
	user.userLeave()

	if space.isEnd() {
		user.LeaveScene()
	} else {
		t := time.Duration(user.leaveSceneTime+10) * time.Second
		user.leaveSceneTimer = user.GetEntities().AddDelayCall(user.LeaveScene, t)
	}
}

//玩家离开场景的处理
func (user *RoomUser) userLeave() {
	space := user.GetSpace().(*Scene)
	if space == nil {
		user.Error("NotifySettle failed, can't get space")
		return
	}
	//GetSpaceResultMgr(space).UserLeave(user.GetID())
	delete(space.members, user.GetID())

	if user.IsFollowedParachute() || user.IsFollowingParachute() {
		space.teamMgr.CancelFollowParachute(user)
	}

	space.ISceneData.onDeath(user)
	space.BroadAliveNum()
	space.BroadAirLeft()

	if !space.isEnd() {
		user.autoDownVehicle()
		space.teamMgr.AddSpaceMemberInfo(user)
	}
	if user.sumData.isSettble {
		user.writeDataReds()
		user.sumData.isSettble = false
	}
}

//isInBuilding 是否在房子里面
func (user *RoomUser) isInBuilding() bool {
	space := user.GetSpace().(*Scene)
	origin := linmath.Vector3{
		X: user.GetPos().X,
		Y: user.GetPos().Y,
		Z: user.GetPos().Z,
	}
	direction := linmath.Vector3{
		X: 0,
		Y: 1,
		Z: 0,
	}
	// dist, pos, _, hit, _ := space.Raycast(origin, direction, 2000, unityLayerBuilding)
	_, _, _, hit, _ := space.Raycast(origin, direction, 2000, unityLayerBuilding)
	if hit {
		// user.Error("检测到建筑物", dist, pos)
		return true
	}

	// dist, pos, _, hit, _ = space.Raycast(origin, direction, 2000, unityLayerFurniture)
	_, _, _, hit, _ = space.Raycast(origin, direction, 2000, unityLayerFurniture)
	if hit {
		// user.Error("检测到家具", dist, pos)
		return true
	}

	return false
}

// HaveAchieved 玩家达成成就
func (user *RoomUser) HaveAchieved(id uint32) {
	user.RPC(common.ServerTypeLobby, "HaveAchieved", id)
}

// 通知玩家最近一局的推荐好友
func (user *RoomUser) recommendStranger() {
	space := user.GetSpace().(*Scene)
	if space == nil {
		return
	}
	var list []uint64
	//分值差最小的三人
	list = space.GetMemberScoreRoundMe(user.GetDBID(), 3)
	//分值最高的三人
	l := len(space.memberModeScore)
	var top int
	if l < 3 {
		top = 0
	} else {
		top = l - 3
	}
	for i := l - 1; i >= top; i-- {
		if space.memberModeScore[i].uid == user.GetDBID() {
			continue
		}
		list = append(list, space.memberModeScore[i].uid)
	}
	//队友
	if space.teamMgr.isTeam {
		team := space.teamMgr.teams[user.GetUserTeamID()]
		if team != nil {
			for _, mem := range team {
				if mem == user.GetDBID() {
					continue
				}
				list = append(list, mem)
			}
		}
	}
	if len(list) == 0 {
		return
	}
	//去重
	var args []interface{}
	for _, uid := range list {
		have := false
		for _, e := range args {
			if uid == e.(uint64) {
				have = true
				break
			}
		}
		if have {
			continue
		}
		args = append(args, uid)
	}
	user.RPC(common.ServerTypeLobby, "FriendRecommendList", args...)
}

func (user *RoomUser) SetDieNotify(proto *protoMsg.DieNotifyRet) {
	user.dieNotify = proto
}

func (user *RoomUser) MedalNotify(willdie bool, headshot bool, injuredType uint32) {
	if willdie {
		user.MedalNotifyRpc(uint32(4))
		return
	}

	if headshot {
		user.MedalNotifyRpc(uint32(3))
		return
	}

	if user.isMeleeWeaponUse() && injuredType == gunAttack {
		user.MedalNotifyRpc(uint32(2))
		return
	}

	user.MedalNotifyRpc(uint32(1))
}

func (user *RoomUser) MedalNotifyRpc(ty uint32) {
	user.RPC(iserver.ServerTypeClient, "MedalNotify", ty)
	user.BroadcastToWatchers(0, "MedalNotify", ty)

	var find bool
	for _, v := range user.medalhave.Data {
		if v.Id == ty {
			v.Num++
			find = true
		}
	}

	if !find {
		item := &protoMsg.MedalDataItem{
			Id:  ty,
			Num: 1,
		}
		user.medalhave.Data = append(user.medalhave.Data, item)
	}

	// user.Debug("medalhave: ", user.medalhave)
	// user.RPC(iserver.ServerTypeClient, "MedalDataList", user.medalhave)
	// user.stateMgr.BroadcastToWatchList("MedalDataList", user.medalhave)

	proto := user.GetMedalData()
	if proto == nil {
		proto = &protoMsg.MedalDataList{}
	}

	find = false
	for _, v := range proto.Data {
		if v.Id == ty {
			v.Num++
			find = true
		}
	}

	if !find {
		item := &protoMsg.MedalDataItem{
			Id:  ty,
			Num: 1,
		}
		proto.Data = append(proto.Data, item)
	}

	user.SetMedalData(proto)
	user.SetMedalDataDirty()
}

func (user *RoomUser) DisposeMultiKill() {
	now := time.Now().Unix()
	multikilllast := int64(common.GetTBSystemValue(common.System_MultiKillLast))

	if now >= user.multikillend {
		user.multikillend = now + multikilllast
		user.multikill = 1
	} else {
		user.multikillend = now + multikilllast
		user.multikill++
	}

	if user.multikill >= 2 {
		user.RPC(iserver.ServerTypeClient, "MultiKill", user.multikill)
		user.BroadcastToWatchers(0, "MultiKill", user.multikill)
	}

	user.Debug("连杀:", user.multikill)
}

func (user *RoomUser) breakdownVehicle(thisid uint64, broke bool) {
	space := user.GetSpace().(*Scene)
	if !space.teamMgr.isTeam {
		user.RPC(iserver.ServerTypeClient, "BreakdownVehicle", thisid, broke)
		return
	}

	team, ok := space.teamMgr.teams[user.GetUserTeamID()]
	if !ok {
		return
	}

	for _, memid := range team {
		tmpUser, ok := space.GetEntityByDBID("Player", memid).(*RoomUser)
		if !ok {
			continue
		}

		tmpUser.RPC(iserver.ServerTypeClient, "BreakdownVehicle", thisid, broke)
	}
}

func (user *RoomUser) GetInsignia() string {
	return user.insignia
}

//IsInAddFuelState 是否正在给载具加油
func (user *RoomUser) IsInAddFuelState() bool {
	return user.addFuelTimer != 0
}

//SucceedAddFuel 成功给载具加燃料
func (user *RoomUser) SucceedAddFuel() {
	car := user.getRidingCar()
	if car == nil {
		return
	}

	item := user.GetItemByThisid(user.addFuelThisid)
	if item == nil {
		user.Error("User doesn't own this item, thisid: ", user.addFuelThisid)
		return
	}

	prop := car.VehicleProp
	oldFuel := prop.FuelLeft

	var fuelMass float32
	if item.GetBaseID() == 3001 {
		fuelMass = float32(common.GetTBSystemValue(197))
	} else if item.GetBaseID() == 3002 {
		fuelMass = float32(common.GetTBSystemValue(198))
	} else if item.GetBaseID() == 3003 {
		fuelMass = float32(common.GetTBSystemValue(199))
	}

	prop.FuelLeft += fuelMass
	if prop.FuelLeft > prop.FuelMax {
		prop.FuelLeft = prop.FuelMax
	}

	car.RefreshInfo()

	//加油玩家扣除道具
	user.MainPack.removeItem(item.GetBaseID(), 1)

	if err := user.RPC(iserver.ServerTypeClient, "FinishAddFuel", uint8(0)); err != nil {
		user.Error("RPC FinishAddFuel err: ", err)
		return
	}

	user.addFuelTimer = 0

	if prop.FuelLeft == prop.FuelMax {
		if car.IsInAddFuelState() {
			car.BreakAddFuelAll(1)
		}
	}

	user.Info("Add fuel to car success, VehicleID: ", prop.VehicleID, " baseid: ", item.GetBaseID(), " oldFuel: ", oldFuel, " newFuel: ", prop.FuelLeft)
}

//BreakAddFuel 加油过程被打断，加油失败。失败原因 1表示油箱已满 2表示位置移动 3表示被攻击 4表示下车 5表示扔掉燃料 6表示加油的过程中攻击
func (user *RoomUser) BreakAddFuel(reason uint8) {
	if user.addFuelTimer == 0 {
		return
	}

	if err := user.RPC(iserver.ServerTypeClient, "FinishAddFuel", reason); err != nil {
		user.Error("RPC FinishAddFuel err: ", err)
		return
	}

	user.GetEntities().RemoveDelayCall(user.addFuelTimer)
	user.addFuelTimer = 0
	user.Info("Add fuel to car failed, baseid: ", user.GetVehicleProp().VehicleID)
}

// setAiPos 设置ai位置
func (user *RoomUser) setAiPos(pos linmath.Vector3) {
	scene := user.GetSpace().(*Scene)
	scene.randomNum++

	playerNum := scene.maxUserSum - scene.aiNum
	log.Debug("aiNum:", scene.aiNum, " playerNum:", playerNum, " maxUserSum:", scene.maxUserSum)

	if scene.aiNum == 0 || playerNum == 0 {
		return
	}

	if playerNum >= scene.aiNum {
		if (scene.randomNum-1)%(playerNum/scene.aiNum) == 0 {
			for entityID, ok := range scene.randomAiPos {
				if !ok {
					continue
				}

				if ai, aiOk := scene.GetEntity(entityID).(*RoomAI); aiOk {
					ai.randomAndSetAiPos(pos)
					delete(scene.randomAiPos, entityID)
					break
				}
			}
		}
	} else {
		var tmp uint32 = 0
		for entityID, ok := range scene.randomAiPos {
			if !ok {
				continue
			}

			if ai, aiOk := scene.GetEntity(entityID).(*RoomAI); aiOk {
				tmp++
				ai.randomAndSetAiPos(pos)
				delete(scene.randomAiPos, entityID)
			}

			if tmp >= scene.aiNum/playerNum && scene.randomNum != playerNum {
				break
			}
		}
	}
}

// IsFollowedParachute 判断玩家是否正在被队友跟随跳伞
func (user *RoomUser) IsFollowedParachute() bool {
	scene := user.GetSpace().(*Scene)
	st := scene.teamMgr

	_, ok := st.followRelations[user.GetID()]
	return ok
}

// IsFollowingParachute 判断玩家是否正在跟随队友跳伞
func (user *RoomUser) IsFollowingParachute() bool {
	scene := user.GetSpace().(*Scene)
	st := scene.teamMgr

	target := st.GetFollowTarget(user.GetUserTeamID(), user.GetID())
	return target != user.GetID()
}

// shellExplodeDamage 炮弹爆炸对周围玩家和载具造成伤害
func (user *RoomUser) shellExplodeDamage(msg *protoMsg.AttackReq, baseid, injuredType, shellType uint32) {
	space := user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	if msg.Origion == nil || msg.Dir == nil {
		user.Error("Invalid attack req")
		return
	}

	center := linmath.Vector3_Invalid()

	if msg.GetDefendid() != 0 {
		//击中场景中的实体(其他玩家和空载具)
		cood, ok := space.GetEntity(msg.GetDefendid()).(iserver.ICoordEntity)
		if ok && msg.GetDefendid() != user.GetID() {
			center = cood.GetPos()
		}

		//击中有人的载具(排除攻击者自己所乘载具)
		if car, ok := space.cars[msg.GetDefendid()]; ok {
			prop := user.GetVehicleProp()
			if !(prop != nil && prop.GetThisid() == car.GetThisid()) {
				carUser := car.GetUser()
				if carUser != nil {
					center = carUser.GetPos()
				}
			}
		}
	}

	if center.IsEqual(linmath.Vector3_Invalid()) {
		orgion := linmath.NewVector3(msg.Origion.X, msg.Origion.Y, msg.Origion.Z)
		dir := linmath.NewVector3(msg.Dir.X, msg.Dir.Y, msg.Dir.Z)

		ndir := linmath.NewVector3(msg.Dir.X, msg.Dir.Y, msg.Dir.Z)
		ndir.Normalize()

		//检测障碍物
		mask := unityLayerGround | unityLayerWater | unityLayerBuilding | unityLayerFurniture | unityLayerPlantAndRock
		_, pos, _, hit, _ := space.Raycast(orgion, ndir, dir.Len(), mask)
		if hit {
			center = pos
		} else {
			center = orgion.Add(dir)
		}
	}

	notify := &protoMsg.ShellExplodeNotify{
		Typ:     shellType,
		Shooter: user.GetID(),
		Baseid:  baseid,
		Center:  &protoMsg.Vector3{center.X, center.Y, center.Z},
	}

	//在aoi范围内广播坦克炮弹爆炸数据
	space.TravsalAOI(user, func(o iserver.ICoordEntity) {
		player, ok := space.GetEntity(o.GetID()).(*RoomUser)
		if !ok {
			return
		}

		player.RPC(iserver.ServerTypeClient, "ShellExplodeSync", notify)
		return
	})

	itemConfig, ok := excel.GetItem(uint64(baseid))
	if !ok {
		user.Error("Can't get item config, baseid: ", baseid)
		return
	}

	damagedCars := make(map[uint64]uint8) //伤害区域内的载具

	space.TravsalRange(&center, int(itemConfig.ThrowHurtRadius), func(o iserver.ICoordEntity) {
		target, ok := o.(iserver.ISpaceEntity)
		if !ok {
			return
		}

		targetPos := target.GetPos()
		targetPos.Y += 0.5 //防止碰撞检测是因为地面的阻挡

		//射线长度及方向
		diff := targetPos.Sub(center)
		rayDis := diff.Len()
		diff.Normalize()

		var entityID uint64
		if defender, ok := space.GetEntity(o.GetID()).(iDefender); ok {
			if prop := defender.GetVehicleProp(); prop != nil {
				entityID = prop.GetThisid()
			}
		}
		user.Debug("target.GetID():", target.GetID(), " entityID:", entityID, " msg.Defendid:", msg.Defendid)

		//检测障碍物
		if target.GetID() != msg.Defendid && entityID != msg.Defendid {
			_, _, _, hit, _ := space.Raycast(center, diff, rayDis, unityLayerBuilding|unityLayerFurniture)
			if hit {
				return
			}
		}

		dis := common.Distance(targetPos, center)
		if dis > itemConfig.ThrowHurtRadius {
			return
		}

		rate := dis / itemConfig.ThrowHurtRadius
		subhp := uint32(itemConfig.ThrowDamage * (1 - rate))

		defender, ok := space.GetEntity(o.GetID()).(iDefender)
		if ok {
			//对有人的载具造成伤害
			prop := defender.GetVehicleProp()
			carid := uint64(0)

			if prop != nil {
				carid = prop.GetThisid()
				if _, ok := damagedCars[carid]; !ok {
					user.explodeDamageVehicle(carid, subhp, center, itemConfig)
					damagedCars[carid] = 1
				}
			}

			if defender.isInTank() {
				return
			}

			if carid != 0 && defender.GetState() == RoomPlayerBaseState_WillDie {
				return
			}

			//对玩家造成伤害
			defender.DisposeSubHp(InjuredInfo{num: subhp, injuredType: injuredType, isHeadshot: false, attackid: user.GetID(), attackdbid: user.GetDBID()})
			user.Infof("Shell explode damage to player %d, damage value: %d\n", defender.GetDBID(), subhp)

			return
		}

		//对空载具造成伤害
		if _, ok := damagedCars[target.GetID()]; !ok {
			user.explodeDamageVehicle(target.GetID(), subhp, center, itemConfig)
			damagedCars[target.GetID()] = 1
		}

		return
	})
}

func (user *RoomUser) CastRpcToAllClient(method string, args ...interface{}) {
	space := user.GetSpace()
	space.TravsalAOI(user, func(o iserver.ICoordEntity) {
		player, ok := space.GetEntity(o.GetID()).(*RoomUser)
		if !ok || player.GetID() == user.GetID() {
			return
		}
		player.RPC(iserver.ServerTypeClient, method, args...)
	})
}

func (user *RoomUser) GetSkillEffectDam(effectID uint32) float32 {
	return user.SkillData.getSkillEffectDam(effectID)
}

func (user *RoomUser) GetRescueTime() uint32 {
	system, ok := excel.GetSystem(10)
	if !ok {
		return 0
	}

	per := user.SkillData.getReduceTypeDam(SE_RescueTime)
	if per == 1 {
		ratio := float32(user.getMaxItemBonus(ItemTypeFastRescue)) / 100.0
		if ratio != 0 {
			return uint32(float32(system.Value) * (1 - ratio))
		}

		return uint32(system.Value)
	}

	return uint32(float32(system.Value) * per)
}

func (user *RoomUser) GetRescueHP() uint32 {
	per := user.SkillData.getSkillEffectDam(SE_HpRecover)
	if per == 0 {
		ratio := float32(user.getMaxItemBonus(ItemTypeDeapRescue)) / 100.0
		if ratio != 0 {
			return uint32(float32(user.GetMaxHP()) * ratio)
		}

		system, ok := excel.GetSystem(15)
		if ok {
			return uint32(system.Value)
		}
	}

	return uint32(float32(user.GetMaxHP()) * per)
}

// GetCallKillEffect 每杀一人触发一个效果
func (user *RoomUser) GetCallKillEffect() {
	effectValue, ok := user.passiveEffect[SE_CallKillEffect]
	if !ok {
		return
	}

	skilleffect := GetSkillEffect(SE_CallKillEffect)
	if skilleffect == nil {
		user.Error("不能获取技能效果", SE_CallKillEffect)
		return
	}

	skilleffect.Start(user, effectValue)
}

// GetVehicleDamReduce 载具内玩家伤害降低百分比
func (user *RoomUser) GetVehicleDamReduce() float32 {
	space := user.GetSpace().(*Scene)
	if space == nil {
		return 1
	}

	if user.GetBaseState() == RoomPlayerBaseState_Ride {
		per := user.SkillData.getReduceTypeDam(SE_VehicleDamReduce)

		return per
	} else if user.GetBaseState() == RoomPlayerBaseState_Passenger {
		prop := user.GetVehicleProp()

		player, _ := space.GetEntity(prop.PilotID).(*RoomUser)
		if player == nil || player.GetBaseState() != RoomPlayerBaseState_Ride {
			return 1
		}

		per := player.SkillData.getReduceTypeDam(SE_VehicleDamReduce)
		return per
	}

	return 1
}

// rpgExplodeDamage 炮弹爆炸对周围玩家和载具造成伤害
func (user *RoomUser) rpgExplodeDamage(msg *protoMsg.ThrowDamageInfo, injuredType, shellType uint32) {
	if msg == nil || msg.Center == nil {
		user.Error("Invalid attack req")
		return
	}

	space := user.GetSpace().(*Scene)
	if space == nil {
		return
	}

	//在aoi范围内广播坦克炮弹爆炸数据
	notify := &protoMsg.ShellExplodeNotify{
		Typ:     shellType,
		Shooter: user.GetID(),
		Baseid:  msg.Baseid,
		Center:  msg.Center,
	}
	space.TravsalAOI(user, func(o iserver.ICoordEntity) {
		player, ok := space.GetEntity(o.GetID()).(*RoomUser)
		if !ok {
			return
		}

		player.RPC(iserver.ServerTypeClient, "ShellExplodeSync", notify)
		return
	})

	itemConfig, ok := excel.GetItem(uint64(msg.Baseid))
	if !ok {
		user.Error("Can't get item config, Baseid: ", msg.Baseid)
		return
	}

	damagedCars := make(map[uint64]uint8) //伤害区域内的载具
	for _, v := range msg.Defends {
		if v == nil {
			continue
		}

		target, ok := space.GetEntity(v.Id).(iserver.ISpaceEntity)
		if !ok {
			user.Error("Get defender failed, id: ", v.Id)
			continue
		}
		targetPos := target.GetPos()
		center := linmath.Vector3{msg.Center.X, msg.Center.Y, msg.Center.Z}

		dis := common.Distance(targetPos, center)
		if dis > itemConfig.ThrowHurtRadius {
			continue
		}

		rate := dis / itemConfig.ThrowHurtRadius
		subhp := uint32(itemConfig.ThrowDamage * (1 - rate))

		defender, ok := space.GetEntity(v.Id).(iDefender)
		if ok { //对有人的载具 或 人 造成伤害

			carid := uint64(0)
			if prop := defender.GetVehicleProp(); prop != nil {
				carid = prop.GetThisid()
				if _, ok := damagedCars[carid]; !ok {
					user.explodeDamageVehicle(carid, subhp, center, itemConfig) //对载具造成伤害
					damagedCars[carid] = 1
				}
			}

			if defender.isInTank() {
				continue
			}

			if carid != 0 && defender.GetState() == RoomPlayerBaseState_WillDie {
				continue
			}

			//对玩家造成伤害
			defender.DisposeSubHp(InjuredInfo{num: subhp, injuredType: injuredType, isHeadshot: false, attackid: user.GetID(), attackdbid: user.GetDBID()})
			user.Infof("Shell explode damage to player %d, damage value: %d\n", defender.GetDBID(), subhp)

		} else if _, ok := damagedCars[target.GetID()]; !ok { //对空载具造成伤害

			user.explodeDamageVehicle(target.GetID(), subhp, center, itemConfig)
			damagedCars[target.GetID()] = 1
		}
	}
}

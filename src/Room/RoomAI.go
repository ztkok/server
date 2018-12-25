package main

import (
	"common"
	"entitydef"
	"excel"
	"math/rand"
	"protoMsg"
	"strings"
	"time"
	"zeus/linmath"
	"zeus/space"

	"sort"

	"github.com/looplab/fsm"
)

// RoomAI 机器人
type RoomAI struct {
	RoomChracter
	entitydef.AIDef
	aiData      excel.AiData
	visionRange float32 //视野范围
	targetPos   linmath.Vector3
	defender    iDefender

	r      *rand.Rand   //  随机数
	aiCtrl *ControlerAI // ai控制器
}

// Init 框架层回调
func (ai *RoomAI) Init(initParam interface{}) {
	ai.r = rand.New(rand.NewSource(time.Now().UnixNano()))
	ai.nosafetime = 0
	ai.bombtime = 0
	ai.MainPack = InitMainPack(ai)
	ai.StateM = InitStateM(ai)
	ai.initProp()
	ai.SetRota(linmath.Vector3{X: 0, Y: 0, Z: 0})
	ai.aiCtrl = NewControlerAI(ai)
	rand.Seed(time.Now().UnixNano())

	attackTimer := ai.aiData.AimTmin + uint64(rand.Int63n(int64(ai.aiData.AimTmax-ai.aiData.AimTmin)))
	ai.GetEntities().RegTimerByObj(ai, ai.aiCtrl.Loop, time.Duration(attackTimer)*time.Millisecond)
}

// Destroy 析构时调用
func (ai *RoomAI) Destroy() {
	// 析构ai
	ai.GetEntities().UnregTimerByObj(ai)
}

func (ai *RoomAI) initProp() {
	space := ai.GetSpace().(*Scene)
	if space == nil {
		return
	}

	aiData := ai.getAIData()
	ai.aiData = aiData
	//获取一个这局比赛中没有使用过的名字
	ai.SetName(space.getAiRandName())
	ai.initAiRole(aiData.Roleid)
	maxHP := ai.GetRoleBaseMaxHP()
	ai.SetMaxHP(maxHP)
	ai.SetHP(maxHP)
	ai.SetChracterMapDataInfo(ai.fillMapData())
	ai.SetChracterMapDataInfoDirty()

	if ai.r.Intn(100) < 50 {
		ai.SetState(RoomPlayerBaseState_Crouch)
	} else {
		ai.SetState(RoomPlayerBaseState_Stand)
	}

	ai.visionRange = float32(aiData.Aiview)
	// pos, err := space.getAiPos()
	// if err != nil {
	// 	//刷新至地图上随机位置
	// 	pos = ai.getRandPoint(linmath.Vector3{
	// 		X: ai.GetSpace().(*Scene).mapdata.Width,
	// 		Y: 0,
	// 		Z: ai.GetSpace().(*Scene).mapdata.Height,
	// 	}, ai.GetSpace().(*Scene).mapdata.Width*3/4)
	// }

	pos := linmath.Vector3{0, 0, 0}
	ai.SetPos(pos)
	ai.initAiItem(aiData.Inititem, space)
	ai.initHead(aiData.Helmetappearance)
	ai.initPack(aiData.Bagappearance)
}

func (ai *RoomAI) getWeight(weight map[interface{}]int) interface{} {
	var sum int
	var sortArray []int
	for _, v := range weight {
		sum += v
		sortArray = append(sortArray, v)
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	index := r.Intn(sum)
	sort.Ints(sortArray)
	var curSum, key int
	for _, v := range sortArray {
		curSum += v
		if index <= curSum {
			key = v
			break
		}
	}
	for k, v := range weight {
		if v == key {
			return k
		}
	}
	return nil
}

// initHead 幻化头盔
func (ai *RoomAI) initHead(param string) {
	prop := ai.GetHeadProp()
	if prop == nil {
		return
	}
	var level uint64
	switch prop.Baseid {
	case 1202:
		level = 1
	case 1204:
		level = 2
	case 1206:
		level = 3
	default:
		return
	}
	index := ai.getWeight(common.StringToMapInt(param, "|", ";"))
	allInfo := strings.Split(index.(string), ":")
	for _, v := range allInfo {
		id := common.StringToUint64(v)
		iInfo, ok := excel.GetItem(id)
		if !ok {
			continue
		}
		if iInfo.Subtype == level {
			prop.Baseid = uint32(id)
			break
		}
	}
	ai.SetHeadProp(prop)
	ai.SetHeadPropDirty()
}

// initPack 幻化背包
func (ai *RoomAI) initPack(param string) {
	prop := ai.GetBackPackProp()
	if prop == nil {
		return
	}
	var level uint64
	switch prop.Baseid {
	case 1603:
		level = 1
	case 1601:
		level = 2
	case 1602:
		level = 3
	default:
		return
	}
	index := ai.getWeight(common.StringToMapInt(param, "|", ";"))
	allInfo := strings.Split(index.(string), ":")
	for _, v := range allInfo {
		id := common.StringToUint64(v)
		iInfo, ok := excel.GetItem(id)
		if !ok {
			continue
		}
		if iInfo.Subtype == level {
			prop.Baseid = uint32(id)
			break
		}
	}
	ai.SetBackPackProp(prop)
	ai.SetBackPackPropDirty()
}
func (ai *RoomAI) initAiRole(param string) {
	role := ai.getWeight(common.StringToMapInt(param, "|", ";"))
	if role == nil {
		return
	}
	ids, ok := role.(string)
	if !ok {
		return
	}
	id := common.StringToUint32(ids)
	ai.SetRole(id)
	ai.SetRoleModel(id)
	ai.SetRoleBase(id)
}

func (ai *RoomAI) initAiItem(param string, space *Scene) {
	sf := GetMapItemRate(space)
	rand.Seed(time.Now().UnixNano())
	rules := strings.Split(param, ";")
	for _, rulestr := range rules {
		rule := common.StringToUint64(rulestr)
		config, ok := sf.mapitemrate[rule]
		if !ok {
			//log.Warn("错误ai配置规则", str)
			continue
		}

		for _, v := range config {
			rate := rand.Intn(10000) + 1
			if rate > v.createrate {
				continue
			}

			for index := 0; index < v.num; index++ {
				randnum := rand.Intn(v.total) + 1
				var percent int
				for _, itemrate := range v.itemrate {
					percent += itemrate.rate
					if percent >= randnum {
						for _, itembaseid := range itemrate.items {
							ai.AddItem(NewItemBullet(uint32(itembaseid), 1))
						}
						break
					}
				}
			}
		}
	}
}

// OnEnterSpace 框架层回调
func (ai *RoomAI) OnEnterSpace() {
	space := ai.GetSpace().(*Scene)
	space.members[ai.GetID()] = true
	space.haveai++

	if space.haveai >= space.aiNum {
		space.BroadAliveNum()
	}
}

// 离开某个状态时, 取消当前所有状态检查timer
func (ai *RoomAI) onLeaveState(e *fsm.Event) {
	if err := ai.GetEntities().UnregTimerByObj(ai); err != nil {
		ai.Error("onLeaveState failed, UnregTimerByObj err: ", err)
	}
}

// UnderAttack 被攻击
func (ai *RoomAI) UnderAttack(entityID uint64) {
	/*
		if ai.fsm.Is("moving") {
			if err := ai.fsm.Event("UnderAttack", entityID); err != nil {
				// ai.Error(err, ai)
			}
		}
	*/
	//ai.aiCtrl.UnderAttack(entityID)
}

// SetState 设置ai状态
func (ai *RoomAI) SetState(v uint8) {
	// ai.Info(ai.GetID(), "设置ai状态", v)
	// ai.AIDef.SetState(v)
	ai.SetBaseState(v)

	spendStr := ""
	if v == RoomPlayerBaseState_Stand {
		spendStr = ai.rolebase.Mvspeed
	} else if v == RoomPlayerBaseState_Down {
		spendStr = ai.rolebase.Crawlspeed
	} else if v == RoomPlayerBaseState_Crouch {
		spendStr = ai.rolebase.CrouchSpeed
	} else {
		return
	}

	s := strings.Split(spendStr, ";")
	if len(s) == 0 {
		return
	}

	// f, err := strconv.ParseFloat(s[0], 32)
	// if err == nil {
	// 	//ai.SetSpeed(float32(f))
	// } else {
	// 	ai.Warn("设置ai速度 fail fail fail", spendStr, v, Stand, err)
	// }

}

// Death 死亡
func (ai *RoomAI) Death(injuredType uint32, attackid uint64) {
	/*
			if err := ai.fsm.Event("die"); err != nil {
				ai.Error(err, ai)
			}

		if err := ai.aiCtrl.fsm.Event("AIDie"); err != nil {
			ai.Error(err, ai)
		}
	*/
	space := ai.GetSpace().(*Scene)
	if space == nil {
		return
	}

	//ai.Info("ai start Death space.members:", len(space.members), " spaceid:", space.GetID())
	ai.SetHP(0)
	ai.SetState(RoomPlayerBaseState_Dead)

	dropid, droppos := ai.MainPack.OnDeath()
	if dropid != 0 {
		attacker, _ := space.GetEntity(attackid).(*RoomUser)
		if attacker != nil {
			attacker.addKillDrop(dropid, droppos)
		}
	}

	if ai.IsBeingWatched() {
		ai.TerminateWatch()
	}

	delete(space.members, ai.GetID())

	if space.aiNumSurplus > 0 {
		space.aiNumSurplus--
	}

	space.BroadAliveNum()

	space.ISceneData.onDeath(ai)
	space.AddDelayCall(ai.LeaveSpace, 3*time.Second)

	if err := ai.GetEntities().UnregTimerByObj(ai); err != nil {
		ai.Error("Ai die, UnregTimerByObj err: ", err)
	}

	ai.Info("AI die, id: ", ai.GetID(), " injuredType: ", injuredType, " attackid: ", attackid, " left: ", len(space.members))
}

// GetUserTeamID AI玩家TeamID为0
func (ai *RoomAI) GetUserTeamID() uint64 {
	return 0
}

func (ai *RoomAI) isAI() bool {
	return true
}

func (ai *RoomAI) isHeadShot(defender iDefender) bool {
	return true
}

// SendChat 发消息, AI不作处理
func (ai *RoomAI) SendChat(str string) {
}

// AddEffect AI不作处理
func (ai *RoomAI) AddEffect(baseid uint32, id uint32) {
}

func (ai *RoomAI) downVehicle(pos, dir linmath.Vector3, prop *protoMsg.VehiclePhysics, reducedam bool) {
}

func (ai *RoomAI) vehicleEngineBroke(autodown bool) {
}

func (ai *RoomAI) compilotDown(pos, dir linmath.Vector3, speed float32) {
}

func (ai *RoomAI) getRidingCar() *Car {
	return nil
}

func (ai *RoomAI) IsinBoat() bool {
	return false
}

// IncrHeadShotNum 当场爆头数自增
func (ai *RoomAI) IncrHeadShotNum() {
}

func (ai *RoomAI) AddEffectHarm(effect uint32) {
}

// AddHp AI增加血量
func (ai *RoomAI) AddHp(num uint32) {
	curHP := ai.GetHP()
	maxHP := ai.GetMaxHP()
	if curHP >= maxHP {
		return
	}

	if curHP+num > maxHP {
		ai.SetHP(maxHP)
	} else {
		ai.SetHP(curHP + num)
	}
}

// DisposeSubHp 减少血量
func (ai *RoomAI) DisposeSubHp(injuredInfo InjuredInfo) {

	result := ai.TmpSubHp(injuredInfo.num, injuredInfo.attackid, &injuredInfo)
	space := ai.GetSpace().(*Scene)
	if space == nil {
		return
	}

	if result == SubHpDie {
		space.BroadAliveNum()

		space.BroadDieNotify(injuredInfo.attackid, ai.GetID(), false, injuredInfo)
	}

	attacker, _ := space.GetEntity(injuredInfo.attackid).(IRoomChracter)
	if attacker != nil {
		if !attacker.isAI() {
			attackUser, _ := space.GetEntity(injuredInfo.attackid).(*RoomUser)
			if attackUser != nil {
				attackUser.DisposeAttackResult(&injuredInfo, result, ai.GetID())
			}
		}
	}
}

// TmpSubHp 处理减血逻辑
func (ai *RoomAI) TmpSubHp(num uint32, attackid uint64, injuredInfo *InjuredInfo) uint32 {

	//space := ai.GetSpace().(*Scene)
	//ai.Info("ai start TmpSubHp space.members:", len(space.members), " spaceid:", space.GetID())
	curHP := ai.GetHP()
	if curHP == 0 {
		return SubHpError
	}

	if curHP > num {
		ai.SetKiller(0)
		ai.SetHP(curHP - num)
		ai.CastRPCToAllClient("OnDamage", uint64(injuredInfo.injuredType))
		injuredInfo.effectHarm = num
		return SubHpNormal
	} else {
		ai.SetKiller(attackid)
		injuredInfo.effectHarm = curHP
		ai.SetHP(0)
		ai.CastRPCToAllClient("OnDamage", uint64(injuredInfo.injuredType))

		ai.Death(injuredInfo.injuredType, attackid)
		return SubHpDie
	}
	//ai.Info("ai end TmpSubHp space.members:", len(space.members), " spaceid:", space.GetID())
	return SubHpError

}

// GetAttack 获得攻击力
func (ai *RoomAI) GetAttack(targetPos linmath.Vector3, headshot bool) uint32 {
	if ai.isMeleeWeaponUse() {
		useweapon := ai.GetMeleeWeaponUse()
		base, ok := excel.GetGun(uint64(useweapon))
		if !ok {
			return 1
		} else {
			return uint32(base.Hurt * base.AiDamageRate)
		}
	}
	if weapon := ai.equips[ai.useweapon]; weapon != nil {
		if base, ok := excel.GetGun(weapon.base.Id); ok {
			hurt := ai.computeAttack(weapon.GetBaseID(), targetPos, headshot)
			return uint32(float32(hurt) * base.AiDamageRate)
		}

	}

	return 1
}

// 在point点的radius半径范围内随机获取可以抵达并且不是水面的点
func (ai *RoomAI) getRandPoint(point linmath.Vector3, radius float32) linmath.Vector3 {
	tryTime := 0
	var pos linmath.Vector3
	waterLevel := ai.GetSpace().(*Scene).mapdata.Water_height
	for {
		if tryTime > 15 {
			pos.X = 1024
			pos.Y = 18.562378
			pos.Z = 1024

			break
		}

		tryTime++
		pos = ai.RandXZ(point, radius)
		pos.Y, _ = ai.GetSpace().GetHeight(pos.X, pos.Z)
		if pos.Y > waterLevel {
			break
		}
	}
	return pos
}

// getSpawnWeight 获取ai类型
func (ai *RoomAI) getAIData() excel.AiData {
	weightInfo := map[interface{}]int{}
	for _, data := range excel.GetAiMap() {
		weightInfo[data] = int(data.SpawnWeight)
	}
	chose := ai.getWeight(weightInfo)
	return chose.(excel.AiData)
}

func (ai *RoomAI) RandXZ(v linmath.Vector3, r float32) linmath.Vector3 {

	pos := linmath.Vector3{}
	pos.Y = 0

	pos.X = float32(ai.r.Float64())*r - r
	pos.Z = float32(ai.r.Float64())*r - r

	return v.Add(pos)
}

// CreateNewEntityState 创建一个新的状态快照，由框架层调用
func (ai *RoomAI) CreateNewEntityState() space.IEntityState {
	return NewRoomNpcState()
}

// SetBaseState 设置 BaseState的值
func (ai *RoomAI) SetBaseState(bs byte) {
	state := ai.GetNpcState()
	state.BaseState = bs
	state.SetModify(true)
}

// GetBaseState 获取BaseState
func (ai *RoomAI) GetBaseState() byte {
	state := ai.GetNpcState()
	return state.BaseState
}

// GetNpcState 获取当前的状态快照
func (ai *RoomAI) GetNpcState() *RoomNpcState {
	return ai.GetStates().GetLastState().(*RoomNpcState)
}

func (ai *RoomAI) AdviceNotify(notifyType uint32, id uint64) {

}

func (ai *RoomAI) SetDieNotify(proto *protoMsg.DieNotifyRet) {
}

func (ai *RoomAI) GetInsignia() string {
	return ""
}

func (ai *RoomAI) isInTank() bool {
	return false
}

func (ai *RoomAI) CastRpcToAllClient(method string, args ...interface{}) {

}

// randomAndSetAiPos 随机并且设置ai的位置
func (ai *RoomAI) randomAndSetAiPos(userPos linmath.Vector3) {
	space := ai.GetSpace().(*Scene)
	pos, err := space.getAiPos(userPos)

	if err != nil {
		//刷新至地图上随机位置
		pos = ai.getRandPoint(linmath.Vector3{
			X: ai.GetSpace().(*Scene).mapdata.Width,
			Y: 0,
			Z: ai.GetSpace().(*Scene).mapdata.Height,
		}, ai.GetSpace().(*Scene).mapdata.Width*3/4)
	}

	ai.SetPos(pos)
}

// IsDead 是否已经死亡
func (ai *RoomAI) IsDead() bool {
	return ai.GetHP() == 0
}

// IsOffline 是否离线
func (ai *RoomAI) IsOffline() bool {
	return false
}

// GetWatchTargetInTeam 在队伍内获取观战目标
func (ai *RoomAI) GetWatchTargetInTeam() uint64 {
	return 0
}

// LeaveScene 离开场景
func (ai *RoomAI) LeaveScene() {
	ai.LeaveSpace()
}

// GetUserType 获取玩家的类型
func (ai *RoomAI) GetUserType() uint32 {
	return RoomUserTypePlayer
}

// GetTeamMembers 获取队伍成员
func (ai *RoomAI) GetTeamMembers() []uint64 {
	return []uint64{ai.GetID()}
}

func (ai *RoomAI) GetSkillEffectDam(effectID uint32) float32 {
	return 0
}

package main

// ProgressBarType 进度条类型
const (
	// USERDOWNTYPE 玩家被击倒
	USERDOWNTYPE = 1
	// USERRESCUETYPE 玩家救援
	USERRESCUETYPE = 2
	// USERBERESCUETYPE 玩家被救援
	USERBERESCUETYPE = 3
)

// injuredType	玩家受伤害类型
const (
	// gunAttack 枪支攻击
	gunAttack = 0
	// fistAttack 拳头攻击
	fistAttack = 1
	// mephitis 毒气伤害
	mephitis = 2
	// bomb 炸弹伤害
	bomb = 3
	// carhit 车撞击
	carhit = 4
	// losthp 失血死亡
	losthp = 5
	// 枪支爆头攻击
	headShotAttack = 6
	// 坠落伤害
	falldam = 7
	// 投掷伤害
	throwdam = 8
	// 下车伤害
	vehicledam = 9
	//carcollision 车撞墙伤害
	carcollision = 10
	//载具爆炸伤害
	carExplode = 11
	//坦克炮弹爆炸
	tankShell = 12
)

const (
	// ActionUseMedicine 使用药品动作
	ActionUseMedicine = 2 // 使用药品动作
)

const (
	RoomUserTypePlayer  = 1 //场景内的玩家
	RoomUserTypeWatcher = 2 //场景外的观战者
)

const (
	RoomUserTypeNormal = 0 // 普通玩家
	RoomUserTypeElite  = 1 // 精英玩家
	RoomUserTypeRed    = 2 // 红军玩家
	RoomUserTypeBlue   = 3 // 蓝军玩家
)

// 技能效果ID
const (
	SE_Recoil           = 1  //机械后坐力效果
	SE_ShrinkDam        = 2  //毒圈伤害降低百分比
	SE_TurnHp           = 3  //按固定值扣血
	SE_WeaponDam        = 4  //武器伤害
	SE_FireSpeed        = 5  //射击速度
	SE_Shield           = 6  //生成护盾吸收固定值伤害
	SE_MapTag           = 7  //地图标记一次一定范围内敌人和空投位置
	SE_RandomItem       = 8  //随机生成一个道具
	SE_AddHpCap         = 9  //血量上限增加百分比
	SE_MoveUseItem      = 10 //移动使用可回血道具
	SE_RescueTime       = 11 //救援时间减少百分比
	SE_HpRecover        = 12 //救援血量恢复至总值百分比
	SE_DamReduce        = 13 //自身伤害削减百分比
	SE_VehicleNoDam     = 14 //载具碰撞伤害减少百分比
	SE_OilLoss          = 15 //开车油耗减少百分比
	SE_ReleaseFog       = 16 //原地瞬时释放烟雾
	SE_CallKillEffect   = 17 //每杀一人触发一个效果
	SE_VehicleDamReduce = 18 //载具内玩家伤害降低百分比
	SE_RandomItem2      = 19 //随机生成一个道具2
	SE_RandomItem3      = 20 //随机生成一个道具3
)

// 技能目标
const (
	SkillTarget_Own          = 0 // 自己
	SkillTarget_Friend       = 1 // 友方
	SkillTarget_OwnAndFriend = 2 // 自己及友方
)

// 攻击部位
const (
	AttackPos_Body = 0 // 身体
	AttackPos_Head = 1 // 头部
	AttackPos_Dong = 2 // 护盾
)

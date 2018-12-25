package common

const (
	//SpaceStatusInit 初始状态
	SpaceStatusInit int = iota + 1
	//SpaceStatusSelect 玩家选角
	SpaceStatusSelect
	//SpaceStatusCreating room游戏初始化
	SpaceStatusCreating
	SpaceStatusInitRoom
	//SpaceStatusWaitIn 玩家进入
	SpaceStatusWaitIn
	//SpaceStatusBegin 游戏开始
	SpaceStatusBegin
	//SpaceStatusBalance 游戏结束
	SpaceStatusBalance
	//SpaceStatusVerify 验证结果中
	SpaceStatusVerify
	// SpaceStatusBalanceDone 结算完成
	SpaceStatusBalanceDone
	//SpaceStatusClose 游戏关闭
	SpaceStatusClose
)

const (
	System_TeamMemberNum           = 1
	System_InitCell                = 3
	System_RefreshMax              = 4
	System_InitRadius              = 5
	System_RoomUserLimit           = 6
	System_MatchWait               = 7
	System_RoomUserMin             = 8
	System_RefreshBomb             = 11
	System_RefreshSafeArea         = 12
	System_SafeAreaNotify          = 13
	System_SafeAreaRadiusRate      = 14
	System_CrouchToDown            = 23
	System_CrouchToStand           = 24
	System_EnergyMax               = 25
	System_FallDamageA             = 26
	System_FallDamageB             = 27
	System_FallDamage              = 28
	System_MailOverdue             = 29
	System_MailMax                 = 30
	System_VehicleDamageA          = 35
	System_VehicleDamageB          = 36
	System_VehicleHitA             = 42
	System_VehicleHitB             = 43
	System_AddHpLimit              = 44
	System_VehicleCollisionA       = 45
	System_VehicleCollisionB       = 46
	System_VehicleCollisionC       = 47
	System_FallHeight              = 49
	System_MatchMinSum             = 53
	System_MatchMinWait            = 55
	System_InitBagPack             = 57
	System_ForceEjectNotify        = 58
	System_MinLoadingTime          = 59
	System_MaxLoadingTime          = 60
	System_SummonAiSum             = 63
	System_InitParachuteID         = 64
	System_RefreshBoxHeight        = 65
	System_GetBoxCold              = 66
	System_RefreshBoxSpeed         = 67
	System_RefreshBoxID            = 68
	System_RefreshFakeBoxID        = 69
	System_KillGetCoin             = 70
	System_DayCoinLimit            = 71
	System_MultiKillLast           = 73
	System_KillDropLast            = 74
	System_SuperDropBoxCold        = 80
	System_DayDiamLimit            = 81
	System_ExpRateMax              = 84 // 经验加成倍率上限
	System_RecruitInitLevel        = 85 // 老带新活动页面开启所需军衔等级
	System_TakeTeacherValidDays    = 86 // 战友绑定活动页面关闭天数
	System_LeaveSceneTime          = 89 // 结算后离开场景的倒计时间
	System_HonourKillDistance      = 91 // 勋章中击杀距离要求
	System_ChatPushCDTime          = 92 // 发送聊天cd时间
	System_TeamCustomCDTime        = 93 // 定制队伍cd时间
	System_WatchTeamLimit          = 101
	System_WatchSpaceLimit         = 102
	System_WatchWait               = 103
	System_EliteFreshTime          = 151 // 精英玩家刷新时间
	System_EliteNum                = 152 // 精英玩家初始数量
	System_EliteWeapon             = 153 // 精英玩家初始武器（作废）
	System_EliteHead               = 154 // 精英玩家初始头盔
	System_ElitePack               = 155 // 精英玩家初始背包
	System_EliteBody               = 156 // 精英玩家初始护甲
	System_EliteAttackE            = 157 // 精英玩家初始友军伤害
	System_EliteAttackN            = 158 // 普通玩家初始友军伤害
	System_EliteAttackOpen         = 159 // 精英伤害开启时间
	System_RefreshSpecialBox       = 160
	System_EliteWeaponBullet       = 161 //初始精英武器的子弹组数量
	System_ElitePosFreshTime       = 162 //初始精英武器的子弹组数量
	System_RedAndBlueFreshTime     = 180 // 红和蓝军玩家区分刷新时间
	System_RedAndBlueAttackS       = 181 // 红和蓝军玩家初始友军伤害
	System_RedAndBlueAttackN       = 182 // 普通玩家初始友军伤害
	System_RedAndBluePosFreshTime  = 183 // 红和蓝军位置通知刷新频率
	System_RefreshSuperBox         = 184 // 超级空投道具id
	System_RedAndBlueAttackOpen    = 185 // 红蓝对决阵营友军伤害开启时间（秒）
	System_PrivilegeGoodStartTime  = 204 // 特惠开始时间（秒）
	System_PrivilegeGoodInterTime7 = 205 // 特惠间隔时间7天（秒）
	System_PrivilegeGoodInterTime1 = 206 // 特惠间隔时间1天（秒）
	System_PrivilegeGoodMaxRound7  = 207 // 周特惠最大轮次
	System_PrivilegeGoodMaxRound1  = 208 // 日特惠最大轮次
	System_ArcadeMatchMaxSum       = 220 // 娱乐模式 房间无条件开启人数/随机AI模式下 总人数上限
	System_ArcadeMatchMaxTime      = 221 // 娱乐模式 匹配等待最长时间（无论多少人都开）
	System_ArcadeMatchMinSum       = 222 // 娱乐模式 真实玩家开房间人数限制 人数下限
	System_ArcadeMatchMinTime      = 223 // 娱乐模式 最小等待时间
	System_ArcadeRefreshSafeArea   = 224 // 娱乐模式 初始安全区刷新时间
	System_ArcadeAirLineRadius     = 225 // 娱乐模式 飞机航线起始点距圆心半径
	System_TankWarMatchMaxSum      = 226 // 坦克大战 房间无条件开启人数/随机AI模式下 总人数上限
	System_TankWarMatchMaxTime     = 227 // 坦克大战 匹配等待最长时间（无论多少人都开）
	System_TankWarMatchMinSum      = 228 // 坦克大战 真实玩家开房间人数限制 人数下限
	System_TankWarMatchMinTime     = 229 // 坦克大战 最小等待时间
	System_TankWarRefreshSafeArea  = 230 // 坦克大战 初始安全区刷新时间
	System_TankWarAirLineRadius    = 231 // 坦克大战 飞机航线起始点距圆心半径
	System_FollowParachuteOffsetX  = 240 // 跟随跳伞X方向偏移量
	System_FollowParachuteOffsetY  = 241 // 跟随跳伞Y方向偏移量
	System_FollowParachuteOffsetZ  = 242 // 跟随跳伞Z方向偏移量
	System_WCupOddsRefreshTime     = 243 // 世界杯赔率刷新时间
	System_WCupDayContestNum       = 244 // 世界杯冠军竞猜单日投注次数
	System_WCupPayRate             = 245 // 世界杯冠军竞猜赔付率
	System_WCupSendReward          = 246 // 世界杯冠军竞猜结算时间（小时）
	System_WCupDayMatchNum         = 248 // 世界杯胜负竞猜单日投注次数
	System_BallStarDayNum          = 249 // 世界杯一球成名单日投注次数
	System_BallStarWheelNum        = 250 // 世界杯一球成名转盘数量
	System_BallStarGridNum         = 251 // 世界杯一球成名格子数量
	System_BallStarUseItem         = 252 // 世界杯一球成名抽奖消耗道具id
	System_WCupOddsTime            = 253
	System_BallStarFreshWeek       = 255  // 世界杯一球成名刷新时间(周几)
	System_BallStarFreshHour       = 256  // 世界杯一球成名刷新时间(小时)
	System_SpecialMaxReplacements  = 258  // 特训任务普通补领上限次数
	System_SpecialAddReplacements  = 259  // 特训任务精英补领额外增加上限次数
	System_DayBraveLimit           = 260  // 每日勇气值上限
	System_CallBoxItem             = 3001 // 信号枪召唤空投itemid
	System_CallBoxRealItem         = 3002 // 信号枪召唤的真实id
	System_CallBoxItem1            = 3003 // 信号枪召唤空投itemid
	System_CallBoxRealItem1        = 3004 // 信号枪召唤的真实id
)

const (
	System2_SpecialTaskClearTime  = 1 //特训任务每日清零时间
	System2_SpecialTaskEnableItem = 4 //特训任务精英资格激活道具
	System2_NameColorItem         = 5 //名字变色道具
)

const (
	Item_BraveTicket      = 2010 //勇者模式入场券
	Item_Coin             = 2201 //金币
	Item_Brick            = 2202 //金砖
	Item_Diam             = 2203 //钻石
	Item_Exp              = 2204 //经验值
	Item_Activeness       = 2205 //活跃度
	Item_ComradePoints    = 2206 //战友积分
	Item_BraveCoin        = 2209 //勇气值
	Item_SpecailExpMedal  = 2210 //特训经验勋章
	Item_SeasonHeroMedal  = 2211 //赛季英雄勋章
	Item_SeasonExpMedal   = 2212 //赛季经验勋章
	Item_ChatTrumpet      = 4100 //聊天喇叭
	Item_DropBoxGun       = 1034 //空投召唤枪
	Item_BombAreaGun      = 1035 //空袭召唤枪
	Item_Tank             = 1910 //坦克
	Item_TankShell        = 2103 //坦克炮弹
	Item_DropTank1        = 1049 //坦克信号枪1
	Item_DropTank         = 1050 //坦克信号枪0
	Item_ClipExtend       = 1405 //弹夹扩容器
	Item_AWMClipExtend    = 1421 //狙击枪弹夹扩容器
	Item_EnergyDrink      = 1117 //能量饮料
	Item_TankBox0         = 1509 //坦克空投0
	Item_TankBox1         = 1510 //坦克空投1
	Item_PackExtendLow    = 1607 //初级战术扩容背包
	Item_PackExtendMiddle = 1608 //中级战术扩容背包
	Item_PackExtendHigh   = 1609 //高级战术扩容背包
)

const (
	Mail_ChampionReward      = 101 //冠军竞猜活动奖励
	Mail_ChampionOut         = 102 //冠军竞猜球队已淘汰
	Mail_TakeTeacher         = 103 //战友绑定
	Mail_BuyMonthCard        = 104 //成功开通月卡
	Mail_MonthCardSoonExpire = 105 //月卡即将过期通知
	Mail_MonthCardExpired    = 106 //月卡过期通知
)

const (
	Match_InitRange  = 6
	Match_InitWait   = 7
	Match_WaitRange  = 8
	Match_InitRank   = 9
	Match_InitPow    = 10
	Match_Timeout    = 11
	Match_Simulator  = 12
	Match_DuoTwo     = 13
	Match_SquadTwo   = 14
	Match_SquadThree = 15
	Match_SquadFour  = 16
	Match_MaxExpand  = 17
)

const (
	Status_Init        = 0
	Status_ShrinkBegin = 1
)

const (
	Bomb_Status_Init        = 0
	Bomb_Status_ShrinkBegin = 1
	Bomb_Status_Disappear   = 2
)

const (
	// notifyCommon 普通通知
	NotifyCommon = 1
	// notifyeError 错误通知
	NotifyError = 2
)

const (

	// StateFree 空闲
	StateFree = 0
	// StateMatchWaiting 匹配等待中
	StateMatchWaiting = 1
	// StateMatching 匹配中
	StateMatching = 2
	// StateGame 游戏中
	StateGame = 3
	// StateOffline 不在线
	StateOffline = 4
	// StateWatch 观战
	StateWatch = 5
)

const (
	//ErrCodeIDInviteJoin 邀请加入房间失效
	ErrCodeIDInviteJoin = 9
	//ErrCodeIDPackCell 背包空间不足
	ErrCodeIDPackCell = 11
)

const (
	// MatchMgrSolo 单人匹配队列
	MatchMgrSolo = "MatchMgrSolo"

	// MatchMgrDuo 双人匹配队列
	MatchMgrDuo = "MatchMgrDuo"

	// MatchMgrSquad 四人匹配队列
	MatchMgrSquad = "MatchMgrSquad"

	// TeamMgr 组队管理器
	TeamMgr = "TeamMgr"
)

// 玩家所在队伍状态
const (
	// PlayerTeamMatching 玩家所在队伍匹配中
	PlayerTeamMatching = 1
)

// QQAppID ID
const QQAppID = 1106393072

const QQAppIDStr = "1106393072"

// WXAppID ID
const WXAppID = "wxa916d09c4b4ef98f"

// GAppID 游客登录时
const GAppID = "G_1106393072"

// MSDKKey KEY
const MSDKKey = "7eead96a3fdb063615b181d7c01480e4"

const (
	PlayerCareerTotalData = "PlayerCareerData"      //玩家生涯总数据redis key标识
	PlayerCareerSoloData  = "PlayerCareerSoloData"  //玩家生涯单排数据redis key标识
	PlayerCareerDuoData   = "PlayerCareerDuoData"   //玩家生涯双排数据redis key标识
	PlayerCareerSquadData = "PlayerCareerSquadData" //玩家生涯四排数据redis key标识
)

const (
	TotalRank = "PlayerTotalRank"
	SoloRank  = "PlayerSoloRank"
	DuoRank   = "PlayerDuoRank"
	SquadRank = "PlayerSquadRank"
)

const (
	BraveBattleRank = "BraveBattleRank"
)

const (
	SeasonRankTypStrWins   = "Wins"  //赛季获胜数排行redis key标识
	SeasonRankTypStrKills  = "Kills" //赛季击杀数排行redis key标识
	SeasonRankTypStrRating = ""      //赛季积分排行redis key标识
)

const (
	MatchModeNormal  uint32 = 0  //普通模式
	MatchModeScuffle uint32 = 1  //快速模式
	MatchModeEndless uint32 = 2  //乱斗模式
	MatchModeElite   uint32 = 3  //精英模式
	MatchModeVersus  uint32 = 4  //红蓝对决
	MatchModeArcade  uint32 = 5  //娱乐模式
	MatchModeTankWar uint32 = 6  //坦克大战
	MatchModeBrave   uint32 = 11 //春节精英赛
)

// 成就条件
const (
	AchievementBattleNum       = 1  //参加普通模式战斗次数
	AchievementDamageTotal     = 2  //普通模式下造成的总伤害量
	AchievementKillTotal       = 3  //普通模式下击败人数
	AchievementTopTen          = 4  //普通模式下前十名次数
	AchievementSoloTopOne      = 5  //普通模式下单人获得第一名
	AchievementDuoTopOne       = 6  //普通模式下双排获得第一名
	AchievementSquadTopOne     = 7  //普通模式下四排获得第一名
	AchievementFistKill        = 8  //普通模式下用拳头击杀
	AchievementGrenadeKill     = 9  //普通模式下用手雷炸死
	AchievementCarKill         = 10 //普通模式下用载具撞死
	AchievementSniperKill      = 11 //普通模式下用狙击枪击杀
	AchievementFirstDie        = 12 //普通模式下第一个死（落地成盒）
	AchievementFirstKill       = 13 //普通模式下第一滴血
	AchievementRunDistance     = 14 //普通模式下跑动距离
	AchievementParachute       = 15 //伞包收集
	AchievementRole            = 16 //人物收集
	AchievementCoin            = 17 //金币收集
	AchievementOpenTreasureBox = 18 //开宝箱
)

// 勋章类型
const (
	HonorKill         = 1 // 击杀数
	HonorHeadShot     = 2 // 爆头数
	HonorRecover      = 3 // 治疗
	HonorHarm         = 4 // 伤害
	HonorMultiKill    = 5 // 连击数
	HonorWin          = 6 // 胜利数
	HonorKillDistance = 7 // 击杀距离
	HonorRank         = 8 // 排名
	HonorRescue       = 9 // 救援次数
)

// reason 变化原因
const (
	RS_Login            = 0  //登陆奖励
	RS_Battle           = 1  //每局奖励
	RS_Activity         = 2  //签到活动奖励
	RS_BuyGoods         = 3  //购买商品
	RS_UpdateAct        = 4  //更新活动
	RS_Mail             = 5  //邮件获取
	RS_SendGift         = 6  //好友送礼
	RS_ChangeName       = 7  //修改角色名
	RS_FirstWinAct      = 8  //当天首次吃鸡活动
	RS_ThreeDayAct      = 9  //3日签到活动
	RS_NewYearAct       = 10 //新年活动
	RS_Achievement      = 11 //成就奖励
	RS_SeasonAward      = 12 //赛季奖励
	RS_TreasureBox      = 13 //开宝箱
	RS_MilitaryRank     = 14 //军衔奖励
	RS_DayTask          = 15 //每日任务奖励
	RS_TeacherPupil     = 16 //结为使徒奖励
	RS_ComradeTask      = 17 //师徒任务
	RS_VeteranRecall    = 18 //老兵召回奖励
	RS_BackBattle       = 19 //重返光荣战场奖励
	RS_Festival         = 20 //节日活动奖励
	RS_Exchange         = 21 //兑换活动奖励
	RS_BallStar         = 22 //一球成名活动奖励
	RS_NormalUse        = 23 //普通使用物品
	RS_WorldCupChampion = 24
	RS_WorldCupMatch    = 25
	RS_SpecialTask      = 26 //特训任务
	RS_Pay              = 27 //充值
	RS_MonthCard        = 28 //月卡
	RS_Chat             = 29 //聊天
	RS_ChallengeTask    = 30 //赛季挑战任务
	RS_PayLevel         = 31 //首次充送
	RS_RollBack         = 32 //购买失败回滚
)

// 货币类型
const (
	MT_MONEY         = 0 //金币
	MT_DIAMOND       = 1 //钻石
	MT_BraveCoin     = 2 //勇气值
	MT_NO            = 3 //无，奖励发放
	MT_ComradePoints = 4 //战友积分
	MT_EXP           = 5 //经验值
	MT_Brick         = 6 //金砖
)

// 任务名
const (
	TaskName_Day       = "Day"       //每日任务
	TaskName_Comrade   = "Comrade"   //战友任务
	TaskName_Special   = "Special"   //特训任务
	TaskName_Challenge = "Challenge" //挑战任务
)

// 任务组解锁条件
const (
	TaskUnlock_Grade      = 1 //等级解锁
	TaskUnlock_PreGroup   = 2 //完成前置任务组
	TaskUnlock_Elite      = 3 //购买英雄勋章
	TaskUnlock_PastDay    = 4 //赛季开始x天后
	TaskUnlock_EnableItem = 5 //拥有激活道具
)

// 任务项 (任务名后的数字表示单局需要达成的数量)
const (
	TaskItem_Login           uint32 = 1  //登陆
	TaskItem_Topten          uint32 = 2  //进入前十
	TaskItem_Game            uint32 = 3  //参加演习
	TaskItem_TeamTopten      uint32 = 4  //在组队模式中进入前十
	TaskItem_Live            uint32 = 5  //存活时间(秒)
	TaskItem_Move            uint32 = 6  //移动距离(米)
	TaskItem_KillEnemy       uint32 = 7  //击败敌人
	TaskItem_ShareResult     uint32 = 8  //分享战绩
	TaskItem_DamageEnemy     uint32 = 9  //对敌人造成伤害
	TaskItem_RescueTeammate  uint32 = 10 //救起队友
	TaskItem_RecoverHp       uint32 = 11 //恢复血量
	TaskItem_Crit            uint32 = 12 //暴击
	TaskItem_DoubleKill      uint32 = 13 //在演习中打出2连击
	TaskItem_OpenBox         uint32 = 14 //打开补给箱
	TaskItem_FriendGame      uint32 = 15 //与好友邀请组队进行演习
	TaskItem_InviteUpLine    uint32 = 16 //邀请好友上线
	TaskItem_FinishTask      uint32 = 17 //完成任务
	TaskItem_CostCoin        uint32 = 18 //消费金币
	TaskItem_CostDiam        uint32 = 19 //消费钻石
	TaskItem_KillEnemy3      uint32 = 20 //单局击杀3人
	TaskItem_Live800         uint32 = 21 //单局存活800秒
	TaskItem_Live1500        uint32 = 22 //单局存活1500秒
	TaskItem_Move20000       uint32 = 23 //单局移动20000米
	TaskItem_Win             uint32 = 24 //吃鸡
	TaskItem_RescueTeammate3 uint32 = 25 //单局救起3人
	TaskItem_RescueTeammate5 uint32 = 26 //单局救起5人
	TaskItem_TripleKill      uint32 = 27 //完成一次3连杀
	TaskItem_RecoverHp300    uint32 = 28 //单局恢复300血
	TaskItem_RecoverHp1000   uint32 = 29 //单局恢复1000血
	TaskItem_RPGKill         uint32 = 30 //rpg击杀
	TaskItem_GetDropBox      uint32 = 31 //拾取空投
	TaskItem_GetDropBox5     uint32 = 32 //单局拾取空投5次
	TaskItem_GilleyWin       uint32 = 33 //穿吉利服获胜
	TaskItem_SignalGun       uint32 = 34 //使用信号枪
	TaskItem_SignalGun3      uint32 = 35 //单局使用信号枪3次
	TaskItem_ShootBullet500  uint32 = 36 //单局使用子弹500个
	TaskItem_FistKill5       uint32 = 37 //单局用拳头击倒5人
	TaskItem_TankKill        uint32 = 38 //坦克击杀
	TaskItem_Swim            uint32 = 39 //游泳
	TaskItem_ParachuteDie    uint32 = 40 //跳伞时被击败
	TaskItem_PistolKill5     uint32 = 41 //单局用手枪击败5人
	TaskItem_CarKill8        uint32 = 42 //单局用载具击败8人
	TaskItem_FallDamage      uint32 = 43 //坠落伤害
	TaskItem_FallDamage500   uint32 = 44 //单局坠落受伤500点
	TaskItem_EnergyDrink20   uint32 = 45 //单局拥有20瓶能量饮料
	TaskItem_AWMKill8        uint32 = 46 //单局用狙击枪击败8人
	TaskItem_LandDie         uint32 = 47 //落地后10秒死亡
)

// 活动任务项
const (
	Act_RescueNum    uint32 = 1  //累计救援次数
	Act_ScuffleNum   uint32 = 2  //累计参加快速模式
	Act_RunDistance  uint32 = 3  //累计移动距离(米)
	Act_KillNum      uint32 = 4  //累计击败玩家
	Act_Headshotnum  uint32 = 5  //累计暴击玩家
	Act_RankTen      uint32 = 6  //累计进入前10名
	Act_Finish       uint32 = 7  //已达成目标
	Act_SurviveTime  uint32 = 8  //累计存活(分钟)
	Act_WinNum       uint32 = 9  //累计胜利
	Act_Effectharm   uint32 = 10 //累计伤害
	Act_Login        uint32 = 11 //累计每日登陆次数
	Act_GameNum      uint32 = 12 //累计参加演习次数
	Act_ShareNum     uint32 = 13 //累计分享战绩次数
	Act_GameTeamNum  uint32 = 14 //累计组队参加演习次数
	Act_OpenBox      uint32 = 15 //累计开启补给箱次数
	Act_RecoverHp    uint32 = 16 //累计恢复血量
	Act_RankTeamTen  uint32 = 17 //累计组队进入前10名
	Act_MultiKillnum uint32 = 18 //累计打出二连击次数
	Act_LoginNoon    uint32 = 19 //累计12:00-14:00登录
	Act_LoginNight   uint32 = 20 //累计18:00-21:00登录
)

// Ai系统配置参数
const (
	AiSys_RandomAiPosFloor = 8 // ai位置下限
	AiSys_RandomAiPosCeil  = 9 // ai位置上限
)

const (
	AdditionBonus_ComradeExp  = 1 //战友组队经验加成
	AdditionBonus_ComradeCoin = 2 //战友组队金币加成
)

// 技能系统常量
const (
	SkillSystem_OpenModeID    = 1 // 技能开放的匹配模式
	SkillSystem_WarSongArea   = 2 // 战歌触发范围（米）
	SkillSystem_WarSongPeople = 3 // 战歌触发人数上限
	SkillSystem_UAVArea       = 4 // 无人机触发范围（米）
)

// 是否主动技能
const (
	Skill_Passive = 0 //被动技能
	Skill_Initive = 1 //主动技能
)

// 技能类型
const (
	Skill_Role = 0 //主角技能
	Skill_Call = 1 //调用技能
	Skill_NPC  = 2 //npc技能
)

const (
	PaySystem_MidasL5SID1     = 1  //米大师L5的SID 正式环境
	PaySystem_MidasL5SID2     = 2  //米大师L5的SID 测试环境
	PaySystem_MidasL5SID3     = 3  //米大师L5的SID 容灾演练
	PaySystem_LevelReset      = 4  //首次充送重置轮次
	PaySystem_MidasIOSAppKey1 = 6  //米大师IOS系统Appkey 正式环境
	PaySystem_MidasANDAppKey1 = 7  //米大师AND系统Appkey 正式环境
	PaySystem_MidasIOSAppKey2 = 8  //米大师IOS系统Appkey 测试环境
	PaySystem_MidasANDAppKey2 = 9  //米大师AND系统Appkey 测试环境
	PaySystem_PayEnv          = 10 //充值环境 1正式环境 2测试环境 3容灾演练
)

const (
	MarketingNameFirstPay     = "FirstPay"     //首充
	MarketingNameMonthCardIOS = "MonthCardIOS" //月卡(IOS系统)
	MarketingNameMonthCardAND = "MonthCardAND" //月卡(Android系统)
	MarketingNamePayLevel     = "PayLevel"     //首次充送
)

const (
	MarketingTypeFirstPay  = 1 //首充
	MarketingTypeMonthCard = 2 //月卡
	MarketingTypePayLevel  = 3 //首次充送
)

const (
	MarketingFirstPay      = 1 //首充赠送奖励
	MarketingMonthCardDay  = 2 //月卡日结奖励
	MarketingMonthCardOnce = 3 //月卡即得奖励
)

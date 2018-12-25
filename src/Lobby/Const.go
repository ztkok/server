package main

const (
	// UserStatusInRoomGaming 玩家游戏中
	UserStatusInRoomGaming = 1
	// UserStatusInRoomQuit 玩家离线
	UserStatusInRoomQuit = 2
	// UserStatusInRoomAi 玩家在被托管状态
	UserStatusInRoomAi = 3
)

// 匹配容器状态
const (
	MatchModelInit = 0
	MatchModelDel  = 1
)

const (
	// TeamCreate 队伍组建
	TeamCreate = 1
	// TeamStart 队伍开始使用
	TeamStart = 2
	// TeamEnd 队伍结束
	TeamEnd = 3
)

//匹配类型
const (
	//MatchKind_Single 单人匹配
	MatchKindSingle = 1
	//MatchKind_Team 组队匹配
	MatchKindTeam = 2
)

const (
	// TwoTeamType 双排
	TwoTeamType = 0
	// FourTeamType 四排
	FourTeamType = 1
)

const (
	// TeamNotMatch 队伍在组队中
	TeamNotMatch = 0
	// TeamMatching 队伍在匹配中
	TeamMatching = 1
)

const (
	// MemberNotReady 成员未准备
	MemberNotReady = 0
	// MemberReadying 成员准备中
	MemberReadying = 1
)

const (
	//目标设置不能被观战
	WatchBattleErrCodeNotAllow = 1
	//观战游戏不存在
	WatchBattleErrCodeNotInGame = 2
	//游戏尚未开始
	WatchBattleErrCodeWaitTime = 3
	//观战人数上限
	WatchBattleErrCodeNumLimit = 4
	//当前不能进行观战
	WatchBattleErrCodeCanNot = 5
)

const (
	//PlayerSettingWatchAble 玩家设置是否可以被观战
	PlayerSettingWatchAble = 1
)

const (
	GoodChannel_Normal     = 1 //普通商品
	GoodChannel_Special    = 2 //特惠商品
	GoodChannel_SeasonPass = 3 //季票商品
)

const (
	GoodType_7Days   = 1 //7天商品
	GoodType_30Days  = 2 //30天商品
	GoodType_Forever = 3 //永久商品
)

// 特惠类型
const (
	SaleType_Day1 = 1 //1=天刷新
	SaleType_Day7 = 7 //7=周刷新
)

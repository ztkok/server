package datadef

import (
	"fmt"
)

// OneRoundData 一局数据表
type OneRoundData struct {
	GameID              uint64  `json:"GameID"`        //本局id
	UID                 uint64  `json:"UID"`           // 用户id
	Season              int     `json:"Season"`        //赛季
	Rank                uint32  `json:"Rank"`          // 本场排名
	KillNum             uint32  `json:"KillNum"`       // 本场击杀数
	HeadShotNum         uint32  `json:"HeadShotNum"`   // 本场爆头数
	EffectHarm          uint32  `json:"EffectHarm"`    // 本场有效伤害
	RecoverUseNum       uint32  `json:"RecoverUseNum"` // 本场治疗道具使用次数
	ShotNum             uint32  `json:"ShotNum"`       //本场开枪次数
	ReviveNum           uint32  `json:"ReviveNum"`     //本场复活道具使用次数
	KillDistance        float32 `json:"KillDistance"`  //本场最远击杀距离
	KillStmNum          uint32  `json:"KillStmNum"`    //本场最大连杀次数
	FinallHp            uint32  `json:"FinallHp"`      //本场最终血量
	RecoverHp           uint32  `json:"RecoverHp"`     //本场总复活血量
	RunDistance         float32 `json:"RunDistance"`   //本场行进距离
	CarUseNum           uint32  `json:"CarUseNum"`     //本场载具使用数量
	CarDestoryNum       uint32  `json:"CarDestoryNum"` //本场载具摧毁数量
	AttackNum           uint32  `json:"AttackNum"`     //本场助攻次数
	SpeedNum            uint32  `json:"SpeedNum"`      //本场加速次数
	Coin                uint32  `json:"Coin"`          //本场奖励金币
	StartUnix           int64   `json:"StartUnix"`
	EndUnix             int64   `json:"EndUnix"`
	StartTime           string  `json:"StartTime"`  //本场开始时间
	EndTime             string  `json:"EndTime"`    //本场结束时间
	BattleType          uint32  `json:"BattleType"` //本场战斗类型（单双四）
	KillRating          float32 `json:"KillRating"` //本场击杀rating
	WinRatig            float32 `json:"WinRatig"`   //本场胜利rating
	SoloKillRating      float32 `json:"SoloKillRating"`
	SoloWinRating       float32 `json:"SoloWinRating"`
	SoloRating          float32 `json:"SoloRating"` //单人总rating分
	SoloRank            uint32  `json:"SoloRank"`   //单人全服排名
	DuoKillRating       float32 `json:"DuoKillRating"`
	DuoWinRating        float32 `json:"DuoWinRating"`
	DuoRating           float32 `json:"DuoRating"` //双人总rating分
	DuoRank             uint32  `json:"DuoRank"`   //双人全服排名
	SquadKillRating     float32 `json:"SquadKillRating"`
	SquadWinRating      float32 `json:"SquadWinRating"`
	SquadRating         float32 `json:"SquadRating"` //四人总rating分
	SquadRank           uint32  `json:"SquadRank"`   //四人全服排名
	TotalRating         float32 `json:"TotalRating"` //总rating分（单双四）
	TotalRank           uint32  `json:"TotalRank"`   //全服总排行
	TopRating           float32 `json:"TopRating"`   //全服最高rating分
	TotalBattleNum      uint32  `json:"TotalBattleNum"`
	TotalFirstNum       uint32  `json:"TotalFirstNum"`
	TotalTopTenNum      uint32  `json:"TotalTopTenNum"`
	TotalHeadShot       uint32  `json:"TotalHeadShot"`
	TotalKillNum        uint32  `json:"TotalKillNum"`
	TotalShotNum        uint32  `json:"TotalShotNum"`
	TotalEffectHarm     uint32  `json:"TotalEffectHarm"`
	TotalSurviveTime    int64   `json:"TotalSurviveTime"`
	TotalDistance       float32 `json:"TotalDistance"`
	TotalRecvItemUseNum uint32  `json:"TotalRecvItemUseNum"`
	SingleMaxKill       uint32  `json:"SingleMaxKill"`
	SingleMaxHeadShot   uint32  `json:"SingleMaxHeadShot"`
	DeadType            uint32  `json:"DeadType"`
	UserName            string  `json:"UserName"` //用户名称
	SkyBox              uint32  `json:"SkyBox"`
	PlatID              uint32  `json:"PlatID"`
	LoginChannel        uint32  `json:"LoginChannel"`
}

func (round *OneRoundData) String() string {
	return fmt.Sprintf("%+v\n", *round)
}

// OneDayData 一天记录数据表
type OneDayData struct {
	DayId          uint64  `json:"DayId"`
	Season         int     `json:"Season"`
	UID            uint64  `json:"UID"`
	NowTime        int64   `json:"NowTime"`
	StartTime      string  `json:"StartTime"`
	Model          uint32  `json:"Model"`
	DayFirstNum    uint32  `json:"DayFirstNum"`
	DayTopTenNum   uint32  `json:"DayTopTenNum"`
	Rating         float32 `json:"Rating"`
	WinRating      float32 `json:"WinRating"`
	KillRating     float32 `json:"KillRating"`
	DayEffectHarm  uint32  `json:"DayEffectHarm"`
	DayShotNum     uint32  `json:"DayShotNum"`
	DaySurviveTime int64   `json:"DaySurviveTime"`
	DayDistance    float32 `json:"DayDistance"`
	DayAttackNum   uint32  `json:"DayAttackNum"`
	DayRecoverNum  uint32  `json:"DayRecoverNum"`
	DayRevivenum   uint32  `json:"DayRevivenum"`
	DayHeadShotNum uint32  `json:"DayHeadShotNum"`
	DayBattleNum   uint32  `json:"DayBattleNum"`
	DayKillNum     uint32  `json:"DayKillNum"` //总击杀数量
	TotalRank      uint32  `json:"TotalRank"`
	Rank           uint32  `json:"Rank"`     //该模式的排行
	UserName       string  `json:"UserName"` //用户名称
}

func (daydata *OneDayData) String() string {
	return fmt.Sprintf("%+v\n", *daydata)
}

// SearchNumData 玩家搜索次数表
type SearchNumData struct {
	UID         uint64  `json:"UID"`
	UserName    string  `json:"UserName"`
	SearchNum   uint32  `json:"SearchNum"`
	TotalRating float32 `json:"TotalRating"`
}

func (searchdata *SearchNumData) String() string {
	return fmt.Sprintf("%+v\n", *searchdata)
}

// GameCareer 生涯概况
type CareerData struct {
	Uid              uint64  `json:"Uid"`              //玩家UID
	TotalBattleNum   uint32  `json:"TotalBattleNum"`   //总场次
	TotalFirstNum    uint32  `json:"TotalFirstNum"`    //吃鸡场次
	TotalTopTenNum   uint32  `json:"TotalTopTenNum"`   //前10场次
	TotalKillNum     uint32  `json:"TotalKillNum"`     //总击杀数量
	TotalHeadShot    uint32  `json:"TotalHeadShot"`    //累积爆头数量
	TotalShotNum     uint32  `json:"TotalShotNum"`     //总开枪次数
	TotalEffectHarm  uint32  `json:"TotalEffectHarm"`  //累积有效伤害
	TotalSurviveTime int64   `json:"TotalSurviveTime"` //总生存时间
	TotalDistance    float32 `json:"TotalDistance"`    //总行进距离
	SoloRating       float32 `json:"SoloRating"`
	SoloRank         uint32  `json:"SoloRank"` //单人排名
	DuoRating        float32 `json:"DuoRating"`
	DuoRank          uint32  `json:"DuoRank"` //双人排名
	SquadRating      float32 `json:"SquadRating"`
	SquadRank        uint32  `json:"SquadRank"`   //四人排名
	TotalRating      float32 `json:"TotalRating"` //综合评分
	TotalRank        uint32  `json:"TotalRank"`   //综合分排名
	TopRating        float32 `json:"TopRating"`   //全服最高评分
	UserName         string  `json:"UserName"`
	Url              string  `json:"Url"`
	QqVip            uint32  `json:"QqVip"`
	GameEenter       string  `json:"GameEenter"`
	Gender           uint8   `json:"Gender"` //性别，0表示未知，1表示男，2表示女
	Level            uint32  `json:"Level"`  //当前级别
	Exp              uint32  `json:"Exp"`    //当前级别积累的经验值
	MaxExp           uint32  `json:"MaxExp"` //当前级别升到下一级需要积累的经验值
	NameColor        uint32  `json:"NameColor"`
}

func (careerdata *CareerData) String() string {
	return fmt.Sprintf("%+v\n", *careerdata)
}

// SyncFriendRankList 同步好友信息列表
type SyncFriendRankList struct {
	Item []*FriendRankInfo `json:"Item"`
}

// FriendRankInfo 好友排名信息
type FriendRankInfo struct {
	Uid         uint64  `json:"Uid"`
	Name        string  `json:"Name"`
	Url         string  `json:"Url"` // 好友头像url
	SoloRating  float32 `json:"SoloRating"`
	DuoRating   float32 `json:"DuoRating"`
	SquadRating float32 `json:"SquadRating"`
	QqVip       uint32  `json:"QqVip"`
	GameEenter  string  `json:"GameEenter"`
	Level       uint32  `json:"Level"` //当前级别
	NameColor   uint32  `json:"NameColor"`
}

// MatchRecord 比赛记录
type MatchRecord struct {
	Matchrecord []*DayRecordData `json:"Matchrecord"` //比赛记录
}

// DayRecordData 每天记录信息
type DayRecordData struct {
	DayId        uint64  `json:"DayId"`   //该天的记录ID
	NowTime      int64   `json:"NowTime"` //战斗时间(那一天)
	Model        uint32  `json:"Model"`
	Rating       float32 `json:"Rating"`
	DayKillNum   uint32  `json:"DayKillNum"`   //总击杀数量
	DayBattleNum uint32  `json:"DayBattleNum"` //该模式当天战斗场次
	DayFirstNum  uint32  `json:"DayFirstNum"`  //该模式吃鸡场次
	DayTopTenNum uint32  `json:"DayTopTenNum"` //前10场次
}

// SettleDayData 每天结算信息
type SettleDayData struct {
	DayId          uint64    `json:"DayId"`        //该天的记录ID
	NowTime        int64     `json:"NowTime"`      //战斗时间(那一天)
	DayFirstNum    uint32    `json:"DayFirstNum"`  //该模式吃鸡场次
	DayTopTenNum   uint32    `json:"DayTopTenNum"` //前10场次
	Rating         float32   `json:"Rating"`
	WinRating      float32   `json:"WinRating"`      //winrating分
	KillRating     float32   `json:"KillRating"`     //killrating分
	DayEffectHarm  uint32    `json:"DayEffectHarm"`  //累积有效伤害
	DayShotNum     uint32    `json:"DayShotNum"`     //当天开枪次数
	DaySurviveTime int64     `json:"DaySurviveTime"` //总生存时间
	DayDistance    float32   `json:"DayDistance"`    //总行进距离
	DayAttackNum   uint32    `json:"DayAttackNum"`   //助攻次数
	DayRecoverNum  uint32    `json:"DayRecoverNum"`  //治疗道具使用次数
	DayRevivenum   uint32    `json:"DayRevivenum"`   //复活次数
	DayHeadShotNum uint32    `json:"DayHeadShotNum"` //爆头数
	DayBattleNum   uint32    `json:"DayBattleNum"`   //该模式当天战斗场次
	TotalRank      uint32    `json:"TotalRank"`      //全服排名
	ServerType     string    `json:"ServerType"`     //服务器类型
	Tag            []*DayTag `json:"Tag"`            //每日标签
}

func (settledata *SettleDayData) String() string {
	return fmt.Sprintf("%+v\n", *settledata)
}

// DayTag 每日标签
type DayTag struct {
	Tag string `json:"Tag"`
}

// RankTrend rating分趋势
type RankTrend struct {
	DaysRating []*DayRating `json:"DaysRating"`
}

// DayRating 那一天的rating分
type DayRating struct {
	Uid       uint64  `json:"Uid"`
	StartTime string  `json:"StartTime"`
	Rating    float32 `json:"Rating"`
}

// ModelDetail 模式详情
type ModelDetail struct {
	Uid            uint64  `json:"Uid"`
	SoloRating     float32 `json:"SoloRating"`
	SoloRank       uint32  `json:"SoloRank"`
	SoloFirstNum   uint32  `json:"SoloFirstNum"`
	SoloTopTenNum  uint32  `json:"SoloTopTenNum"`
	SoloBattleNum  uint32  `json:"SoloBattleNum"`
	SoloKillNum    uint32  `json:"SoloKillNum"`
	DuoRating      float32 `json:"DuoRating"`
	DuoRank        uint32  `json:"DuoRank"`
	DuoFirstNum    uint32  `json:"DuoFirstNum"`
	DuoTopTenNum   uint32  `json:"DuoTopTenNum"`
	DuoBattleNum   uint32  `json:"DuoBattleNum"`
	DuoKillNum     uint32  `json:"DuoKillNum"`
	SquadRating    float32 `json:"SquadRating"`
	SquadRank      uint32  `json:"SquadRank"`
	SquadFirstNum  uint32  `json:"SquadFirstNum"`
	SquadTopTenNum uint32  `json:"SquadTopTenNum"`
	SquadBattleNum uint32  `json:"SquadBattleNum"`
	SquadKillNum   uint32  `json:"SquadKillNum"`
}

/**********************************赛季排行***************************/
// 玩家赛季积分排行数据
type PlayerRatingRankInfo struct {
	Uid            uint64  `json:"Uid"`
	Name           string  `json:"Name"`
	NameColor      uint32  `json:"NameColor"`
	Url            string  `json:"Url"` // 好友头像url
	QqVip          uint32  `json:"QqVip"`
	GameEenter     string  `json:"GameEenter"`
	Rank           uint32  `json:"Rank"`           //名次
	Rating         uint32  `json:"Rating"`         //赛季积分
	KDA            float32 `json:"KDA"`            //战损比
	TotalBattleNum uint32  `json:"TotalBattleNum"` //总场次
}

type SeasonRatingRank struct {
	Season  int                     //排行赛季id
	Infos   []*PlayerRatingRankInfo `json:"Infos"`   //全服前N名赛季积分排行数据
	ReqInfo *PlayerRatingRankInfo   `json:"ReqInfo"` //请求者排行数据
}

// 玩家赛季获胜数排行数据
type PlayerWinsRankInfo struct {
	Uid            uint64 `json:"Uid"`
	Name           string `json:"Name"`
	NameColor      uint32 `json:"NameColor"`
	Url            string `json:"Url"` //好友头像url
	QqVip          uint32 `json:"QqVip"`
	GameEenter     string `json:"GameEenter"`
	Rank           uint32 `json:"Rank"`           //名次
	Wins           uint32 `json:"Wins"`           //赛季获胜数
	TopTenNum      uint32 `json:"TopTenNum"`      //赛季前十数
	TotalBattleNum uint32 `json:"TotalBattleNum"` //总场次
}

type SeasonWinsRank struct {
	Season  int                   //排行赛季id
	Infos   []*PlayerWinsRankInfo `json:"Infos"`   //全服前N名赛季获胜数排行数据
	ReqInfo *PlayerWinsRankInfo   `json:"ReqInfo"` //请求者排行数据
}

// 玩家赛季击杀数排行数据
type PlayerKillsRankInfo struct {
	Uid            uint64 `json:"Uid"`
	Name           string `json:"Name"`
	NameColor      uint32 `json:"NameColor"`
	Url            string `json:"Url"` // 好友头像url
	QqVip          uint32 `json:"QqVip"`
	GameEenter     string `json:"GameEenter"`
	Rank           uint32 `json:"Rank"`           //名次
	Kills          uint32 `json:"Kills"`          //赛季击杀数
	MaxKills       uint32 `json:"MaxKills"`       //赛季单场最多击杀数
	TotalBattleNum uint32 `json:"TotalBattleNum"` //总场次
}

type SeasonKillsRank struct {
	Season  int                    //排行赛季id
	Infos   []*PlayerKillsRankInfo `json:"Infos"`   //全服前N名赛季击杀数排行数据
	ReqInfo *PlayerKillsRankInfo   `json:"ReqInfo"` //请求者排行数据
}

/**********************************redis存储数据struct***************************/

// CareerBase 生涯数据录入redis哈希表PlayerCareerData表中
type CareerBase struct {
	UID                uint64
	UserName           string
	TotalBattleNum     uint32
	FirstNum           uint32
	FirstStamp         int64 //玩家在本局比赛中获胜，记录时间戳
	TopTenNum          uint32
	TotalKillNum       uint32
	KillStamp          int64 //玩家在本局比赛中击杀了敌人，记录时间戳
	TotalHeadShot      uint32
	Totalshotnum       uint32
	TotalEffectHarm    uint32
	SurviveTime        int64
	TotalDistance      float32
	SoloWinRating      float32
	SoloKillRating     float32
	DuoWinRating       float32
	DuoKillRating      float32
	SquadWinRating     float32
	SquadKillRating    float32
	SoloRating         float32
	DuoRating          float32
	SquadRating        float32
	TotalRating        float32
	TotalRank          uint32
	SoloRank           uint32
	DuoRank            uint32
	SquadRank          uint32
	TopRating          float32
	SingleMaxKill      uint32
	SingleMaxHeadShot  uint32
	RecvItemUseNum     uint32
	CarUseNum          uint32
	CarDestroyNum      uint32
	Coin               uint32
	TotalCarDistance   float32
	TotalTopRating     float32 //历史最高rating
	TotalTopRank       uint32  //历史最高排名
	TotalTopSoloRating float32 //历史最高solorating
	TotalTopSoloRank   uint32  //历史最高solorank
	TotalTopDuoRating  float32 //历史最高DuoRating
	TotalTopDuoRank    uint32
	TotalTopSquadRaing float32
	TotalTopSquadRank  uint32
	WinStreak          uint32 //连胜次数
	FailStreak         uint32 //连败次数
}

func (careerBase *CareerBase) String() string {
	return fmt.Sprintf("%+v\n", *careerBase)
}

// DayData 一天数据录入redis中PlayerDayData表中
type DayData struct {
	UID            uint64
	UserName       string //用户名称
	Season         int
	NowTime        int64
	StartTime      string
	DayFirstNum    uint32
	DayTopTenNum   uint32
	DayEffectHarm  uint32
	DayShotNum     uint32
	DaySurviveTime int64
	DayDistance    float32
	DayCarDistance float32
	DayAttackNum   uint32
	DayRecoverNum  uint32
	DayRevivenum   uint32
	DayHeadShotNum uint32
	DayBattleNum   uint32
	DayKillNum     uint32 //总击杀数量
	DayCoin        uint32
	TotalRating    float32
	TotalRank      uint32
}

func (dayData *DayData) String() string {
	return fmt.Sprintf("%+v\n", *dayData)
}

/**********************************************[手Q]分享宝箱接口******************************************/

// QQChestReq [手Q]分享宝箱请求
type QQChestReq struct {
	AppID       string `json:"appid"`       //应用在平台的唯一id
	OpenID      string `json:"openid"`      //用户在某个应用的唯一标识
	AccessToken string `json:"accessToken"` //	第三方调用凭证，通过获取凭证接口获得
	Pf          string `json:"pf"`          //宝箱发送者平台信息
	Actid       uint32 `json:"actid"`       //活动号
	Num         uint32 `json:"num"`         //物品的总数量
	Peoplenum   uint32 `json:"peoplenum"`   //人数
	Type        uint32 `json:"type"`        //宝箱类型，填0
	Secret      string `json:"secret"`      //手Q申请的secret；游戏接入该功能的时候由对应腾讯产品经理联系手Q获取，非appkey
}

func (qqChestReq *QQChestReq) String() string {
	return fmt.Sprintf("%+v\n", *qqChestReq)
}

// QQChestRsp [手Q]分享宝箱应答
type QQChestRsp struct {
	Ret   int32  `json:"ret"`   //返回码 0：正确，其它：失败
	Msg   string `json:"msg"`   //ret非0，则表示“错误码，错误提示”，详细注释参见错误码描述
	BoxID string `json:"boxid"` //宝箱id
}

/**********************************************[微信]分享接口******************************************/

// WXChestRsp [微信]创建福袋接口应答
type WXChestRsp struct {
	Ret  int32  `json:"ret"`  //返回码 0：正确，其它：失败
	Msg  string `json:"msg"`  //ret非0，则表示“错误码，错误提示”，详细注释参见错误码描述
	Data Body   `json:"data"` //
}

func (wxChestRsp *WXChestRsp) String() string {
	return fmt.Sprintf("%+v\n", *wxChestRsp)
}

// Body [微信]创建福袋接口应答包体
type Body struct {
	Url string `json:"url"` //	url
}

// SyncBraveRankList 勇者战场信息列表
type SyncBraveRankList struct {
	Item []*BraveBattleRankInfo `json:"Item"`
}

// BraveBattleRankInfo 勇者战场排名信息
type BraveBattleRankInfo struct {
	Uid        uint64 `json:"Uid"`
	Name       string `json:"Name"`
	Url        string `json:"Url"` // 好友头像url
	Brave      uint32 `json:"Brave"`
	QqVip      uint32 `json:"QqVip"`
	GameEenter string `json:"GameEenter"`
	Rank       uint32 `json:"Rank"`
	NameColor  uint32 `json:"NameColor"`
}

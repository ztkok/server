package main

import (
	"common"
	"db"
	"excel"
	"fmt"
	"protoMsg"
	"strconv"
	"strings"
	"time"
	"zeus/dbservice"
	"zeus/tlog"
)

// tlogLiveFlow 个人信息流水数据打印
func (user *LobbyUser) tlogLiveFlow() {
	if user.loginMsg != nil {
		msg := &protoMsg.LiveFlow{}
		msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
		msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
		msg.VGameAppID = user.loginMsg.VGameAppID
		if msg.VGameAppID == "" {
			msg.VGameAppID = "NoVGameAppID"
		}
		msg.PlatID = user.loginMsg.PlatID
		msg.IZoneAreaID = 0
		msg.VOpenID = user.loginMsg.VOpenID
		if msg.VOpenID == "NoOpenID" || msg.VOpenID == "" {
			msg.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
		}
		msg.VRoleName = user.GetName()
		if msg.VRoleName == "" {
			msg.VRoleName = "no username"
		} else {
			r := strings.NewReplacer("|", "-", "｜", "+", ",", "=")
			msg.VRoleName = r.Replace(msg.VRoleName)
		}
		msg.Level = user.GetLevel()
		msg.PlayerFriendsNum = user.GetFriendsNum()
		msg.ClientVersion = user.loginMsg.ClientVersion
		if msg.ClientVersion == "" {
			msg.ClientVersion = "1.0.0"
		}
		msg.SystemHardware = user.loginMsg.SystemHardware
		if msg.SystemHardware == "" {
			msg.SystemHardware = "NoSysHardware"
		}
		msg.TelecomOper = user.loginMsg.TelecomOper
		if msg.TelecomOper == "" {
			msg.TelecomOper = "NoOp"
		}
		msg.Network = user.loginMsg.Network
		if msg.Network == "" {
			msg.Network = "NoNetwork"
		} else {
			msg.Network = strings.Replace(msg.Network, "|", "-", -1)
		}
		msg.LoginChannel = user.loginMsg.LoginChannel

		season := common.GetSeason()
		careerdata, err := common.GetPlayerCareerData(user.GetDBID(), season, 0)
		if err != nil {
			user.Error("GetPlayerCareerData err: ", err)
			return
		}

		msg.Score = careerdata.TotalRating
		msg.ScoreRank = careerdata.TotalRank
		msg.Touts = careerdata.TotalBattleNum
		msg.Wins = careerdata.FirstNum
		msg.Topwins = careerdata.TopTenNum
		if careerdata.TotalBattleNum > careerdata.FirstNum {
			msg.KD = careerdata.TotalKillNum / (careerdata.TotalBattleNum - careerdata.FirstNum)
		}
		if careerdata.TotalBattleNum != 0 {
			msg.BeatAvg = careerdata.TotalKillNum / careerdata.TotalBattleNum
			msg.WinRate = careerdata.FirstNum / careerdata.TotalBattleNum
			msg.TopWinRate = careerdata.TopTenNum / careerdata.TotalBattleNum
			msg.HurtAvg = careerdata.TotalEffectHarm / careerdata.TotalBattleNum
			msg.TimeAvg = careerdata.SurviveTime / int64(careerdata.TotalBattleNum)
			msg.DistanceAvg = careerdata.TotalDistance / float32(careerdata.TotalBattleNum)
		}
		if careerdata.TotalKillNum != 0 {
			msg.CriticalRate = careerdata.TotalHeadShot / careerdata.TotalKillNum
		}
		msg.BestScore = careerdata.TotalTopRating
		msg.BestRank = careerdata.TotalTopRank
		msg.SingBestScore = careerdata.TotalTopSoloRating
		msg.SingBestRank = careerdata.TotalTopSoloRank
		msg.DuoBestScore = careerdata.TotalTopDuoRating
		msg.DuoBestRank = careerdata.TotalTopDuoRank
		msg.SquadBestScore = careerdata.TotalTopSquadRaing
		msg.SquadRank = careerdata.TotalTopSquadRank

		tlog.Format(msg)
	} else {
		user.Warn("tlogLiveFlow loginMsg is nil")
	}
}

/***************************tlog聊天流水表********************************/

/*
	openMic: 打开麦克风操作为0，1代表关闭
	chatType: 1系统、2世界、3组队
	msgType: 0 为文字信息，1 为语音信息，2为快捷文字信息
*/

// tlogChatFlow tlog聊天流水表
func (user *LobbyUser) tlogChatFlow(openMic, chatType, msgType uint32) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg != nil {
		msg := &protoMsg.ChatFlow{}
		msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
		msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
		msg.VGameAppID = loginMsg.VGameAppID
		if msg.VGameAppID == "" {
			msg.VGameAppID = "NoVGameAppID"
		}
		msg.PlatID = loginMsg.PlatID
		msg.IZoneAreaID = 0
		msg.VOpenID = loginMsg.VOpenID
		if msg.VOpenID == "NoOpenID" || msg.VOpenID == "" {
			msg.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
		}
		msg.VRoleName = user.GetName()
		if msg.VRoleName == "" {
			msg.VRoleName = "no username"
		} else {
			r := strings.NewReplacer("|", "-", "｜", "+", ",", "=")
			msg.VRoleName = r.Replace(msg.VRoleName)
		}
		msg.VRoleProfession = user.GetRoleModel()
		msg.Level = user.GetLevel()
		msg.PlayerFriendsNum = user.GetFriendsNum()
		msg.ClientVersion = loginMsg.ClientVersion
		if msg.ClientVersion == "" {
			msg.ClientVersion = "1.0.0"
		}
		msg.SystemHardware = loginMsg.SystemHardware
		if msg.SystemHardware == "" {
			msg.SystemHardware = "NoSysHardware"
		}
		msg.TelecomOper = loginMsg.TelecomOper
		if msg.TelecomOper == "" {
			msg.TelecomOper = "NoOp"
		}
		msg.Network = loginMsg.Network
		if msg.Network == "" {
			msg.Network = "NoNetwork"
		} else {
			msg.Network = strings.Replace(msg.Network, "|", "-", -1)
		}
		msg.LoginChannel = loginMsg.LoginChannel
		msg.OpenMic = openMic
		msg.ChatType = chatType
		msg.MsgType = msgType
		msg.BattleID = user.spaceID
		msg.TeamID = user.GetTeamID()

		tlog.Format(msg)
	} else {
		user.Warn("tlogChatFlow loginMsg is nil")
	}
}

/***************************tlog操作模式表********************************/

// tlogOperFlow tlog操作模式表
func (user *LobbyUser) tlogOperFlow(msg *protoMsg.OperFlow) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg != nil {
		// msg := &protoMsg.OperFlow{}
		msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
		msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
		msg.VGameAppID = loginMsg.VGameAppID
		if msg.VGameAppID == "" {
			msg.VGameAppID = "NoVGameAppID"
		}
		msg.PlatID = loginMsg.PlatID
		msg.IZoneAreaID = 0
		msg.VOpenID = loginMsg.VOpenID
		if msg.VOpenID == "NoOpenID" || msg.VOpenID == "" {
			msg.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
		}
		msg.VRoleName = user.GetName()
		if msg.VRoleName == "" {
			msg.VRoleName = "no username"
		} else {
			r := strings.NewReplacer("|", "-", "｜", "+", ",", "=")
			msg.VRoleName = r.Replace(msg.VRoleName)
		}
		msg.VRoleProfession = user.GetRoleModel()
		msg.Level = user.GetLevel()
		msg.PlayerFriendsNum = user.GetFriendsNum()
		msg.ClientVersion = loginMsg.ClientVersion
		if msg.ClientVersion == "" {
			msg.ClientVersion = "1.0.0"
		}
		msg.SystemHardware = loginMsg.SystemHardware
		if msg.SystemHardware == "" {
			msg.SystemHardware = "NoSysHardware"
		}
		msg.TelecomOper = loginMsg.TelecomOper
		if msg.TelecomOper == "" {
			msg.TelecomOper = "NoOp"
		}
		msg.Network = loginMsg.Network
		if msg.Network == "" {
			msg.Network = "NoNetwork"
		} else {
			msg.Network = strings.Replace(msg.Network, "|", "-", -1)
		}
		msg.LoginChannel = loginMsg.LoginChannel
		if user.GetPlayerGameState() == common.StateGame {
			msg.OperScene = 1
		} else {
			msg.OperScene = 0
		}

		tlog.Format(msg)
	} else {
		user.Warn("tlogOperFlow loginMsg is nil")
	}
}

/***************************tlog战场匹配表********************************/
/*
	matchType: 匹配模式 0单人，1双人，2四人
	matchScene: 匹配场景，1为大厅，2为素质广场
	result: 1为成功，0为失败，2为取消
	TeamType：1系统分配， 2 好友拉入
*/

// tlogMatchFlow tlog战场匹配表
func (user *LobbyUser) tlogMatchFlow(matchType, matchScene uint32, matchTime int64, result, realNum, skybox, stateEnterTeam uint32) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg != nil {
		msg := &protoMsg.MatchFlow{}
		msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
		msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
		msg.VGameAppID = loginMsg.VGameAppID
		if msg.VGameAppID == "" {
			msg.VGameAppID = "NoVGameAppID"
		}
		msg.PlatID = loginMsg.PlatID
		msg.IZoneAreaID = 0
		msg.VOpenID = loginMsg.VOpenID
		if msg.VOpenID == "NoOpenID" || msg.VOpenID == "" {
			msg.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
		}
		msg.VRoleName = user.GetName()
		if msg.VRoleName == "" {
			msg.VRoleName = "no username"
		} else {
			r := strings.NewReplacer("|", "-", "｜", "+", ",", "=")
			msg.VRoleName = r.Replace(msg.VRoleName)
		}
		msg.VRoleProfession = user.GetRoleModel()
		msg.Level = user.GetLevel()
		msg.PlayerFriendsNum = user.GetFriendsNum()
		msg.ClientVersion = loginMsg.ClientVersion
		if msg.ClientVersion == "" {
			msg.ClientVersion = "1.0.0"
		}
		msg.SystemHardware = loginMsg.SystemHardware
		if msg.SystemHardware == "" {
			msg.SystemHardware = "NoSysHardware"
		}
		msg.TelecomOper = loginMsg.TelecomOper
		if msg.TelecomOper == "" {
			msg.TelecomOper = "NoOp"
		}
		msg.Network = loginMsg.Network
		if msg.Network == "" {
			msg.Network = "NoNetwork"
		} else {
			msg.Network = strings.Replace(msg.Network, "|", "-", -1)
		}
		msg.LoginChannel = loginMsg.LoginChannel
		msg.MatchType = matchType
		msg.MatchScene = matchScene
		msg.MatchTime = matchTime
		msg.Result = result
		msg.RealNum = realNum
		msg.BattleID = user.spaceID
		msg.SkyBox = skybox
		msg.TeamID = user.GetTeamID()
		msg.TeamType = stateEnterTeam
		msg.MatchMode = user.matchMode

		tlog.Format(msg)
	} else {
		user.Warn("tlogMatchFlow loginMsg is nil")
	}
}

/***************************tlog货币流水表********************************/
// addOrReduce 货币变化
const (
	ADD    = 0 //加
	REDUCE = 1 //减
)

// tlogMoneyFlow tlog货币流水表（iMoney货币数量，reason变化原因，addOrReduce增加(0)还是减少(1), iMoneyType货币类型comment.Const中定义）
func (user *LobbyUser) tlogMoneyFlow(iMoney, reason, addOrReduce, iMoneyType uint32, nowCount uint64) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg != nil {
		msg := &protoMsg.MoneyFlow{}
		msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
		msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
		msg.VGameAppID = loginMsg.VGameAppID
		if msg.VGameAppID == "" {
			msg.VGameAppID = "NoVGameAppID"
		}
		msg.PlatID = loginMsg.PlatID
		msg.IZoneAreaID = 0
		msg.VOpenID = loginMsg.VOpenID
		if msg.VOpenID == "NoOpenID" || msg.VOpenID == "" {
			msg.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
		}
		msg.Level = user.GetLevel()
		msg.IMoney = iMoney
		msg.Reason = reason
		msg.AddOrReduce = addOrReduce
		msg.IMoneyType = iMoneyType
		msg.IMoneyCount = nowCount
		msg.MatchMode = user.matchMode

		tlog.Format(msg)
	} else {
		user.Warn("tlogMoneyFlow loginMsg is nil")
	}
}

/********************************tlog社交流水表**************************************/
// snsType 类型
const (
	SNSTYPE_SHOWOFF       = 0 //炫耀
	SNSTYPE_INVITE        = 1 //邀请
	SNSTYPE_SENDHEART     = 2 //送心
	SNSTYPE_RECEIVEHEART  = 3 //收取心
	SNSTYPE_SENDEMAIL     = 4 //发邮件
	SNSTYPE_RECEIVEEMAIL  = 5 //收邮件
	SNSTYPE_SHARE         = 6 //分享
	SNSTYPE_INVITESUCCESS = 7 //邀请成功
)

// tlogSnsFlow tlog社交流水表
func (user *LobbyUser) tlogSnsFlow(count, snsType uint32, acceptOpenID string) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg != nil {
		msg := &protoMsg.SnsFlow{}
		msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
		msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
		msg.VGameAppID = loginMsg.VGameAppID
		if msg.VGameAppID == "" {
			msg.VGameAppID = "NoVGameAppID"
		}
		msg.PlatID = loginMsg.PlatID
		msg.IZoneAreaID = 0
		msg.ActorOpenID = loginMsg.VOpenID
		if msg.ActorOpenID == "NoOpenID" || msg.ActorOpenID == "" {
			msg.ActorOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
		}
		msg.Count = count
		msg.SnsType = snsType
		msg.AcceptOpenID = acceptOpenID
		if msg.AcceptOpenID == "" {
			msg.AcceptOpenID = "NoOpenID"
		}
		msg.PlayerFriendsNum = user.GetFriendsNum()

		tlog.Format(msg)
	} else {
		user.Warn("tlogSnsFlow loginMsg is nil")
	}
}

/********************************tlog新手引导流水表**************************************/

// tlogGuideFlow tlog新手引导流水表
func (user *LobbyUser) tlogGuideFlow(prog uint8) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg != nil {
		msg := &protoMsg.GuideFlow{}
		msg.GameSvrId = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
		msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
		msg.VGameAppid = loginMsg.VGameAppID
		if msg.VGameAppid == "" {
			msg.VGameAppid = "NoVGameAppID"
		}
		msg.PlatID = loginMsg.PlatID
		msg.IZoneAreaID = 0
		msg.Vopenid = loginMsg.VOpenID
		if msg.Vopenid == "NoOpenID" || msg.Vopenid == "" {
			msg.Vopenid, _ = dbservice.Account(user.GetDBID()).GetUsername()
		}
		msg.VRoleName = user.GetName()
		if msg.VRoleName == "" {
			msg.VRoleName = "no username"
		} else {
			r := strings.NewReplacer("|", "-", "｜", "+", ",", "=")
			msg.VRoleName = r.Replace(msg.VRoleName)
		}
		msg.VRoleProfession = user.GetRoleModel()
		msg.Level = user.GetLevel()
		msg.PlayerFriendsNum = user.GetFriendsNum()
		msg.ClientVersion = loginMsg.ClientVersion
		if msg.ClientVersion == "" {
			msg.ClientVersion = "1.0.0"
		}
		msg.SystemHardware = loginMsg.SystemHardware
		if msg.SystemHardware == "" {
			msg.SystemHardware = "NoSysHardware"
		}
		msg.TelecomOper = loginMsg.TelecomOper
		if msg.TelecomOper == "" {
			msg.TelecomOper = "NoOp"
		}
		msg.Network = loginMsg.Network
		if msg.Network == "" {
			msg.Network = "NoNetwork"
		} else {
			msg.Network = strings.Replace(msg.Network, "|", "-", -1)
		}
		msg.LoginChannel = loginMsg.LoginChannel
		msg.IGuideID = uint32(prog)
		tlog.Format(msg)
	} else {
		user.Warn("tlogGuideFlow loginMsg is nil")
	}
}

/********************************tlog断线重连流水表**************************************/

// tlogReConnetionFlow tlog断线重连流水表
func (user *LobbyUser) tlogReConnetionFlow() {
	loginMsg := user.GetPlayerLogin()
	if loginMsg != nil {
		msg := &protoMsg.ReConnetionFlow{}
		msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
		msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
		msg.VGameAppID = loginMsg.VGameAppID
		if msg.VGameAppID == "" {
			msg.VGameAppID = "NoVGameAppID"
		}
		msg.PlatID = loginMsg.PlatID
		msg.IZoneAreaID = 0
		msg.VOpenID = loginMsg.VOpenID
		if msg.VOpenID == "NoOpenID" || msg.VOpenID == "" {
			msg.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
		}
		msg.VRoleName = user.GetName()
		if msg.VRoleName == "" {
			msg.VRoleName = "no username"
		} else {
			r := strings.NewReplacer("|", "-", "｜", "+", ",", "=")
			msg.VRoleName = r.Replace(msg.VRoleName)
		}
		msg.VRoleProfession = user.GetRoleModel()
		msg.Level = user.GetLevel()
		msg.PlayerFriendsNum = user.GetFriendsNum()
		msg.ClientVersion = loginMsg.ClientVersion
		if msg.ClientVersion == "" {
			msg.ClientVersion = "1.0.0"
		}
		msg.SystemHardware = loginMsg.SystemHardware
		if msg.SystemHardware == "" {
			msg.SystemHardware = "NoSysHardware"
		}
		msg.TelecomOper = loginMsg.TelecomOper
		if msg.TelecomOper == "" {
			msg.TelecomOper = "NoOp"
		}
		msg.Network = loginMsg.Network
		if msg.Network == "" {
			msg.Network = "NoNetwork"
		} else {
			msg.Network = strings.Replace(msg.Network, "|", "-", -1)
		}
		msg.LoginChannel = loginMsg.LoginChannel

		tlog.Format(msg)
	} else {
		user.Warn("tlogReConnetionFlow loginMsg is nil")
	}
}

/********************************tlog道具流水表**************************************/

/*
	iMoneyType: 钱的类型, 0金币，1钻石, 2勇气值, 3 无,奖励发放
	addOrReduce: 道具增加 0, 减少 1
*/

// iGoodsType 道具类型
const (
	iGoodsType_Role     = 1 //角色
	iGoodsType_Umbrella = 2 //伞包
)

// tlogItemFlow tlog道具流水表
func (user *LobbyUser) tlogItemFlow(iGoodsId, count, reason, iMoney, iMoneyType, addOrReduce, reduceReason, leftUseTime uint32) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg != nil {
		msg := &protoMsg.ItemFlow{}
		msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
		msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
		msg.VGameAppID = loginMsg.VGameAppID
		if msg.VGameAppID == "" {
			msg.VGameAppID = "NoVGameAppID"
		}
		msg.PlatID = loginMsg.PlatID
		msg.IZoneAreaID = 0
		msg.VOpenID = loginMsg.VOpenID
		if msg.VOpenID == "NoOpenID" || msg.VOpenID == "" {
			msg.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
		}
		msg.Level = user.GetLevel()

		goodsConfig, _ := excel.GetStore(uint64(iGoodsId))
		msg.IGoodsType = uint32(goodsConfig.Type) // 商品类型

		msg.IGoodsId = iGoodsId
		msg.Count = count

		// 获取数据库信息
		info, err := db.PlayerGoodsUtil(user.GetDBID()).GetGoodsInfo(uint32(goodsConfig.RelationID))
		if err == nil && info != nil {
			msg.AfterCount = info.Sum
		}

		msg.Reason = reason
		msg.SubReason = 0
		msg.IMoney = iMoney
		msg.IMoneyType = iMoneyType
		msg.AddOrReduce = addOrReduce
		msg.ReduceReason = reduceReason
		msg.LeftUseTime = leftUseTime

		tlog.Format(msg)
	} else {
		user.Warn("tlogItemFlow loginMsg is nil")
	}
}

/********************************tlog物品存量流水表**************************************/

// behaveType 状态
const (
	login  = 1
	logout = 2
)

// tlogGoodRecordFlow tlog物品存量流水表
func (user *LobbyUser) tlogGoodRecordFlow(behaveType uint32) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg != nil {
		msg := &protoMsg.GoodRecordFlow{}
		msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
		msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
		msg.VGameAppID = loginMsg.VGameAppID
		if msg.VGameAppID == "" {
			msg.VGameAppID = "NoVGameAppID"
		}
		msg.PlatID = loginMsg.PlatID
		msg.IZoneAreaID = 0
		msg.VOpenID = loginMsg.VOpenID
		if msg.VOpenID == "NoOpenID" || msg.VOpenID == "" {
			msg.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
		}
		msg.VRoleName = user.GetName()
		if msg.VRoleName == "" {
			msg.VRoleName = "no username"
		} else {
			r := strings.NewReplacer("|", "-", "｜", "+", ",", "=")
			msg.VRoleName = r.Replace(msg.VRoleName)
		}
		msg.VRoleProfession = user.GetRoleModel()
		msg.Level = user.GetLevel()
		msg.PlayerFriendsNum = user.GetFriendsNum()
		msg.ClientVersion = loginMsg.ClientVersion
		if msg.ClientVersion == "" {
			msg.ClientVersion = "1.0.0"
		}
		msg.SystemHardware = loginMsg.SystemHardware
		if msg.SystemHardware == "" {
			msg.SystemHardware = "NoSysHardware"
		}
		msg.TelecomOper = loginMsg.TelecomOper
		if msg.TelecomOper == "" {
			msg.TelecomOper = "NoOp"
		}
		msg.Network = loginMsg.Network
		if msg.Network == "" {
			msg.Network = "NoNetwork"
		} else {
			msg.Network = strings.Replace(msg.Network, "|", "-", -1)
		}
		msg.LoginChannel = loginMsg.LoginChannel
		msg.BehaveType = behaveType
		msg.RoleStock, msg.BrollyStock = user.getGoodsStock()

		tlog.Format(msg)
	} else {
		user.Warn("tlogGoodRecordFlow loginMsg is nil")
	}
}

// getGoodsStock 获取物品库存 [[1656,2],[1657,3],[1658,2]]
func (user *LobbyUser) getGoodsStock() (string, string) {
	roleStock, brollyStock := "[", "["

	goodsInfo := db.PlayerGoodsUtil(user.GetDBID()).GetAllGoodsInfo()
	for _, v := range goodsInfo {
		if v == nil {
			continue
		}

		storeData, ok := excel.GetStore(uint64(v.Id))
		if !ok {
			continue
		}

		if storeData.Type == 1 {
			roleStock += "[" + strconv.FormatUint(uint64(v.Id), 10) + "," + strconv.FormatUint(uint64(v.Sum), 10) + "],"
		} else if storeData.Type == 2 {
			brollyStock += "[" + strconv.FormatUint(uint64(v.Id), 10) + "," + strconv.FormatUint(uint64(v.Sum), 10) + "],"
		}
	}

	roleStock = strings.TrimSuffix(roleStock, ",") + "]"
	brollyStock = strings.TrimSuffix(brollyStock, ",") + "]"

	return roleStock, brollyStock
}

//InsigniaFlow 勋章系统
func (user *LobbyUser) InsigniaFlow(id uint32) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg == nil {
		user.Warn("PlayerLogin msg null:tlogBattleResult")
		return
	}

	msg := &protoMsg.InsigniaFlow{}
	msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
	msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
	msg.VGameAppID = loginMsg.VGameAppID
	if msg.VGameAppID == "" {
		msg.VGameAppID = "NoVGameAppID"
	}
	msg.PlatID = loginMsg.PlatID
	msg.IZoneAreaID = 0
	msg.VOpenID = loginMsg.VOpenID
	if msg.VOpenID == "NoOpenID" || msg.VOpenID == "" {
		msg.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
	}
	msg.VRoleName = user.GetName()

	msg.VRoleProfession = user.GetRoleModel()
	msg.Level = user.GetLevel()
	msg.PlayerFriendsNum = user.GetFriendsNum()
	msg.ClientVersion = loginMsg.ClientVersion
	if msg.ClientVersion == "" {
		msg.ClientVersion = "1.0.0"
	}
	msg.SystemHardware = loginMsg.SystemHardware
	if msg.SystemHardware == "" {
		msg.SystemHardware = "NoSysHardware"
	}
	msg.TelecomOper = loginMsg.TelecomOper
	if msg.TelecomOper == "" {
		msg.TelecomOper = "NoOp"
	}
	msg.Network = loginMsg.Network
	if msg.Network == "" {
		msg.Network = "NoNetwork"
	} else {
		msg.Network = strings.Replace(msg.Network, "|", "-", -1)
	}
	msg.LoginChannel = loginMsg.LoginChannel
	msg.LogType = 2
	msg.InsigniaUse = id
	tlog.Format(msg)
}

//AchievementFlow 勋章系统
func (user *LobbyUser) AchievementFlow(id uint32, level uint32, exp uint32) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg == nil {
		user.Warn("PlayerLogin msg null:tlogBattleResult")
		return
	}

	msg := &protoMsg.AchievementFlow{}
	msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
	msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
	msg.VGameAppID = loginMsg.VGameAppID
	if msg.VGameAppID == "" {
		msg.VGameAppID = "NoVGameAppID"
	}
	msg.PlatID = loginMsg.PlatID
	msg.IZoneAreaID = 0
	msg.VOpenID = loginMsg.VOpenID
	if msg.VOpenID == "NoOpenID" || msg.VOpenID == "" {
		msg.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
	}
	msg.VRoleName = user.GetName()

	msg.VRoleProfession = user.GetRoleModel()
	msg.Level = user.GetLevel()
	msg.PlayerFriendsNum = user.GetFriendsNum()
	msg.ClientVersion = loginMsg.ClientVersion
	if msg.ClientVersion == "" {
		msg.ClientVersion = "1.0.0"
	}
	msg.SystemHardware = loginMsg.SystemHardware
	if msg.SystemHardware == "" {
		msg.SystemHardware = "NoSysHardware"
	}
	msg.TelecomOper = loginMsg.TelecomOper
	if msg.TelecomOper == "" {
		msg.TelecomOper = "NoOp"
	}
	msg.Network = loginMsg.Network
	if msg.Network == "" {
		msg.Network = "NoNetwork"
	} else {
		msg.Network = strings.Replace(msg.Network, "|", "-", -1)
	}
	msg.LoginChannel = loginMsg.LoginChannel

	msg.AchievementID = id
	msg.AchievementLevel = level
	msg.AchievementExp = exp

	tlog.Format(msg)
}

//幻化装统
func (user *LobbyUser) WearInGameFlow(id uint32, kind uint32) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg == nil {
		user.Warn("PlayerLogin msg null:tlogBattleResult")
		return
	}

	msg := &protoMsg.WearInGameFlow{}
	msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
	msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
	msg.VGameAppID = loginMsg.VGameAppID
	if msg.VGameAppID == "" {
		msg.VGameAppID = "NoVGameAppID"
	}
	msg.PlatID = loginMsg.PlatID
	msg.IZoneAreaID = 0
	msg.VOpenID = loginMsg.VOpenID
	if msg.VOpenID == "NoOpenID" || msg.VOpenID == "" {
		msg.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
	}
	msg.VRoleName = user.GetName()

	msg.VRoleProfession = user.GetRoleModel()
	msg.Level = user.GetLevel()
	msg.PlayerFriendsNum = user.GetFriendsNum()
	msg.ClientVersion = loginMsg.ClientVersion
	if msg.ClientVersion == "" {
		msg.ClientVersion = "1.0.0"
	}
	msg.SystemHardware = loginMsg.SystemHardware
	if msg.SystemHardware == "" {
		msg.SystemHardware = "NoSysHardware"
	}
	msg.TelecomOper = loginMsg.TelecomOper
	if msg.TelecomOper == "" {
		msg.TelecomOper = "NoOp"
	}
	msg.Network = loginMsg.Network
	if msg.Network == "" {
		msg.Network = "NoNetwork"
	} else {
		msg.Network = strings.Replace(msg.Network, "|", "-", -1)
	}
	msg.LoginChannel = loginMsg.LoginChannel

	msg.EquipID = id
	msg.ActionType = kind

	tlog.Format(msg)
}

// tlogTreasureBoxFlow 开宝箱
func (user *LobbyUser) tlogTreasureBoxFlow(boxId, openNum, iMoneyType, iMoney, itemId uint32) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg == nil {
		user.Warn("tlogTreasureBoxFlow loginMsg null!")
		return
	}

	msg := &protoMsg.TreasureBoxFlow{}
	msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
	msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
	msg.VGameAppID = loginMsg.VGameAppID
	if msg.VGameAppID == "" {
		msg.VGameAppID = "NoVGameAppID"
	}
	msg.PlatID = loginMsg.PlatID
	msg.IZoneAreaID = 0
	msg.VOpenID = loginMsg.VOpenID
	if msg.VOpenID == "NoOpenID" || msg.VOpenID == "" {
		msg.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
	}
	msg.VRoleName = user.GetName()

	msg.VRoleProfession = user.GetRoleModel()
	msg.Level = user.GetLevel()
	msg.PlayerFriendsNum = user.GetFriendsNum()
	msg.ClientVersion = loginMsg.ClientVersion
	if msg.ClientVersion == "" {
		msg.ClientVersion = "1.0.0"
	}
	msg.SystemHardware = loginMsg.SystemHardware
	if msg.SystemHardware == "" {
		msg.SystemHardware = "NoSysHardware"
	}
	msg.TelecomOper = loginMsg.TelecomOper
	if msg.TelecomOper == "" {
		msg.TelecomOper = "NoOp"
	}
	msg.Network = loginMsg.Network
	if msg.Network == "" {
		msg.Network = "NoNetwork"
	} else {
		msg.Network = strings.Replace(msg.Network, "|", "-", -1)
	}
	msg.LoginChannel = loginMsg.LoginChannel

	msg.BoxID = boxId
	msg.OpenNum = openNum
	msg.IMoneyType = iMoneyType
	msg.IMoney = iMoney
	msg.ItemID = itemId

	tlog.Format(msg)
}

//MilitaryRankFlow 军衔
func (user *LobbyUser) MilitaryRankFlow() {
	loginMsg := user.GetPlayerLogin()
	if loginMsg == nil {
		user.Warn("PlayerLogin msg null:tlogBattleResult")
		return
	}

	msg := &protoMsg.MilitaryRankFlow{}
	msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
	msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
	msg.VGameAppID = loginMsg.VGameAppID
	if msg.VGameAppID == "" {
		msg.VGameAppID = "NoVGameAppID"
	}
	msg.PlatID = loginMsg.PlatID
	msg.IZoneAreaID = 0
	msg.VOpenID = loginMsg.VOpenID
	if msg.VOpenID == "NoOpenID" || msg.VOpenID == "" {
		msg.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
	}
	msg.VRoleName = user.GetName()

	msg.VRoleProfession = user.GetRoleModel()
	msg.Level = user.GetLevel()
	msg.PlayerFriendsNum = user.GetFriendsNum()
	msg.ClientVersion = loginMsg.ClientVersion
	if msg.ClientVersion == "" {
		msg.ClientVersion = "1.0.0"
	}
	msg.SystemHardware = loginMsg.SystemHardware
	if msg.SystemHardware == "" {
		msg.SystemHardware = "NoSysHardware"
	}
	msg.TelecomOper = loginMsg.TelecomOper
	if msg.TelecomOper == "" {
		msg.TelecomOper = "NoOp"
	}
	msg.Network = loginMsg.Network
	if msg.Network == "" {
		msg.Network = "NoNetwork"
	} else {
		msg.Network = strings.Replace(msg.Network, "|", "-", -1)
	}
	msg.LoginChannel = loginMsg.LoginChannel

	msg.Exp = user.GetExp()
	msg.MilitaryRank = common.GetMilitaryRankByLevel(user.GetLevel())

	tlog.Format(msg)
}

//DayTaskFlow 每日任务
func (user *LobbyUser) DayTaskFlow(id, finished uint32) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg == nil {
		user.Warn("PlayerLogin msg null:tlogBattleResult")
		return
	}

	msg := &protoMsg.DayTaskFlow{}
	msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
	msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
	msg.VGameAppID = loginMsg.VGameAppID
	if msg.VGameAppID == "" {
		msg.VGameAppID = "NoVGameAppID"
	}
	msg.PlatID = loginMsg.PlatID
	msg.IZoneAreaID = 0
	msg.VOpenID = loginMsg.VOpenID
	if msg.VOpenID == "NoOpenID" || msg.VOpenID == "" {
		msg.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
	}
	msg.VRoleName = user.GetName()

	msg.VRoleProfession = user.GetRoleModel()
	msg.Level = user.GetLevel()
	msg.PlayerFriendsNum = user.GetFriendsNum()
	msg.ClientVersion = loginMsg.ClientVersion
	if msg.ClientVersion == "" {
		msg.ClientVersion = "1.0.0"
	}
	msg.SystemHardware = loginMsg.SystemHardware
	if msg.SystemHardware == "" {
		msg.SystemHardware = "NoSysHardware"
	}
	msg.TelecomOper = loginMsg.TelecomOper
	if msg.TelecomOper == "" {
		msg.TelecomOper = "NoOp"
	}
	msg.Network = loginMsg.Network
	if msg.Network == "" {
		msg.Network = "NoNetwork"
	} else {
		msg.Network = strings.Replace(msg.Network, "|", "-", -1)
	}
	msg.LoginChannel = loginMsg.LoginChannel

	msg.TaskItemID = id
	msg.TaskItemFinished = finished

	tlog.Format(msg)
}

//ActivenessFlow 活跃度表
func (user *LobbyUser) ActivenessFlow(day, week uint32) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg == nil {
		user.Warn("PlayerLogin msg null:tlogBattleResult")
		return
	}

	msg := &protoMsg.ActivenessFlow{}
	msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
	msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
	msg.VGameAppID = loginMsg.VGameAppID
	if msg.VGameAppID == "" {
		msg.VGameAppID = "NoVGameAppID"
	}
	msg.PlatID = loginMsg.PlatID
	msg.IZoneAreaID = 0
	msg.VOpenID = loginMsg.VOpenID
	if msg.VOpenID == "NoOpenID" || msg.VOpenID == "" {
		msg.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
	}
	msg.VRoleName = user.GetName()

	msg.VRoleProfession = user.GetRoleModel()
	msg.Level = user.GetLevel()
	msg.PlayerFriendsNum = user.GetFriendsNum()
	msg.ClientVersion = loginMsg.ClientVersion
	if msg.ClientVersion == "" {
		msg.ClientVersion = "1.0.0"
	}
	msg.SystemHardware = loginMsg.SystemHardware
	if msg.SystemHardware == "" {
		msg.SystemHardware = "NoSysHardware"
	}
	msg.TelecomOper = loginMsg.TelecomOper
	if msg.TelecomOper == "" {
		msg.TelecomOper = "NoOp"
	}
	msg.Network = loginMsg.Network
	if msg.Network == "" {
		msg.Network = "NoNetwork"
	} else {
		msg.Network = strings.Replace(msg.Network, "|", "-", -1)
	}
	msg.LoginChannel = loginMsg.LoginChannel

	msg.DayActiveness = day
	msg.WeekActiveness = week

	tlog.Format(msg)
}

//ComradeTaskFlow 战友任务
func (user *LobbyUser) ComradeTaskFlow(id, finished uint32) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg == nil {
		user.Warn("PlayerLogin msg null:tlogBattleResult")
		return
	}

	msg := &protoMsg.DayTaskFlow{}
	msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
	msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
	msg.VGameAppID = loginMsg.VGameAppID
	if msg.VGameAppID == "" {
		msg.VGameAppID = "NoVGameAppID"
	}
	msg.PlatID = loginMsg.PlatID
	msg.IZoneAreaID = 0
	msg.VOpenID = loginMsg.VOpenID
	if msg.VOpenID == "NoOpenID" || msg.VOpenID == "" {
		msg.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
	}
	msg.VRoleName = user.GetName()

	msg.VRoleProfession = user.GetRoleModel()
	msg.Level = user.GetLevel()
	msg.PlayerFriendsNum = user.GetFriendsNum()
	msg.ClientVersion = loginMsg.ClientVersion
	if msg.ClientVersion == "" {
		msg.ClientVersion = "1.0.0"
	}
	msg.SystemHardware = loginMsg.SystemHardware
	if msg.SystemHardware == "" {
		msg.SystemHardware = "NoSysHardware"
	}
	msg.TelecomOper = loginMsg.TelecomOper
	if msg.TelecomOper == "" {
		msg.TelecomOper = "NoOp"
	}
	msg.Network = loginMsg.Network
	if msg.Network == "" {
		msg.Network = "NoNetwork"
	} else {
		msg.Network = strings.Replace(msg.Network, "|", "-", -1)
	}
	msg.LoginChannel = loginMsg.LoginChannel

	msg.TaskItemID = id
	msg.TaskItemFinished = finished

	tlog.Format(msg)
}

//TaskFlow 任务
func (user *LobbyUser) TaskFlow(name string, taskType uint8, id, finished uint32) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg == nil {
		user.Warn("PlayerLogin msg null:tlogBattleResult")
		return
	}

	msg := &protoMsg.TaskFlow{}
	msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
	msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
	msg.VGameAppID = loginMsg.VGameAppID
	if msg.VGameAppID == "" {
		msg.VGameAppID = "NoVGameAppID"
	}
	msg.PlatID = loginMsg.PlatID
	msg.IZoneAreaID = 0
	msg.VOpenID = loginMsg.VOpenID
	if msg.VOpenID == "NoOpenID" || msg.VOpenID == "" {
		msg.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
	}
	msg.VRoleName = user.GetName()

	msg.VRoleProfession = user.GetRoleModel()
	msg.Level = user.GetLevel()
	msg.PlayerFriendsNum = user.GetFriendsNum()
	msg.ClientVersion = loginMsg.ClientVersion
	if msg.ClientVersion == "" {
		msg.ClientVersion = "1.0.0"
	}
	msg.SystemHardware = loginMsg.SystemHardware
	if msg.SystemHardware == "" {
		msg.SystemHardware = "NoSysHardware"
	}
	msg.TelecomOper = loginMsg.TelecomOper
	if msg.TelecomOper == "" {
		msg.TelecomOper = "NoOp"
	}
	msg.Network = loginMsg.Network
	if msg.Network == "" {
		msg.Network = "NoNetwork"
	} else {
		msg.Network = strings.Replace(msg.Network, "|", "-", -1)
	}
	msg.LoginChannel = loginMsg.LoginChannel

	msg.TaskName = name
	msg.TaskItemID = id
	msg.TaskType = uint32(taskType)
	msg.TaskItemFinished = finished

	tlog.Format(msg)
}

//SpecialExpFlow 特训经验表
func (user *LobbyUser) SpecialExpFlow(exp uint32) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg == nil {
		user.Warn("PlayerLogin msg null:tlogBattleResult")
		return
	}

	msg := &protoMsg.SpecialExpFlow{}
	msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
	msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
	msg.VGameAppID = loginMsg.VGameAppID
	if msg.VGameAppID == "" {
		msg.VGameAppID = "NoVGameAppID"
	}
	msg.PlatID = loginMsg.PlatID
	msg.IZoneAreaID = 0
	msg.VOpenID = loginMsg.VOpenID
	if msg.VOpenID == "NoOpenID" || msg.VOpenID == "" {
		msg.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
	}
	msg.VRoleName = user.GetName()

	msg.VRoleProfession = user.GetRoleModel()
	msg.Level = user.GetLevel()
	msg.PlayerFriendsNum = user.GetFriendsNum()
	msg.ClientVersion = loginMsg.ClientVersion
	if msg.ClientVersion == "" {
		msg.ClientVersion = "1.0.0"
	}
	msg.SystemHardware = loginMsg.SystemHardware
	if msg.SystemHardware == "" {
		msg.SystemHardware = "NoSysHardware"
	}
	msg.TelecomOper = loginMsg.TelecomOper
	if msg.TelecomOper == "" {
		msg.TelecomOper = "NoOp"
	}
	msg.Network = loginMsg.Network
	if msg.Network == "" {
		msg.Network = "NoNetwork"
	} else {
		msg.Network = strings.Replace(msg.Network, "|", "-", -1)
	}
	msg.LoginChannel = loginMsg.LoginChannel

	msg.SpecialExp = exp

	tlog.Format(msg)
}

//SeasonExpFlow 赛季经验表
func (user *LobbyUser) SeasonExpFlow(exp uint32) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg == nil {
		user.Warn("PlayerLogin msg null:tlogBattleResult")
		return
	}

	msg := &protoMsg.SeasonExpFlow{}
	msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
	msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
	msg.VGameAppID = loginMsg.VGameAppID
	if msg.VGameAppID == "" {
		msg.VGameAppID = "NoVGameAppID"
	}
	msg.PlatID = loginMsg.PlatID
	msg.IZoneAreaID = 0
	msg.VOpenID = loginMsg.VOpenID
	if msg.VOpenID == "NoOpenID" || msg.VOpenID == "" {
		msg.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
	}
	msg.VRoleName = user.GetName()

	msg.VRoleProfession = user.GetRoleModel()
	msg.Level = user.GetLevel()
	msg.PlayerFriendsNum = user.GetFriendsNum()
	msg.ClientVersion = loginMsg.ClientVersion
	if msg.ClientVersion == "" {
		msg.ClientVersion = "1.0.0"
	}
	msg.SystemHardware = loginMsg.SystemHardware
	if msg.SystemHardware == "" {
		msg.SystemHardware = "NoSysHardware"
	}
	msg.TelecomOper = loginMsg.TelecomOper
	if msg.TelecomOper == "" {
		msg.TelecomOper = "NoOp"
	}
	msg.Network = loginMsg.Network
	if msg.Network == "" {
		msg.Network = "NoNetwork"
	} else {
		msg.Network = strings.Replace(msg.Network, "|", "-", -1)
	}
	msg.LoginChannel = loginMsg.LoginChannel

	msg.SeasonExp = exp

	tlog.Format(msg)
}

// tlogPreferenceFlow 偏好流水表
func (user *LobbyUser) tlogPreferenceFlow(itemType, preType uint32) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg == nil {
		user.Warn("PlayerLogin msg null:tlogPreferenceFlow")
		return
	}

	msg := &protoMsg.PreferenceFlow{}
	msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
	msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
	msg.VGameAppID = loginMsg.VGameAppID
	if msg.VGameAppID == "" {
		msg.VGameAppID = "NoVGameAppID"
	}
	msg.PlatID = loginMsg.PlatID
	msg.IZoneAreaID = 0
	msg.VOpenID = loginMsg.VOpenID
	if msg.VOpenID == "NoOpenID" || msg.VOpenID == "" {
		msg.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
	}
	msg.VRoleName = user.GetName()

	msg.VRoleProfession = user.GetRoleModel()
	msg.Level = user.GetLevel()
	msg.PlayerFriendsNum = user.GetFriendsNum()
	msg.ClientVersion = loginMsg.ClientVersion
	if msg.ClientVersion == "" {
		msg.ClientVersion = "1.0.0"
	}
	msg.SystemHardware = loginMsg.SystemHardware
	if msg.SystemHardware == "" {
		msg.SystemHardware = "NoSysHardware"
	}
	msg.TelecomOper = loginMsg.TelecomOper
	if msg.TelecomOper == "" {
		msg.TelecomOper = "NoOp"
	}
	msg.Network = loginMsg.Network
	if msg.Network == "" {
		msg.Network = "NoNetwork"
	} else {
		msg.Network = strings.Replace(msg.Network, "|", "-", -1)
	}
	msg.LoginChannel = loginMsg.LoginChannel

	msg.ItemType = itemType
	msg.PreType = preType

	tlog.Format(msg)
}

// tlogFestivalFlow 节日活动表
func (user *LobbyUser) tlogFestivalFlow(taskID uint32, rewards map[uint32]uint32) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg == nil {
		user.Warn("PlayerLogin msg null:tlogFestivalFlow")
		return
	}

	msg := &protoMsg.FestivalFlow{}
	msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
	msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
	msg.VGameAppID = loginMsg.VGameAppID
	if msg.VGameAppID == "" {
		msg.VGameAppID = "NoVGameAppID"
	}
	msg.PlatID = loginMsg.PlatID
	msg.IZoneAreaID = 0
	msg.VOpenID = loginMsg.VOpenID
	if msg.VOpenID == "NoOpenID" || msg.VOpenID == "" {
		msg.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
	}
	msg.VRoleName = user.GetName()

	msg.VRoleProfession = user.GetRoleModel()
	msg.Level = user.GetLevel()
	msg.PlayerFriendsNum = user.GetFriendsNum()
	msg.ClientVersion = loginMsg.ClientVersion
	if msg.ClientVersion == "" {
		msg.ClientVersion = "1.0.0"
	}
	msg.SystemHardware = loginMsg.SystemHardware
	if msg.SystemHardware == "" {
		msg.SystemHardware = "NoSysHardware"
	}
	msg.TelecomOper = loginMsg.TelecomOper
	if msg.TelecomOper == "" {
		msg.TelecomOper = "NoOp"
	}
	msg.Network = loginMsg.Network
	if msg.Network == "" {
		msg.Network = "NoNetwork"
	} else {
		msg.Network = strings.Replace(msg.Network, "|", "-", -1)
	}
	msg.LoginChannel = loginMsg.LoginChannel

	msg.TaskID = taskID
	msg.Rewards = common.MapToString(rewards)

	tlog.Format(msg)
}

// tlogExchangeFlow 兑换活动表
func (user *LobbyUser) tlogExchangeFlow(actID, exchangeID, exchangeNum uint32, rewards map[uint32]uint32) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg == nil {
		user.Warn("PlayerLogin msg null:tlogExchangeFlow")
		return
	}

	msg := &protoMsg.ExchangeFlow{}
	msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
	msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
	msg.VGameAppID = loginMsg.VGameAppID
	if msg.VGameAppID == "" {
		msg.VGameAppID = "NoVGameAppID"
	}
	msg.PlatID = loginMsg.PlatID
	msg.IZoneAreaID = 0
	msg.VOpenID = loginMsg.VOpenID
	if msg.VOpenID == "NoOpenID" || msg.VOpenID == "" {
		msg.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
	}
	msg.VRoleName = user.GetName()

	msg.VRoleProfession = user.GetRoleModel()
	msg.Level = user.GetLevel()
	msg.PlayerFriendsNum = user.GetFriendsNum()
	msg.ClientVersion = loginMsg.ClientVersion
	if msg.ClientVersion == "" {
		msg.ClientVersion = "1.0.0"
	}
	msg.SystemHardware = loginMsg.SystemHardware
	if msg.SystemHardware == "" {
		msg.SystemHardware = "NoSysHardware"
	}
	msg.TelecomOper = loginMsg.TelecomOper
	if msg.TelecomOper == "" {
		msg.TelecomOper = "NoOp"
	}
	msg.Network = loginMsg.Network
	if msg.Network == "" {
		msg.Network = "NoNetwork"
	} else {
		msg.Network = strings.Replace(msg.Network, "|", "-", -1)
	}
	msg.LoginChannel = loginMsg.LoginChannel

	msg.ActID = actID
	msg.ExchangeID = exchangeID
	msg.ExchangeNum = exchangeNum
	msg.Rewards = common.MapToString(rewards)

	tlog.Format(msg)
}

// tlogBallStarFlow 一球成名活动表
func (user *LobbyUser) tlogBallStarFlow(position, sum, rewardsID, rewardsNum uint32) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg == nil {
		user.Warn("PlayerLogin msg null:tlogBallStarFlow")
		return
	}

	msg := &protoMsg.BallStarFlow{}
	msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
	msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
	msg.VGameAppID = loginMsg.VGameAppID
	if msg.VGameAppID == "" {
		msg.VGameAppID = "NoVGameAppID"
	}
	msg.PlatID = loginMsg.PlatID
	msg.IZoneAreaID = 0
	msg.VOpenID = loginMsg.VOpenID
	if msg.VOpenID == "NoOpenID" || msg.VOpenID == "" {
		msg.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
	}
	msg.VRoleName = user.GetName()

	msg.VRoleProfession = user.GetRoleModel()
	msg.Level = user.GetLevel()
	msg.PlayerFriendsNum = user.GetFriendsNum()
	msg.ClientVersion = loginMsg.ClientVersion
	if msg.ClientVersion == "" {
		msg.ClientVersion = "1.0.0"
	}
	msg.SystemHardware = loginMsg.SystemHardware
	if msg.SystemHardware == "" {
		msg.SystemHardware = "NoSysHardware"
	}
	msg.TelecomOper = loginMsg.TelecomOper
	if msg.TelecomOper == "" {
		msg.TelecomOper = "NoOp"
	}
	msg.Network = loginMsg.Network
	if msg.Network == "" {
		msg.Network = "NoNetwork"
	} else {
		msg.Network = strings.Replace(msg.Network, "|", "-", -1)
	}
	msg.LoginChannel = loginMsg.LoginChannel

	msg.Position = position
	msg.Sum = sum
	msg.RewardsID = rewardsID
	msg.RewardsNum = rewardsNum

	tlog.Format(msg)
}

// tlogWorldCupChampionFlow 冠军竞猜投注日志
func (user *LobbyUser) tlogWorldCupChampionFlow(teamId uint32, odds float32) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg == nil {
		user.Warn("PlayerLogin msg null:tlogBallStarFlow")
		return
	}

	msg := &protoMsg.WorldCupChampionFlow{}
	msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
	msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
	msg.VGameAppID = loginMsg.VGameAppID
	if msg.VGameAppID == "" {
		msg.VGameAppID = "NoVGameAppID"
	}
	msg.PlatID = loginMsg.PlatID
	msg.IZoneAreaID = 0
	msg.VOpenID = loginMsg.VOpenID
	if msg.VOpenID == "NoOpenID" || msg.VOpenID == "" {
		msg.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
	}
	msg.VRoleName = user.GetName()

	msg.VRoleProfession = user.GetRoleModel()
	msg.Level = user.GetLevel()
	msg.PlayerFriendsNum = user.GetFriendsNum()
	msg.ClientVersion = loginMsg.ClientVersion
	if msg.ClientVersion == "" {
		msg.ClientVersion = "1.0.0"
	}
	msg.SystemHardware = loginMsg.SystemHardware
	if msg.SystemHardware == "" {
		msg.SystemHardware = "NoSysHardware"
	}
	msg.TelecomOper = loginMsg.TelecomOper
	if msg.TelecomOper == "" {
		msg.TelecomOper = "NoOp"
	}
	msg.Network = loginMsg.Network
	if msg.Network == "" {
		msg.Network = "NoNetwork"
	} else {
		msg.Network = strings.Replace(msg.Network, "|", "-", -1)
	}
	msg.LoginChannel = loginMsg.LoginChannel

	msg.TeamID = teamId
	msg.Odds = fmt.Sprintf("%.4f", odds)
	tlog.Format(msg)
}

// tlogWorldCupMatchFlow 胜负竞猜投注日志
func (user *LobbyUser) tlogWorldCupMatchFlow(matchId, kind uint32, odds float32) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg == nil {
		user.Warn("PlayerLogin msg null:tlogBallStarFlow")
		return
	}

	msg := &protoMsg.WorldCupMatchFlow{}
	msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
	msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
	msg.VGameAppID = loginMsg.VGameAppID
	if msg.VGameAppID == "" {
		msg.VGameAppID = "NoVGameAppID"
	}
	msg.PlatID = loginMsg.PlatID
	msg.IZoneAreaID = 0
	msg.VOpenID = loginMsg.VOpenID
	if msg.VOpenID == "NoOpenID" || msg.VOpenID == "" {
		msg.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
	}
	msg.VRoleName = user.GetName()

	msg.VRoleProfession = user.GetRoleModel()
	msg.Level = user.GetLevel()
	msg.PlayerFriendsNum = user.GetFriendsNum()
	msg.ClientVersion = loginMsg.ClientVersion
	if msg.ClientVersion == "" {
		msg.ClientVersion = "1.0.0"
	}
	msg.SystemHardware = loginMsg.SystemHardware
	if msg.SystemHardware == "" {
		msg.SystemHardware = "NoSysHardware"
	}
	msg.TelecomOper = loginMsg.TelecomOper
	if msg.TelecomOper == "" {
		msg.TelecomOper = "NoOp"
	}
	msg.Network = loginMsg.Network
	if msg.Network == "" {
		msg.Network = "NoNetwork"
	} else {
		msg.Network = strings.Replace(msg.Network, "|", "-", -1)
	}
	msg.LoginChannel = loginMsg.LoginChannel

	msg.MatchId = matchId
	msg.Kind = kind
	msg.Odds = fmt.Sprintf("%.4f", odds)
	tlog.Format(msg)
}

//BattleBookingFlow 预约好友表
func (user *LobbyUser) BattleBookingFlow(booker, target uint64, resp uint32) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg == nil {
		user.Warn("PlayerLogin msg null:tlogBattleResult")
		return
	}

	msg := &protoMsg.BattleBookingFlow{}
	msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
	msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
	msg.VGameAppID = loginMsg.VGameAppID
	if msg.VGameAppID == "" {
		msg.VGameAppID = "NoVGameAppID"
	}
	msg.PlatID = loginMsg.PlatID
	msg.IZoneAreaID = 0
	msg.VOpenID = loginMsg.VOpenID
	if msg.VOpenID == "NoOpenID" || msg.VOpenID == "" {
		msg.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
	}
	msg.VRoleName = user.GetName()

	msg.VRoleProfession = user.GetRoleModel()
	msg.Level = user.GetLevel()
	msg.PlayerFriendsNum = user.GetFriendsNum()
	msg.ClientVersion = loginMsg.ClientVersion
	if msg.ClientVersion == "" {
		msg.ClientVersion = "1.0.0"
	}
	msg.SystemHardware = loginMsg.SystemHardware
	if msg.SystemHardware == "" {
		msg.SystemHardware = "NoSysHardware"
	}
	msg.TelecomOper = loginMsg.TelecomOper
	if msg.TelecomOper == "" {
		msg.TelecomOper = "NoOp"
	}
	msg.Network = loginMsg.Network
	if msg.Network == "" {
		msg.Network = "NoNetwork"
	} else {
		msg.Network = strings.Replace(msg.Network, "|", "-", -1)
	}
	msg.LoginChannel = loginMsg.LoginChannel

	msg.Booker = booker
	msg.BookTarget = target
	msg.Resp = resp

	tlog.Format(msg)
}

const (
	WatchEnter = 1
	WatchLeave = 2
)

func (user *LobbyUser) WatchFlow(op uint32, targetId, battleID uint64, num, watchTime, loading uint32) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg != nil {
		msg := &protoMsg.WatchFlow{}
		msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
		msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
		msg.VGameAppID = loginMsg.VGameAppID
		if msg.VGameAppID == "" {
			msg.VGameAppID = "NoVGameAppID"
		}
		msg.PlatID = loginMsg.PlatID
		msg.IZoneAreaID = 0
		msg.VOpenID = loginMsg.VOpenID
		if msg.VOpenID == "NoOpenID" || msg.VOpenID == "" {
			msg.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
		}
		msg.VRoleName = user.GetName()
		if msg.VRoleName == "" {
			msg.VRoleName = "no username"
		} else {
			r := strings.NewReplacer("|", "-", "｜", "+", ",", "=")
			msg.VRoleName = r.Replace(msg.VRoleName)
		}
		msg.VRoleProfession = user.GetRoleModel()
		msg.Level = user.GetLevel()
		msg.PlayerFriendsNum = user.GetFriendsNum()
		msg.ClientVersion = loginMsg.ClientVersion
		if msg.ClientVersion == "" {
			msg.ClientVersion = "1.0.0"
		}
		msg.SystemHardware = loginMsg.SystemHardware
		if msg.SystemHardware == "" {
			msg.SystemHardware = "NoSysHardware"
		}
		msg.TelecomOper = loginMsg.TelecomOper
		if msg.TelecomOper == "" {
			msg.TelecomOper = "NoOp"
		}
		msg.Network = loginMsg.Network
		if msg.Network == "" {
			msg.Network = "NoNetwork"
		} else {
			msg.Network = strings.Replace(msg.Network, "|", "-", -1)
		}
		msg.LoginChannel = loginMsg.LoginChannel

		msg.WatchType = op
		msg.WatchID = user.GetDBID()
		msg.BeWatchID = targetId
		msg.BeWatchName = user.friendMgr.getFriendName(targetId)
		msg.WatchNum = num
		msg.BattleID = battleID
		msg.WatchTime = watchTime
		msg.Loading = loading

		tlog.Format(msg)
	} else {
		user.Warn("tlog WatchFlow loginMsg is nil")
	}
}

// MonthCardFlow 月卡tlog
func (user *LobbyUser) MonthCardFlow(beginTime, endTime int64) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg == nil {
		user.Warn("PlayerLogin msg null:tlogBattleResult")
		return
	}

	msg := &protoMsg.MonthCardFlow{}
	msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
	msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
	msg.VGameAppID = loginMsg.VGameAppID
	if msg.VGameAppID == "" {
		msg.VGameAppID = "NoVGameAppID"
	}
	msg.PlatID = loginMsg.PlatID
	msg.IZoneAreaID = 0
	msg.VOpenID = loginMsg.VOpenID
	if msg.VOpenID == "NoOpenID" || msg.VOpenID == "" {
		msg.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
	}
	msg.VRoleName = user.GetName()

	msg.VRoleProfession = user.GetRoleModel()
	msg.Level = user.GetLevel()
	msg.PlayerFriendsNum = user.GetFriendsNum()
	msg.ClientVersion = loginMsg.ClientVersion
	if msg.ClientVersion == "" {
		msg.ClientVersion = "1.0.0"
	}
	msg.SystemHardware = loginMsg.SystemHardware
	if msg.SystemHardware == "" {
		msg.SystemHardware = "NoSysHardware"
	}
	msg.TelecomOper = loginMsg.TelecomOper
	if msg.TelecomOper == "" {
		msg.TelecomOper = "NoOp"
	}
	msg.Network = loginMsg.Network
	if msg.Network == "" {
		msg.Network = "NoNetwork"
	} else {
		msg.Network = strings.Replace(msg.Network, "|", "-", -1)
	}
	msg.LoginChannel = loginMsg.LoginChannel

	msg.BeginTime = beginTime
	msg.EndTime = endTime

	tlog.Format(msg)
}

// BuyGoodsFlow 购买道具tlog
func (user *LobbyUser) BuyGoodsFlow(billno string, moneyType, amount uint32, succGoods, failGoods map[uint32]uint32, succBuy, succRoll uint32) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg == nil {
		user.Warn("PlayerLogin msg null:tlogBattleResult")
		return
	}

	msg := &protoMsg.BuyGoodsFlow{}
	msg.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
	msg.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
	msg.VGameAppID = loginMsg.VGameAppID
	if msg.VGameAppID == "" {
		msg.VGameAppID = "NoVGameAppID"
	}
	msg.PlatID = loginMsg.PlatID
	msg.IZoneAreaID = 0
	msg.VOpenID = loginMsg.VOpenID
	if msg.VOpenID == "NoOpenID" || msg.VOpenID == "" {
		msg.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
	}
	msg.VRoleName = user.GetName()

	msg.VRoleProfession = user.GetRoleModel()
	msg.Level = user.GetLevel()
	msg.PlayerFriendsNum = user.GetFriendsNum()
	msg.ClientVersion = loginMsg.ClientVersion
	if msg.ClientVersion == "" {
		msg.ClientVersion = "1.0.0"
	}
	msg.SystemHardware = loginMsg.SystemHardware
	if msg.SystemHardware == "" {
		msg.SystemHardware = "NoSysHardware"
	}
	msg.TelecomOper = loginMsg.TelecomOper
	if msg.TelecomOper == "" {
		msg.TelecomOper = "NoOp"
	}
	msg.Network = loginMsg.Network
	if msg.Network == "" {
		msg.Network = "NoNetwork"
	} else {
		msg.Network = strings.Replace(msg.Network, "|", "-", -1)
	}
	msg.LoginChannel = loginMsg.LoginChannel

	msg.BillNo = billno
	msg.IMoneyType = moneyType
	msg.IMoney = amount
	msg.SuccGoods = common.MapUint32ToString(succGoods, ":", ";")
	msg.FailGoods = common.MapUint32ToString(failGoods, ":", ";")
	msg.SuccBuy = succBuy
	msg.SuccRoll = succRoll

	tlog.Format(msg)
}

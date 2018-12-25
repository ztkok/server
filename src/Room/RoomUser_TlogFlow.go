package main

import (
	"common"
	"fmt"
	"protoMsg"
	"strings"
	"time"
	"zeus/dbservice"
	"zeus/linmath"
	"zeus/tlog"

	log "github.com/cihub/seelog"
	"github.com/garyburd/redigo/redis"
)

// tlogRoundFlow tlog本局结束经分流水数据
func (user *RoomUser) tlogRoundFlow() {
	loginMsg := user.GetPlayerLogin()
	if loginMsg == nil {
		log.Warn("PlayerLogin msg null:tlogBattleResult")
		return
	}

	scene := user.GetSpace().(*Scene)
	if scene == nil {
		return
	}
	roundData := &protoMsg.RoundFlow{}

	roundData.GameSvrID = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
	roundData.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
	roundData.VGameAppID = loginMsg.VGameAppID
	if roundData.VGameAppID == "" {
		roundData.VGameAppID = "NoVGameAppID"
	}
	roundData.PlatID = loginMsg.PlatID
	roundData.IZoneAreaID = loginMsg.IZoneAreaID
	roundData.VOpenID = loginMsg.VOpenID
	if roundData.VOpenID == "NoOpenID" || roundData.VOpenID == "" {
		roundData.VOpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
	}
	roundData.BattleID = scene.GetID() //	(必填)关卡id或者副本id[副本ID必须有序]，对局时填0
	roundData.BattleType = user.sumData.battletype
	roundData.RoundScore = 0
	roundData.RoundTime = uint32(user.sumData.endgametime - user.sumData.begingametime)
	if user.sumData.rank == 1 {
		roundData.Result = 1
	} else {
		roundData.Result = 0
	}
	roundData.Rank = user.sumData.rank
	roundData.Gold = user.sumData.coin
	roundData.TotalBattleNum = user.sumData.careerdata.TotalBattleNum
	roundData.FirstNum = user.sumData.careerdata.FirstNum
	roundData.TopTenNum = user.sumData.careerdata.TopTenNum
	if roundData.TotalBattleNum != 0 {
		roundData.FirstRate = fmt.Sprintf("%.4f", float32(roundData.FirstNum)/float32(roundData.TotalBattleNum))
	} else {
		roundData.FirstRate = "0.0000"
	}
	if roundData.TotalBattleNum != 0 {
		roundData.TopTenRate = fmt.Sprintf("%.4f", float32(roundData.TopTenNum)/float32(roundData.TotalBattleNum))
	} else {
		roundData.TopTenRate = "0.0000"
	}
	roundData.SingleMaxKill = user.sumData.careerdata.SingleMaxKill
	roundData.TotalKillNum = user.sumData.careerdata.TotalKillNum
	if roundData.TotalBattleNum != 0 {
		roundData.AverageKillNum = fmt.Sprintf("%.4f", float32(roundData.TotalKillNum)/float32(roundData.TotalBattleNum))
	} else {
		roundData.AverageKillNum = "0.0000"
	}
	roundData.SingleMaxHeadShot = user.sumData.careerdata.SingleMaxHeadShot
	roundData.TotalHeadShot = user.sumData.careerdata.TotalHeadShot
	roundData.TotalEffectHarm = user.sumData.careerdata.TotalEffectHarm
	if roundData.TotalBattleNum != 0 {
		roundData.AverageEffectHarm = fmt.Sprintf("%.4f", float32(roundData.TotalEffectHarm)/float32(roundData.TotalBattleNum))
	} else {
		roundData.AverageEffectHarm = "0.0000"
	}
	if user.sumData.careerdata.TotalKillNum != 0 {
		roundData.HeadShotRate = fmt.Sprintf("%.4f", float32(roundData.TotalHeadShot)/float32(user.sumData.careerdata.TotalKillNum))
	} else {
		roundData.HeadShotRate = "0.0000"
	}
	roundData.RecvItemUseNum = user.sumData.careerdata.RecvItemUseNum
	roundData.CarUseNum = user.sumData.careerdata.CarUseNum
	roundData.CarDestroyNum = user.sumData.careerdata.CarDestroyNum
	roundData.KillRating = user.sumData.killRating
	roundData.WinRating = user.sumData.winRating
	roundData.RoundRating = user.sumData.rating
	roundData.TotalRating = user.sumData.careerdata.TotalRating

	roundData.AINum = scene.aiNum
	roundData.PlayerNum = scene.maxUserSum - scene.aiNum
	roundData.PlayerRunDistance = user.sumData.runDistance / 1000
	roundData.CarRunDistance = user.sumData.carDistance / 1000
	roundData.DEADTYPE = user.sumData.deadType
	roundData.RecoverHp = user.sumData.recoverHp
	roundData.BandageNum = user.sumData.bandageNum
	roundData.MedicalBoxNum = user.sumData.medicalBoxNum
	roundData.PainkillerNum = user.sumData.painkillerNum
	roundData.EnergyNum = user.sumData.speednum
	roundData.HeadShotNum = user.headshotnum
	roundData.EffectHarm = user.effectharm
	roundData.ShotNum = user.sumData.shotnum
	roundData.ReviveNum = user.sumData.revivenum
	roundData.KillDistance = user.sumData.killdistance
	roundData.KillStmNum = user.sumData.killstmnum
	roundData.RCarUseNum = user.sumData.carusernum
	roundData.RCarDestoryNum = user.sumData.carDestoryNum
	roundData.AttackNum = user.sumData.attacknum
	roundData.SkyBox = scene.skybox
	roundData.Kill = user.kill
	roundData.TotalSurviveTime = user.sumData.careerdata.SurviveTime
	roundData.TotalDistance = user.sumData.careerdata.TotalDistance
	roundData.TotalRank = user.sumData.careerdata.TotalRank
	roundData.SoloRating = user.sumData.careerdata.SoloRating
	roundData.DuoRating = user.sumData.careerdata.DuoRating
	roundData.SquadRating = user.sumData.careerdata.SquadRating
	roundData.SoloRank = user.sumData.careerdata.SoloRank
	roundData.DuoRank = user.sumData.careerdata.DuoRank
	roundData.SquadRank = user.sumData.careerdata.SquadRank
	roundData.TopRating = user.sumData.careerdata.TopRating
	roundData.TotalCoin = user.sumData.careerdata.Coin
	roundData.TotalCarDistance = user.sumData.careerdata.TotalCarDistance
	if roundData.BattleType != 0 {
		roundData.TeamID = user.GetUserTeamID()
	}
	roundData.GunID = user.sumData.gunID
	roundData.SightID = user.sumData.sightID
	roundData.SilenceID = user.sumData.silienceID
	roundData.MagazineID = user.sumData.magazineID
	roundData.StockID = user.sumData.stockID
	roundData.HandleID = user.sumData.handleID
	roundData.OpenIDByKill = user.sumData.openIDByKill
	roundData.GunIDByKill = user.sumData.gunIDByKill
	roundData.SightIDByKill = user.sumData.sightIDByKill
	roundData.SilenceIDByKill = user.sumData.silienceIDByKill
	roundData.MagazineIDByKill = user.sumData.magazineIDByKill
	roundData.StockIDByKill = user.sumData.stockIDByKill
	roundData.HandleIDByKill = user.sumData.handleIDByKill
	roundData.DeadIsHead = user.sumData.deadIsHead
	roundData.TankKillNum = user.sumData.tankKillNum
	roundData.TankUseTime = uint32(user.sumData.tankUseTime)
	roundData.WatchType = user.sumData.watchType
	roundData.WatchTime = user.sumData.watchEndTime - user.sumData.watchStartTime
	if roundData.PlayerNum > scene.loadFulPlayerNum {
		roundData.LoadFailPlayerNum = roundData.PlayerNum - scene.loadFulPlayerNum
	}
	roundData.VRoleName = user.sumData.userName
	roundData.MatchMode = scene.GetMatchMode()
	roundData.ParachuteType = uint32(user.parachuteType)

	tlog.Format(roundData)
}

// tlogBattleResult tlog战场结果表
func (user *RoomUser) tlogBattleResult() {
	loginMsg := user.GetPlayerLogin()
	if loginMsg == nil {
		log.Warn("PlayerLogin msg null:tlogBattleResult")
		return
	}

	msg := &protoMsg.BattleResult{}
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
	msg.VRoleName = user.sumData.userName

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
	space := user.GetSpace().(*Scene)
	if space != nil {
		msg.Mode = space.mapdata.Id
		msg.BattleID = space.GetID()
		msg.SkyBox = space.skybox

		msg.Times = space.refreshzone.GetRefreshCount() - space.mapdata.Id*1000
		if msg.Times > 10 {
			msg.Times = 0
		}
	}
	msg.TeamType = user.sumData.battletype
	if msg.TeamType != 0 {
		msg.TeamID = user.GetUserTeamID()
	}
	mapid := user.GetPos()
	msg.MapIDX = mapid.X
	msg.MapIDY = mapid.Y
	msg.MapIDZ = mapid.Z
	msg.MapType = user.judgeMapType(mapid)
	msg.RoundTime = user.sumData.endgametime - user.sumData.begingametime
	msg.ResultType = user.sumData.deadType
	msg.RoundScore = user.sumData.rating
	msg.Rank = user.sumData.rank
	msg.Hurt = user.effectharm
	if user.kill != 0 {
		msg.Cirt = user.headshotnum / user.kill
	}
	msg.MoneyProduce = user.sumData.coin
	msg.Distance = user.sumData.runDistance
	msg.MatchMode = space.GetMatchMode()

	tlog.Format(msg)
}

// tlogGunFlow tlog枪支伤害信息表
func (user *RoomUser) tlogGunFlow(dps uint32, injuredResult uint32, isshothead bool, killDistance float32) {
	if user.isMeleeWeaponUse() {
		return
	}

	gunID, sightID, silenceID, magazineID, stockID, handleID := user.getUseGunInfo()

	loginMsg := user.GetPlayerLogin()
	if loginMsg == nil {
		log.Warn("PlayerLogin msg null:tlogGunFlow")
		return
	}

	msg := &protoMsg.GunFlow{}
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
	msg.VRoleName = user.sumData.userName
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

	space := user.GetSpace().(*Scene)
	if space != nil {
		msg.Mode = space.mapdata.Id
		msg.BattleID = space.GetID()
	}
	msg.TeamType = user.sumData.battletype
	if msg.TeamType != 0 {
		msg.TeamID = user.GetUserTeamID()
	}
	mapid := user.GetPos()
	msg.MapIDX = mapid.X
	msg.MapIDY = mapid.Y
	msg.MapIDZ = mapid.Z
	msg.GunID = gunID
	msg.SightID = sightID
	msg.SilenceID = silenceID
	msg.MagazineID = magazineID
	msg.Distance = killDistance
	msg.Dps = dps
	if injuredResult == SubHpDie || injuredResult == WillDieSubHpToWatch || injuredResult == WillDieSubHpToDead {
		msg.Kill = 1
	}
	if isshothead {
		msg.HeadKill = 1
	}
	msg.StockID = stockID
	msg.HandleID = handleID
	msg.MatchMode = space.GetMatchMode()

	tlog.Format(msg)
}

// behaveType 分类
const (
	behavetype_ground    = 0 //落地
	behavetype_pickup    = 1 //拾取
	behavetype_throw     = 2 //丢弃
	behavetype_install   = 3 //安装配件
	behavetype_dismount  = 4 //卸下配件
	behavetype_usecar    = 5 //使用载具
	behavetype_jump      = 6 //跳伞
	behavetype_forcejump = 7 //强制跳伞
)

// tlogBattleFlow tlog战场流水表
func (user *RoomUser) tlogBattleFlow(behaveType uint32, afItemID, beItemID uint64, afItemLevel, beItemLevel, bagItemNum uint32) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg == nil {
		log.Warn("PlayerLogin msg null:tlogBattleFlow")
		return
	}

	msg := &protoMsg.BattleFlow{}
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
	msg.VRoleName = user.sumData.userName
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

	space := user.GetSpace().(*Scene)
	if space != nil {
		msg.Mode = space.mapdata.Id
		msg.BattleID = space.GetID()
		msg.SkyBox = space.skybox

		msg.Times = space.refreshzone.GetRefreshCount() - space.mapdata.Id*1000
		if msg.Times > 10 {
			msg.Times = 0
		}
	}
	msg.TeamType = user.sumData.battletype
	if msg.TeamType != 0 {
		msg.TeamID = user.GetUserTeamID()
	}
	msg.BehaveType = behaveType
	mapid := user.GetPos()
	msg.MapIDX = mapid.X
	msg.MapIDY = mapid.Y
	msg.MapIDZ = mapid.Z
	msg.MapType = user.judgeMapType(mapid)
	msg.AfIemID = afItemID
	msg.BeItemID = beItemID
	msg.AfItemLevel = afItemLevel
	msg.BeItemLevel = beItemLevel
	msg.BagItemNum = bagItemNum
	msg.RoundTime = time.Now().Unix() - user.sumData.begingametime
	msg.Distance = user.sumData.runDistance
	msg.MatchMode = space.GetMatchMode()

	tlog.Format(msg)
}

// judgeMapType 判断玩家位置（安全区内为1，安全区外毒气圈内为2，安全区外毒气圈外为3，轰炸区内为4）
func (user *RoomUser) judgeMapType(mapid linmath.Vector3) uint32 {
	space := user.GetSpace().(*Scene)
	if space == nil {
		return 0
	}
	if space.refreshzone == nil {
		return 0
	}

	if common.Distance(mapid, space.refreshzone.nextsafecenter) <= space.refreshzone.nextsaferadius {
		return 1
	} else if GetRefreshZoneMgr(space).InBombArea(mapid) {
		return 4
	} else if common.Distance(mapid, space.safecenter) <= space.saferadius {
		return 2
	} else {
		return 3
	}
}

// tlogSecGameStartFlow (安全tlog)游戏开始流水表
func (user *RoomUser) tlogSecGameStartFlow() {
	loginMsg := user.GetPlayerLogin()
	if loginMsg == nil {
		log.Warn("PlayerLogin msg null:tlogSecGameStartFlow")
		return
	}

	msg := &protoMsg.SecGameStartFlow{}
	msg.GameSvrId = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
	msg.DtEventTime = time.Unix(user.sumData.begingametime, 0).Format("2006-01-02 15:04:05")
	msg.GameAppID = loginMsg.VGameAppID
	if msg.GameAppID == "" {
		msg.GameAppID = "NoVGameAppID"
	}
	msg.OpenID = loginMsg.VOpenID
	if msg.OpenID == "NoOpenID" || msg.OpenID == "" {
		msg.OpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
	}
	msg.PlatID = loginMsg.PlatID
	msg.AreaID = loginMsg.LoginChannel

	space := user.GetSpace().(*Scene)
	if space == nil {
		log.Warn("获取scene fail")
		return
	}

	msg.SvrRoleID = user.GetDBID()
	args := []interface{}{
		"Picture",
	}
	values, valueErr := dbservice.EntityUtil("Player", msg.SvrRoleID).GetValues(args)
	if valueErr != nil || len(values) != 1 {
		log.Error("获取url错误")
		return
	}
	tmpUrl, urlErr := redis.String(values[0], nil)
	if urlErr == nil {
		msg.PicUrl = tmpUrl
	}
	if msg.PicUrl == "" {
		msg.PicUrl = "no PicUrl"
	}
	msg.ZoneID = 0
	msg.BattleID = space.GetID()
	msg.ClientStartTime = msg.DtEventTime
	msg.UserName = user.sumData.userName
	msg.SvrUserMoney1 = user.GetCoin()
	msg.SvrUserMoney2 = 0
	msg.SvrUserMoney3 = 0

	season := common.GetSeason()
	careerdata, err := common.GetPlayerCareerData(user.GetDBID(), season, 0)
	if err != nil {
		user.Error("GetPlayerCareerData err: ", err)
		return
	}

	msg.SvrRoundRank = careerdata.TotalRating
	msg.SvrRoundRank1 = careerdata.SoloRating
	msg.SvrRoundRank2 = careerdata.DuoRating
	msg.SvrRoundRank3 = careerdata.SquadRating
	msg.SvrRoleType = user.GetRoleModel()
	// msg.SvrMapid = space.get
	msg.SvrWeatherid = space.skybox
	msg.SvrItemList = "0"
	msg.WaitStartTime = msg.DtEventTime
	msg.WaitEndTime = msg.DtEventTime
	msg.RoleType = 0
	msg.Mapid = 0
	msg.Weatherid = 0
	msg.ItemList = "0"

	msg.TeamType = 0
	msg.AutoMatch = 0
	msg.TeamPlayer1 = "0"
	msg.TeamPlayer2 = "0"
	msg.TeamPlayer3 = "0"

	if !space.teamMgr.isTeam {
		msg.GameType = 0
	} else {
		msg.TeamID = user.GetUserTeamID()

		if space.teamMgr.teamType == 0 {
			msg.GameType = 1
			msg.TeamType = 2
			msg.AutoMatch = 0
		} else if space.teamMgr.teamType == 1 {
			msg.GameType = 2
			msg.TeamType = 4
			msg.AutoMatch = 0
		}

		user.secGameStartFlow(msg)
	}
	msg.PlayerCount = space.maxUserSum

	tlog.Format(msg)
}

// secGameStartFlow 组队队友信息获取(开始流水)
func (user *RoomUser) secGameStartFlow(msg *protoMsg.SecGameStartFlow) {
	space := user.GetSpace().(*Scene)
	if space == nil {
		log.Warn("duoSecGameStartFlow 获取scene fail")
		return
	}
	season := common.GetSeason()

	i := 0
	for _, v := range space.teamMgr.teams[msg.TeamID] {
		if v == msg.SvrRoleID {
			continue
		}
		i++
		teamPlayer, err := dbservice.Account(v).GetUsername()
		if err != nil {
			log.Error("tlogSecGameEndFlow 获取openid fail")
			return
		}

		careerdata, err := common.GetPlayerCareerData(v, season, 0)
		if err != nil {
			user.Error("GetPlayerCareerData err: ", err)
			return
		}

		if i == 1 {
			msg.TeamPlayer1 = teamPlayer
			msg.TeamPlayer1Rank = careerdata.TotalRating
		} else if i == 2 {
			msg.TeamPlayer2 = teamPlayer
			msg.TeamPlayer2Rank = careerdata.TotalRating
		} else if i == 3 {
			msg.TeamPlayer3 = teamPlayer
			msg.TeamPlayer3Rank = careerdata.TotalRating
		}
	}
}

// tlogSecGameEndFlow (安全tlog)结算日志
func (user *RoomUser) tlogSecGameEndFlow() {
	loginMsg := user.GetPlayerLogin()
	if loginMsg == nil {
		log.Warn("PlayerLogin msg null:tlogSecGameEndFlow")
		return
	}
	msg := &protoMsg.SecGameEndFlow{}
	msg.GameSvrId = fmt.Sprintf("%d", GetSrvInst().GetSrvID())
	msg.DtEventTime = time.Unix(user.sumData.endgametime, 0).Format("2006-01-02 15:04:05")
	msg.GameAppID = loginMsg.VGameAppID
	if msg.GameAppID == "" {
		msg.GameAppID = "NoVGameAppID"
	}
	msg.OpenID = loginMsg.VOpenID
	if msg.OpenID == "NoOpenID" || msg.OpenID == "" {
		msg.OpenID, _ = dbservice.Account(user.GetDBID()).GetUsername()
	}
	msg.PlatID = loginMsg.PlatID
	msg.AreaID = loginMsg.LoginChannel
	msg.ZoneID = 0

	space := user.GetSpace().(*Scene)
	if space == nil {
		log.Warn("获取scene fail")
		return
	}
	msg.BattleID = space.GetID()
	msg.ClientStartTime = msg.DtEventTime
	msg.UserName = user.sumData.userName
	msg.RoleID = user.GetDBID()

	msg.RoleType = 0
	msg.OverTime = user.sumData.endgametime - user.sumData.begingametime
	msg.EndType = user.sumData.deadType
	msg.KillCount = user.kill
	msg.AssistsCount = user.sumData.attacknum
	msg.SaveCount = 0
	if user.sumData.deadType == 9 {
		msg.DropCount = user.sumData.revivenum
		msg.AliveType = 0
	} else {
		msg.DropCount = user.sumData.revivenum + 1
		msg.AliveType = 1
	}
	msg.RebornCount = msg.DropCount
	msg.GoldGet = user.sumData.coin
	msg.DiamondGet = 0
	msg.ExpGet = 0
	msg.WinRank = user.sumData.rank
	msg.TotalPlayers = space.maxUserSum
	msg.RankEnd = user.sumData.careerdata.TotalRating

	msg.TeamPlayer1 = "0"
	msg.TeamPlayer2 = "0"
	msg.TeamPlayer3 = "0"

	if space.teamMgr.isTeam {
		msg.TeamID = user.GetUserTeamID()
		user.secGameEndFlow(msg)
	}

	tlog.Format(msg)
}

// secGameEndFlow 组队信息获取(结算流水)
func (user *RoomUser) secGameEndFlow(msg *protoMsg.SecGameEndFlow) {
	space := user.GetSpace().(*Scene)
	if space == nil {
		log.Warn("duoSecGameStartFlow 获取scene fail")
		return
	}

	i := 0
	for _, v := range space.teamMgr.teams[msg.TeamID] {
		if v == msg.RoleID {
			continue
		}
		i++

		openID, _ := dbservice.Account(v).GetUsername()
		if i == 1 {
			msg.TeamPlayer1 = openID
		} else if i == 2 {
			msg.TeamPlayer2 = openID
		} else if i == 3 {
			msg.TeamPlayer3 = openID
		}

		memberInfo := space.teamMgr.GetSpaceMemberInfo(v)
		if memberInfo == nil {
			continue
		}

		var TeamPlayerAliveType uint32
		user, ok := space.GetEntityByDBID("Player", v).(*RoomUser)
		if ok {
			if user.sumData.deadType == 9 { //吃鸡
				TeamPlayerAliveType = 1
			} else {
				TeamPlayerAliveType = 0
			}
		}

		if i == 1 {
			msg.TeamPlayer1AliveType = TeamPlayerAliveType
			msg.TeamPlayer1Kill = memberInfo.killnum
		} else if i == 2 {
			msg.TeamPlayer2AliveType = TeamPlayerAliveType
			msg.TeamPlayer2Kill = memberInfo.killnum
		} else if i == 3 {
			msg.TeamPlayer3AliveType = TeamPlayerAliveType
			msg.TeamPlayer3Kill = memberInfo.killnum
		}
	}
}

//tlogInsigniaFlow 勋章的tlog日志
func (user *RoomUser) tlogInsigniaFlow(id uint32, num uint32) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg == nil {
		log.Warn("PlayerLogin msg null:tlogBattleResult")
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
	msg.VRoleName = user.sumData.userName

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
	msg.LogType = 1
	msg.InsigniaID = id
	msg.InsigniaNum = num
	tlog.Format(msg)
}

//AchievementFlow 勋章系统
func (user *RoomUser) tlogAchievementFlow(id uint32, level uint32, exp uint32) {
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

//tlogCheaterReportFlow 举报
func (user *RoomUser) tlogCheaterReportFlow(uid uint64, id uint32, reportNum map[uint32]uint32) {
	loginMsg := user.GetPlayerLogin()
	if loginMsg == nil {
		user.Warn("PlayerLogin msg null:tlogCheaterReportFlow")
		return
	}

	msg := &protoMsg.CheaterReportFlow{}
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

	msg.ReportOpenID, _ = dbservice.Account(uid).GetUsername()
	name, err := dbservice.EntityUtil("Player", uid).GetValue("Name")
	if err == nil {
		msg.ReportName, _ = redis.String(name, nil)
	}
	if msg.ReportOpenID == "" {
		msg.ReportOpenID = "Ai"
	}
	if msg.ReportName == "" {
		msg.ReportName = "Ai"
	}
	msg.ReportTypeID = id
	msg.ReportTypeSum = common.MapToString(reportNum)

	tlog.Format(msg)
}

// tlogSkillUseFlow 技能使用流水表
func (user *RoomUser) tlogSkillUseFlow() {
	loginMsg := user.GetPlayerLogin()
	if loginMsg == nil {
		user.Warn("PlayerLogin msg null:tlogSkillUseFlow")
		return
	}

	msg := &protoMsg.SkillUseFlow{}
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

	space := user.GetSpace().(*Scene)
	if space != nil {
		msg.MatchMode = space.GetMatchMode()
	}

	var initSkillID uint32
	for k, _ := range user.SkillMgr.initiveSkillData {
		initSkillID = k
	}
	msg.InitiveSkillID = initSkillID

	var passSkillID uint32
	for k, _ := range user.SkillMgr.passiveSkillData {
		passSkillID = k
	}
	msg.PassiveSkillID = passSkillID

	tlog.Format(msg)
}

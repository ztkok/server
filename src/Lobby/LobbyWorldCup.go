package main

import (
	"common"
	"db"
	"excel"
	"protoMsg"
	"strings"
	"time"
	"zeus/iserver"

	"strconv"

	log "github.com/cihub/seelog"
)

func (p *LobbyUserMsgProc) RPC_GetChampionContestInfo() {
	msg := &protoMsg.ChampionContestInfo{}
	util := db.NewPlayerWorldCupUtil(p.user.GetDBID())
	contest := util.GetContest()
	for k, v := range contest {
		msg.Info = append(msg.Info, &protoMsg.ContestInfo{
			TeamId: common.StringToUint32(k),
			Bouns:  common.StringToUint64(v),
		})
	}
	for _, v := range excel.GetWorldCupChampionMap() {
		if v.Out == 0 {
			msg.Team = append(msg.Team, uint32(v.Id))
		}
	}
	p.user.RPC(iserver.ServerTypeClient, "ChampionContestInfo", msg)
}

func (p *LobbyUserMsgProc) RPC_GetChampionContestRecord() {
	msg := &protoMsg.ChampionContestRecord{}
	util := db.NewPlayerWorldCupUtil(p.user.GetDBID())
	for _, r := range util.GetContestRecord() {
		msg.Record = append(msg.Record, &protoMsg.ContestRecord{
			TeamId: r.Id,
			Stamp:  r.Stamp,
			Odds:   r.Odds,
		})
	}
	p.user.RPC(iserver.ServerTypeClient, "ChampionContestRecord", msg)
}
func (p *LobbyUserMsgProc) RPC_ChampionContest(teamId uint32) {
	//活动是否开启
	activity, ok := GetSrvInst().activityMgr.GetActivity(worldCupChampion).(*WorldCupChampionActivity)
	if activity == nil || !ok {
		return
	}
	t, _ := activity.checkOpenDate()
	if t != 0 {
		return
	}
	// 队伍是否被淘汰
	teamData, ok := excel.GetWorldCupChampion(uint64(teamId))
	if !ok || teamData.Out == 1 {
		return
	}

	util := db.NewPlayerWorldCupUtil(p.user.GetDBID())
	// 次数最大值
	nums := common.StringToUint32(util.GetContestTime(teamId))
	if nums >= uint32(common.GetTBSystemValue(common.System_WCupDayContestNum)) {
		p.user.AdviceNotify(common.NotifyCommon, 82)
		return
	}
	if !p.user.storeMgr.ReduceGoods(2207, 1, common.RS_WorldCupChampion) {
		return
	}
	now := util.AddContest(teamId)
	odds := util.GetChampionOdds()
	util.AddContestRecord(teamId, odds[teamId])
	p.user.tlogWorldCupChampionFlow(teamId, odds[teamId])
	p.user.RPC(iserver.ServerTypeClient, "ChampionContest", &protoMsg.ContestInfo{
		TeamId: teamId,
		Bouns:  now,
	})
}

func (p *LobbyUserMsgProc) RPC_GetChampionOddsInfo() {
	activity, ok := GetSrvInst().activityMgr.GetActivity(worldCupChampion).(*WorldCupChampionActivity)
	if !ok || activity == nil {
		return
	}
	msg := &protoMsg.ChampionOddsInfo{}
	for k, v := range activity.GetChampionOdds() {
		msg.Info = append(msg.Info, &protoMsg.OddsInfo{
			TeamId: k,
			Odds:   v,
		})
	}
	p.user.RPC(iserver.ServerTypeClient, "ChampionOddsInfo", msg)
}

func (p *LobbyUserMsgProc) RPC_WorldCupMatchRecordInfo() {
	msg := &protoMsg.WorldCupMatchRecordInfo{}
	util := db.NewPlayerWorldCupUtil(p.user.GetDBID())
	for _, r := range util.GetMatchContestRecord() {
		info, ok := excel.GetWorldCupBattle(uint64(r.MatchId))
		if !ok {
			continue
		}
		msg.Info = append(msg.Info, &protoMsg.WorldCupMatchRecord{
			MatchId: r.MatchId,
			Stamp:   r.Stamp,
			Kind:    r.Index,
			Odds:    r.Odds,
			Team:    []string{info.TeamA, info.TeamB},
		})
	}
	p.user.RPC(iserver.ServerTypeClient, "WorldCupMatchRecordInfo", msg)
}
func (p *LobbyUserMsgProc) RPC_WorldCupMatchInfo() {
	activity, ok := GetSrvInst().activityMgr.GetActivity(worldCupMatch).(*WorldCupMatchActivity)
	if activity == nil || !ok {
		return
	}
	msg := &protoMsg.WorldCupMatchInfo{}
	util := db.NewPlayerWorldCupUtil(p.user.GetDBID())
	contest := util.GetMatchContest()
	reward := util.GetMatchRewardAll()
	for _, battle := range excel.GetWorldCupBattleMap() {
		s := common.Uint64ToString(battle.Id)
		battleInfo := &protoMsg.WorldCupMatch{
			MatchId:    uint32(battle.Id),
			Score:      []uint32{uint32(battle.ScoreA), uint32(battle.ScoreB)},
			Odds:       activity.GetMatchOdds(uint32(battle.Id)),
			IsDone:     1,
			Team:       []string{battle.TeamA, battle.TeamB},
			ScoreFinal: []uint32{uint32(battle.ScoreAfinal), uint32(battle.ScoreBfinal)},
		}
		battleInfo.Contest = []uint64{0, 0, 0}
		if _, ok := contest[s]; ok {
			battleInfo.Contest = contest[s]
		}
		if _, ok := reward[s]; ok {
			battleInfo.IsReward = 1
		}
		if battle.ScoreB == 100 || battle.ScoreA == 100 {
			battleInfo.IsDone = 0
		}
		msg.Info = append(msg.Info, battleInfo)
	}
	p.user.RPC(iserver.ServerTypeClient, "WorldCupMatchInfo", msg)
}

func (p *LobbyUserMsgProc) RPC_WorldCupMatchContest(matchId uint32, index uint32) {
	//活动是否开启
	activity, ok := GetSrvInst().activityMgr.GetActivity(worldCupMatch).(*WorldCupMatchActivity)
	if activity == nil || !ok {
		p.user.Debugf("not exists ")
		return
	}
	t, _ := activity.checkOpenDate()
	if t != 0 {
		p.user.Debugf("not start ", t)
		return
	}

	// 比赛是否开始
	battleData, ok := excel.GetWorldCupBattle(uint64(matchId))
	if !ok || common.TimeStringToTime(battleData.Time).Unix()-int64(common.GetTBSystemValue(common.System_WCupOddsTime)) < time.Now().Unix() {
		return
	}
	_, e1 := strconv.Atoi(battleData.TeamA)
	_, e2 := strconv.Atoi(battleData.TeamB)
	if e1 != nil || e2 != nil {
		return
	}
	if battleData.ScoreB != 100 && battleData.ScoreA != 100 {
		return
	}
	if index > 3 {
		return
	}
	util := db.NewPlayerWorldCupUtil(p.user.GetDBID())
	// 次数最大值
	nums := common.StringToUint32(util.GetContestMatchTime())
	if nums >= uint32(common.GetTBSystemValue(common.System_WCupDayMatchNum)) {
		p.user.AdviceNotify(common.NotifyCommon, 82)
		return
	}
	if !p.user.storeMgr.ReduceGoods(2207, 1, common.RS_WorldCupMatch) {
		return
	}
	now := util.AddMatchContest(matchId, index)
	odds := activity.GetMatchOdds(matchId)
	if len(odds) != 3 || odds[int(index)] == 0 {
		return
	}
	util.AddMatchContestRecord(matchId, index, odds[int(index)])
	reward := util.IsMatchReward(matchId)
	info := &protoMsg.WorldCupMatch{
		MatchId:    matchId,
		Contest:    now,
		Score:      []uint32{uint32(battleData.ScoreA), uint32(battleData.ScoreB)},
		Odds:       activity.GetMatchOdds(uint32(battleData.Id)),
		IsDone:     1,
		Team:       []string{battleData.TeamA, battleData.TeamB},
		IsReward:   0,
		ScoreFinal: []uint32{uint32(battleData.ScoreAfinal), uint32(battleData.ScoreBfinal)},
	}
	if reward {
		info.IsReward = 1
	}
	if battleData.ScoreB == 100 || battleData.ScoreA == 100 {
		info.IsDone = 0
	}
	p.user.RPC(iserver.ServerTypeClient, "WorldCupMatchContest", info)
	p.user.tlogWorldCupMatchFlow(matchId, index, odds[index])
}

func (p *LobbyUserMsgProc) RPC_WorldCupMatchReward(matchId uint32) {
	battleData, ok := excel.GetWorldCupBattle(uint64(matchId))
	if !ok || battleData.ScoreA == 100 || battleData.ScoreB == 100 {
		return
	}
	util := db.NewPlayerWorldCupUtil(p.user.GetDBID())
	if util.IsMatchReward(matchId) {
		return
	}
	var winner uint32
	if battleData.ScoreA > battleData.ScoreB {
		winner = 0
	} else if battleData.ScoreA < battleData.ScoreB {
		winner = 2
	} else {
		winner = 1
	}
	var nums float32
	for _, record := range util.GetMatchContestRecord() {
		if record.MatchId != matchId {
			continue
		}
		if record.Index == winner {
			nums += record.Odds
		}
	}
	if nums == 0 {
		return
	}
	p.user.storeMgr.GetGoods(2207, uint32(nums), common.RS_WorldCupMatch, common.MT_NO, 0)
	util.SetMatchReward(matchId, nums)
	p.user.RPC(iserver.ServerTypeClient, "WorldCupMatchReward", matchId)
}

// NewWorldCupMatchActivity 世界杯胜负竞猜活动
func NewWorldCupMatchActivity() *WorldCupMatchActivity {
	t := WorldCupMatchActivity{}
	t.id = worldCupMatch
	return &t
}

type WorldCupMatchActivity struct {
	Activity
}

func (wcup *WorldCupMatchActivity) doHour() {
	now := time.Now().Unix()
	for _, battle := range excel.GetWorldCupBattleMap() {
		if now > common.TimeStringToTime(battle.Time).Unix() {
			continue
		}
		wcup.GetMatchOdds(uint32(battle.Id))
	}

}

// GetMatchOdds 获取某个比赛的赔率
func (wcup *WorldCupMatchActivity) GetMatchOdds(matchId uint32) []float32 {
	util := db.NewPlayerWorldCupUtil(0)
	data := util.GetMatchOdds(matchId)
	if len(data) > 0 {
		return data
	}

	battle, ok := excel.GetWorldCupBattle(uint64(matchId))
	if !ok {
		return []float32{0, 0, 0}
	}
	ret := []float32{
		battle.RateA,
		battle.RateEqual,
		battle.RateB,
	}
	if ret[0] == 0 {
		return ret
	}
	bonus := util.GetMatchBonus(matchId)
	if len(bonus) == 0 {
		return ret
	}
	var sum float64
	for _, v := range bonus {
		sum += common.StringToFloat64(v)
	}

	for k, v := range bonus {
		t := common.StringToFloat64(v)
		if t == 0 {
			continue
		}
		tmp := strings.Split(k, "_")
		ret[common.StringToUint32(tmp[1])] = float32(common.RoundFloat(1/(t/sum)*float64(common.GetTBSystemValue(common.System_WCupPayRate))/100, 2))
	}
	expire := common.GetNextHoursStamp()
	//赔率不再更新
	if common.TimeStringToTime(battle.Time).Unix()-300 < time.Now().Unix() {
		expire = 0
	}
	util.SetMatchOdds(ret, matchId, expire)
	ret = util.GetMatchOdds(matchId)
	return ret
}

//NewWorldCupChampionActivity 世界杯冠军竞猜活动
func NewWorldCupChampionActivity() *WorldCupChampionActivity {
	t := WorldCupChampionActivity{}
	t.id = worldCupChampion
	return &t
}

type WorldCupChampionActivity struct {
	Activity
}

func (wcup *WorldCupChampionActivity) doHour() {
	wcup.GetChampionOdds()
}

func (wcup *WorldCupChampionActivity) GetChampionOdds() map[uint32]float32 {
	util := db.NewPlayerWorldCupUtil(0)
	data := util.GetChampionOdds()
	if len(data) > 0 {
		return data
	}
	ret := map[uint32]float32{}
	bonus := util.GetTeamBonus()
	for _, v := range excel.GetWorldCupChampionMap() {
		ret[uint32(v.Id)] = float32(v.Rate)
	}

	var sum float64
	for _, v := range bonus {
		sum += common.StringToFloat64(v)
	}

	for k, v := range bonus {
		t := common.StringToFloat64(v)
		ret[common.StringToUint32(k)] = float32(common.RoundFloat(1/(t/sum)*float64(common.GetTBSystemValue(common.System_WCupPayRate))/100, 2))
	}

	util.SetChampionOdds(ret, common.GetNextHoursDur())
	ret = util.GetChampionOdds()
	return ret
}

// endReward 活动结束后 结算
func (wcup *WorldCupChampionActivity) endReward() {
	dur := common.GetTBSystemValue(common.System_WCupSendReward) * 3600
	if wcup.tmEnd[0]+int64(dur) > time.Now().Unix() {
		return
	}

	var winner uint32
	for _, v := range excel.GetWorldCupChampionMap() {
		if v.Out == 0 {
			winner = uint32(v.Id)
			break
		}
	}

	for _, uid := range GetSrvInst().GetOnlineIds() {
		util := db.NewPlayerWorldCupUtil(uid)
		if util.IsChampionReward() {
			continue
		}

		recordArr := util.GetContestRecord()
		if len(recordArr) == 0 {
			continue
		}

		var nums float32
		for _, r := range recordArr {
			if r.Id == winner {
				nums += r.Odds
			}
		}
		util.ChampionReward()

		if nums > 0 {
			mail, ok := excel.GetMail(common.Mail_ChampionReward)
			if !ok {
				log.Error("NotFound Mail ", common.Mail_ChampionReward)
			}
			sendObjMail(uid, "", 0, mail.MailTitle, mail.Mail, "", "", map[uint32]uint32{2207: uint32(nums)})
		} else {
			mail, ok := excel.GetMail(common.Mail_ChampionOut)
			if !ok {
				log.Error("NotFound Mail ", common.Mail_ChampionOut)
			}
			sendObjMail(uid, "", 0, mail.MailTitle, mail.Mail, "", "", nil)
		}
	}
}

func (srv *LobbySrv) GetOnlineIds() []uint64 {
	var ids []uint64
	srv.TravsalEntity("Player", func(entity iserver.IEntity) {
		ids = append(ids, entity.GetDBID())
	})
	return ids
}

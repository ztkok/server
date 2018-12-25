package data

import (
	"database/sql"
	"datadef"
	"fmt"
	"time"

	log "github.com/cihub/seelog"
	_ "github.com/go-sql-driver/mysql" //mysql连接库
)

type CareerService struct {
	mysqlDB *sql.DB
}

// NewMysqlService 新建一个sql连接服务
func NewCareerService(user, pwd, addr, db string, maxConns, maxIdleConns int, maxLifetime time.Duration) *CareerService {
	log.Info("user:", user, " addr:", addr, " db:", db)
	s := &CareerService{}
	var err error
	s.mysqlDB, err = sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", user, pwd, addr, db))

	s.mysqlDB.SetMaxOpenConns(maxConns)
	s.mysqlDB.SetMaxIdleConns(maxIdleConns)
	s.mysqlDB.SetConnMaxLifetime(maxLifetime)

	if err != nil {
		log.Error(err)
		return nil
	}

	return s
}

// InsertDatas 插入每局数据
func (csv *CareerService) InsertRoundDatas(table string, req *datadef.OneRoundData) error {
	if req == nil {
		log.Error("insert数据err")
		return nil
	}

	sqlStr := fmt.Sprintf(
		`INSERT INTO %s(GameID,UID,Season,Rank,KillNum,HeadShotNum,EffectHarm,RecoverUseNum,ShotNum,ReviveNum,KillDistance,
		KillStmNum,FinallHp,RecoverHp,RunDistance,CarUseNum,CarDestoryNum,AttackNum,SpeedNum,Coin,StartUnix,EndUnix,StartTime,
		EndTime,BattleType,KillRating,WinRatig,SoloKillRating,SoloWinRating,SoloRating,SoloRank,DuoKillRating,DuoWinRating,
		DuoRating,DuoRank,SquadKillRating,SquadWinRating,SquadRating,SquadRank,TotalRating,TotalRank,TopRating,TotalBattleNum,
		TotalFirstNum,TotalTopTenNum,TotalHeadShot,TotalKillNum,TotalShotNum,TotalEffectHarm,TotalSurviveTime,TotalDistance,
		TotalRecvItemUseNum,SingleMaxKill,SingleMaxHeadShot,DeadType,UserName,SkyBox,PlatID,LoginChannel) VALUES 
		(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		table)
	log.Info("InsertRoundDatas:", sqlStr)

	_, err := csv.mysqlDB.Exec(sqlStr, req.GameID, req.UID, req.Season, req.Rank, req.KillNum, req.HeadShotNum, req.EffectHarm, req.RecoverUseNum,
		req.ShotNum, req.ReviveNum, req.KillDistance, req.KillStmNum, req.FinallHp, req.RecoverHp, req.RunDistance, req.CarUseNum, req.CarDestoryNum,
		req.AttackNum, req.SpeedNum, req.Coin, req.StartUnix, req.EndUnix, req.StartTime, req.EndTime, req.BattleType, req.KillRating, req.WinRatig,
		req.SoloKillRating, req.SoloWinRating, req.SoloRating, req.SoloRank, req.DuoKillRating, req.DuoWinRating, req.DuoRating, req.DuoRank,
		req.SquadKillRating, req.SquadWinRating, req.SquadRating, req.SquadRank, req.TotalRating, req.TotalRank, req.TopRating, req.TotalBattleNum,
		req.TotalFirstNum, req.TotalTopTenNum, req.TotalHeadShot, req.TotalKillNum, req.TotalShotNum, req.TotalEffectHarm, req.TotalSurviveTime,
		req.TotalDistance, req.TotalRecvItemUseNum, req.SingleMaxKill, req.SingleMaxHeadShot, req.DeadType, req.UserName, req.SkyBox, req.PlatID, req.LoginChannel)
	log.Info("RoundData: ", req)
	if err != nil {
		return err
	}

	return nil
}

// InsertDayDatas 插入每天统计的数据
func (csv *CareerService) InsertDayDatas(table string, req *datadef.OneDayData) error {
	if req == nil {
		log.Error("insert数据err")
		return nil
	}

	sqlStr := fmt.Sprintf(
		`INSERT INTO %s(DayId,Season,UID,NowTime,StartTime,Model,DayFirstNum,DayTopTenNum,Rating,WinRating,KillRating,DayEffectHarm,
		DayShotNum,DaySurviveTime,DayDistance,DayAttackNum,DayRecoverNum,DayRevivenum,DayHeadShotNum,DayBattleNum,DayKillNum,TotalRank,Rank,UserName) 
		VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`, table)
	log.Info("InsertDayDatas: ", sqlStr)

	_, err := csv.mysqlDB.Exec(sqlStr, req.DayId, req.Season, req.UID, req.NowTime, req.StartTime, req.Model, req.DayFirstNum, req.DayTopTenNum,
		req.Rating, req.WinRating, req.KillRating, req.DayEffectHarm, req.DayShotNum, req.DaySurviveTime, req.DayDistance, req.DayAttackNum,
		req.DayRecoverNum, req.DayRevivenum, req.DayHeadShotNum, req.DayBattleNum, req.DayKillNum, req.TotalRank, req.Rank, req.UserName)
	log.Info("DayDatas: ", req)
	if err != nil {
		return err
	}

	return nil
}

// UpdataDayDatas 更新每天统计的数据
func (csv *CareerService) UpdateDayDatas(table string, dayid, uid uint64, model uint32, req *datadef.OneDayData) error {
	if req == nil {
		log.Error("insert数据err")
		return nil
	}

	sqlStr := fmt.Sprintf(`UPDATE %s SET DayId=?,NowTime=?,StartTime=?,DayFirstNum=DayFirstNum+?,DayTopTenNum=DayTopTenNum+?,Rating=?,WinRating=?,
		KillRating=?,DayEffectHarm=DayEffectHarm+?,DayShotNum=DayShotNum+?,DaySurviveTime=DaySurviveTime+?,DayDistance=DayDistance+?,
		DayAttackNum=DayAttackNum+?,DayRecoverNum=DayRecoverNum+?,DayRevivenum=DayRevivenum+?,DayHeadShotNum=DayHeadShotNum+?,
		DayBattleNum=DayBattleNum+?,DayKillNum=DayKillNum+?,TotalRank=?,Rank=? WHERE UID=%d AND DayID=%d AND Model=%d`,
		table, uid, dayid, model)
	log.Info("UpdateDayDatas ", sqlStr)

	_, err := csv.mysqlDB.Exec(sqlStr, req.DayId, req.NowTime, req.StartTime, req.DayFirstNum, req.DayTopTenNum, req.Rating, req.WinRating,
		req.KillRating, req.DayEffectHarm, req.DayShotNum, req.DaySurviveTime, req.DayDistance, req.DayAttackNum, req.DayRecoverNum,
		req.DayRevivenum, req.DayHeadShotNum, req.DayBattleNum, req.DayKillNum, req.TotalRank, req.Rank)
	log.Info("Update DayDatas ", req)
	if err != nil {
		return err
	}

	return nil
}

// IsDayDatas 判断每天记录表中是否有该天的记录
func (csv *CareerService) IsDayDatas(table string, dayid, uid uint64, model uint32) (bool, error) {
	var gameID string
	sqlStr := fmt.Sprintf(`SELECT DayID FROM %s WHERE UID=%d AND DayID=%d AND Model=%d`,
		table, uid, dayid, model)
	log.Info("IsDayDatas:", sqlStr)
	err := csv.mysqlDB.QueryRow(sqlStr).Scan(&gameID)
	switch {
	case err == sql.ErrNoRows:
		return false, nil
	case err != nil:
		return false, err
	default:
		return true, nil
	}
}

// InsertSearchNumDatas 插入玩家搜索次数表
func (csv *CareerService) InsertSearchNumDatas(table string, req *datadef.SearchNumData) error {
	if req == nil {
		log.Error("insert数据err")
		return nil
	}

	sqlStr := fmt.Sprintf(`INSERT INTO %s(UID,UserName,SearchNum,TotalRating) VALUES (?,?,?,?)`, table)
	log.Info("InsertSearchNumDatas:", sqlStr)

	_, err := csv.mysqlDB.Exec(sqlStr, req.UID, req.UserName, req.SearchNum, req.TotalRating)
	if err != nil {
		return err
	}

	log.Info("Insert SearchNumDatas ", req)
	return nil
}

// UpdateSearchNumDatas 更新玩家搜索次数表
func (csv *CareerService) UpdateSearchNumDatas(table string, uid uint64, req *datadef.SearchNumData) error {
	if req == nil {
		log.Error("insert数据err")
		return nil
	}

	sqlStr := fmt.Sprintf(`UPDATE %s SET UID=?,UserName=?,SearchNum=SearchNum+?,TotalRating=? WHERE UID=%d`, table, uid)
	log.Info("UpdateSearchNumDatas:", sqlStr)

	_, err := csv.mysqlDB.Exec(sqlStr, req.UID, req.UserName, req.SearchNum, req.TotalRating)
	if err != nil {
		return err
	}

	log.Info("Update SearchNumDatas:", req)
	return nil
}

// IsSearchNumDatas 判断玩家搜索次数表是否有该玩家的记录
func (csv *CareerService) IsSearchNumDatas(table string, uid uint64) (bool, error) {
	var tmpUid uint64
	sqlStr := fmt.Sprintf(`SELECT UID FROM %s WHERE UID=%d`, table, uid)
	log.Info("IsSearchNumDatas:", sqlStr)
	err := csv.mysqlDB.QueryRow(sqlStr).Scan(&tmpUid)
	switch {
	case err == sql.ErrNoRows:
		return false, nil
	case err != nil:
		return false, err
	default:
		return true, nil
	}
}

// QueryCareerData 查询一局CareerData数据
func (csv *CareerService) QueryCareerData(table string, season int, uid uint64) (*datadef.CareerData, error) {
	careerdata := datadef.CareerData{}
	careerdata.Uid = uid
	sqlStr := fmt.Sprintf(
		`SELECT UID,SoloRating,SoloRank,DuoRating,DuoRank,SquadRating,SquadRank,TotalRating,TotalRank,
		TopRating,TotalBattleNum,TotalFirstNum,TotalTopTenNum,TotalHeadShot,TotalKillNum,TotalShotNum,TotalEffectHarm,
		TotalSurviveTime,TotalDistance,UserName FROM %s WHERE UID=%d AND Season=%d ORDER BY StartUnix DESC LIMIT 1`, table, uid, season)
	log.Info("QueryCareerData:", sqlStr)
	err := csv.mysqlDB.QueryRow(sqlStr).Scan(
		&careerdata.Uid, &careerdata.SoloRating, &careerdata.SoloRank, &careerdata.DuoRating, &careerdata.DuoRank, &careerdata.SquadRating,
		&careerdata.SquadRank, &careerdata.TotalRating, &careerdata.TotalRank, &careerdata.TopRating, &careerdata.TotalBattleNum,
		&careerdata.TotalFirstNum, &careerdata.TotalTopTenNum, &careerdata.TotalHeadShot, &careerdata.TotalKillNum, &careerdata.TotalShotNum,
		&careerdata.TotalEffectHarm, &careerdata.TotalSurviveTime, &careerdata.TotalDistance, &careerdata.UserName)
	log.Info("CareerData:", &careerdata)
	if err != nil {
		return &careerdata, err
	}
	return &careerdata, nil
}

// QueryMatchRecord 查询num天记录数据
func (csv *CareerService) QueryMatchRecord(table string, season int, uid uint64, num int) (*datadef.MatchRecord, error) {
	matchrecord := &datadef.MatchRecord{}
	matchrecord.Matchrecord = make([]*datadef.DayRecordData, 0)

	sqlStr := fmt.Sprintf(`SELECT DayId,NowTime,Model,DayFirstNum,DayTopTenNum,Rating,DayBattleNum,DayKillNum
		 FROM %s WHERE UID=%d AND Season=%d ORDER BY NowTime DESC LIMIT %d`, table, uid, season, num)
	log.Info("QueryMatchRecord:", sqlStr)
	rows, err := csv.mysqlDB.Query(sqlStr)
	if err != nil {
		return matchrecord, err
	}
	defer rows.Close()

	for rows.Next() {
		dayRecordData := datadef.DayRecordData{}
		err := rows.Scan(&dayRecordData.DayId, &dayRecordData.NowTime, &dayRecordData.Model, &dayRecordData.DayFirstNum,
			&dayRecordData.DayTopTenNum, &dayRecordData.Rating, &dayRecordData.DayBattleNum, &dayRecordData.DayKillNum)
		if err != nil {
			return matchrecord, err
		}
		matchrecord.Matchrecord = append(matchrecord.Matchrecord, &dayRecordData)
	}

	return matchrecord, nil
}

// QueryOneDayData 查询具体哪天记录数据
func (csv *CareerService) QueryOneDayData(table string, season int, uid uint64, dayid uint64) (*datadef.OneDayData, error) {
	req := datadef.OneDayData{}
	req.DayId = dayid
	req.UID = uid
	sqlStr := fmt.Sprintf(`SELECT * FROM %s WHERE DayId=%d AND UID=%d AND Season=%d`, table, dayid, uid, season)
	log.Info("QueryOneDayData:", sqlStr)

	err := csv.mysqlDB.QueryRow(sqlStr).Scan(
		&req.DayId, &req.Season, &req.UID, &req.NowTime, &req.StartTime, &req.Model, &req.DayFirstNum, &req.DayTopTenNum, &req.Rating,
		&req.WinRating, &req.KillRating, &req.DayEffectHarm, &req.DayShotNum, &req.DaySurviveTime, &req.DayDistance, &req.DayAttackNum,
		&req.DayRecoverNum, &req.DayRevivenum, &req.DayHeadShotNum, &req.DayBattleNum, &req.DayKillNum, &req.TotalRank, &req.Rank, &req.UserName)
	log.Info("OneDayData:", &req)
	if err != nil {
		return &req, err
	}

	return &req, nil
}

// QuerySearchNumData 获取搜索量最高的前num名玩家
func (csv *CareerService) QuerySearchNumData(table string, num int) ([]*datadef.SearchNumData, error) {
	searchNumDatas := make([]*datadef.SearchNumData, 0)

	sqlStr := fmt.Sprintf(`SELECT UID,UserName,SearchNum,TotalRating FROM %s ORDER BY SearchNum DESC LIMIT %d`, table, num)
	log.Info("QuerySearchNumData", sqlStr)

	rows, err := csv.mysqlDB.Query(sqlStr)
	if err != nil {
		return searchNumDatas, err
	}
	defer rows.Close()

	for rows.Next() {
		searchNumData := datadef.SearchNumData{}
		err := rows.Scan(&searchNumData.UID, &searchNumData.UserName, &searchNumData.SearchNum, &searchNumData.TotalRating)
		if err != nil {
			return searchNumDatas, err
		}
		searchNumDatas = append(searchNumDatas, &searchNumData)
	}

	return searchNumDatas, nil
}

// QueryRankTrend 获取rating趋势数据
func (csv *CareerService) QueryRankTrend(table string, season int, uid uint64, model uint32, timeStart string, timeEnd string) (*datadef.RankTrend, error) {
	rankTrend := &datadef.RankTrend{}
	rankTrend.DaysRating = make([]*datadef.DayRating, 0)

	sqlStr := fmt.Sprintf(`SELECT UID,StartTime,Rating FROM %s WHERE UID=%d AND Model=%d AND Season=%d AND "%s"<=
		StartTime<="%s"`, table, uid, model, season, timeStart, timeEnd)
	log.Info("QueryRankTrend:", sqlStr)

	rows, err := csv.mysqlDB.Query(sqlStr)
	if err != nil {
		return rankTrend, err
	}
	defer rows.Close()

	for rows.Next() {
		dayRating := datadef.DayRating{}
		err := rows.Scan(&dayRating.Uid, &dayRating.StartTime, &dayRating.Rating)
		if err != nil {
			return rankTrend, err
		}
		rankTrend.DaysRating = append(rankTrend.DaysRating, &dayRating)
	}

	return rankTrend, nil
}

// IsModelDetail 判断是否存在该模式详情
func (csv *CareerService) IsModelDetail(table string, season int, uid uint64, model uint32) (bool, error) {
	var num uint64
	sqlStr := fmt.Sprintf(`SELECT UID FROM %s WHERE UID=%d AND Model=%d AND Season=%d`,
		table, uid, model, season)
	log.Info("IsModelDetail:", sqlStr)

	err := csv.mysqlDB.QueryRow(sqlStr).Scan(&num)
	switch {
	case err == sql.ErrNoRows:
		return false, nil
	case err != nil:
		return false, err
	default:
		return true, nil
	}
}

// QueryModelDetail 查询模式详情
func (csv *CareerService) QueryModelDetail(table string, season int, uid uint64, model uint32) ([4]uint32, error) {
	var num [4]uint32
	sqlStr := fmt.Sprintf(`SELECT SUM(DayFirstNum),SUM(DayTopTenNum),SUM(DayBattleNum),SUM(DayKillNum) FROM %s WHERE UID=%d AND Model=%d AND Season=%d`,
		table, uid, model, season)
	log.Info("QueryModelDetail:", sqlStr)
	if err := csv.mysqlDB.QueryRow(sqlStr).Scan(&num[0], &num[1], &num[2], &num[3]); err != nil {
		return num, err
	}

	return num, nil
}

package idip

import (
	"common"
	"datadef"
	"db"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
	"zeus/dbservice"

	log "github.com/cihub/seelog"
	"github.com/garyburd/redigo/redis"
)

/*

 IDIP BAN_RANK_REQ 禁止参与排行榜接口

*/

// AqDoBanRankReq 禁止参与排行榜接口请求 4129
type AqDoBanRankReq struct {
	AreaID        uint32 `json:"AreaId"`        // 所在大区ID
	PlatID        uint8  `json:"PlatId"`        // 平台
	OpenID        string `json:"OpenId"`        // 用户OpenId
	RankData      uint8  `json:"RankData"`      //榜单数据（0 否，1 是）
	RankType      uint32 `json:"RankType"`      // 榜单类型（1=生涯榜，2=好友榜 ，99=所有榜单）
	BanTime       uint32 `json:"BanTime"`       // 禁止参与排行榜时长（秒）
	BanRankReason string `json:"BanRankReason"` // 禁止排行榜提示内容（文本）
	Source        uint32 `json:"Source"`        //渠道号，由前端生成，不需填写
	Serial        string `json:"Serial"`        //流水号，由前端生成，不需要填写
}

// GetCmdID 获取命令ID
func (req *AqDoBanRankReq) GetCmdID() uint32 {
	return 4129
}

// GetAreaID 获取AreaID
func (req *AqDoBanRankReq) GetAreaID() uint32 {
	return req.AreaID
}

// Do 执行封停角色请求
func (req *AqDoBanRankReq) Do() (interface{}, int32, error) {
	num, err := req.clearCareerRank()
	if err != nil {
		return nil, num, err
	}

	// 踢角色下线
	_, _, err = kickPlayer(req.OpenID, req.BanRankReason)
	if err != nil {
		return nil, ErrAqDoBanRankReqFailed, err
	}

	return &AqDoBanRankRsp{0, "success"}, num, nil
}

// clearRank 清空好友榜
func (req *AqDoBanRankReq) clearCareerRank() (int32, error) {
	season := common.GetSeason()
	uid, err := dbservice.GetUID(req.OpenID)
	if err != nil {
		return ErrAqDoBanRankReqFailed, err
	}
	args := []interface{}{
		"Name",
	}
	values, valueErr := dbservice.EntityUtil("Player", uid).GetValues(args)
	if valueErr != nil || len(values) != 1 {
		log.Error("获取url、qqvip错误")
		return ErrAqDoBanRankReqFailed, err
	}
	tmpUserName, urlErr := redis.String(values[0], nil)
	if urlErr != nil {
		log.Error("获取username错误")
		return ErrAqDoBanRankReqFailed, err
	}

	careerdata := &datadef.CareerBase{}
	playerCareerDataUtil := db.PlayerCareerDataUtil(common.PlayerCareerTotalData, uid, season)
	if err := playerCareerDataUtil.GetRoundData(careerdata); err != nil {
		return ErrAqDoBanRankReqFailed, err
	}
	careerdata.UID = uid
	careerdata.UserName = tmpUserName
	careerdata.SoloWinRating = 800
	careerdata.SoloKillRating = 800
	careerdata.DuoWinRating = 800
	careerdata.DuoKillRating = 800
	careerdata.SquadWinRating = 800
	careerdata.SquadKillRating = 800
	careerdata.SoloRating = 0
	careerdata.DuoRating = 0
	careerdata.SquadRating = 0
	careerdata.TotalRating = 0
	careerdata.TotalRank = 0
	careerdata.SoloRank = 0
	careerdata.DuoRank = 0
	careerdata.SquadRank = 0
	careerdata.TotalTopRating = 0
	careerdata.TotalTopRank = 0
	careerdata.TotalTopSoloRating = 0
	careerdata.TotalTopSoloRank = 0
	careerdata.TotalTopDuoRating = 0
	careerdata.TotalTopDuoRank = 0
	careerdata.TotalTopSquadRaing = 0
	careerdata.TotalTopSquadRank = 0
	err = playerCareerDataUtil.SetRoundData(careerdata)
	if err != nil {
		return ErrAqDoBanRankReqFailed, err
	}

	oneRoundData := &datadef.OneRoundData{}
	oneRoundData.GameID = uint64(time.Now().Unix())
	oneRoundData.UID = uid
	oneRoundData.Season = season
	oneRoundData.StartUnix = time.Now().Unix()
	oneRoundData.EndUnix = time.Now().Unix()
	oneRoundData.StartTime = time.Unix(oneRoundData.StartUnix, 0).Format("2006-01-02 15:04:05")
	oneRoundData.EndTime = time.Unix(oneRoundData.EndUnix, 0).Format("2006-01-02 15:04:05")
	oneRoundData.UserName = tmpUserName
	oneRoundData.SoloKillRating = careerdata.SoloKillRating
	oneRoundData.SoloWinRating = careerdata.SoloWinRating
	oneRoundData.SoloRating = careerdata.SoloRating
	oneRoundData.DuoKillRating = careerdata.DuoKillRating
	oneRoundData.DuoWinRating = careerdata.DuoWinRating
	oneRoundData.DuoRating = careerdata.DuoRating
	oneRoundData.SquadKillRating = careerdata.SquadKillRating
	oneRoundData.SquadWinRating = careerdata.SquadWinRating
	oneRoundData.SquadRating = careerdata.SquadRating
	oneRoundData.TotalRating = careerdata.TotalRating
	oneRoundData.TotalBattleNum = careerdata.TotalBattleNum
	oneRoundData.TotalFirstNum = careerdata.FirstNum
	oneRoundData.TotalTopTenNum = careerdata.TopTenNum
	oneRoundData.TotalHeadShot = careerdata.TotalHeadShot
	oneRoundData.TotalKillNum = careerdata.TotalKillNum
	oneRoundData.TotalShotNum = careerdata.Totalshotnum
	oneRoundData.TotalEffectHarm = careerdata.TotalEffectHarm
	oneRoundData.TotalSurviveTime = careerdata.SurviveTime
	oneRoundData.TotalDistance = careerdata.TotalDistance
	oneRoundData.TotalRecvItemUseNum = careerdata.RecvItemUseNum
	oneRoundData.SingleMaxKill = careerdata.SingleMaxKill
	oneRoundData.SingleMaxHeadShot = careerdata.SingleMaxHeadShot

	data, err := json.Marshal(oneRoundData)
	if err != nil {
		log.Error(err)
		return ErrAqDoBanRankReqFailed, err
	}

	dataCenterInnerAddr, err := db.GetDataCenterAddr("DataCenterInnerAddr")
	if err != nil {
		log.Error(err)
		return ErrAqDoBanRankReqFailed, err
	}

	resp, err := http.Post("http://"+dataCenterInnerAddr+"/dataCenter", "application/json", strings.NewReader(string(data)))
	if err != nil {
		log.Error("post data 失败:", err)
		return ErrAqDoBanRankReqFailed, err
	}
	log.Info("dataCenterInnerAddr:", dataCenterInnerAddr, " post data 成功")

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		return ErrAqDoBanRankReqFailed, err
	}

	log.Info(string(body))

	return 0, nil
}

// AqDoBanRankRsp 禁止参与排行榜接口(AQ)应答 4130
type AqDoBanRankRsp struct {
	Result int32  `json:"Result"` // 结果
	RetMsg string `json:"RetMsg"` // 返回消息
}

// GetCmdID 获取命令ID
func (rsp *AqDoBanRankRsp) GetCmdID() uint32 {
	return 4130
}

package idip

import (
	"common"
	"db"
	"fmt"
	"net/url"
	"strconv"
	"zeus/dbservice"

	log "github.com/cihub/seelog"
	"github.com/garyburd/redigo/redis"
)

// QueryRoleInfoReq 查询角色信息请求 4109
type QueryRoleInfoReq struct {
	AreaID uint32 `json:"AreaId"` // 所在大区ID
	PlatID uint8  `json:"PlatId"` // 平台
	OpenID string `json:"OpenId"` // 用户OpenId
	RoleID string `json:"RoleId"` // 角色ID
	PageNo uint8  `json:"PageNo"` // 页码
}

// GetCmdID 获取命令ID
func (req *QueryRoleInfoReq) GetCmdID() uint32 {
	return 4109
}

// GetAreaID 获取AreaID
func (req *QueryRoleInfoReq) GetAreaID() uint32 {
	return req.AreaID
}

// Do 执行查询请求
func (req *QueryRoleInfoReq) Do() (interface{}, int32, error) {
	ack := &QueryRoleInfoRsp{}
	ack.TotalPageNo = 1
	ack.TotalCount = 1
	ack.RoleInfoCount = 1
	ack.RoleInfo = make([]*RoleInfo, 0, 1)

	info := &RoleInfo{}
	var err error
	if req.RoleID != "" && req.RoleID != "0" {
		err = info.InitByUID(req.RoleID)
	} else if req.OpenID != "" {
		err = info.InitByUserName(req.OpenID)
	} else {
		err = fmt.Errorf("RoleID and OpenID nil")
	}
	if err != nil {
		if err == ErrPlayerUnknown {
			return nil, 1, ErrPlayerUnknown
		}
		return nil, ErrQueryRoleInfoFailed, err
	}
	ack.RoleInfo = append(ack.RoleInfo, info)
	return ack, 0, nil
}

/* --------------------------------------------------------------------------*/

// QueryRoleInfoRsp 查询角色信息应答
type QueryRoleInfoRsp struct {
	TotalPageNo   uint8       `json:"TotalPageNo"`    // 总页码
	TotalCount    uint32      `json:"TotalCount"`     // 总数量
	RoleInfoCount uint32      `json:"RoleInfo_count"` // 角色信息的最大数量
	RoleInfo      []*RoleInfo `json:"RoleInfo"`       // 角色信息
}

// GetCmdID 获取命令ID
func (rsp *QueryRoleInfoRsp) GetCmdID() uint32 {
	return 4110
}

/* --------------------------------------------------------------------------*/

// RoleInfo 角色信息
type RoleInfo struct {
	RoleID          string `json:"RoleId"`           // 角色ID
	RoleName        string `json:"RoleName"`         // 角色名称
	Gender          string `json:"Gender"`           // 性别
	Job             string `json:"Job"`              // 职业
	RoleType        uint32 `json:"RoleType"`         // 角色类型
	Exp             uint32 `json:"Exp"`              // 经验值
	Fight           uint32 `json:"Fight"`            // 战斗力
	JewelNum        uint32 `json:"JewelNum"`         // 钻石存量
	GoldNum         uint32 `json:"GoldNum"`          // 金币存量
	AddRecharge     uint32 `json:"AddRecharge"`      // 累计充值
	JewelConsume    uint32 `json:"JewelConsume"`     // 钻石消耗累积
	RegisterTime    uint32 `json:"RegisterTime"`     // 注册时间
	BanTime         uint32 `json:"BanTime"`          // 封号时长
	AccTimes        uint32 `json:"AccTimes"`         // 累计登录时长
	LastLogoutTime  uint32 `json:"LastLoginOutTime"` // 最后登出时间
	IsOnline        uint32 `json:"IsOnline"`         // 当前是否在线（0是 1 否）
	JoinGameNum     uint32 `json:"JoinGameNum"`      // 参与比赛总场次
	EatChicken      uint32 `json:"EatChicken"`       // 吃鸡场次
	ComEvaluation   string `json:"ComEvaluation"`    // 综合评分
	GetTopTenNum    uint32 `json:"GetTopTenNum"`     // 获得前10名场次
	SessionAverTime string `json:"SessionAverTime"`  // 场均存活时间
	SessionAdvDis   string `json:"SessionAdvDis"`    // 场均前进距离
	SessionHarm     string `json:"SessionHarm"`      // 场均伤害量
	OneRating       string `json:"OneRating"`        // 单人rating
	TwoRating       string `json:"TwoRating"`        // 双排rating
	FourRating      string `json:"FourRating"`       // 四排rating
	ComScoreRank    uint32 `json:"ComScoreRank"`     // 综合分排行
}

// InitByUID 根据UID获取信息
func (info *RoleInfo) InitByUID(uid string) error {
	if uid == "0" {
		return ErrPlayerUnknown
	}

	info.RoleID = uid

	// 从用户属性信息中(Player表)获取Name, OnlineTime, LogoutTime三个字段
	args := []interface{}{
		"Name",
		"OnlineTime",
		"LogoutTime",
		"Coin",
	}
	uid64, _ := strconv.ParseUint(uid, 10, 64)
	values, err := dbservice.EntityUtil("Player", uid64).GetValues(args)
	if err != nil {
		return err
	}
	name, _ := redis.String(values[0], nil)
	info.RoleName = url.QueryEscape(name)
	onlineTime, _ := redis.Int64(values[1], nil)
	info.AccTimes = uint32(onlineTime)
	logoutTime, _ := redis.Int64(values[2], nil)
	info.LastLogoutTime = uint32(logoutTime)
	goldNum, _ := redis.Int64(values[3], nil)
	info.GoldNum = uint32(goldNum)

	// 从PlayerInfo表中获取注册时间
	var regTime int64
	regTime, err = db.PlayerInfoUtil(uid64).GetRegisterTime()
	if err == nil {
		info.RegisterTime = uint32(regTime)
	} else {
		log.Error("IDIP查询角色注册信息失败 ", err)
	}

	// 根据是否存在Session表判断是否在线
	var isOnline bool
	isOnline, err = dbservice.SessionUtil(uid64).IsExisted()
	if err == nil {
		if isOnline {
			info.IsOnline = 0
		} else {
			info.IsOnline = 1
		}
	} else {
		log.Error("IDIP查询角色是否在线 ", err)
	}

	// 未处理的字段, TODO
	info.Gender = "unknown"
	info.Job = "unknown"
	// info.RoleType = 0
	// info.Exp = 0
	// info.Fight = 0
	// info.JewelNum = 0
	// info.AddRecharge = 0
	// info.JewelConsume = 0
	// info.BanTime = 0

	season := common.GetSeason()
	careerdata, err := common.GetPlayerCareerData(uid64, season, 0)
	if err != nil {
		return err
	}

	info.JoinGameNum = careerdata.TotalBattleNum                                          // 参与比赛总场次
	info.EatChicken = careerdata.FirstNum                                                 // 吃鸡场次
	info.ComEvaluation = strconv.FormatFloat(float64(careerdata.TotalRating), 'f', 2, 64) // 综合评分
	info.GetTopTenNum = careerdata.TopTenNum                                              // 获得前10名场次
	if careerdata.TotalBattleNum != 0 {
		info.SessionAverTime = strconv.FormatFloat(float64(careerdata.SurviveTime/int64(careerdata.TotalBattleNum)), 'f', 2, 64)   // 场均存活时间
		info.SessionAdvDis = strconv.FormatFloat(float64(careerdata.TotalDistance/float32(careerdata.TotalBattleNum)), 'f', 2, 64) // 场均前进距离
		info.SessionHarm = strconv.FormatFloat(float64(careerdata.TotalEffectHarm/careerdata.TotalBattleNum), 'f', 2, 64)          // 场均伤害量
	} else {
		info.SessionAverTime = "0" // 场均存活时间
		info.SessionAdvDis = "0"   // 场均前进距离
		info.SessionHarm = "0"     // 场均伤害量
	}
	info.OneRating = strconv.FormatFloat(float64(careerdata.SoloRating), 'f', 2, 64)   // 单人rating
	info.TwoRating = strconv.FormatFloat(float64(careerdata.DuoRating), 'f', 2, 64)    // 双排rating
	info.FourRating = strconv.FormatFloat(float64(careerdata.SquadRating), 'f', 2, 64) // 四排rating
	info.ComScoreRank = careerdata.TotalRank                                           // 综合分排行
	return nil
}

// InitByUserName 根据账号名获取信息
func (info *RoleInfo) InitByUserName(username string) error {
	uid, err := dbservice.GetUID(username)
	if uid == 0 {
		return ErrPlayerUnknown
	} else if err != nil {
		return err
	}

	return info.InitByUID(strconv.FormatUint(uid, 10))
}

/* --------------------------------------------------------------------------*/

// QueryOpenidViaRolenameInfoReq 昵称反查账号请求 4117
type QueryOpenidViaRolenameInfoReq struct {
	AreaID   uint32 `json:"AreaId"`   // 服务器：微信（1），手Q（2）
	PlatID   uint8  `json:"PlatId"`   // 平台：IOS（0），安卓（1）
	RoleName string `json:"RoleName"` // 角色名称
}

// GetCmdID 获取命令ID
func (req *QueryOpenidViaRolenameInfoReq) GetCmdID() uint32 {
	return 4117
}

// GetAreaID 获取AreaID
func (req *QueryOpenidViaRolenameInfoReq) GetAreaID() uint32 {
	return req.AreaID
}

// Do 执行请求
func (req *QueryOpenidViaRolenameInfoReq) Do() (interface{}, int32, error) {
	ack := &QueryOpenidViaRolenameInfoRsp{}
	ack.AreaID = req.AreaID
	ack.PlatID = req.PlatID

	rolename, err := url.QueryUnescape(req.RoleName)
	if err != nil {
		return nil, ErrDecodeOpenidViaRolenameInfoFailed, err
	}

	ack.RoleID = db.GetIDByName(rolename)
	if ack.RoleID == 0 {
		return nil, 1, ErrPlayerUnknown
	}
	ack.OpenID, err = dbservice.Account(ack.RoleID).GetUsername()
	if err != nil {
		return nil, ErrQueryOpenidViaRolenameInfoFailed, err
	}

	ack.RoleName = req.RoleName

	args := []interface{}{
		"LoginTime",
	}
	values, err := dbservice.EntityUtil("Player", ack.RoleID).GetValues(args)
	if err != nil {
		return nil, ErrQueryOpenidViaRolenameInfoFailed, err
	}

	time, err := redis.String(values[0], nil)
	if err != nil {
		return nil, ErrQueryOpenidViaRolenameInfoFailed, err
	}
	ack.LastLoginTime = time

	return ack, 0, nil
}

/* --------------------------------------------------------------------------*/

// QueryOpenidViaRolenameInfoRsp 昵称反查账号应答 4118
type QueryOpenidViaRolenameInfoRsp struct {
	AreaID        uint32 `json:"AreaId"`        // 服务器：微信（1），手Q（2）
	PlatID        uint8  `json:"PlatId"`        // 平台：IOS（0），安卓（1）
	OpenID        string `json:"OpenId"`        // openid
	RoleID        uint64 `json:"RoleId"`        // 角色ID
	RoleName      string `json:"RoleName"`      // 角色名称
	LastLoginTime string `json:"LastLoginTime"` // 最后登陆时间 unix时间戳
}

// GetCmdID 获取命令ID
func (rsp *QueryOpenidViaRolenameInfoRsp) GetCmdID() uint32 {
	return 4118
}

/* --------------------------------------------------------------------------*/

// QueryActiveUsrReq 查询帐号白名单权限请求 4123
type QueryActiveUsrReq struct {
	AreaID uint32 `json:"AreaId"` // 所在大区ID
	PlatID uint8  `json:"PlatId"` // 平台：IOS（0），安卓（1）
	OpenID string `json:"OpenId"` // 用户OpenId
	Source uint32 `json:"Source"` // 渠道号，由前端生成
	Serial string `json:"Serial"` // 流水号
}

// GetCmdID 获取命令ID
func (req *QueryActiveUsrReq) GetCmdID() uint32 {
	return 4123
}

// GetAreaID 获取AreaID
func (req *QueryActiveUsrReq) GetAreaID() uint32 {
	return req.AreaID
}

// Do 查询帐号白名单权限请求
func (req *QueryActiveUsrReq) Do() (interface{}, int32, error) {
	uid, _ := dbservice.GetUID(req.OpenID)
	if uid == 0 {
		return &QueryActiveUsrRsp{1, "fail", 0}, 0, ErrPlayerUnknown
	}

	userGrade, err := dbservice.Account(uid).GetGrade()
	if err != nil {
		return nil, ErrQueryActiveUsrFailed, err
	}

	return &QueryActiveUsrRsp{0, "success", userGrade}, 0, nil
}

// QueryActiveUsrRsp 查询帐号白名单权限应答 4124
type QueryActiveUsrRsp struct {
	Result int32  `json:"Result"` // 结果: 0 成功；1 帐号不存在；其它失败
	RetMsg string `json:"RetMsg"` // 返回消息
	Grade  uint32 `json:"Grade"`  //级别：1表示内部 2表示外部玩家, 0 账号不存在
}

// GetCmdID 获取命令ID
func (rsp *QueryActiveUsrRsp) GetCmdID() uint32 {
	return 4124
}

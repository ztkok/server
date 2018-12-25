package idip

import (
	"common"
	"db"
	"fmt"
	"net/url"
	"time"
	"zeus/dbservice"
	"zeus/entity"
	//log "github.com/cihub/seelog"
)

/*

 IDIP Ban-Unban 封停解封相关

*/

// DoBanUsrReq 封停角色请求 4111
type DoBanUsrReq struct {
	AreaID    uint32 `json:"AreaId"`    // 所在大区ID
	PlatID    uint8  `json:"PlatId"`    // 平台
	OpenID    string `json:"OpenId"`    // 用户OpenId
	BanTerm   int32  `json:"BanTerm"`   // XX秒，-1表示永久
	BanReason string `json:"BanReason"` // 封号原因
	Source    uint32 `json:"Source"`    // 渠道号，由前端生成，不需填写
	Serial    string `json:"Serial"`    // 流水号，由前端生成，不需要填写
}

// GetCmdID 获取命令ID
func (req *DoBanUsrReq) GetCmdID() uint32 {
	return 4111
}

// GetAreaID 获取AreaID
func (req *DoBanUsrReq) GetAreaID() uint32 {
	return req.AreaID
}

// Do 执行封停角色请求
func (req *DoBanUsrReq) Do() (interface{}, int32, error) {

	targetID, err := dbservice.GetUID(req.OpenID)
	if err != nil || targetID == 0 {
		return nil, 1, ErrPlayerUnknown
	}

	reason, err := url.QueryUnescape(req.BanReason)
	if err != nil {
		return nil, ErrDecodeNoticeContentFailed, err
	}

	data := &db.BanAccount{
		Uid:         targetID,
		OpenID:      req.OpenID,
		BanDuration: req.BanTerm,
		BanReason:   reason,
	}

	if data.BanDuration > 0 {
		data.EndTime = uint32(int32(time.Now().Unix()) + req.BanTerm)
	}

	if !db.AddBanAccount(data) {
		return nil, ErrBanAccountFailed, fmt.Errorf("data error")
	}

	return &DoBanUsrRsp{Result: 0, RetMsg: "ban user success"}, 0, nil
}

// DoBanUsrRsp 封停角色应答 4112
type DoBanUsrRsp struct {
	Result int32  `json:"Result"` // 结果：成功(0)，玩家不存在(1)，失败(其他)
	RetMsg string `json:"RetMsg"` // 返回消息
}

// GetCmdID 获取命令ID
func (rsp *DoBanUsrRsp) GetCmdID() uint32 {
	return 4112
}

/* --------------------------------------------------------------------------*/

// DoUnbanUsrReq 解封角色 4113
type DoUnbanUsrReq struct {
	AreaID uint32 `json:"AreaId"` // 所在大区ID
	PlatID uint8  `json:"PlatId"` // 平台
	OpenID string `json:"OpenId"` // 用户OpenId
	Source uint32 `json:"Source"` // 渠道号，由前端生成，不需填写
	Serial string `json:"Serial"` // 流水号，由前端生成，不需要填写
}

// GetCmdID 获取命令ID
func (req *DoUnbanUsrReq) GetCmdID() uint32 {
	return 4113
}

// GetAreaID 获取AreaID
func (req *DoUnbanUsrReq) GetAreaID() uint32 {
	return req.AreaID
}

// Do 执行解封角色请求
func (req *DoUnbanUsrReq) Do() (interface{}, int32, error) {

	targetID, err := dbservice.GetUID(req.OpenID)
	if err != nil || targetID == 0 {
		return nil, 1, ErrPlayerUnknown
	}

	if db.GetBanAccountData(req.OpenID) == nil {
		return &DoUnbanUsrRsp{Result: 0, RetMsg: "Account not banned"}, 0, nil
	}

	db.UnbanAccount(req.OpenID)

	return &DoUnbanUsrRsp{Result: 0, RetMsg: "unban user success"}, 0, nil
}

// DoUnbanUsrRsp 解封角色应答 4114
type DoUnbanUsrRsp struct {
	Result int32  `json:"Result"` // 结果：成功(0)，玩家不存在(1)，失败(其他)
	RetMsg string `json:"RetMsg"` // 返回消息
}

// GetCmdID 获取命令ID
func (rsp *DoUnbanUsrRsp) GetCmdID() uint32 {
	return 4114
}

/* --------------------------------------------------------------------------*/

// 查询角色封停状态 5030
type QueryRoleBanStateReq struct {
	AreaID uint32 `json:"AreaId"` // 所在大区ID
	PlatID uint8  `json:"PlatId"` // 平台
	OpenID string `json:"OpenId"` // 用户OpenId
	RoleID string `json:"RoleId"` // 角色ID
}

// GetCmdID 获取命令ID
func (req *QueryRoleBanStateReq) GetCmdID() uint32 {
	return 5030
}

// GetAreaID 获取AreaID
func (req *QueryRoleBanStateReq) GetAreaID() uint32 {
	return req.AreaID
}

// Do 执行查询角色封停状态请求
func (req *QueryRoleBanStateReq) Do() (interface{}, int32, error) {

	uidByOpenID, err := dbservice.GetUID(req.OpenID)
	if err != nil || uidByOpenID == 0 {
		return nil, ErrQueryBanRoleFailed, fmt.Errorf("账号不存在")
	}

	uidByRoleID := db.GetIDByName(req.RoleID)
	if uidByRoleID == 0 {
		return nil, ErrQueryBanRoleFailed, fmt.Errorf("角色不存在")
	}

	if uidByOpenID != uidByRoleID {
		return nil, ErrQueryBanRoleFailed, fmt.Errorf("查询账号下没有该角色！")
	}

	ack := &DoQueryRoleBanStateRsp{}
	ack.SBanRoleRsp = make([]BanRoleRsp, 0)

	item := BanRoleRsp{
		OpenID: req.OpenID,
		RoleID: req.RoleID,
	}
	data := db.GetBanRoleData(req.RoleID)
	if data == nil {
		item.State = 0
	} else {
		item.State = 1
		item.BanDuration = data.BanDuration
		item.EndTime = data.EndTime
		item.BanReason = data.BanReason
	}
	ack.SBanRoleRsp = append(ack.SBanRoleRsp, item)

	return ack, 0, nil
}

// BanRoleRspData 封停角色应答数据
type BanRoleRsp struct {
	OpenID      string `json:"OpenID"`
	RoleID      string `json:"RoleID"`
	State       uint32 `json:"State"`       // 封号状态 0 未封号 1 封号
	BanDuration int32  `json:"BanDuration"` // 封停时长，-1表示永久封号
	EndTime     uint32 `json:"EndTime"`     // 封号截止日期
	BanReason   string `json:"BanReason"`   // 封号原因：（自定义文字，玩家登录时客户端可见）
}

// DoBanRoleRsp 查询角色封停状态应答 5031
type DoQueryRoleBanStateRsp struct {
	SBanRoleRsp []BanRoleRsp `json:"BanRoleRsp"` // 封停角色列表
}

// GetCmdID 获取命令ID
func (rsp *DoQueryRoleBanStateRsp) GetCmdID() uint32 {
	return 5031
}

/* --------------------------------------------------------------------------*/

// 封停账号 5040
type BanAccountReq struct {
	AreaID      uint32 `json:"AreaId"`      // 所在大区ID
	PlatID      uint8  `json:"PlatId"`      // 平台
	OpenID      string `json:"OpenId"`      // 用户OpenId
	BanDuration int32  `json:"BanDuration"` // 封停时长，-1表示永久封号
	BanReason   string `json:"BanReason"`   // 封号原因：（自定义文字，玩家登录时客户端可见）
}

// GetCmdID 获取命令ID
func (req *BanAccountReq) GetCmdID() uint32 {
	return 5040
}

// GetAreaID 获取AreaID
func (req *BanAccountReq) GetAreaID() uint32 {
	return req.AreaID
}

// Do 执行封停账号请求
func (req *BanAccountReq) Do() (interface{}, int32, error) {

	targetID, err := dbservice.GetUID(req.OpenID)
	if err != nil || targetID == 0 {
		return nil, ErrBanAccountFailed, fmt.Errorf("账号不存在")
	}

	data := &db.BanAccount{
		Uid:         targetID,
		OpenID:      req.OpenID,
		BanDuration: req.BanDuration,
		BanReason:   req.BanReason,
	}

	if data.BanDuration > 0 {
		data.EndTime = uint32(int32(time.Now().Unix()) + req.BanDuration)
	}

	if !db.AddBanAccount(data) {
		return nil, ErrBanAccountFailed, fmt.Errorf("请求执行封账号数据错误!")
	}

	return &DoBanAccountRsp{Result: 0, RetMsg: "封停账号成功!"}, 0, nil

}

// DoBanRoleRsp 封停账号应答 5041
type DoBanAccountRsp struct {
	Result int32  `json:"Result"` // 结果：成功(0)，玩家不存在(1)，失败(其他)
	RetMsg string `json:"RetMsg"` // 返回消息
}

// GetCmdID 获取命令ID
func (rsp *DoBanAccountRsp) GetCmdID() uint32 {
	return 5041
}

/* --------------------------------------------------------------------------*/

// 解封账号 5050
type UnbanAccountReq struct {
	AreaID uint32 `json:"AreaId"` // 所在大区ID
	PlatID uint8  `json:"PlatId"` // 平台
	OpenID string `json:"OpenId"` // 用户OpenId
	RoleID string `json:"RoleId"` // 角色ID
}

// GetCmdID 获取命令ID
func (req *UnbanAccountReq) GetCmdID() uint32 {
	return 5050
}

// GetAreaID 获取AreaID
func (req *UnbanAccountReq) GetAreaID() uint32 {
	return req.AreaID
}

// Do 执行解封账号请求
func (req *UnbanAccountReq) Do() (interface{}, int32, error) {

	targetID, err := dbservice.GetUID(req.OpenID)
	if err != nil || targetID == 0 {
		return nil, ErrUnbanAccountFailed, fmt.Errorf("解封账号不存在!")
	}

	if db.GetBanAccountData(req.OpenID) == nil {
		return nil, ErrUnbanAccountFailed, fmt.Errorf("账号没有被封禁!")
	}

	db.UnbanAccount(req.OpenID)

	return &DoUnbanAccountRsp{Result: 0, RetMsg: "解封账号成功!"}, 0, nil

}

// DoBanAccountRsp 解封账号应答 5051
type DoUnbanAccountRsp struct {
	Result int32  `json:"Result"` // 结果：成功(0)，玩家不存在(1)，失败(其他)
	RetMsg string `json:"RetMsg"` // 返回消息
}

// GetCmdID 获取命令ID
func (rsp *DoUnbanAccountRsp) GetCmdID() uint32 {
	return 5051
}

/* --------------------------------------------------------------------------*/

// 查询账号封停状态 5060
type QueryAccountBanStateReq struct {
	AreaID uint32 `json:"AreaId"` // 所在大区ID
	PlatID uint8  `json:"PlatId"` // 平台
	OpenID string `json:"OpenId"` // 用户OpenId
}

// GetCmdID 获取命令ID
func (req *QueryAccountBanStateReq) GetCmdID() uint32 {
	return 5060
}

// GetAreaID 获取AreaID
func (req *QueryAccountBanStateReq) GetAreaID() uint32 {
	return req.AreaID
}

// Do 执行查询账号封停状态请求
func (req *QueryAccountBanStateReq) Do() (interface{}, int32, error) {

	targetID, err := dbservice.GetUID(req.OpenID)
	if err != nil || targetID == 0 {
		return nil, ErrQueryBanAccountFailed, fmt.Errorf("角色不存在")
	}

	ack := &DoQueryAccountBanStateRsp{}
	ack.SBanAccountRsp = make([]BanAccountRsp, 0)

	item := BanAccountRsp{
		OpenID: req.OpenID,
	}
	data := db.GetBanAccountData(req.OpenID)
	if data == nil {
		item.State = 0
	} else {
		item.State = 1
		item.BanDuration = data.BanDuration
		item.EndTime = data.EndTime
		item.BanReason = data.BanReason
	}
	ack.SBanAccountRsp = append(ack.SBanAccountRsp, item)

	return ack, 0, nil
}

// BanAccountRsp 封停账号应答数据
type BanAccountRsp struct {
	OpenID      string `json:"OpenID"`
	State       uint32 `json:"State"`       // 封号状态 0 未封号 1 封号
	BanDuration int32  `json:"BanDuration"` // 封停时长，-1表示永久封号
	EndTime     uint32 `json:"EndTime"`     // 封号截止日期
	BanReason   string `json:"BanReason"`   // 封号原因：（自定义文字，玩家登录时客户端可见）
}

// DoBanRoleRsp 查询账号封停状态应答 5061
type DoQueryAccountBanStateRsp struct {
	SBanAccountRsp []BanAccountRsp `json:"BanAccountRsp"` // 封停账号列表
}

// GetCmdID 获取命令ID
func (rsp *DoQueryAccountBanStateRsp) GetCmdID() uint32 {
	return 5061
}

/* --------------------------------------------------------------------------*/

// DoKickingPlayerReq 踢出角色下线 4125
type DoKickingPlayerReq struct {
	AreaID uint32 `json:"AreaId"` // 所在大区ID
	PlatID uint8  `json:"PlatId"` // 平台
	OpenID string `json:"OpenId"` // 用户OpenId
	RoleID string `json:"RoleID"` // RoleID
	Source uint32 `json:"Source"` // 渠道号，由前端生成
	Serial string `json:"Serial"` // 流水号
}

// GetCmdID 获取命令ID
func (req *DoKickingPlayerReq) GetCmdID() uint32 {
	return 4125
}

// GetAreaID 获取AreaID
func (req *DoKickingPlayerReq) GetAreaID() uint32 {
	return req.AreaID
}

// Do 执行踢出角色下线请求
func (req *DoKickingPlayerReq) Do() (interface{}, int32, error) {
	return kickPlayer(req.OpenID, "你被踢下线了")
}

// DoKickingPlayerRsp 踢出角色下线应答 4126
type DoKickingPlayerRsp struct {
	Result int32  `json:"Result"` // 结果：成功(0)，玩家不在线(1)，失败(其他)
	RetMsg string `json:"RetMsg"` // 返回消息
}

// GetCmdID 获取命令ID
func (rsp *DoKickingPlayerRsp) GetCmdID() uint32 {
	return 4126
}

/***************************************************************************/

// kickPlayer 踢角色下线
func kickPlayer(openid string, banAccountReason string) (interface{}, int32, error) {
	uid, err := dbservice.GetUID(openid)
	if err != nil {
		return nil, ErrKickingPlayerFailed, err
	}

	entityID, err := dbservice.SessionUtil(uid).GetUserEntityID()
	if err != nil {
		return &DoKickingPlayerRsp{1, "Player Offline"}, 0, nil
	}

	srvLobbyID, spaceID, err := dbservice.EntitySrvUtil(entityID).GetSrvInfo(common.ServerTypeLobby)
	if srvLobbyID != 0 && err == nil {
		proxy := entity.NewEntityProxy(srvLobbyID, spaceID, entityID)
		if err := proxy.RPC(common.ServerTypeLobby, "KickingPlayer", banAccountReason); err != nil {
			return nil, ErrKickingPlayerFailed, err
		}
	}

	srvRoomID, spaceID, err := dbservice.EntitySrvUtil(entityID).GetSrvInfo(common.ServerTypeRoom)
	if srvRoomID != 0 && err == nil {
		proxy := entity.NewEntityProxy(srvRoomID, spaceID, entityID)
		if err := proxy.RPC(common.ServerTypeRoom, "KickingPlayer", banAccountReason); err != nil {
			return nil, ErrKickingPlayerFailed, err
		}
	}

	return &DoKickingPlayerRsp{0, "success"}, 0, nil
}

/**************************************************************************************************
*********************************************(安全idip)********************************************
**************************************************************************************************/

// AqDoBanAccountReq (安全idip)封号接口请求 4131
type AqDoBanAccountReq struct {
	AreaID           uint32 `json:"AreaId"`           // 所在大区ID
	PlatID           uint8  `json:"PlatId"`           // 平台
	OpenID           string `json:"OpenId"`           // 用户OpenId
	BanTime          int32  `json:"BanTime"`          // XX秒，-1表示永久
	BanAccountReason string `json:"BanAccountReason"` // 封号原因
	Source           uint32 `json:"Source"`           // 渠道号，由前端生成，不需填写
	Serial           string `json:"Serial"`           // 流水号，由前端生成，不需要填写
}

// GetCmdID 获取命令ID
func (req *AqDoBanAccountReq) GetCmdID() uint32 {
	return 4131
}

// GetAreaID 获取AreaID
func (req *AqDoBanAccountReq) GetAreaID() uint32 {
	return req.AreaID
}

// Do 执行封停账号请求
func (req *AqDoBanAccountReq) Do() (interface{}, int32, error) {
	targetID, err := dbservice.GetUID(req.OpenID)
	if err != nil || targetID == 0 {
		return nil, ErrBanAccountFailed, ErrPlayerUnknown
	}

	reason, err := url.QueryUnescape(req.BanAccountReason)
	if err != nil {
		return nil, ErrDecodeNoticeContentFailed, err
	}

	data := &db.BanAccount{
		Uid:         targetID,
		OpenID:      req.OpenID,
		BanDuration: req.BanTime,
		BanReason:   reason,
	}

	if data.BanDuration > 0 {
		data.EndTime = uint32(int32(time.Now().Unix()) + req.BanTime)
	}

	if !db.AddBanAccount(data) {
		return nil, ErrBanAccountFailed, fmt.Errorf("data error")
	}

	// 踢角色下线
	_, _, err = kickPlayer(req.OpenID, req.BanAccountReason)
	if err != nil {
		return nil, ErrDecodeNoticeContentFailed, err
	}

	return &DoBanUsrRsp{Result: 0, RetMsg: "success"}, 0, nil
}

// AqDoBanAccountRsp 封停账号应答 4132
type AqDoBanAccountRsp struct {
	Result int32  `json:"Result"` // 结果：成功(0)，玩家不存在(1)，失败(其他)
	RetMsg string `json:"RetMsg"` // 返回消息
}

// GetCmdID 获取命令ID
func (rsp *AqDoBanAccountRsp) GetCmdID() uint32 {
	return 4132
}

/* --------------------------------------------------------------------------*/

// AqDoRelievePublishReq (安全idip)解除处罚接口 4133
type AqDoRelievePublishReq struct {
	AreaID            uint32 `json:"AreaId"`            // 所在大区ID
	PlatID            uint8  `json:"PlatId"`            // 平台
	OpenID            string `json:"OpenId"`            // 用户OpenId
	RoleID            string `json:"RoleId"`            // 角色ID
	RelieveAssignPlay uint8  `json:"RelieveAssignPlay"` //解除禁止指定玩法状态（0 否，1 是）
	RelieveRackBan    uint8  `json:"RelieveRackBan"`    //解除禁止排行榜限制（0 否，1 是）
	RelieveAccountBan uint8  `json:"RelieveAccountBan"` //解除封号（0 否，1 是）
	RelieveChatBan    uint8  `json:"RelieveChatBan"`    //解除禁言（0 否，1 是）
	RelieveZeroProfit uint8  `json:"RelieveZeroProfit"` //解除零收益（0 否，1 是）
	Source            uint32 `json:"Source"`            //渠道号，由前端生成，不需填写
	Serial            string `json:"Serial"`            //流水号，由前端生成，不需要填写
}

// GetCmdID 获取命令ID
func (req *AqDoRelievePublishReq) GetCmdID() uint32 {
	return 4133
}

// GetAreaID 获取AreaID
func (req *AqDoRelievePublishReq) GetAreaID() uint32 {
	return req.AreaID
}

// Do 执行解封账号请求
func (req *AqDoRelievePublishReq) Do() (interface{}, int32, error) {

	targetID, err := dbservice.GetUID(req.OpenID)
	if err != nil || targetID == 0 {
		return nil, ErrUnbanAccountFailed, fmt.Errorf("解封账号不存在!")
	}

	if db.GetBanAccountData(req.OpenID) == nil {
		return nil, ErrUnbanAccountFailed, fmt.Errorf("账号没有被封禁!")
	}

	db.UnbanAccount(req.OpenID)

	return &AqDoRelievePublishRsp{Result: 0, RetMsg: "success"}, 0, nil
}

// AqDoRelievePublishRsp 解除处罚接口应答 4134
type AqDoRelievePublishRsp struct {
	Result int32  `json:"Result"` // 结果：成功(0)，玩家不存在(1)，失败(其他)
	RetMsg string `json:"RetMsg"` // 返回消息
}

// GetCmdID 获取命令ID
func (rsp *AqDoRelievePublishRsp) GetCmdID() uint32 {
	return 4134
}

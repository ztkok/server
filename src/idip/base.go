package idip

import (
	"encoding/json"
	"errors"
	"fmt"
)

/*

 IDIP请求基础数据结构

*/

// IDo 命令执行接口
type IDo interface {
	Do() (interface{}, int32, error)
}

// IAreaIDGetter 获取命令AreaID
type IAreaIDGetter interface {
	GetAreaID() uint32
}

// ICmdIDGetter 获取命令ID
type ICmdIDGetter interface {
	GetCmdID() uint32
}

// ErrPlayerUnknown 玩家不存在时返回统一错误
var ErrPlayerUnknown = errors.New("Player Unknown")

const (
	// ErrReadBodyFailed 读Body失败
	ErrReadBodyFailed = -101

	// ErrJSONDecodeFailed 解析json出错
	ErrJSONDecodeFailed = -102

	// ErrUnsupportCmd 不支持的命令
	ErrUnsupportCmd = -103

	// ErrDoCmdFailed 执行命令失败
	ErrDoCmdFailed = -104

	// ErrTransportFailed 转发请求失败
	ErrTransportFailed = -105

	// ErrQueryRoleInfoFailed 查询角色信息失败
	ErrQueryRoleInfoFailed = -110

	// ErrQueryOpenidViaRolenameInfoFailed 昵称反查账号请求失败
	ErrQueryOpenidViaRolenameInfoFailed = -111

	// ErrDecodeOpenidViaRolenameInfoFailed 解码失败
	ErrDecodeOpenidViaRolenameInfoFailed = -112

	// ErrActiveFailed 激活帐号失败
	ErrActiveFailed = -120

	// ErrSendAnnuonce 发送走马灯失败
	ErrSendAnnuonce = -130

	// ErrDelAnnuonce 删除走马灯失败
	ErrDelAnnuonce = -131

	// ErrDecodeNoticeContentFailed 解码失败
	ErrDecodeNoticeContentFailed = -132

	// ErrBanRoleFailed 封停角色失败
	ErrBanRoleFailed = -141

	// ErrUnbanRoleFailed 解封角色失败
	ErrUnbanRoleFailed = -142

	// ErrQueryBanRoleFailed 查询封停角色失败
	ErrQueryBanRoleFailed = -143

	// ErrBanAccountFailed 封停账号失败
	ErrBanAccountFailed = -144

	// ErrUnbanAccountFailed 解封账号失败
	ErrUnbanAccountFailed = -145

	// ErrQueryBanAccountFailed 查询封停账号失败
	ErrQueryBanAccountFailed = -146

	// ErrDecodeMailTitleFailed 解码失败
	ErrDecodeMailTitleFailed = -150

	// ErrDecodeMailContentFailed 解码失败
	ErrDecodeMailContentFailed = -151

	// ErrDecodeMailLineFailed 解码失败
	ErrDecodeMailLineFailed = -152

	// ErrDeocdeMailButtonConFailed 解码失败
	ErrDeocdeMailButtonConFailed = -153

	// ErrAqDoBanRankReqFailed 禁止参与排行榜失败
	ErrAqDoBanRankReqFailed = -154

	// ErrKickingPlayerFailed 踢出角色下线失败
	ErrKickingPlayerFailed = -155

	// ErrQueryActiveUsrFailed 查询帐号白名单权限失败
	ErrQueryActiveUsrFailed = -156
)

// Header IDIP消息头
type Header struct {
	PacketLen    uint32 `json:"PacketLen"`    // 包长
	CmdID        uint32 `json:"Cmdid"`        // 命令ID
	SeqID        uint32 `json:"Seqid"`        // 流水号
	ServiceName  string `json:"ServiceName"`  // 服务名
	SendTime     uint32 `json:"SendTime"`     // 发送时间YYYYMMDD对应的整数
	Version      uint32 `json:"Version"`      // 版本号
	Authenticate string `json:"Authenticate"` // 加密串
	Result       int32  `json:"Result"`       // 错误码,返回码类型：0：处理成功，需要解开包体获得详细信息,1：处理成功，但包体返回为空，不需要处理包体（eg：查询用户角色，用户角色不存在等），-1: 网络通信异常,-2：超时,-3：数据库操作异常,-4：API返回异常,-5：服务器忙,-6：其他错误,小于-100 ：用户自定义错误，需要填写szRetErrMsg
	RetErrMsg    string `json:"RetErrMsg"`    // 错误信息
}

// DataPaket IDIP数据包
type DataPaket struct {
	Head Header      `json:"head"` // 包头信息
	Body interface{} `json:"body"` // 包体信息
}

// RequestLog IDIP 敏感接口tlog日志结构, 比如单发邮件
type RequestLog struct {
	DtEventTime string // (必填)游戏事件的时间, 格式 YYYY-MM-DD HH:MM:SS
	AreaID      uint32 // 大区ID
	VOpenID     string // (必填)用户OPENID号
	ItemID      int    // (必填)道具id，非道具填0
	ItemNum     int    // 物品数量
	Serial      string // (必填)流水号前端传入，默认0
	Source      int    // 渠道ID, 默认0
	Cmd         int    // 指令ID
}

// Name 结构体名字
func (r *RequestLog) Name() string {
	return "IDIPRequest"
}

// RequestLogNoOpenID IDIP 敏感接口tlog日志结构, 没有OpenID的请求, 比如群发邮件
type RequestLogNoOpenID struct {
	DtEventTime string // (必填)游戏事件的时间, 格式 YYYY-MM-DD HH:MM:SS
	AreaID      uint32 // 大区ID
	ItemID      int    // (必填)道具id，非道具填0
	ItemNum     int    // 物品数量
	Serial      string // (必填)流水号前端传入，默认0
	Source      int    // 渠道ID, 默认0
	Cmd         int    // 指令ID
}

// Name 结构体名字
func (r *RequestLogNoOpenID) Name() string {
	return "IDIPRequest"
}

// Decode 解包
func Decode(data []byte) (*DataPaket, int32, string, error) {
	req := &DataPaket{}

	err := json.Unmarshal(data, req)
	if err != nil {
		return nil, ErrJSONDecodeFailed, err.Error(), err
	}

	req.Body, err = getBody(req.Head.CmdID)
	if err != nil {
		return nil, ErrUnsupportCmd, err.Error(), err
	}

	err = json.Unmarshal(data, req)
	if err != nil {
		return nil, ErrJSONDecodeFailed, err.Error(), err
	}
	return req, 0, "", nil
}

func getBody(cmd uint32) (interface{}, error) {
	switch cmd {
	case 4097: // 激活帐号 （开白名单）请求
		return &DoActiveUsrReq{}, nil
	case 4099: // 发走马灯
		return &DoSendMarqueeReq{}, nil
	case 4101: // 删除走马灯
		return &DoDeleteMarqueeReq{}, nil
	case 4103: // 群发邮件
		return &DoSendItemReq{}, nil
	case 4105: // 查询走马灯
		return &DoQueryMarqueeReq{}, nil
	case 4107:
		return &QueryMailAllInfoReq{}, nil
	case 4109: // 查询角色信息
		return &QueryRoleInfoReq{}, nil
	case 4111: // 封停角色
		return &DoBanUsrReq{}, nil
	case 4113: // 解封角色
		return &DoUnbanUsrReq{}, nil
	case 4115: // 邮件赠送物品请求
		return &DoSendItemMailReq{}, nil
	case 4117: // 昵称反查账号请求
		return &QueryOpenidViaRolenameInfoReq{}, nil
	case 5030: // 查询封停角色
		return &QueryRoleBanStateReq{}, nil
	case 5040: // 封停账号
		return &BanAccountReq{}, nil
	case 5050: // 解封封停账号
		return &UnbanAccountReq{}, nil
	case 5060: // 查询封停账号
		return &QueryAccountBanStateReq{}, nil
	case 4125: // 踢角色下线
		return &DoKickingPlayerReq{}, nil
	case 4129: // (安全idip)禁止参与排行榜
		return &AqDoBanRankReq{}, nil
	case 4123: // 查询帐号白名单权限
		return &QueryActiveUsrReq{}, nil
	case 4131: // (安全idip)封号接口请求
		return &AqDoBanAccountReq{}, nil
	case 4133: // (安全idip)解除处罚接口
		return &AqDoRelievePublishReq{}, nil
	default:
		return nil, fmt.Errorf("Unknown request %d", cmd)
	}
}

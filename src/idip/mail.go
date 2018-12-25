package idip

import (
	"db"
	"encoding/json"
	"math"
	"net/url"
	"sort"
	"time"
	"zeus/iserver"
	"zeus/serializer"
	"zeus/tlog"

	"github.com/cihub/seelog"
)

/*

 IDIP 邮件相关

*/

// SMailInfo 邮件信息对象
type SMailInfo struct {
	MailID       uint32 `json:"MailId"`       // 邮件ID
	SendTime     uint32 `json:"SendTime"`     // 发送时间
	MailTitle    string `json:"MailTitle"`    // 邮件标题
	MailContent  string `json:"MailContent"`  // 邮件内容
	MinLevel     uint16 `json:"MinLevel"`     // 最小领取等级（默认0）
	MaxLevel     uint32 `json:"MaxLevel"`     // 最大领取等级（默认0）
	ItemOneID    uint32 `json:"ItemOneId"`    // 道具ID1
	ItemOneNum   uint32 `json:"ItemOneNum"`   // 道具数量1
	ItemTwoID    uint32 `json:"ItemTwoId"`    // 道具ID2
	ItemTwoNum   uint32 `json:"ItemTwoNum"`   // 道具数量2
	ItemThreeID  uint32 `json:"ItemThreeId"`  // 道具ID3
	ItemThreeNum uint32 `json:"ItemThreeNum"` // 道具数量3
	ItemFourID   uint32 `json:"ItemFourId"`   // 道具ID4
	ItemFourNum  uint32 `json:"ItemFourNum"`  // 道具数量4
	ItemFiveID   uint32 `json:"ItemFiveId"`   // 道具ID5
	ItemFiveNum  uint32 `json:"ItemFiveNum"`  // 道具数量5
	Hyperlink    string `json:"Hyperlink"`    // 超链接
	ButtonCon    string `json:"ButtonCon"`    // 按钮内容：(可以为空、为空时则不显示该超链接的按钮。不为空时则按钮显示输入的文字、如”点击查看“按钮)
}

/* --------------------------------------------------------------------------*/

// DoSendItemReq 群发邮件请求 4103
type DoSendItemReq struct {
	AreaID        uint32     `json:"AreaId"`         // 所在大区ID
	PlatID        uint8      `json:"PlatId"`         // 平台
	MailTitle     string     `json:"MailTitle"`      // 邮件标题
	MailContent   string     `json:"MailContent"`    // 邮件内容
	LevelFloor    uint32     `json:"LevelFloor"`     // 等级下限
	LevelTop      uint32     `json:"LevelTop"`       // 等级上限
	SendTime      uint32     `json:"SendTime"`       // 发送时间
	ItemDataCount uint32     `json:"ItemData_count"` // 道具列表的最大数量
	ItemData      []ItemInfo `json:"ItemData"`       // 道具列表
	Line          string     `json:"Line"`           // 超链接
	ButtonCon     string     `json:"ButtonCon"`      // 按钮内容：(可以为空、为空时则不显示该超链接的按钮。不为空时则按钮显示输入的文字、如”点击查看“按钮)
	Source        uint32     `json:"Source"`         // 渠道号
	Serial        string     `json:"Serial"`         // 流水号
}

// GetCmdID 获取命令ID
func (req *DoSendItemReq) GetCmdID() uint32 {
	return 4103
}

// GetAreaID 获取AreaID
func (req *DoSendItemReq) GetAreaID() uint32 {
	return req.AreaID
}

// Do 执行群发邮件请求
func (req *DoSendItemReq) Do() (interface{}, int32, error) {
	ack := &DoSendItemRsp{}
	ack.Result = 0
	ack.RetMsg = "Success"

	title, err := url.QueryUnescape(req.MailTitle)
	if err != nil {
		return nil, ErrDecodeMailTitleFailed, err
	}
	content, err := url.QueryUnescape(req.MailContent)
	if err != nil {
		return nil, ErrDecodeMailContentFailed, err
	}
	line, err := url.QueryUnescape(req.Line)
	if err != nil {
		return nil, ErrDecodeMailLineFailed, err
	}
	btCon, err := url.QueryUnescape(req.ButtonCon)
	if err != nil {
		return nil, ErrDeocdeMailButtonConFailed, err
	}

	req.MailTitle = title
	req.MailContent = content
	req.Line = line
	req.ButtonCon = btCon
	data, err := json.Marshal(req)
	if err == nil {
		iserver.GetSrvInst().FireRPC("SendAllMail", serializer.Serialize(data))
	}

	log := &RequestLogNoOpenID{}
	log.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
	log.AreaID = req.AreaID
	log.Serial = req.Serial
	log.Source = int(req.Source)
	log.Cmd = int(req.GetCmdID())
	for _, v := range req.ItemData {
		log.ItemID = int(v.ItemID)
		log.ItemNum = int(v.ItemNum)
		tlog.Format(log)
		seelog.Info("发送邮件000", log)
	}
	seelog.Info("发送邮件1111", log)

	return ack, 0, nil
}

/* --------------------------------------------------------------------------*/

// DoSendItemRsp 群发邮件应答 4104
type DoSendItemRsp struct {
	Result int32  `json:"Result"` // 结果
	RetMsg string `json:"RetMsg"` // 返回消息
}

// GetCmdID 获取命令ID
func (rsp *DoSendItemRsp) GetCmdID() uint32 {
	return 4104
}

/* --------------------------------------------------------------------------*/

// QueryMailAllInfoReq 查询全服邮件请求 4107
type QueryMailAllInfoReq struct {
	AreaID    uint32 `json:"AreaId"`    // 所在大区ID
	PlatID    uint8  `json:"PlatId"`    // 平台
	BeginTime uint32 `json:"BeginTime"` // 开始时间
	EndTime   uint32 `json:"EndTime"`   // 结束时间
	PageNo    uint8  `json:"PageNo"`    // 页码
}

// GetCmdID 获取命令ID
func (req *QueryMailAllInfoReq) GetCmdID() uint32 {
	return 4107
}

// GetAreaID 获取AreaID
func (req *QueryMailAllInfoReq) GetAreaID() uint32 {
	return req.AreaID
}

// Do 执行查询全服邮件应答
func (req *QueryMailAllInfoReq) Do() (interface{}, int32, error) {
	ack := &QueryMailAllInfoRsp{}

	mails := db.SliIdipMail(db.IdipMailUtil().GetMails(req.BeginTime, req.EndTime))
	sort.Sort(mails)

	ack.TotalCount = uint32(len(mails))
	ack.TotalPageNo = uint8(math.Ceil(float64(ack.TotalCount) / float64(maxMailInfos)))

	startIndex := uint32(req.PageNo-1) * maxMailInfos
	endIndex := uint32(req.PageNo) * maxMailInfos
	if endIndex > uint32(ack.TotalCount) {
		endIndex = uint32(ack.TotalCount)
	}

	for i, d := range mails {

		if i < int(startIndex) {
			continue
		}

		if i >= int(endIndex) {
			break
		}

		tmp := SMailInfo{
			MailID:       d.MailID,
			SendTime:     d.SendTime,
			MailTitle:    url.QueryEscape(d.MailTitle),
			MailContent:  url.QueryEscape(d.MailContent),
			MinLevel:     d.MinLevel,
			MaxLevel:     d.MaxLevel,
			ItemOneID:    d.ItemOneID,
			ItemOneNum:   d.ItemOneNum,
			ItemTwoID:    d.ItemTwoID,
			ItemTwoNum:   d.ItemTwoNum,
			ItemThreeID:  d.ItemThreeID,
			ItemThreeNum: d.ItemThreeNum,
			ItemFourID:   d.ItemFourID,
			ItemFourNum:  d.ItemFourNum,
			ItemFiveID:   d.ItemFiveID,
			ItemFiveNum:  d.ItemFiveNum,
			Hyperlink:    url.QueryEscape(d.Hyperlink),
			ButtonCon:    url.QueryEscape(d.ButtonCon),
		}

		ack.MailInfo = append(ack.MailInfo, tmp)
	}

	ack.MailInfoCount = uint32(len(ack.MailInfo))

	return ack, 0, nil
}

/* --------------------------------------------------------------------------*/

const maxMailInfos = 10 // 单次应答最大邮件信息数量

// QueryMailAllInfoRsp 查询全服邮件应答 4108
type QueryMailAllInfoRsp struct {
	TotalCount    uint32      `json:"TotalCount"`     // 总数量
	TotalPageNo   uint8       `json:"TotalPageNo"`    // 总页码
	MailInfoCount uint32      `json:"MailInfo_count"` // 邮件信息列表的最大数量
	MailInfo      []SMailInfo `json:"MailInfo"`       // 邮件信息列表
}

// GetCmdID 获取命令ID
func (rsp *QueryMailAllInfoRsp) GetCmdID() uint32 {
	return 4108
}

/* --------------------------------------------------------------------------*/

// DoSendItemMailReq 邮件赠送物品请求 4115
type DoSendItemMailReq struct {
	AreaID      uint32 `json:"AreaId"`      // 所在大区ID
	PlatID      uint8  `json:"PlatId"`      // 平台
	OpenID      string `json:"OpenId"`      // 用户OpenId
	RoleID      uint64 `json:"RoleId"`      // 角色ID
	MailTitle   string `json:"MailTitle"`   // 邮件标题
	MailContent string `json:"MailContent"` // 邮件内容
	ItemID      uint64 `json:"ItemId"`      // 道具ID
	ItemNum     uint32 `json:"ItemNum"`     // 道具数量
	Source      uint32 `json:"Source"`      // 渠道号
	Serial      string `json:"Serial"`      // 流水号
}

// GetCmdID 获取命令ID
func (req *DoSendItemMailReq) GetCmdID() uint32 {
	return 4115
}

// GetAreaID 获取AreaID
func (req *DoSendItemMailReq) GetAreaID() uint32 {
	return req.AreaID
}

// Do 给单独玩家发邮件请求执行
func (req *DoSendItemMailReq) Do() (interface{}, int32, error) {
	ack := &DoSendItemMailRsp{}
	ack.Result = 0
	ack.RetMsg = "Success"

	title, err := url.QueryUnescape(req.MailTitle)
	if err != nil {
		return nil, ErrDecodeMailTitleFailed, err
	}
	content, err := url.QueryUnescape(req.MailContent)
	if err != nil {
		return nil, ErrDecodeMailContentFailed, err
	}

	req.MailTitle = title
	req.MailContent = content
	data, err := json.Marshal(req)
	if err == nil {
		iserver.GetSrvInst().FireRPC("SendOneMail", serializer.Serialize(data))
	}

	log := &RequestLog{}
	log.DtEventTime = time.Now().Format("2006-01-02 15:04:05")
	log.AreaID = req.AreaID
	log.VOpenID = req.OpenID
	log.ItemID = int(req.ItemID)
	log.ItemNum = int(req.ItemNum)
	log.Serial = req.Serial
	log.Source = int(req.Source)
	log.Cmd = int(req.GetCmdID())
	tlog.Format(log)

	return ack, 0, nil
}

/* --------------------------------------------------------------------------*/

// DoSendItemMailRsp 邮件赠送物品应答 4116
type DoSendItemMailRsp struct {
	Result int32  `json:"Result"` // 结果: 0 激活成功；1 帐号曾经激活；其它失败
	RetMsg string `json:"RetMsg"` // 返回消息
}

// GetCmdID 获取命令ID
func (rsp *DoSendItemMailRsp) GetCmdID() uint32 {
	return 4116
}

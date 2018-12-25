package idip

import (
	"db"
	"math"
	"net/url"
	"sort"
	"strconv"
	"zeus/iserver"
	"zeus/serializer"
)

/*

 IDIP 走马灯相关

*/

// DoSendMarqueeReq 发走马灯请求 4099
type DoSendMarqueeReq struct {
	AreaID        uint32 `json:"AreaId"`        // 服务器：微信（1），手Q（2）
	PlatID        uint8  `json:"PlatId"`        // 平台：IOS（0），安卓（1）
	SlidingTime   uint32 `json:"SlidingTime"`   // 滚动间隔时间 ：**秒/次
	NoticeContent string `json:"NoticeContent"` // 公告内容
	BeginTime     uint32 `json:"BeginTime"`     // 开始时间
	EndTime       uint32 `json:"EndTime"`       // 结束时间
	Source        uint32 `json:"Source"`        // 渠道号，由前端生成，不需要填写
	Serial        string `json:"Serial"`        // 流水号，由前端生成，不需要填写
}

// GetCmdID 获取命令ID
func (req *DoSendMarqueeReq) GetCmdID() uint32 {
	return 4099
}

// GetAreaID 获取AreaID
func (req *DoSendMarqueeReq) GetAreaID() uint32 {
	return req.AreaID
}

// Do 执行发走马灯请求
const MarqueeMinInternalTime = 60

func (req *DoSendMarqueeReq) Do() (interface{}, int32, error) {
	ack := &DoSendMarqueeRsp{}

	content, err := url.QueryUnescape(req.NoticeContent)
	if err != nil {
		return nil, ErrDecodeNoticeContentFailed, err
	}

	if req.SlidingTime < MarqueeMinInternalTime {
		req.SlidingTime = MarqueeMinInternalTime
	}

	data := &db.AnnuonceData{
		ID:           db.GetAnnuonceGlobalID(),
		ServerID:     req.AreaID,
		PlatID:       req.PlatID,
		InternalTime: req.SlidingTime,
		Content:      content,
		StartTime:    req.BeginTime,
		EndTime:      req.EndTime,
		Source:       req.Source,
		Serial:       req.Serial,
	}

	// 数据存入库
	if db.AddAnnuonceData(data) == false {
		ack.Result = ErrSendAnnuonce
		return ack, ErrSendAnnuonce, nil
	}

	iserver.GetSrvInst().FireRPC("AddAnnuonce", serializer.Serialize(data.ID))
	ack.Result = 0
	return ack, 0, nil
}

/* --------------------------------------------------------------------------*/

// DoSendMarqueeRsp 发走马灯应答 4100
type DoSendMarqueeRsp struct {
	Result int32  `json:"Result"` // 结果：成功(0)，玩家不存在(1)，失败(其他)
	RetMsg string `json:"RetMsg"` // 返回消息
}

// GetCmdID 获取命令ID
func (rsp *DoSendMarqueeRsp) GetCmdID() uint32 {
	return 4100
}

/* --------------------------------------------------------------------------*/

// DoDeleteMarqueeReq 删除走马灯请求 4101
type DoDeleteMarqueeReq struct {
	AreaID  uint32 `json:"AreaId"`  // 服务器：微信（1），手Q（2）
	PlatID  uint8  `json:"PlatId"`  // 平台：IOS（0），安卓（1）
	EventID string `json:"EventId"` // 事件ID
}

// GetCmdID 获取命令ID
func (req *DoDeleteMarqueeReq) GetCmdID() uint32 {
	return 4101
}

// GetAreaID 获取AreaID
func (req *DoDeleteMarqueeReq) GetAreaID() uint32 {
	return req.AreaID
}

// Do 执行发走马灯请求
func (req *DoDeleteMarqueeReq) Do() (interface{}, int32, error) {
	ack := &DoDeleteMarqueeRsp{}

	id, err := strconv.ParseInt(req.EventID, 10, 64)
	if err != nil {
		ack.Result = ErrDelAnnuonce
		return ack, ErrDelAnnuonce, nil
	}

	if db.DelAnnuoncing(uint64(id)) {
		iserver.GetSrvInst().FireRPC("DelAnnuoncing", serializer.Serialize(uint64(id)))
	} else {
		ack.Result = ErrDelAnnuonce
		return ack, ErrDelAnnuonce, nil
	}

	ack.Result = 0
	return ack, 0, nil
}

/* --------------------------------------------------------------------------*/

// DoDeleteMarqueeRsp 删除走马灯应答 4102
type DoDeleteMarqueeRsp struct {
	Result int32  `json:"Result"` // 结果：成功(0)，玩家不存在(1)，失败(其他)
	RetMsg string `json:"RetMsg"` // 返回消息
}

// GetCmdID 获取命令ID
func (rsp *DoDeleteMarqueeRsp) GetCmdID() uint32 {
	return 4102
}

/* --------------------------------------------------------------------------*/

// DoQueryMarqueeReq 查询走马灯请求 4105
type DoQueryMarqueeReq struct {
	AreaID    uint32 `json:"AreaId"`    // 服务器：微信（1），手Q（2）
	PlatID    uint8  `json:"PlatId"`    // 平台：IOS（0），安卓（1）
	BeginTime uint32 `json:"BeginTime"` // 开始时间
	EndTime   uint32 `json:"EndTime"`   // 结束时间
	PageNo    uint8  `json:"PageNo"`    // 页码
}

// GetCmdID 获取命令ID
func (req *DoQueryMarqueeReq) GetCmdID() uint32 {
	return 4105
}

// GetAreaID 获取AreaID
func (req *DoQueryMarqueeReq) GetAreaID() uint32 {
	return req.AreaID
}

// Do 执行查询走马灯请求
func (req *DoQueryMarqueeReq) Do() (interface{}, int32, error) {

	ack := &DoQueryMarqueeRsp{}

	Sdata := db.SliAnData(db.QueryAnnuonceData(req.BeginTime, req.EndTime))
	sort.Sort(Sdata)

	ack.TotalCount = uint32(len(Sdata))
	ack.TotalPageNo = uint8(math.Ceil(float64(ack.TotalCount) / float64(maxMarqueeInfos)))

	startIndex := (req.PageNo - 1) * maxMarqueeInfos
	endIndex := uint32(req.PageNo) * maxMarqueeInfos
	if endIndex > ack.TotalCount {
		endIndex = ack.TotalCount
	}

	for i, d := range Sdata {

		if i < int(startIndex) {
			continue
		}

		if i >= int(endIndex) {
			break
		}

		tmp := SMarqueeInfo{
			HouseID:       strconv.FormatInt(int64(d.ID), 10),
			SlidingTime:   d.InternalTime,
			BeginTime:     d.StartTime,
			EndTime:       d.EndTime,
			NoticeContent: url.QueryEscape(d.Content),
		}

		ack.MarqueeInfo = append(ack.MarqueeInfo, tmp)

	}

	ack.MarqueeInfoCount = uint32(len(ack.MarqueeInfo))

	return ack, 0, nil
}

/* --------------------------------------------------------------------------*/

const maxMarqueeInfos = 30 // 单次应答最大走马灯信息数量

// SMarqueeInfo 走马灯信息
type SMarqueeInfo struct {
	HouseID       string `json:"HouseId"`       // 走马灯ID
	SlidingTime   uint32 `json:"SlidingTime"`   // 滚动间隔时间 ：**秒/次
	BeginTime     uint32 `json:"BeginTime"`     // 开始时间（时间戳）
	EndTime       uint32 `json:"EndTime"`       // 结束时间（时间戳）
	NoticeContent string `json:"NoticeContent"` // 公告内容
}

// DoQueryMarqueeRsp 查询走马灯应答 4106
type DoQueryMarqueeRsp struct {
	TotalCount       uint32         `json:"TotalCount"`        // 总数
	TotalPageNo      uint8          `json:"TotalPageNo"`       // 总页码
	MarqueeInfoCount uint32         `json:"MarqueeInfo_count"` // 当前应答中包含的总数
	MarqueeInfo      []SMarqueeInfo `json:"MarqueeInfo"`       // 走马灯信息
}

// GetCmdID 获取命令ID
func (rsp *DoQueryMarqueeRsp) GetCmdID() uint32 {
	return 4106
}

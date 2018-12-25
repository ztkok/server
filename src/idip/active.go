package idip

import (
	"zeus/dbservice"
	"zeus/login"
)

/*

 IDIP 激活帐号相关

*/

/* --------------------------------------------------------------------------*/

// DoActiveUsrReq 激活帐号 （开白名单）请求 4097
type DoActiveUsrReq struct {
	AreaID uint32 `json:"AreaId"` // 服务器：微信（1），手Q（2）
	PlatID uint8  `json:"PlatId"` // 平台：IOS（0），安卓（1）
	OpenID string `json:"OpenId"` // openid
	Grade  uint32 `json:"Grade"`  // 级别：1表示内部 2表示外部玩家
	Source uint32 `json:"Source"` // 渠道号，由前端生成，不需要填写
	Serial string `json:"Serial"` // 流水号，由前端生成，不需要填写
}

// Do 执行激活帐号请求
func (req *DoActiveUsrReq) Do() (interface{}, int32, error) {
	uid, err := dbservice.GetUID(req.OpenID)
	if uid == 0 {
		// 没有帐号的情况下需要先创建帐号
		app := login.NewApp()
		_, err := app.DoCreateNewUser(req.OpenID, "", req.Grade)
		if err != nil {
			return nil, ErrActiveFailed, err
		}
		return &DoActiveUsrRsp{Result: 0}, 0, nil
	}
	if err != nil {
		return nil, ErrActiveFailed, err
	}

	util := dbservice.Account(uid)
	curGrade, err := util.GetGrade()
	if err != nil {
		return nil, ErrActiveFailed, err
	}
	if curGrade == req.Grade {
		return &DoActiveUsrRsp{Result: 1}, 0, nil
	}

	err = util.SetGrade(req.Grade)
	if err != nil {
		return nil, ErrActiveFailed, err
	}
	return &DoActiveUsrRsp{Result: 0}, 0, nil
}

// GetCmdID 获取命令ID
func (req *DoActiveUsrReq) GetCmdID() uint32 {
	return 4097
}

// GetAreaID 获取AreaID
func (req *DoActiveUsrReq) GetAreaID() uint32 {
	return req.AreaID
}

/* --------------------------------------------------------------------------*/

// DoActiveUsrRsp 激活帐号 （开白名单）应答 4098
type DoActiveUsrRsp struct {
	Result int32  `json:"Result"` // 结果: 0 激活成功；1 帐号曾经激活；其它失败
	RetMsg string `json:"RetMsg"` // 返回消息
}

// GetCmdID 获取命令ID
func (rsp *DoActiveUsrRsp) GetCmdID() uint32 {
	return 4098
}

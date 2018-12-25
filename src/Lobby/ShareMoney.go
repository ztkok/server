package main

import (
	"crypto/md5"
	"datadef"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"protoMsg"
	"strings"
	"time"
	"zeus/iserver"
	"zeus/l5"
)

// Share Values
const (
	MSDKKey   = "7eead96a3fdb063615b181d7c01480e4"
	AppID     = "1106393072"
	Peoplenum = 10
	Secret    = "20163eea001bc90aac2b6e9f4d35eff5"

	WXShareKey = "2OU7wuxevLy8QHZENKFhfHzhS0kTMruD"
	WXNoticeID = 90119676
	WXNum      = 10

	WXModId = 64028801
	WXCmdId = 65536
)

// GenSig 生成签名
func GenQQSig(timestamp int64) string {
	h := md5.New()
	str := fmt.Sprintf("%s%d", MSDKKey, timestamp)
	h.Write([]byte(str))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// doQQShare QQ分享post请求
func (user *LobbyUser) doQQShare(reqAddr string, data []byte, ret interface{}) error {
	user.Info("Do QQ share")
	resp, err := http.Post(reqAddr, "application/json", strings.NewReader(string(data)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(body, ret); err != nil {
		return err
	}
	return nil
}

// QQShareRMB [手Q]分享宝箱接口
func (user *LobbyUser) QQShareRMB(shareRMBMoney *protoMsg.ShareRMBMoney) {
	user.Info("QQ share RMB")
	timestamp := time.Now().Unix()
	sig := GenQQSig(timestamp)

	qqChestReq := &datadef.QQChestReq{}
	qqChestReq.AppID = AppID
	qqChestReq.OpenID = shareRMBMoney.OpenID
	qqChestReq.AccessToken = shareRMBMoney.AccessToken
	qqChestReq.Actid = shareRMBMoney.ActID
	if qqChestReq.Actid == 155 {
		qqChestReq.Num = 17
	} else if qqChestReq.Actid == 156 {
		qqChestReq.Num = 1600
	}
	qqChestReq.Peoplenum = Peoplenum
	qqChestReq.Type = 0
	qqChestReq.Secret = Secret

	data, err := json.Marshal(qqChestReq)
	if err != nil {
		user.Error("QQShareRMB failed, Marshal err: ", err)
		return
	}
	url := fmt.Sprintf(`%s/relation/qq_gain_chestV2?timestamp=%d&appid=%s&sig=%s&openid=%s&encode=2`,
		GetSrvInst().msdkAddr, timestamp, AppID, sig, qqChestReq.OpenID)

	rsp := &datadef.QQChestRsp{}
	if err := user.doQQShare(url, data, rsp); err != nil {
		user.Error("doQQShare err: ", err)
		return
	}

	shareRMBMoney.Ret = rsp.Ret
	shareRMBMoney.Msg = rsp.Msg
	shareRMBMoney.BoxID = rsp.BoxID

	user.Info("QQ share RMB url: ", url, " ret: ", rsp.Ret, " msg: ", rsp.Msg, " boxid: ", rsp.BoxID)

	if err := user.RPC(iserver.ServerTypeClient, "ShareRMBMoney", shareRMBMoney); err != nil {
		user.Error("RPC ShareRMBMoney err：", err)
	}
}

/********************************************************************************************/

// GenSig 生成签名
func GenWXSig(actID uint32, openID string, serial string, timestamp int64) string {
	h := md5.New()
	str := fmt.Sprintf(`actid=%d&noticeid=%d&num=%d&openid=%s&serial=%s&stamp=%d&key=%s`,
		actID, WXNoticeID, WXNum, openID, serial, timestamp, WXShareKey)
	h.Write([]byte(str))
	return fmt.Sprintf("%x", h.Sum(nil))
}

// WXShareRMB [微信]创建福袋接口
func (user *LobbyUser) WXShareRMB(shareRMBMoney *protoMsg.ShareRMBMoney) {
	ip, port := l5.CallService(WXModId, WXCmdId)
	if len(ip) == 0 {
		user.Error("CallService l5 failed")
		return
	}

	timestamp := time.Now().Unix()
	serial := fmt.Sprintf("%s%d", shareRMBMoney.OpenID, timestamp)
	sign := GenWXSig(shareRMBMoney.ActID, shareRMBMoney.OpenID, serial, timestamp)

	url := fmt.Sprintf("http://%s:%d/game/innerblessbag?actid=%d&noticeid=%d&num=%d&openid=%s&serial=%s&stamp=%d&sign=%s",
		ip, port, shareRMBMoney.ActID, WXNoticeID, WXNum, shareRMBMoney.OpenID, serial, timestamp, sign)

	resp, err := http.Get(url)
	if err != nil {
		user.Info("WXShareRMB failed, Get err：", err)
		return
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		user.Error("WXShareRMB failed, ReadAll err: ", err)
		return
	}

	rsp := &datadef.WXChestRsp{}
	if err = json.Unmarshal(body, rsp); err != nil {
		user.Error("WXShareRMB failed, Unmarshal err: ", err)
		return
	}

	shareRMBMoney.Ret = rsp.Ret
	shareRMBMoney.Msg = rsp.Msg
	shareRMBMoney.Url = rsp.Data.Url

	user.RPC(iserver.ServerTypeClient, "ShareRMBMoney", shareRMBMoney)
	user.Info("Weixin share RMB, url: ", url)
}

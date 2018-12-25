package pay

import (
	"common"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"net/url"
	"sort"
	"strings"
	"time"
	"zeus/l5"

	log "github.com/cihub/seelog"
)

var (
	payModId, payCmdId   int    //米大师L5使用的SID
	iosAppKey, andAppKey string //ios、android系统使用的appkey
)

func init() {
	var l5id, keyid1, keyid2 uint64
	env := common.GetPaySystemValue(common.PaySystem_PayEnv)

	switch env {
	case "1":
		{
			l5id = common.PaySystem_MidasL5SID1
			keyid1 = common.PaySystem_MidasIOSAppKey1
			keyid2 = common.PaySystem_MidasANDAppKey1
		}
	case "2":
		{
			l5id = common.PaySystem_MidasL5SID2
			keyid1 = common.PaySystem_MidasIOSAppKey2
			keyid2 = common.PaySystem_MidasANDAppKey2
		}
	case "3":
		{
			l5id = common.PaySystem_MidasL5SID3
			keyid1 = common.PaySystem_MidasIOSAppKey1
			keyid2 = common.PaySystem_MidasANDAppKey1
		}
	}

	value := common.GetPaySystemValue(l5id)
	ss := strings.Split(value, ":")
	if len(ss) == 2 {
		payModId = common.StringToInt(ss[0])
		payCmdId = common.StringToInt(ss[1])
	}

	iosAppKey = common.GetPaySystemValue(keyid1)
	andAppKey = common.GetPaySystemValue(keyid2)
}

// 根据登录类型获取midas需要用的appid
func GetMidasAppidByType(typ string) string {
	if typ == "android" {
		return AndroidAppid
	} else if typ == "iap" {
		return IOSAppid
	}

	return ""
}

// 根据登录类型获取midas需要用的appid
func GetMidasAppKeyByType(typ string) string {
	if typ == "android" {
		return andAppKey
	} else if typ == "iap" {
		return iosAppKey
	}

	return ""
}

// 根据操作系统类型获取msdk需要用的appid
func GetMSDKAppidByType(typ string) string {
	if typ == "desktop_m_qq" {
		return common.QQAppIDStr
	} else if typ == "desktop_m_wx" {
		return common.WXAppID
	} else if typ == "desktop_m_guest" {
		return common.GAppID
	}

	return ""
}

// 获取全局唯一的订单号
func GetBillNoByUid(uid uint64) string {
	res := time.Now().Format("20060102150405")

	nanos := common.Int64ToString(int64(time.Now().Nanosecond()))
	for i := 0; i < 9-len(nanos); i++ {
		nanos = "0" + nanos
	}

	res += nanos
	res += common.Uint64ToString(uid)

	return res
}

// 生成URL
func GenUrl(os string, method string, uri string, values *url.Values) string {
	site := GetMidasSite()
	if len(site) == 0 {
		return ""
	}

	params := values.Encode()
	sig := GenSig(os, method, uri, values)

	return site + uri + "?" + params + "&sig=" + sig
}

// 生成签名
func GenSig(os string, method string, uri string, values *url.Values) string {
	//关键字排序
	ks := []string{}
	for k := range *values {
		ks = append(ks, k)
	}
	sort.Strings(ks)

	//拼接
	var params string
	for _, v := range ks {
		if len(params) != 0 {
			params += "&"
		}
		params += (v + "=" + values.Get(v))
	}

	srcStr := method + "&" + url.QueryEscape(uri) + "&" + url.QueryEscape(params)
	secretKey := GetMidasAppKeyByType(os) + "&"

	//加密及编码
	mac := hmac.New(sha1.New, []byte(secretKey))
	mac.Write([]byte(srcStr))
	s := []byte(mac.Sum(nil))
	sig := base64.StdEncoding.EncodeToString(s)

	return sig
}

// 获取米大师的地址
func GetMidasSite() string {
	ip, port := l5.CallService(payModId, payCmdId)
	if len(ip) == 0 {
		log.Error("CallService l5 failed")
		return ""
	}

	return "http://" + ip + ":" + common.UintToString(port)
}

// 获取月卡的数据库标识
func GetMonthCardDBFlag(os string) string {
	if os == "android" {
		return common.MarketingNameMonthCardAND
	} else if os == "iap" {
		return common.MarketingNameMonthCardIOS
	}

	return ""
}

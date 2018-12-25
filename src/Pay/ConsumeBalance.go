package pay

import (
	"common"
	"encoding/json"
	"fmt"
	"io/ioutil"

	log "github.com/cihub/seelog"

	"net/http"
	"net/url"
	"strconv"
	"time"
)

// ConsumeBalance 购买物品
func ConsumeBalance(param []byte, msdkAddr string, amount uint32, uid uint64) (*MidasDeductVirtualCoinResult, string, error) {
	if amount == 0 {
		return nil, "", nil
	}

	reqMsg := &DeductVirtualCoin{}
	err := json.Unmarshal(param, reqMsg)
	log.Debug("DeductVirtualCoin: ", string(param), " uid: ", uid, " amount: ", amount)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return nil, "", err
	}

	nowStr := strconv.FormatInt(time.Now().Unix(), 10)
	urlPath := "/v3/r/mpay/pay_m"
	billno := GetBillNoByUid(uid)

	value := url.Values{}
	value.Add("openid", reqMsg.Openid)
	value.Add("openkey", reqMsg.Openkey)
	value.Add("pay_token", reqMsg.PayToken)
	value.Add("appid", GetMidasAppidByType(reqMsg.Os))
	value.Add("ts", nowStr)
	value.Add("pf", reqMsg.Pf)
	value.Add("pfkey", reqMsg.PfKey)
	value.Add("zoneid", "1")
	value.Add("amt", common.Uint32ToString(amount))
	value.Add("billno", billno)
	value.Add("accounttype", "common")

	urlInfo := GenUrl(reqMsg.Os, "GET", urlPath, &value)
	log.Debug("HTTP URL: ", urlInfo)
	if len(urlInfo) == 0 {
		return nil, "", fmt.Errorf("%s", "GenUrl failed")
	}

	req, err := http.NewRequest("GET", urlInfo, nil)
	if err != nil {
		return nil, "", err
	}

	cookies := []*http.Cookie{
		{
			Name:     "session_id",
			Value:    reqMsg.SessionID,
			HttpOnly: true,
		},
		{
			Name:     "session_type",
			Value:    reqMsg.SessionType,
			HttpOnly: true,
		},
		{
			Name:     "org_loc",
			Value:    urlPath,
			HttpOnly: true,
		},
	}

	for _, v := range cookies {
		req.AddCookie(v)
	}

	respInfo := &MidasDeductVirtualCoinResult{}
	client := &http.Client{}

	for i := 0; i < 3; i++ {
		resp, err := client.Do(req)
		if err != nil {
			return nil, "", err
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, "", err
		}

		defer resp.Body.Close()

		log.Debug("MidasDeductVirtualCoinResult body: ", string(body))

		err = json.Unmarshal(body, respInfo)
		if err != nil {
			log.Error("Unmarshal err: ", err)
			return nil, "", err
		}

		// 未获得订单状态，重试
		if respInfo.Ret == 3000111 {
			continue
		}

		// 扣除成功发放道具 获得前次请求成功发放道具
		if respInfo.Ret == 0 || respInfo.Ret == 1002215 {
			respInfo.Ret = 0
			return respInfo, reqMsg.Os, nil
		}
	}

	return respInfo, reqMsg.Os, nil
}

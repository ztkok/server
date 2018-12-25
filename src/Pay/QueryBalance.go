package pay

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	log "github.com/cihub/seelog"

	"net/url"
)

// QueryBalance 查询剩余虚拟币
func QueryBalance(r []byte, msdkAddr string) (*MidasQueryBalanceResult, string, error) {
	reqMsg := &QueryBalaceByClient{}
	err := json.Unmarshal(r, reqMsg)
	if err != nil {
		return nil, "", err
	}

	nowStr := strconv.FormatInt(time.Now().Unix(), 10)
	urlPath := "/v3/r/mpay/get_balance_m"

	value := url.Values{}
	value.Add("openid", reqMsg.Openid)
	value.Add("openkey", reqMsg.Openkey)
	value.Add("pay_token", reqMsg.PayToken)
	value.Add("appid", GetMidasAppidByType(reqMsg.Os))
	value.Add("ts", nowStr)
	value.Add("pf", reqMsg.Pf)
	value.Add("pfkey", reqMsg.PfKey)
	value.Add("zoneid", "1")
	//value.Add("accounttype","common")

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
		}}
	/*
		c4 := &http.Cookie{
			Name:     "appip",
			Value:    "",
			HttpOnly: true,
		}
	*/

	for _, c := range cookies {
		req.AddCookie(c)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	respInfo := &MidasQueryBalanceResult{}
	err = json.Unmarshal(body, respInfo)
	if err != nil {
		return nil, "", err
	}

	return respInfo, reqMsg.Os, nil
}

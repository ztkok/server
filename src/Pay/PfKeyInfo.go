package pay

import (
	"common"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	log "github.com/cihub/seelog"
)

// GetpfAndpfKey 获取pf和pfKey
func GetpfAndpfKey(data *GetpfAndpfkeyData, curTime string, msdkAddr string) (*GetpfAndpfkeyRet, error) {
	respInfo := &GetpfAndpfkeyRet{}

	cmdData, err := json.Marshal(data)
	if err != nil {
		respInfo.Ret = ErrPayJSONDecodeFailed
		log.Error("Marshal err: ", err)
		return nil, err
	}

	h := md5.New()
	sigPara := common.MSDKKey + curTime

	io.WriteString(h, sigPara)
	token := fmt.Sprintf("%x", h.Sum(nil))

	urlinfo := msdkAddr + "/auth/get_pfval?" +
		"timestamp=" + curTime +
		"&appid=" + data.Appid +
		"&sig=" + token +
		"&openid=" + data.Openid +
		"&encode=2"

	resp, err := http.Post(urlinfo, "application/x-www-form-urlencoded", strings.NewReader(string(cmdData)))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, respInfo)
	log.Debug("GetpfAndpfKey body: ", string(body))
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return nil, err
	}

	if respInfo.Ret != 0 {
		return nil, fmt.Errorf("GetPf Error %s Ret %d", respInfo.Msg, respInfo.Ret)
	}

	return respInfo, nil
}

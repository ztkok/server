package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type CallbackrewardData struct {
	Id          uint64
	ActivityNum uint64
	CallNum     uint64
	Reward      string
}

var callbackreward map[uint64]CallbackrewardData
var callbackrewardLock sync.RWMutex

func LoadCallbackreward() {
	callbackrewardLock.Lock()
	defer callbackrewardLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/callbackreward.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	callbackreward = nil
	err = json.Unmarshal(data, &callbackreward)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetCallbackrewardMap() map[uint64]CallbackrewardData {
	callbackrewardLock.RLock()
	defer callbackrewardLock.RUnlock()

	callbackreward2 := make(map[uint64]CallbackrewardData)
	for k, v := range callbackreward {
		callbackreward2[k] = v
	}

	return callbackreward2
}

func GetCallbackreward(key uint64) (CallbackrewardData, bool) {
	callbackrewardLock.RLock()
	defer callbackrewardLock.RUnlock()

	val, ok := callbackreward[key]

	return val, ok
}

func GetCallbackrewardMapLen() int {
	return len(callbackreward)
}

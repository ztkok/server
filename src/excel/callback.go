package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type CallbackData struct {
	Id                     uint64
	Exittime               uint64
	Repeattime             uint64
	Clicktimes             uint64
	ClickrewardType        uint64
	ClickrewardNum         uint64
	ClickrewardMailTitle   string
	ClickrewardMailContent string
	CallrewardType         uint64
	CallrewardNum          uint64
	CallrewardMailTitle    string
	CallrewardMailContent  string
}

var callback map[uint64]CallbackData
var callbackLock sync.RWMutex

func LoadCallback() {
	callbackLock.Lock()
	defer callbackLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/callback.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	callback = nil
	err = json.Unmarshal(data, &callback)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetCallbackMap() map[uint64]CallbackData {
	callbackLock.RLock()
	defer callbackLock.RUnlock()

	callback2 := make(map[uint64]CallbackData)
	for k, v := range callback {
		callback2[k] = v
	}

	return callback2
}

func GetCallback(key uint64) (CallbackData, bool) {
	callbackLock.RLock()
	defer callbackLock.RUnlock()

	val, ok := callback[key]

	return val, ok
}

func GetCallbackMapLen() int {
	return len(callback)
}

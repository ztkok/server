package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type BackrewardData struct {
	Id          uint64
	ActivityNum uint64
	Reward      string
}

var backreward map[uint64]BackrewardData
var backrewardLock sync.RWMutex

func LoadBackreward() {
	backrewardLock.Lock()
	defer backrewardLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/backreward.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	backreward = nil
	err = json.Unmarshal(data, &backreward)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetBackrewardMap() map[uint64]BackrewardData {
	backrewardLock.RLock()
	defer backrewardLock.RUnlock()

	backreward2 := make(map[uint64]BackrewardData)
	for k, v := range backreward {
		backreward2[k] = v
	}

	return backreward2
}

func GetBackreward(key uint64) (BackrewardData, bool) {
	backrewardLock.RLock()
	defer backrewardLock.RUnlock()

	val, ok := backreward[key]

	return val, ok
}

func GetBackrewardMapLen() int {
	return len(backreward)
}

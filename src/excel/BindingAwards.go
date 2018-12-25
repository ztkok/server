package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type BindingAwardsData struct {
	Id      uint64
	Type    uint64
	Require uint64
	Awards  string
}

var BindingAwards map[uint64]BindingAwardsData
var BindingAwardsLock sync.RWMutex

func LoadBindingAwards() {
	BindingAwardsLock.Lock()
	defer BindingAwardsLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/BindingAwards.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	BindingAwards = nil
	err = json.Unmarshal(data, &BindingAwards)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetBindingAwardsMap() map[uint64]BindingAwardsData {
	BindingAwardsLock.RLock()
	defer BindingAwardsLock.RUnlock()

	BindingAwards2 := make(map[uint64]BindingAwardsData)
	for k, v := range BindingAwards {
		BindingAwards2[k] = v
	}

	return BindingAwards2
}

func GetBindingAwards(key uint64) (BindingAwardsData, bool) {
	BindingAwardsLock.RLock()
	defer BindingAwardsLock.RUnlock()

	val, ok := BindingAwards[key]

	return val, ok
}

func GetBindingAwardsMapLen() int {
	return len(BindingAwards)
}

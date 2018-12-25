package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type NewyearTargetPoolData struct {
	Id uint64
}

var newyearTargetPool map[uint64]NewyearTargetPoolData
var newyearTargetPoolLock sync.RWMutex

func LoadNewyearTargetPool() {
	newyearTargetPoolLock.Lock()
	defer newyearTargetPoolLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/newyearTargetPool.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	newyearTargetPool = nil
	err = json.Unmarshal(data, &newyearTargetPool)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetNewyearTargetPoolMap() map[uint64]NewyearTargetPoolData {
	newyearTargetPoolLock.RLock()
	defer newyearTargetPoolLock.RUnlock()

	newyearTargetPool2 := make(map[uint64]NewyearTargetPoolData)
	for k, v := range newyearTargetPool {
		newyearTargetPool2[k] = v
	}

	return newyearTargetPool2
}

func GetNewyearTargetPool(key uint64) (NewyearTargetPoolData, bool) {
	newyearTargetPoolLock.RLock()
	defer newyearTargetPoolLock.RUnlock()

	val, ok := newyearTargetPool[key]

	return val, ok
}

func GetNewyearTargetPoolMapLen() int {
	return len(newyearTargetPool)
}

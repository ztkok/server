package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type DragonrewardData struct {
	Id     uint64
	Single string
}

var dragonreward map[uint64]DragonrewardData
var dragonrewardLock sync.RWMutex

func LoadDragonreward() {
	dragonrewardLock.Lock()
	defer dragonrewardLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/dragonreward.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	dragonreward = nil
	err = json.Unmarshal(data, &dragonreward)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetDragonrewardMap() map[uint64]DragonrewardData {
	dragonrewardLock.RLock()
	defer dragonrewardLock.RUnlock()

	dragonreward2 := make(map[uint64]DragonrewardData)
	for k, v := range dragonreward {
		dragonreward2[k] = v
	}

	return dragonreward2
}

func GetDragonreward(key uint64) (DragonrewardData, bool) {
	dragonrewardLock.RLock()
	defer dragonrewardLock.RUnlock()

	val, ok := dragonreward[key]

	return val, ok
}

func GetDragonrewardMapLen() int {
	return len(dragonreward)
}

package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type AisysData struct {
	Id          uint64
	Label       string
	Valuestring string
	Valueuint   uint64
	Valuefloat  float32
}

var aisys map[uint64]AisysData
var aisysLock sync.RWMutex

func LoadAisys() {
	aisysLock.Lock()
	defer aisysLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/aisys.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	aisys = nil
	err = json.Unmarshal(data, &aisys)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetAisysMap() map[uint64]AisysData {
	aisysLock.RLock()
	defer aisysLock.RUnlock()

	aisys2 := make(map[uint64]AisysData)
	for k, v := range aisys {
		aisys2[k] = v
	}

	return aisys2
}

func GetAisys(key uint64) (AisysData, bool) {
	aisysLock.RLock()
	defer aisysLock.RUnlock()

	val, ok := aisys[key]

	return val, ok
}

func GetAisysMapLen() int {
	return len(aisys)
}

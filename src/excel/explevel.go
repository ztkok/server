package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type ExplevelData struct {
	Id    uint64
	Value uint64
}

var explevel map[uint64]ExplevelData
var explevelLock sync.RWMutex

func LoadExplevel() {
	explevelLock.Lock()
	defer explevelLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/explevel.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	explevel = nil
	err = json.Unmarshal(data, &explevel)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetExplevelMap() map[uint64]ExplevelData {
	explevelLock.RLock()
	defer explevelLock.RUnlock()

	explevel2 := make(map[uint64]ExplevelData)
	for k, v := range explevel {
		explevel2[k] = v
	}

	return explevel2
}

func GetExplevel(key uint64) (ExplevelData, bool) {
	explevelLock.RLock()
	defer explevelLock.RUnlock()

	val, ok := explevel[key]

	return val, ok
}

func GetExplevelMapLen() int {
	return len(explevel)
}

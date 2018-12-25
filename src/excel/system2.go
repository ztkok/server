package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type System2Data struct {
	Id    uint64
	Value string
	Desc  string
}

var system2 map[uint64]System2Data
var system2Lock sync.RWMutex

func LoadSystem2() {
	system2Lock.Lock()
	defer system2Lock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/system2.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	system2 = nil
	err = json.Unmarshal(data, &system2)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetSystem2Map() map[uint64]System2Data {
	system2Lock.RLock()
	defer system2Lock.RUnlock()

	system22 := make(map[uint64]System2Data)
	for k, v := range system2 {
		system22[k] = v
	}

	return system22
}

func GetSystem2(key uint64) (System2Data, bool) {
	system2Lock.RLock()
	defer system2Lock.RUnlock()

	val, ok := system2[key]

	return val, ok
}

func GetSystem2MapLen() int {
	return len(system2)
}

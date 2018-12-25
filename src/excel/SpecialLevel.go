package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type SpecialLevelData struct {
	Id         uint64
	Level      uint64
	Requirenum uint64
	Awards     string
	Starttime  string
}

var SpecialLevel map[uint64]SpecialLevelData
var SpecialLevelLock sync.RWMutex

func LoadSpecialLevel() {
	SpecialLevelLock.Lock()
	defer SpecialLevelLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/SpecialLevel.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	SpecialLevel = nil
	err = json.Unmarshal(data, &SpecialLevel)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetSpecialLevelMap() map[uint64]SpecialLevelData {
	SpecialLevelLock.RLock()
	defer SpecialLevelLock.RUnlock()

	SpecialLevel2 := make(map[uint64]SpecialLevelData)
	for k, v := range SpecialLevel {
		SpecialLevel2[k] = v
	}

	return SpecialLevel2
}

func GetSpecialLevel(key uint64) (SpecialLevelData, bool) {
	SpecialLevelLock.RLock()
	defer SpecialLevelLock.RUnlock()

	val, ok := SpecialLevel[key]

	return val, ok
}

func GetSpecialLevelMapLen() int {
	return len(SpecialLevel)
}

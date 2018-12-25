package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type SpecialReplaceData struct {
	Id    uint64
	Items string
}

var SpecialReplace map[uint64]SpecialReplaceData
var SpecialReplaceLock sync.RWMutex

func LoadSpecialReplace() {
	SpecialReplaceLock.Lock()
	defer SpecialReplaceLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/SpecialReplace.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	SpecialReplace = nil
	err = json.Unmarshal(data, &SpecialReplace)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetSpecialReplaceMap() map[uint64]SpecialReplaceData {
	SpecialReplaceLock.RLock()
	defer SpecialReplaceLock.RUnlock()

	SpecialReplace2 := make(map[uint64]SpecialReplaceData)
	for k, v := range SpecialReplace {
		SpecialReplace2[k] = v
	}

	return SpecialReplace2
}

func GetSpecialReplace(key uint64) (SpecialReplaceData, bool) {
	SpecialReplaceLock.RLock()
	defer SpecialReplaceLock.RUnlock()

	val, ok := SpecialReplace[key]

	return val, ok
}

func GetSpecialReplaceMapLen() int {
	return len(SpecialReplace)
}

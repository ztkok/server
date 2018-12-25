package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type ReversionWardData struct {
	Id                   uint64
	BonusID              uint64
	BonusNum             uint64
	IosrewardVersion     string
	AndroidrewardVersion string
}

var ReversionWard map[uint64]ReversionWardData
var ReversionWardLock sync.RWMutex

func LoadReversionWard() {
	ReversionWardLock.Lock()
	defer ReversionWardLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/ReversionWard.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	ReversionWard = nil
	err = json.Unmarshal(data, &ReversionWard)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetReversionWardMap() map[uint64]ReversionWardData {
	ReversionWardLock.RLock()
	defer ReversionWardLock.RUnlock()

	ReversionWard2 := make(map[uint64]ReversionWardData)
	for k, v := range ReversionWard {
		ReversionWard2[k] = v
	}

	return ReversionWard2
}

func GetReversionWard(key uint64) (ReversionWardData, bool) {
	ReversionWardLock.RLock()
	defer ReversionWardLock.RUnlock()

	val, ok := ReversionWard[key]

	return val, ok
}

func GetReversionWardMapLen() int {
	return len(ReversionWard)
}

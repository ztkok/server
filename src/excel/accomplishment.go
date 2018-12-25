package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type AccomplishmentData struct {
	Id         uint64
	Experience []uint32
	Mode       uint64
	Condition1 uint64
	Amount     []uint32
}

var accomplishment map[uint64]AccomplishmentData
var accomplishmentLock sync.RWMutex

func LoadAccomplishment() {
	accomplishmentLock.Lock()
	defer accomplishmentLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/accomplishment.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	accomplishment = nil
	err = json.Unmarshal(data, &accomplishment)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetAccomplishmentMap() map[uint64]AccomplishmentData {
	accomplishmentLock.RLock()
	defer accomplishmentLock.RUnlock()

	accomplishment2 := make(map[uint64]AccomplishmentData)
	for k, v := range accomplishment {
		accomplishment2[k] = v
	}

	return accomplishment2
}

func GetAccomplishment(key uint64) (AccomplishmentData, bool) {
	accomplishmentLock.RLock()
	defer accomplishmentLock.RUnlock()

	val, ok := accomplishment[key]

	return val, ok
}

func GetAccomplishmentMapLen() int {
	return len(accomplishment)
}

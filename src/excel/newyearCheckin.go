package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type NewyearCheckinData struct {
	Id          uint64
	ActivityNum uint64
	TargetStage uint64
	TargetType  uint64
	TargetNum   uint64
	RandomType  string
	RandomNum   string
	BonusID     uint64
	BonusNum    uint64
}

var newyearCheckin map[uint64]NewyearCheckinData
var newyearCheckinLock sync.RWMutex

func LoadNewyearCheckin() {
	newyearCheckinLock.Lock()
	defer newyearCheckinLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/newyearCheckin.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	newyearCheckin = nil
	err = json.Unmarshal(data, &newyearCheckin)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetNewyearCheckinMap() map[uint64]NewyearCheckinData {
	newyearCheckinLock.RLock()
	defer newyearCheckinLock.RUnlock()

	newyearCheckin2 := make(map[uint64]NewyearCheckinData)
	for k, v := range newyearCheckin {
		newyearCheckin2[k] = v
	}

	return newyearCheckin2
}

func GetNewyearCheckin(key uint64) (NewyearCheckinData, bool) {
	newyearCheckinLock.RLock()
	defer newyearCheckinLock.RUnlock()

	val, ok := newyearCheckin[key]

	return val, ok
}

func GetNewyearCheckinMapLen() int {
	return len(newyearCheckin)
}

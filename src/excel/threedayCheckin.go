package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type ThreedayCheckinData struct {
	Id          uint64
	BonusID     uint64
	BonusNum    uint64
	ActivityNum uint64
}

var threedayCheckin map[uint64]ThreedayCheckinData
var threedayCheckinLock sync.RWMutex

func LoadThreedayCheckin() {
	threedayCheckinLock.Lock()
	defer threedayCheckinLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/threedayCheckin.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	threedayCheckin = nil
	err = json.Unmarshal(data, &threedayCheckin)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetThreedayCheckinMap() map[uint64]ThreedayCheckinData {
	threedayCheckinLock.RLock()
	defer threedayCheckinLock.RUnlock()

	threedayCheckin2 := make(map[uint64]ThreedayCheckinData)
	for k, v := range threedayCheckin {
		threedayCheckin2[k] = v
	}

	return threedayCheckin2
}

func GetThreedayCheckin(key uint64) (ThreedayCheckinData, bool) {
	threedayCheckinLock.RLock()
	defer threedayCheckinLock.RUnlock()

	val, ok := threedayCheckin[key]

	return val, ok
}

func GetThreedayCheckinMapLen() int {
	return len(threedayCheckin)
}

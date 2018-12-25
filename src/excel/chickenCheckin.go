package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type ChickenCheckinData struct {
	Id          uint64
	BonusID     uint64
	BonusNum    uint64
	ActivityNum uint64
}

var chickenCheckin map[uint64]ChickenCheckinData
var chickenCheckinLock sync.RWMutex

func LoadChickenCheckin() {
	chickenCheckinLock.Lock()
	defer chickenCheckinLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/chickenCheckin.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	chickenCheckin = nil
	err = json.Unmarshal(data, &chickenCheckin)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetChickenCheckinMap() map[uint64]ChickenCheckinData {
	chickenCheckinLock.RLock()
	defer chickenCheckinLock.RUnlock()

	chickenCheckin2 := make(map[uint64]ChickenCheckinData)
	for k, v := range chickenCheckin {
		chickenCheckin2[k] = v
	}

	return chickenCheckin2
}

func GetChickenCheckin(key uint64) (ChickenCheckinData, bool) {
	chickenCheckinLock.RLock()
	defer chickenCheckinLock.RUnlock()

	val, ok := chickenCheckin[key]

	return val, ok
}

func GetChickenCheckinMapLen() int {
	return len(chickenCheckin)
}

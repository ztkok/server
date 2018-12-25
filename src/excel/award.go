package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type AwardData struct {
	Id       uint64
	Season   uint64
	Matchtyp uint64
	Ranktyp  uint64
	Level    string
	Awards   string
}

var award map[uint64]AwardData
var awardLock sync.RWMutex

func LoadAward() {
	awardLock.Lock()
	defer awardLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/award.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	award = nil
	err = json.Unmarshal(data, &award)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetAwardMap() map[uint64]AwardData {
	awardLock.RLock()
	defer awardLock.RUnlock()

	award2 := make(map[uint64]AwardData)
	for k, v := range award {
		award2[k] = v
	}

	return award2
}

func GetAward(key uint64) (AwardData, bool) {
	awardLock.RLock()
	defer awardLock.RUnlock()

	val, ok := award[key]

	return val, ok
}

func GetAwardMapLen() int {
	return len(award)
}

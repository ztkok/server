package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type OddsData struct {
	Id    uint64
	Type  uint64
	Group string
	Tips  string
	Turn  uint64
}

var odds map[uint64]OddsData
var oddsLock sync.RWMutex

func LoadOdds() {
	oddsLock.Lock()
	defer oddsLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/odds.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	odds = nil
	err = json.Unmarshal(data, &odds)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetOddsMap() map[uint64]OddsData {
	oddsLock.RLock()
	defer oddsLock.RUnlock()

	odds2 := make(map[uint64]OddsData)
	for k, v := range odds {
		odds2[k] = v
	}

	return odds2
}

func GetOdds(key uint64) (OddsData, bool) {
	oddsLock.RLock()
	defer oddsLock.RUnlock()

	val, ok := odds[key]

	return val, ok
}

func GetOddsMapLen() int {
	return len(odds)
}

package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type MilitaryrankData struct {
	Id      uint64
	Name    string
	Value   uint64
	Award   string
	Content string
}

var militaryrank map[uint64]MilitaryrankData
var militaryrankLock sync.RWMutex

func LoadMilitaryrank() {
	militaryrankLock.Lock()
	defer militaryrankLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/militaryrank.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	militaryrank = nil
	err = json.Unmarshal(data, &militaryrank)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetMilitaryrankMap() map[uint64]MilitaryrankData {
	militaryrankLock.RLock()
	defer militaryrankLock.RUnlock()

	militaryrank2 := make(map[uint64]MilitaryrankData)
	for k, v := range militaryrank {
		militaryrank2[k] = v
	}

	return militaryrank2
}

func GetMilitaryrank(key uint64) (MilitaryrankData, bool) {
	militaryrankLock.RLock()
	defer militaryrankLock.RUnlock()

	val, ok := militaryrank[key]

	return val, ok
}

func GetMilitaryrankMapLen() int {
	return len(militaryrank)
}

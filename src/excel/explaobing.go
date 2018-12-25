package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type ExplaobingData struct {
	Id         uint64
	Roundlimit uint64
	Levellimit uint64
	Value      uint64
}

var explaobing map[uint64]ExplaobingData
var explaobingLock sync.RWMutex

func LoadExplaobing() {
	explaobingLock.Lock()
	defer explaobingLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/explaobing.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	explaobing = nil
	err = json.Unmarshal(data, &explaobing)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetExplaobingMap() map[uint64]ExplaobingData {
	explaobingLock.RLock()
	defer explaobingLock.RUnlock()

	explaobing2 := make(map[uint64]ExplaobingData)
	for k, v := range explaobing {
		explaobing2[k] = v
	}

	return explaobing2
}

func GetExplaobing(key uint64) (ExplaobingData, bool) {
	explaobingLock.RLock()
	defer explaobingLock.RUnlock()

	val, ok := explaobing[key]

	return val, ok
}

func GetExplaobingMapLen() int {
	return len(explaobing)
}

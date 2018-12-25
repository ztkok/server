package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type BraverewardData struct {
	Id     uint64
	Single uint64
	Two    uint64
	Four   uint64
}

var bravereward map[uint64]BraverewardData
var braverewardLock sync.RWMutex

func LoadBravereward() {
	braverewardLock.Lock()
	defer braverewardLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/bravereward.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	bravereward = nil
	err = json.Unmarshal(data, &bravereward)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetBraverewardMap() map[uint64]BraverewardData {
	braverewardLock.RLock()
	defer braverewardLock.RUnlock()

	bravereward2 := make(map[uint64]BraverewardData)
	for k, v := range bravereward {
		bravereward2[k] = v
	}

	return bravereward2
}

func GetBravereward(key uint64) (BraverewardData, bool) {
	braverewardLock.RLock()
	defer braverewardLock.RUnlock()

	val, ok := bravereward[key]

	return val, ok
}

func GetBraverewardMapLen() int {
	return len(bravereward)
}

package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type ActivenessData struct {
	Id        uint64
	Resetloop uint64
	Maxactive uint64
	Awards    string
}

var activeness map[uint64]ActivenessData
var activenessLock sync.RWMutex

func LoadActiveness() {
	activenessLock.Lock()
	defer activenessLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/activeness.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	activeness = nil
	err = json.Unmarshal(data, &activeness)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetActivenessMap() map[uint64]ActivenessData {
	activenessLock.RLock()
	defer activenessLock.RUnlock()

	activeness2 := make(map[uint64]ActivenessData)
	for k, v := range activeness {
		activeness2[k] = v
	}

	return activeness2
}

func GetActiveness(key uint64) (ActivenessData, bool) {
	activenessLock.RLock()
	defer activenessLock.RUnlock()

	val, ok := activeness[key]

	return val, ok
}

func GetActivenessMapLen() int {
	return len(activeness)
}

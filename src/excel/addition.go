package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type AdditionData struct {
	Id    uint64
	Value uint64
	Desc  string
}

var addition map[uint64]AdditionData
var additionLock sync.RWMutex

func LoadAddition() {
	additionLock.Lock()
	defer additionLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/addition.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	addition = nil
	err = json.Unmarshal(data, &addition)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetAdditionMap() map[uint64]AdditionData {
	additionLock.RLock()
	defer additionLock.RUnlock()

	addition2 := make(map[uint64]AdditionData)
	for k, v := range addition {
		addition2[k] = v
	}

	return addition2
}

func GetAddition(key uint64) (AdditionData, bool) {
	additionLock.RLock()
	defer additionLock.RUnlock()

	val, ok := addition[key]

	return val, ok
}

func GetAdditionMapLen() int {
	return len(addition)
}

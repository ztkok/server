package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type NewmapawardData struct {
	Id     uint64
	Single uint64
	Two    uint64
	Four   uint64
}

var Newmapaward map[uint64]NewmapawardData
var NewmapawardLock sync.RWMutex

func LoadNewmapaward() {
	NewmapawardLock.Lock()
	defer NewmapawardLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/Newmapaward.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	Newmapaward = nil
	err = json.Unmarshal(data, &Newmapaward)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetNewmapawardMap() map[uint64]NewmapawardData {
	NewmapawardLock.RLock()
	defer NewmapawardLock.RUnlock()

	Newmapaward2 := make(map[uint64]NewmapawardData)
	for k, v := range Newmapaward {
		Newmapaward2[k] = v
	}

	return Newmapaward2
}

func GetNewmapaward(key uint64) (NewmapawardData, bool) {
	NewmapawardLock.RLock()
	defer NewmapawardLock.RUnlock()

	val, ok := Newmapaward[key]

	return val, ok
}

func GetNewmapawardMapLen() int {
	return len(Newmapaward)
}

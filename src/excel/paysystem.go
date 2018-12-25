package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type PaysystemData struct {
	Id            uint64
	Desc          string
	Awards        string
	Condition     uint64
	MarketingType uint64
}

var paysystem map[uint64]PaysystemData
var paysystemLock sync.RWMutex

func LoadPaysystem() {
	paysystemLock.Lock()
	defer paysystemLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/paysystem.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	paysystem = nil
	err = json.Unmarshal(data, &paysystem)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetPaysystemMap() map[uint64]PaysystemData {
	paysystemLock.RLock()
	defer paysystemLock.RUnlock()

	paysystem2 := make(map[uint64]PaysystemData)
	for k, v := range paysystem {
		paysystem2[k] = v
	}

	return paysystem2
}

func GetPaysystem(key uint64) (PaysystemData, bool) {
	paysystemLock.RLock()
	defer paysystemLock.RUnlock()

	val, ok := paysystem[key]

	return val, ok
}

func GetPaysystemMapLen() int {
	return len(paysystem)
}

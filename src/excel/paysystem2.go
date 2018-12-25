package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type Paysystem2Data struct {
	Id    uint64
	Value string
	Desc  string
}

var paysystem2 map[uint64]Paysystem2Data
var paysystem2Lock sync.RWMutex

func LoadPaysystem2() {
	paysystem2Lock.Lock()
	defer paysystem2Lock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/paysystem2.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	paysystem2 = nil
	err = json.Unmarshal(data, &paysystem2)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetPaysystem2Map() map[uint64]Paysystem2Data {
	paysystem2Lock.RLock()
	defer paysystem2Lock.RUnlock()

	paysystem22 := make(map[uint64]Paysystem2Data)
	for k, v := range paysystem2 {
		paysystem22[k] = v
	}

	return paysystem22
}

func GetPaysystem2(key uint64) (Paysystem2Data, bool) {
	paysystem2Lock.RLock()
	defer paysystem2Lock.RUnlock()

	val, ok := paysystem2[key]

	return val, ok
}

func GetPaysystem2MapLen() int {
	return len(paysystem2)
}

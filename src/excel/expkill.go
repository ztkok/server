package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type ExpkillData struct {
	Id     uint64
	Single uint64
	Two    uint64
	Four   uint64
}

var expkill map[uint64]ExpkillData
var expkillLock sync.RWMutex

func LoadExpkill() {
	expkillLock.Lock()
	defer expkillLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/expkill.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	expkill = nil
	err = json.Unmarshal(data, &expkill)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetExpkillMap() map[uint64]ExpkillData {
	expkillLock.RLock()
	defer expkillLock.RUnlock()

	expkill2 := make(map[uint64]ExpkillData)
	for k, v := range expkill {
		expkill2[k] = v
	}

	return expkill2
}

func GetExpkill(key uint64) (ExpkillData, bool) {
	expkillLock.RLock()
	defer expkillLock.RUnlock()

	val, ok := expkill[key]

	return val, ok
}

func GetExpkillMapLen() int {
	return len(expkill)
}

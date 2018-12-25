package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type ExpmodeData struct {
	Id       uint64
	Name     string
	Value    uint64
	Valuenew uint64
}

var expmode map[uint64]ExpmodeData
var expmodeLock sync.RWMutex

func LoadExpmode() {
	expmodeLock.Lock()
	defer expmodeLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/expmode.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	expmode = nil
	err = json.Unmarshal(data, &expmode)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetExpmodeMap() map[uint64]ExpmodeData {
	expmodeLock.RLock()
	defer expmodeLock.RUnlock()

	expmode2 := make(map[uint64]ExpmodeData)
	for k, v := range expmode {
		expmode2[k] = v
	}

	return expmode2
}

func GetExpmode(key uint64) (ExpmodeData, bool) {
	expmodeLock.RLock()
	defer expmodeLock.RUnlock()

	val, ok := expmode[key]

	return val, ok
}

func GetExpmodeMapLen() int {
	return len(expmode)
}

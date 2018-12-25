package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type TankmodeData struct {
	Id     uint64
	Single string
	Two    string
	Four   string
}

var tankmode map[uint64]TankmodeData
var tankmodeLock sync.RWMutex

func LoadTankmode() {
	tankmodeLock.Lock()
	defer tankmodeLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/tankmode.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	tankmode = nil
	err = json.Unmarshal(data, &tankmode)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetTankmodeMap() map[uint64]TankmodeData {
	tankmodeLock.RLock()
	defer tankmodeLock.RUnlock()

	tankmode2 := make(map[uint64]TankmodeData)
	for k, v := range tankmode {
		tankmode2[k] = v
	}

	return tankmode2
}

func GetTankmode(key uint64) (TankmodeData, bool) {
	tankmodeLock.RLock()
	defer tankmodeLock.RUnlock()

	val, ok := tankmode[key]

	return val, ok
}

func GetTankmodeMapLen() int {
	return len(tankmode)
}

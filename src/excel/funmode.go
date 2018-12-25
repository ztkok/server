package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type FunmodeData struct {
	Id     uint64
	Single string
	Two    string
	Four   string
}

var funmode map[uint64]FunmodeData
var funmodeLock sync.RWMutex

func LoadFunmode() {
	funmodeLock.Lock()
	defer funmodeLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/funmode.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	funmode = nil
	err = json.Unmarshal(data, &funmode)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetFunmodeMap() map[uint64]FunmodeData {
	funmodeLock.RLock()
	defer funmodeLock.RUnlock()

	funmode2 := make(map[uint64]FunmodeData)
	for k, v := range funmode {
		funmode2[k] = v
	}

	return funmode2
}

func GetFunmode(key uint64) (FunmodeData, bool) {
	funmodeLock.RLock()
	defer funmodeLock.RUnlock()

	val, ok := funmode[key]

	return val, ok
}

func GetFunmodeMapLen() int {
	return len(funmode)
}

package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type LuandourewardData struct {
	Id     uint64
	Single string
	Two    string
	Four   string
}

var luandoureward map[uint64]LuandourewardData
var luandourewardLock sync.RWMutex

func LoadLuandoureward() {
	luandourewardLock.Lock()
	defer luandourewardLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/luandoureward.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	luandoureward = nil
	err = json.Unmarshal(data, &luandoureward)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetLuandourewardMap() map[uint64]LuandourewardData {
	luandourewardLock.RLock()
	defer luandourewardLock.RUnlock()

	luandoureward2 := make(map[uint64]LuandourewardData)
	for k, v := range luandoureward {
		luandoureward2[k] = v
	}

	return luandoureward2
}

func GetLuandoureward(key uint64) (LuandourewardData, bool) {
	luandourewardLock.RLock()
	defer luandourewardLock.RUnlock()

	val, ok := luandoureward[key]

	return val, ok
}

func GetLuandourewardMapLen() int {
	return len(luandoureward)
}

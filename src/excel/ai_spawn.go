package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type Ai_spawnData struct {
	Id             uint64
	Label          string
	Vbase          float32
	Timetobeginacc uint64
	Vadd           float32
	Vmax           float32
	N_ai_max       uint64
}

var ai_spawn map[uint64]Ai_spawnData
var ai_spawnLock sync.RWMutex

func LoadAi_spawn() {
	ai_spawnLock.Lock()
	defer ai_spawnLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/ai_spawn.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	ai_spawn = nil
	err = json.Unmarshal(data, &ai_spawn)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetAi_spawnMap() map[uint64]Ai_spawnData {
	ai_spawnLock.RLock()
	defer ai_spawnLock.RUnlock()

	ai_spawn2 := make(map[uint64]Ai_spawnData)
	for k, v := range ai_spawn {
		ai_spawn2[k] = v
	}

	return ai_spawn2
}

func GetAi_spawn(key uint64) (Ai_spawnData, bool) {
	ai_spawnLock.RLock()
	defer ai_spawnLock.RUnlock()

	val, ok := ai_spawn[key]

	return val, ok
}

func GetAi_spawnMapLen() int {
	return len(ai_spawn)
}

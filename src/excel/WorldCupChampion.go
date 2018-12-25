package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type WorldCupChampionData struct {
	Id   uint64
	Rate float32
	Out  uint64
}

var WorldCupChampion map[uint64]WorldCupChampionData
var WorldCupChampionLock sync.RWMutex

func LoadWorldCupChampion() {
	WorldCupChampionLock.Lock()
	defer WorldCupChampionLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/WorldCupChampion.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	WorldCupChampion = nil
	err = json.Unmarshal(data, &WorldCupChampion)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetWorldCupChampionMap() map[uint64]WorldCupChampionData {
	WorldCupChampionLock.RLock()
	defer WorldCupChampionLock.RUnlock()

	WorldCupChampion2 := make(map[uint64]WorldCupChampionData)
	for k, v := range WorldCupChampion {
		WorldCupChampion2[k] = v
	}

	return WorldCupChampion2
}

func GetWorldCupChampion(key uint64) (WorldCupChampionData, bool) {
	WorldCupChampionLock.RLock()
	defer WorldCupChampionLock.RUnlock()

	val, ok := WorldCupChampion[key]

	return val, ok
}

func GetWorldCupChampionMapLen() int {
	return len(WorldCupChampion)
}

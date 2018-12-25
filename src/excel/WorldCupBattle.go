package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type WorldCupBattleData struct {
	Id          uint64
	Time        string
	TeamA       string
	ScoreA      uint64
	ScoreAfinal uint64
	TeamB       string
	ScoreB      uint64
	ScoreBfinal uint64
	RateA       float32
	RateEqual   float32
	RateB       float32
}

var WorldCupBattle map[uint64]WorldCupBattleData
var WorldCupBattleLock sync.RWMutex

func LoadWorldCupBattle() {
	WorldCupBattleLock.Lock()
	defer WorldCupBattleLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/WorldCupBattle.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	WorldCupBattle = nil
	err = json.Unmarshal(data, &WorldCupBattle)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetWorldCupBattleMap() map[uint64]WorldCupBattleData {
	WorldCupBattleLock.RLock()
	defer WorldCupBattleLock.RUnlock()

	WorldCupBattle2 := make(map[uint64]WorldCupBattleData)
	for k, v := range WorldCupBattle {
		WorldCupBattle2[k] = v
	}

	return WorldCupBattle2
}

func GetWorldCupBattle(key uint64) (WorldCupBattleData, bool) {
	WorldCupBattleLock.RLock()
	defer WorldCupBattleLock.RUnlock()

	val, ok := WorldCupBattle[key]

	return val, ok
}

func GetWorldCupBattleMapLen() int {
	return len(WorldCupBattle)
}

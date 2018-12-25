package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type AchievementLevelData struct {
	Level      uint64
	Experience uint64
	Reward1    []uint32
	Reward2    []uint32
}

var AchievementLevel map[uint64]AchievementLevelData
var AchievementLevelLock sync.RWMutex

func LoadAchievementLevel() {
	AchievementLevelLock.Lock()
	defer AchievementLevelLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/AchievementLevel.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	AchievementLevel = nil
	err = json.Unmarshal(data, &AchievementLevel)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetAchievementLevelMap() map[uint64]AchievementLevelData {
	AchievementLevelLock.RLock()
	defer AchievementLevelLock.RUnlock()

	AchievementLevel2 := make(map[uint64]AchievementLevelData)
	for k, v := range AchievementLevel {
		AchievementLevel2[k] = v
	}

	return AchievementLevel2
}

func GetAchievementLevel(key uint64) (AchievementLevelData, bool) {
	AchievementLevelLock.RLock()
	defer AchievementLevelLock.RUnlock()

	val, ok := AchievementLevel[key]

	return val, ok
}

func GetAchievementLevelMapLen() int {
	return len(AchievementLevel)
}

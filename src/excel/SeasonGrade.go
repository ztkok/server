package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type SeasonGradeData struct {
	Id           uint64
	Value        uint64
	NormalAwards string
	EliteAwards  string
}

var SeasonGrade map[uint64]SeasonGradeData
var SeasonGradeLock sync.RWMutex

func LoadSeasonGrade() {
	SeasonGradeLock.Lock()
	defer SeasonGradeLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/SeasonGrade.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	SeasonGrade = nil
	err = json.Unmarshal(data, &SeasonGrade)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetSeasonGradeMap() map[uint64]SeasonGradeData {
	SeasonGradeLock.RLock()
	defer SeasonGradeLock.RUnlock()

	SeasonGrade2 := make(map[uint64]SeasonGradeData)
	for k, v := range SeasonGrade {
		SeasonGrade2[k] = v
	}

	return SeasonGrade2
}

func GetSeasonGrade(key uint64) (SeasonGradeData, bool) {
	SeasonGradeLock.RLock()
	defer SeasonGradeLock.RUnlock()

	val, ok := SeasonGrade[key]

	return val, ok
}

func GetSeasonGradeMapLen() int {
	return len(SeasonGrade)
}

package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type SkillSystemData struct {
	Id     uint64
	Value1 uint64
	Value2 string
	Name   string
}

var skillSystem map[uint64]SkillSystemData
var skillSystemLock sync.RWMutex

func LoadSkillSystem() {
	skillSystemLock.Lock()
	defer skillSystemLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/skillSystem.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	skillSystem = nil
	err = json.Unmarshal(data, &skillSystem)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetSkillSystemMap() map[uint64]SkillSystemData {
	skillSystemLock.RLock()
	defer skillSystemLock.RUnlock()

	skillSystem2 := make(map[uint64]SkillSystemData)
	for k, v := range skillSystem {
		skillSystem2[k] = v
	}

	return skillSystem2
}

func GetSkillSystem(key uint64) (SkillSystemData, bool) {
	skillSystemLock.RLock()
	defer skillSystemLock.RUnlock()

	val, ok := skillSystem[key]

	return val, ok
}

func GetSkillSystemMapLen() int {
	return len(skillSystem)
}

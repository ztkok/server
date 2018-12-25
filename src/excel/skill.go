package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type SkillData struct {
	Id          uint64
	Name        string
	SkillType   uint64
	SkillRole   string
	SkillShop   uint64
	SkillTarget uint64
	Active      uint64
	Cold        uint64
	SkillEffect string
}

var skill map[uint64]SkillData
var skillLock sync.RWMutex

func LoadSkill() {
	skillLock.Lock()
	defer skillLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/skill.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	skill = nil
	err = json.Unmarshal(data, &skill)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetSkillMap() map[uint64]SkillData {
	skillLock.RLock()
	defer skillLock.RUnlock()

	skill2 := make(map[uint64]SkillData)
	for k, v := range skill {
		skill2[k] = v
	}

	return skill2
}

func GetSkill(key uint64) (SkillData, bool) {
	skillLock.RLock()
	defer skillLock.RUnlock()

	val, ok := skill[key]

	return val, ok
}

func GetSkillMapLen() int {
	return len(skill)
}

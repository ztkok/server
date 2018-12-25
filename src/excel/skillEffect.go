package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type SkillEffectData struct {
	Id    uint64
	Name  string
	Param string
}

var skillEffect map[uint64]SkillEffectData
var skillEffectLock sync.RWMutex

func LoadSkillEffect() {
	skillEffectLock.Lock()
	defer skillEffectLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/skillEffect.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	skillEffect = nil
	err = json.Unmarshal(data, &skillEffect)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetSkillEffectMap() map[uint64]SkillEffectData {
	skillEffectLock.RLock()
	defer skillEffectLock.RUnlock()

	skillEffect2 := make(map[uint64]SkillEffectData)
	for k, v := range skillEffect {
		skillEffect2[k] = v
	}

	return skillEffect2
}

func GetSkillEffect(key uint64) (SkillEffectData, bool) {
	skillEffectLock.RLock()
	defer skillEffectLock.RUnlock()

	val, ok := skillEffect[key]

	return val, ok
}

func GetSkillEffectMapLen() int {
	return len(skillEffect)
}

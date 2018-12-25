package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type SkillmodeData struct {
	Id     uint64
	Single string
	Two    string
	Four   string
}

var skillmode map[uint64]SkillmodeData
var skillmodeLock sync.RWMutex

func LoadSkillmode() {
	skillmodeLock.Lock()
	defer skillmodeLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/skillmode.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	skillmode = nil
	err = json.Unmarshal(data, &skillmode)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetSkillmodeMap() map[uint64]SkillmodeData {
	skillmodeLock.RLock()
	defer skillmodeLock.RUnlock()

	skillmode2 := make(map[uint64]SkillmodeData)
	for k, v := range skillmode {
		skillmode2[k] = v
	}

	return skillmode2
}

func GetSkillmode(key uint64) (SkillmodeData, bool) {
	skillmodeLock.RLock()
	defer skillmodeLock.RUnlock()

	val, ok := skillmode[key]

	return val, ok
}

func GetSkillmodeMapLen() int {
	return len(skillmode)
}

package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type AiData struct {
	Id               uint64
	Roleid           string
	Helmetappearance string
	Bagappearance    string
	Rolename         string
	Aiview           uint64
	Hitheadrate      uint64
	Inititem         string
	Dmin             uint64
	Dmax             uint64
	HideTmin         float32
	HideTmax         float32
	Wrun             uint64
	Wdown            uint64
	AimTmin          uint64
	AimTmax          uint64
	Ptohide          float32
	AccFix           float32
	DetectFix        float32
	AttackAgainFix   float32
	SpawnWeight      uint64
	AccOnMoving      float32
	DetectOnhide     float32
}

var ai map[uint64]AiData
var aiLock sync.RWMutex

func LoadAi() {
	aiLock.Lock()
	defer aiLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/ai.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	ai = nil
	err = json.Unmarshal(data, &ai)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetAiMap() map[uint64]AiData {
	aiLock.RLock()
	defer aiLock.RUnlock()

	ai2 := make(map[uint64]AiData)
	for k, v := range ai {
		ai2[k] = v
	}

	return ai2
}

func GetAi(key uint64) (AiData, bool) {
	aiLock.RLock()
	defer aiLock.RUnlock()

	val, ok := ai[key]

	return val, ok
}

func GetAiMapLen() int {
	return len(ai)
}

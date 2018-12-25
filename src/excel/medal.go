package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type MedalData struct {
	Id         uint64
	Condition  uint64
	Lv2_amount uint64
	Lv3_amount uint64
	Lv4_amount uint64
	Lv5_amount uint64
	Lv1_image  string
	Lv2_image  string
	Lv3_image  string
	Lv4_image  string
	Lv5_image  string
}

var medal map[uint64]MedalData
var medalLock sync.RWMutex

func LoadMedal() {
	medalLock.Lock()
	defer medalLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/medal.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	medal = nil
	err = json.Unmarshal(data, &medal)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetMedalMap() map[uint64]MedalData {
	medalLock.RLock()
	defer medalLock.RUnlock()

	medal2 := make(map[uint64]MedalData)
	for k, v := range medal {
		medal2[k] = v
	}

	return medal2
}

func GetMedal(key uint64) (MedalData, bool) {
	medalLock.RLock()
	defer medalLock.RUnlock()

	val, ok := medal[key]

	return val, ok
}

func GetMedalMapLen() int {
	return len(medal)
}

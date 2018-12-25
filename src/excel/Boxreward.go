package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type BoxrewardData struct {
	Id      uint64
	Reward1 string
	Reward2 string
}

var Boxreward map[uint64]BoxrewardData
var BoxrewardLock sync.RWMutex

func LoadBoxreward() {
	BoxrewardLock.Lock()
	defer BoxrewardLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/Boxreward.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	Boxreward = nil
	err = json.Unmarshal(data, &Boxreward)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetBoxrewardMap() map[uint64]BoxrewardData {
	BoxrewardLock.RLock()
	defer BoxrewardLock.RUnlock()

	Boxreward2 := make(map[uint64]BoxrewardData)
	for k, v := range Boxreward {
		Boxreward2[k] = v
	}

	return Boxreward2
}

func GetBoxreward(key uint64) (BoxrewardData, bool) {
	BoxrewardLock.RLock()
	defer BoxrewardLock.RUnlock()

	val, ok := Boxreward[key]

	return val, ok
}

func GetBoxrewardMapLen() int {
	return len(Boxreward)
}

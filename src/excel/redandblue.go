package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type RedandblueData struct {
	Id     uint64
	Single string
}

var redandblue map[uint64]RedandblueData
var redandblueLock sync.RWMutex

func LoadRedandblue() {
	redandblueLock.Lock()
	defer redandblueLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/redandblue.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	redandblue = nil
	err = json.Unmarshal(data, &redandblue)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetRedandblueMap() map[uint64]RedandblueData {
	redandblueLock.RLock()
	defer redandblueLock.RUnlock()

	redandblue2 := make(map[uint64]RedandblueData)
	for k, v := range redandblue {
		redandblue2[k] = v
	}

	return redandblue2
}

func GetRedandblue(key uint64) (RedandblueData, bool) {
	redandblueLock.RLock()
	defer redandblueLock.RUnlock()

	val, ok := redandblue[key]

	return val, ok
}

func GetRedandblueMapLen() int {
	return len(redandblue)
}

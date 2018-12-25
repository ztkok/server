package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type FestivalData struct {
	Id         uint64
	Flush      uint64
	Taskpool   uint64
	Requirenum uint64
	Awards     string
}

var festival map[uint64]FestivalData
var festivalLock sync.RWMutex

func LoadFestival() {
	festivalLock.Lock()
	defer festivalLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/festival.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	festival = nil
	err = json.Unmarshal(data, &festival)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetFestivalMap() map[uint64]FestivalData {
	festivalLock.RLock()
	defer festivalLock.RUnlock()

	festival2 := make(map[uint64]FestivalData)
	for k, v := range festival {
		festival2[k] = v
	}

	return festival2
}

func GetFestival(key uint64) (FestivalData, bool) {
	festivalLock.RLock()
	defer festivalLock.RUnlock()

	val, ok := festival[key]

	return val, ok
}

func GetFestivalMapLen() int {
	return len(festival)
}

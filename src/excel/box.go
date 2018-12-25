package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type BoxData struct {
	Id          uint64
	Name        string
	StartDate   string
	EndDate     string
	Refreshrate uint64
	Money       string
	Pond        string
	Item        string
}

var box map[uint64]BoxData
var boxLock sync.RWMutex

func LoadBox() {
	boxLock.Lock()
	defer boxLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/box.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	box = nil
	err = json.Unmarshal(data, &box)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetBoxMap() map[uint64]BoxData {
	boxLock.RLock()
	defer boxLock.RUnlock()

	box2 := make(map[uint64]BoxData)
	for k, v := range box {
		box2[k] = v
	}

	return box2
}

func GetBox(key uint64) (BoxData, bool) {
	boxLock.RLock()
	defer boxLock.RUnlock()

	val, ok := box[key]

	return val, ok
}

func GetBoxMapLen() int {
	return len(box)
}

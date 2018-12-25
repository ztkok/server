package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type ChangenameData struct {
	Times    uint64
	Length   uint64
	Infotxt  string
	Cost     uint64
	CloseBtn uint64
}

var changename map[uint64]ChangenameData
var changenameLock sync.RWMutex

func LoadChangename() {
	changenameLock.Lock()
	defer changenameLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/changename.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	changename = nil
	err = json.Unmarshal(data, &changename)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetChangenameMap() map[uint64]ChangenameData {
	changenameLock.RLock()
	defer changenameLock.RUnlock()

	changename2 := make(map[uint64]ChangenameData)
	for k, v := range changename {
		changename2[k] = v
	}

	return changename2
}

func GetChangename(key uint64) (ChangenameData, bool) {
	changenameLock.RLock()
	defer changenameLock.RUnlock()

	val, ok := changename[key]

	return val, ok
}

func GetChangenameMapLen() int {
	return len(changename)
}

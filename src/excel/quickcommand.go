package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type QuickcommandData struct {
	Id          uint64
	Talkcontent string
	Sound       string
}

var quickcommand map[uint64]QuickcommandData
var quickcommandLock sync.RWMutex

func LoadQuickcommand() {
	quickcommandLock.Lock()
	defer quickcommandLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/quickcommand.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	quickcommand = nil
	err = json.Unmarshal(data, &quickcommand)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetQuickcommandMap() map[uint64]QuickcommandData {
	quickcommandLock.RLock()
	defer quickcommandLock.RUnlock()

	quickcommand2 := make(map[uint64]QuickcommandData)
	for k, v := range quickcommand {
		quickcommand2[k] = v
	}

	return quickcommand2
}

func GetQuickcommand(key uint64) (QuickcommandData, bool) {
	quickcommandLock.RLock()
	defer quickcommandLock.RUnlock()

	val, ok := quickcommand[key]

	return val, ok
}

func GetQuickcommandMapLen() int {
	return len(quickcommand)
}

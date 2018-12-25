package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type Cheater_reportData struct {
	Type uint64
	Desc string
}

var cheater_report map[uint64]Cheater_reportData
var cheater_reportLock sync.RWMutex

func LoadCheater_report() {
	cheater_reportLock.Lock()
	defer cheater_reportLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/cheater_report.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	cheater_report = nil
	err = json.Unmarshal(data, &cheater_report)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetCheater_reportMap() map[uint64]Cheater_reportData {
	cheater_reportLock.RLock()
	defer cheater_reportLock.RUnlock()

	cheater_report2 := make(map[uint64]Cheater_reportData)
	for k, v := range cheater_report {
		cheater_report2[k] = v
	}

	return cheater_report2
}

func GetCheater_report(key uint64) (Cheater_reportData, bool) {
	cheater_reportLock.RLock()
	defer cheater_reportLock.RUnlock()

	val, ok := cheater_report[key]

	return val, ok
}

func GetCheater_reportMapLen() int {
	return len(cheater_report)
}

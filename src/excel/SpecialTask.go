package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type SpecialTaskData struct {
	Id         uint64
	Taskpool   uint64
	Groupid    uint64
	Tasktype   uint64
	Starttime  string
	Requirenum uint64
	Awards     string
}

var SpecialTask map[uint64]SpecialTaskData
var SpecialTaskLock sync.RWMutex

func LoadSpecialTask() {
	SpecialTaskLock.Lock()
	defer SpecialTaskLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/SpecialTask.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	SpecialTask = nil
	err = json.Unmarshal(data, &SpecialTask)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetSpecialTaskMap() map[uint64]SpecialTaskData {
	SpecialTaskLock.RLock()
	defer SpecialTaskLock.RUnlock()

	SpecialTask2 := make(map[uint64]SpecialTaskData)
	for k, v := range SpecialTask {
		SpecialTask2[k] = v
	}

	return SpecialTask2
}

func GetSpecialTask(key uint64) (SpecialTaskData, bool) {
	SpecialTaskLock.RLock()
	defer SpecialTaskLock.RUnlock()

	val, ok := SpecialTask[key]

	return val, ok
}

func GetSpecialTaskMapLen() int {
	return len(SpecialTask)
}

package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type ComradeTaskData struct {
	Id         uint64
	Taskpool   uint64
	Groupid    uint64
	Requirenum uint64
	Starttime  string
	Awards     string
}

var ComradeTask map[uint64]ComradeTaskData
var ComradeTaskLock sync.RWMutex

func LoadComradeTask() {
	ComradeTaskLock.Lock()
	defer ComradeTaskLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/ComradeTask.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	ComradeTask = nil
	err = json.Unmarshal(data, &ComradeTask)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetComradeTaskMap() map[uint64]ComradeTaskData {
	ComradeTaskLock.RLock()
	defer ComradeTaskLock.RUnlock()

	ComradeTask2 := make(map[uint64]ComradeTaskData)
	for k, v := range ComradeTask {
		ComradeTask2[k] = v
	}

	return ComradeTask2
}

func GetComradeTask(key uint64) (ComradeTaskData, bool) {
	ComradeTaskLock.RLock()
	defer ComradeTaskLock.RUnlock()

	val, ok := ComradeTask[key]

	return val, ok
}

func GetComradeTaskMapLen() int {
	return len(ComradeTask)
}

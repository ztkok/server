package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type TaskData struct {
	Id         uint64
	Requirenum uint64
	Taskpool   uint64
	Awards     string
}

var task map[uint64]TaskData
var taskLock sync.RWMutex

func LoadTask() {
	taskLock.Lock()
	defer taskLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/task.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	task = nil
	err = json.Unmarshal(data, &task)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetTaskMap() map[uint64]TaskData {
	taskLock.RLock()
	defer taskLock.RUnlock()

	task2 := make(map[uint64]TaskData)
	for k, v := range task {
		task2[k] = v
	}

	return task2
}

func GetTask(key uint64) (TaskData, bool) {
	taskLock.RLock()
	defer taskLock.RUnlock()

	val, ok := task[key]

	return val, ok
}

func GetTaskMapLen() int {
	return len(task)
}

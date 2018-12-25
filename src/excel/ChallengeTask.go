package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type ChallengeTaskData struct {
	Id         uint64
	Taskpool   uint64
	Requirenum uint64
	Awards     string
	Pic        string
}

var ChallengeTask map[uint64]ChallengeTaskData
var ChallengeTaskLock sync.RWMutex

func LoadChallengeTask() {
	ChallengeTaskLock.Lock()
	defer ChallengeTaskLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/ChallengeTask.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	ChallengeTask = nil
	err = json.Unmarshal(data, &ChallengeTask)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetChallengeTaskMap() map[uint64]ChallengeTaskData {
	ChallengeTaskLock.RLock()
	defer ChallengeTaskLock.RUnlock()

	ChallengeTask2 := make(map[uint64]ChallengeTaskData)
	for k, v := range ChallengeTask {
		ChallengeTask2[k] = v
	}

	return ChallengeTask2
}

func GetChallengeTask(key uint64) (ChallengeTaskData, bool) {
	ChallengeTaskLock.RLock()
	defer ChallengeTaskLock.RUnlock()

	val, ok := ChallengeTask[key]

	return val, ok
}

func GetChallengeTaskMapLen() int {
	return len(ChallengeTask)
}

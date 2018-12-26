package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type QuestData struct {
	Id     uint64
	Quest  string
	Answer string
}

var Quest map[uint64]QuestData
var QuestLock sync.RWMutex

func LoadQuest() {
	QuestLock.Lock()
	defer QuestLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/Quest.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	Quest = nil
	err = json.Unmarshal(data, &Quest)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetQuestMap() map[uint64]QuestData {
	QuestLock.RLock()
	defer QuestLock.RUnlock()

	Quest2 := make(map[uint64]QuestData)
	for k, v := range Quest {
		Quest2[k] = v
	}

	return Quest2
}

func GetQuest(key uint64) (QuestData, bool) {
	QuestLock.RLock()
	defer QuestLock.RUnlock()

	val, ok := Quest[key]

	return val, ok
}

func GetQuestMapLen() int {
	return len(Quest)
}

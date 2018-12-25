package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type ActivityData struct {
	Id          uint64
	BonusID     uint64
	BonusNum    uint64
	ActivityNum uint64
}

var activity map[uint64]ActivityData
var activityLock sync.RWMutex

func LoadActivity() {
	activityLock.Lock()
	defer activityLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/activity.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	activity = nil
	err = json.Unmarshal(data, &activity)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetActivityMap() map[uint64]ActivityData {
	activityLock.RLock()
	defer activityLock.RUnlock()

	activity2 := make(map[uint64]ActivityData)
	for k, v := range activity {
		activity2[k] = v
	}

	return activity2
}

func GetActivity(key uint64) (ActivityData, bool) {
	activityLock.RLock()
	defer activityLock.RUnlock()

	val, ok := activity[key]

	return val, ok
}

func GetActivityMapLen() int {
	return len(activity)
}

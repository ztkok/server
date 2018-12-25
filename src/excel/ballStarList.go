package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type BallStarListData struct {
	Id      uint64
	Num     uint64
	Rewards string
}

var ballStarList map[uint64]BallStarListData
var ballStarListLock sync.RWMutex

func LoadBallStarList() {
	ballStarListLock.Lock()
	defer ballStarListLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/ballStarList.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	ballStarList = nil
	err = json.Unmarshal(data, &ballStarList)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetBallStarListMap() map[uint64]BallStarListData {
	ballStarListLock.RLock()
	defer ballStarListLock.RUnlock()

	ballStarList2 := make(map[uint64]BallStarListData)
	for k, v := range ballStarList {
		ballStarList2[k] = v
	}

	return ballStarList2
}

func GetBallStarList(key uint64) (BallStarListData, bool) {
	ballStarListLock.RLock()
	defer ballStarListLock.RUnlock()

	val, ok := ballStarList[key]

	return val, ok
}

func GetBallStarListMapLen() int {
	return len(ballStarList)
}

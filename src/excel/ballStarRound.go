package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type BallStarRoundData struct {
	Id         uint64
	RewardType uint64
	MiniNum    uint64
	Weight     uint64
	Rewards    string
}

var ballStarRound map[uint64]BallStarRoundData
var ballStarRoundLock sync.RWMutex

func LoadBallStarRound() {
	ballStarRoundLock.Lock()
	defer ballStarRoundLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/ballStarRound.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	ballStarRound = nil
	err = json.Unmarshal(data, &ballStarRound)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetBallStarRoundMap() map[uint64]BallStarRoundData {
	ballStarRoundLock.RLock()
	defer ballStarRoundLock.RUnlock()

	ballStarRound2 := make(map[uint64]BallStarRoundData)
	for k, v := range ballStarRound {
		ballStarRound2[k] = v
	}

	return ballStarRound2
}

func GetBallStarRound(key uint64) (BallStarRoundData, bool) {
	ballStarRoundLock.RLock()
	defer ballStarRoundLock.RUnlock()

	val, ok := ballStarRound[key]

	return val, ok
}

func GetBallStarRoundMapLen() int {
	return len(ballStarRound)
}

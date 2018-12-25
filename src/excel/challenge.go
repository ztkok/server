package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type ChallengeData struct {
	Id          uint64
	Name        string
	Season      uint64
	GroupId     uint64
	Unlock      uint64
	UnlockValue uint64
	UnlockWord  string
	Task        string
	Pic         string
}

var challenge map[uint64]ChallengeData
var challengeLock sync.RWMutex

func LoadChallenge() {
	challengeLock.Lock()
	defer challengeLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/challenge.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	challenge = nil
	err = json.Unmarshal(data, &challenge)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetChallengeMap() map[uint64]ChallengeData {
	challengeLock.RLock()
	defer challengeLock.RUnlock()

	challenge2 := make(map[uint64]ChallengeData)
	for k, v := range challenge {
		challenge2[k] = v
	}

	return challenge2
}

func GetChallenge(key uint64) (ChallengeData, bool) {
	challengeLock.RLock()
	defer challengeLock.RUnlock()

	val, ok := challenge[key]

	return val, ok
}

func GetChallengeMapLen() int {
	return len(challenge)
}

package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type MatchData struct {
	Id    uint64
	Value float32
	Name  string
}

var match map[uint64]MatchData
var matchLock sync.RWMutex

func LoadMatch() {
	matchLock.Lock()
	defer matchLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/match.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	match = nil
	err = json.Unmarshal(data, &match)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetMatchMap() map[uint64]MatchData {
	matchLock.RLock()
	defer matchLock.RUnlock()

	match2 := make(map[uint64]MatchData)
	for k, v := range match {
		match2[k] = v
	}

	return match2
}

func GetMatch(key uint64) (MatchData, bool) {
	matchLock.RLock()
	defer matchLock.RUnlock()

	val, ok := match[key]

	return val, ok
}

func GetMatchMapLen() int {
	return len(match)
}

package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type MatchmodeData struct {
	Id              uint64
	Name            string
	State           uint64
	Opentime        string
	Closetime       string
	Opennum         uint64
	Looptime        uint64
	Modeid          uint64
	Seasonopentime  string
	Seasonclosetime string
	Itemid          uint64
	Itemnum         uint64
	Itemprice       uint64
	Matchnum        uint64
	Solorank        uint64
	Duorank         uint64
	Squadrank       uint64
	Waittime        uint64
	Mapid           uint64
	Itemrule        string
}

var matchmode map[uint64]MatchmodeData
var matchmodeLock sync.RWMutex

func LoadMatchmode() {
	matchmodeLock.Lock()
	defer matchmodeLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/matchmode.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	matchmode = nil
	err = json.Unmarshal(data, &matchmode)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetMatchmodeMap() map[uint64]MatchmodeData {
	matchmodeLock.RLock()
	defer matchmodeLock.RUnlock()

	matchmode2 := make(map[uint64]MatchmodeData)
	for k, v := range matchmode {
		matchmode2[k] = v
	}

	return matchmode2
}

func GetMatchmode(key uint64) (MatchmodeData, bool) {
	matchmodeLock.RLock()
	defer matchmodeLock.RUnlock()

	val, ok := matchmode[key]

	return val, ok
}

func GetMatchmodeMapLen() int {
	return len(matchmode)
}

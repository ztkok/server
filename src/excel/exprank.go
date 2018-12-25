package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type ExprankData struct {
	Id     uint64
	Single uint64
	Two    uint64
	Four   uint64
}

var exprank map[uint64]ExprankData
var exprankLock sync.RWMutex

func LoadExprank() {
	exprankLock.Lock()
	defer exprankLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/exprank.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	exprank = nil
	err = json.Unmarshal(data, &exprank)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetExprankMap() map[uint64]ExprankData {
	exprankLock.RLock()
	defer exprankLock.RUnlock()

	exprank2 := make(map[uint64]ExprankData)
	for k, v := range exprank {
		exprank2[k] = v
	}

	return exprank2
}

func GetExprank(key uint64) (ExprankData, bool) {
	exprankLock.RLock()
	defer exprankLock.RUnlock()

	val, ok := exprank[key]

	return val, ok
}

func GetExprankMapLen() int {
	return len(exprank)
}

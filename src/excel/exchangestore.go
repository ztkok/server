package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type ExchangestoreData struct {
	Id           uint64
	Activityid   uint64
	ExchangeA    string
	ExchangeB    string
	ExchangeMax  uint64
	ExchangeTips uint64
	ExchangeList uint64
}

var exchangestore map[uint64]ExchangestoreData
var exchangestoreLock sync.RWMutex

func LoadExchangestore() {
	exchangestoreLock.Lock()
	defer exchangestoreLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/exchangestore.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	exchangestore = nil
	err = json.Unmarshal(data, &exchangestore)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetExchangestoreMap() map[uint64]ExchangestoreData {
	exchangestoreLock.RLock()
	defer exchangestoreLock.RUnlock()

	exchangestore2 := make(map[uint64]ExchangestoreData)
	for k, v := range exchangestore {
		exchangestore2[k] = v
	}

	return exchangestore2
}

func GetExchangestore(key uint64) (ExchangestoreData, bool) {
	exchangestoreLock.RLock()
	defer exchangestoreLock.RUnlock()

	val, ok := exchangestore[key]

	return val, ok
}

func GetExchangestoreMapLen() int {
	return len(exchangestore)
}

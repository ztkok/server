package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type SellingData struct {
	Id        uint64
	Type      uint64
	Sontype   uint64
	GoodsID   uint64
	Goods     string
	Price1    string
	Discount1 string
	Price2    string
	Discount2 string
	Pic       string
}

var selling map[uint64]SellingData
var sellingLock sync.RWMutex

func LoadSelling() {
	sellingLock.Lock()
	defer sellingLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/selling.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	selling = nil
	err = json.Unmarshal(data, &selling)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetSellingMap() map[uint64]SellingData {
	sellingLock.RLock()
	defer sellingLock.RUnlock()

	selling2 := make(map[uint64]SellingData)
	for k, v := range selling {
		selling2[k] = v
	}

	return selling2
}

func GetSelling(key uint64) (SellingData, bool) {
	sellingLock.RLock()
	defer sellingLock.RUnlock()

	val, ok := selling[key]

	return val, ok
}

func GetSellingMapLen() int {
	return len(selling)
}

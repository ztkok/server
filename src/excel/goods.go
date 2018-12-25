package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type GoodsData struct {
	Id        uint64
	Group     uint64
	GoodsID   uint64
	Goods     string
	Price1    string
	Discount1 string
	Price2    string
	Discount2 string
	ShowType  uint64
	Pic       string
	Picsize   string
	Icon      string
	Name      string
}

var goods map[uint64]GoodsData
var goodsLock sync.RWMutex

func LoadGoods() {
	goodsLock.Lock()
	defer goodsLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/goods.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	goods = nil
	err = json.Unmarshal(data, &goods)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetGoodsMap() map[uint64]GoodsData {
	goodsLock.RLock()
	defer goodsLock.RUnlock()

	goods2 := make(map[uint64]GoodsData)
	for k, v := range goods {
		goods2[k] = v
	}

	return goods2
}

func GetGoods(key uint64) (GoodsData, bool) {
	goodsLock.RLock()
	defer goodsLock.RUnlock()

	val, ok := goods[key]

	return val, ok
}

func GetGoodsMapLen() int {
	return len(goods)
}

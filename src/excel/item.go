package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type ItemData struct {
	Id                    uint64
	Type                  uint64
	Subtype               uint64
	Addcell               uint64
	Additem               string
	Reducerate            uint64
	Reducedam             uint64
	Effect                uint64
	Effectparam           uint64
	Addpack               uint64
	Addlimit              uint64
	Uselimit              uint64
	Weaponindex           uint64
	Reward                uint64
	Lensoffsetscale       float32
	Headratedelta         uint64
	Changebullettimescale float32
	Processtime           float32
	ThrowDamage           float32
	ThrowHurtRadius       float32
	MarkAoiRange          float32
	HpCondition           uint64
	Light                 uint64
	AutoPickup            uint64
	IsGreatWarrior        uint64
	AddValue              string
	GunCustom             uint64
	ShowOff               uint64
}

var item map[uint64]ItemData
var itemLock sync.RWMutex

func LoadItem() {
	itemLock.Lock()
	defer itemLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/item.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	item = nil
	err = json.Unmarshal(data, &item)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetItemMap() map[uint64]ItemData {
	itemLock.RLock()
	defer itemLock.RUnlock()

	item2 := make(map[uint64]ItemData)
	for k, v := range item {
		item2[k] = v
	}

	return item2
}

func GetItem(key uint64) (ItemData, bool) {
	itemLock.RLock()
	defer itemLock.RUnlock()

	val, ok := item[key]

	return val, ok
}

func GetItemMapLen() int {
	return len(item)
}

package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type SeasontikData struct {
	Id         uint64
	Goods      string
	Price      string
	Discount   string
	Smallpic   string
	Bigpic     string
	Name       string
	Shortwords string
	Content    string
}

var Seasontik map[uint64]SeasontikData
var SeasontikLock sync.RWMutex

func LoadSeasontik() {
	SeasontikLock.Lock()
	defer SeasontikLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/Seasontik.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	Seasontik = nil
	err = json.Unmarshal(data, &Seasontik)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetSeasontikMap() map[uint64]SeasontikData {
	SeasontikLock.RLock()
	defer SeasontikLock.RUnlock()

	Seasontik2 := make(map[uint64]SeasontikData)
	for k, v := range Seasontik {
		Seasontik2[k] = v
	}

	return Seasontik2
}

func GetSeasontik(key uint64) (SeasontikData, bool) {
	SeasontikLock.RLock()
	defer SeasontikLock.RUnlock()

	val, ok := Seasontik[key]

	return val, ok
}

func GetSeasontikMapLen() int {
	return len(Seasontik)
}

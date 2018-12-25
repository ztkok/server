package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type PromotionData struct {
	Id                   uint64
	StartshowDate        string
	EndshowDate          string
	StartDate            string
	EndDate              string
	Name                 string
	Des                  string
	Days                 string
	IosrewardVersion     string
	AndroidrewardVersion string
	Activitytype         uint64
}

var Promotion map[uint64]PromotionData
var PromotionLock sync.RWMutex

func LoadPromotion() {
	PromotionLock.Lock()
	defer PromotionLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/Promotion.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	Promotion = nil
	err = json.Unmarshal(data, &Promotion)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetPromotionMap() map[uint64]PromotionData {
	PromotionLock.RLock()
	defer PromotionLock.RUnlock()

	Promotion2 := make(map[uint64]PromotionData)
	for k, v := range Promotion {
		Promotion2[k] = v
	}

	return Promotion2
}

func GetPromotion(key uint64) (PromotionData, bool) {
	PromotionLock.RLock()
	defer PromotionLock.RUnlock()

	val, ok := Promotion[key]

	return val, ok
}

func GetPromotionMapLen() int {
	return len(Promotion)
}

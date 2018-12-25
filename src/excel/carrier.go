package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type CarrierData struct {
	Id              uint64
	Name            string
	Subtype         uint64
	Seat            uint64
	InitFuelRange   string
	MaxFuel         float32
	Consumption     float32
	Factor          string
	InitShellsRange string
	MaxShells       uint64
	AngleLimit      float32
	NoDamage        string
	Weapon          uint64
	Telescope       uint64
}

var carrier map[uint64]CarrierData
var carrierLock sync.RWMutex

func LoadCarrier() {
	carrierLock.Lock()
	defer carrierLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/carrier.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	carrier = nil
	err = json.Unmarshal(data, &carrier)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetCarrierMap() map[uint64]CarrierData {
	carrierLock.RLock()
	defer carrierLock.RUnlock()

	carrier2 := make(map[uint64]CarrierData)
	for k, v := range carrier {
		carrier2[k] = v
	}

	return carrier2
}

func GetCarrier(key uint64) (CarrierData, bool) {
	carrierLock.RLock()
	defer carrierLock.RUnlock()

	val, ok := carrier[key]

	return val, ok
}

func GetCarrierMapLen() int {
	return len(carrier)
}

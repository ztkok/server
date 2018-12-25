package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type GunData struct {
	Id                  uint64
	Name                string
	Consumebullet       uint64
	Changeonebullettime float32
	Changebullettime    float32
	Changebullettime2   float32
	Shootinterval1      float32
	Shootinterval2      float32
	Shootinterval3      float32
	Shootinterval4      float32
	Shootinterval5      float32
	Shootinterval6      float32
	Shootinterval7      float32
	Shootinterval8      float32
	Bulletnum           int
	Hrecoilscale        float32
	Vrecoilscale        float32
	Offset              float32
	Distance            float32
	Reducedistance      float32
	Headshotrate        uint64
	Clipcap             uint64
	Hurt                float32
	Headhurt            float32
	Reformitems         string
	Backupscope         uint64
	Defaultscope        uint64
	Bmin                uint64
	Bmax                uint64
	ContinueTrh         float32
	AiDamageRate        float32
	Aishotinterval      float32
	Takeintime          float32
	Takeouttime         float32
	Closelenstime       float32
	Shotheadlimit       uint64
	GunID               uint64
	DoubleSightID       uint64
	Magazinedelta       string
}

var gun map[uint64]GunData
var gunLock sync.RWMutex

func LoadGun() {
	gunLock.Lock()
	defer gunLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/gun.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	gun = nil
	err = json.Unmarshal(data, &gun)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetGunMap() map[uint64]GunData {
	gunLock.RLock()
	defer gunLock.RUnlock()

	gun2 := make(map[uint64]GunData)
	for k, v := range gun {
		gun2[k] = v
	}

	return gun2
}

func GetGun(key uint64) (GunData, bool) {
	gunLock.RLock()
	defer gunLock.RUnlock()

	val, ok := gun[key]

	return val, ok
}

func GetGunMapLen() int {
	return len(gun)
}

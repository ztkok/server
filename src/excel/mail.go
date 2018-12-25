package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type MailData struct {
	MailId    uint64
	MailTitle string
	Mail      string
}

var mail map[uint64]MailData
var mailLock sync.RWMutex

func LoadMail() {
	mailLock.Lock()
	defer mailLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/mail.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	mail = nil
	err = json.Unmarshal(data, &mail)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetMailMap() map[uint64]MailData {
	mailLock.RLock()
	defer mailLock.RUnlock()

	mail2 := make(map[uint64]MailData)
	for k, v := range mail {
		mail2[k] = v
	}

	return mail2
}

func GetMail(key uint64) (MailData, bool) {
	mailLock.RLock()
	defer mailLock.RUnlock()

	val, ok := mail[key]

	return val, ok
}

func GetMailMapLen() int {
	return len(mail)
}

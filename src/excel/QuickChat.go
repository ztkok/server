package excel

import (
	"encoding/json"
	"io/ioutil"
	"sync"

	log "github.com/cihub/seelog"
)

type QuickChatData struct {
	Id uint64
}

var QuickChat map[uint64]QuickChatData
var QuickChatLock sync.RWMutex

func LoadQuickChat() {
	QuickChatLock.Lock()
	defer QuickChatLock.Unlock()

	data, err := ioutil.ReadFile("../res/excel/QuickChat.json")
	if err != nil {
		log.Error("ReadFile err: ", err)
		return
	}

	QuickChat = nil
	err = json.Unmarshal(data, &QuickChat)
	if err != nil {
		log.Error("Unmarshal err: ", err)
		return
	}
}

func GetQuickChatMap() map[uint64]QuickChatData {
	QuickChatLock.RLock()
	defer QuickChatLock.RUnlock()

	QuickChat2 := make(map[uint64]QuickChatData)
	for k, v := range QuickChat {
		QuickChat2[k] = v
	}

	return QuickChat2
}

func GetQuickChat(key uint64) (QuickChatData, bool) {
	QuickChatLock.RLock()
	defer QuickChatLock.RUnlock()

	val, ok := QuickChat[key]

	return val, ok
}

func GetQuickChatMapLen() int {
	return len(QuickChat)
}

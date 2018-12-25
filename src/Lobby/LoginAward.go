package main

import (
	"common"
	"db"
	"excel"
	"strings"
	"time"
	"zeus/iserver"
)

func (user *LobbyUser) loginAward() {
	if user.GetLoginChannel() == 1 {
		if canGetLoginAward(user, "wxlogin") {
			setLoginAward(user, "wxlogin")
			sendLoginMail(user, 1)
		}

	} else if user.GetLoginChannel() == 2 {
		if canGetLoginAward(user, "qqlogin") {
			setLoginAward(user, "qqlogin")
			sendLoginMail(user, 2)
		}

		if user.GetQQVIP() == 1 {
			if canGetLoginAward(user, "qqviplogin") {
				setLoginAward(user, "qqviplogin")
				sendLoginMail(user, 3)
			}

		} else if user.GetQQVIP() == 2 {
			if canGetLoginAward(user, "qqsviplogin") {
				setLoginAward(user, "qqsviplogin")
				sendLoginMail(user, 4)
			}
		}
	}
}

func canGetLoginAward(user *LobbyUser, str string) bool {
	lastlogin, err := db.PlayerInfoUtil(user.GetDBID()).GetLoginAward(str)
	if err != nil {
		return false
	}

	lasttime := time.Unix(lastlogin, 0)
	cur := time.Now()
	if lasttime.Year() == cur.Year() && lasttime.YearDay() == cur.YearDay() {
		return false
	}

	return true
}

func setLoginAward(user *LobbyUser, str string) {
	db.PlayerInfoUtil(user.GetDBID()).SetLoginAward(str, time.Now().Unix())
}

func sendLoginMail(user *LobbyUser, baseid uint64) {
	base, ok := excel.GetDetail(baseid)
	if !ok {
		return
	}

	item := strings.Split(base.Dailylogin, ";")
	if len(item) != 2 {
		return
	}
	id := common.StringToUint32(item[0])
	num := common.StringToUint32(item[1])

	objs := make(map[uint32]uint32)
	objs[id] = num
	sendObjMail(user.GetDBID(), "", 0, base.DailyLoginMailTitle, base.DailyLoginMail, "", "", objs)
	user.RPC(iserver.ServerTypeClient, "AddNewMail")
}

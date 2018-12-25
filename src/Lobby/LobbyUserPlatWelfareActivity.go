package main

import (
	"zeus/iserver"
)

/*--------------------------------平台福利活动------------------------------------*/

// syncPlatWelfareActivity 同步平台福利活动信息
func (a *ActivityMgr) syncPlatWelfareActivity() {
	a.user.Debug("syncPlatWelfareActivity")
	if err := a.user.RPC(iserver.ServerTypeClient, "LimitTencent"); err != nil {
		a.user.Error(err)
	}
}

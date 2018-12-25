package main

import (
	pay "Pay"
	"common"
	"db"
	"encoding/json"
	"excel"
	"fmt"
	"protoMsg"
	"time"
	"zeus/iserver"
)

const (
	BrickRMBRatio = 10
)

// RPC_QueryBalanceReq 玩家查询充值货币数量
func (proc *LobbyUserMsgProc) RPC_QueryBalanceReq(param []byte) {
	proc.user.Debug("QueryBalanceReq: ", string(param))

	var info *pay.MidasQueryBalanceResult
	var os string
	var err error

	for i := 0; i < 5; i++ {
		info, os, err = pay.QueryBalance(param, GetSrvInst().msdkAddr)
		if err != nil {
			proc.user.Error("QueryBalance err: ", err)
			continue
		}

		if info.Ret == 0 {
			break
		}
	}

	d, e := json.Marshal(info)
	if e != nil {
		proc.user.Error("Marshal Error ", e)
		return
	}

	proc.user.Debugf("info： %+v\n", info)
	proc.user.RPC(iserver.ServerTypeClient, "QueryBalanceRet", string(d))

	if info == nil {
		proc.user.Error("info is nil! err:", err)
		return
	}

	if info.Ret == 0 {
		if len(proc.user.GetPayOS()) == 0 {
			proc.user.SetPayOS(os)
			proc.user.syncMonthCard()
		}

		//更新账户余额
		amount := proc.user.getPayAmount(info, os)
		proc.user.updateBalance(amount, 0, common.RS_Pay, os)

		//首次充值
		if proc.user.isFirstPay(amount) {
			proc.user.firstPay()
		}

		//首次充送
		level := proc.user.getFirstPayLevel(amount)
		if level != 0 {
			proc.user.firstPayLevel(param, level)
		}

		//购买月卡
		if proc.user.isBuyNewMonthCard(info.Tss_list) {
			proc.user.buyMonthCard(param, info.Tss_list)
		}
	}
}

// RPC_ConsumeBalanceReq 玩家消耗充值货币
func (proc *LobbyUserMsgProc) RPC_ConsumeBalanceReq(param []byte) {
	proc.user.Debug("ConsumeBalanceReq: ", string(param))

	//待修复
	// amount := uint32(0)
	// ret, _ := proc.user.consumeBalance(param, amount, common.RS_BuyGoods)
	// proc.user.RPC(iserver.ServerTypeClient, "ConsumeBalanceRet", ret, proc.user.GetBrick())
}

// RPC_GetPresentReq 玩家领取充值相关奖励
func (proc *LobbyUserMsgProc) RPC_GetPresentReq(param []byte, typ uint32) {
	proc.user.Debug("GetPresentReq: ", string(param), " typ: ", typ)

	ret := uint32(0)
	msg := ""

	defer func() {
		proc.user.RPC(iserver.ServerTypeClient, "GetPresentRet", ret, msg, typ)
	}()

	awards := make(map[uint32]uint32)
	reason := uint32(0)

	switch typ {
	case common.MarketingTypeFirstPay:
		{
			awards = proc.user.getFirstPayAwards()
			reason = common.RS_Pay
		}
	case common.MarketingTypeMonthCard:
		{
			awards = proc.user.getMonthCardAwards()
			reason = common.RS_MonthCard
		}
	}

	if len(awards) == 0 {
		ret = 1
		return
	}

	bricks := awards[common.Item_Brick]
	delete(awards, common.Item_Brick)

	if bricks > 0 {
		ret, msg = proc.user.presentBricks(param, bricks, reason)
	}

	if ret == 0 {
		if len(awards) > 0 {
			proc.user.storeMgr.GetAwards(awards, reason, false, false)
		}

		switch typ {
		case common.MarketingTypeFirstPay:
			proc.user.setFirstPayAwardsDrawed()
		case common.MarketingTypeMonthCard:
			proc.user.setMonthCardAwardsDrawed()
		}
	}
}

// getPayAmount 获取玩家的充值金额
func (user *LobbyUser) getPayAmount(info *pay.MidasQueryBalanceResult, os string) uint32 {
	var amount uint32

	if os == "iap" {
		if info.Balance >= uint32(user.GetBrick1()) {
			amount = info.Balance - uint32(user.GetBrick1())
		} else {
			user.adjustBalance(info.Balance, os)
		}
	} else if os == "android" {
		if info.Balance >= uint32(user.GetBrick2()) {
			amount = info.Balance - uint32(user.GetBrick2())
		} else {
			user.adjustBalance(info.Balance, os)
		}
	}

	return amount
}

// adjustBalance 当数据库记录的账户余额异常时，使用米大师记录的数据校正数据库的记录。
func (user *LobbyUser) adjustBalance(balance uint32, os string) {
	if os == "iap" {
		user.SetBrick1(uint64(balance))
		user.SetBrick(uint64(balance))
	} else if os == "android" {
		user.SetBrick2(uint64(balance))
		user.SetBrick(uint64(balance))
	}
}

// updateBalance 更新账户余额
// oper表示操作类型 0表示增加 1表示减少
func (user *LobbyUser) updateBalance(amount, oper, reason uint32, os string) {
	if os == "iap" {
		if oper == 0 {
			user.SetBrick1(user.GetBrick1() + uint64(amount))
			user.SetBrick(user.GetBrick1())
		} else if oper == 1 {
			user.SetBrick1(user.GetBrick1() - uint64(amount))
			user.SetBrick(user.GetBrick1())
		}
	} else if os == "android" {
		if oper == 0 {
			user.SetBrick2(user.GetBrick2() + uint64(amount))
			user.SetBrick(user.GetBrick2())
		} else if oper == 1 {
			user.SetBrick2(user.GetBrick2() - uint64(amount))
			user.SetBrick(user.GetBrick2())
		}
	}

	user.tlogMoneyFlow(amount, reason, oper, common.MT_Brick, user.GetBrick()) //tlog货币流水表
}

// consumeBalance 消耗金砖余额
func (user *LobbyUser) consumeBalance(param []byte, amount uint32, reason uint32) (uint32, string) {
	var ret uint32
	var billno string

	for i := 0; i < 5; i++ {
		start := time.Now()
		info, os, err := pay.ConsumeBalance(param, GetSrvInst().msdkAddr, amount, user.GetDBID())
		if err != nil || info == nil {
			ret = 1
			user.Error("ConsumeBalance err: ", err)
			continue
		}

		user.Debug("ConsumeBalance time: ", time.Since(start).Nanoseconds()/1000000, " ms")
		user.Debugf("info： %+v\n", info)

		ret = uint32(info.Ret)
		billno = info.Billno

		if ret == 0 {
			user.updateBalance(amount, 1, reason, os)
			break
		}
	}

	return ret, billno
}

// cancelPay 取消支付
func (user *LobbyUser) cancelPay(param []byte, billno string, amount uint32, reason uint32) uint32 {
	var ret uint32

	for i := 0; i < 5; i++ {
		start := time.Now()
		info, os, err := pay.CancelPay(param, GetSrvInst().msdkAddr, amount, billno)
		if err != nil || info == nil {
			ret = 1
			user.Error("CancelPay err: ", err)
			continue
		}

		user.Debug("CancelPay time: ", time.Since(start).Nanoseconds()/1000000, " ms")
		user.Debugf("info： %+v\n", info)

		ret = uint32(info.Ret)
		if ret == 0 {
			user.updateBalance(amount, 0, reason, os)
			break
		}
	}

	return ret
}

// presentBricks 赠送金砖
func (user *LobbyUser) presentBricks(param []byte, amount uint32, reason uint32) (uint32, string) {
	var ret uint32
	var msg string

	for i := 0; i < 5; i++ {
		start := time.Now()
		info, os, err := pay.Present(param, GetSrvInst().msdkAddr, amount, user.GetDBID())
		if err != nil || info == nil {
			ret = 1
			user.Error("Present err: ", err)
			continue
		}

		user.Debug("Present time: ", time.Since(start).Nanoseconds()/1000000, " ms")
		user.Debugf("info： %+v\n", info)

		ret = uint32(info.Ret)
		msg = info.Msg

		if ret == 0 {
			user.updateBalance(amount, 0, reason, os)
			break
		}
	}

	return ret, msg
}

/*-------------------------首充系统-----------------------------*/
// isFirstPay 是否首充
func (user *LobbyUser) isFirstPay(amount uint32) bool {
	util := db.PlayerPayUtil(user.GetDBID(), common.MarketingNameFirstPay)
	if util.GetFirstPayRecord() != 0 {
		return false
	}

	var condition uint32
	item, ok := excel.GetPaysystem(common.MarketingFirstPay)
	if ok {
		condition = uint32(item.Condition)
	}

	if (amount / BrickRMBRatio) >= condition {
		return true
	}

	return false
}

// firstPay 玩家首次充值
func (user *LobbyUser) firstPay() {
	util := db.PlayerPayUtil(user.GetDBID(), common.MarketingNameFirstPay)
	util.SetFirstPayRecord(1)
	user.syncFirstPay()
}

// syncFirstPay 同步首充
func (user *LobbyUser) syncFirstPay() {
	util := db.PlayerPayUtil(user.GetDBID(), common.MarketingNameFirstPay)
	ret := util.GetFirstPayRecord()
	user.RPC(iserver.ServerTypeClient, "FirstPayNotify", ret)
}

// getFirstPayAwards 获取首充奖励
func (user *LobbyUser) getFirstPayAwards() map[uint32]uint32 {
	awards := make(map[uint32]uint32)
	util := db.PlayerPayUtil(user.GetDBID(), common.MarketingNameFirstPay)

	if util.GetFirstPayRecord() == 1 {
		awards = common.GetMarketingAwards(common.MarketingFirstPay)
	}

	return awards
}

// setFirstPayAwardsDrawed 记录玩家领取首充奖励
func (user *LobbyUser) setFirstPayAwardsDrawed() {
	util := db.PlayerPayUtil(user.GetDBID(), common.MarketingNameFirstPay)
	util.SetFirstPayRecord(2)
	user.syncFirstPay()
}

/*-------------------------月卡系统-----------------------------*/
// syncMonthCard 同步月卡
func (user *LobbyUser) syncMonthCard() {
	name := pay.GetMonthCardDBFlag(user.GetPayOS())
	util := db.PlayerPayUtil(user.GetDBID(), name)

	own := uint8(1)
	if util.IsOwnValidMonthCard() {
		own = 2
	}

	draw := uint8(1)
	if util.GetOnceAwardUndrawedNum() == 0 && util.IsDayAwardsDrawed() {
		draw = 2
	}

	left := util.GetMonthCardEndTime() - time.Now().Unix() - 60
	if left < 0 {
		left = 0
	}

	user.RPC(iserver.ServerTypeClient, "MonthCardInfoNotify", own, draw, left)
	user.Info("MonthCardInfoNotify, own: ", own, " draw: ", draw, " left: ", left)
}

// monthCardExpireNotify 月卡过期邮件通知
func (user *LobbyUser) monthCardExpireNotify() {
	os := user.GetPayOS()
	name := pay.GetMonthCardDBFlag(os)
	util := db.PlayerPayUtil(user.GetDBID(), name)

	if util.NeedNotifyExpire(1) { //通知玩家月卡即将过期
		mail, ok := excel.GetMail(common.Mail_MonthCardSoonExpire)
		if ok {
			title := mail.MailTitle
			content := fmt.Sprintf(mail.Mail, util.GetMonthCardLeftDays())
			sendObjMail(user.GetDBID(), os, 0, title, content, "", "", nil)
			user.MailNotify()
			util.SetExpireNotifyed(1)
		}
	} else if util.NeedNotifyExpire(2) { //通知玩家月卡已经过期
		mail, ok := excel.GetMail(common.Mail_MonthCardExpired)
		if ok {
			title := mail.MailTitle
			content := mail.Mail
			sendObjMail(user.GetDBID(), os, 0, title, content, "", "", nil)
			user.MailNotify()
			util.SetExpireNotifyed(2)
		}
	}
}

// isBuyNewMonthCard 判断玩家是否购买了新的月卡
func (user *LobbyUser) isBuyNewMonthCard(list []pay.MidasQueryBalanceTssList) bool {
	if len(list) == 0 {
		return false
	}

	name := pay.GetMonthCardDBFlag(user.GetPayOS())
	util := db.PlayerPayUtil(user.GetDBID(), name)
	expire := util.GetMonthCardEndTime()

	for _, item := range list {
		end, err := time.ParseInLocation("2006-01-02 15:04:05", item.Endtime, time.Local)
		if err != nil {
			user.Error("ParseInLocation err: ", err)
			continue
		}

		endStamp := common.GetDayBeginStamp(end.Unix())
		if endStamp > expire {
			return true
		}
	}

	return false
}

// buyMonthCard 购买月卡
func (user *LobbyUser) buyMonthCard(param []byte, list []pay.MidasQueryBalanceTssList) {
	os := user.GetPayOS()
	name := pay.GetMonthCardDBFlag(os)
	util := db.PlayerPayUtil(user.GetDBID(), name)
	expire := util.GetMonthCardEndTime()

	for _, item := range list {
		end, err := time.ParseInLocation("2006-01-02 15:04:05", item.Endtime, time.Local)
		if err != nil {
			user.Error("ParseInLocation err: ", err)
			continue
		}

		endStamp := common.GetDayBeginStamp(end.Unix())
		if endStamp <= expire {
			continue
		}

		begin, err := time.ParseInLocation("2006-01-02 15:04:05", item.Begintime, time.Local)
		if err != nil {
			user.Error("ParseInLocation err: ", err)
			return
		}

		beginStamp := common.GetDayBeginStamp(begin.Unix())
		util.SetMonthCardRecord(beginStamp, endStamp)
		user.MonthCardFlow(beginStamp, endStamp)

		period := common.GetMonthCardPeriod() * (24 * 60 * 60)
		cardNum := uint32(endStamp-beginStamp+60) / period
		util.SetMonthCardNum(cardNum)

		// 购买月卡后达成首充
		amount := common.GetMonthCardPrice() * BrickRMBRatio
		if user.isFirstPay(amount) {
			user.firstPay()
		}

		mail, ok := excel.GetMail(common.Mail_BuyMonthCard)
		if ok {
			title := mail.MailTitle
			content := fmt.Sprintf(mail.Mail, util.GetMonthCardLeftDays())
			sendObjMail(user.GetDBID(), os, 0, title, content, "", "", nil)
			user.MailNotify()
		}
	}

	user.syncMonthCard()
}

// getMonthCardAwards 获取玩家可以领取的月卡奖励
func (user *LobbyUser) getMonthCardAwards() map[uint32]uint32 {
	awards := make(map[uint32]uint32)
	name := pay.GetMonthCardDBFlag(user.GetPayOS())
	util := db.PlayerPayUtil(user.GetDBID(), name)
	num := util.GetOnceAwardUndrawedNum()

	for i := uint32(0); i < num; i++ {
		tm := common.GetMarketingAwards(common.MarketingMonthCardOnce)
		awards = common.CombineMapUint32(awards, tm)
	}

	if !util.IsDayAwardsDrawed() {
		tm := common.GetMarketingAwards(common.MarketingMonthCardDay)
		awards = common.CombineMapUint32(awards, tm)
	}

	return awards
}

// setMonthCardAwardsDrawed 记录玩家领取月卡奖励
func (user *LobbyUser) setMonthCardAwardsDrawed() {
	name := pay.GetMonthCardDBFlag(user.GetPayOS())
	util := db.PlayerPayUtil(user.GetDBID(), name)
	num := util.GetOnceAwardUndrawedNum()

	for i := uint32(0); i < num; i++ {
		util.IncrOnceAwardDrawNum()
	}

	if !util.IsDayAwardsDrawed() {
		util.SetDayAwardsDrawed()
	}

	user.syncMonthCard()
}

// monthCardRegularCheck 月卡定时检测
func (user *LobbyUser) monthCardRegularCheck() {
	name := pay.GetMonthCardDBFlag(user.GetPayOS())
	util := db.PlayerPayUtil(user.GetDBID(), name)
	left := util.GetMonthCardEndTime() - time.Now().Unix()

	if left >= 0 && left < 60 {
		user.syncMonthCard()
	}

	user.monthCardExpireNotify()
}

/*-------------------------首次充送-----------------------------*/
// getFirstPayLevel 获取玩家充值达到的新档次
func (user *LobbyUser) getFirstPayLevel(amount uint32) uint32 {
	max := common.GetMaxPayLevel(amount / BrickRMBRatio)
	if max == 0 {
		return 0
	}

	util := db.PlayerPayUtil(user.GetDBID(), common.MarketingNamePayLevel)
	if !util.IsPayLevelMatched(max) {
		return max
	}

	return 0
}

// firstPayLevel 发放首次充送奖励
func (user *LobbyUser) firstPayLevel(param []byte, level uint32) {
	awards := common.GetMarketingAwards(uint64(level))
	bricks := awards[common.Item_Brick]
	if bricks == 0 {
		return
	}

	ret, _ := user.presentBricks(param, bricks, common.RS_PayLevel)
	if ret == 0 {
		value := common.GetPaySystemValue(common.PaySystem_LevelReset)
		round := common.StringToUint32(value)

		util := db.PlayerPayUtil(user.GetDBID(), common.MarketingNamePayLevel)
		myround := util.GetPayLevelResetRound()

		if round != 0 && myround == 0 {
			util.SetPayLevelResetRound(round)
		}

		util.AddPayLevelRecord(level)
		user.syncPayLevel()
	}
}

// payLevelRegularCheck 定时检测是否需要重置首次充送的记录
func (user *LobbyUser) payLevelRegularCheck() {
	value := common.GetPaySystemValue(common.PaySystem_LevelReset)
	round := common.StringToUint32(value)

	util := db.PlayerPayUtil(user.GetDBID(), common.MarketingNamePayLevel)
	myround := util.GetPayLevelResetRound()

	if myround != round {
		util.SetPayLevelResetRound(round)
		user.syncPayLevel()
	}
}

// syncPayLevel 同步首次充送
func (user *LobbyUser) syncPayLevel() {
	util := db.PlayerPayUtil(user.GetDBID(), common.MarketingNamePayLevel)
	ret := &protoMsg.Uint32Array{
		List: util.GetPayLevelRecord(),
	}
	user.RPC(iserver.ServerTypeClient, "PayLevelNotify", ret)
	user.Debugf("PayLevelNotify: %+v\n", ret)
}

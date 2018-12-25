package main

import (
	"common"
	"db"
	"excel"
	"fmt"
	"protoMsg"
	"time"
	"zeus/iserver"
)

// needDisplayTakeTeacher 客户端是否需要显示绑定战友界面
func (user *LobbyUser) needDisplayTakeTeacher() bool {
	util := db.PlayerInfoUtil(user.GetDBID())

	record, err := util.GetOldBringNewRecord()
	if err != nil || record == nil {
		user.Error("GetOldBringNewRecord err: ", err)
		return false
	}

	if record.TakeTeacherDeadTime <= time.Now().Unix() {
		return false
	}

	return true
}

// needDisplayReceivePupil 客户端是否需要显示新兵招募界面
func (user *LobbyUser) needDisplayReceivePupil() bool {
	minLevel := uint32(common.GetTBSystemValue(common.System_RecruitInitLevel))
	if user.GetLevel() >= minLevel {
		return true
	}

	return false
}

// syncOldBringNew 同步以老带新标记信息
func (user *LobbyUser) syncOldBringNew() {
	var bind, bindHint, recruit, recruitHint bool
	util := db.PlayerInfoUtil(user.GetDBID())

	record, err := util.GetOldBringNewRecord()
	if err != nil {
		user.Error("GetOldBringNewRecord err: ", err)
	}

	if record != nil {
		if record.TakeTeacherDeadTime > time.Now().Unix() {
			bind = true

			if len(record.Teachers) == 0 {
				bindHint = true
			}

			for _, info := range record.TakeTeacherAwards {
				if info.State == 1 {
					bindHint = true
					break
				}
			}
		}
	}

	if user.GetLevel() >= uint32(common.GetTBSystemValue(common.System_RecruitInitLevel)) {
		recruit = true

		if record != nil {
			for _, info := range record.ReceivePupilAwards {
				if info.State == 1 {
					recruitHint = true
					break
				}
			}
		}
	}

	user.RPC(iserver.ServerTypeClient, "OldBringNewRet", bind, bindHint, recruit, recruitHint)
	user.Info("OldBringNewRet, bind: ", bind, " bindHint: ", bindHint, " recruit: ", recruit, " recruitHint: ", recruitHint)
}

// syncOldBringNewDetail 同步以老带新详细信息
func (user *LobbyUser) syncOldBringNewDetail(typ uint8) {
	detail := &protoMsg.OldBringNewDetail{}
	awardM := excel.GetBindingAwardsMap()
	util := db.PlayerInfoUtil(user.GetDBID())

	record, err := util.GetOldBringNewRecord()
	if err != nil {
		user.Error("GetOldBringNewRecord err: ", err)
	}

	switch typ {
	case 1:
		detail.Typ = 1

		if record != nil {
			detail.Teachers = record.Teachers

			left := record.TakeTeacherDeadTime - time.Now().Unix()
			if left < 0 {
				left = 0
			}

			detail.TimeLeft = uint32(left)
		}

		for _, v := range awardM {
			if v.Type != 1 {
				continue
			}

			award := &protoMsg.AwardInfo{
				Id: uint32(v.Id),
			}

			if record != nil {
				for _, info := range record.TakeTeacherAwards {
					if info.Id == award.Id {
						award.State = uint32(info.State)
					}
				}
			}

			detail.Awards = append(detail.Awards, award)
		}

	case 2:
		detail.Typ = 2

		if record != nil {
			detail.Pupils = record.Pupils
		}

		for _, v := range awardM {
			if v.Type != 2 {
				continue
			}

			award := &protoMsg.AwardInfo{
				Id: uint32(v.Id),
			}

			if record != nil {
				for _, info := range record.ReceivePupilAwards {
					if info.Id == award.Id {
						award.State = uint32(info.State)
					}
				}
			}

			detail.Awards = append(detail.Awards, award)
		}
	}

	user.RPC(iserver.ServerTypeClient, "OldBringNewDetailRet", detail)
	user.Infof("OldBringNewDetailRet: %+v\n", detail)
}

// canTakeTeacher 是否可拜师
func (user *LobbyUser) canTakeTeacher() bool {
	util := db.PlayerInfoUtil(user.GetDBID())

	record, err := util.GetOldBringNewRecord()
	if err != nil || record == nil {
		user.Error("GetOldBringNewRecord err: ", err)
		return false
	}

	if record.TakeTeacherDeadTime <= time.Now().Unix() {
		return false
	}

	if len(record.Teachers) >= 1 {
		return false
	}

	return true
}

// takeTeacher 拜师
func (user *LobbyUser) takeTeacher(uid uint64) {
	ret := uint8(1)

	if user.canTakeTeacher() && user.friendMgr.isBindableFriend(uid) {

		util := db.PlayerInfoUtil(user.GetDBID())
		util.AddTeacher(uid)

		//更新拜师奖励状态
		teachers := util.GetTeachers()
		ids := common.GetAvailableAwards(1, teachers)
		util.SetOldBringNewDrawable(1, ids)
		user.syncOldBringNewDetail(1)

		if teachers == 1 {
			user.comradeTaskInfoNotify()
		}

		util = db.PlayerInfoUtil(uid)
		util.AddPupil(user.GetDBID())

		//更新收徒奖励状态
		pupils := util.GetPupils()
		ids = common.GetAvailableAwards(2, pupils)
		util.SetOldBringNewDrawable(2, ids)

		mail, ok := excel.GetMail(common.Mail_TakeTeacher)
		if ok {
			title := mail.MailTitle
			content := fmt.Sprintf(mail.Mail, user.GetName())
			sendObjMail(uid, "", 0, title, content, "", "", nil)
		}

		//通知好友
		user.friendMgr.SendProxyInfo(uid, "ReceivePupil")

		ret = 0
	}

	user.RPC(iserver.ServerTypeClient, "TakeTeacherRet", ret, uid)
	user.Info("TakeTeacherRet, ret: ", ret, " uid: ", uid)
}

// receivePupil 收徒
func (user *LobbyUser) receivePupil() {
	user.MailNotify()
	user.syncOldBringNewDetail(2)

	pupils := db.PlayerInfoUtil(user.GetDBID()).GetPupils()
	if pupils == 1 {
		user.comradeTaskInfoNotify()
	}
}

// releaseTeacherPupil 解除师徒关系
func (user *LobbyUser) releaseTeacherPupil(uid uint64) {
	util1 := db.PlayerInfoUtil(user.GetDBID())
	util2 := db.PlayerInfoUtil(uid)

	if util1.IsTeacher(uid) {

		util1.RemoveTeacher(uid)
		util2.RemovePupil(user.GetDBID())

	} else if util1.IsPupil(uid) {

		util1.RemovePupil(uid)
		util2.RemoveTeacher(user.GetDBID())

	}
}

// drawOldBringNewAwards 领取以老带新相关奖励
func (user *LobbyUser) drawOldBringNewAwards(typ uint8, id uint32) {
	ret := uint8(1)
	util := db.PlayerInfoUtil(user.GetDBID())

	if util.CanDrawOldBringNew(typ, id) {
		awards := common.GetOldBringNewAwards(id)
		user.storeMgr.GetAwards(awards, common.RS_TeacherPupil, false, false)

		util.SetOldBringNewDrawed(typ, id)
		ret = 0
	}

	user.RPC(iserver.ServerTypeClient, "DrawOldBringNewAwardsRet", ret)
	user.Info("DrawOldBringNewAwardsRet, ret: ", ret)
}

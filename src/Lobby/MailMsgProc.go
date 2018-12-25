package main

import (
	"common"
	"db"
	"excel"
	"protoMsg"
	"time"
	"zeus/iserver"

	log "github.com/cihub/seelog"
)

func (p *LobbyUserMsgProc) RPC_ReqGetMailList() {
	p.user.checkGlobalMail()
	checkMail(p.user.GetDBID())

	mails := db.MailUtil(p.user.GetDBID()).GetMails()
	os := p.user.GetPayOS()
	var msg protoMsg.RetMailList

	for _, v := range mails {
		if v.Os != "" && v.Os != os {
			continue
		}
		msg.Mails = append(msg.Mails, v)
	}

	if len(msg.Mails) > 0 {
		p.user.RPC(iserver.ServerTypeClient, "RetMailList", &msg)
		p.user.Info("Request to get mail list, len: ", len(msg.Mails))
	}
}

func (p *LobbyUserMsgProc) RPC_ReqMailInfo(mailid uint64) {
	p.user.Info("Request to get mail info, mailid: ", mailid)
	mail := db.MailUtil(p.user.GetDBID()).GetMail(mailid)

	if mail != nil {
		mail.Haveread = true
		db.MailUtil(p.user.GetDBID()).SaveMail(mail)

		p.user.RPC(iserver.ServerTypeClient, "RetMailInfo", mail)
	}
}

func (p *LobbyUserMsgProc) RPC_DelMail(proto *protoMsg.DelMail) {
	p.user.Info("Delete mail, mailid: ", proto.Mailid)

	for _, v := range proto.Mailid {
		mail := db.MailUtil(p.user.GetDBID()).GetMail(v)
		if mail == nil {
			continue
		}
		if !mail.Haveget && len(mail.Objs) != 0 {
			continue
		}

		db.MailUtil(p.user.GetDBID()).RemMail(v)
	}
}

func (p *LobbyUserMsgProc) RPC_GetMailObj(mailid uint64, param []byte) {
	p.user.Info("Get mail obj, mailid: ", mailid)
	mail := db.MailUtil(p.user.GetDBID()).GetMail(mailid)
	ret := p.user.drawMailObjs(mail, param)
	p.user.RPC(iserver.ServerTypeClient, "GetMailObj", mailid, ret)
}

// RPC_GetAllMailObjsReq 一键领取全部邮件附件
func (p *LobbyUserMsgProc) RPC_GetAllMailObjsReq(param []byte) {
	p.user.checkGlobalMail()
	checkMail(p.user.GetDBID())

	msg := &protoMsg.MailInfoList{}
	util := db.MailUtil(p.user.GetDBID())
	mails := util.GetMails()
	os := p.user.GetPayOS()

	for _, v := range mails {
		if v.Os != "" && v.Os != os {
			continue
		}

		if !p.user.drawMailObjs(v, param) {
			continue
		}

		msg.List = append(msg.List, util.GetMail(v.Mailid))
	}

	p.user.RPC(iserver.ServerTypeClient, "GetAllMailObjsRet", msg)
	p.user.Info("GetAllMailObjsRet, len(msg.List) = ", len(msg.List))
}

func checkMail(dbid uint64) {
	overdue := int64(common.GetTBSystemValue(common.System_MailOverdue))
	max := common.GetTBSystemValue(common.System_MailMax)
	now := time.Now().Unix()
	del := make([]uint64, 0)

	mails := db.MailUtil(dbid).GetMails()
	leftmail := make(map[uint64]*protoMsg.MailInfo)
	for _, v := range mails {
		if now >= int64(v.Gettime)+overdue {
			del = append(del, v.Mailid)
			continue
		}
		leftmail[v.Mailid] = v
	}

	if uint(len(leftmail)) > max {
		num := uint(len(leftmail)) - max
		var i uint
		for k, v := range leftmail {
			if v.Haveget || len(v.Objs) == 0 {
				delete(leftmail, k)
				del = append(del, k)
				i++
			}

			if i >= num {
				break
			}
		}
	}

	if uint(len(leftmail)) > max {
		num := uint(len(leftmail)) - max
		var i uint
		for k, _ := range leftmail {
			delete(leftmail, k)
			del = append(del, k)
			i++

			if i >= num {
				break
			}
		}
	}

	log.Infof("Player %d delete %d mails\n", dbid, len(del))
	for _, v := range del {
		db.MailUtil(dbid).RemMail(v)
	}
}

func (user *LobbyUser) MailNotify() {
	var havenew bool
	mails := db.MailUtil(user.GetDBID()).GetMails()
	for _, v := range mails {
		if !v.Haveread {
			havenew = true
		}
	}

	if havenew {
		user.RPC(iserver.ServerTypeClient, "AddNewMail")
	}
}

func (user *LobbyUser) checkGlobalMail() {
	overdue := int64(common.GetTBSystemValue(common.System_MailOverdue))
	now := time.Now().Unix()
	havesend := db.PlayerGlobalMailUtil(user.GetDBID()).GetAll()
	mails := db.GetGlobalMails()
	for _, v := range mails {
		mailid := common.Uint64ToString(v.Mailid)
		if _, ok := havesend[mailid]; !ok && now <= int64(v.Gettime)+overdue {
			db.PlayerGlobalMailUtil(user.GetDBID()).AddMail(v.Mailid)

			objs := make(map[uint32]uint32, 0)
			for _, obj := range v.Objs {
				objs[obj.Id] = obj.Num
			}
			sendObjMail(user.GetDBID(), v.Os, v.Mailtype, v.Title, v.Text, v.Url, v.Button, objs)

			user.tlogSnsFlow(1, SNSTYPE_RECEIVEEMAIL, "0") //tlog社交流水表(SNSTYPE_RECEIVEEMAIL)
		}
	}
}

// drawMailObjs 领取邮件奖励
func (user *LobbyUser) drawMailObjs(mail *protoMsg.MailInfo, param []byte) bool {
	if mail == nil || mail.Haveget || len(mail.Objs) == 0 {
		return false
	}

	bricks := getMailBricks(mail.Objs)
	if bricks != 0 {
		ret, _ := user.presentBricks(param, bricks, common.RS_Mail)
		if ret != 0 {
			return false
		}
	}

	for _, v := range mail.Objs {
		if v.Id != common.Item_Brick {
			user.storeMgr.GetGoods(v.Id, v.Num, common.RS_Mail, common.MT_NO, 0) //5 邮件获取
		}
	}

	mail.Haveread = true
	mail.Haveget = true
	db.MailUtil(user.GetDBID()).SaveMail(mail)

	return true
}

func sendObjMail(dbid uint64, os string, mailtype uint32, title string, text string, url string, button string, objs map[uint32]uint32) {
	mail := &protoMsg.MailInfo{}

	mailid := common.CreateNewMailID()
	if mailid == 0 {
		return
	}

	mail.Mailid = mailid
	mail.Mailtype = mailtype
	mail.Gettime = uint64(time.Now().Unix())
	mail.Haveread = false
	mail.Title = title
	mail.Text = text
	mail.Url = url
	mail.Button = button
	mail.Haveget = false
	mail.Os = os

	for k, v := range objs {
		_, ok := excel.GetStore(uint64(k))
		if ok {
			obj := &protoMsg.MailObject{Id: k, Num: v}
			mail.Objs = append(mail.Objs, obj)
		}
	}

	db.MailUtil(dbid).AddMail(mail)
	log.Info("Send mail to player, title: ", title, " mailid: ", mailid, "	uid: ", dbid)
}

// getMailBricks 获取邮件中发放的金砖数量
func getMailBricks(objs []*protoMsg.MailObject) uint32 {
	var num uint32
	for _, v := range objs {
		if v.Id == common.Item_Brick {
			num += v.Num
		}
	}
	return num
}

package main

import (
	"common"
	"db"
	"excel"
	"idip"
	"protoMsg"
	"time"

	log "github.com/cihub/seelog"
)

func sendMail(dbid uint64, os string, mailtype uint32, title string, text string, url string, button string) {
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

	db.MailUtil(dbid).AddMail(mail)
	log.Info("发送邮件", mailid, title)
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
	log.Debug("发送邮件", dbid, mailid, title)
}

func sendGlobalObjMail(os string, mailtype uint32, title string, text string, url string, button string, objs map[uint32]uint32) {
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

	db.AddGlobalMail(mail)
	log.Debug("发送全服邮件", mailid, title)
}

func sendIdipMail(msg *idip.DoSendItemReq) {
	mailid := common.CreateNewMailID()
	if mailid == 0 {
		return
	}

	data := &db.IdipMailData{}
	data.MailID = uint32(mailid)
	data.SendTime = uint32(time.Now().Unix()) //msg.SendTime
	data.MailTitle = msg.MailTitle
	data.MailContent = msg.MailContent
	data.MinLevel = uint16(msg.LevelFloor)
	data.MaxLevel = msg.LevelTop
	data.Hyperlink = msg.Line
	data.ButtonCon = msg.ButtonCon
	for i := 0; i < len(msg.ItemData); i++ {
		tmp := msg.ItemData[i]
		if i == 0 {
			data.ItemOneID = tmp.ItemID
			data.ItemOneNum = uint32(tmp.ItemNum)
		}
		if i == 1 {
			data.ItemTwoID = tmp.ItemID
			data.ItemTwoNum = uint32(tmp.ItemNum)
		}
		if i == 2 {
			data.ItemThreeID = tmp.ItemID
			data.ItemThreeNum = uint32(tmp.ItemNum)
		}
		if i == 3 {
			data.ItemFourID = tmp.ItemID
			data.ItemFourNum = uint32(tmp.ItemNum)
		}
		if i == 4 {
			data.ItemFiveID = tmp.ItemID
			data.ItemFiveNum = uint32(tmp.ItemNum)
		}
	}

	db.IdipMailUtil().AddMail(data)
	log.Info("发送idip邮件", mailid)
}

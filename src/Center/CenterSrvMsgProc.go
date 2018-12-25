package main

import (
	"common"
	"encoding/json"
	"idip"
	"zeus/dbservice"
	"zeus/entity"
	"zeus/iserver"

	log "github.com/cihub/seelog"
)

// CenterSrvMsgProc 服务器消息处理类
type CenterSrvMsgProc struct {
	srv *Server
}

func (p *CenterSrvMsgProc) RPC_SendOneMail(data []byte) {
	log.Info("send one mail")

	var msg *idip.DoSendItemMailReq
	err := json.Unmarshal(data, &msg)
	if err != nil {
		log.Error("json解析失败")
		return
	}

	objs := make(map[uint32]uint32)
	objs[uint32(msg.ItemID)] = uint32(msg.ItemNum)

	uid, err := dbservice.GetUID(msg.OpenID)
	if err != nil {
		log.Error("未找到账号", msg.OpenID)
		return
	}
	sendObjMail(uid, "", 0, msg.MailTitle, msg.MailContent, "", "", objs)

	entityID, err := dbservice.SessionUtil(uid).GetUserEntityID()
	if err != nil {
		return
	}

	srvID, spaceID, err := dbservice.EntitySrvUtil(entityID).GetSrvInfo(common.ServerTypeLobby)

	proxy := entity.NewEntityProxy(srvID, spaceID, entityID)
	proxy.RPC(iserver.ServerTypeClient, "AddNewMail")
}

func (p *CenterSrvMsgProc) RPC_SendAllMail(data []byte) {
	log.Info("send all mail ")

	var msg *idip.DoSendItemReq
	err := json.Unmarshal(data, &msg)
	if err != nil {
		log.Error("json解析失败")
		return
	}

	objs := make(map[uint32]uint32)
	for _, v := range msg.ItemData {
		objs[v.ItemID] = uint32(v.ItemNum)
	}

	sendIdipMail(msg)

	/*
		go func(msg *idip.DoSendItemReq, users []uint64) {

			for _, v := range users {
				sendObjMail(v, "", 0, msg.MailTitle, msg.MailContent, msg.Line, msg.ButtonCon, objs)
			}

		}(msg, common.GetUsers())
	*/
	sendGlobalObjMail("", 0, msg.MailTitle, msg.MailContent, msg.Line, msg.ButtonCon, objs)

	// 通知所有在线玩家
	GetSrvInst().FireEvent(iserver.RPCChannel, "AddNewMail")
}

// RPC_AddAnnuouce 添加公告
func (p *CenterSrvMsgProc) RPC_AddAnnuonce(id uint64) {
	log.Info("add Annuonce")

	p.srv.annuonceMgr.AddAnnuouce(id)
}

// RPC_DelAnnuoucing 删除进行中的公告
func (p *CenterSrvMsgProc) RPC_DelAnnuoncing(id uint64) {
	log.Info("add Annuoncing")

	p.srv.annuonceMgr.DelAnnuoucing(id)
}

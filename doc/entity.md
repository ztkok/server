## Player实体

1. 登录成功后连接gateway验证后创建 Player并同步到其它成功注册Player实体的服务器
```go
func (mgr *GateUserMgr) addUser(uid uint64) *GateUser {

	user, err := GetSrvInst().CreateEntityAll("Player", uid, "", true)
	if err != nil {
		log.Error("Add user failed", err, "UID:", uid)
		return nil
	}

	return user.(*GateUser)
}
```

    CreateEntityAll内部分同步到其它服务器
    目前只在Lobby上同步创建了Player实体

2. Room上Player实体的创建（因为实现了iserver.ISpaceEntity接口的实体不能注册到全局，所以Room上的Player不会随着Gateway的创建而创建）

```go
// 客户端给lobby发送进入场景的消息
func (proc *LobbyUserMsgProc) RPC_EnterScene(spaceid uint64) {
    //通知场景实体这个玩家进入了
	proc.user.EnterSpace(spaceid)

	log.Info("角色进入地图 ", proc.user)
}
```

```go
// LobbyUser实体发送EnterSpaceReq消息到RoomServer的场景实体中
func (e *Entity) EnterSpace(spaceID uint64) {
	if e.IsSpace() {
		log.Error("space entity couldn't move into space", e)
		return
	}

	util := dbservice.EntitySrvUtil(e.GetID())

	srvID, err := dbservice.SpaceUtil(spaceID).GetSrvID()
	if err != nil {
		log.Error("get srvID error", err, e)
		return
	}

	_, oldSpaceID, err := util.GetSpaceInfo()
	if err != nil {
		log.Error("get spaceID error", err, e)
		return
	}

	if oldSpaceID != 0 {
		if oldSpaceID == spaceID {
			log.Error("target spaceID is same as the space which entity in ... ", e)
			return
		}
	}

	msg := &msgdef.EnterSpaceReq{
		SrvID:      srvID,
		SpaceID:    spaceID,
		EntityType: e.entityType,
		EntityID:   e.entityID,
		DBID:       e.dbid,
		InitParam:  serializer.Serialize(e.initParam),
		OldSrvID:   0,
		OldSpaceID: 0,
	}

	if err := iserver.GetSrvInst().PostMsgToSpace(srvID, spaceID, msg); err != nil {
		log.Error(err, e)
	}
}
```

```go
// 场景实体收到消息后创建RoomUser实体
func (proc *SpaceMsgProc) MsgProc_EnterSpaceReq(content msgdef.IMsg) {
	msg := content.(*msgdef.EnterSpaceReq)

	params := serializer.UnSerialize(msg.InitParam)
	if len(params) < 1 {
		log.Error("Unmarshal initparam error ", msg.InitParam)
		return
	}

	err := proc.space.AddEntity(msg.EntityType, msg.EntityID, msg.DBID, params[0], false, false)
	if err != nil {
		log.Error("space add entity error", err, proc.space)
		return
	}
}
```


```go
// RoomUser实体创建的初始化函数里会调用onEnterSpace，地址等场景实体信息发给客户端
func (e *Entity) onEnterSpace() {
	ic, ok := e.GetRealPtr().(IEnterSpace)
	if ok {
		ic.OnEnterSpace()
	}

	if e.IsWatcher() {
		msg := &msgdef.EnterSpace{
			SpaceID:   e.GetSpace().GetID(),
			MapName:   e.GetSpace().GetInitParam().(string),
			EntityID:  e.GetID(),
			Addr:      iserver.GetSrvInst().GetCurSrvInfo().OuterAddress,
			TimeStamp: e.GetSpace().GetTimeStamp(),
		}
		if err := e.Post(iserver.ServerTypeClient, msg); err != nil {
			log.Error(err, e, msg)
		}

		e.aoies = append(e.aoies, AOIInfo{true, e})
	}
}
```

```go
// 客户端连接RoomServer进入场景实体
func (proc *SpaceSessesMsgProc) MsgProc_SessVertified(content interface{}) {

	uid := content.(uint64)

	sess := proc.srv.clientSrv.GetSession(uid)
	if sess == nil {
		seelog.Error("couldn't found sess ", uid)
		return
	}

	// Source 为1 代表是 SpaceSess
	sess.Send(&msgdef.ClientVertifySucceedRet{
		Source:   1,
		UID:      uid,
		SourceID: iserver.GetSrvInst().GetSrvID(),
		Type:     0,
	})

	seelog.Debug("space sess establish !! ", content)
}

func (proc *SpaceSessesMsgProc) MsgProc_SpaceUserConnect(content interface{}) {
	msg := content.(*msgdef.SpaceUserConnect)
	sess := proc.srv.clientSrv.GetSession(msg.UID)
	if sess == nil {
		seelog.Error("space user connected but not find sess ", msg.UID)
		return
	}

	space := iserver.GetSrvInst().GetEntity(msg.SpaceID)
	if space == nil {
		seelog.Error("couldn't find space ", msg.SpaceID)
		return
	}

	imh, ok := space.(msghandler.IMsgHandlers)
	if !ok {
		seelog.Error("this not go happen")
		return
	}

	imh.FireMsg("SpaceUserSess", sess)
}
```

```go
// 通知客户端连接场景成功
func (proc *SpaceMsgProc) MsgProc_SpaceUserSess(sess iserver.ISess) {
	ie := proc.space.GetEntityByDBID("Player", sess.GetID())
	if ie == nil {
		log.Error("there is no player in space ui = ", sess.GetID())
		sess.Close()
		return
	}

	ise, ok := ie.(iserver.ISpaceEntity)
	if !ok {
		log.Error("conert to ispaceentity error , strange!! ")
		return
	}

	ise.SetClient(sess)
	if err := ise.Post(iserver.ServerTypeClient, &msgdef.SpaceUserConnectSucceedRet{}); err != nil {
		log.Error(err)
	}
}
```

    最后，游戏中的消息都投递到RoomUserMsgProc类中的消息处理中


## Space实体

```go
// 匹配成功创建场景实体
func (ws *WaitingScene) Go() {
	str := common.MatchMgrSolo
	if !ws.singleMatch {
		if ws.teamType == TwoTeamType {
			str = common.MatchMgrDuo
		} else if ws.teamType == FourTeamType {
			str = common.MatchMgrSquad
		}
	}

	mapStr := fmt.Sprintf("%d:%s:%d", ws.mapid, str, GetSrvInst().GetSrvID())
	// log.Error("开启一局比赛!", mapStr)

	e, err := GetSrvInst().CreateEntityAll("Space", 0, mapStr, false)
	if err != nil {
		log.Error(err)
		return
	}

	ws.spaceID = e.GetID()
}
```


### zeus net

zeus net库包括3部分：
  - zeus/sess
  - zeus/server
  - zeus net 应用层（如，zeus/src/GatewayServer）


### zeus/sess

这块内容有一定的可制定化特点，除了网络基本功能外，额外提供以下功能：
  - IMsgDeliver / pushMsg / pushRawMsg

    通过实现具体IMsgDeliver接口，来达到控制消息如何投递处理。如，并发处理消息还是在主循环按帧率来处理消息。

  - IMsgHandlers / SetMsgHandler

    通过该接口，方便底层处理一些统一的消息处理操作，并可以让应用层接管消息处理


### zeus/sess 重点代码分析

  - ARQNetSrv::recvLoop
```go
func (srv *ARQNetSrv) recvLoop(conn net.Conn) {

	iconn := newConn(conn)
	//TODO: fixme 客户端只连接不发验证包的情况下, 会占用端口

	if srv.forwardMode {
		for {
			msg, msgID, rawMsg, err := readARQMsgForward(conn)
			if err != nil {
				srv.msgDeliver.connErr(iconn)
				log.Error("tcp read message error ", err, conn.RemoteAddr())
				break
			}

			if msg != nil {
				srv.msgDeliver.pushMsg(iconn, msg)
			}

			if rawMsg != nil {
				srv.msgDeliver.pushRawMsg(iconn, msgID, rawMsg)
			}
		}

	} else {
		for {
			msg, err := readARQMsg(conn)
			if err != nil {
				srv.msgDeliver.connErr(iconn)
				log.Error("tcp read message error ", err, conn.RemoteAddr())
				break
			}

			srv.msgDeliver.pushMsg(iconn, msg)
		}
	}
}
```
    1. 每个net session对象各自1个goroutine
    1. srv.msgDeliver.pushMsg(iconn, msg) / srv.msgDeliver.pushRawMsg(iconn, msgID, rawMsg)。通过 pushMsg、pushRawMsg的具体实现，来控制消息是否按主循环帧率来运行、或是并发运行


  - SessMgr::IMsgDeliver

    目前，reus/sess/SessMgr实现了IMsgDeliver接口，代码如下：

```go
func (mgr *SessMgr) pushMsg(conn IConn, msg msgdef.IMsg) {
	mgr.doMsg(conn, msg)
}

func (mgr *SessMgr) pushRawMsg(conn IConn, msgID int, msg []byte) {
	mgr.doRawMsg(conn, msgID, msg)
}

func (mgr *SessMgr) doMsg(conn IConn, msg msgdef.IMsg) {

	sess := mgr.fetchSess(conn)
	if sess.IsVertified() {
		if msg.Name() != "HeartBeat" {
			sess.FireMsg(msg.Name(), msg)
		}
	} else if msg.Name() == "ClientVertifyReq" {
		if err := mgr.vertifySess(msg, sess); err != nil {
			sess.Send(&msgdef.ClientVertifyFailedRet{})
			sess.Close()
			log.Info("client vertify failed", conn.String(), err)
		} else {
			sess.DoNormalMsg("SessVertified", msg.(*msgdef.ClientVertifyReq).UID)
		}
	} else {
		log.Warn("receive message , but sess is not vertified ", msg.Name())
	}

	sess.Touch()
}

func (mgr *SessMgr) doRawMsg(conn IConn, msgID int, msg []byte) {
	sess := mgr.fetchSess(conn)
	sess.DoNormalMsg("MsgForward", &RawMsg{conn, msgID, msg})

	sess.Touch()
}
```

   1. 统一的登录验证过程。不管客户端连接服务器、服务器连接服务器，服务器处理过程一样。

      ClientVertifyReq -> ClientVertifyFailedRet

      ClientVertifyReq -> SessVertified

      ( 具体的 MsgProc_SessVertified 实现，内部会发送ClientVertifyFailedRet或者ClientVertifySucceedRet  )

      因此应用层可以注册 MsgProc_SessVertified、MsgProc_ClientVertifyFailedRet、MsgProc_ClientVertifySucceedRet

      同时可以看出验证过程是并发的

   1. 转发处理，是并发的。

   1. sess.FireMsg(msg.Name(), msg)

      这是一句很微妙的语句，把消息放到session对象自己的消息队列中，消息处理行为取决于应用层。

        a. 未做SetMsgHandler 改变 消息处理器的，使用的默认的服务器消息处理器。那么它按照的主循环帧率来处理

        b. Entity对象都会重新注册自己的消息处理器。并通常的Entitys管理器都会让它的entity开goroutine并发处理消息


### zeus/server

  主要做下面一些事情：
  - 通用服务器间互连处理
  - 通用的登录验证流程中的事件触发
  - Entity一些通用消息处理，暂未细看



### zeus/server 重点代码分析

  - 服务器间互连处理

```go
func (srv *SrvNet) refresh() {
	srv.RefreshSrvInfo()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	srvNetTicker := time.NewTicker(time.Duration(rand.Intn(5)+15) * time.Second)
	defer srvNetTicker.Stop()

	util := dbservice.ServerUtil(srv.srvID)
	for {
		select {
		case <-ticker.C:
			if err := util.RefreshExpire(15); err != nil {
				log.Error(err)
			}
		case <-srvNetTicker.C:
			srv.RefreshSrvInfo()
		}
	}
}
func (srv *SrvNet) RefreshSrvInfo() {

	remoteSrvList, err := serverMgr.GetServerMgr().GetServerList()
	if err != nil {
		log.Error("fetch server info failed", err)
		return
	}
	// srv.srvInfos = &sync.Map{}
	for _, srvInfo := range remoteSrvList {
		srv.tryConnectToSrv(srvInfo)
		// srv.srvInfos.Store(srvInfo.ServerID, srvInfo)
	}
}
func (srv *SrvNet) tryConnectToSrv(info *iserver.ServerInfo) {

	if !srv.isNeedConnectedToSrv(info) {
		return
	}

	srv.pendingSesses.Store(info.ServerID, nil)
	go func() {

		s, err := sess.Dial("tcp", info.InnerAddress)
		if err != nil {
			srv.pendingSesses.Delete(info.ServerID)

			log.Errorf("Connect failed. %d to %d Addr %s, error:%v", srv.srvID, info.ServerID, info.InnerAddress, err)
			return
		}

		srv.pendingSesses.Store(info.ServerID, s)

		s.SetID(info.ServerID)
		s.SetMsgHandler(srv.msgSrv)

		s.Send(&msgdef.ClientVertifyReq{
			Source: srv.srvType,
			UID:    srv.srvID,
			Token:  srv.token,
		})

		log.Info("SrvNet try connect to ", info.ServerID)

	}()
}
```
  定时查看连接新服务器。其他互连相关略，如断开连接、连接成功等等


  - 登录验证流程中的事件触发

```go
func (srv *SrvNet) MsgProc_ClientVertifySucceedRet(content msgdef.IMsg) {
	msg := content.(*msgdef.ClientVertifySucceedRet)
	srv.onServerConnected(msg.SourceID)
}

func (srv *SrvNet) MsgProc_SessVertified(content interface{}) {

	uid := content.(uint64)

	srv.msgSrv.GetSession(uid).Send(&msgdef.ClientVertifySucceedRet{
		Source:   0,
		UID:      0,
		SourceID: srv.srvID,
		Type:     0,
	})

	log.Info("SrvNet server ", srv.srvID, "  recevice a connect from server ", uid)
}
```


### zeus net 应用层

主要变化是，可以在MsgProc_SessVertified事件中，接管session的消息处理器。从而达到处理逻辑消息，以及消息处理方式。

这里举一个例子：

```go
func (user *GateUser) Init(initParam interface{}) {
	user.RegMsgProc(&GateUserMsgProc{user: user})

	var err error
	// 获取并保存token
	if user.token, err = dbservice.SessionUtil(user.GetDBID()).GetToken(); err != nil {
		log.Error(err, user)
	}

	sess := GetSrvInst().clientSrv.GetSession(user.GetDBID())
	user.SetClient(sess)
	user.onLoginSuccessFinal()

	log.Info("GateUser Inited ", user)
}
func (e *Entity) SetClient(s iserver.ISess) {
	e.clientSess = s
	if s == nil {
		return
	}

	e.clientSess.SetMsgHandler(e.IMsgHandlers)
}
```

应用层网络会话对象，GateUser，是一个entity。通过entity::SetClient，让GateUser自己接管了消息处理方式。

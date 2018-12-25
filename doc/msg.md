## 消息注册

    基本上消息注册都是通过反射的方式，把一个对象上的MsgProc_和RPC_开头的函数注册到消息处理器MsgHandlers中
    RPCMsg消息体作了RPC的包装，用以方便的约定消息处理，对于此消息对投递到消息处理器的Rpc队列中


## 消息处理

```go
// DoMsg 将缓冲的消息一次性处理
func (handlers *MsgHandlers) DoMsg() {
	if !handlers.enable {
		return
	}

	var procName string
	defer func() {
		if err := recover(); err != nil {
			log.Error(err, procName)
			if viper.GetString("Config.Recover") == "0" {
				panic(fmt.Sprintln(err, procName))
			}
		}
	}()

	for {
		info, err := handlers.msgFireInfo.Pop()
		if err != nil {
			break
		}

		msg := info.(*msgFireInfo)
		e := handlers.DoNormalMsg(msg.name, msg.content)
		if e != nil {
			log.Error(e)
		}
	}

	for {
		info, err := handlers.rpcFireInfo.Pop()
		if err != nil {
			break
		}

		msg := info.(*rpcFireInfo)
		handlers.DoRPCMsg(msg.methodName, msg.data)
	}
}
```

`rpcFireInfo` 是RPC调用队列，用`FireRPC()`添加，
`msgFireInfo`是事件消息队列，用`FireMsg()`添加。
```
	FireMsg(name string, content interface{})
	FireRPC(methodName string, data []byte)
```

## 消息投递

    1. 连接上的消息包直接投递到消息处理器的消息列队中
```go
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
```
    2. 来自其它服转发过来的消息，由服务器的消息处理继续投递到对应实体的消息队列中
```go
func (es *Entities) MsgProc_EntityMsgTransport(content msgdef.IMsg) {
	msg := content.(*msgdef.EntityMsgTransport)
	ie := es.GetEntity(msg.EntityID)
	if ie == nil {
		//innerMsg, err := sess.DecodeMsg(msg.MsgContent[3], msg.MsgContent[4:])
		// if err != nil {
		// 	log.Error("Decode innerMsg failed", err, msg)
		// 	return
		// }

		//log.Error("EntityMsgTransport failed, entity not existed ", innerMsg.Name(), "   ", msg)
		return
	}

	te := ie.(iEntityCtrl)
	if msg.SrvType == iserver.ServerTypeClient {
		if sess := te.GetClientSess(); sess != nil {
			sess.SendRaw(msg.MsgContent)
		}
	} else {
		innerMsg, err := sess.DecodeMsg(msg.MsgContent[3], msg.MsgContent[4:])
		if err != nil {
			log.Error("Decode innerMsg failed", err, msg)
			return
		}

		if !iserver.GetSrvInst().IsSrvValid() {
			iserver.GetSrvInst().HandlerSrvInvalid(ie.GetID())
		} else {
			te.FireMsg(innerMsg.Name(), innerMsg)
		}
	}
}
```
    3. RPCMsg处理时把解出的消息投递到Rpc队列中
```go
// MsgProc_RPCMsg 实体间RPC调用消息处理
func (e *Entity) MsgProc_RPCMsg(content msgdef.IMsg) {
	msg := content.(*msgdef.RPCMsg)
	// 从客户端收到的消息要判断要调用的服务器类型

	if msg.ServerType == iserver.GetSrvInst().GetSrvType() {
		e.FireRPC(msg.MethodName, msg.Data)
	} else {
		e.Post(msg.ServerType, msg)
	}
}
```


## 消息体

	各个服务器在启动时基本都会调用common.InitMsg()，把protoMap.go中消息ID对应的消息类型的映射添加到msgdef中
	客户端通过gateway拉取proto.json的内容（和protoMap.go一样的内容）用以消息解析
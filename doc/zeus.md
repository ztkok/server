# zeus 包说明

svn://192.168.150.63/mb/zeus/branch_1/src/zeus

包名			| 说明
----------------|----------------------------------
admin			| HTTP 管理控制台
common			| ByteStream 字节流，方便二进制消息的序列和反序列            
dbservice		| Redis 服务         
entity			| 实体            
events			| 本地事件分发器
gameconfig		| 配置表工具，仅generator工具使用        
global			| 支持分布式结构的全局数据结构            
iserver			| 接口定义           
l5				| libqos_client封装                
linmath			| Vector2, Vector3           
login			| 登录服务             
msgdef			| 消息定义            
msghandler		| 消息处理器        
nav				| 寻路               
pool			| 缓冲区池              
safecontainer	| 安全容器     
serializer		| 序列化        
server			| 服务器            
serverMgr		| 服务器管理         
sess			| 会话              
space			| 空间实体            
timer			| 定时器            
tlog			| Tlog日志             
tsssdk			| tss_sdk           
unitypx			| unity physix 物理引擎           
zlog			| 初始化日志库              

## iserver

接口定义

### `iserver.IServer`

```go
// IServer 基础Server提供的接口
type IServer interface {
	ISrvNet		// 服务器的网状结构
	IEntities
	...
	Run()
	...
}
```

```go
// IEntities 用于实体的管理
type IEntities interface {
	msghandler.IMsgHandlers
	timer.ITimer
	events.IEvents  // 事件分发器接口
	...
}
```

## server

### `server.Server`

除了 Login 是 HTTP REST 服务器，其他服包含 `IServer` 接口，
其具体类型为 `server.Server`.

以 Gateway 为例：

```go
// GatewaySrv 网关服务器
type GatewaySrv struct {
	iserver.IServer
	...
}
```

```go
// GetSrvInst 获取服务器全局实例
func GetSrvInst() *GatewaySrv {
	...
	srvInst = &GatewaySrv{}
	srvInst.IServer = server.NewServer(...)
	...
	return srvInst
}
```

## serializer
```
// Serialize 序列化
func Serialize(args ...interface{}) []byte
// UnSerialize 反序列化
func UnSerialize(data []byte) []interface{} {
```
支持基本类型和proto消息类型。
这样一些简单的RPC接口只需直接输入基本数据类型，而不必定义proto消息。
缺点是没有proto来定义协议，协议定义分散在源码中，并且失去proto消息的版本兼容功能，
个人认为得不偿失。正确的做法是用proto来定义RPC接口。

## msghandler

### `msghandler.IMsgHandlers`

消息处理模块的接口, 以下类型包含该接口：

* `entity.Entities`
* `ebtity.Entity`
* `iserver.IEntities`
* `iserver.ISess`
* `sess.IMsgServer`
* `sess.NetSess`
* `sess.SessMgr`

### `msghandler.MsgHandlers`

消息处理中心, 是 IMsgHandlers 的实现。

#### `RegMsgProc()`
注册消息处理对象

```
// 其中 proc 是一个对象，包含是类似于 MsgProc_XXXXX的一系列函数，分别用来处理不同的消息
func (handlers *MsgHandlers) RegMsgProc(proc interface{})
```

proc 中还有 RPC_XXXXX 系列函数，用来处理 RPC 消息。

## msgdef

MessageConst.go 中定义了所有消息号，如：
```go
const (
	ClientVertifyReqMsgID = 1
	...
	EntityAOISMsgID = 66
)
```

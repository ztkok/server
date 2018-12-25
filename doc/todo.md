# TODO

## jinqing

* `user.RPC(iserver.ServerTypeClient, ...)` 可提取成 user.RPCClient(...)
不必总是输入 ServerTypeClient。

* serializer test fail
```
d:/go/bin/go.exe test -v [E:/svn/mb/zeus/branch_1/src/zeus/serializer]
# zeus/serializer
.\serializer_test.go:10:10: cannot assign 1 values to 2 variables
.\serializer_test.go:14:11: cannot assign 1 values to 2 variables
.\serializer_test.go:25:10: undefined: SerializeNew
.\serializer_test.go:28:9: undefined: UnSerializeNew
FAIL	zeus/serializer [build failed]
错误: 进程退出代码 2.
```

* server, sess 包放入 zeus/net 包，可能 msghandler 也应该放入

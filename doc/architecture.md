# 服务器架构

服务器有以下几个进程：

1. Gateway 网关
1. Login 登录服
1. Room 房间服
1. Center 中心服
1. Match 匹配服
1. IDIPServer IDIP服
1. DataCenter 数据中心服
1. Lobby 大厅服

客户端先短连 Login, 获取 Gateway 地址，然后
客户端长连 Gateway, 如匹配进入战斗则再长连 Room.

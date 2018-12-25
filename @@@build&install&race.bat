set GOOS=windows
go install -race Login
go install -race Lobby
go install -race Room
go install -race Center
go install -race Match
go install -race IDIPServer
go install -race DataCenter
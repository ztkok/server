set GOPATH=F:\@@@@@grsm20181018\zeus;F:\@@@@@grsm20181018\server
go install Login
go install Lobby
go install Room
go install Center
go install Match
go install IDIPServer
go install DataCenter

go build Gateway
copy F:\@@@@@grsm20181018\zeus\Gateway.exe  F:\@@@@@grsm20181018\server\bin /y

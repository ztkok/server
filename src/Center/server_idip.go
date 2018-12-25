package main

import (
	"encoding/json"
	"idip"
	"io/ioutil"
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
	log "github.com/cihub/seelog"
)

// 启动idip服务
func (srv *Server) startIDIP() {
	api := rest.NewApi()
	api.Use(rest.DefaultCommonStack...)
	router, err := rest.MakeRouter(
		rest.Post("/idip", srv.idipQuery),
	)
	if err != nil {
		log.Error(err)
	}
	api.SetApp(router)
	err = http.ListenAndServe(srv.idipAddr+":"+srv.idipPort, api.MakeHandler())
	if err != nil {
		log.Error("listen error", err)
	}
}

// 处理idip指令
func (srv *Server) idipQuery(w rest.ResponseWriter, r *rest.Request) {
	req := &idip.DataPaket{}
	ack := &idip.DataPaket{}
	defer func() {
		iCmd, ok := ack.Body.(idip.ICmdIDGetter)
		if ok {
			ack.Head.CmdID = iCmd.GetCmdID()
		}

		data, _ := json.Marshal(ack)
		ack.Head.PacketLen = uint32(len(data))
		w.Header().Set("Content-Type", "text/html;charset=utf-8")
		w.WriteJson(ack)
	}()

	data, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		log.Error(err)
		ack.Head.Result = idip.ErrReadBodyFailed
		ack.Head.RetErrMsg = err.Error()
		return
	}

	if len(data) <= 12 { //len("data_packet=")
		log.Error("数据包格式错误", string(data))
		ack.Head.Result = idip.ErrJSONDecodeFailed
		ack.Head.RetErrMsg = "数据包格式错误"
		return
	}

	var decodeRet int32
	var decodeErrString string
	req, decodeRet, decodeErrString, err = idip.Decode(data[12:])
	if err != nil {
		log.Error(err, string(data))
		ack.Head.Result = decodeRet
		ack.Head.RetErrMsg = decodeErrString
		return
	}

	ack.Head = req.Head
	iDo, ok := req.Body.(idip.IDo)
	if !ok {
		log.Error("无法执行命令, 缺少接口 ", string(data))
		ack.Head.Result = idip.ErrDoCmdFailed
		ack.Head.RetErrMsg = "无法执行命令, 缺少接口"
		return
	}
	ack.Body, ack.Head.Result, err = iDo.Do()
	if err != nil {
		ack.Head.RetErrMsg = err.Error()
	}
}

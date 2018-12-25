package main

import (
	"encoding/json"
	"idip"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/ant0ine/go-json-rest/rest"
	log "github.com/cihub/seelog"
	"github.com/spf13/viper"
)

var srvInst *Server

// GetSrvInst 获取服务器全局实例
func GetSrvInst() *Server {
	if srvInst == nil {
		srvInst = &Server{}

		srvInst.addr = viper.GetString("IDIP.Addr")
		srvInst.port = viper.GetString("IDIP.Port")
		srvInst.route = make(map[uint32]string)

		wxAddr := viper.GetString("IDIP.WXAddr")
		wxPort := viper.GetString("IDIP.WXPort")
		srvInst.route[1] = wxAddr + ":" + wxPort

		qqAddr := viper.GetString("IDIP.QQAddr")
		qqPort := viper.GetString("IDIP.QQPort")
		srvInst.route[2] = qqAddr + ":" + qqPort
	}

	return srvInst
}

// Server 中心服务器
type Server struct {
	addr string
	port string

	route map[uint32]string
}

// Start 启动服务器
func (srv *Server) Start() {
	api := rest.NewApi()
	api.Use(rest.DefaultCommonStack...)
	router, err := rest.MakeRouter(
		rest.Post("/idip", srv.idipQuery),
	)
	if err != nil {
		log.Error(err)
	}
	api.SetApp(router)
	err = http.ListenAndServe(srv.addr+":"+srv.port, api.MakeHandler())
	if err != nil {
		log.Error("listen error", err)
		return
	}
}

// 转发idip指令到具体的Center服务器
func (srv *Server) idipQuery(w rest.ResponseWriter, r *rest.Request) {
	req := &idip.DataPaket{}
	ack := &idip.DataPaket{}
	skip := false
	defer func() {
		if skip {
			return
		}

		data, _ := json.Marshal(ack)
		ack.Head.PacketLen = uint32(len(data))
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
		log.Error("数据包格式错误")
		ack.Head.Result = idip.ErrJSONDecodeFailed
		ack.Head.RetErrMsg = "数据包格式错误"
		return
	}

	var decodeRet int32
	var decodeErrString string
	req, decodeRet, decodeErrString, err = idip.Decode(data[12:])
	if err != nil {
		log.Error(err)
		ack.Head.Result = decodeRet
		ack.Head.RetErrMsg = decodeErrString
		return
	}

	ack.Head = req.Head

	areaID := req.Body.(idip.IAreaIDGetter).GetAreaID()
	tarAddr := srv.route[areaID]
	resp, err := http.Post(tarAddr+"/idip", "application/x-www-form-urlencoded", strings.NewReader(string(data)))
	if err != nil {
		log.Error("转发请求失败,", tarAddr)
		ack.Head.Result = idip.ErrTransportFailed
		ack.Head.RetErrMsg = err.Error()
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(err)
		ack.Head.Result = idip.ErrReadBodyFailed
		ack.Head.RetErrMsg = err.Error()
		return
	}

	w.(http.ResponseWriter).Write(body)
	skip = true
}

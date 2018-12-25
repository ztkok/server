package main

import (
	"net/http"
	"strings"
	"zeus/tlog"
	"zeus/zlog"

	_ "net/http/pprof"

	log "github.com/cihub/seelog"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func main() {
	viper.SetConfigFile("../res/config/server.json")
	if err := viper.ReadInConfig(); err != nil {
		panic("加载配置文件失败")
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		if strings.Contains(e.Name, "server.json") {
			log.Info("Reload server.json success")
		}
	})

	logDir := viper.GetString("Config.LogDir")
	logLevel := viper.GetString("Config.LogLevel")
	zlog.Init(logDir, logLevel)
	defer log.Flush()
	defer tlog.Flush()

	if viper.GetBool("Config.Debug") {
		go func() {
			http.ListenAndServe("localhost:7066", nil)
		}()
	}

	GetSrvInst().Run()
}

package main

import (
    "flag"
    "log"

    "goatherd/config"
    "goatherd/model/master"
)

var configFile string

func init() {
    flag.StringVar(&configFile, "config", "./etc/goatherd.conf", "goatherd配置")
}

func main() {
    flag.Parse()

    var conf, err = config.LoadNewConfig(configFile)
    if err != nil {
        log.Fatal("master config load faild:" + err.Error())
    }
    master.StartNewRpcServer(conf.Master[0])
}

package main

import (
    "flag"
    "goatherd/config"
    "goatherd/master"
    "log"
    "os"
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

    var masterConf = conf.Master[0]
    if output, err := os.OpenFile(masterConf.LogFile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666); err != nil {
        log.Fatal("master log file open faild")
    } else {
        log.SetOutput(output)
        log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
    }
    master.StartNewRpcServer(masterConf)
}

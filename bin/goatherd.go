package main

import (
    "errors"
    "flag"
    "log"
    "sunteng/commons/util/toml_util"

    "goatherd/collie"
)

var (
    port       int
    name       string
    join       string
    sheepFile  string
    configFile string
)

func init() {
    flag.IntVar(&port, "port", 0, "collie服务端口")
    flag.StringVar(&name, "name", "collie", "collie服务name")
    flag.StringVar(&join, "join", "", "lead服务地址")
    flag.StringVar(&configFile, "config", "/etc/goatherd.toml", "collie配置文件路径")
    flag.StringVar(&sheepFile, "sheep_config", "", "sheep配置文件路径")
}

type Config struct {
    collie.Config
}

func LoadConfig(file string) (conf *Config, err error) {
    conf = new(Config)
    log.Printf("load config : %s", file)
    if err = toml_util.DecodeFile(file, conf); err != nil {
        err = errors.New("config decode faild:" + err.Error())
        return
    }
    return
}

func main() {
    flag.Parse()

    var err error
    defer func() {
        if err != nil {
            log.Print("Error : ", err.Error())
        }
    }()

    conf, err := LoadConfig(configFile)
    if err != nil {
        return
    }
    conf.ConfigPath = configFile

    if sheepFile != "" {
        sheepConf, err := LoadConfig(sheepFile)
        if err != nil {
            return
        }
        conf.ContexConfig = sheepConf.ContexConfig
    }

    if port != 0 {
        conf.Http.Port = port
    }

    // log.Fatalf("config : %+v", conf)
    log.Print("goatherd will start at:", conf.Http.Port)
    log.Fatal(collie.NewHttpServe(conf.Config, join))
}

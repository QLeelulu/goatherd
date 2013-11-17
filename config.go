package goatherd

import (
    "errors"
    "github.com/BurntSushi/toml"
    // "io/ioutil"
    "fmt"
    "os"
    "path"
)

const (
    configFileName = "goatherd.conf"
)

type Config struct {
    Master   MasterConfig
    Programs map[string]Program
}

type MasterConfig struct {
    Host string
    Port int
}

func loadConf() (*Config, error) {
    configPaths := []string{
        path.Join("./", configFileName),
        path.Join("./etc/", configFileName),
        path.Join("/etc/", configFileName),
    }
    var conf Config
    var err error
    for _, configFile := range configPaths {
        _, err = toml.DecodeFile(configFile, &conf)
        if err != nil {
            if os.IsNotExist(err) {
                continue
            } else {
                return nil, err
            }
        } else {
            break
        }
    }
    if err != nil {
        if os.IsNotExist(err) {
            return nil, errors.New(fmt.Sprintf("Can not found config file, found at %+v.", configPaths))
        } else {
            return nil, err
        }
    }

    return &conf, nil
}

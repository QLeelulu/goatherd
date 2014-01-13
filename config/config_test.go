package config

import (
    // "log"
    "github.com/BurntSushi/toml"

    "testing"
)

func TestLoadConf(t *testing.T) {
    /* conf, err := loadConf() */
    /* if err != nil { */
    /* t.Fatal(err) */
    /* } */
    /* for _, master := range conf.Master { */
    /* t.Logf("master : %+v\n", master.Id) */
    /* for _, slave := range master.Slave { */
    /* t.Logf("  slave : %+v\n", slave.Id) */
    /* for _, process := range slave.Process { */
    /* t.Logf("    process : %+v\n", process.Id) */
    /* } */
    /* } */
    /* } */
}

func TestSlice(t *testing.T) {
    file := "/Users/cici/goatherd/etc/master.conf"
    conf := []MasterConfig{}
    if _, err := toml.DecodeFile(file, conf); err != nil {
        t.Fatal("decode faild:", err.Error())
    }
}

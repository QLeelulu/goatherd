package goatherd

import (
    // "log"
    "testing"
)

func TestLoadConf(t *testing.T) {
    conf, err := loadConf()
    if err != nil {
        t.Fatal(err)
    }
    t.Logf("%+v\n", conf)
}

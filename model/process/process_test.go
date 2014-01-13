package process

import (
    "io/ioutil"
    "log"
    "testing"
    "time"

    "goatherd/config"
)

var TestScript = []byte(`ls /data`)
var TestPath = "/data/resources/test.sh"

func init() {
    if err := ioutil.WriteFile(TestPath, TestScript, 0777); err != nil {
        log.Fatal("init script faild:", err.Error())
    }
}

var conf = config.ProcessConfig{
    Id:               "test",
    Command:          "sh " + TestPath,
    AutoRestart:      true,
    AutoRestartDelay: 1,
}

func TestReload(t *testing.T) {
    var ctrl, err = NewCtrl(conf)
    if err != nil {
        t.Fatalf("newprocessctrl faild:%s:%+v", err.Error(), conf)
    }

    if err = ctrl.start(); err != nil {
        t.Fatalf("start faild:%s:%+v", err.Error(), *ctrl)
    }
    time.Sleep(time.Second * 5)
}

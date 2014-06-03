package process

import (
    "fmt"
    "io/ioutil"
    "testing"

    "goatherd/config"
)

var script = "test.sh"

var conf = config.ProcessConfig{
    Id:      "test",
    Command: fmt.Sprintf("sh %s", script),
}

func init() {
    var cmd = "for i in `seq 1 3`;do echo hello;sleep 1;done"
    /* var cmd = "ls /data" */
    if err := ioutil.WriteFile(script, []byte(cmd), 0777); err != nil {
        panic("script init faild")
    }
}

func TestStart(t *testing.T) {
    var ctrl, err = NewCtrl(conf)
    if err != nil {
        t.Fatalf("newprocessctrl faild:%s:%+v", err.Error(), conf)
    }

    if err = ctrl.Start(); err != nil {
        t.Fatalf("start faild:%s:%+v", err.Error(), *ctrl)
    }

    if ctrl.GetStatus() != PROCESS_STATUS_STARTED {
        t.Fatal("start faild")
    }

    if err = ctrl.Stop(); err != nil {
        t.Fatalf("stop faild:%s:%+v", err.Error(), *ctrl)
    }

    if ctrl.GetStatus() != PROCESS_STATUS_STOPED {
        t.Fatal("stop faild")
    }
}

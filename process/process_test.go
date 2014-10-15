package process

import (
    "fmt"
    "io/ioutil"
    "os"
    "testing"

    "sunteng/commons/constant"
)

var script = "test.sh"

var conf = Config{
    Name:    "test",
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
    defer os.Remove("test.sh")
    var ctrl, err = NewCtrl(conf)
    if err != nil {
        t.Fatalf("newprocessctrl faild:%s:%+v", err.Error(), conf)
    }

    if err = ctrl.Start(); err != nil {
        t.Fatalf("start faild:%s:%+v", err.Error(), *ctrl)
    }

    if ctrl.GetStatus() != constant.STATUS_STARTED {
        t.Fatal("start faild")
    }

    if err = ctrl.Stop(); err != nil {
        t.Fatalf("stop faild:%s:%+v", err.Error(), *ctrl)
    }

    if ctrl.GetStatus() != constant.STATUS_STOPPED {
        t.Fatal("stop faild")
    }
}

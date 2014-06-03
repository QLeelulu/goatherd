package guard

import (
    "goatherd/config"
    "goatherd/constant"
    "io/ioutil"
    "testing"
    "time"
)

var TestPath = "/usr/local/goatherd/test"

func TestGuard(t *testing.T) {
    client := NewClientDefault()
    arg := NewArg("guard", constant.ACT_START)
    ret := NewRet()
    if err := client.Call(*arg, ret); err != nil {
        t.Fatal("start faild:", err.Error())
    } else {
        t.Log("start done")
    }

    time.Sleep(time.Second * 1)
    arg = NewArg("guard", constant.ACT_GET)
    ret = NewRet()
    if err := client.Call(*arg, ret); err != nil {
        t.Fatal("get faild:", err.Error())
    } else {
        t.Logf("get done:%+v", *ret.Conf)
    }

    time.Sleep(time.Second * 1)
    arg = NewArg("guard", constant.ACT_STOP)
    ret = NewRet()
    if err := client.Call(*arg, ret); err != nil {
        t.Fatal("stop faild:", err.Error())
    } else {
        t.Logf("stop done")
    }

    time.Sleep(time.Second * 1)
}

func TestGuardServerStart(t *testing.T) {
    client := NewClientDefault()
    arg := NewArg("guard", constant.ACT_START)
    ret := NewRet()
    if err := client.Call(*arg, ret); err != nil {
        t.Fatal("start faild:", err.Error())
    } else {
        t.Log("start done")
    }
}

func TestGuardServerStatus(t *testing.T) {
    client := NewClientDefault()
    arg := NewArg("guard", constant.ACT_STATUS)
    ret := NewRet()
    if err := client.Call(*arg, ret); err != nil {
        t.Fatal("status guard faild:", err.Error())
    } else {
        t.Logf("status guard done:%+v", ret.Status)
    }
}

func TestGuardServerGet(t *testing.T) {
    client := NewClientDefault()
    arg := NewArg("guard", constant.ACT_GET)
    ret := NewRet()
    if err := client.Call(*arg, ret); err != nil {
        t.Fatal("get guard faild:", err.Error())
    } else {
        t.Logf("get guard done:%+v", *ret.Conf)
    }
}

func TestGuardServerStop(t *testing.T) {
    client := NewClientDefault()
    arg := NewArg("guard", constant.ACT_STOP)
    ret := NewRet()
    if err := client.Call(*arg, ret); err != nil {
        t.Fatal("stop guard faild:", err.Error())
    } else {
        t.Log("stop guard done")
    }
}

func TestProcessAdd(t *testing.T) {
    process_file := TestPath + "/process_test.sh"
    script := "while true;do echo `date` process_test;sleep 1;done"
    err := ioutil.WriteFile(process_file, []byte(script), 0777)
    if err != nil {
        t.Fatal("write process script faild")
    }

    client := NewClientDefault()
    arg := NewArg("guard:process", constant.ACT_ADD)
    processConf := &config.ProcessConfig{
        Id:      "process",
        Command: "sh " + process_file,
        FileOut: TestPath + "/process_test.log",
    }
    arg.ProcessArgs[processConf.Id] = &ProcessArg{
        Id:   processConf.Id,
        Conf: processConf,
    }
    ret := NewRet()
    if err := client.Call(*arg, ret); err != nil {
        t.Fatal("add process faild:", err.Error())
    } else {
        t.Log("add process done")
    }
}

func TestProcessGet(t *testing.T) {
    client := NewClientDefault()
    arg := NewArg("guard:process", constant.ACT_GET)
    ret := NewRet()
    if client.Call(*arg, ret); ret.Err != nil {
        t.Fatal("get config process faild:", ret.Err.Error())
    } else {
        t.Logf("get config process done:%+v", *ret.ProcessRets["process"].Conf)
    }
}

func TestProcessStatus(t *testing.T) {
    client := NewClientDefault()
    arg := NewArg("guard:process", constant.ACT_STATUS)
    ret := NewRet()
    if err := client.Call(*arg, ret); err != nil {
        t.Fatal("status process faild:", err.Error())
    } else {
        t.Logf("status process done:%+v", ret.ProcessRets["process"].Status)
    }
}

func TestProcessStart(t *testing.T) {
    client := NewClientDefault()
    arg := NewArg("guard:process", constant.ACT_START)
    ret := NewRet()
    if err := client.Call(*arg, ret); err != nil {
        t.Fatal("start process faild:", err.Error())
    } else {
        t.Logf("start process done")
    }
}

func TestProcessStop(t *testing.T) {
    client := NewClientDefault()
    arg := NewArg("guard:process", constant.ACT_STOP)
    ret := NewRet()
    if err := client.Call(*arg, ret); err != nil {
        t.Fatal("stop process faild:", err.Error())
    } else {
        t.Logf("stop process done")
    }
}

func TestProcessDel(t *testing.T) {
    client := NewClientDefault()
    arg := NewArg("guard:process", constant.ACT_DEL)
    ret := NewRet()
    if err := client.Call(*arg, ret); err != nil {
        t.Fatal("status faild:", err.Error())
    } else {
        t.Logf("status done:%+v", ret.ProcessRets["process"].Status)
    }
}

func TestProcessUpdate(t *testing.T) {
    process_file := TestPath + "/process_test_new.sh"
    script := "while true;do echo `date` process new;sleep 1;done"
    err := ioutil.WriteFile(process_file, []byte(script), 0777)
    if err != nil {
        t.Fatal("write process script faild")
    }

    client := NewClientDefault()
    arg := NewArg("guard:process", constant.ACT_UPDATE)
    processConf := &config.ProcessConfig{
        Id:      "process",
        Command: "sh " + process_file,
        FileOut: TestPath + "process_test_new.log",
    }
    arg.ProcessArgs[processConf.Id] = &ProcessArg{
        Id:   processConf.Id,
        Conf: processConf,
    }
    ret := NewRet()
    if err := client.Call(*arg, ret); err != nil {
        t.Fatal("update process faild:", err.Error())
    } else {
        t.Log("update process done")
    }
}

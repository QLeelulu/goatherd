package master

import (
    "goatherd/config"
    "goatherd/constant"
    "goatherd/guard"
    "goatherd/library/sshclient"
    "io/ioutil"
    "net/rpc"
    "testing"
)

var TestDir = "/usr/local/goatherd/test"
var TestHost = "127.0.0.1"
var TestPort = 8019

func TestGuardAdd(t *testing.T) {
    var arg = NewArg("master:guard", constant.ACT_ADD)
    guardConf := &config.GuardConfig{
        Id:      "guard",
        Host:    TestHost,
        Port:    8020,
        LogFile: TestDir + "/guard.log",
        PidFile: TestDir + "/guard.pid",
    }
    arg.GuardArgs[guardConf.Id] = &guard.Arg{
        Id:   guardConf.Id,
        Conf: guardConf,
    }
    var ret = NewRet("master")
    var client = NewClient(TestHost, TestPort)
    if client.Call(*arg, ret); ret.Err != nil {
        t.Fatal("add faild:", ret.Err.Error())
    } else {
        t.Log("add done")
    }
}

func TestGuardGet(t *testing.T) {
    var arg = NewArg("master:guard", constant.ACT_GET)
    var ret = NewRet("master")
    var client = NewClient(TestHost, TestPort)
    if client.Call(*arg, ret); ret.Err != nil {
        t.Fatal("get faild:", ret.Err.Error())
    } else {
        t.Logf("get done:%+v", *ret.GuardRets["guard"].Conf)
    }
}

func TestProcessGet(t *testing.T) {
    var arg = NewArg("master:guard:process", constant.ACT_GET)
    var ret = NewRet("master")
    var client = NewClient(TestHost, TestPort)
    if client.Call(*arg, ret); ret.Err != nil {
        t.Fatal("get faild:", ret.Err.Error())
    } else {
        t.Logf("get done:%+v", *ret.GuardRets["guard"].ProcessRets["process"].Conf)
    }
}

func TestProcessStart(t *testing.T) {
    var arg = NewArg("master:guard:process", constant.ACT_START)
    var ret = NewRet("master")
    var client = NewClient(TestHost, TestPort)
    if client.Call(*arg, ret); ret.Err != nil {
        t.Fatal("start faild:", ret.Err.Error())
    } else {
        t.Log("start done")
    }
}

func TestProcessStop(t *testing.T) {
    var arg = NewArg("master:guard:process", constant.ACT_STOP)
    var ret = NewRet("master")
    var client = NewClient(TestHost, TestPort)
    if client.Call(*arg, ret); ret.Err != nil {
        t.Fatal("stop faild:", ret.Err.Error())
    } else {
        t.Log("stop done")
    }
}

func TestProcessStatus(t *testing.T) {
    var arg = NewArg("master:guard:process", constant.ACT_STATUS)
    var ret = NewRet("master")
    var client = NewClient(TestHost, TestPort)
    if client.Call(*arg, ret); ret.Err != nil {
        t.Fatal("status faild:", ret.Err.Error())
    } else {
        t.Logf("status:%#v", ret.GuardRets["guard"].ProcessRets["process"].Status)
    }
}

var TestPath = "/usr/local/goatherd/test"

func TestProcessAdd(t *testing.T) {
    process_file := TestPath + "/process_test.sh"
    script := "while true;do echo `date` process_test;sleep 1;done"
    err := ioutil.WriteFile(process_file, []byte(script), 0777)
    if err != nil {
        t.Fatal("write process script faild")
    }

    var ret = NewRet("master")
    var arg = Arg{
        Id:     "master:guard:process",
        Action: constant.ACT_ADD,
        GuardArgs: map[string]*guard.Arg{
            "guard": &guard.Arg{
                Id: "guard",
                ProcessArgs: map[string]*guard.ProcessArg{
                    "process": &guard.ProcessArg{
                        Id: "process",
                        Conf: &config.ProcessConfig{
                            Id:      "process",
                            Command: "sh " + process_file,
                            FileOut: TestPath + "/process_test.log",
                        },
                    },
                },
            },
        },
    }

    var client = NewClient(TestHost, TestPort)
    if client.Call(arg, ret); ret.Err != nil {
        t.Fatal("get faild:", ret.Err.Error())
    } else {
        t.Log("get done")
    }
}
func TestGuardStart(t *testing.T) {
    var arg = NewArg("master:guard", constant.ACT_START)
    var ret = NewRet("master")
    var client = NewClient(TestHost, TestPort)
    if client.Call(*arg, ret); ret.Err != nil {
        t.Fatal("start faild:", ret.Err.Error())
    } else {
        t.Log("start done")
    }
}

func TestGuardStatus(t *testing.T) {
    var arg = NewArg("master:guard", constant.ACT_STATUS)
    var ret = NewRet("master")
    var client = NewClient(TestHost, TestPort)
    if client.Call(*arg, ret); ret.Err != nil {
        t.Fatal("status faild:", ret.Err.Error())
    } else {
        t.Log("status done:", ret.GuardRets["guard"].Status)
    }
}

func TestGuardStop(t *testing.T) {
    var arg = NewArg("master:guard", constant.ACT_STOP)
    var ret = NewRet("master")
    var client = NewClient(TestHost, TestPort)
    if client.Call(*arg, ret); ret.Err != nil {
        t.Fatal("stop faild:", ret.Err.Error())
    } else {
        t.Log("stop done")
    }
}

func TestEvalTest(t *testing.T) {
    var arg = "ifconfig"
    var ret = new(int)
    if client, err := rpc.Dial(SOCKET_TYPE, "127.0.0.1:8019"); err == nil {
        defer client.Close()
        if err = client.Call("Server.EvalTest", arg, ret); err != nil {
            t.Fatal("call faild:", err.Error())
        }
    } else {
        t.Fatal("dail faild")
    }
}

func TestSsh(t *testing.T) {
    client := &sshclient.Handel{
        Host:       "127.0.0.1",
        Port:       "22",
        User:       "cici",
        PrivateKey: "/Users/cici/.ssh/id_dsa",
        FileOut:    "test.out",
        FileErr:    "test.err",
    }
    arg := "source /etc/profile;ifconfig"
    if err := client.Run(arg); err != nil {
        t.Fatal("ssh run : " + err.Error())
    }
}

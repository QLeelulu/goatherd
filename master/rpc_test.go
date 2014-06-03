package master

import (
    "errors"
    "goatherd/config"
    "goatherd/constant"
    "goatherd/guard"
    "testing"
)

func TestGoString(t *testing.T) {
    var ret = &Ret{
        Id:     "master",
        Err:    errors.New("shit"),
        Status: constant.PROCESS_STATUS_STOPED,
        Conf:   &config.MasterConfig{},
        GuardRets: map[string]*guard.Ret{
            "guard": &guard.Ret{
                Id:     "guard",
                Status: constant.PROCESS_STATUS_STARTED,
                ProcessRets: map[string]*guard.ProcessRet{
                    "process": &guard.ProcessRet{
                        Id:     "process",
                        Status: constant.PROCESS_STATUS_STARTING,
                    },
                },
            },
        },
    }
    t.Logf("ret:%#v", ret)
}

package master

import (
    "goatherd/config"
    "goatherd/model/slave"
)

type Arg struct {
    Action    int
    IsTarget  bool
    Id        string
    Conf      *config.MasterConfig
    SlaveArgs map[string]slave.Arg
}

func NewArg() *Arg {
    return &Arg{
        SlaveArgs: make(map[string]slave.Arg),
    }
}

type Ret struct {
    Err       error
    Id        string
    Conf      *config.MasterConfig
    SlaveRets map[string]slave.Ret
}

func NewRet() *Ret {
    return &Ret{
        SlaveRets: make(map[string]slave.Ret),
    }
}

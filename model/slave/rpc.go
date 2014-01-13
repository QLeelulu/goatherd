package slave

import (
    "goatherd/config"
    "goatherd/model/process"
)

type Arg struct {
    Action      int
    Id          string
    Conf        *config.SlaveConfig
    ProcessArgs map[string]process.Arg
}

func NewArg(name string) *Arg {
    return &Arg{
        Id:          name,
        ProcessArgs: make(map[string]process.Arg),
    }
}

type Ret struct {
    Err         error
    Id          string
    ProcessRets map[string]process.Ret
}

func NewRet() *Ret {
    return &Ret{
        ProcessRets: make(map[string]process.Ret),
    }
}

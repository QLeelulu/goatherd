package guard

import (
    "fmt"
    "goatherd/config"
    "goatherd/constant"
    "strings"
)

type ProcessArg struct {
    Id       string
    Action   constant.ACT_ID
    IsTarget bool
    Conf     *config.ProcessConfig
}

func (this *ProcessArg) GoString() (output string) {
    output += fmt.Sprintf("Id:%s ", this.Id)
    output += fmt.Sprintf("Action:%#v ", this.Action)
    output += fmt.Sprintf("IsTarget:%t ", this.IsTarget)
    output += fmt.Sprintf("Conf:%+v ", this.Conf)
    return
}

type ProcessRet struct {
    Id     string
    Err    error
    Status constant.PROCESS_STATUS
    Conf   *config.ProcessConfig
}

func (this *ProcessRet) GoString() (output string) {
    output += fmt.Sprintf("Id:%s ", this.Id)
    if this.Err == nil {
        output += fmt.Sprintf("Err:nil ")
    } else {
        output += fmt.Sprintf("Err:%s ", this.Err.Error())
    }
    output += fmt.Sprintf("Status:%#v ", this.Status)
    output += fmt.Sprintf("Conf:%+v ", this.Conf)
    return
}

type Arg struct {
    Action      constant.ACT_ID
    Id          string
    IsTarget    bool
    Conf        *config.GuardConfig
    ProcessArgs map[string]*ProcessArg
}

func (this *Arg) GoString() (output string) {
    output += fmt.Sprintf("Id:%s ", this.Id)
    output += fmt.Sprintf("Action:%#v ", this.Action)
    output += fmt.Sprintf("IsTarget:%t ", this.IsTarget)
    output += fmt.Sprintf("Conf:%+v ", this.Conf)

    output += "ProcessArgs:{"
    for id, parg := range this.ProcessArgs {
        output += fmt.Sprintf("%s:{%#v} ", id, parg)
    }
    output += "}"
    return
}

func NewArg(name string, action constant.ACT_ID) *Arg {
    return &Arg{
        Id:          name,
        Action:      action,
        ProcessArgs: make(map[string]*ProcessArg),
    }
}

func (this Arg) CheckTarget() bool {
    return this.IsTarget || (this.Id != "" && strings.Index(this.Id, ":") == -1)
}

type Ret struct {
    Id          string
    Err         error
    Status      constant.PROCESS_STATUS
    Conf        *config.GuardConfig
    ProcessRets map[string]*ProcessRet
}

func (this *Ret) GoString() (output string) {
    output += fmt.Sprintf("Id:%s ", this.Id)
    if this.Err == nil {
        output += fmt.Sprintf("Err:nil ")
    } else {
        output += fmt.Sprintf("Err:%s ", this.Err.Error())
    }
    output += fmt.Sprintf("Status:%#v ", this.Status)
    output += fmt.Sprintf("Conf:%+v ", this.Conf)

    output += "ProcessRets:{"
    for id, gret := range this.ProcessRets {
        output += fmt.Sprintf("%s:{%#v} ", id, gret)
    }
    output += "}"
    return
}

func NewRet() *Ret {
    return &Ret{
        ProcessRets: make(map[string]*ProcessRet),
    }
}

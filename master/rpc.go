package master

import (
    "fmt"
    "goatherd/config"
    "goatherd/constant"
    "goatherd/guard"
    "strings"
)

type Arg struct {
    Id        string
    Action    constant.ACT_ID
    IsTarget  bool
    Conf      *config.MasterConfig
    GuardArgs map[string]*guard.Arg
}

func (this *Arg) GoString() (output string) {
    output += fmt.Sprintf("Id:%s ", this.Id)
    output += fmt.Sprintf("Action:%#v ", this.Action)
    output += fmt.Sprintf("IsTarget:%t ", this.IsTarget)
    output += fmt.Sprintf("Conf:%+v ", this.Conf)

    output += "GuardArgs:{"
    for id, garg := range this.GuardArgs {
        output += fmt.Sprintf("%s:{%#v} ", id, garg)
    }
    output += "}"
    return
}

func (this Arg) CheckTarget() bool {
    return this.IsTarget || (this.Id != "" && strings.Index(this.Id, ":") == -1)
}

func NewArg(id string, action constant.ACT_ID) *Arg {
    return &Arg{
        Id:        id,
        Action:    action,
        GuardArgs: make(map[string]*guard.Arg),
    }
}

/* func CloneArgWithAction(action int, arg Arg) (newArg Arg) { */
/* newArg = Arg{ */
/* Id:        arg.Id, */
/* Action:    action, */
/* GuardArgs: make(map[string]guard.Arg), */
/* } */
/* for id, garg := range arg.GuardArgs { */
/* newArg.GuardArgs[id] = guard.CloneArgWithAction(action, garg) */
/* } */
/* return */
/* } */

type Ret struct {
    Id        string
    Err       error
    Status    constant.PROCESS_STATUS
    Conf      *config.MasterConfig
    GuardRets map[string]*guard.Ret
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

    output += "GuardRets:{"
    for id, gret := range this.GuardRets {
        output += fmt.Sprintf("%s:{%#v} ", id, gret)
    }
    output += "}"
    return
}

func NewRet(id string) *Ret {
    return &Ret{
        GuardRets: make(map[string]*guard.Ret),
    }
}

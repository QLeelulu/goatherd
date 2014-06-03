package guard

import (
    "errors"
    "fmt"
    "log"
    "net"
    "net/rpc"
    "strconv"
    "strings"
    "sync"

    "goatherd/config"
    "goatherd/constant"
    "goatherd/process"
)

const (
    SOCKET_TYPE = "tcp"
)

type Server struct {
    Id           string
    mutex        sync.RWMutex
    conf         config.GuardConfig
    ProcessCtrls map[string]*process.Ctrl
}

func NewServer(id string, port int) *Server {
    return &Server{
        Id: id,
        conf: config.GuardConfig{
            Id:   id,
            Port: port,
        },
        ProcessCtrls: make(map[string]*process.Ctrl),
    }
}

// guardId:processid:processid
func (this *Server) checkArg(arg *Arg) (err error) {
    //check guard
    segs := strings.Split(arg.Id, ":")

    guardId := segs[0]
    if guardId != this.Id && guardId != "" && arg.Action != constant.ACT_ADD {
        return errors.New("guard check id not match:" + guardId)
    }
    arg.Id = guardId

    //check process
    if len(segs) < 2 {
        arg.IsTarget = true
    } else if len(segs) > 2 {
        err = errors.New(fmt.Sprintf("bad process target id:arg.Id(%s),server.Id(%s)", arg.Id, this.Id))
    } else {
        processIds := []string{}
        if processId := segs[1]; processId != "" {
            if _, ok := this.ProcessCtrls[processId]; ok || arg.Action == constant.ACT_ADD {
                processIds = append(processIds, processId)
            } else {
                err = errors.New("process check id not found:" + processId)
            }
        } else {
            for processId, _ = range this.ProcessCtrls {
                processIds = append(processIds, processId)
            }
        }
        //expand
        for _, processId := range processIds {
            parg, ok := arg.ProcessArgs[processId]
            if !ok {
                parg = &ProcessArg{
                    Id:     processId,
                    Action: arg.Action,
                }
                arg.ProcessArgs[processId] = parg
            } else {
                parg.Id = processId
                parg.Action = arg.Action
            }
            if len(segs) == 2 {
                parg.IsTarget = true
            }
        }
    }
    log.Printf("guard arg : %#v", arg)
    return
}

func (this *Server) Eval(arg Arg, ret *Ret) (err error) {
    ret.ProcessRets = make(map[string]*ProcessRet)
    if err = this.checkArg(&arg); err != nil {
        return
    }
    ret.Id = arg.Id

    this.mutex.Lock()
    defer this.mutex.Unlock()
    if arg.IsTarget {
        switch arg.Action {
        case constant.ACT_GET:
            ret.Conf = &this.conf
        default:
            err = errors.New("target should be process,guard id:" + arg.Id)
        }
        if err != nil {
            ret.Err = err
        }
    } else {
        for id, parg := range arg.ProcessArgs {
            pret := new(ProcessRet)
            if ctrl, ok := this.ProcessCtrls[id]; ok {
                switch parg.Action {
                case constant.ACT_UPDATE:
                    if parg.Conf == nil {
                        err = errors.New("process config missed")
                    } else {
                        err = ctrl.Update(*parg.Conf)
                    }
                case constant.ACT_START:
                    err = ctrl.Start()
                case constant.ACT_STOP:
                    err = ctrl.Stop()
                case constant.ACT_GET:
                    _conf := ctrl.GetConfig()
                    pret.Conf = &_conf
                case constant.ACT_STATUS:
                    pret.Status = ctrl.GetStatus()
                case constant.ACT_ADD:
                    err = errors.New("process already exist")
                case constant.ACT_DEL:
                    delete(this.ProcessCtrls, id)
                }
            } else {
                switch parg.Action {
                case constant.ACT_ADD:
                    if parg.Conf == nil {
                        err = errors.New("process config missed")
                    } else {
                        if err = parg.Conf.Test(); err == nil {
                            this.ProcessCtrls[id], err = process.NewCtrl(*parg.Conf)
                        }
                    }
                default:
                    err = errors.New("process id not found")
                }
            }
            pret.Err = err
            pret.Id = id
            ret.ProcessRets[id] = pret
            if pret.Err != nil && ret.Err == nil {
                ret.Err = err
            }
            log.Printf("ret:%+v", ret.ProcessRets)
        }
    }
    return
}

/* func (this *Server) EvalParallel(args Arg, ret *Ret) (err error) { */
/* //parallel call */
/* for id, pargs := range args.ProcessArgs { */
/* if ctrl, ok := this.ProcessCtrls[id]; ok { */
/* var retChan = make(chan process.Ret) */
/* go func(retChan chan process.Ret) { */
/* var pret = new(process.Ret) */
/* ctrl.Eval(pargs, pret) */
/* retChan <- *pret */
/* }(retChan) */
/* defer func(id string, retChan chan process.Ret) { */
/* ret.ProcessRets[id] = <-retChan */
/* }(id, retChan) */
/* } */
/* } */
/* return */
/* } */

func StartNewRpcServer(id string, port int) (err error) {
    var server = rpc.NewServer()
    server.Register(NewServer(id, port))

    l, err := net.Listen("tcp", ":"+strconv.Itoa(port))
    if err == nil {
        server.Accept(l)
    }
    return
}

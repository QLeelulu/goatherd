package slave

import (
    "log"
    "net"
    "net/rpc"
    "strconv"
    "sync"

    "goatherd/config"
    "goatherd/constant"
    "goatherd/model/process"
)

const (
    SOCKET_TYPE = "tcp"
)

type Server struct {
    Id           string
    Host         string
    Port         string
    PId          int
    mutex        sync.RWMutex
    conf         config.SlaveConfig
    processCtrls map[string]process.Ctrl
}

func NewServer() *Server {
    return &Server{
        processCtrls: make(map[string]process.Ctrl),
    }
}

func (this *Server) add(args Arg, ret *Ret) (err error) {
    for id, _ := range args.ProcessArgs {
        if pret, ok := ret.ProcessRets[id]; ok && pret.Err == nil {
            this.processCtrls[id] = process.Ctrl{}
        }
    }
    return
}

func (this *Server) del(args Arg, ret *Ret) (err error) {
    for id, _ := range args.ProcessArgs {
        if pret, ok := ret.ProcessRets[id]; ok && pret.Err == nil {
            delete(this.processCtrls, id)
        }
    }
    return
}

func (this *Server) CheckAlive(args Arg, ret *Ret) (err error) {
    return
}

func (this *Server) Eval(args Arg, ret *Ret) (err error) {
    //lock and unlock
    switch args.Action {
    case constant.ACT_CREATE:
        this.mutex.Lock()
        defer this.mutex.Unlock()
        this.add(args, ret)
    case constant.ACT_DESTROY:
        this.mutex.Lock()
        defer this.mutex.Unlock()
        defer this.del(args, ret)
    default:
        this.mutex.RLock()
        defer this.mutex.RUnlock()
    }
    //parallel call
    for id, pargs := range args.ProcessArgs {
        if ctrl, ok := this.processCtrls[id]; ok {
            var retChan = make(chan process.Ret)
            go func(retChan chan process.Ret) {
                var pret = new(process.Ret)
                ctrl.Eval(pargs, pret)
                retChan <- *pret
            }(retChan)
            defer func(id string, retChan chan process.Ret) {
                ret.ProcessRets[id] = <-retChan
            }(id, retChan)
        }
    }
    return
}

func StartNewRpcServer(port int) (err error) {
    var server = rpc.NewServer()
    server.Register(NewServer())

    l, e := net.Listen("tcp", ":"+strconv.Itoa(port))
    if e != nil {
        log.Fatal("listen error:", e)
    }
    server.Accept(l)
    return
}

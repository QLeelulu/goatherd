package master

import (
    "errors"
    "log"
    "net"
    "net/rpc"
    "strconv"
    "sync"

    "goatherd/config"
    "goatherd/constant"
    "goatherd/model/slave"
)

type Server struct {
    Id     string
    mutex  sync.RWMutex
    conf   config.MasterConfig
    slaves map[string]*slave.Client
}

func NewServer(conf config.MasterConfig) *Server {
    return &Server{
        Id:     conf.Id,
        conf:   conf,
        slaves: make(map[string]*slave.Client),
    }
}

func (this *Server) addSlave(slaveArg slave.Arg) (err error) {
    var slaveConf = slaveArg.Conf
    if slaveConf == nil {
        return errors.New("slave config needed")
    }
    this.mutex.Lock()
    defer this.mutex.Unlock()
    if _, ok := this.slaves[slaveConf.Id]; ok {
        return errors.New("slave already exist")
    }
    this.slaves[slaveConf.Id] = slave.NewClientWithSsh(
        slaveConf.Id,
        slaveConf.Host,
        slaveConf.Port,
        "",
        slaveConf.LogFile,
        slaveConf.LogFile,
    )
    return
}

func (this *Server) reloadSlave(slaveArg slave.Arg) (err error) {
    var slaveConf = slaveArg.Conf
    if slaveConf == nil {
        return errors.New("slave config needed")
    }
    this.mutex.Lock()
    defer this.mutex.Unlock()
    this.slaves[slaveArg.Id] = slave.NewClientWithSsh(
        slaveConf.Id,
        slaveConf.Host,
        slaveConf.Port,
        "",
        slaveConf.LogFile,
        slaveConf.LogFile,
    )
    return
}

func (this *Server) delSlave(slaveArg slave.Arg) (err error) {
    this.mutex.Lock()
    defer this.mutex.Unlock()
    if _, ok := this.slaves[slaveArg.Id]; !ok {
        return errors.New("slave not exist")
    }
    delete(this.slaves, slaveArg.Id)
    return
}

func (this *Server) startSlave(slaveArg slave.Arg) (err error) {
    this.mutex.Lock()
    defer this.mutex.Unlock()

    if slaveClient, ok := this.slaves[slaveArg.Id]; ok {
        err = slaveClient.SshStart()
    } else {
        err = errors.New(slaveArg.Id + " not found")
    }
    return
}

func (this *Server) stopSlave(slaveArg slave.Arg) (err error) {
    if slaveClient, ok := this.slaves[slaveArg.Id]; ok {
        err = slaveClient.SshKill()
    } else {
        err = errors.New(slaveArg.Id + " not found")
    }
    return
}

func (this *Server) callSlave(slaveArg slave.Arg, slaveRet *slave.Ret) (err error) {
    if slaveClient, ok := this.slaves[slaveArg.Id]; ok {
        err = slaveClient.Call(slaveArg, slaveRet)
    } else {
        err = errors.New(slaveArg.Id + " not found")
    }
    return
}

func (this *Server) getConfig(ret *Ret) (err error) {
    ret.Conf = &this.conf
    return
}

func (this *Server) Eval(arg Arg, ret *Ret) (err error) {
    switch arg.Action {
    case constant.ACT_GET:
        err = this.getConfig(ret)
        return
    }
    for id, sarg := range arg.SlaveArgs {
        var sret = &slave.Ret{}
        defer func(id string, sret *slave.Ret) {
            ret.SlaveRets[id] = *sret
        }(id, sret)

        switch arg.Action {
        case constant.ACT_CALL:
            err = this.callSlave(sarg, sret)
        case constant.ACT_ADD:
            err = this.addSlave(sarg)
        case constant.ACT_MOD:
            err = this.reloadSlave(sarg)
        case constant.ACT_DEL:
            err = this.delSlave(sarg)
        case constant.ACT_START:
            err = this.startSlave(sarg)
        case constant.ACT_STOP:
            err = this.stopSlave(sarg)
        default:
            println("yamidie")
        }
    }
    return
}

func StartNewRpcServer(conf config.MasterConfig) (err error) {
    var masterServer = NewServer(conf)

    var rpcServer = rpc.NewServer()
    rpcServer.Register(masterServer)

    l, e := net.Listen("tcp", ":"+strconv.Itoa(conf.Port))
    if e != nil {
        log.Fatal("listen error:", e)
    }
    log.Print("master server will start at:", conf.Port)
    rpcServer.Accept(l)
    return
}

/* func (this *Server) Eval(args Arg, rets *Ret) (err error) { */
/* for id, sarg := range args.SArgs { */
/* switch args.Action{ */
/* case constant.ACT_ADD */
/* } */
/* if slaveClient, ok := this.slaves[name]; ok { */
/* if err = slaveClient.Go(sarg); err != nil { */
/* rets.SRets[name] = slave.Ret{ */
/* Name: name, */
/* Err:  err, */
/* } */
/* } else { */
/* defer func(name string, slaveClient *slave.Client) (err error) { */
/* if sret, err := slaveClient.Wait(); err != nil { */
/* rets.SRets[name] = slave.Ret{ */
/* Name: name, */
/* Err:  err, */
/* } */
/* } else { */
/* rets.SRets[name] = *sret */
/* } */
/* return */
/* }(name, slaveClient) */
/* } */
/* } else { */
/* rets.SRets[name] = slave.Ret{ */
/* Name: name, */
/* Err:  errors.New("slave not found"), */
/* } */
/* } */
/* } */
/* return */
/* } */

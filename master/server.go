package master

import (
    "errors"
    "log"
    "net"
    "net/rpc"
    "strconv"
    "strings"
    "sync"

    "goatherd/config"
    "goatherd/constant"
    "goatherd/guard"
    "goatherd/library/sshclient"
)

type Server struct {
    Id           string
    mutex        sync.RWMutex
    conf         config.MasterConfig
    GuardClients map[string]*guard.Client
}

func NewServer(conf config.MasterConfig) *Server {
    return &Server{
        Id:           conf.Id,
        conf:         conf,
        GuardClients: make(map[string]*guard.Client),
    }
}

/* func (this *Server) addGuard(garg guard.Arg) (sclient *guard.Client, err error) { */
/* var guardConf = garg.Conf */
/* if guardConf == nil { */
/* err = errors.New("guard config needed") */
/* } */
/* this.mutex.Lock() */
/* defer this.mutex.Unlock() */

/* var ok bool */
/* if sclient, ok = this.GuardClients[guardConf.Id]; !ok { */
/* sclient = guard.NewClientWithSsh( */
/* guardConf.Id, */
/* guardConf.Host, */
/* guardConf.Port, */
/* "", */
/* guardConf.LogFile, */
/* ) */
/* this.GuardClients[guardConf.Id] = sclient */
/* } */

/* return */
/* } */

/* func (this *Server) Reload(arg Arg) (err error) { */
/* if arg.IsTarget { */
/* if arg.Conf == nil { */
/* return errors.New("guard config needed") */
/* } */

/* this.mutex.Lock() */
/* defer this.mutex.Unlock() */
/* stopArg := CloneArgWithAction(constant.ACT_STOP, arg) */
/* if _, err = this.Stop(stopArg); err != nil { */
/* return */
/* } */

/* this.conf = *arg.Conf */

/* startArg := CloneArgWithAction(constant.ACT_START, arg) */
/* if _, err = this.Start(arg); err != nil { */
/* return */
/* } */
/* } */

/* return */
/* } */

// masterId:guardid:processid
func (this *Server) checkArg(arg *Arg) (err error) {
    //check master
    segs := strings.Split(arg.Id, ":")

    masterId := segs[0]
    if masterId != this.Id && masterId != "" && arg.Action != constant.ACT_ADD {
        return errors.New("bad master target id:" + masterId)
    }
    arg.Id = masterId

    //check guard
    if len(segs) < 2 {
        arg.IsTarget = true
    } else {
        guardIds := []string{}
        if guardId := segs[1]; guardId != "" {
            if _, ok := this.GuardClients[guardId]; ok || arg.Action == constant.ACT_ADD {
                guardIds = append(guardIds, guardId)
            } else {
                err = errors.New("guard check id not found:" + guardId)
            }
        } else {
            for guardId, _ = range this.GuardClients {
                guardIds = append(guardIds, guardId)
            }
        }
        for _, guardId := range guardIds {
            newSegs := []string{guardId}
            if len(segs) >= 2 {
                newSegs = append(newSegs, segs[2:]...)
            }
            newId := strings.Join(newSegs, ":")

            garg, ok := arg.GuardArgs[guardId]
            if !ok {
                garg = guard.NewArg(newId, arg.Action)
                arg.GuardArgs[guardId] = garg
            } else {
                garg.Id = newId
                garg.Action = arg.Action
            }
            if len(segs) == 2 {
                garg.IsTarget = true
            }
        }
    }
    log.Printf("master arg : %#v", arg)
    return
}

func (this *Server) EvalTest(arg string, ret *int) (err error) {
    client := &sshclient.Handel{
        Host:       "127.0.0.1",
        Port:       "22",
        User:       "cici",
        PrivateKey: "/Users/cici/.ssh/id_dsa",
        /* FileOut:    "test.out", */
        /* FileErr:    "test.err", */
    }
    if err = client.Start(arg); err != nil {
        return
    }
    if err = client.Wait(); err != nil {
        return
    }
    return
}

func (this *Server) Eval(arg Arg, ret *Ret) (err error) {
    /* log.Printf("master eval:%+v", arg) */
    ret.GuardRets = make(map[string]*guard.Ret)
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
            err = errors.New("action not supperted for master:" + arg.Id)
        }
        ret.Err = err
    } else {
        for id, garg := range arg.GuardArgs {
            var gret = new(guard.Ret)
            if gclient, ok := this.GuardClients[id]; ok {
                switch garg.Action {
                case constant.ACT_ADD:
                    if garg.IsTarget {
                        err = errors.New("guard already exist")
                    } else {
                        err = gclient.Call(*garg, gret)
                    }
                case constant.ACT_DEL:
                    if garg.IsTarget {
                        delete(this.GuardClients, id)
                    } else {
                        err = gclient.Call(*garg, gret)
                    }
                case constant.ACT_GET:
                    if garg.IsTarget {
                        gret.Conf = gclient.Conf
                    } else {
                        err = gclient.Call(*garg, gret)
                    }
                /* case constant.ACT_UPDATE: */
                /* if garg.Conf == nil { */
                /* err = errors.New("process config missed") */
                /* } else { */
                /* fallthrough */
                /* } */
                case constant.ACT_START:
                    fallthrough
                case constant.ACT_STOP:
                    fallthrough
                case constant.ACT_STATUS:
                    fallthrough
                case constant.ACT_CALL:
                    fallthrough
                default:
                    err = gclient.Call(*garg, gret)
                }
            } else {
                switch garg.Action {
                case constant.ACT_ADD:
                    if garg.IsTarget {
                        if garg.Conf == nil {
                            err = errors.New("guard config missed")
                        } else {
                            if err = garg.Conf.Test(); err == nil {
                                this.GuardClients[id] = guard.NewClientWithConfig(garg.Conf, this.conf)
                            } else {
                                log.Print("bad guard config:", err.Error())
                            }
                        }
                    } else {
                        err = errors.New("guard not exist")
                    }
                default:
                    err = errors.New("guard id not found")
                }
            }
            gret.Err = err
            gret.Id = id
            ret.GuardRets[id] = gret
            if gret.Err != nil && ret.Err == nil {
                ret.Err = err
            }
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
/* for id, garg := range args.SArgs { */
/* switch args.Action{ */
/* case constant.ACT_ADD */
/* } */
/* if sclient, ok := this.GuardClients[name]; ok { */
/* if err = sclient.Go(garg); err != nil { */
/* rets.SRets[name] = guard.Ret{ */
/* Name: name, */
/* Err:  err, */
/* } */
/* } else { */
/* defer func(name string, sclient *guard.Client) (err error) { */
/* if gret, err := sclient.Wait(); err != nil { */
/* rets.SRets[name] = guard.Ret{ */
/* Name: name, */
/* Err:  err, */
/* } */
/* } else { */
/* rets.SRets[name] = *gret */
/* } */
/* return */
/* }(name, sclient) */
/* } */
/* } else { */
/* rets.SRets[name] = guard.Ret{ */
/* Name: name, */
/* Err:  errors.New("guard not found"), */
/* } */
/* } */
/* } */
/* return */
/* } */

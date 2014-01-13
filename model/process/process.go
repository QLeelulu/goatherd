package process

import (
    "errors"
    "fmt"
    "log"
    "os"
    "os/exec"
    "strings"
    "sync"
    "syscall"
    "time"

    "goatherd/config"
    "goatherd/constant"
)

const (
    PROCESS_STATUS_INIT = iota
    PROCESS_STATUS_STARTING
    PROCESS_STATUS_START_FAILED
    PROCESS_STATUS_STARTED
    PROCESS_STATUS_STOPING
    PROCESS_STATUS_STOPED
)

type Arg struct {
    Action int
    Config *config.ProcessConfig
}

type Ret struct {
    Err   error
    Id    string
    Value interface{}
}

type Ctrl struct {
    Id     string
    Cmd    *exec.Cmd
    conf   config.ProcessConfig
    mutex  sync.Mutex
    Status int
}

func NewCtrl(conf config.ProcessConfig) (ctrl *Ctrl, err error) {
    ctrl = new(Ctrl)
    err = ctrl.reload(conf)
    return
}

func (this *Ctrl) reload(conf config.ProcessConfig) (err error) {
    //cmd
    cmdArr := strings.Split(conf.Command, " ")
    cmdId := cmdArr[0]
    cmdArgs := []string{}
    if len(cmdArr) > 1 {
        cmdArgs = cmdArr[1:]
    }
    var cmd = exec.Command(cmdId, cmdArgs...)
    //stdin
    if conf.FileIn == "" {
        cmd.Stdin = os.Stdin
    } else if fd, err := os.Open(conf.FileIn); err == nil {
        cmd.Stdin = fd
    } else {
        return errors.New("process config error : bad input file : " + err.Error())
    }
    //stdout
    if conf.FileOut == "" {
        cmd.Stdout = os.Stdout
    } else if fd, err := os.OpenFile(conf.FileOut, os.O_WRONLY|os.O_CREATE, 0666); err == nil {
        cmd.Stdout = fd
    } else {
        return errors.New("process config error : bad output file : " + err.Error())
    }
    //stderr
    if conf.FileErr == "" {
        cmd.Stderr = os.Stderr
    } else if fd, err := os.OpenFile(conf.FileErr, os.O_WRONLY|os.O_CREATE, 0666); err == nil {
        cmd.Stderr = fd
    } else {
        return errors.New("process config error : bad error file : " + err.Error())
    }
    //kill signal
    if conf.StopSignal == uint(syscall.Signal(0)) {
        conf.StopSignal = uint(syscall.SIGKILL)
    }
    this.conf = conf
    this.Cmd = cmd
    return
}

func (this *Ctrl) create(conf config.ProcessConfig) (err error) {
    if err = this.reload(conf); err != nil {
        return
    }
    if conf.AutoStart {
        err = this.start()
    }
    return
}

func (this *Ctrl) update(conf *config.ProcessConfig) (err error) {
    if conf == nil {
        return errors.New("config is needed")
    }
    //update
    err = this.reload(*conf)
    err = this.restart()
    return
}

func (this *Ctrl) start() (err error) {
    this.Status = PROCESS_STATUS_STARTING
    log.Printf("str:%+v", *this)
    err = this.Cmd.Start()
    if err != nil {
        this.Status = PROCESS_STATUS_START_FAILED
        log.Printf("start faild:%s:%+v", err.Error(), *this)
        return err
    }
    go this.watch()
    this.Status = PROCESS_STATUS_STARTED
    return nil
}

func (this *Ctrl) watch() {
    err := this.Cmd.Wait()
    fmt.Println(this.conf.Id, "has exit.")
    if err != nil {
        fmt.Printf("[%s] got error: %s\n", this.Id, err.Error())
    }

    // conf is runing
    if this.Status == PROCESS_STATUS_STARTED {
        if this.conf.AutoRestart {
            fmt.Println(this.conf.Id, "is auto restarting.")
            time.Sleep(time.Duration(this.conf.AutoRestartDelay))
            /* this.reload(this.conf) */
            this.start()
        }
    }
}

func (this *Ctrl) restart() (err error) {
    err = this.stop()
    err = this.start()
    return err
}

func (this *Ctrl) stop() error {
    if this.Cmd.Process == nil {
        return errors.New(fmt.Sprintf("%s not started.", this.Id))
    }

    if this.Cmd.ProcessState.Exited() {
        return errors.New(fmt.Sprintf("%s has stoped.", this.Id))
    }

    err := this.Cmd.Process.Signal(syscall.Signal(this.conf.StopSignal))
    return err
}

func (this *Ctrl) tail() (err error) {
    return
}

func (this *Ctrl) tailf() (err error) {
    return
}

func (this *Ctrl) Eval(args Arg, ret *Ret) (err error) {

    switch args.Action {
    case constant.ACT_START:
        err = this.start()
    case constant.ACT_STOP, constant.ACT_DESTROY:
        err = this.stop()
    case constant.ACT_RESTART:
        err = this.restart()
    case constant.ACT_UPDATE, constant.ACT_CREATE:
        err = this.update(args.Config)
    case constant.ACT_TAIL:
        err = this.tail()
    case constant.ACT_TAILF:
        err = this.tailf()
    }
    ret.Err = err
    return
}

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
    PROCESS_STATUS_CRASHED
)

type Ctrl struct {
    Id     string
    Cmd    *exec.Cmd
    Conf   config.ProcessConfig
    mutex  sync.Mutex
    Status constant.PROCESS_STATUS
}

func NewCtrl(conf config.ProcessConfig) (ctrl *Ctrl, err error) {
    ctrl = new(Ctrl)
    err = ctrl.Reload(conf)
    return
}

func (this *Ctrl) GetConfig() (conf config.ProcessConfig) {
    return this.Conf
}

func (this *Ctrl) GetStatus() (status constant.PROCESS_STATUS) {
    return this.Status
}

func (this *Ctrl) Reload(conf config.ProcessConfig) (err error) {
    this.Conf = conf
    return
}

func (this *Ctrl) Create(conf config.ProcessConfig) (err error) {
    if err = conf.Test(); err != nil {
        return
    }

    if err = this.Reload(conf); err != nil {
        return
    }

    if conf.AutoStart {
        err = this.Start()
    }
    return
}

func (this *Ctrl) Update(conf config.ProcessConfig) (err error) {
    if err = conf.Test(); err != nil {
        return
    }

    if this.Status == constant.PROCESS_STATUS_STARTED {
        if err = this.Stop(); err != nil {
            return
        }
    }

    if err = this.Reload(conf); err != nil {
        return
    }

    if err = this.Start(); err != nil {
        return
    }

    return
}

func (this *Ctrl) Start() (err error) {
    var conf = this.Conf
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
        this.Conf.StopSignal = uint(syscall.SIGKILL)
    }
    this.Cmd = cmd
    this.Status = constant.PROCESS_STATUS_STARTING
    err = this.Cmd.Start()
    if err != nil {
        this.Status = constant.PROCESS_STATUS_START_FAILED
        log.Printf("start faild:%s:%+v", err.Error(), *this)
        return err
    }
    go this.watch()
    this.Status = constant.PROCESS_STATUS_STARTED
    time.Sleep(time.Second)
    return nil
}

func (this *Ctrl) watch() {
    err := this.Cmd.Wait()
    fmt.Println(this.Conf.Id, "has exit.")
    if err != nil {
        fmt.Printf("[%s] got error: %s\n", this.Id, err.Error())
    }

    // conf is runing
    if this.Status == constant.PROCESS_STATUS_STARTED {
        if this.Conf.AutoRestart {
            fmt.Println(this.Conf.Id, "is auto restarting.")
            time.Sleep(time.Duration(this.Conf.AutoRestartDelay))
            this.Start()
        }
    }
}

func (this *Ctrl) Restart() (err error) {
    if err = this.Stop(); err == nil {
        err = this.Start()
    } else {
        log.Printf("stop err:", err.Error())
    }
    return err
}

func (this *Ctrl) Stop() (err error) {
    if this.Cmd.Process == nil {
        return errors.New(fmt.Sprintf("%s not started.", this.Id))
    }

    if this.Cmd.ProcessState == nil {
        log.Print("process state unvailable")
    } else if this.Cmd.ProcessState.Exited() {
        return errors.New(fmt.Sprintf("%s has stoped.", this.Id))
    }

    this.Status = PROCESS_STATUS_STOPING
    log.Printf("signal:%d", this.Conf.StopSignal)
    if err = this.Cmd.Process.Signal(syscall.Signal(this.Conf.StopSignal)); err == nil {
        this.Status = constant.PROCESS_STATUS_STOPED
    } else {
        log.Printf("stop faild:", err.Error())
    }
    return err
}

func (this *Ctrl) Tail() (err error) {
    return
}

func (this *Ctrl) Tailf() (err error) {
    return
}

func (this *Ctrl) Hello() {
    println("hello goatherd")
}

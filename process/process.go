package process

import (
    "errors"
    "fmt"
    "os"
    "os/exec"
    "strings"
    "sync"
    "syscall"
    "time"

    "sunteng/commons/constant"
    "sunteng/commons/log"
    "sunteng/commons/util"
)

type Sheep struct {
    Config
    Cmd    *exec.Cmd
    mutex  sync.Mutex
    Status string
}

func NewSheep() (ctrl *Sheep) {
    ctrl = new(Sheep)
    return
}

func (this *Sheep) GetStatus() (status string) {
    return this.Status
}

func (this *Sheep) GetConfig() (conf Config) {
    return this.Config
}

func (this *Sheep) SetConfig(conf Config) (err error) {
    this.Config = conf
    return
}

func (this *Sheep) Create(conf Config) (err error) {
    if err = conf.Check(); err != nil {
        return
    }

    if err = this.SetConfig(conf); err != nil {
        return
    }

    if conf.AutoStart {
        err = this.Start()
    }
    return
}

func (this *Sheep) Update(conf Config) (err error) {
    if err = conf.Check(); err != nil {
        return
    }

    if this.Status == constant.STATUS_STARTED {
        if err = this.Stop(); err != nil {
            return
        }
    }

    if err = this.SetConfig(conf); err != nil {
        return
    }

    if err = this.Start(); err != nil {
        return
    }

    return
}

func (this *Sheep) Start() (err error) {
    log.Logf("sheep start : %+v", this.Name)
    //cmd
    cmdArr := strings.Split(this.Command, " ")
    cmdId := cmdArr[0]
    cmdArgs := []string{}
    if len(cmdArr) > 1 {
        cmdArgs = cmdArr[1:]
    }
    var cmd = exec.Command(cmdId, cmdArgs...)
    //stdin
    if this.FileIn == "" {
        cmd.Stdin = os.Stdin
    } else if fd, err := os.Open(this.FileIn); err == nil {
        cmd.Stdin = fd
    } else {
        return errors.New("process config error : bad input file : " + err.Error())
    }
    //stdout
    if this.FileOut == "" {
        cmd.Stdout = os.Stdout
    } else if fd, err := os.OpenFile(this.FileOut, os.O_WRONLY|os.O_CREATE, 0666); err == nil {
        cmd.Stdout = fd
    } else {
        return errors.New("process config error : bad output file : " + err.Error())
    }
    //stderr
    if this.FileErr == "" {
        cmd.Stderr = os.Stderr
    } else if fd, err := os.OpenFile(this.FileErr, os.O_WRONLY|os.O_CREATE, 0666); err == nil {
        cmd.Stderr = fd
    } else {
        return errors.New("process config error : bad error file : " + err.Error())
    }
    //kill signal
    if this.StopSignal == uint(syscall.Signal(0)) {
        this.StopSignal = uint(syscall.SIGKILL)
    }
    this.Cmd = cmd
    this.Status = constant.STATUS_STARTING
    err = this.Cmd.Start()
    if err != nil {
        this.Status = constant.STATUS_START_FAILED
        log.Errorf("start faild:%s:%+v", err.Error(), *this)
        return err
    }
    log.Logf("sheep %s started", this.Name)
    go this.watch()
    this.Status = constant.STATUS_STARTED
    time.Sleep(time.Second)
    return nil
}

func (this *Sheep) watch() {
    err := this.Cmd.Wait()
    fmt.Println(this.Name, "has exit.")
    if err != nil {
        fmt.Printf("[%s] got error: %s\n", this.Name, err.Error())
    }

    // conf is runing
    if this.Status == constant.STATUS_STARTED {
        if this.AutoRestart {
            fmt.Println(this.Name, "is auto restarting.")
            time.Sleep(time.Duration(this.AutoRestartDelay))
            this.Start()
        }
    }
}

func (this *Sheep) Restart() (err error) {
    if err = this.Stop(); err == nil {
        err = this.Start()
    } else {
        log.Log("stop err:", err.Error())
    }
    return err
}

func (this *Sheep) Stop() (err error) {
    if this.Cmd.Process == nil {
        return errors.New(fmt.Sprintf("%s not started.", this.Name))
    }

    if this.Cmd.ProcessState == nil {
        log.Log("process state unvailable")
    } else if this.Cmd.ProcessState.Exited() {
        return errors.New(fmt.Sprintf("%s has stoped.", this.Name))
    }

    this.Status = constant.STATUS_STOPING
    log.Logf("signal:%d", this.StopSignal)
    if err = this.Cmd.Process.Signal(syscall.Signal(this.StopSignal)); err == nil {
        this.Status = constant.STATUS_STOPPED
    } else {
        log.Error("stop faild:", err.Error())
    }
    return err
}

func (this *Sheep) Tail(num int, forever bool) (lines [][]byte, err error) {
    return util.ReadLastLines(this.FileOut, num)
}

func (this *Sheep) Tailf() (err error) {
    return
}

func (this *Sheep) Hello() {
    println("hello goatherd")
}

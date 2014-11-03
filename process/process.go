package process

import (
    "errors"
    "fmt"
    "sync"
    "syscall"
    "time"

    "sunteng/commons/constant"
    "sunteng/commons/log"
    "sunteng/commons/util"
)

type Sheep struct {
    *Config
    mutex  sync.Mutex
    Status string
}

func NewSheep() *Sheep {
    return new(Sheep)
}

func (this *Sheep) GetStatus() string {
    return this.Status
}

func (this *Sheep) GetConfig() *Config {
    return this.Config
}

func (this *Sheep) SetConfig(conf *Config) (err error) {
    this.Config = conf
    return
}

func (this *Sheep) Create(conf *Config) (err error) {
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

func (this *Sheep) Update(conf *Config) (err error) {
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
    if err = this.Config.InitAll(); err != nil {
        return
    }

    this.Status = constant.STATUS_STARTING

    err = this.cmd.Start()
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
    err := this.cmd.Wait()
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
    if this.cmd.Process == nil {
        return errors.New(fmt.Sprintf("%s not started.", this.Name))
    }

    if this.cmd.ProcessState == nil {
        log.Log("process state unvailable")
    } else if this.cmd.ProcessState.Exited() {
        return errors.New(fmt.Sprintf("%s has stoped.", this.Name))
    }

    this.Status = constant.STATUS_STOPING
    log.Logf("signal:%d", this.StopSignal)
    if err = this.cmd.Process.Signal(syscall.Signal(this.StopSignal)); err == nil {
        this.Status = constant.STATUS_STOPPED
    } else {
        log.Error("stop faild:", err.Error())
    }
    return err
}

func (this *Sheep) Tail(num int, forever bool) (lines [][]byte, err error) {
    fileOut := this.GetDataFile(this.Name + ".log")
    return util.ReadLastLines(fileOut, num)
}

func (this *Sheep) Tailf() (err error) {
    return
}

func (this *Sheep) Hello() {
    println("hello goatherd")
}

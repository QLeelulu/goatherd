package process

import (
    "errors"
    "os"
    "os/exec"
    "strings"
    "sunteng/commons/confutil"
    "syscall"
)

type Config struct {
    confutil.DaemonBase
    Tags             []string
    Collie           string
    Command          string
    AutoStart        bool
    AutoRestart      bool
    AutoRestartDelay uint
    StartRetries     uint
    NumProcs         uint
    StopSignal       uint
    Environment      map[string]string
    cmd              *exec.Cmd
}

func (this *Config) Check() error {
    // if !util.DirExist(this.DataDir) {
    // return errors.New("config home not avalaible : " + this.Home)
    // }
    return nil
}

func (this *Config) InitAll() (err error) {
    // daemon base
    if err = this.DaemonBase.InitAll(); err != nil {
        return
    }

    //cmd
    cmdArr := strings.Split(this.Command, " ")
    cmdId := cmdArr[0]
    cmdArgs := []string{}
    if len(cmdArr) > 1 {
        cmdArgs = cmdArr[1:]
    }
    this.cmd = exec.Command(cmdId, cmdArgs...)

    //stdout
    fileOut := this.GetDataFile(this.Name + ".log")
    if this.cmd.Stdout, err = os.OpenFile(fileOut, os.O_WRONLY|os.O_CREATE, 0666); err != nil {
        return errors.New("process config error : bad output file : " + err.Error())
    }

    //stderr
    fileErr := this.GetDataFile(this.Name + ".err")
    if this.cmd.Stderr, err = os.OpenFile(fileErr, os.O_WRONLY|os.O_CREATE, 0666); err != nil {
        return errors.New("process config error : bad error file : " + err.Error())
    }

    //kill signal
    if this.StopSignal == uint(syscall.Signal(0)) {
        this.StopSignal = uint(syscall.SIGKILL)
    }
    return
}

func (this *Config) Expand(modelConfig Config) {
    if this.Name == "" {
        this.Name = modelConfig.Name
    }
    if this.DataDir == "" {
        this.DataDir = modelConfig.DataDir
    }
    if this.Collie == "" {
        this.Collie = modelConfig.Collie
    }
    if this.Command == "" {
        this.Command = modelConfig.Command
    }
    if this.AutoRestartDelay == 0 {
        this.AutoRestartDelay = modelConfig.AutoRestartDelay
    }
    if this.StartRetries == 0 {
        this.StartRetries = modelConfig.StartRetries
    }
    if this.NumProcs == 0 {
        this.NumProcs = modelConfig.NumProcs
    }
    if this.StopSignal == 0 {
        this.StopSignal = modelConfig.StopSignal
    }
    return
}

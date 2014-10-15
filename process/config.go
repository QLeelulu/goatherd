package process

import (
    "os"
    "sunteng/commons/util"
)

type Config struct {
    Name             string
    Command          string
    FileIn           string
    FileOut          string
    FileErr          string
    FilePId          string
    AutoStart        bool
    AutoRestart      bool
    AutoRestartDelay uint
    StartRetries     uint
    NumProcs         uint
    StopSignal       uint
    Environment      map[string]string
}

func (this *Config) Check() (err error) {
    if err = util.TryOpenFile(this.FileIn, os.O_RDONLY); err != nil {
        return
    }
    if err = util.TryOpenFile(this.FileOut, os.O_WRONLY|os.O_CREATE); err != nil {
        return
    }
    if err = util.TryOpenFile(this.FileErr, os.O_WRONLY|os.O_CREATE); err != nil {
        return
    }
    if err = util.TryOpenFile(this.FilePId, os.O_WRONLY|os.O_CREATE); err != nil {
        return
    }
    return
}

func (this *Config) Expand(modelConfig Config) {
    if this.Name == "" {
        this.Name = modelConfig.Name
    }
    if this.Command == "" {
        this.Command = modelConfig.Command
    }
    if this.FileOut == "" {
        this.FileOut = modelConfig.FileOut
    }
    if this.FileErr == "" {
        this.FileErr = modelConfig.FileErr
    }
    if this.FilePId == "" {
        this.FilePId = modelConfig.FilePId
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

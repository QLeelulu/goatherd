package collie

import (
    "goatherd/process"
    "sunteng/commons/confutil"
)

type ContexConfig struct {
    ConfigPath   string
    ProcessModel process.Config
    Process      map[string]*process.Config
}

type Config struct {
    Http confutil.NetBase
    confutil.DaemonBase
    ContexConfig
}

func (this *ContexConfig) Expand() {
    for name, processConf := range this.Process {
        processConf.Name = name
        processConf.Expand(this.ProcessModel)
    }
}

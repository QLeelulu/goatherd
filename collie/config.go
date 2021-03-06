package collie

import (
    "goatherd/process"
    "sunteng/commons/confutil"
)

type PeerConfigMap map[string]*PeerConfig
type PeerConfig struct {
    Name      string
    ElectAddr string
    HttpAddr  string
}

type ContexConfig struct {
    confutil.DaemonBase
    ConfigPath   string
    ProcessModel process.Config
    Process      map[string]*process.Config
}

type Config struct {
    *ContexConfig
    Elect confutil.NetBase
    Http  confutil.NetBase
}

func (this *ContexConfig) Expand() {
    this.ProcessModel.DataDir = this.GetDataDir()
    this.ProcessModel.Collie = this.Name
    for name, processConf := range this.Process {
        processConf.Name = name
        processConf.Expand(this.ProcessModel)
    }
}

func (this *Config) GetPeerConfig() *PeerConfig {
    return &PeerConfig{
        Name:      this.Name,
        ElectAddr: this.Elect.GetAddr(),
        HttpAddr:  this.Http.GetAddr(),
    }
}

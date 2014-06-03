package config

import (
    "errors"
    "log"
    "os"
    "path"

    "github.com/BurntSushi/toml"
    "goatherd/library/utils"
)

const (
    DEFAULT_MASTER_ID   = "master"
    DEFAULT_MASTER_HOST = "127.0.0.1"
    DEFAULT_MASTER_PORT = 8018
    DEFAULT_NETWORK     = "tcp"
    DEFAULT_SIGNAL      = 7
)

type Config struct {
    Master     []MasterConfig
    MasterFile []string
}

type ProcessConfig struct {
    Id               string
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

type GuardConfig struct {
    Id           string
    Host         string
    Port         int
    LogFile      string
    PidFile      string
    ProcessModel ProcessConfig
    Process      []ProcessConfig
    ProcessFile  []string
}

type MasterConfig struct {
    Id          string
    Host        string
    Port        int
    LogFile     string
    PidFile     string
    User        string
    AuthKeyFile string
    GuardModel  GuardConfig
    Guard       []GuardConfig
    GuardFile   []string
}

//guard.conf
func loadNewProcess(configFile string) (conf []ProcessConfig, err error) {
    log.Printf("load process config file:%+v", configFile)
    var newConf = new(GuardConfig)
    if _, err = toml.DecodeFile(configFile, newConf); err != nil {
        err = errors.New("process config decode faild:" + err.Error())
        return
    }
    conf = newConf.Process
    for i, _ := range conf {
        pconf := &conf[i]
        if err = pconf.Test(); err != nil {
            return
        }
    }
    return
}

func (this ProcessConfig) Test() (err error) {
    if err = utils.TryOpenFile(this.FileIn, os.O_RDONLY); err != nil {
        return
    }
    if err = utils.TryOpenFile(this.FileOut, os.O_WRONLY|os.O_CREATE); err != nil {
        return
    }
    if err = utils.TryOpenFile(this.FileErr, os.O_WRONLY|os.O_CREATE); err != nil {
        return
    }
    return
}

//process.conf
func (this *GuardConfig) checkInclude(configDir string) (err error) {
    for _, processFile := range this.ProcessFile {
        if !path.IsAbs(processFile) {
            processFile = path.Join(configDir, processFile)
        }
        if processConf, err := loadNewProcess(processFile); err != nil {
            return err
        } else {
            this.Process = append(this.Process, processConf...)
        }
    }
    return
}

func (this GuardConfig) Test() (err error) {
    if err = utils.TryOpenFile(this.LogFile, os.O_WRONLY|os.O_CREATE); err != nil {
        return
    }

    for _, pconf := range this.Process {
        if err = pconf.Test(); err != nil {
            return
        }
    }
    return
}

//guard.conf
func loadNewGuard(configFile string) (conf []GuardConfig, err error) {
    log.Printf("load guard config file:%+v", configFile)
    var newConf = new(MasterConfig)
    if _, err = toml.DecodeFile(configFile, newConf); err != nil {
        err = errors.New("guard config decode faild:" + err.Error())
        return
    }
    conf = newConf.Guard

    for i, _ := range conf {
        guardConf := &conf[i]
        if err = guardConf.checkInclude(path.Dir(configFile)); err != nil {
            return
        }
        err = guardConf.Test()
    }
    return
}

func (this MasterConfig) Test() (err error) {
    if err = utils.TryOpenFile(this.LogFile, os.O_WRONLY|os.O_CREATE); err != nil {
        return
    }

    if err = utils.TryOpenFile(this.AuthKeyFile, os.O_RDONLY); err != nil {
        return
    }

    for _, gconf := range this.Guard {
        if err = gconf.Test(); err != nil {
            return
        }
    }
    return
}

func (this *MasterConfig) checkInclude(configDir string) (err error) {
    for i, _ := range this.Guard {
        guard := &this.Guard[i]
        if err = guard.checkInclude(configDir); err != nil {
            return
        }
    }

    for _, guardFile := range this.GuardFile {
        if !path.IsAbs(guardFile) {
            guardFile = path.Join(configDir, guardFile)
        }
        if guardConf, err := loadNewGuard(guardFile); err != nil {
            return err
        } else {
            this.Guard = append(this.Guard, guardConf...)
        }
    }
    return
}

func loadNewMaster(configFile string) (conf []MasterConfig, err error) {
    log.Printf("load master config file:%+v", configFile)
    var newConf = new(Config)
    if _, err = toml.DecodeFile(configFile, newConf); err != nil {
        err = errors.New("master config decode faild:" + err.Error())
        return
    }
    conf = newConf.Master

    var configDir = path.Dir(configFile)
    for i, _ := range conf {
        masterConf := &conf[i]
        if err = masterConf.checkInclude(configDir); err != nil {
            return
        }
        if err = masterConf.Test(); err != nil {
            return
        }
        /* log.Printf("master conf:%+v", masterConf) */
    }

    return
}

func (this *Config) checkInclude(configDir string) (err error) {
    for i, _ := range this.Master {
        master := &this.Master[i]
        if err = master.checkInclude(configDir); err != nil {
            return
        }
    }

    for _, masterFile := range this.MasterFile {
        if !path.IsAbs(masterFile) {
            masterFile = path.Join(configDir, masterFile)
        }
        if masterConf, err := loadNewMaster(masterFile); err != nil {
            return err
        } else {
            this.Master = append(this.Master, masterConf...)
        }
    }
    return
}

func (this Config) Test() (err error) {
    if len(this.Master) < 1 {
        err = errors.New("master config missed")
    }
    return
}

func LoadNewConfig(configFile string) (conf *Config, err error) {
    if !path.IsAbs(configFile) {
        pwd, _ := os.Getwd()
        configFile = path.Join(pwd, configFile)
    }
    log.Printf("load goatherd config file:%+v", configFile)

    conf = new(Config)
    if _, err = toml.DecodeFile(configFile, conf); err != nil {
        err = errors.New("config decode faild:" + err.Error())
        return
    }

    var configDir = path.Dir(configFile)
    if err = conf.checkInclude(configDir); err != nil {
        return
    }

    if err = conf.Test(); err != nil {
        return
    }
    return
}

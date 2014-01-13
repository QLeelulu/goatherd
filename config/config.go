package config

import (
    "errors"
    "log"
    "os"
    "path"

    "github.com/BurntSushi/toml"
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
    AutoStart        bool
    AutoRestart      bool
    AutoRestartDelay uint
    StartRetries     uint
    NumProcs         uint
    StopSignal       uint
    Environment      map[string]string
}

type SlaveConfig struct {
    Id             string
    Host           string
    Port           int
    LogFile        string
    BinPath        string
    DefaultProcess ProcessConfig
    ProcessModel   ProcessConfig
    Process        []ProcessConfig
    ProcessFile    []string
}

type MasterConfig struct {
    Id          string
    Host        string
    Port        int
    LogFile     string
    AuthKeyFile string
    SlaveModel  SlaveConfig
    Slave       []SlaveConfig
    SlaveFile   []string
}

//slave.conf
func loadNewProcess(configFile string) (conf []ProcessConfig, err error) {
    log.Printf("load process config file:%+v", configFile)
    var newConf = new(SlaveConfig)
    if _, err = toml.DecodeFile(configFile, newConf); err != nil {
        err = errors.New("process config decode faild:" + err.Error())
        return
    }
    conf = newConf.Process
    return
}

//process.conf
func (this *SlaveConfig) checkInclude(configDir string) (err error) {
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

//slave.conf
func loadNewSlave(configFile string) (conf []SlaveConfig, err error) {
    log.Printf("load slave config file:%+v", configFile)
    var newConf = new(MasterConfig)
    if _, err = toml.DecodeFile(configFile, newConf); err != nil {
        err = errors.New("slave config decode faild:" + err.Error())
        return
    }
    conf = newConf.Slave

    for i, _ := range conf {
        slaveConf := &conf[i]
        if err = slaveConf.checkInclude(path.Dir(configFile)); err != nil {
            return
        }
    }
    return
}

func (this *MasterConfig) checkInclude(configDir string) (err error) {
    for i, _ := range this.Slave {
        slave := &this.Slave[i]
        if err = slave.checkInclude(configDir); err != nil {
            return
        }
    }

    for _, slaveFile := range this.SlaveFile {
        if !path.IsAbs(slaveFile) {
            slaveFile = path.Join(configDir, slaveFile)
        }
        if slaveConf, err := loadNewSlave(slaveFile); err != nil {
            return err
        } else {
            this.Slave = append(this.Slave, slaveConf...)
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
    err = conf.checkInclude(configDir)

    return
}

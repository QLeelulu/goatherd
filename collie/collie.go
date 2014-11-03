package collie

import (
    "bytes"
    "errors"

    "goatherd/process"
    "sunteng/commons/constant"
    "sunteng/commons/log"
)

type Collied struct {
    conf   ContexConfig
    sheeps map[string]*process.Sheep
}

func NewCollied() *Collied {
    return &Collied{
        sheeps: make(map[string]*process.Sheep),
    }
}

func (this *Collied) LoadConfig(conf *ContexConfig) (err error) {
    conf.Expand()
    for name, processConf := range conf.Process {
        sheep := process.NewSheep()
        if err = sheep.Create(processConf); err != nil {
            log.Errorf("load sheep : %+v --- %s", name, err.Error())
            return
        }
        // log.Logf("load sheep : %+v", processConf)
        this.sheeps[name] = sheep
    }
    this.conf.ProcessModel = conf.ProcessModel
    this.conf.ConfigPath = conf.ConfigPath
    return
}

func (this *Collied) GetCollieConfig() interface{} {
    return this.conf
}

func (this *Collied) GetSheepConfig(name string) (interface{}, error) {
    sheep, ok := this.sheeps[name]
    if !ok {
        return nil, errors.New("get sheep config faild : " + name)
    }
    return sheep.Config, nil
}

func (this *Collied) SheepGetStatus(name string) (string, error) {
    sheep, ok := this.sheeps[name]
    if !ok {
        return constant.STATUS_UNKNOWN, errors.New("get sheep config faild : " + name)
    }
    return sheep.Status, nil
}

// to do : forever
func (this *Collied) SheepTail(name string, num int, forever bool) (interface{}, error) {
    sheep, ok := this.sheeps[name]
    if !ok {
        return nil, errors.New(name + " not found")
    }
    lines, err := sheep.Tail(num, forever)
    if err != nil {
        return nil, err
    }
    return bytes.Join(lines, []byte("\n")), nil
}

func (this *Collied) SheepReload(conf *process.Config) error {
    conf.Expand(this.conf.ProcessModel)
    if err := this.SheepDel(conf.Name); err != nil {
        return nil
    }
    return this.SheepAdd(conf)
}

func (this *Collied) SheepRestart(name string) error {
    sheep, ok := this.sheeps[name]
    if !ok {
        return errors.New(name + " not found")
    }
    return sheep.Restart()
}

func (this *Collied) SheepStart(name string) error {
    sheep, ok := this.sheeps[name]
    if !ok {
        return errors.New(name + " not found")
    }
    return sheep.Start()
}

func (this *Collied) SheepStop(name string) error {
    sheep, ok := this.sheeps[name]
    if !ok {
        return errors.New(name + " not found")
    }
    return sheep.Stop()
}

func (this *Collied) SheepAdd(conf *process.Config) error {
    conf.Expand(this.conf.ProcessModel)

    var sheep = process.NewSheep()
    var err = sheep.Create(conf)
    if err != nil {
        return err
    }

    this.sheeps[sheep.Name] = sheep
    return nil
}

func (this *Collied) SheepDel(name string) error {
    sheep, ok := this.sheeps[name]
    if !ok {
        return errors.New(name + " not found")
    }
    if sheep.Status != constant.STATUS_STOPPED {
        if err := sheep.Stop(); err != nil {
            return err
        }
    }
    delete(this.sheeps, sheep.Name)
    return nil
}

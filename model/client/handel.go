package client

import (
    "goatherd/config"
    "goatherd/constant"
    "goatherd/model/master"
)

var defaultHandel *Handel

type Handel struct {
    masterId string
    sclient  *master.Client
}

/* func HandelApp() *Handel { */
/* if defaultHandel == nil { */
/* conf := config.ConfigApp().Master */
/* defaultHandel = NewHandel(conf.Id, master.Host, master.Port) */
/* } */
/* return defaultHandel */
/* } */

func NewHandel() *Handel {
    return NewHandelWithConfig("", "", 0)
}

func NewHandelWithConfig(masterId, host string, port int) *Handel {
    if port == 0 {
        port = config.DEFAULT_MASTER_PORT
    }
    if host == "" {
        host = config.DEFAULT_MASTER_HOST
    }
    if masterId == "" {
        masterId = config.DEFAULT_MASTER_ID
    }
    return &Handel{
        masterId: masterId,
        sclient: &master.Client{
            Host: host,
            Port: port,
        },
    }
}

func (this *Handel) GetConfig() (conf config.MasterConfig, err error) {
    var arg = master.Arg{
        Id:     this.masterId,
        Action: constant.ACT_GET,
    }
    var ret = master.NewRet()
    if err = this.sclient.Call(arg, ret); err != nil {
        return
    }
    conf = *ret.Conf
    return
}

package collie

import (
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "io/ioutil"
    "net/http"
    "path"
    "strconv"

    "goatherd/process"
    "sunteng/commons/confutil"
    "sunteng/commons/log"
    "sunteng/commons/util"
    "sunteng/commons/util/toml_util"
    "sunteng/commons/web"
)

type httpServer struct {
    ctx   *Contex
    elect *electServer
    confutil.NetBase
    confutil.DaemonBase
}

func NewHttpServe(conf Config, leader string) (err error) {
    var server = new(httpServer)

    // 初始化daemon配置
    if err = conf.DaemonBase.InitAll(); err != nil {
        log.Noticef("new http serve daemon base init faild: %+v ---  %s\n ", conf, err.Error())
        return
    }
    server.DaemonBase = conf.DaemonBase

    // 初始化collie配置
    server.ctx = NewContex()
    if err = server.ctx.LoadConfig(conf.ContexConfig); err != nil {
        log.Errorf("new http serve contex init faild: %+v\n", conf)
        return
    }

    // 初始化http配置
    if err = conf.Http.Check(); err != nil {
        log.Noticef("new http serve net base init faild: %+v ---  %s\n ", conf, err.Error())
        return
    }
    server.NetBase = conf.Http

    if server.elect, err = NewElectServer(conf, leader); err != nil {
        log.Error("new elect server faild : ", err.Error())
        return
    }

    // 持久化配置
    if err = server.Persistence(); err != nil {
        log.Errorf("persistence faild : %s", err.Error())
        return
    }

    // 启动http服务
    err = conf.Http.StartHttp(server)
    return
}

func (this *httpServer) Persistence() (err error) {
    conf, err := this.doCollieGetConfig()
    if err != nil {
        return
    }

    buf, err := toml_util.Encode(conf)
    if err != nil {
        return
    }
    err = ioutil.WriteFile(this.ctx.conf.ConfigPath, []byte(buf), 0666)
    return
}

func (this *httpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    peers := this.elect.doPeers()

    // get peers
    var ret web.ApiResponse
    var collies = make(PeerConfigMap)
    if len(r.URL.Query()["collie"]) == 0 {
        collies = peers
    } else {
        for _, collie := range r.URL.Query()["collie"] {
            if peer, ok := peers[collie]; ok {
                collies[collie] = peer
            } else {
                ret.Set(collie, web.ApiResponse{500, "collie unvalaible : " + collie, nil})
            }
        }
    }

    for name, peer := range collies {
        var resp web.ApiResponse
        if name == this.Name {
            resp = this.router(r)
        } else {
            query := r.URL.Query()
            query["collie"] = []string{name}
            resp = this.proxy(fmt.Sprintf("http://%s%s?%s", peer.HttpAddr, r.URL.Path, query.Encode()), r.Body)
        }

        ret.Set(name, resp)
    }
    ret.Write(w)
}

func (this *httpServer) proxy(addr string, body io.Reader) (res web.ApiResponse) {
    var err error
    defer func() {
        if err != nil {
            res = web.ApiResponse{StatusCode: 500, StatusTxt: err.Error()}
        }
    }()
    resp, err := http.Post(addr, "application/toml", body)
    if err != nil {
        return
    }
    defer resp.Body.Close()

    data, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return
    }

    if err = json.Unmarshal(data, &res); err != nil {
        return
    }
    return
}

func (this *httpServer) router(r *http.Request) (resp web.ApiResponse) {
    var data interface{}
    var code int
    var err error
    switch path.Dir(r.URL.Path) {
    case "/sheep":
        data, code, err = this.sheepRouter(r)
    case "/collie":
        data, code, err = this.collieRouter(r)
    default:
        code, err = 400, errors.New("bad path : "+r.URL.Path)
    }
    if err != nil {
        resp = web.ApiResponse{
            StatusCode: 500,
            StatusTxt:  err.Error(),
            Data:       data,
        }
    } else {
        resp = web.ApiResponse{
            StatusCode: code,
            StatusTxt:  "ok",
            Data:       data,
        }
    }
    return
}

func (this *httpServer) sheepRouter(r *http.Request) (data interface{}, code int, err error) {
    code = 200
    switch path.Base(r.URL.Path) {
    case "add":
        data, err = this.doSheepAdd(r)
    case "del":
        data, err = this.doSheepDel(r)
    case "start":
        data, err = this.doSheepStart(r)
    case "stop":
        data, err = this.doSheepStop(r)
    case "restart":
        data, err = this.doSheepRestart(r)
    case "reload":
        data, err = this.doSheepReload(r)
    case "tail":
        data, err = this.doSheepTail(r)
    case "list":
    case "status":
        data, err = this.doSheepGetStatus(r)
    case "get_config":
        data, err = this.doSheepGetConfig(r)
    default:
        code, err = 400, errors.New("bad path : "+r.URL.Path)
    }
    return
}

func (this *httpServer) collieRouter(r *http.Request) (data interface{}, code int, err error) {
    switch path.Base(r.URL.Path) {
    case "get_config":
        data, err = this.doCollieGetConfig()
    case "set_config":
        data, err = this.doCollieGetConfig()
    case "list":
    default:
        code, err = 400, errors.New("bad path : "+r.URL.Path)
    }
    return
}

func (this *httpServer) readBody(r *web.ReqParams) (ctx ContexConfig, err error) {
    if len(r.Body) == 0 {
        return
    }
    if err = toml_util.Decode([]byte(r.Body), &ctx); err != nil {
        return
    }
    if ctx.ProcessModel.Name == "" {
        ctx.ProcessModel = this.ctx.conf.ProcessModel
    }
    ctx.Expand()
    for name, proc := range ctx.Process {
        if proc.Collie != "" && proc.Collie != this.Name {
            delete(ctx.Process, name)
        }
    }
    return
}

func (this *httpServer) doCollieGetConfig() (Config, error) {
    var conf = Config{
        Http:       this.NetBase,
        DaemonBase: this.DaemonBase,
        ContexConfig: ContexConfig{
            ProcessModel: this.ctx.conf.ProcessModel,
            Process:      make(map[string]*process.Config),
        },
        Elect: this.elect.NetBase,
    }
    for name, sheep := range this.ctx.sheeps {
        conf.Process[name] = &sheep.Config
    }
    return conf, nil
}

func (this *httpServer) doSheepGetConfig(r *http.Request) (util.WaitRetMap, error) {
    reqParams, err := web.NewReqParams(r)
    if err != nil {
        return nil, web.HTTPError{400, "INVALID_REQUEST"}
    }

    names, err := reqParams.GetAll("name")
    if err != nil {
        for name, _ := range this.ctx.sheeps {
            names = append(names, name)
        }
    }

    rets, err := util.MultiWait(names, func(name interface{}) util.WaitRet {
        data, err := this.ctx.GetSheepConfig(name.(string))
        return util.WaitRet{err, data}
    })
    return rets, err
}

func (this *httpServer) doSheepGetStatus(r *http.Request) (util.WaitRetMap, error) {
    reqParams, err := web.NewReqParams(r)
    if err != nil {
        return nil, web.HTTPError{400, "INVALID_REQUEST"}
    }

    names, err := reqParams.GetAll("name")
    if err != nil {
        for name, _ := range this.ctx.sheeps {
            names = append(names, name)
        }
    }

    rets, err := util.MultiWait(names, func(name interface{}) util.WaitRet {
        data, err := this.ctx.SheepGetStatus(name.(string))
        return util.WaitRet{err, data}
    })
    return rets, err
}

func (this *httpServer) doSheepTail(r *http.Request) (util.WaitRetMap, error) {
    reqParams, err := web.NewReqParams(r)
    if err != nil {
        return nil, web.HTTPError{400, "INVALID_REQUEST"}
    }

    names, err := reqParams.GetAll("name")
    if err != nil {
        for name, _ := range this.ctx.sheeps {
            names = append(names, name)
        }
    }

    var num int
    numStr, err := reqParams.Get("num")
    if err != nil {
        num = 10
    } else if num, err = strconv.Atoi(numStr); err != nil {
        return nil, web.HTTPError{400, "MISSING_ARG_NUM"}
    }

    forever, _ := reqParams.Get("forever")

    rets, err := util.MultiWait(names, func(name interface{}) util.WaitRet {
        data, err := this.ctx.SheepTail(name.(string), num, forever == "true")
        return util.WaitRet{err, data}
    })
    return rets, err
}

func (this *httpServer) doSheepReload(r *http.Request) (util.WaitRetMap, error) {
    reqParams, err := web.NewReqParams(r)
    if err != nil {
        return nil, web.HTTPError{400, "INVALID_REQUEST"}
    }

    ctx, err := this.readBody(reqParams)
    if err != nil {
        return nil, web.HTTPError{400, "INVALID_BODY"}
    }

    rets, err := util.MultiWait(ctx.Process, func(conf interface{}) util.WaitRet {
        return util.WaitRet{nil, this.ctx.SheepReload(*conf.(*process.Config))}
    })

    this.Persistence()
    return rets, err
}

func (this *httpServer) doSheepRestart(r *http.Request) (util.WaitRetMap, error) {
    reqParams, err := web.NewReqParams(r)
    if err != nil {
        return nil, web.HTTPError{400, "INVALID_REQUEST"}
    }

    names, err := reqParams.GetAll("name")
    if err != nil {
        for name, _ := range this.ctx.sheeps {
            names = append(names, name)
        }
    }

    rets, err := util.MultiWait(names, func(name interface{}) util.WaitRet {
        return util.WaitRet{nil, this.ctx.SheepRestart(name.(string))}
    })

    return rets, err
}

func (this *httpServer) doSheepStart(r *http.Request) (util.WaitRetMap, error) {
    reqParams, err := web.NewReqParams(r)
    if err != nil {
        return nil, web.HTTPError{400, "INVALID_REQUEST"}
    }

    names, err := reqParams.GetAll("name")
    if err != nil {
        for name, _ := range this.ctx.sheeps {
            names = append(names, name)
        }
    }

    rets, err := util.MultiWait(names, func(name interface{}) util.WaitRet {
        return util.WaitRet{nil, this.ctx.SheepStart(name.(string))}
    })

    return rets, err
}

func (this *httpServer) doSheepStop(r *http.Request) (util.WaitRetMap, error) {
    reqParams, err := web.NewReqParams(r)
    if err != nil {
        return nil, web.HTTPError{400, "INVALID_REQUEST"}
    }

    names, err := reqParams.GetAll("name")
    if err != nil {
        for name, _ := range this.ctx.sheeps {
            names = append(names, name)
        }
    }

    rets, err := util.MultiWait(names, func(name interface{}) util.WaitRet {
        return util.WaitRet{nil, this.ctx.SheepStop(name.(string))}
    })

    return rets, err
}

func (this *httpServer) doSheepAdd(r *http.Request) (util.WaitRetMap, error) {
    reqParams, err := web.NewReqParams(r)
    if err != nil {
        return nil, web.HTTPError{400, "INVALID_REQUEST"}
    }

    ctx, err := this.readBody(reqParams)
    if err != nil {
        return nil, web.HTTPError{400, "INVALID_BODY"}
    }

    rets, err := util.MultiWait(ctx.Process, func(conf interface{}) util.WaitRet {
        return util.WaitRet{nil, this.ctx.SheepAdd(*conf.(*process.Config))}
    })

    this.Persistence()
    return rets, err
}

func (this *httpServer) doSheepDel(r *http.Request) (util.WaitRetMap, error) {
    reqParams, err := web.NewReqParams(r)
    if err != nil {
        return nil, web.HTTPError{400, "INVALID_REQUEST"}
    }

    names, err := reqParams.GetAll("name")
    if err != nil {
        for name, _ := range this.ctx.sheeps {
            names = append(names, name)
        }
    }

    rets, err := util.MultiWait(names, func(name interface{}) util.WaitRet {
        return util.WaitRet{nil, this.ctx.SheepDel(name.(string))}
    })

    this.Persistence()
    return rets, err
}

package collie

import (
    "errors"
    "io/ioutil"
    "net/http"
    "strconv"
    "strings"

    "goatherd/process"
    "sunteng/commons/confutil"
    "sunteng/commons/log"
    "sunteng/commons/util"
    "sunteng/commons/util/toml_util"
    "sunteng/commons/web"
)

type httpServer struct {
    ctx *Contex
    confutil.NetBase
    confutil.DaemonBase
}

func NewHttpServe(conf Config) (err error) {
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
    server.NetBase = conf.Http

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
    // log.Logf("persistence : %s", buf)
    return
}

func (this *httpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if this.sheepRouter(w, r) == nil {
        return
    }
    if this.collieRouter(w, r) == nil {
        return
    }
    web.ApiResponse(w, 404, "NOT_FOUND", nil)
}

func (this *httpServer) sheepRouter(w http.ResponseWriter, r *http.Request) (err error) {
    var target = "/sheep/"
    if !strings.HasPrefix(r.URL.Path, target) {
        err = errors.New("bad prefix")
        return
    }
    var action = strings.TrimPrefix(r.URL.Path, target)
    switch action {
    case "add":
        web.APIResponse(w, r, func() (data interface{}, err error) { return this.doSheepAdd(r) })
    case "del":
        web.APIResponse(w, r, func() (data interface{}, err error) { return this.doSheepDel(r) })
    case "start":
        web.APIResponse(w, r, func() (data interface{}, err error) { return this.doSheepStart(r) })
    case "stop":
        web.APIResponse(w, r, func() (data interface{}, err error) { return this.doSheepStop(r) })
    case "restart":
        web.APIResponse(w, r, func() (data interface{}, err error) { return this.doSheepRestart(r) })
    case "reload":
        web.APIResponse(w, r, func() (data interface{}, err error) { return this.doSheepReload(r) })
    case "tail":
        web.APIResponse(w, r, func() (data interface{}, err error) { return this.doSheepTail(r) })
    case "list":
    case "status":
        web.APIResponse(w, r, func() (data interface{}, err error) { return this.doSheepGetStatus(r) })
    case "get_config":
        web.APIResponse(w, r, func() (data interface{}, err error) { return this.doSheepGetConfig(r) })
    default:
        err = errors.New("bad path : " + action)
    }
    return
}

func (this *httpServer) collieRouter(w http.ResponseWriter, r *http.Request) (err error) {
    var target = "/collie/"
    if !strings.HasPrefix(r.URL.Path, target) {
        err = errors.New("bad prefix")
        return
    }
    var action = strings.TrimPrefix(r.URL.Path, target)
    switch action {
    case "get_config":
        web.APIResponse(w, r, func() (data interface{}, err error) {
            return this.doCollieGetConfig()
        })
    default:
        err = errors.New("bad path : " + action)
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
    }
    for name, sheep := range this.ctx.sheeps {
        conf.Process[name] = &sheep.Config
    }
    return conf, nil
}

func (this *httpServer) doSheepGetConfig(r *http.Request) (interface{}, error) {
    reqParams, err := web.NewReqParams(r)
    if err != nil {
        return nil, web.HTTPError{400, "INVALID_REQUEST"}
    }

    name, err := reqParams.Get("name")
    if err != nil {
        return nil, web.HTTPError{400, "MISSING_ARG_NAME"}
    }

    return this.ctx.GetSheepConfig(name)
}

func (this *httpServer) doSheepGetStatus(r *http.Request) (interface{}, error) {
    reqParams, err := web.NewReqParams(r)
    if err != nil {
        return nil, web.HTTPError{400, "INVALID_REQUEST"}
    }

    var names []string
    all, err := reqParams.Get("all")
    if err == nil && all == "true" {
        for name, _ := range this.ctx.sheeps {
            names = append(names, name)
        }
    } else {
        names, err = reqParams.GetAll("name")
        if err != nil {
            return nil, web.HTTPError{400, "MISSING_ARG_NAME"}
        }
    }

    var res = make(map[string]string)
    err = util.MultiWaitMap("sheep status", names, func(name interface{}) error {
        status, _err := this.ctx.SheepGetStatus(name.(string))
        res[name.(string)] = status
        return _err
    })
    return res, err
}

func (this *httpServer) doSheepTail(r *http.Request) (interface{}, error) {
    reqParams, err := web.NewReqParams(r)
    if err != nil {
        return nil, web.HTTPError{400, "INVALID_REQUEST"}
    }

    name, err := reqParams.Get("name")
    if err != nil {
        return nil, web.HTTPError{400, "MISSING_ARG_NAME"}
    }

    var num int
    numStr, err := reqParams.Get("num")
    if err != nil {
        num = 10
    } else if num, err = strconv.Atoi(numStr); err != nil {
        return nil, web.HTTPError{400, "MISSING_ARG_NUM"}
    }

    forever, _ := reqParams.Get("forever")
    return this.ctx.SheepTail(name, num, forever == "true")
}

func (this *httpServer) doSheepReload(r *http.Request) (interface{}, error) {
    reqParams, err := web.NewReqParams(r)
    if err != nil {
        return nil, web.HTTPError{400, "INVALID_REQUEST"}
    }

    if len(reqParams.Body) == 0 {
        return nil, web.HTTPError{400, "MSG_EMPTY"}
    }
    var ctx ContexConfig
    if err = toml_util.Decode([]byte(reqParams.Body), &ctx); err != nil {
        return nil, err
    }
    if ctx.ProcessModel.Name == "" {
        ctx.ProcessModel = this.ctx.conf.ProcessModel
    }
    ctx.Expand()

    err = util.MultiWaitMap("sheep reload", ctx.Process, func(conf interface{}) error {
        return this.ctx.SheepReload(*conf.(*process.Config))
    })

    err = this.Persistence()
    return "ok", err
}

func (this *httpServer) doSheepRestart(r *http.Request) (interface{}, error) {
    reqParams, err := web.NewReqParams(r)
    if err != nil {
        return nil, web.HTTPError{400, "INVALID_REQUEST"}
    }

    names, err := reqParams.GetAll("name")
    if err != nil {
        return nil, web.HTTPError{400, "MISSING_ARG_NAME"}
    }

    err = util.MultiWait("sheep restart", names, func(name string) error {
        return this.ctx.SheepRestart(name)
    })

    return "ok", err
}

func (this *httpServer) doSheepStart(r *http.Request) (interface{}, error) {
    reqParams, err := web.NewReqParams(r)
    if err != nil {
        return nil, web.HTTPError{400, "INVALID_REQUEST"}
    }

    names, err := reqParams.GetAll("name")
    if err != nil {
        return nil, web.HTTPError{400, "MISSING_ARG_NAME"}
    }

    err = util.MultiWait("sheep start", names, func(name string) error {
        return this.ctx.SheepStart(name)
    })

    return "ok", err
}

func (this *httpServer) doSheepStop(r *http.Request) (interface{}, error) {
    reqParams, err := web.NewReqParams(r)
    if err != nil {
        return nil, web.HTTPError{400, "INVALID_REQUEST"}
    }

    names, err := reqParams.GetAll("name")
    if err != nil {
        return nil, web.HTTPError{400, "MISSING_ARG_NAME"}
    }

    err = util.MultiWait("sheep stop", names, func(name string) error {
        return this.ctx.SheepStop(name)
    })

    return "ok", err
}

func (this *httpServer) doSheepAdd(r *http.Request) (interface{}, error) {
    reqParams, err := web.NewReqParams(r)
    if err != nil {
        return nil, web.HTTPError{400, "INVALID_REQUEST"}
    }

    if len(reqParams.Body) == 0 {
        return nil, web.HTTPError{400, "MSG_EMPTY"}
    }
    var ctx ContexConfig
    if err = toml_util.Decode([]byte(reqParams.Body), &ctx); err != nil {
        return nil, err
    }
    if ctx.ProcessModel.Name == "" {
        ctx.ProcessModel = this.ctx.conf.ProcessModel
    }
    ctx.Expand()

    err = util.MultiWaitMap("sheep add", ctx.Process, func(conf interface{}) error {
        return this.ctx.SheepAdd(*conf.(*process.Config))
    })

    err = this.Persistence()
    return "ok", err
}

func (this *httpServer) doSheepDel(r *http.Request) (interface{}, error) {
    reqParams, err := web.NewReqParams(r)
    if err != nil {
        return nil, web.HTTPError{400, "INVALID_REQUEST"}
    }

    names, err := reqParams.GetAll("name")
    if err != nil {
        return nil, web.HTTPError{400, "MISSING_ARG_NAME"}
    }

    err = util.MultiWait("sheep del", names, func(name string) error {
        return this.ctx.SheepDel(name)
    })
    err = this.Persistence()
    return "ok", err
}

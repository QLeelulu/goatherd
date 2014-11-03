package collie

import (
    "bytes"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "os"
    "sunteng/commons/confutil"
    "sunteng/commons/log"
    "sunteng/commons/web"
    "time"

    "github.com/goraft/raft"
)

type electServer struct {
    raftServer raft.Server
    mux        *http.ServeMux
    confutil.DaemonBase
    confutil.NetBase
    peerConfig *PeerConfig
    peers      PeerConfigMap
}

func NewElectServer(conf Config, leader string) (server *electServer, err error) {
    server = &electServer{
        DaemonBase: conf.DaemonBase,
        NetBase:    conf.Elect,
        mux:        http.NewServeMux(),
        peerConfig: conf.GetPeerConfig(),
    }

    if err = conf.Elect.Check(); err != nil {
        return
    }

    if err = server.InitRaftServer(); err != nil {
        return
    }

    if err = server.InitRaftCluster(leader); err != nil {
        return
    }

    go func() {
        if err = server.ServeHttp(); err != nil {
            os.Exit(1)
        }
    }()

    go func() {
        time.Sleep(time.Second)
        if err = server.syncLoop(); err != nil {
            os.Exit(2)
        }
    }()
    return
}

func (this *electServer) syncLoop() error {
    this.syncCluster()
    timeup := time.After(time.Minute)
    tick := time.Tick(10 * time.Second)
    var syncd bool
    for {
        select {
        case <-tick:
            if err := this.syncCluster(); err == nil {
                syncd = true
            }
        case <-timeup:
            if !syncd {
                return errors.New("sync timeup")
            }
        }
    }
    return nil
}

func (this *electServer) ServeHttp() (err error) {
    this.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { this.Router(w, r) })
    return this.StartHttp(this.mux)
}

func (this *electServer) InitRaftServer() (err error) {
    // raft.SetLogLevel(raft.Trace)
    transporter := raft.NewHTTPTransporter("/raft", 200*time.Millisecond)
    this.raftServer, err = raft.NewServer(this.Name, this.GetDataDir(), transporter, nil, nil, "")
    if err != nil {
        return
    }
    transporter.Install(this.raftServer, this.mux)
    err = this.raftServer.Start()
    return
}

func (this *electServer) InitRaftCluster(leader string) (err error) {
    if leader != "" {
        if !this.raftServer.IsLogEmpty() {
            log.Log("cannot join with exist log")
            return
        }
        if err = this.JoinLeader(leader); err != nil {
            return
        }
    } else if this.raftServer.IsLogEmpty() {
        if err = this.JoinSelf(); err != nil {
            return
        }
    } else {
        log.Notice("recovered from log")
    }
    return
}

func (this *electServer) JoinSelf() (err error) {
    log.Logf("attemping to join self")
    _, err = this.raftServer.Do(&raft.DefaultJoinCommand{
        Name:             this.Name,
        ConnectionString: this.HttpAddr(),
    })
    return
}

func (this *electServer) JoinLeader(leader string) (err error) {
    log.Logf("attemping to join leader : %s", leader)
    command := raft.DefaultJoinCommand{
        Name:             this.Name,
        ConnectionString: this.HttpAddr(),
    }
    var b bytes.Buffer
    if err = json.NewEncoder(&b).Encode(command); err != nil {
        return
    }
    resp, err := http.Post("http://"+leader+"/join", "application/json", &b)
    if err != nil {
        return
    }
    defer resp.Body.Close()
    if resp.StatusCode != 200 {
        return fmt.Errorf("bad status code : %+v", resp.StatusCode)
    }
    return
}

func (this *electServer) Router(w http.ResponseWriter, r *http.Request) {
    var err error
    switch r.URL.Path {
    case "/join":
        err = this.doJoin(r)
    case "/peers":
        err = json.NewEncoder(w).Encode(this.doPeers())
    case "/leader":
        err = json.NewEncoder(w).Encode(this.doLeader())
    case "/peer_config":
        err = json.NewEncoder(w).Encode(this.doPeerConfig())
    case "/peer_exchange":
        err = this.doExchangePeerConfig(r, w)
    case "/stat":
        err = this.doStat(w)
    default:
        resp := web.ApiResponse{404, "NOT_FOUND", nil}
        resp.Write(w)
    }
    if err != nil {
        resp := web.ApiResponse{502, err.Error(), nil}
        resp.Write(w)
        log.Errorf("%s handel faild : %s", r.URL.Path, err.Error())
    }
}

func (this *electServer) doStat(w http.ResponseWriter) error {
    return json.NewEncoder(w).Encode(map[string]interface{}{
        "leader":  this.raftServer.Leader(),
        "running": this.raftServer.Running(),
        "peers":   this.raftServer.Peers(),
        "stat":    this.raftServer.GetState(),
        "nodes":   this.peers,
    })
}

func (this *electServer) syncCluster() (err error) {
    b, err := json.Marshal(this.peerConfig)
    if err != nil {
        return
    }

    var peers = make(PeerConfigMap)
    for name, peer := range this.raftServer.Peers() {
        resp, err := http.Post(peer.ConnectionString+"/peer_exchange", "application/json", bytes.NewBuffer(b))
        if err != nil {
            log.Errorf("post faild : %s", err.Error())
            continue
        }
        defer resp.Body.Close()
        if resp.StatusCode != 200 {
            log.Errorf("%s : resp code not 200", peer.ConnectionString)
            continue
        }
        var peerConfig = new(PeerConfig)
        if err = json.NewDecoder(resp.Body).Decode(peerConfig); err != nil {
            log.Errorf("decode response from %s faild : %s", peer.ConnectionString, err.Error())
            continue
        }
        peers[name] = peerConfig
    }
    peers[this.Name] = this.peerConfig

    this.peers = peers
    return
}

func (this *electServer) doExchangePeerConfig(r *http.Request, w http.ResponseWriter) error {
    var peerConfig = new(PeerConfig)
    if err := json.NewDecoder(r.Body).Decode(peerConfig); err != nil {
        return err
    } else {
        this.peers[peerConfig.Name] = peerConfig
        json.NewEncoder(w).Encode(this.peerConfig)
    }
    return nil
}

func (this *electServer) doPeerConfig() *PeerConfig {
    return this.peerConfig
}

func (this *electServer) doPeers() PeerConfigMap {
    return this.peers
}

func (this *electServer) doLeader() string {
    return this.raftServer.Leader()
}

// func (this *electServer) doMachines(r *http.Request) (map[string]string, error) {
// return
// }

// func (this *electServer) doPeers(r *http.Request) (map[string]string, error) {
// log.Noticef("peers : %+v", this.raftServer.Peers())
// log.Noticef("lead : %+v", this.raftServer.Leader())
// log.Noticef("stat : %+v", this.raftServer.GetState())
// log.Noticef("name : %+v", this.raftServer.Name())
// log.Noticef("path : %+v", this.raftServer.Path())
// var peers = make(map[string]string)
// for name, peer := range this.raftServer.Peers() {
// peers[name] = peer.ConnectionString
// }
// return peers, nil
// }

func (this *electServer) doJoin(r *http.Request) error {
    var command raft.DefaultJoinCommand
    if err := json.NewDecoder(r.Body).Decode(&command); err != nil {
        return web.HTTPError{400, "MSG_ERROR"}
    }
    _, err := this.raftServer.Do(&command)
    if err != nil {
        log.Logf("raft do faild : %+v", err.Error())
        return err
    }
    log.Logf("raft join success")
    return nil
}

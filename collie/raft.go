package collie

import (
    "bytes"
    "encoding/json"
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

    if err = server.syncCluster(); err != nil {
        return
    }

    go func() {
        if err = server.ServeHttp(); err != nil {
            os.Exit(1)
        }
    }()
    return
}

func (this *electServer) ServeHttp() (err error) {
    this.mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { this.Router(w, r) })
    return this.StartHttp(this.mux)
}

func (this *electServer) InitRaftServer() (err error) {
    transporter := raft.NewHTTPTransporter("/raft", 200*time.Millisecond)
    this.raftServer, err = raft.NewServer(this.Name, this.GetDataDir(), transporter, nil, nil, "")
    if err != nil {
        return
    }
    transporter.Install(this.raftServer, this.mux)
    this.raftServer.Start()
    return
}

func (this *electServer) InitRaftCluster(leader string) (err error) {
    if leader != "" {
        if !this.raftServer.IsLogEmpty() {
            log.Log("cannot join with exist log")
        }
        if err = this.JoinLeader(leader); err != nil {
            return
        }
    } else if this.raftServer.IsLogEmpty() {
        if err = this.JoinSelf(); err != nil {
            return
        }
    } else {
        log.Log("recovered from log")
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
    resp.Body.Close()
    return
}

func (this *electServer) Router(w http.ResponseWriter, r *http.Request) {
    var err error
    switch r.URL.Path {
    case "/join":
        err = this.doJoin(r)
    case "/peers":
        err = json.NewEncoder(w).Encode(this.doPeers())
    case "/peer_config":
        err = json.NewEncoder(w).Encode(this.doPeerConfig())
    case "/peer_exchange":
        this.doExchangePeerConfig(r, w)
    default:
        resp := web.ApiResponse{404, "NOT_FOUND", nil}
        resp.Write(w)
    }
    if err != nil {
        resp := web.ApiResponse{502, err.Error(), nil}
        resp.Write(w)
    }
}

func (this *electServer) syncCluster() (err error) {
    var b bytes.Buffer
    if err = json.NewEncoder(&b).Encode(this.peerConfig); err != nil {
        return
    }

    var peers = make(PeerConfigMap)
    for name, peer := range this.raftServer.Peers() {
        resp, err := http.Post(peer.ConnectionString+"/peer_exchange", "application/json", &b)
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

func (this *electServer) doExchangePeerConfig(r *http.Request, w http.ResponseWriter) {
    var peerConfig = new(PeerConfig)
    if err := json.NewDecoder(r.Body).Decode(peerConfig); err != nil {
        w.WriteHeader(204)
    } else {
        this.peers[peerConfig.Name] = peerConfig
        json.NewEncoder(w).Encode(this.peerConfig)
    }
}

func (this *electServer) doPeerConfig() *PeerConfig {
    return this.peerConfig
}

func (this *electServer) doPeers() PeerConfigMap {
    return this.peers
}

// func (this *electServer) doLeader() (string, string) {
// return this.raftServer.Leader(), ""
// }

func (this *electServer) doLeader(r *http.Request) (leader string, err error) {
    return
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
    return err
}

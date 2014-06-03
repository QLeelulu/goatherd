package guard

import (
    "code.google.com/p/go.crypto/ssh"
    "log"

    "errors"
    "fmt"
    "goatherd/config"
    "goatherd/constant"
    "goatherd/library/sshclient"
    "net"
    "net/rpc"
)

const (
    DEFAULT_GUARD_ID   = "guard"
    DEFAULT_GUARD_HOST = "127.0.0.1"
    DEFAULT_GUARD_PORT = 8020
    DEFAULT_GUARD_USER = "cici"
    DEFAULT_GUARD_BIN  = "/usr/local/goatherd/bin/goatherd"
    DEFAULT_GUARD_LOG  = "/data/logs/goatherd/goatherd.log"
    DEFAULT_GUARD_PID  = "/data/logs/goatherd/goatherd.pid"
    DEFAULT_GUARD_KEY  = "/Users/cici/.ssh/id_dsa"
)

type Client struct {
    Id        string
    Host      string
    Port      int
    FilePid   string
    FileOut   string
    Conf      *config.GuardConfig
    con       *net.Conn
    client    *rpc.Client
    call      *rpc.Call
    sshHandel *sshclient.Handel
}

func NewClient(host string, port int) (client *Client) {
    client = &Client{
        Host: host,
        Port: port,
    }
    return
}

func NewClientDefault() (client *Client) {
    client = NewClientWithSsh(DEFAULT_GUARD_ID, DEFAULT_GUARD_HOST, DEFAULT_GUARD_PORT, DEFAULT_GUARD_LOG, DEFAULT_GUARD_PID, DEFAULT_GUARD_KEY, DEFAULT_GUARD_USER)
    return
}

func NewClientWithConfig(gconf *config.GuardConfig, mconf config.MasterConfig) (client *Client) {
    client = NewClientWithSsh(gconf.Id, gconf.Host, gconf.Port, gconf.LogFile, gconf.PidFile, mconf.AuthKeyFile, mconf.User)
    client.Conf = gconf
    return
}

func NewClientWithSsh(id, host string, port int, fileOut, filePId, key, user string) (client *Client) {
    client = &Client{
        Id:      id,
        Host:    host,
        Port:    port,
        FilePid: filePId,
        FileOut: fileOut,
        sshHandel: &sshclient.Handel{
            Host:       host,
            User:       user,
            PrivateKey: key,
        },
    }
    return
}

func (this *Client) SshKill() (err error) {
    return this.SshSignal(7)
}

func (this *Client) SshSignal(signal int) (err error) {
    if this.sshHandel == nil {
        return errors.New("ssh handel not initialized")
    }
    if err = this.sshHandel.ResetSession(); err != nil {
        return
    }
    if this.FilePid == "" {
        err = errors.New("pid path should set")
    } else {
        var cmd = fmt.Sprintf("source /etc/profile;cat %s | xargs kill -%d", this.FilePid, signal)
        this.sshHandel.Start(cmd)
    }
    return
}

func (this *Client) SshStatus() (status constant.PROCESS_STATUS, err error) {
    if this.sshHandel == nil {
        err = errors.New("ssh handel not initialized")
        return
    }
    if err = this.sshHandel.ResetSession(); err != nil {
        return
    }

    if this.FilePid == "" {
        err = errors.New("pid path should set")
    } else {
        var cmd = fmt.Sprintf("source /etc/profile;pid=$(tail -1 %s 2>/dev/null);if [ -z $pid ];then exit 1;else ps -p ${pid};fi", this.FilePid)
        fmt.Println("status :", cmd)
        _err := this.sshHandel.Run(cmd)
        if _err != nil {
            fmt.Println("status run:", _err.Error())
        } else {
            fmt.Println("status done")
        }
        if _, ok := _err.(*ssh.ExitError); ok {
            status = constant.PROCESS_STATUS_STOPED
        } else if _err == nil {
            status = constant.PROCESS_STATUS_STARTED
        } else {
            err = _err
        }
    }
    return
}

func (this *Client) SshStart() (err error) {
    if this.sshHandel == nil {
        return errors.New("ssh handel not initialized")
    }
    if err = this.sshHandel.ResetSession(); err != nil {
        return
    }

    if this.FilePid == "" {
        err = errors.New("pid path should set")
    } else if this.FileOut == "" {
        err = errors.New("log path should set")
    } else {
        var cmd = fmt.Sprintf("source /etc/profile;if nohup goatherd -id=%s -port=%d -log=%s > /dev/null &;then echo $! > %s;fi", this.Id, this.Port, this.FileOut, this.FilePid)
        log.Printf("start cmd:%s\n", cmd)
        err = this.sshHandel.Run(cmd)
    }
    return
}

func (this *Client) SshWait() (err error) {
    if this.sshHandel == nil {
        return errors.New("session not started")
    }
    err = this.sshHandel.Wait()
    return
}

func (this *Client) Dial() (err error) {
    var addr = fmt.Sprintf("%s:%d", this.Host, this.Port)
    this.client, err = rpc.Dial(SOCKET_TYPE, addr)
    return
}

func (this *Client) Call(arg Arg, ret *Ret) (err error) {
    /* log.Printf("guard client call:%+v", arg) */
    defer func() {
        /* log.Printf("call:%+v", arg) */
        ret.Err = err
    }()

    if arg.CheckTarget() {
        switch arg.Action {
        case constant.ACT_START:
            err = this.SshStart()
        case constant.ACT_STOP:
            err = this.SshKill()
        case constant.ACT_STATUS:
            ret.Status, err = this.SshStatus()
        case constant.ACT_UPDATE:
            if err = this.SshKill(); err == nil {
                err = this.SshStart()
            }
        default:
            err = errors.New("method not support for guard")
        }
    } else {
        if this.client == nil {
            if err = this.Dial(); err != nil {
                return
            }
            defer this.Close()
        }
        err = this.client.Call("Server.Eval", arg, ret)
    }
    return
}

func (this *Client) Go(arg Arg) (err error) {
    if this.client == nil {
        if err = this.Dial(); err != nil {
            return
        }
    }
    var ret = NewRet()
    this.call = this.client.Go("GuardServer.Eval", arg, &ret, nil)
    return
}

func (this *Client) Wait() (ret *Ret, err error) {
    retCall := <-this.call.Done
    ret = retCall.Reply.(*Ret)
    err = retCall.Error
    return
}

func (this *Client) Close() (err error) {
    if this.client != nil {
        err = this.client.Close()
        this.client = nil
    }
    return
}

package slave

import (
    "errors"
    "fmt"
    "goatherd/library/sshclient"
    "net"
    "net/rpc"
)

type Client struct {
    Id        string
    Host      string
    Port      int
    BinPath   string
    ProcessId int
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

func NewClientWithSsh(id, host string, port int, fileIn, fileOut, fileErr string) (client *Client) {
    client = &Client{
        Id:   id,
        Host: host,
        Port: port,
        sshHandel: &sshclient.Handel{
            Host:    host,
            FileIn:  fileIn,
            FileOut: fileOut,
            FileErr: fileErr,
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

    var cmd = fmt.Sprintf("kill -%d %d", signal, this.ProcessId)
    this.sshHandel.Start(cmd)
    return
}

func (this *Client) SshStart() (err error) {
    if this.sshHandel == nil {
        return errors.New("ssh handel not initialized")
    }
    if err = this.sshHandel.ResetSession(); err != nil {
        return
    }

    var cmd = fmt.Sprintf("%s -port=:%d", this.Port)
    this.sshHandel.Run(cmd)
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
    var addr = fmt.Sprint("%s:%d", this.Host, this.Port)
    this.client, err = rpc.Dial(SOCKET_TYPE, addr)
    return
}

func (this *Client) Call(args Arg, ret *Ret) (err error) {
    if this.client == nil {
        if err = this.Dial(); err != nil {
            return
        }
        defer this.Close()
    }
    err = this.client.Call("SlaveServer.Eval", args, ret)
    return
}

func (this *Client) Go(args Arg) (err error) {
    if this.client == nil {
        if err = this.Dial(); err != nil {
            return
        }
    }
    var ret = NewRet()
    this.call = this.client.Go("SlaveServer.Eval", args, &ret, nil)
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

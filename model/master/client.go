package master

import (
    "fmt"
    "log"
    "net"
    "net/rpc"
)

const (
    SOCKET_TYPE = "tcp"
)

type Client struct {
    Host   string
    Port   int
    con    *net.Conn
    client *rpc.Client
    call   *rpc.Call
}

func (this *Client) Dial() (err error) {
    var addr = fmt.Sprintf("%s:%d", this.Host, this.Port)
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
    log.Printf("call:%+v", *this)
    err = this.client.Call("Server.Eval", args, ret)
    return
}

func (this *Client) Close() (err error) {
    if this.client != nil {
        err = this.client.Close()
        this.client = nil
    }
    return
}

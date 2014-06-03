package master

import (
    "errors"
    "fmt"
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

func NewClient(host string, port int) *Client {
    return &Client{
        Host: host,
        Port: port,
    }
}

func (this *Client) Dial() (err error) {
    var addr = fmt.Sprintf("%s:%d", this.Host, this.Port)
    this.client, err = rpc.Dial(SOCKET_TYPE, addr)
    return
}

func (this *Client) Call(arg Arg, ret *Ret) (err error) {
    /* log.Printf("master client call:%+v", arg) */
    defer func() {
        /* log.Printf("call:%+v", arg) */
        ret.Err = err
    }()

    if arg.CheckTarget() {
        switch arg.Action {
        default:
            err = errors.New("method not support for target")
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

func (this *Client) Close() (err error) {
    if this.client != nil {
        err = this.client.Close()
        this.client = nil
    }
    return
}

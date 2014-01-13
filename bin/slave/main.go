package main

import (
    "flag"

    "goatherd/model/slave"
)

var port int

func init() {
    flag.IntVar(&port, "port", 8000, "slave服务端口")
}

func main() {
    flag.Parse()

    slave.StartNewRpcServer(port)
}

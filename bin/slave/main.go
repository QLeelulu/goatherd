package main

import (
    "log"
    "os"
    "strconv"

    "goatherd/model/slave"
)

func main() {
    if len(os.Args) != 2 {
        log.Fatal("bad params")
    }
    if port, err := strconv.Atoi(os.Args[1]); err != nil {
        log.Fatal("bad port")
    } else {
        slave.StartNewRpcServer(port)
    }
}

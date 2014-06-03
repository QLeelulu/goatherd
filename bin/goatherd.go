package main

import (
    "flag"
    "log"
    "os"

    "goatherd/guard"
)

var port int
var id string
var logfile string

func init() {
    flag.IntVar(&port, "port", 8018, "guard服务端口")
    flag.StringVar(&id, "id", "guard", "guard服务id")
    flag.StringVar(&logfile, "log", "", "guard日志输出路径")
}

func main() {
    flag.Parse()
    if logfile != "" {
        if output, err := os.OpenFile(logfile, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666); err != nil {
            log.Fatal("guard log file open faild")
        } else {
            log.SetOutput(output)
            log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
        }
    }
    log.Print("goatherd will start at:", port)
    log.Fatal(guard.StartNewRpcServer(id, port))
}

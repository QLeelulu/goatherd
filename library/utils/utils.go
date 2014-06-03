package utils

import (
    "net"
    "os"
    "runtime"
    "strconv"
    "syscall"
)

const (
    DEFAULT_NETWORK = "tcp"
)

func Fork() (pid uintptr, err error) {
    pid, pid1, sysErr := syscall.RawSyscall(syscall.SYS_FORK, 0, 0, 0)
    if sysErr != 0 {
        return
    }
    if runtime.GOOS == "darwin" && pid1 == 1 {
        pid = 0
    }
    return
}

func TryDial(host string, port int) (err error) {
    var conn net.Conn
    if conn, err = net.Dial(DEFAULT_NETWORK, net.JoinHostPort(host, strconv.Itoa(port))); err != nil {
        return
    }
    defer conn.Close()
    return
}

func TryOpenFile(filePath string, flag int) (err error) {
    if filePath == "" {
        return
    }
    var fd *os.File
    if fd, err = os.OpenFile(filePath, flag, 0666); err != nil {
        return
    }
    defer fd.Close()
    return
}

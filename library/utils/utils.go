package base

import (
    "runtime"
    "syscall"
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

package goatherd

import (
    "bytes"
    "errors"
    "fmt"
    "os/exec"
    "strings"
    "sync"
)

const (
    RUNNER_STATUS_INIT = iota
    RUNNER_STATUS_STARTING
    RUNNER_STATUS_START_FAILED
    RUNNER_STATUS_STARTED
    RUNNER_STATUS_STOPING
    RUNNER_STATUS_STOPED
)

type Runner struct {
    Name    string
    program *Program
    Cmd     *exec.Cmd
    mutex   sync.Mutex
    Status  int

    out bytes.Buffer
}

func (self *Runner) Start() error {
    self.mutex.Lock()
    defer self.mutex.Unlock()

    self.Status = RUNNER_STATUS_STARTING
    cmdArr := strings.Split(self.program.Command, " ")
    cmdName := cmdArr[0]
    cmdArgs := []string{}
    if len(cmdArr) > 1 {
        cmdArgs = cmdArr[1:]
    }
    self.Cmd = exec.Command(cmdName, cmdArgs...)
    self.Cmd.Stdout = &self.out
    self.Cmd.Stderr = &self.out
    err := self.Cmd.Start()
    if err != nil {
        self.Status = RUNNER_STATUS_START_FAILED
        return err
    }
    go self.watch()
    self.Status = RUNNER_STATUS_STARTED
    return nil
}

func (self *Runner) watch() {
    err := self.Cmd.Wait()
    fmt.Println(self.Name, "has exit.")
    if err != nil {
        fmt.Printf("[%s] got error: %s\n", self.Name, err.Error())

        fmt.Printf("The date is %q\n", self.out.String())
    }

    // process is runing
    if self.Status == RUNNER_STATUS_STARTED {
        if self.program.AutoRestart {
            fmt.Println(self.Name, "is auto restarting.")
            self.Start()
        }
    }
}

func (self *Runner) Restart() (err error) {
    self.mutex.Lock()
    defer self.mutex.Unlock()

    err = self.Stop()
    err = self.Start()
    return err
}

func (self *Runner) Stop() error {
    self.mutex.Lock()
    defer self.mutex.Unlock()

    if self.Cmd.Process == nil {
        return errors.New(fmt.Sprintf("%s not started.", self.Name))
    }

    if self.Cmd.ProcessState.Exited() {
        return errors.New(fmt.Sprintf("%s has stoped.", self.Name))
    }

    err := self.Cmd.Process.Kill()
    return err
}

package sshclient

import (
    "code.google.com/p/go.crypto/ssh"
    "errors"
    "os"

    _ "crypto/sha1"
)

const (
    SSH_PORT = "22"
)

type Handel struct {
    Host       string
    Port       string
    User       string
    PrivateKey string
    FileIn     string
    FileOut    string
    FileErr    string
    FilePId    string
    session    *ssh.Session
    client     *ssh.ClientConn
    auth       *ssh.ClientConfig
}

func (this *Handel) ResetSession() (err error) {
    if this.session != nil {
        err = this.session.Close()
        this.session = nil
    }
    return
}

func (this *Handel) check() (err error) {
    if this.auth == nil {
        keys := new(keychain)
        if err = keys.LoadPEM(this.PrivateKey); err != nil {
            return
        }
        this.auth = &ssh.ClientConfig{
            User: this.User,
            Auth: []ssh.ClientAuth{
                ssh.ClientAuthKeyring(keys),
            },
        }
    }
    if this.Port == "" {
        this.Port = SSH_PORT
    }
    if this.client == nil {
        this.client, err = ssh.Dial("tcp", this.Host+":"+this.Port, this.auth)
        if err != nil {
            return
        }
    }
    if this.session == nil {
        this.session, err = this.client.NewSession()
        if err != nil {
            return
        }
    }

    //stdin
    if this.FileIn == "" {
        this.session.Stdin = os.Stdin
    } else if fd, err := os.Open(this.FileIn); err == nil {
        this.session.Stdin = fd
    } else {
        return errors.New("process config error : bad input file : " + err.Error())
    }
    //stdout
    if this.FileOut == "" {
        this.session.Stdout = os.Stdout
    } else if fd, err := os.OpenFile(this.FileOut, os.O_WRONLY|os.O_CREATE, 0666); err == nil {
        this.session.Stdout = fd
    } else {
        return errors.New("process config error : bad output file : " + err.Error())
    }
    //stderr
    if this.FileErr == "" {
        this.session.Stderr = os.Stderr
    } else if fd, err := os.OpenFile(this.FileErr, os.O_WRONLY|os.O_CREATE, 0666); err == nil {
        this.session.Stderr = fd
    } else {
        return errors.New("process config error : bad error file : " + err.Error())
    }

    return
}

func (this *Handel) reset() {
    if this.session != nil {
        this.session.Close()
        this.session = nil
    }
    if this.client != nil {
        this.client = nil
    }
}

func (this *Handel) Run(cmd string) (err error) {
    defer this.reset()
    if err = this.check(); err != nil {
        return
    }
    err = this.session.Run(cmd)
    return
}

func (this *Handel) Start(cmd string) (err error) {
    if err = this.check(); err != nil {
        return
    }
    err = this.session.Start(cmd)
    return
}

func (this *Handel) Wait() (err error) {
    if this.session == nil || this.client == nil {
        err = errors.New("session not created")
        return
    }
    defer this.reset()
    err = this.session.Wait()
    return
}

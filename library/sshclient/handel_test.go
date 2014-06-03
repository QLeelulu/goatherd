package sshclient

import "testing"

func TestTestHandel(t *testing.T) {
    client := &Handel{
        Host:       "c41.clicki.cn",
        Port:       "22",
        User:       "root",
        PrivateKey: "/Users/cici/.ssh/id_dsa",
        FileOut:    "test.out",
        FileErr:    "test.err",
    }
    var cmd = "/root/cici/test/main"
    if err := client.Start(cmd); err != nil {
        t.Fatal("client start faild:", err.Error())
    }
    if err := client.Wait(); err != nil {
        t.Fatal("client wait faild:", err.Error())
    }
}

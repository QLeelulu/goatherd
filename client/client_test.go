package client

import "testing"

func TestGetClient(t *testing.T) {
    conf, err := NewHandel().GetConfig()
    if err != nil {
        t.Log("get config faild:", err.Error())
    }
    t.Logf("master config:%+v", conf)
}

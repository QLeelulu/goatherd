package goatherd

import (
    "dsp_masky/library/pvsz"
    "errors"
    "goatherd/library/myhttp"
    "net/http"
    "path"
    "strings"
    "sunteng/commons/db/myetcd"
)

type Contex struct {
    *Goatherd
}

var Root string

func (this *httpServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    if err := goatherdRouter(w, r); err != nil {
    }
}

func (this *httpServer) goatherdRouter(w http.ResponseWriter, r *http.Request) (err error) {
    var target = "/goatherd/"
    if !strings.HasPrefix(r.URL.Path, target) {
        return
    }
    var action = strings.TrimPrefix(r.URL.Path, target)
    switch action {
    case "list":
        myhttp.APIResponse(w, r, func() (data interface{}, err error) { return this.doSheepList(r) })
    case "ping":
    default:
        err = errors.New("bad http url")
        return
    }
}

func (this *httpServer) doSheepList(r *http.Request) (interface{}, error) {
    children, err := GetChildren(this.Name)
    var nodes []map[string]interface{}
    for _, child := range children {
        status, err := GetStatus(child)
        if err != nil {
            status = myetcd.STATUS_UNKNOWN
        }
        nodes[child] = map[string]interface{}{"status": status}
    }
    return map[string]interface{}{
        "nodes": nodes,
    }, nil
}

func GetStatus(name string) (status string, err error) {
    etcdCtl := myetcd.NewClient()
    etcdRoad := pvsz.NewRoad(Root)

    statusPath := etcdRoad.FormatKey(myetcd.KEY_STATUS, father)
    resp, err := this.Slipper.Get(statusPath, false, false)
    if err != nil {
        return
    }
    if myetcd.ErrorKeyNotFound(err) {
        status = myetcd.STATUS_STOPPED
    }
    return
}

func GetChildren(father string) (children []string, err error) {
    etcdCtl := myetcd.NewClient()
    etcdRoad := pvsz.NewRoad(Root)

    // check status
    status, err := GetStatus(father)
    if status == myetcd.STATUS_STOPPED {
        err = errors.New(father + " has stopped")
        return
    }

    // get children
    fatherPath = etcdRoad.FormatKey(myetcd.KEY_FAMILY, father)
    resp, err := this.Slipper.Get(home, false, true)
    if err != nil {
        return
    }
    for _, nodes := range resp.Node.Nodes {
        children = append(children, path.Base(nodes.Key))
    }
    return
}

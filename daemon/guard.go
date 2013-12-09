package base

//客户,客户端访问Guard,查询和控制Slave的任务
type Guest interface {
}

//用户指令,Guest和Guard的通讯包
type Card interface {
    Marshal(buf []byte)
}

//奴隶,定期向Guard发送report，并接受来自Guard的Command，每个slave使用独立的配置（可reload，支持多group和*匹配）
type Slave interface {
    Listen() error
    Excute() error
    Report() error
    Feed() error
}

//哨兵,独立进程,监听（可以是文件、jsonrpc、protorpc、mq、db。。。）来自Guest的消息，并发送Command到Slave
type Guard interface {
    Listen() error
    Accept() (card Card, err error)
    Notify(cmd Command)
}

//系统指令,Guard和Slave间通讯
type Command interface {
    Unmarshal(buf []byte)
}

//队长,根据配置，部署多个Guard
type Captain interface {
    AddGuard(guard Guard) error
    DelGuard(guard Guard) error
    CleanGuards(guard Guard) error
    ListGuards() map[int]Guard
}

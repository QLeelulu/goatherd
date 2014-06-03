package constant

const (
    ACT_NULL ACT_ID = iota
    ACT_STATUS
    ACT_START
    ACT_STOP
    ACT_RESTART
    ACT_UPDATE
    ACT_CHECK
    ACT_CREATE
    ACT_DESTROY
    ACT_TAIL
    ACT_CALL
    ACT_ADD
    ACT_DEL
    ACT_GET
    ACT_MOD
    ACT_TAILF
    ACT_CONFIG_GET
    ACT_CONFIG_SET
    ACT_CONFIG_RESET
    ACT_CONFIG_REWRITE
    ACT_HEARTBEAT
    ACT_ID_MAX
)

const (
    PROCESS_STATUS_INIT PROCESS_STATUS = iota
    PROCESS_STATUS_STARTING
    PROCESS_STATUS_START_FAILED
    PROCESS_STATUS_STARTED
    PROCESS_STATUS_STOPING
    PROCESS_STATUS_STOPED
    PROCESS_STATUS_CRASHED
    PROCESS_STATUS_MAX
)

type ACT_ID int

var ACT_ID_MAP = make([]string, ACT_ID_MAX)

func (this ACT_ID) GoString() string {
    return "act_" + ACT_ID_MAP[this]
}

type PROCESS_STATUS int

func (this PROCESS_STATUS) GoString() string {
    return "process_status_" + PROCESS_STATUS_MAP[this]
}

var PROCESS_STATUS_MAP = make([]string, PROCESS_STATUS_MAX)

func init() {
    ACT_ID_MAP[ACT_STATUS] = "status"
    ACT_ID_MAP[ACT_START] = "start"
    ACT_ID_MAP[ACT_STOP] = "stop"
    ACT_ID_MAP[ACT_RESTART] = "restart"
    ACT_ID_MAP[ACT_UPDATE] = "update"
    ACT_ID_MAP[ACT_CHECK] = "check"
    ACT_ID_MAP[ACT_CREATE] = "create"
    ACT_ID_MAP[ACT_DESTROY] = "destroy"
    ACT_ID_MAP[ACT_TAIL] = "tail"
    ACT_ID_MAP[ACT_CALL] = "call"
    ACT_ID_MAP[ACT_ADD] = "add"
    ACT_ID_MAP[ACT_DEL] = "del"
    ACT_ID_MAP[ACT_GET] = "get"
    ACT_ID_MAP[ACT_MOD] = "mod"
    ACT_ID_MAP[ACT_CONFIG_GET] = "config get"
    ACT_ID_MAP[ACT_CONFIG_SET] = "config set"
    ACT_ID_MAP[ACT_CONFIG_RESET] = "config reset"
    ACT_ID_MAP[ACT_HEARTBEAT] = "heart beat"

    PROCESS_STATUS_MAP[PROCESS_STATUS_STARTING] = "starting"
    PROCESS_STATUS_MAP[PROCESS_STATUS_START_FAILED] = "start faild"
    PROCESS_STATUS_MAP[PROCESS_STATUS_STARTED] = "started"
    PROCESS_STATUS_MAP[PROCESS_STATUS_STOPING] = "stoping"
    PROCESS_STATUS_MAP[PROCESS_STATUS_STOPED] = "stoped"
    PROCESS_STATUS_MAP[PROCESS_STATUS_CRASHED] = "crashed"
}

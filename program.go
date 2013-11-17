package goatherd

type Program struct {
    Name         string
    Command      string
    AutoStart    bool
    AutoRestart  bool
    StartRetries int
    NumProcs     int
    Environment  map[string]string
}

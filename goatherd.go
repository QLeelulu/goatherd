package goatherd

import "log"

var programs map[string]*Program
var config *Config

func init() {
    var err error
    config, err = loadConf()
    if err != nil {
        log.Fatalf("load config file faild: %s", err.Error())
    }

    programs = make(map[string]*Program)
    for name, program := range config.Programs {
        program.Name = name
        programs[name] = &program
    }
    /* fmt.Printf("config:%+v", *config) */
}

func RunAll() {
    for name, program := range programs {
        runner := Runner{}
        runner.Name = name
        runner.program = program
        err := runner.Start()
        if err != nil {
            log.Println("start", name, "got error:", err)
        }
    }
}

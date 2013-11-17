package main

import (
    "goatherd"
    "time"
)

func main() {
    goatherd.RunAll()
    for {
        time.Sleep(time.Hour)
    }
}

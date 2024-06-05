package main

import (
	"github.com/bluenviron/mediamtx/control"
	"os"
)

func main() {
	if len(os.Args) > 2 {
		control.RunControl(os.Args[1], "/ci", os.Args[2])
	} else if len(os.Args) > 1 {
		control.RunControl(":7002", "/ci", os.Args[1])
	} else {
		control.RunControl(":7002", "/ci", "ci.ini") // message broker URL is ws://127.0.0.1:7000/ci
	}
}

package main

import (
	//"internal/api"
	"github.com/bluenviron/mediamtx/internal/api"
	"os"
)

func main() {
	if len(os.Args) > 2 {
		api.RunControl(os.Args[1], os.Args[2])
	} else if len(os.Args) > 1 {
		api.RunControl("tcp://test.mosquitto.org:1883", os.Args[1])
	} else {
		api.RunControl("tcp://test.mosquitto.org:1883", "")
	}
}

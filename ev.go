package main

import (
	"github.com/bluenviron/mediamtx/internal/api"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		api.RunEvents(os.Args[1])
	} else {
		api.RunEvents("tcp://test.mosquitto.org:1883")
	}
}

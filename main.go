// main executable.
package main

import (
	"github.com/bluenviron/mediamtx/internal/api"
	"os"

	"github.com/bluenviron/mediamtx/internal/core"
)

func main() {
	go api.RunServer("tcp://test.mosquitto.org:1883")

	s, ok := core.New(os.Args[1:])
	if !ok {
		os.Exit(1)
	}

	s.Wait()
}

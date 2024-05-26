// main executable.
package main

import (
	"github.com/bluenviron/mediamtx/internal/api"
	"os"

	"github.com/bluenviron/mediamtx/internal/core"
)

func main() {
	go api.RunBroker()

	s, ok := core.New(os.Args[1:])
	if !ok {
		os.Exit(1)
	}
	//control.RunControl(":7002", "/ci", "ci.ini")

	s.Wait()
}

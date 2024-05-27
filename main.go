// main executable.
package main

import (
	"github.com/bluenviron/mediamtx/control"
	"os"

	"github.com/bluenviron/mediamtx/internal/core"
)

func main() {
	go control.RunBroker()

	s, ok := core.New(os.Args[1:])
	if !ok {
		os.Exit(1)
	}
	//control.RunControl(":7002", "/ci", "ci.ini")

	s.Wait()
}

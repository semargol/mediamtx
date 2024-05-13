// main executable.
package main

import (
	"github.com/bluenviron/mediamtx/internal/core"
	"os"
)

func main() {
	s, ok := core.New(os.Args[1:])
	if !ok {
		os.Exit(1)
	}
	//c := api.NewControl("127.0.0.1:7002", "127.0.0.1:7000")
	//go c.Once()

	s.Wait()
}

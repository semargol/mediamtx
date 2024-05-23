// main executable.
package main

import (
	"github.com/bluenviron/mediamtx/internal/api"
	"os"

	"github.com/bluenviron/mediamtx/internal/core"
)

func main() {
	// start ApiBroker
	//apiBroker := api.NewApiBroker("127.0.0.1:7000")
	//go apiBroker.Listen() // run ApiBroker forever
	go api.RunBroker()
	//go api.CreateServerAndListen("127.0.0.1:7001", "127.0.0.1:7000")

	s, ok := core.New(os.Args[1:])
	if !ok {
		os.Exit(1)
	}
	//c := api.NewControl("127.0.0.1:7002", "127.0.0.1:7000")
	//go c.Once()

	s.Wait()
}

package main

import "github.com/bluenviron/mediamtx/internal/api"

func main() {
	/*
		c := api.NewControl("127.0.0.1:7002", "127.0.0.1:7000")
		if len(os.Args) > 1 {
			c.Init(os.Args[1])
		} else {
			c.Init("")
		}
		c.Commands()
	*/
	api.RunControl()
}

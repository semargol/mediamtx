package main

import "github.com/bluenviron/mediamtx/internal/api"

func main() {
	c := api.NewControl("127.0.0.1:7002", "127.0.0.1:7000")
	c.Once()
}

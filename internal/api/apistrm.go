package api

import (
	"time"
)

type StreamApiServer struct {
	//broker *ApiBroker
	server *ApiServer
}

// NewStrmApiServer allocates a StrmApiServer.
func NewStreamApiServer(network string, address string, readTimeout time.Duration, api *API) (*StreamApiServer, error) {
	//ln, err := net.Listen(network, address)
	//if err != nil {
	//	return nil, err
	//}

	//h := handler
	//h = &handlerFilterRequests{h}
	//h = &handlerFilterRequests{h}
	//h = &handlerServerHeader{h}
	//h = &handlerLogger{h, parent}
	//h = &handlerExitOnPanic{h}

	s := &StreamApiServer{}
	//s.broker = NewApiBroker("127.0.0.1:7000")
	s.server = NewApiServer("127.0.0.1:7001", "127.0.0.1:7000", api)
	//go s.broker.Listen()
	go s.server.Listen()

	//c := NewControl("127.0.0.1:7002", "127.0.0.1:7000")
	//c.Once()
	//c.Once()
	//c.Once()

	return s, nil
}

func (s *StreamApiServer) Close() {
	_ = s.server.udpConn.Close()
}

package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/bluenviron/mediamtx/internal/conf"
)

var GlobalApiServerStruct ApiServer

var fromServerToBroker = make(chan *Message, 10)
var fromBrokerToServer = make(chan *Message, 10)
var interrupt = make(chan os.Signal, 1)
var done = make(chan struct{})

// NewStrmApiServer allocates a StrmApiServer.
func NewStreamApiServer(api *API) (*StreamApiServer, error) {
	s := &StreamApiServer{}
	s.server = newApiServer(api)
	go s.server.Listen()
	return s, nil
}

func (s *StreamApiServer) Close() {
}

func newApiServer(api *API) *ApiServer {
	ctx, cancel := context.WithCancel(context.Background())
	var s *ApiServer = new(ApiServer)
	s.ctx = ctx
	s.cancel = cancel

	defaultStrmConf := conf.InitializeDefaultStrmConf()
	conf.StrmGlobalConf = defaultStrmConf
	s.strmConf = &defaultStrmConf
	s.api = api

	return s
}

/*
type StrmInterface interface {
	SetRtp(id uint, params map[string]string)
	SetRtsp(id uint, params map[string]string)
}

type CiMessage struct {
	Ci    string
	Verb  string
	Comp  string
	Param map[string]string
}
*/

type StreamApiServer struct {
	//broker *ApiBroker
	server *ApiServer
}

type ApiServer struct {
	api        *API
	strmConf   *conf.StrmConf
	eventsChan chan string
	ctx        context.Context
	mutex      sync.Mutex
	cancel     func()
	respev     Message
}

func (s *ApiServer) GetStrmConf() *conf.StrmConf {
	return s.strmConf
}

func (s *ApiServer) sendTo(msg Message) {
	fromServerToBroker <- &msg
	//t.Transceiver.SendTo(msg, t.brokerAddr)
}

//func (t *Server) SendToAll(msg Message, subscriberAddrList map[string]*net.UDPAddr) {
//	t.transceiver.sendToAll(msg, subscriberAddrList)
//}

//func (t *ApiServer) ReceiveFrom(msec int) (msg Message, fromAddr *net.UDPAddr, err error) {
//	msg = * <- FromBrokerToServer
//	//msg, fromAddr, err = t.Transceiver.ReceiveFrom(msec)
//	return msg, "control", nil
//}

func (s *ApiServer) receive(msec int) (msg *Message, err error) {
	select {
	case msg = <-fromBrokerToServer:
		return msg, nil
	case <-time.After(time.Duration(msec) * time.Millisecond):
		return nil, errors.New("timeout")
	}
	return nil, nil
}

func (s *ApiServer) sendEvent(event Message) {
	s.sendTo(event)
}

func (s *ApiServer) getEventsChan() chan string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.eventsChan
}

func (s *ApiServer) SetEventsChan(newChan chan string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.eventsChan = newChan
}

func (s *ApiServer) StartEventListener() {
	go func() {
		for {
			select {
			case event := <-s.getEventsChan():
				eventMsg := Message{0, "msg", "evn", "", "", make(map[string]string), nil}
				eventMsg.Data = map[string]string{"status": event}
				s.sendEvent(eventMsg)
				fmt.Println("event: ", eventMsg)
			case <-s.ctx.Done():
				return
			default:
				time.Sleep(100 * time.Millisecond)
				break
			}
		}
	}()
}

func (s *ApiServer) updateEventsChan() {
	go func() {
		var prevEventsChan chan string
		for {
			select {
			case <-s.ctx.Done():
				return
			default:
				r := s.api.Parent.GetRTSPServer()
				if r != nil && r.EventsChan != nil {
					if r.EventsChan != prevEventsChan {
						// fmt.Println("r.EventsChan: ", r.EventsChan)
						s.SetEventsChan(r.EventsChan)
						prevEventsChan = r.EventsChan
					}
				}
				// Sleep for a while before checking again
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()
}

func (s *ApiServer) Listen() {
	log.Println("Start api server")
	//s.PublishAt("evn")
	//s.PublishAt("res")
	//s.SubscribeAt("req")
	var request Message
	var response Message
	s.updateEventsChan()
	s.StartEventListener()
	//var from *net.UDPAddr
	//var err error
	for {
		reqmsg, err := s.receive(10)
		if err == nil {
			request = *reqmsg
			fmt.Println("request: ", request)
			switch request.Verb + "/" + request.Noun {
			case "log/on":
				{
					s.strmConf.LogLavel = 1
					ConfigSync(s)
				}
			case "log/off":
				{
					s.strmConf.LogLavel = 4
					ConfigSync(s)
				}
			case "get/config":
				{
					jsonConf, readableConf, err := GetStrmConfig(s)
					if err != nil {
						log.Fatalf("Error getting stream config: %v", err)
					}

					fmt.Println("JSON Configuration:")
					fmt.Println(jsonConf)

					fmt.Println("\nReadable Configuration:")
					fmt.Println(readableConf)

					response = request
					response.Conf = s.strmConf
				}
			case "add/pipe":
				{
					response, _ = ApiAddPipe(s, &request)
				}
			case "del/pipe":
				{
					response, _ = ApiDelPipe(s, &request)
				}
			case "set/pipe":
				{
					response, _ = ApiSetPipe(s, &request)
				}
			case "get/pipe":
				{
					response, _ = ApiGetPipe(s, &request)
				}

			// case "add/rtp":
			// 	{
			// 		response, _ = ApiAddRtp(s.api, &request)
			// 	}
			// case "del/rtp":
			// 	{
			// 		response, _ = ApiDelRtp(s.api, &request)
			// 	}
			case "set/rtpr":
				{
					conf.STRMGlobalConfiguration.SetRtpr(request.Data) // also set buf size = 0
					// response, _ = ApiSetRtp(s.api, &request)
					response, _ = ApiUpdatePipeConfig(s, &request, "RTPR")
				}
			case "get/rtpr":
				{
					// response, _ = ApiGetRtp(s.api, &request)
					response, _ = ApiGetSubConfigField(s, &request, "RTPR")
					//fmt.Println("response: ", response)
				}
			case "set/rtps":
				{
					conf.STRMGlobalConfiguration.SetRtps(request.Data)
					// response, _ = ApiSetRtp(s.api, &request)
					response, _ = ApiUpdatePipeConfig(s, &request, "RTPS")
				}
			case "get/rtps":
				{
					// response, _ = ApiGetRtp(s.api, &request)
					response, _ = ApiGetSubConfigField(s, &request, "RTPS")
					//fmt.Println("response: ", response)
				}
			case "set/rtspcl":
				{
					//conf.STRMGlobalConfiguration.SetRtspcl(request.Data)
					// response, _ = ApiSetRtp(s.api, &request)
					response, _ = ApiUpdatePipeConfig(s, &request, "RTSPCL")
				}
			case "get/rtspcl":
				{
					// response, _ = ApiGetRtp(s.api, &request)
					response, _ = ApiGetSubConfigField(s, &request, "RTSPCL")
					//fmt.Println("response: ", response)
				}
			case "set/rtspsrv":
				{
					response, _ = ApiSetRtsp(s, &request)
				}
			case "get/rtspsrv":
				{
					response, _ = ApiGetRtsp(s, &request)
				}
			case "set/buf":
				{
					conf.STRMGlobalConfiguration.SetBuf(request.Data)
					//response, _ = ApiUpdatePipeConfig(s, &request, "BUF")
					response.Data["result"] = "unknown command"
				}
			case "get/buf":
				{
					response, _ = ApiGetSubConfigField(s, &request, "BUF")
					//response.Data["result"] = "unknown command"
				}

			default:
				{
					response = request
					response.Data["result"] = "unknown command"
				}
			}
			//response = request
			response.Name = "msg"
			response.Topic = "res"
			response.Corr = request.Corr
			s.sendTo(response)
			//fmt.Pgetrintln("Server Sent response: ", response)
			//response.Topic = "evn"
			//s.sendEvent(response)
		}
	}
}

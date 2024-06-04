package api

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bluenviron/mediamtx/internal/conf"
)

var FromServerToBroker = make(chan *Message, 10)
var FromBrokerToServer = make(chan *Message, 10)

func CreateServerAndListen(sep string, bep string, api *API) {
	var serverApp *ApiServer = NewApiServer(sep, bep, api)
	serverApp.Listen()
}

func NewApiServer(serverEp string, brokerEp string, api *API) *ApiServer {
	ctx, cancel := context.WithCancel(context.Background())
	var s *ApiServer = new(ApiServer)
	s.ctx = ctx
	s.cancel = cancel

	defaultStrmConf := conf.InitializeDefaultStrmConf()
	conf.StrmGlobalConf = &defaultStrmConf
	s.strmConf = &defaultStrmConf
	/*
		s.Transceiver.Open(serverEp)
		addr, err := net.ResolveUDPAddr("udp", brokerEp)
		if err != nil {
			fmt.Println("Error resolving ApiServer UDP address:", err)
			return nil
		}
		s.brokerAddr = addr
	*/
	s.api = api

	//ConfigSync(s)

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

type ApiServer struct {
	//Transceiver
	//brokerAddr *net.UDPAddr
	api        *API
	strmConf   *conf.StrmConf
	eventsChan chan string
	ctx        context.Context
	mutex      sync.Mutex
	cancel     func()
	respev     Message
}

func (t *ApiServer) GetStrmConf() *conf.StrmConf {
	return t.strmConf
	//t.Transceiver.SendTo(msg, t.brokerAddr)
}

func (t *ApiServer) SendTo(msg Message) {
	FromServerToBroker <- &msg
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

func (t *ApiServer) Receive(msec int) (msg *Message, err error) {
	select {
	case msg = <-FromBrokerToServer:
		return msg, nil
	case <-time.After(time.Duration(msec) * time.Millisecond):
		return nil, errors.New("timeout")
	}
	return nil, nil
}

//func (t *ApiServer) SubscribeAt(topic string) {
//	t.Transceiver.SubscribeAt(topic, t.brokerAddr)
//}

//func (t *ApiServer) PublishAt(topic string) {
//	t.Transceiver.PublishAt(topic, t.brokerAddr)
//}

func (t *ApiServer) SendEvent(event Message) {
	t.SendTo(event)
}

func (t *ApiServer) SubscribeAt(topic string) {
	var sub Message = Message{0, "sub", topic, "", "", make(map[string]string), nil}
	t.SendTo(sub)
}

func (t *ApiServer) PublishAt(topic string) {
	var pub Message = Message{0, "pub", topic, "", "", make(map[string]string), nil}
	t.SendTo(pub)
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
				s.SendEvent(eventMsg)
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
	s.PublishAt("evn")
	s.PublishAt("res")
	s.SubscribeAt("req")
	var request Message
	var response Message
	s.updateEventsChan()
	s.StartEventListener()
	//var from *net.UDPAddr
	//var err error
	for {
		reqmsg, err := s.Receive(10)
		if err == nil {
			request = *reqmsg
			fmt.Println("request: ", request)
			switch request.Verb + "/" + request.Noun {
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
			s.SendTo(response)
			//fmt.Pgetrintln("Server Sent response: ", response)
			response.Topic = "evn"
			s.SendEvent(response)
		}
	}
}

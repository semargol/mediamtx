package api

import (
	"fmt"
	"net"

	"github.com/bluenviron/mediamtx/internal/conf"
)

func CreateServerAndListen(sep string, bep string, api *API) {
	var serverApp *ApiServer = NewApiServer(sep, bep, api)
	serverApp.Listen()
}

func NewApiServer(serverEp string, brokerEp string, api *API) *ApiServer {
	var s *ApiServer = new(ApiServer)
	defaultStrmConf := conf.InitializeDefaultStrmConf()
	s.strmConf = &defaultStrmConf
	s.transceiver.open(serverEp)
	addr, err := net.ResolveUDPAddr("udp", brokerEp)
	if err != nil {
		fmt.Println("Error resolving ApiServer UDP address:", err)
		return nil
	}
	s.brokerAddr = addr
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
	transceiver
	brokerAddr *net.UDPAddr
	api        *API
	strmConf   *conf.StrmConf
}

func (t *ApiServer) SendTo(msg Message) {
	t.transceiver.sendTo(msg, t.brokerAddr)
}

//func (t *Server) SendToAll(msg Message, subscriberAddrList map[string]*net.UDPAddr) {
//	t.transceiver.sendToAll(msg, subscriberAddrList)
//}

func (t *ApiServer) ReceiveFrom(msec int) (msg Message, fromAddr *net.UDPAddr, err error) {
	msg, fromAddr, err = t.transceiver.receiveFrom(msec)
	return
}

func (t *ApiServer) SubscribeAt(topic string) {
	t.transceiver.SubscribeAt(topic, t.brokerAddr)
}

func (t *ApiServer) PublishAt(topic string) {
	t.transceiver.PublishAt(topic, t.brokerAddr)
}

func (t *ApiServer) SendEvent(event Message) {
	t.transceiver.sendTo(event, t.brokerAddr)
}

func (s *ApiServer) Listen() {
	fmt.Println("Start api server at ", s.endPoint.String(), " with broker at ", s.brokerAddr.String())
	s.PublishAt("evn")
	s.PublishAt("res")
	s.SubscribeAt("req")
	var request Message
	var response Message
	//var from *net.UDPAddr
	var err error
	for {
		request, _, err = s.ReceiveFrom(10)
		if err == nil {
			fmt.Println("request: ", request)
			//println("Server Received request: ", request.String())
			if request.Verb == "init" {
				ConfigSync(s)
			}
			switch request.Verb + "/" + request.Noun {
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

			// case "add/rtsp":
			// 	{
			// 		response, _ = ApiAddRtsp(s.api, &request)
			// 	}
			// case "del/rtsp":
			// 	{
			// 		response, _ = ApiDelRtsp(s.api, &request)
			// 	}
			case "set/rtsp":
				{
					response, _ = ApiSetRtsp(s.api, &request)
				}
			case "get/rtsp":
				{
					response, _ = ApiGetRtsp(s.api, &request)
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
			//response.Data = request.Data
			s.SendTo(response)
			//fmt.Pgetrintln("Server Sent response: ", response)
			response.Topic = "evn"
			s.SendEvent(response)
		}
	}
}

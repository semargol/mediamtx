package control

import (
	"encoding/json"
	"github.com/bluenviron/mediamtx/internal/api"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

// var fromServer = make(chan *api.Message, 10)
var fromControl = make(chan *api.Message, 10)

//var serverAddr net.UDPAddr

// var serverConnection api.Transceiver
var controlConnection *websocket.Conn = nil
var topicList TopicList
var inputJson bool = false

func init() {
	//serverAddrRef, _ := net.ResolveUDPAddr("udp", "127.0.0.1:7001")
	//serverAddr = *serverAddrRef
	topicList.TopicMap = make(map[string]*Topic)
}

func ListenControlConnections() {
	http.HandleFunc("/cihtml", strmhtml)
	http.HandleFunc("/ci", strm)
	http.HandleFunc("/", home)
	err := http.ListenAndServe(":7002", nil)
	if err != nil {
		panic(err)
	}
}

var controlBrokerUpgrader = websocket.Upgrader{} // use default options

func strm(w http.ResponseWriter, r *http.Request) {
	if controlConnection != nil {
		log.Print("Only one control connection allowed")
		return
	}
	var err error = nil
	controlConnection, err = controlBrokerUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("controlBrokerUpgraderError:", err)
		return
	}
	defer func() {
		err := controlConnection.Close()
		if err != nil {
		}
		controlConnection = nil
	}()
	inputJson = true
	topicList.publish("req", "control")
	topicList.subscribe("res", "control")
	RunControlReader()
}

/*
	func OpenServerConnection() {
		//serverConnection.Open(":7000")
	}

	func CloseControlConnection() {
		//
	}

	func CloseServerConnection() {
		//serverConnection.close()
	}
*/
func RunControlReader() {
	//var msg Message
	for {
		mt, buf, err := controlConnection.ReadMessage()
		if err != nil || mt != websocket.TextMessage {
			log.Println("controlBrokerReadError:", err)
			//CloseControlConnection()  will be closed in strm
			//controlConnection = nil
			break
		}

		//msg.Parse(string(buf))
		//txt := msg.String()
		var msg *api.Message = new(api.Message)
		uerr := json.Unmarshal(buf, msg)
		if uerr != nil {
			msg.Parse(string(buf))
			bytes, merr := json.Marshal(msg)
			if merr != nil {
				continue
			}
			uerr = json.Unmarshal(bytes, msg)
			if uerr != nil {
				continue
			}
			msg.Topic = "req"
		}
		//fmt.Println("BROK BrokerControlReader: ", msg) //%s, type: %d", message, mt)
		fromControl <- msg
		//serverBroker.topicList.push(msg, from, &serverBroker.transceiver)
	}
}

/*
func RunServerReader() {
	log.Println("Start api broker")
	//var msg Message
	//var from *net.UDPAddr
	//var err error
	for {
		//var msg *api.Message = new(api.Message)
		//*msg, from, err = serverConnection.ReceiveFrom(10)
		//msg := <- api.FromServerToBroker
		//if err == nil {
			//serverAddr = *from
			//log.Println("BROK BrokerServerReader: ", from, " ", msg) //%s, type: %d", message, mt)
			//fromServer <- msg
		//}
	}
}
*/

func pushMessage(msg *api.Message, from string) {
	switch msg.Name {
	case "pub":
		topicList.publish(msg.Topic, from)
	case "sub":
		topicList.subscribe(msg.Topic, from)
	case "rem":
		topicList.unsubscribe(msg.Topic, from)
		topicList.unpublish(msg.Topic, from)
	case "msg":
		topic, ok := topicList.TopicMap[msg.Topic]
		if ok {
			for _, addr := range topic.SubscriberList {
				if addr != from {
					if addr == "server" {
						api.FromBrokerToServer <- msg
						//_, _ = serverConnection.UdpConn.WriteToUDP(data, &serverAddr)
					}
					if addr == "control" && inputJson == true {
						data, _ := json.Marshal(msg)
						_ = controlConnection.WriteMessage(websocket.TextMessage, data)
						//log.Println("BROK SendToControl: ", msg) //%s, type: %d", message, mt)
					}
					if addr == "control" && inputJson == false {
						response(msg)
						//_ = controlConnection.WriteMessage(websocket.TextMessage, []byte(msg.String()))
					}
					//fmt.Println("send to ", addr, "msg", msg)
				}
			}
		}
	}
}

func RunBroker() {
	go ListenControlConnections()
	for {
		select {
		case msg := <-api.FromServerToBroker:
			//log.Println("BROK ReadFromServer:  ", msg)
			pushMessage(msg, "server")
		case msg := <-fromControl:
			//log.Println("BROK ReadFromControl: ", msg)
			pushMessage(msg, "control")
		case <-interrupt:
			log.Println("BROK interrupt")
			return
		case <-done:
			log.Println("BROK done")
			return
		case <-time.After(time.Second * 2):
			if inputJson == false && controlConnection != nil {
				msg := api.Message{0, "msg", "evn", "tic", "", make(map[string]string), nil}
				msg.Data["result"] = "100"
				msg.Data["description"] = "timer event occurs every 2000 msec"
				event(&msg)
			}
		}
	}
}

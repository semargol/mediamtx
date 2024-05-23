package api

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
)

var fromServer = make(chan *Message, 10)
var fromControl = make(chan *Message, 10)

var serverAddr net.UDPAddr
var serverConnection transceiver
var controlConnection *websocket.Conn = nil
var topicList TopicList

func init() {
	serverAddrRef, _ := net.ResolveUDPAddr("udp", "127.0.0.1:7001")
	serverAddr = *serverAddrRef
	topicList.TopicMap = make(map[string]*Topic)
}

func OpenControlConnection() {
	http.HandleFunc("/strm", strm)
	//http.HandleFunc("/", home)
	err := http.ListenAndServe(":7002", nil)
	if err != nil {
		panic(err)
	}
}

var controlBrokerUpgrader = websocket.Upgrader{} // use default options

func strm(w http.ResponseWriter, r *http.Request) {
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
	RunControlReader()
}

func OpenServerConnection() {
	serverConnection.open(":7000")
}

func CloseControlConnection() {
	//
}

func CloseServerConnection() {
	//serverConnection.close()
}

func RunControlReader() {
	//var msg Message
	for {
		mt, buf, err := controlConnection.ReadMessage()
		if err != nil || mt != websocket.TextMessage {
			log.Println("controlBrokerReadError:", err)
			break
		}

		//msg.Parse(string(buf))
		//txt := msg.String()
		var msg *Message = new(Message)
		err = json.Unmarshal(buf, msg)
		//fmt.Println("BROK BrokerControlReader: ", msg) //%s, type: %d", message, mt)
		fromControl <- msg
		//serverBroker.topicList.push(msg, from, &serverBroker.transceiver)
	}
}

func RunServerReader() {
	fmt.Println("Start api broker at ", serverConnection.endPoint.String())
	//var msg Message
	var from *net.UDPAddr
	var err error
	for {
		var msg *Message = new(Message)
		*msg, from, err = serverConnection.receiveFrom(10)
		if err == nil {
			serverAddr = *from
			//log.Println("BROK BrokerServerReader: ", from, " ", msg) //%s, type: %d", message, mt)
			fromServer <- msg
		}
	}
}

func pushMessage(msg *Message, from string) {
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
			data, _ := json.Marshal(msg)
			for _, addr := range topic.SubscriberList {
				if addr != from {
					if addr == "server" {
						_, _ = serverConnection.udpConn.WriteToUDP(data, &serverAddr)
					}
					if addr == "control" {
						_ = controlConnection.WriteMessage(websocket.TextMessage, data)
					}
					//fmt.Println("send to ", addr, "msg", msg)
				}
			}
		}
	}

}

func RunBroker() {
	OpenServerConnection()
	go OpenControlConnection()
	go RunServerReader()
	//go RunControlReader()
	for {
		select {
		case msg := <-fromServer:
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
		}
	}
	CloseControlConnection()
	CloseServerConnection()
}

/*
type ApiServerBroker struct {
	transceiver
	topicList TopicList
}

var serverBroker *ApiServerBroker = nil

func CreateServerBrokerAndListen(bep string) {
	serverBroker = NewApiBroker(bep)
	serverBroker.Listen()
}

func NewApiBroker(ep string) *ApiServerBroker {
	http.HandleFunc("/strm", strm)
	http.HandleFunc("/", home)
	err := http.ListenAndServe(":7000", nil)
	if err != nil {
		panic(err)
	}

	var b *ApiServerBroker = new(ApiServerBroker)
	b.transceiver.open(ep)
	b.topicList = newTopicList()
	return b
}

func (b *ApiServerBroker) Listen() {
	fmt.Println("Start api broker at ", b.endPoint.String())
	var msg Message
	var from *net.UDPAddr
	var err error
	for {
		msg, from, err = b.receiveFrom(10)
		if err == nil {
			b.topicList.push(msg, from, &b.transceiver)
		}
	}
}

type ApiControlBroker struct {
	//transceiver
	topicList TopicList
}

var controlBroker *ApiControlBroker = nil

func CreateControlBrokerAndListen(bep string) {
	serverBroker = NewApiBroker(bep)
	serverBroker.Listen()
}

var controlBrokerUpgrader = websocket.Upgrader{} // use default options

func strm(w http.ResponseWriter, r *http.Request) {
	c, err := controlBrokerUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("controlBrokerUpgraderError:", err)
		return
	}
	defer c.Close()

	var msg Message
	from, _ := net.ResolveUDPAddr("udp", ":12000")

	for {
		mt, buf, err := c.ReadMessage()
		if err != nil || mt != websocket.TextMessage {
			log.Println("controlBrokerReadError:", err)
			break
		}

		err = json.Unmarshal(buf, &msg)
		log.Printf("controlBrokerRead: ", msg) //%s, type: %d", message, mt)

		serverBroker.topicList.push(msg, from, &serverBroker.transceiver)
		//err = c.WriteMessage(mt, message)
		//if err != nil {
		//	log.Println("write err:", err)
		//	break
		//}

		//log.Printf("send: %s, type: %d", "sdjhnfvviwerbg", 1)
		//err = c.WriteMessage(1, []byte("sdjhnfvviwerbg"))
		//if err != nil {
		//	log.Println("write err:", err)
		//	break
		//}
	}
}
*/

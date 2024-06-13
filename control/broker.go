package control

import (
	"encoding/json"
	"github.com/bluenviron/mediamtx/internal/api"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

var fromControl = make(chan *api.Message, 10) // channel to receive incoming directives

var controlConnection *websocket.Conn = nil
var topicList TopicList
var inputJson bool = false

func init() {
	topicList.TopicMap = make(map[string]*Topic)
}

func ListenControlConnections() {
	http.HandleFunc("/cihtml", strmhtml)
	http.HandleFunc("/ci", strm)             // to connect from ci
	http.HandleFunc("/", home)               // to connect from browser
	err := http.ListenAndServe(":7002", nil) // temporary fixed port 7002
	if err != nil {
		panic(err)
	}
}

var controlBrokerUpgrader = websocket.Upgrader{} // use default options

// function to serve ci connection
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
	inputJson = true // json encoding for ci
	topicList.publish("req", "control")
	topicList.subscribe("res", "control")
	RunControlReader()
}

// read json encoded command from control web socket and send to channel fromControl
func RunControlReader() {
	for {
		mt, buf, err := controlConnection.ReadMessage()
		if err != nil || mt != websocket.TextMessage {
			log.Println("controlBrokerReadError:", err)
			break
		}

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
		fromControl <- msg
	}
}

// execute directive
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

// dispatch messages from
func RunBroker() {
	go ListenControlConnections()
	for {
		select {
		case msg := <-api.FromServerToBroker: // response from server
			pushMessage(msg, "server")
		case msg := <-fromControl: // request  from control
			pushMessage(msg, "control")
		case <-interrupt:
			log.Println("BROK interrupt")
			return
		case <-done:
			log.Println("BROK done")
			return
		case <-time.After(time.Second * 2): // wait 2 second
			if inputJson == false && controlConnection != nil {
				msg := api.Message{0, "msg", "evn", "tic", "", make(map[string]string), nil}
				msg.Data["result"] = "100"
				msg.Data["description"] = "timer event occurs every 2000 msec"
				event(&msg) // test time events
			}
		}
	}
}

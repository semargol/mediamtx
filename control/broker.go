package control

import (
	"encoding/json"
	"github.com/bluenviron/mediamtx/internal/api"
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"net"
	"net/http"
	"time"
)

var fromServer = make(chan *api.Message, 10)
var fromControl = make(chan *api.Message, 10)

var serverAddr net.UDPAddr
var serverConnection api.Transceiver
var controlConnection *websocket.Conn = nil
var topicList TopicList
var inputJson bool = false

func init() {
	serverAddrRef, _ := net.ResolveUDPAddr("udp", "127.0.0.1:7001")
	serverAddr = *serverAddrRef
	topicList.TopicMap = make(map[string]*Topic)
}

func OpenControlConnection() {
	http.HandleFunc("/cihtml", strmhtml)
	http.HandleFunc("/ci", strm)
	http.HandleFunc("/", home)
	//http.HandleFunc("/", home)
	err := http.ListenAndServe(":7002", nil)
	if err != nil {
		panic(err)
	}
}

var controlBrokerUpgrader = websocket.Upgrader{} // use default options

func __home(w http.ResponseWriter, r *http.Request) {
	__homeTemplate.Execute(w, "ws://"+r.Host+"/ci")
}

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

func OpenServerConnection() {
	serverConnection.Open(":7000")
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

func RunServerReader() {
	log.Println("Start api broker at ", serverConnection.EndPoint.String())
	//var msg Message
	var from *net.UDPAddr
	var err error
	for {
		var msg *api.Message = new(api.Message)
		*msg, from, err = serverConnection.ReceiveFrom(10)
		if err == nil {
			serverAddr = *from
			//log.Println("BROK BrokerServerReader: ", from, " ", msg) //%s, type: %d", message, mt)
			fromServer <- msg
		}
	}
}

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
			data, _ := json.Marshal(msg)
			for _, addr := range topic.SubscriberList {
				if addr != from {
					if addr == "server" {
						_, _ = serverConnection.UdpConn.WriteToUDP(data, &serverAddr)
					}
					if addr == "control" && inputJson == true {
						_ = controlConnection.WriteMessage(websocket.TextMessage, data)
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
		case <-time.After(time.Second * 2):
			if inputJson == false && controlConnection != nil {
				msg := api.Message{0, "msg", "evn", "tic", "", make(map[string]string), nil}
				msg.Data["result"] = "100"
				msg.Data["description"] = "timer event occurs every 2000 msec"
				event(&msg)
			}
		}
	}
	CloseControlConnection()
	CloseServerConnection()
}

var __homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>  
window.addEventListener("load", function(evt) {

    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;

    var print = function(message) {
        var d = document.createElement("div");
        d.textContent = message;
        output.appendChild(d);
        output.scroll(0, output.scrollHeight);
    };

    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };

    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
        ws.send(input.value);
        return false;
    };

    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };

});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server, 
"Send" to send a message to the server and "Close" to close the connection. 
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output" style="max-height: 70vh;overflow-y: scroll;"></div>
</td></tr></table>
</body>
</html>
`))

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

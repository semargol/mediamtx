package api

import (
	"bufio"
	"encoding/json"
	"time"

	//"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"os"
	"strings"
)

var cmdNumber int = 1000
var connection *websocket.Conn = nil

var interrupt = make(chan os.Signal, 1)
var done = make(chan struct{})
var fromBroker = make(chan *Message, 10)

// var toBroker = make(chan *Message)
var fromConsole = make(chan *Message, 10)

func OpenBrokerConnection() {
	u := url.URL{Scheme: "ws", Host: ":7002", Path: "/strm"}
	log.Printf("CTRL connecting to %s", u.String())

	var err error = nil
	connection, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("OpenBrokerReader:", err)
	}
}

func RunBrokerReader() {
	//defer close(done)
	for {
		mt, buf, err := connection.ReadMessage()
		if err != nil {
			log.Println("ReadFromBrokerError: ", err)
			return
		} else if mt != websocket.TextMessage {
			log.Println("ReadFromBrokerError: ", "not a text message")
			return
		} else {
			var msg *Message = new(Message)
			err = json.Unmarshal(buf, msg)
			//log.Println("CTRL ReadFromBroker:  ", msg) //%s, type: %d", message, mt)
			fromBroker <- msg
		}
	}
}

/*
func RunBrokerSender() {
	//defer close(done)
	for {
		select {
		case msg := <-toBroker:
			err := connection.WriteMessage(websocket.TextMessage, []byte(msg.String()))
			if err != nil {
				log.Println("WriteToBrokerError:", err)
			}
		}
	}
}
*/

func CloseBrokerConnection() {
	_ = connection.Close()
}

func OpenConsoleReader() {

}

func RunConsoleReader() {
	for {
		fmt.Print(cmdNumber+1, ">")
		text, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		text, _ = strings.CutSuffix(text, "\n")
		text, _ = strings.CutSuffix(text, "\r")
		msg := new(Message)
		msg.Parse(text)
		cmdNumber++
		msg.Corr = cmdNumber
		msg.Topic = "req"
		fmt.Println(msg) //%s, type: %d", message, mt)
		fromConsole <- msg
		time.Sleep(50 * time.Millisecond)
	}
}

func CloseConsoleReader() {

}

func ConsoleReader() {
	OpenConsoleReader()
	RunConsoleReader()
	CloseConsoleReader()
}

func RunControl() {
	OpenBrokerConnection()
	go RunBrokerReader()
	go ConsoleReader()

	//var sub Message = Message{0, "sub", "res", "", "", make(map[string]string)}
	//_ = connection.WriteMessage(websocket.TextMessage, []byte(sub.String()))
	//var pub Message = Message{0, "pub", "req", "", "", make(map[string]string)}
	//_ = connection.WriteMessage(websocket.TextMessage, []byte(pub.String()))

	fromConsole <- &Message{0, "sub", "res", "", "", make(map[string]string)}
	fromConsole <- &Message{0, "pub", "req", "", "", make(map[string]string)}

	for {
		select {
		case msg := <-fromBroker:
			fmt.Println(msg.Corr, " ", msg)
			//fmt.Println("ReadFromBroker: ", msg)
		case msg := <-fromConsole:
			bytes, _ := json.Marshal(msg)
			err := connection.WriteMessage(websocket.TextMessage, bytes)
			if err != nil {
				log.Println("WriteToBrokerError:", err)
			}
		case <-interrupt:
			log.Println("interrupt")
			return
		case <-done:
			log.Println("done")
			return
		}
	}
}

/*
func CreateControlAndListen(bep string) {
	//flag.Parse()
	//log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: ":12000", Path: "/strm"}
	log.Printf("connecting to %s", u.String())

	connection, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer connection.Close()

	done := make(chan struct{})

	//go ReadFromBroker()
		go func() {
			defer close(done)
			for {

				mt, message, err := connection.ReadMessage()
				if err != nil {
					log.Println("read:", err)
					return
				}
				log.Printf("recv: %s, type: %d", message, mt)
			}
		}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}
*/

/*
type Control struct {
	transceiver
	brokerAddr *net.UDPAddr
	cmdNumber  int
}

var controlApp Control

func CreateControlAndTest(bep string) *Control {
	var controlApp *Control = NewControl(cep, bep)
	//Test(controlApp)
	return controlApp
}

func (c *Control) SendTo(msg Message) {
	c.transceiver.sendTo(msg, c.brokerAddr)
}

//func (t *Control) SendToAll(msg Message, subscriberAddrList map[string]*net.UDPAddr) {
//	t.transceiver.sendToAll(msg, subscriberAddrList)
//}

func (c *Control) ReceiveFrom(msec int) (msg Message, fromAddr *net.UDPAddr, err error) {
	msg, fromAddr, err = c.transceiver.receiveFrom(msec)
	return
}

func (c *Control) SubscribeAt(topic string) {
	c.transceiver.SubscribeAt(topic, c.brokerAddr)
}

func (c *Control) PublishAt(topic string) {
	c.transceiver.PublishAt(topic, c.brokerAddr)
}

func NewControl(controlEp string, brokerEp string) *Control {
	var c *Control = new(Control)
	c.transceiver.open(controlEp)
	c.brokerAddr, _ = net.ResolveUDPAddr("udp", brokerEp)
	fmt.Println("Start control application at ", c.endPoint.String(), " with broker at ", c.brokerAddr.String())
	return c
}

func (c *Control) OneCommand(text string) {
	text, _ = strings.CutSuffix(text, "\n")
	text, _ = strings.CutSuffix(text, "\r")
	var request Message = NewMessage(text)
	c.cmdNumber++
	request.Corr = c.cmdNumber
	request.Topic = "req"
	//fmt.Println("msg ", request)
	c.SendTo(request)
	response, _, err := c.ReceiveFrom(10000)
	if err != nil {
		fmt.Println("timeout, more than 10000 msec", response)
	} else {
		if response.Data == nil {
			fmt.Println(c.cmdNumber, "??")
		} else if response.Data["result"] == "" {
			fmt.Println(c.cmdNumber, "__", response.Data)
		} else {
			fmt.Println(c.cmdNumber, response.Data)
		}
	}
}

func (c *Control) Init(path string) {
	c.cmdNumber = 1000
	c.PublishAt("req")
	c.SubscribeAt("res")

	if len(path) > 0 {
		var file *os.File
		file, err := os.Open(path)
		if err == nil {
			fmt.Println("ci.ini = ", path)
			reader := bufio.NewReader(file)
			for {
				text, eof := reader.ReadString('\n')
				if eof == io.EOF {
					break
				}
				fmt.Println(">", text)
				c.OneCommand(text)
			}
		} else {
			fmt.Println("Ini file not found ", path)
		}
	} else {
		fmt.Println("No ini file")
	}
}

func (c *Control) Commands() {
	for {
		fmt.Print(c.cmdNumber+1, ">")
		text, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		text, _ = strings.CutSuffix(text, "\n")
		text, _ = strings.CutSuffix(text, "\r")

		if text == "help" || text == "h" || text == "?" {
			showHelp()
		} else {
			c.OneCommand(text)
		}
	}
}

func (c *Control) Events() {
	c.SubscribeAt("evn")
	for {
		event, _, err := c.ReceiveFrom(100)
		if err == nil {
			fmt.Println(event)
		}
	}
}

func showHelp() {
	fmt.Println("Usage: <verb> <noun> { <key>=<value> }")
	fmt.Println("<verb>: add,del,get,set  - required action")
	fmt.Println("<noun>: pipe,rtp,rtsp    - target")
	fmt.Println("<key>=<value>            - parameters")
	fmt.Println("Examples:")
	fmt.Println("add pipe id=1")
	fmt.Println("set rtsp id=1 path=stream_1")
	fmt.Println("set rtsp port=554")
}
*/

package api

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func CreateControlAndTest(cep string, bep string) *Control {
	var controlApp *Control = NewControl(cep, bep)
	//Test(controlApp)
	return controlApp
}

type Control struct {
	transceiver
	brokerAddr *net.UDPAddr
	cmdNumber  int
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

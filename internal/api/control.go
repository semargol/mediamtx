package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"io"
	"os"
	"strings"
)

var cmdNumber int = 1001

func onConsoleRequest(message *Message) {
	//fmt.Printf("%s\n", message.Sprint())
}

func onConsoleResponse(message *Message) {
	//fmt.Printf("%s\n", message.Sprint())
}

func onConsoleEvent(message *Message) {
	//fmt.Printf("%s\n", message.Sprint())
}

func onConsoleConfig(message *Message) {
	//fmt.Printf("%s\n", message.Sprint())
}

func onConsoleMqttRequest(client mqtt.Client, message mqtt.Message) {
	var msg Message
	_ = json.Unmarshal(message.Payload(), &msg)
	onConsoleRequest(&msg)
	//fmt.Printf("Received message: %s from topic: %s\n", message.Payload(), message.Topic())
}

func onConsoleMqttResponse(client mqtt.Client, message mqtt.Message) {
	var msg Message
	_ = json.Unmarshal(message.Payload(), &msg)
	onConsoleResponse(&msg)
	//fmt.Printf("Received message: %s from topic: %s\n", message.Payload(), message.Topic())
}

func onConsoleMqttEvent(client mqtt.Client, message mqtt.Message) {
	var msg Message
	_ = json.Unmarshal(message.Payload(), &msg)
	onConsoleEvent(&msg)
	//fmt.Printf("Received message: %s from topic: %s\n", message.Payload(), message.Topic())
}

func onConsoleMqttConfig(client mqtt.Client, message mqtt.Message) {
	var msg Message
	_ = json.Unmarshal(message.Payload(), &msg)
	onConsoleConfig(&msg)
	//fmt.Printf("Received message: %s from topic: %s\n", message.Payload(), message.Topic())
}

func RunControl(addr string, init string) {
	OpenMqttConnection(addr)
	SubscribeMqttConnection("req", onConsoleMqttRequest)
	SubscribeMqttConnection("res", onConsoleMqttResponse)
	SubscribeMqttConnection("evn", onConsoleMqttEvent)
	SubscribeMqttConnection("cfg", onConsoleMqttConfig)
	if init != "" {
		DoScript(init)
	}
	RunConsoleReader()
}

func RunConsoleReader() {
	for {
		fmt.Print(cmdNumber, ">")
		text, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		text, _ = strings.CutSuffix(text, "\n")
		text, _ = strings.CutSuffix(text, "\r")
		DoCommand(text)
	}
}

func DoCommand(text string) {
	if len(text) > 0 {
		if text[0] == '@' {
			DoScript(text[1:])
		} else {
			var msg Message
			msg.Parse(text)
			msg.Corr = cmdNumber
			msg.Topic = "req"
			SendRequest(&msg)
			cmdNumber++
		}
	}
}

func DoScript(path string) {
	if len(path) > 0 {
		//var file *os.File
		file, err := os.Open(path)
		if err == nil {
			reader := bufio.NewReader(file)
			for {
				text, eof := reader.ReadString('\n')
				if eof == io.EOF {
					break
				}
				text, _ = strings.CutSuffix(text, "\n")
				text, _ = strings.CutSuffix(text, "\r")
				DoCommand(text)
			}
		} else {
			fmt.Println("Ini file not found ", path)
		}
	} else {
		fmt.Println("No ini file")
	}
}

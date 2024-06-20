package api

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"time"
)

var evRequest chan Message = make(chan Message)
var evResponse chan Message = make(chan Message)
var evEvent chan Message = make(chan Message)
var evConfig chan Message = make(chan Message)

func RunEvents(addr string) {
	OpenMqttConnection(addr)
	SubscribeMqttConnection("req", onMqttRequest)
	SubscribeMqttConnection("res", onMqttResponse)
	SubscribeMqttConnection("evn", onMqttEvent)
	SubscribeMqttConnection("cfg", onMqttConfig)
	for {
		select {
		case msg := <-evRequest:
			onRequest(msg)
			break
		case msg := <-evResponse:
			onResponse(msg)
			break
		case msg := <-evEvent:
			onEvent(msg)
			break
		case msg := <-evConfig:
			onConfig(msg)
			break
		case <-interrupt:
			log.Println("interrupt")
			return
		case <-done:
			log.Println("done")
			return
		case <-time.After(time.Second * 2): // wait 2 second
		}
	}
}

func onRequest(message Message) {
	fmt.Printf("%s\n", message.Sprint())
}

func onResponse(message Message) {
	m := message.SprintAll()
	for n := 0; n < len(m); n++ {
		fmt.Printf("%s\n", m[n])
	}
}

func onEvent(message Message) {
	for _, t := range message.SprintAll() {
		fmt.Printf("%s\n", t)
	}
}

func onConfig(message Message) {
	for _, t := range message.SprintCfg() {
		fmt.Printf("%s\n", t)
	}
}

func onMqttRequest(client mqtt.Client, message mqtt.Message) {
	var msg Message
	_ = json.Unmarshal(message.Payload(), &msg)
	evRequest <- msg
	//onRequest(msg)
	//fmt.Printf("Received message: %s from topic: %s\n", message.Payload(), message.Topic())
}

func onMqttResponse(client mqtt.Client, message mqtt.Message) {
	var msg Message
	_ = json.Unmarshal(message.Payload(), &msg)
	evResponse <- msg
	//onResponse(msg)
	//fmt.Printf("Received message: %s from topic: %s\n", message.Payload(), message.Topic())
}

func onMqttEvent(client mqtt.Client, message mqtt.Message) {
	var msg Message
	_ = json.Unmarshal(message.Payload(), &msg)
	evEvent <- msg
	//onEvent(msg)
	//fmt.Printf("Received message: %s from topic: %s\n", message.Payload(), message.Topic())
}

func onMqttConfig(client mqtt.Client, message mqtt.Message) {
	var msg Message
	_ = json.Unmarshal(message.Payload(), &msg)
	evConfig <- msg
	//onConfig(msg)
	//fmt.Printf("Received message: %s from topic: %s\n", message.Payload(), message.Topic())
}

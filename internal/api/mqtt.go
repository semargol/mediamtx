package api

import (
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/websocket"
	"time"
)

var mqttConnection mqtt.Client
var websConnection *websocket.Conn
var rootTopic string = "strm/ci"

func SendRequest(message *Message) {
	//onReceive_requestChannel(message, message)
	bytes, _ := json.Marshal(message)
	if mqttConnection.IsConnected() {
		token := mqttConnection.Publish(rootTopic+"/req", 0, false, string(bytes))
		token.WaitTimeout(time.Millisecond * 500)
	} else if websConnection != nil {
		_ = websConnection.WriteMessage(websocket.TextMessage, bytes)
	}
}

func SendResponse(message *Message) {
	//onReceive_requestChannel(message, message)
	bytes, _ := json.Marshal(message)
	if mqttConnection.IsConnected() {
		token := mqttConnection.Publish(rootTopic+"/res", 0, false, string(bytes))
		token.WaitTimeout(time.Millisecond * 500)
	} else if websConnection != nil {
		_ = websConnection.WriteMessage(websocket.TextMessage, bytes)
	}
}

func SendEvent(message *Message) {
	onReceive_requestChannel(message, message)
	bytes, _ := json.Marshal(message)
	if mqttConnection.IsConnected() {
		token := mqttConnection.Publish(rootTopic+"/evn", 0, false, string(bytes))
		token.WaitTimeout(time.Millisecond * 500)
	} else if websConnection != nil {
		_ = websConnection.WriteMessage(websocket.TextMessage, bytes)
	}
}

func SendConfig(message *Message) {
	onReceive_requestChannel(message, message)
	bytes, _ := json.Marshal(message)
	if mqttConnection.IsConnected() {
		token := mqttConnection.Publish(rootTopic+"/cfg", 0, false, string(bytes))
		token.WaitTimeout(time.Millisecond * 500)
	} else if websConnection != nil {
		_ = websConnection.WriteMessage(websocket.TextMessage, bytes)
	}
}

/*
func onResponse(message Message) {
	fmt.Printf("Received response: %s  %s\n", message.Topic, message.String())
}

func onEvent(message Message) {
	fmt.Printf("Received event:    %s  %s\n", message.Topic, message.String())
}

func onConfig(message Message) {
	fmt.Printf("Received config:   %s  %s\n", message.Topic, message.String())
}
*/

func SubscribeMqttConnection(topic string, callback mqtt.MessageHandler) {
	if token := mqttConnection.Subscribe(rootTopic+"/"+topic, 0, callback); token.Wait() && token.Error() != nil {
		panic(fmt.Sprint("Error subscribing to topic", topic, ":", token.Error()))
	}
	//fmt.Println("Subscribed to", topic)
}

func OpenMqttConnection(addr string) {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(addr)

	mqttConnection = mqtt.NewClient(opts)
	if token := mqttConnection.Connect(); token.Wait() && token.Error() != nil {
		panic(fmt.Sprint("Error connecting to MQTT broker:", token.Error()))
	}
	fmt.Println("Connected to mqtt at", addr)
	/*
		if token := mqttConnection.Subscribe(rootTopic+"/req", 0, onMqttRequest); token.Wait() && token.Error() != nil {
			panic(fmt.Sprint("Error subscribing to topic REQ:", token.Error()))
		}
		fmt.Println("Subscribed to REQ")

		if token := mqttConnection.Subscribe(rootTopic+"/res", 0, onMqttResponse); token.Wait() && token.Error() != nil {
			panic(fmt.Sprint("Error subscribing to topic RES:", token.Error()))
		}
		fmt.Println("Subscribed to RES")

		if token := mqttConnection.Subscribe(rootTopic+"/evn", 0, onMqttEvent); token.Wait() && token.Error() != nil {
			panic(fmt.Sprint("Error subscribing to topic EVN:", token.Error()))
		}
		fmt.Println("Subscribed to EVN")

		if token := mqttConnection.Subscribe(rootTopic+"/cfg", 0, onMqttConfig); token.Wait() && token.Error() != nil {
			panic(fmt.Sprint("Error subscribing to topic CFG:", token.Error()))
		}
		fmt.Println("Subscribed to CFG")
	*/
}

/*
func OpenWebsConnection(addr string) {
	var err error = nil
	websConnection, _, err = websocket.DefaultDialer.Dial(addr, nil)
	if err != nil {
		panic(fmt.Sprint("Error connecting to WS broker:", err))
	}
	fmt.Println("Connected to websocket at", addr)

	go RunWebsConnectionReader()
}

func RunWebsConnectionReader() {
	for {
		mt, buf, err := websConnection.ReadMessage()
		if err != nil {
			log.Println("ReadFromBrokerError: ", err)
			return
		} else if mt != websocket.TextMessage {
			log.Println("ReadFromBrokerError: ", "not a text message")
			return
		} else {
			var msg *Message = new(Message)
			err = json.Unmarshal(buf, msg)
			if err == nil {
				//log.Println("CTRL ReadFromBroker:  ", msg) //%s, type: %d", message, mt)
				if msg.Topic == rootTopic+"/res" {
					onResponse(msg)
				}
				if msg.Topic == rootTopic+"/evn" {
					onEvent(*msg)
				}
				if msg.Topic == rootTopic+"/cfg" {
					onConfig(*msg)
				}
				//onResponse(*msg)
				//fromBroker <- msg
				continue
			}

			//	var cfg *conf.StrmConf = new(conf.StrmConf)
			//	err = json.Unmarshal(buf, cfg)
			//	if err == nil {
			//		fromBrokerConfig <- cfg
			//		continue
			//	}

			log.Println("CTRL ReadFromBroker not recognized: ", string(buf)) //%s, type: %d", message, mt)
		}
	}
}

func onMqttRequest(client mqtt.Client, message mqtt.Message) {
	var msg Message
	_ = json.Unmarshal(message.Payload(), &msg)
	onRequest(msg)
	//fmt.Printf("Received message: %s from topic: %s\n", message.Payload(), message.Topic())
}

func onMqttResponse(client mqtt.Client, message mqtt.Message) {
	var msg Message
	_ = json.Unmarshal(message.Payload(), &msg)
	onResponse(msg)
	//fmt.Printf("Received message: %s from topic: %s\n", message.Payload(), message.Topic())
}

func onMqttEvent(client mqtt.Client, message mqtt.Message) {
	var msg Message
	_ = json.Unmarshal(message.Payload(), &msg)
	onEvent(msg)
	//fmt.Printf("Received message: %s from topic: %s\n", message.Payload(), message.Topic())
}

func onMqttConfig(client mqtt.Client, message mqtt.Message) {
	var msg Message
	_ = json.Unmarshal(message.Payload(), &msg)
	onConfig(msg)
	//fmt.Printf("Received message: %s from topic: %s\n", message.Payload(), message.Topic())
}
*/

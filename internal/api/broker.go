package api

import (
	"fmt"
	"net"
)

func CreateBrokerAndListen(bep string) {
	var brokerApp *ApiBroker = NewApiBroker(bep)
	brokerApp.Listen()
}

type ApiBroker struct {
	transceiver
	topicList TopicList
}

func NewApiBroker(ep string) *ApiBroker {
	var b *ApiBroker = new(ApiBroker)
	b.transceiver.open(ep)
	//b.EndPoint, _ = net.ResolveUDPAddr("udp", ep)
	b.topicList = newTopicList()
	return b
}

func (b *ApiBroker) Listen() {
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

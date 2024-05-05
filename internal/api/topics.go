package api

import (
	"net"
)

type TopicList struct {
	topicMap map[string]*Topic
}

func newTopicList() TopicList {
	return TopicList{make(map[string]*Topic)}
}

func (this *TopicList) subscribe(name string, addr *net.UDPAddr) {
	topic, ok := this.topicMap[name]
	if !ok {
		topic = newTopic(name)
		this.topicMap[name] = topic
	}
	topic.addSubscriber(addr)
}

func (this *TopicList) publish(name string, addr *net.UDPAddr) {
	topic, ok := this.topicMap[name]
	if !ok {
		topic = newTopic(name)
		this.topicMap[name] = topic
	}
	topic.addPublisher(addr)
}

func (this *TopicList) unsubscribe(name string, addr *net.UDPAddr) {
	topic, ok := this.topicMap[name]
	if ok {
		topic.delSubscriber(addr)
	}
}

func (this *TopicList) unpublish(name string, addr *net.UDPAddr) {
	topic, ok := this.topicMap[name]
	if ok {
		topic.delPublisher(addr)
	}
}

func (this *TopicList) push(msg Message, from *net.UDPAddr, trans *transceiver) {
	//fmt.Println("push msg:", msg)
	if this.topicMap == nil {
		this.topicMap = make(map[string]*Topic)
	}
	switch msg.Name {
	case "pub":
		this.publish(msg.Topic, from)
	case "sub":
		this.subscribe(msg.Topic, from)
	case "rem":
		this.unsubscribe(msg.Topic, from)
		this.unpublish(msg.Topic, from)
	case "msg":
		topic, ok := this.topicMap[msg.Topic]
		if ok {
			trans.sendToAll(msg, topic.SubscriberList)
		}
	}
}

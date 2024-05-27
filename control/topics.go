package control

type TopicList struct {
	TopicMap map[string]*Topic
}

func newTopicList() TopicList {
	return TopicList{make(map[string]*Topic)}
}

func (this *TopicList) subscribe(name string, addr string) {
	topic, ok := this.TopicMap[name]
	if !ok {
		topic = newTopic(name)
		this.TopicMap[name] = topic
	}
	topic.addSubscriber(addr)
}

func (this *TopicList) publish(name string, addr string) {
	topic, ok := this.TopicMap[name]
	if !ok {
		topic = newTopic(name)
		this.TopicMap[name] = topic
	}
	topic.addPublisher(addr)
}

func (this *TopicList) unsubscribe(name string, addr string) {
	topic, ok := this.TopicMap[name]
	if ok {
		topic.delSubscriber(addr)
	}
}

func (this *TopicList) unpublish(name string, addr string) {
	topic, ok := this.TopicMap[name]
	if ok {
		topic.delPublisher(addr)
	}
}

/*
func (this *TopicList) push(msg Message, from string, trans *transceiver) {
	//fmt.Println("push msg:", msg)
	if this.TopicMap == nil {
		this.TopicMap = make(map[string]*Topic)
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
		topic, ok := this.TopicMap[msg.Topic]
		if ok {
			//trans.sendToAll(msg, topic.SubscriberList)
		}
	}
}
*/

package control

type TopicList struct {
	TopicMap map[string]*Topic
}

// subscribe to topic
func (this *TopicList) subscribe(name string, addr string) {
	topic, ok := this.TopicMap[name]
	if !ok {
		topic = newTopic(name)
		this.TopicMap[name] = topic
	}
	topic.addSubscriber(addr)
}

// publish topic
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

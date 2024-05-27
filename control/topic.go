package control

type Topic struct {
	Name           string
	SubscriberList map[string]string // registered subscribers
	PublisherList  map[string]string // registered publishers
}

func newTopic(name string) *Topic {
	t := new(Topic)
	t.Name = name
	t.SubscriberList = make(map[string]string)
	t.PublisherList = make(map[string]string)
	return t
}

func (t *Topic) addSubscriber(ip string) {
	t.SubscriberList[ip] = ip
}

func (t *Topic) delSubscriber(ip string) {
	delete(t.SubscriberList, ip)
}

func (t *Topic) addPublisher(ip string) {
	t.PublisherList[ip] = ip
}

func (t *Topic) delPublisher(ip string) {
	delete(t.PublisherList, ip)
}

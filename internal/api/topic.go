package api

import "net"

type Topic struct {
	Name           string
	SubscriberList map[string]*net.UDPAddr // registered subscribers
	PublisherList  map[string]*net.UDPAddr // registered publishers
}

func newTopic(name string) *Topic {
	t := new(Topic)
	t.Name = name
	t.SubscriberList = make(map[string]*net.UDPAddr)
	t.PublisherList = make(map[string]*net.UDPAddr)
	return t
}

func (t *Topic) addSubscriber(ip *net.UDPAddr) {
	t.SubscriberList[ip.String()] = ip
}

func (t *Topic) delSubscriber(ip *net.UDPAddr) {
	delete(t.SubscriberList, ip.String())
}

func (t *Topic) addPublisher(ip *net.UDPAddr) {
	t.PublisherList[ip.String()] = ip
}

func (t *Topic) delPublisher(ip *net.UDPAddr) {
	delete(t.PublisherList, ip.String())
}

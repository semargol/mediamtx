package api

import (
	"encoding/json"
	"net"
	"time"
)

type transceiver struct {
	endPoint *net.UDPAddr
	udpConn  *net.UDPConn
}

func (t *transceiver) open(ep string) {
	t.endPoint, _ = net.ResolveUDPAddr("udp", ep)
	t.udpConn, _ = net.ListenUDP("udp", t.endPoint)
}

func (t *transceiver) sendTo(msg Message, addr *net.UDPAddr) {
	//fmt.Println("send to ", addr, "msg", msg)
	data, _ := json.Marshal(msg)
	_, _ = t.udpConn.WriteToUDP(data, addr)
}

func (t *transceiver) sendToAll(msg Message, subscriberList map[string]*net.UDPAddr) {
	data, _ := json.Marshal(msg)
	for _, addr := range subscriberList {
		//fmt.Println("send to ", addr, "msg", msg)
		_, _ = t.udpConn.WriteToUDP(data, addr)
	}
}

func (t *transceiver) receiveFrom(msec int) (Message, *net.UDPAddr, error) {
	var msg Message
	buf := make([]byte, 1024)
	deadline := time.Now().Add(time.Duration(msec) * time.Millisecond)
	_ = t.udpConn.SetReadDeadline(deadline)
	n, addr, err := t.udpConn.ReadFromUDP(buf)
	if err != nil {
		return msg, addr, err
	} else {
		err = json.Unmarshal(buf[:n], &msg)
		//fmt.Println("read from ", addr, "msg", msg)
		return msg, addr, err
	}
}

func (t *transceiver) SubscribeAt(topic string, at *net.UDPAddr) {
	var sub Message = Message{0, "sub", topic, "", "", make(map[string]string), nil}
	t.sendTo(sub, at)
}

func (t *transceiver) PublishAt(topic string, at *net.UDPAddr) {
	var pub Message = Message{0, "pub", topic, "", "", make(map[string]string), nil}
	t.sendTo(pub, at)
}

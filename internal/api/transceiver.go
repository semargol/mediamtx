package api

import (
	"encoding/json"
	"net"
	"time"
)

type Transceiver struct {
	EndPoint *net.UDPAddr
	UdpConn  *net.UDPConn
}

func (t *Transceiver) Open(ep string) {
	t.EndPoint, _ = net.ResolveUDPAddr("udp", ep)
	t.UdpConn, _ = net.ListenUDP("udp", t.EndPoint)
}

func (t *Transceiver) SendTo(msg Message, addr *net.UDPAddr) {
	//fmt.Println("send to ", addr, "msg", msg)
	data, _ := json.Marshal(msg)
	_, _ = t.UdpConn.WriteToUDP(data, addr)
}

func (t *Transceiver) SendToAll(msg Message, subscriberList map[string]*net.UDPAddr) {
	data, _ := json.Marshal(msg)
	for _, addr := range subscriberList {
		//fmt.Println("send to ", addr, "msg", msg)
		_, _ = t.UdpConn.WriteToUDP(data, addr)
	}
}

func (t *Transceiver) ReceiveFrom(msec int) (Message, *net.UDPAddr, error) {
	var msg Message
	buf := make([]byte, 1024)
	deadline := time.Now().Add(time.Duration(msec) * time.Millisecond)
	_ = t.UdpConn.SetReadDeadline(deadline)
	n, addr, err := t.UdpConn.ReadFromUDP(buf)
	if err != nil {
		return msg, addr, err
	} else {
		err = json.Unmarshal(buf[:n], &msg)
		//fmt.Println("read from ", addr, "msg", msg)
		return msg, addr, err
	}
}

func (t *Transceiver) SubscribeAt(topic string, at *net.UDPAddr) {
	var sub Message = Message{0, "sub", topic, "", "", make(map[string]string), nil}
	t.SendTo(sub, at)
}

func (t *Transceiver) PublishAt(topic string, at *net.UDPAddr) {
	var pub Message = Message{0, "pub", topic, "", "", make(map[string]string), nil}
	t.SendTo(pub, at)
}

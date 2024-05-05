package api

import (
	"encoding/json"
	"strings"
)

type Message struct {
	Name  string            // msg, pub[lish], sub[scribe], rem[ove]
	Topic string            // req[est], res[ponse]
	Verb  string            // add, set, get, del
	Noun  string            // pipe, rtp, rtsp
	Data  map[string]string // port=7777, mode=on|off
}

func NewMessage(text string) Message {
	msg := Message{}
	msg.Parse(text)
	return msg
}

func (msg *Message) Parse(text string) {
	var verb string
	var noun string
	var pair map[string]string = make(map[string]string)
	for {
		if strings.Contains(text, " =") {
			text = strings.Replace(text, " =", "=", -1)
		} else if strings.Contains(text, "= ") {
			text = strings.Replace(text, "= ", "=", -1)
		} else {
			break
		}
	}
	words := strings.Split(text, " ")
	msg.Name = "msg"
	if len(words) > 0 {
		verb = words[0]
		msg.Verb = verb
	}
	if len(words) > 1 {
		noun = words[1]
		msg.Noun = noun
		msg.Topic = verb + "/" + noun
	}
	var n int
	n = 2
	for {
		if n < len(words) {
			key, value, _ := strings.Cut(words[n], "=")
			pair[key] = value
			n = n + 1
		} else {
			break
		}
	}
	msg.Data = pair
}

func (msg *Message) String() string {
	b, _ := json.Marshal(msg)
	return string(b)
}

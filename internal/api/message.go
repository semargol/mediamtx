package api

import (
	"encoding/json"
	"github.com/bluenviron/mediamtx/internal/conf"
	"strings"
)

type Message struct {
	Corr  int               // correlation number
	Name  string            // msg, pub[lish], sub[scribe], rem[ove]
	Topic string            // req[est], res[ponse]
	Verb  string            // add, set, get, del
	Noun  string            // pipe, rtp, rtsp
	Data  map[string]string // port=7777, mode=on|off
	Conf  *conf.StrmConf
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
		if strings.Contains(text, "  ") {
			text = strings.Replace(text, "  ", " ", -1)
		} else if strings.Contains(text, " =") {
			text = strings.Replace(text, " =", "=", -1)
		} else if strings.Contains(text, "= ") {
			text = strings.Replace(text, "= ", "=", -1)
		} else {
			break
		}
	}
	words := strings.Split(text, " ")
	msg.Name = "msg"
	var n int
	n = 0
	if n < len(words) {
		if !strings.Contains(words[n], "=") {
			verb = words[n]
			msg.Verb = verb
			n++
		} else {
			msg.Verb = ""
		}
	}
	if n < len(words) {
		if !strings.Contains(words[n], "=") {
			noun = words[n]
			msg.Noun = noun
			n++
		} else {
			msg.Noun = ""
		}
	}
	msg.Topic = verb + "/" + noun
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

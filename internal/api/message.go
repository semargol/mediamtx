package api

import (
	"encoding/json"
	"fmt"
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

func (msg *Message) Sprint() string {
	//b, _ := json.Marshal(msg)
	text := fmt.Sprintf("%4d: %3s %-4s %-7s", msg.Corr, msg.Topic, msg.Verb, msg.Noun)
	for k, v := range msg.Data {
		text += fmt.Sprintf(" %s=%s", k, v)
	}
	return text
}

func (msg *Message) SprintCfg() map[int]string {
	var txt = make(map[int]string)
	//txt[len(txt)] = msg.Sprint()
	c := msg.Conf
	if c != nil {
		txt[len(txt)] = fmt.Sprintf("RTSP  addr=%s state=%s", c.RTSPSRV.Address, c.RTSPSRV.State)
		for _, p := range c.Pipes {
			r := p.RTPR
			x := p.RTSPCL
			s := p.RTPS
			txt[len(txt)] = fmt.Sprintf("PIPE  id=%d name=%s type=%s state=%s source=%s", p.ID, p.Name, p.Type, p.State, p.Source)
			txt[len(txt)] = fmt.Sprintf("       RTP-R   name=%s ror=%s video=%s,%d,%d audio=%s,%d,%d", r.Name, r.RunOnReady, r.VideoCodec, r.VideoPT, r.VideoPort, r.AudioCodec, r.AudioPT, r.AudioPort)
			txt[len(txt)] = fmt.Sprintf("       RTP-S   name=%s ror=%s video=%s,%s,%s audio=%s,%s,%s", s.Name, " ", s.VideoCodec, "PT", s.VideoURL, "opus", "PT", s.AudioURL)
			txt[len(txt)] = fmt.Sprintf("       RTSP-CL url=%s", x.Url)
		}
	}
	return txt
}

func (msg *Message) SprintAll() map[int]string {
	var txt = make(map[int]string)
	txt[len(txt)] = msg.Sprint()
	c := msg.Conf
	if c != nil {
		txt[len(txt)] = fmt.Sprintf("RTSP  addr=%s state=%s", c.RTSPSRV.Address, c.RTSPSRV.State)
		for _, p := range c.Pipes {
			r := p.RTPR
			x := p.RTSPCL
			s := p.RTPS
			txt[len(txt)] = fmt.Sprintf("PIPE  id=%d name=%s type=%s state=%s source=%s", p.ID, p.Name, p.Type, p.State, p.Source)
			txt[len(txt)] = fmt.Sprintf("       RTP-R   name=%s ror=%s video=%s,%d,%d audio=%s,%d,%d", r.Name, r.RunOnReady, r.VideoCodec, r.VideoPT, r.VideoPort, r.AudioCodec, r.AudioPT, r.AudioPort)
			txt[len(txt)] = fmt.Sprintf("       RTP-S   name=%s ror=%s video=%s,%s,%s audio=%s,%s,%sn", s.Name, " ", s.VideoCodec, "PT", s.VideoURL, "opus", "PT", s.AudioURL)
			txt[len(txt)] = fmt.Sprintf("       RTSP-CL url=%s", x.Url)
		}
	}
	return txt
}

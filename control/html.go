package control

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/bluenviron/mediamtx/internal/api"
	"github.com/gorilla/websocket"
)

func RunHtmlReader() {
	//var msg Message
	for {
		mt, buf, err := controlConnection.ReadMessage()
		if err != nil || mt != websocket.TextMessage {
			log.Println("controlBrokerReadError:", err)
			//CloseControlConnection()  will be closed in strm
			//controlConnection = nil
			break
		}

		//var msg *api.Message = new(api.Message)
		//msg.Parse(string(buf))
		//fmt.Println("html got ", msg)

		text := string(buf)
		if len(text) > 0 {
			cmdNumber++
			fmt.Print(cmdNumber, ">", text)
			fmt.Println()
			msg := new(api.Message)
			msg.Parse(text)
			msg.Corr = cmdNumber
			msg.Topic = "req"
			//fmt.Println(msg) //%s, type: %d", message, mt)
			fromControl <- msg
			_ = controlConnection.WriteMessage(websocket.TextMessage, []byte(strconv.Itoa(cmdNumber)+">"+text))
			//time.Sleep(50 * time.Millisecond)
		}
		/*
			bytes, merr := json.Marshal(msg)
			if merr != nil {
				continue
			}
			uerr := json.Unmarshal(bytes, msg)
			if uerr != nil {
				continue
			}
		*/

		//msg.Corr = cmdNumber
		//cmdNumber++
		//msg.Topic = "req"

		//fmt.Println("BROK BrokerControlReader: ", msg) //%s, type: %d", message, mt)
		//fromControl <- msg
		//serverBroker.topicList.push(msg, from, &serverBroker.transceiver)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/cihtml")
}

func strmhtml(w http.ResponseWriter, r *http.Request) {
	if controlConnection != nil {
		log.Print("Only one control connection allowed")
		return
	}
	var err error = nil
	controlConnection, err = controlBrokerUpgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("controlBrokerUpgraderError:", err)
		return
	}
	defer func() {
		err := controlConnection.Close()
		if err != nil {
		}
		controlConnection = nil
	}()
	inputJson = false
	topicList.publish("req", "control")
	topicList.subscribe("res", "control")
	topicList.subscribe("evn", "control")
	topicList.subscribe("cfg", "control")
	RunHtmlReader()
}

func response(msg *api.Message) {
	if msg.Topic == "res" {
		response_or_event(msg)
	}
}

func event(msg *api.Message) {
	if msg.Topic == "evn" {
		response_or_event(msg)
	}
}

func response_or_event(msg *api.Message) {
	if msg.Data["result"] != "" {
		//fmt.Println(msg.Corr, " ", msg)
		rs := fmt.Sprintf("     %s %s %s", msg.Data["result"], msg.Data["errorMsg"], msg.Data["description"])
		_ = controlConnection.WriteMessage(websocket.TextMessage, []byte(rs))
	} else {
		rs := fmt.Sprintf("     ok  %s", msg.Data["status"])
		rp := ""
		for key, val := range msg.Data {
			if key != "status" && val != "" {
				rp += fmt.Sprintf(" %s=%s", key, val)
			}
		}
		_ = controlConnection.WriteMessage(websocket.TextMessage, []byte(rs))
		_ = controlConnection.WriteMessage(websocket.TextMessage, []byte(rp))
	}
	c := msg.Conf
	if c != nil {
		rs := fmt.Sprintf("     RTSPSRV:  addr=%s state=%s\n", c.RTSPSRV.Address, c.RTSPSRV.State)
		_ = controlConnection.WriteMessage(websocket.TextMessage, []byte(rs))
		for _, p := range c.Pipes {
			r := p.RTPR
			pp := fmt.Sprintf("     PIPE:     id=%d name=%s type=%s state=%s source=%s\n", p.ID, p.Name, p.Type, p.State, p.Source)
			rt := fmt.Sprintf("         RTPR: name=%s ror=%s video=%s,%d,%s audio=%s,%d,%s\n", r.Name, r.RunOnReady, r.VideoCodec, r.VideoPT, r.VideoURL, r.AudioCodec, r.AudioPT, r.AudioURL)
			_ = controlConnection.WriteMessage(websocket.TextMessage, []byte(pp))
			_ = controlConnection.WriteMessage(websocket.TextMessage, []byte(rt))
		}
	}
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>  
window.addEventListener("load", function(evt) {

    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;

    var print = function(message) {
        var d = document.createElement("div");
        d.textContent = message;
        output.appendChild(d);
        output.scroll(0, output.scrollHeight);
    };

    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print(evt.data);
            /*print("RESPONSE: " + evt.data);*/
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };

    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        /*print("SEND: " + input.value);*/
        ws.send(input.value);
        return false;
    };

    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };

});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server, 
"Send" to send a message to the server and "Close" to close the connection. 
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output" style="max-height: 70vh;overflow-y: scroll;"></div>
</td></tr></table>
</body>
</html>
`))

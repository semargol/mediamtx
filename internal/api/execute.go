package api

import "encoding/json"

// type Message struct {		// examples
//		Name  string            // msg, pub (publish), sub (subscribe), rem (remove = unpublish & unsubscribe)
//		Topic string            // req (request), res (response)
//		Verb  string            // add, set, get, del
//		Noun  string            // pipe, rtp, rtsp
//		Data  map[string]string // port=7777, mode=on|off, path=AV_2, sdp=1.sdp
// }

func globals(api *API, response *Message) {
	response.Data["rtcp"] = string(api.Conf.RTCPAddress)
	response.Data["rtp"] = string(api.Conf.RTPAddress)
	response.Data["rtsp"] = string(api.Conf.RTSPAddress)
	//response.Data["paths"] = string(json.Marshal(api.Conf.Paths))
	b, _ := api.Conf.ReadTimeout.MarshalJSON()
	var s string
	_ = json.Unmarshal(b, &s)
	response.Data["timeout"] = s
	response.Data["result"] = "OK"
}

func ApiAddPipe(api *API, req *Message) (Message, error) {
	id := req.Data["id"]
	response := Message{"msg", "res", req.Verb, req.Noun, make(map[string]string)}
	response.Data["result"] = "OK"
	response.Data["id"] = id
	return response, nil
}

func ApiAddRtp(api *API, req *Message) (Message, error) {
	id := req.Data["id"]
	response := Message{"msg", "res", req.Verb, req.Noun, make(map[string]string)}
	response.Data["result"] = "OK"
	response.Data["id"] = id
	return response, nil
}

func ApiAddRtsp(api *API, req *Message) (Message, error) {
	id := req.Data["id"]
	response := Message{"msg", "res", req.Verb, req.Noun, make(map[string]string)}
	response.Data["result"] = "OK"
	response.Data["id"] = id
	return response, nil
}

func ApiDelPipe(api *API, req *Message) (Message, error) {
	id := req.Data["id"]
	response := Message{"msg", "res", req.Verb, req.Noun, make(map[string]string)}
	response.Data["result"] = "OK"
	response.Data["id"] = id
	return response, nil
}

func ApiDelRtp(api *API, req *Message) (Message, error) {
	id := req.Data["id"]
	response := Message{"msg", "res", req.Verb, req.Noun, make(map[string]string)}
	response.Data["result"] = "OK"
	response.Data["id"] = id
	return response, nil
}

func ApiDelRtsp(api *API, req *Message) (Message, error) {
	id := req.Data["id"]
	response := Message{"msg", "res", req.Verb, req.Noun, make(map[string]string)}
	response.Data["result"] = "OK"
	response.Data["id"] = id
	return response, nil
}

func ApiSetPipe(api *API, req *Message) (Message, error) {
	response := Message{"msg", "res", req.Verb, req.Noun, make(map[string]string)}
	response.Data["result"] = "OK"
	return response, nil
}

func ApiSetRtp(api *API, req *Message) (Message, error) {
	response := Message{"msg", "res", req.Verb, req.Noun, make(map[string]string)}
	response.Data["result"] = "OK"
	return response, nil
}

func ApiSetRtsp(api *API, req *Message) (Message, error) {
	response := Message{"msg", "res", req.Verb, req.Noun, make(map[string]string)}
	response.Data["result"] = "OK"
	return response, nil
}

func ApiGetPipe(api *API, req *Message) (Message, error) {
	response := Message{"msg", "res", req.Verb, req.Noun, make(map[string]string)}
	globals(api, &response)
	return response, nil
}

func ApiGetRtp(api *API, req *Message) (Message, error) {
	response := Message{"msg", "res", req.Verb, req.Noun, make(map[string]string)}
	globals(api, &response)
	return response, nil
}

func ApiGetRtsp(api *API, req *Message) (Message, error) {
	response := Message{"msg", "res", req.Verb, req.Noun, make(map[string]string)}
	globals(api, &response)
	return response, nil
}

/*
func ApiSetRtp(api *api.API, req *Message) (Message, error) {
	id := req.Data["id"]
	path := req.Data["path"]
	port := req.Data["port"]
	file := req.Data["file"]
	response := Message{"msg", "res", req.Verb, req.Noun, make(map[string]string)}
	response.Data["result"] = "OK"
	return response, nil
}
*/

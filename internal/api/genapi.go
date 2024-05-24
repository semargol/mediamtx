package api

import (
    "fmt"
    "strconv"
    "strings"
)


func onReceive_requestChannel(request *Message, response *Message) {
    cmd := strings.ToLower(request.Verb + request.Noun)
    switch cmd {
        case "addpipe":
            onRequest_addpipe(request, response)
            break
        case "delpipe":
            onRequest_delpipe(request, response)
            break
        case "getpipe":
            onRequest_getpipe(request, response)
            break
        case "getrtpr":
            onRequest_getrtpr(request, response)
            break
        case "getrtps":
            onRequest_getrtps(request, response)
            break
        case "getrtspcl":
            onRequest_getrtspcl(request, response)
            break
        case "getrtspsrv":
            onRequest_getrtspsrv(request, response)
            break
        case "setpipe":
            onRequest_setpipe(request, response)
            break
        case "setrtpr":
            onRequest_setrtpr(request, response)
            break
        case "setrtps":
            onRequest_setrtps(request, response)
            break
        case "setrtspcl":
            onRequest_setrtspcl(request, response)
            break
        case "setrtspsrv":
            onRequest_setrtspsrv(request, response)
            break
        case "sub":
            break
        case "pub":
            break
        default:
            fmt.Println("message", "\""+request.Verb+request.Noun+"\"", "not defined")
            break
    }
}
var allowed_addpipe map[string]string

func init() {
    allowed_addpipe = make(map[string]string)
    allowed_addpipe["dir"] = "_genyml.ApiListAttribute_"
    allowed_addpipe["id"] = "_genyml.ApiRangeAttribute_"
}

func onRequest_addpipe_allkeys() string {
    return "dir,id,noun,verb"
}

func onRequest_addpipe_allowed(key string, val string) (keyok bool, valok bool, vlist string) {
    switch key {
    case "dir":
          vlist = "sender|reader"
          if val == "sender" { return true, true, vlist }
          if val == "reader" { return true, true, vlist }
          return true, false, vlist
    case "id":
          vlist = "1..4"
          ival, eval := strconv.Atoi(val)
          if eval != nil { return true, false, vlist }
          if ival >= 1 && ival <= 4 { return true, true, vlist }
          return true, false, vlist
    case "noun":
    case "verb":
        default:
            vlist = "?"
            return false, false, vlist
    }
    vlist = "?"
    return false, false, vlist
}

func onRequest_addpipe(request *Message, response *Message) {
    klist := onRequest_addpipe_allkeys()
    for rkey, rval := range request.Data {
        keyok, valok, vlist := onRequest_addpipe_allowed(rkey, rval)
        if !keyok {
            fmt.Println("message addpipe key not allowed", rkey, "not in", klist)
        } else if !valok {
            fmt.Println("message addpipe key value not allowed", rkey, "=",vlist)
        }
    }
    for _, akey := range strings.Split(klist, ",") {
        if akey == "name" { continue }
        if akey == "noun" { continue }
        if akey == "verb" { continue }
        rval, ok := request.Data[akey]
        if !ok || rval == "" {
            _, _, vlist := onRequest_addpipe_allowed(akey, rval)
            //if rval != "" {
                fmt.Println("message addpipe key missed", akey, "=", vlist)
            //}
        }
    }
}
var allowed_delpipe map[string]string

func init() {
    allowed_delpipe = make(map[string]string)
    allowed_delpipe["id"] = "_genyml.ApiRangeAttribute_"
}

func onRequest_delpipe_allkeys() string {
    return "id,noun,verb"
}

func onRequest_delpipe_allowed(key string, val string) (keyok bool, valok bool, vlist string) {
    switch key {
    case "id":
          vlist = "1..4"
          ival, eval := strconv.Atoi(val)
          if eval != nil { return true, false, vlist }
          if ival >= 1 && ival <= 4 { return true, true, vlist }
          return true, false, vlist
    case "noun":
    case "verb":
        default:
            vlist = "?"
            return false, false, vlist
    }
    vlist = "?"
    return false, false, vlist
}

func onRequest_delpipe(request *Message, response *Message) {
    klist := onRequest_delpipe_allkeys()
    for rkey, rval := range request.Data {
        keyok, valok, vlist := onRequest_delpipe_allowed(rkey, rval)
        if !keyok {
            fmt.Println("message delpipe key not allowed", rkey, "not in", klist)
        } else if !valok {
            fmt.Println("message delpipe key value not allowed", rkey, "=",vlist)
        }
    }
    for _, akey := range strings.Split(klist, ",") {
        if akey == "name" { continue }
        if akey == "noun" { continue }
        if akey == "verb" { continue }
        rval, ok := request.Data[akey]
        if !ok || rval == "" {
            _, _, vlist := onRequest_delpipe_allowed(akey, rval)
            //if rval != "" {
                fmt.Println("message delpipe key missed", akey, "=", vlist)
            //}
        }
    }
}
var allowed_getpipe map[string]string

func init() {
    allowed_getpipe = make(map[string]string)
    allowed_getpipe["dir"] = "_genyml.ApiSimpleAttribute_"
    allowed_getpipe["id"] = "_genyml.ApiRangeAttribute_"
    allowed_getpipe["sink"] = "_genyml.ApiSimpleAttribute_"
    allowed_getpipe["source"] = "_genyml.ApiSimpleAttribute_"
    allowed_getpipe["state"] = "_genyml.ApiSimpleAttribute_"
}

func onRequest_getpipe_allkeys() string {
    return "dir,id,noun,sink,source,state,verb"
}

func onRequest_getpipe_allowed(key string, val string) (keyok bool, valok bool, vlist string) {
    switch key {
    case "dir":
          vlist = ""
          if val == "" { return true, true, vlist }
          return true, false, vlist
    case "id":
          vlist = "1..4"
          ival, eval := strconv.Atoi(val)
          if eval != nil { return true, false, vlist }
          if ival >= 1 && ival <= 4 { return true, true, vlist }
          return true, false, vlist
    case "noun":
    case "sink":
          vlist = ""
          if val == "" { return true, true, vlist }
          return true, false, vlist
    case "source":
          vlist = ""
          if val == "" { return true, true, vlist }
          return true, false, vlist
    case "state":
          vlist = ""
          if val == "" { return true, true, vlist }
          return true, false, vlist
    case "verb":
        default:
            vlist = "?"
            return false, false, vlist
    }
    vlist = "?"
    return false, false, vlist
}

func onRequest_getpipe(request *Message, response *Message) {
    klist := onRequest_getpipe_allkeys()
    for rkey, rval := range request.Data {
        keyok, valok, vlist := onRequest_getpipe_allowed(rkey, rval)
        if !keyok {
            fmt.Println("message getpipe key not allowed", rkey, "not in", klist)
        } else if !valok {
            fmt.Println("message getpipe key value not allowed", rkey, "=",vlist)
        }
    }
    for _, akey := range strings.Split(klist, ",") {
        if akey == "name" { continue }
        if akey == "noun" { continue }
        if akey == "verb" { continue }
        rval, ok := request.Data[akey]
        if !ok || rval == "" {
            _, _, vlist := onRequest_getpipe_allowed(akey, rval)
            //if rval != "" {
                fmt.Println("message getpipe key missed", akey, "=", vlist)
            //}
        }
    }
}
var allowed_getrtpr map[string]string

func init() {
    allowed_getrtpr = make(map[string]string)
    allowed_getrtpr["acodec"] = "_genyml.ApiSimpleAttribute_"
    allowed_getrtpr["id"] = "_genyml.ApiRangeAttribute_"
    allowed_getrtpr["port"] = "_genyml.ApiSimpleAttribute_"
    allowed_getrtpr["pt"] = "_genyml.ApiSimpleAttribute_"
    allowed_getrtpr["vcodec"] = "_genyml.ApiSimpleAttribute_"
}

func onRequest_getrtpr_allkeys() string {
    return "acodec,id,noun,port,pt,vcodec,verb"
}

func onRequest_getrtpr_allowed(key string, val string) (keyok bool, valok bool, vlist string) {
    switch key {
    case "acodec":
          vlist = ""
          if val == "" { return true, true, vlist }
          return true, false, vlist
    case "id":
          vlist = "1..4"
          ival, eval := strconv.Atoi(val)
          if eval != nil { return true, false, vlist }
          if ival >= 1 && ival <= 4 { return true, true, vlist }
          return true, false, vlist
    case "noun":
    case "port":
          vlist = ""
          if val == "" { return true, true, vlist }
          return true, false, vlist
    case "pt":
          vlist = ""
          if val == "" { return true, true, vlist }
          return true, false, vlist
    case "vcodec":
          vlist = ""
          if val == "" { return true, true, vlist }
          return true, false, vlist
    case "verb":
        default:
            vlist = "?"
            return false, false, vlist
    }
    vlist = "?"
    return false, false, vlist
}

func onRequest_getrtpr(request *Message, response *Message) {
    klist := onRequest_getrtpr_allkeys()
    for rkey, rval := range request.Data {
        keyok, valok, vlist := onRequest_getrtpr_allowed(rkey, rval)
        if !keyok {
            fmt.Println("message getrtpr key not allowed", rkey, "not in", klist)
        } else if !valok {
            fmt.Println("message getrtpr key value not allowed", rkey, "=",vlist)
        }
    }
    for _, akey := range strings.Split(klist, ",") {
        if akey == "name" { continue }
        if akey == "noun" { continue }
        if akey == "verb" { continue }
        rval, ok := request.Data[akey]
        if !ok || rval == "" {
            _, _, vlist := onRequest_getrtpr_allowed(akey, rval)
            //if rval != "" {
                fmt.Println("message getrtpr key missed", akey, "=", vlist)
            //}
        }
    }
}
var allowed_getrtps map[string]string

func init() {
    allowed_getrtps = make(map[string]string)
    allowed_getrtps["acodec"] = "_genyml.ApiSimpleAttribute_"
    allowed_getrtps["id"] = "_genyml.ApiRangeAttribute_"
    allowed_getrtps["port"] = "_genyml.ApiSimpleAttribute_"
    allowed_getrtps["pt"] = "_genyml.ApiSimpleAttribute_"
    allowed_getrtps["vcodec"] = "_genyml.ApiSimpleAttribute_"
}

func onRequest_getrtps_allkeys() string {
    return "acodec,id,noun,port,pt,vcodec,verb"
}

func onRequest_getrtps_allowed(key string, val string) (keyok bool, valok bool, vlist string) {
    switch key {
    case "acodec":
          vlist = ""
          if val == "" { return true, true, vlist }
          return true, false, vlist
    case "id":
          vlist = "1..4"
          ival, eval := strconv.Atoi(val)
          if eval != nil { return true, false, vlist }
          if ival >= 1 && ival <= 4 { return true, true, vlist }
          return true, false, vlist
    case "noun":
    case "port":
          vlist = ""
          if val == "" { return true, true, vlist }
          return true, false, vlist
    case "pt":
          vlist = ""
          if val == "" { return true, true, vlist }
          return true, false, vlist
    case "vcodec":
          vlist = ""
          if val == "" { return true, true, vlist }
          return true, false, vlist
    case "verb":
        default:
            vlist = "?"
            return false, false, vlist
    }
    vlist = "?"
    return false, false, vlist
}

func onRequest_getrtps(request *Message, response *Message) {
    klist := onRequest_getrtps_allkeys()
    for rkey, rval := range request.Data {
        keyok, valok, vlist := onRequest_getrtps_allowed(rkey, rval)
        if !keyok {
            fmt.Println("message getrtps key not allowed", rkey, "not in", klist)
        } else if !valok {
            fmt.Println("message getrtps key value not allowed", rkey, "=",vlist)
        }
    }
    for _, akey := range strings.Split(klist, ",") {
        if akey == "name" { continue }
        if akey == "noun" { continue }
        if akey == "verb" { continue }
        rval, ok := request.Data[akey]
        if !ok || rval == "" {
            _, _, vlist := onRequest_getrtps_allowed(akey, rval)
            //if rval != "" {
                fmt.Println("message getrtps key missed", akey, "=", vlist)
            //}
        }
    }
}
var allowed_getrtspcl map[string]string

func init() {
    allowed_getrtspcl = make(map[string]string)
    allowed_getrtspcl["id"] = "_genyml.ApiRangeAttribute_"
    allowed_getrtspcl["url"] = "_genyml.ApiSimpleAttribute_"
}

func onRequest_getrtspcl_allkeys() string {
    return "id,noun,url,verb"
}

func onRequest_getrtspcl_allowed(key string, val string) (keyok bool, valok bool, vlist string) {
    switch key {
    case "id":
          vlist = "1..4"
          ival, eval := strconv.Atoi(val)
          if eval != nil { return true, false, vlist }
          if ival >= 1 && ival <= 4 { return true, true, vlist }
          return true, false, vlist
    case "noun":
    case "url":
          vlist = ""
          if val == "" { return true, true, vlist }
          return true, false, vlist
    case "verb":
        default:
            vlist = "?"
            return false, false, vlist
    }
    vlist = "?"
    return false, false, vlist
}

func onRequest_getrtspcl(request *Message, response *Message) {
    klist := onRequest_getrtspcl_allkeys()
    for rkey, rval := range request.Data {
        keyok, valok, vlist := onRequest_getrtspcl_allowed(rkey, rval)
        if !keyok {
            fmt.Println("message getrtspcl key not allowed", rkey, "not in", klist)
        } else if !valok {
            fmt.Println("message getrtspcl key value not allowed", rkey, "=",vlist)
        }
    }
    for _, akey := range strings.Split(klist, ",") {
        if akey == "name" { continue }
        if akey == "noun" { continue }
        if akey == "verb" { continue }
        rval, ok := request.Data[akey]
        if !ok || rval == "" {
            _, _, vlist := onRequest_getrtspcl_allowed(akey, rval)
            //if rval != "" {
                fmt.Println("message getrtspcl key missed", akey, "=", vlist)
            //}
        }
    }
}
var allowed_getrtspsrv map[string]string

func init() {
    allowed_getrtspsrv = make(map[string]string)
    allowed_getrtspsrv["id"] = "_genyml.ApiRangeAttribute_"
    allowed_getrtspsrv["path"] = "_genyml.ApiSimpleAttribute_"
}

func onRequest_getrtspsrv_allkeys() string {
    return "id,noun,path,verb"
}

func onRequest_getrtspsrv_allowed(key string, val string) (keyok bool, valok bool, vlist string) {
    switch key {
    case "id":
          vlist = "1..4"
          ival, eval := strconv.Atoi(val)
          if eval != nil { return true, false, vlist }
          if ival >= 1 && ival <= 4 { return true, true, vlist }
          return true, false, vlist
    case "noun":
    case "path":
          vlist = ""
          if val == "" { return true, true, vlist }
          return true, false, vlist
    case "verb":
        default:
            vlist = "?"
            return false, false, vlist
    }
    vlist = "?"
    return false, false, vlist
}

func onRequest_getrtspsrv(request *Message, response *Message) {
    klist := onRequest_getrtspsrv_allkeys()
    for rkey, rval := range request.Data {
        keyok, valok, vlist := onRequest_getrtspsrv_allowed(rkey, rval)
        if !keyok {
            fmt.Println("message getrtspsrv key not allowed", rkey, "not in", klist)
        } else if !valok {
            fmt.Println("message getrtspsrv key value not allowed", rkey, "=",vlist)
        }
    }
    for _, akey := range strings.Split(klist, ",") {
        if akey == "name" { continue }
        if akey == "noun" { continue }
        if akey == "verb" { continue }
        rval, ok := request.Data[akey]
        if !ok || rval == "" {
            _, _, vlist := onRequest_getrtspsrv_allowed(akey, rval)
            //if rval != "" {
                fmt.Println("message getrtspsrv key missed", akey, "=", vlist)
            //}
        }
    }
}
var allowed_setpipe map[string]string

func init() {
    allowed_setpipe = make(map[string]string)
    allowed_setpipe["id"] = "_genyml.ApiRangeAttribute_"
    allowed_setpipe["state"] = "_genyml.ApiListAttribute_"
}

func onRequest_setpipe_allkeys() string {
    return "id,noun,state,verb"
}

func onRequest_setpipe_allowed(key string, val string) (keyok bool, valok bool, vlist string) {
    switch key {
    case "id":
          vlist = "1..4"
          ival, eval := strconv.Atoi(val)
          if eval != nil { return true, false, vlist }
          if ival >= 1 && ival <= 4 { return true, true, vlist }
          return true, false, vlist
    case "noun":
    case "state":
          vlist = "start|stop"
          if val == "start" { return true, true, vlist }
          if val == "stop" { return true, true, vlist }
          return true, false, vlist
    case "verb":
        default:
            vlist = "?"
            return false, false, vlist
    }
    vlist = "?"
    return false, false, vlist
}

func onRequest_setpipe(request *Message, response *Message) {
    klist := onRequest_setpipe_allkeys()
    for rkey, rval := range request.Data {
        keyok, valok, vlist := onRequest_setpipe_allowed(rkey, rval)
        if !keyok {
            fmt.Println("message setpipe key not allowed", rkey, "not in", klist)
        } else if !valok {
            fmt.Println("message setpipe key value not allowed", rkey, "=",vlist)
        }
    }
    for _, akey := range strings.Split(klist, ",") {
        if akey == "name" { continue }
        if akey == "noun" { continue }
        if akey == "verb" { continue }
        rval, ok := request.Data[akey]
        if !ok || rval == "" {
            _, _, vlist := onRequest_setpipe_allowed(akey, rval)
            //if rval != "" {
                fmt.Println("message setpipe key missed", akey, "=", vlist)
            //}
        }
    }
}
var allowed_setrtpr map[string]string

func init() {
    allowed_setrtpr = make(map[string]string)
    allowed_setrtpr["acodec"] = "_genyml.ApiSimpleAttribute_"
    allowed_setrtpr["id"] = "_genyml.ApiRangeAttribute_"
    allowed_setrtpr["port"] = "_genyml.ApiRangeAttribute_"
    allowed_setrtpr["pt"] = "_genyml.ApiRangeAttribute_"
    allowed_setrtpr["vcodec"] = "_genyml.ApiListAttribute_"
}

func onRequest_setrtpr_allkeys() string {
    return "acodec,id,noun,port,pt,vcodec,verb"
}

func onRequest_setrtpr_allowed(key string, val string) (keyok bool, valok bool, vlist string) {
    switch key {
    case "acodec":
          vlist = "opus"
          if val == "opus" { return true, true, vlist }
          return true, false, vlist
    case "id":
          vlist = "1..4"
          ival, eval := strconv.Atoi(val)
          if eval != nil { return true, false, vlist }
          if ival >= 1 && ival <= 4 { return true, true, vlist }
          return true, false, vlist
    case "noun":
    case "port":
          vlist = "5000..6000"
          ival, eval := strconv.Atoi(val)
          if eval != nil { return true, false, vlist }
          if ival >= 5000 && ival <= 6000 { return true, true, vlist }
          return true, false, vlist
    case "pt":
          vlist = "96..99"
          ival, eval := strconv.Atoi(val)
          if eval != nil { return true, false, vlist }
          if ival >= 96 && ival <= 99 { return true, true, vlist }
          return true, false, vlist
    case "vcodec":
          vlist = "h264|h265"
          if val == "h264" { return true, true, vlist }
          if val == "h265" { return true, true, vlist }
          return true, false, vlist
    case "verb":
        default:
            vlist = "?"
            return false, false, vlist
    }
    vlist = "?"
    return false, false, vlist
}

func onRequest_setrtpr(request *Message, response *Message) {
    klist := onRequest_setrtpr_allkeys()
    for rkey, rval := range request.Data {
        keyok, valok, vlist := onRequest_setrtpr_allowed(rkey, rval)
        if !keyok {
            fmt.Println("message setrtpr key not allowed", rkey, "not in", klist)
        } else if !valok {
            fmt.Println("message setrtpr key value not allowed", rkey, "=",vlist)
        }
    }
    for _, akey := range strings.Split(klist, ",") {
        if akey == "name" { continue }
        if akey == "noun" { continue }
        if akey == "verb" { continue }
        rval, ok := request.Data[akey]
        if !ok || rval == "" {
            _, _, vlist := onRequest_setrtpr_allowed(akey, rval)
            //if rval != "" {
                fmt.Println("message setrtpr key missed", akey, "=", vlist)
            //}
        }
    }
}
var allowed_setrtps map[string]string

func init() {
    allowed_setrtps = make(map[string]string)
    allowed_setrtps["acodec"] = "_genyml.ApiSimpleAttribute_"
    allowed_setrtps["id"] = "_genyml.ApiRangeAttribute_"
    allowed_setrtps["port"] = "_genyml.ApiRangeAttribute_"
    allowed_setrtps["pt"] = "_genyml.ApiRangeAttribute_"
    allowed_setrtps["vcodec"] = "_genyml.ApiListAttribute_"
}

func onRequest_setrtps_allkeys() string {
    return "acodec,id,noun,port,pt,vcodec,verb"
}

func onRequest_setrtps_allowed(key string, val string) (keyok bool, valok bool, vlist string) {
    switch key {
    case "acodec":
          vlist = "opus"
          if val == "opus" { return true, true, vlist }
          return true, false, vlist
    case "id":
          vlist = "1..4"
          ival, eval := strconv.Atoi(val)
          if eval != nil { return true, false, vlist }
          if ival >= 1 && ival <= 4 { return true, true, vlist }
          return true, false, vlist
    case "noun":
    case "port":
          vlist = "5000..6000"
          ival, eval := strconv.Atoi(val)
          if eval != nil { return true, false, vlist }
          if ival >= 5000 && ival <= 6000 { return true, true, vlist }
          return true, false, vlist
    case "pt":
          vlist = "96..99"
          ival, eval := strconv.Atoi(val)
          if eval != nil { return true, false, vlist }
          if ival >= 96 && ival <= 99 { return true, true, vlist }
          return true, false, vlist
    case "vcodec":
          vlist = "h264|h265"
          if val == "h264" { return true, true, vlist }
          if val == "h265" { return true, true, vlist }
          return true, false, vlist
    case "verb":
        default:
            vlist = "?"
            return false, false, vlist
    }
    vlist = "?"
    return false, false, vlist
}

func onRequest_setrtps(request *Message, response *Message) {
    klist := onRequest_setrtps_allkeys()
    for rkey, rval := range request.Data {
        keyok, valok, vlist := onRequest_setrtps_allowed(rkey, rval)
        if !keyok {
            fmt.Println("message setrtps key not allowed", rkey, "not in", klist)
        } else if !valok {
            fmt.Println("message setrtps key value not allowed", rkey, "=",vlist)
        }
    }
    for _, akey := range strings.Split(klist, ",") {
        if akey == "name" { continue }
        if akey == "noun" { continue }
        if akey == "verb" { continue }
        rval, ok := request.Data[akey]
        if !ok || rval == "" {
            _, _, vlist := onRequest_setrtps_allowed(akey, rval)
            //if rval != "" {
                fmt.Println("message setrtps key missed", akey, "=", vlist)
            //}
        }
    }
}
var allowed_setrtspcl map[string]string

func init() {
    allowed_setrtspcl = make(map[string]string)
    allowed_setrtspcl["id"] = "_genyml.ApiRangeAttribute_"
    allowed_setrtspcl["url"] = "_genyml.ApiSimpleAttribute_"
}

func onRequest_setrtspcl_allkeys() string {
    return "id,noun,url,verb"
}

func onRequest_setrtspcl_allowed(key string, val string) (keyok bool, valok bool, vlist string) {
    switch key {
    case "id":
          vlist = "1..4"
          ival, eval := strconv.Atoi(val)
          if eval != nil { return true, false, vlist }
          if ival >= 1 && ival <= 4 { return true, true, vlist }
          return true, false, vlist
    case "noun":
    case "url":
          vlist = ""
          if val == "" { return true, true, vlist }
          return true, false, vlist
    case "verb":
        default:
            vlist = "?"
            return false, false, vlist
    }
    vlist = "?"
    return false, false, vlist
}

func onRequest_setrtspcl(request *Message, response *Message) {
    klist := onRequest_setrtspcl_allkeys()
    for rkey, rval := range request.Data {
        keyok, valok, vlist := onRequest_setrtspcl_allowed(rkey, rval)
        if !keyok {
            fmt.Println("message setrtspcl key not allowed", rkey, "not in", klist)
        } else if !valok {
            fmt.Println("message setrtspcl key value not allowed", rkey, "=",vlist)
        }
    }
    for _, akey := range strings.Split(klist, ",") {
        if akey == "name" { continue }
        if akey == "noun" { continue }
        if akey == "verb" { continue }
        rval, ok := request.Data[akey]
        if !ok || rval == "" {
            _, _, vlist := onRequest_setrtspcl_allowed(akey, rval)
            //if rval != "" {
                fmt.Println("message setrtspcl key missed", akey, "=", vlist)
            //}
        }
    }
}
var allowed_setrtspsrv map[string]string

func init() {
    allowed_setrtspsrv = make(map[string]string)
    allowed_setrtspsrv["id"] = "_genyml.ApiRangeAttribute_"
    allowed_setrtspsrv["path"] = "_genyml.ApiSimpleAttribute_"
}

func onRequest_setrtspsrv_allkeys() string {
    return "id,noun,path,verb"
}

func onRequest_setrtspsrv_allowed(key string, val string) (keyok bool, valok bool, vlist string) {
    switch key {
    case "id":
          vlist = "1..4"
          ival, eval := strconv.Atoi(val)
          if eval != nil { return true, false, vlist }
          if ival >= 1 && ival <= 4 { return true, true, vlist }
          return true, false, vlist
    case "noun":
    case "path":
          vlist = ""
          if val == "" { return true, true, vlist }
          return true, false, vlist
    case "verb":
        default:
            vlist = "?"
            return false, false, vlist
    }
    vlist = "?"
    return false, false, vlist
}

func onRequest_setrtspsrv(request *Message, response *Message) {
    klist := onRequest_setrtspsrv_allkeys()
    for rkey, rval := range request.Data {
        keyok, valok, vlist := onRequest_setrtspsrv_allowed(rkey, rval)
        if !keyok {
            fmt.Println("message setrtspsrv key not allowed", rkey, "not in", klist)
        } else if !valok {
            fmt.Println("message setrtspsrv key value not allowed", rkey, "=",vlist)
        }
    }
    for _, akey := range strings.Split(klist, ",") {
        if akey == "name" { continue }
        if akey == "noun" { continue }
        if akey == "verb" { continue }
        rval, ok := request.Data[akey]
        if !ok || rval == "" {
            _, _, vlist := onRequest_setrtspsrv_allowed(akey, rval)
            //if rval != "" {
                fmt.Println("message setrtspsrv key missed", akey, "=", vlist)
            //}
        }
    }
}

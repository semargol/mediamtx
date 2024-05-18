package api

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/bluenviron/mediamtx/internal/conf"
)

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

func errorf(format string, a ...interface{}) error {
	err := fmt.Errorf(format, a...)
	fmt.Println("Error logged:", err)
	return err
}

func getError(req *Message, errorCode int) (Message, int) {
	response := Message{req.Corr, "msg", "res", req.Verb, req.Noun, make(map[string]string)}
	response.Data["result"] = strconv.Itoa(errorCode)
	response.Data["errorMsg"] = getErrorDescription(errorCode, true)
	return response, errorCode
}

func setField(p *conf.OptionalPath, fieldName, newValue string) {
	v := reflect.ValueOf(p.Values)

	// Dereference the pointer if it points to a structure
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	// Check if the provided value is valid
	if !v.IsValid() {
		fmt.Println("Provided value is invalid")
		return
	}

	// Try to find the field by name in the structure
	field := v.FieldByName(fieldName)
	if !field.IsValid() {
		fmt.Println(fieldName + " field is not found in the struct")
		return
	}

	// Check and handle if the field is a pointer
	if field.Kind() == reflect.Ptr {
		// Initialize the field if it is nil
		if field.IsNil() {
			field.Set(reflect.New(field.Type().Elem()))
		}
		field = field.Elem()
	}

	// Ensure the field is of type string
	if field.Kind() != reflect.String {
		fmt.Println(fieldName + " field is not a string")
		return
	}

	// Check if the field can be set
	if !field.CanSet() {
		fmt.Println("Cannot set " + fieldName + " field")
		return
	}

	// Set the new value for the field
	field.SetString(newValue)
}

// Configuration synchronization between mediamtx and strm_server
func ConfigSync(t *ApiServer) {
	t.api.mutex.Lock()
	defer t.api.mutex.Unlock()
	newConf := *t.api.Conf
	newConf.SetDefaults()
	newConf.Paths = nil
	newConf.OptionalPaths = nil
	for _, pipeConfig := range t.strmConf.Pipes {
		if pipeConfig.Source == "RTPR" {
			newConf.AddPath(pipeConfig.Name, nil)
			newConf.Validate()
			setField(newConf.OptionalPaths[pipeConfig.Name], "Source", pipeConfig.RTPR.VideoURL)
			setField(newConf.OptionalPaths[pipeConfig.Name], "AudioSource", pipeConfig.RTPR.AudioURL)
			//fmt.Println("AudioSource", pipeConfig.RTPR.AudioURL)
			newConf.Validate()
			//fmt.Println("newConf AudioSource", newConf.Paths[pipeConfig.Name].AudioSource)
		}
	}
	//newConf.Validate()del
	t.api.Conf = &newConf
	t.api.Parent.APIConfigSet(&newConf)
}

func ExtractID(msg *Message) (int, error) {
	if msg == nil {
		return 0, errorf("Message is nil")
	}

	idStr, ok := msg.Data["id"]
	if !ok {
		return 0, errorf("ID not found in message data")
	}

	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, errorf("Error converting id to int: %v", err)
	}

	return idInt, nil
}

func DeletePipeByID(t *ApiServer, id int) error {
	if t == nil || t.strmConf == nil || t.strmConf.Pipes == nil {
		return errorf("no configuration or pipes found")
	}
	if _, ok := t.strmConf.Pipes[id]; !ok {
		return errorf("no pipe found with ID %d", id)
	}

	delete(t.strmConf.Pipes, id)
	return nil
}

func ApiUpdatePipeConfig(api *ApiServer, req *Message, configType string) (Message, int) {
	id, err := ExtractID(req)
	if err != nil {
		return getError(req, 100)
	}

	pipe, exists := api.strmConf.Pipes[id]
	if !exists {
		return getError(req, 101)
	}

	// Obtain the field (RTPR or RTPS) using reflection
	v := reflect.ValueOf(&pipe).Elem()
	field := v.FieldByName(configType)
	if !field.IsValid() {
		return getError(req, 104)
	}

	// Create a map to match lowercased field names to reflect.Value fields
	fieldsMap := make(map[string]reflect.Value)
	fType := field.Type()
	for i := 0; i < fType.NumField(); i++ {
		fieldName := fType.Field(i).Name
		fieldsMap[strings.ToLower(fieldName)] = field.FieldByName(fieldName)
	}

	// Iterate over the provided data fields and set them, ignoring case and ID field
	for key, value := range req.Data {
		lowerKey := strings.ToLower(key)
		if lowerKey == "id" {
			continue // skip the ID field as it should not be changed
		}

		subField, found := fieldsMap[lowerKey]
		if !found {
			return getError(req, 104)
		}

		if subField.CanSet() {
			switch subField.Kind() {
			case reflect.String:
				subField.SetString(value)
			case reflect.Int:
				intValue, err := strconv.Atoi(value)
				if err != nil {
					return getError(req, 105)
				}
				subField.SetInt(int64(intValue))
			default:
				return getError(req, 105)
			}
		}
	}

	// Update the pipe configuration after making changes
	api.strmConf.Pipes[id] = pipe
	//fmt.Println("api.strmConf.Pipes[id]: ", api.strmConf.Pipes[id])
	ConfigSync(api)
	return Message{
		Name: "success",
		Data: map[string]string{"status": "Configuration updated successfully"},
	}, 0
}

func ApiAddPipe(t *ApiServer, req *Message) (Message, int) {
	// Use ExtractID to get the ID from the message
	id, err := ExtractID(req)
	if err != nil {
		return getError(req, 100)
	}

	// Check for existing ID
	if t.strmConf.Pipes == nil {
		t.strmConf.Pipes = make(map[int]conf.PipeConfig)
	}
	if _, exists := t.strmConf.Pipes[id]; exists {
		return getError(req, 102)
	}
	// Create a new PipeConfig and add it to the map
	newPipe := conf.PipeConfig{
		ID:    id,
		State: "active", // Example default state
		Type:  "sending",
	}
	if t.strmConf.Pipes == nil {
		t.strmConf.Pipes = make(map[int]conf.PipeConfig)
	}
	t.strmConf.Pipes[id] = newPipe

	r, e := ApiSetPipe(t, req)
	if e != 0 {
		return r, e
	}
	ConfigSync(t)
	// Return success message
	return Message{
		Name: "success",
		Data: map[string]string{
			"status": "Pipe added successfully",
		},
	}, 0
}

func ApiAddRtp(api *API, req *Message) (Message, error) {
	id := req.Data["id"]
	response := Message{req.Corr, "msg", "res", req.Verb, req.Noun, make(map[string]string)}
	response.Data["result"] = "OK"
	response.Data["id"] = id
	return response, nil
}

func ApiAddRtsp(api *API, req *Message) (Message, error) {
	id := req.Data["id"]
	response := Message{req.Corr, "msg", "res", req.Verb, req.Noun, make(map[string]string)}
	response.Data["result"] = "OK"
	response.Data["id"] = id
	return response, nil
}

func ApiDelPipe(api *ApiServer, req *Message) (Message, int) {
	response := Message{req.Corr, "msg", "res", req.Verb, req.Noun, make(map[string]string)}
	id, _ := ExtractID(req)
	err := DeletePipeByID(api, id)
	if err != nil {
		return getError(req, 103)
	} else {
		ConfigSync(api)
	}
	response.Data["result"] = "OK"
	response.Data["id"] = strconv.Itoa(id)
	return response, 0
}

func ApiDelRtp(api *API, req *Message) (Message, error) {
	id := req.Data["id"]
	response := Message{req.Corr, "msg", "res", req.Verb, req.Noun, make(map[string]string)}
	response.Data["result"] = "OK"
	response.Data["id"] = id
	return response, nil
}

func ApiDelRtsp(api *API, req *Message) (Message, error) {
	id := req.Data["id"]
	response := Message{req.Corr, "msg", "res", req.Verb, req.Noun, make(map[string]string)}
	response.Data["result"] = "OK"
	response.Data["id"] = id
	return response, nil
}

// ApiSetPipe updates a specific field in the PipeConfig for a given Pipe ID from Message, case-insensitively
func ApiSetPipe(t *ApiServer, req *Message) (Message, int) {
	// Extracting pipe ID from the request
	id, err := ExtractID(req)
	if err != nil {
		return getError(req, 100)
	}

	// Retrieving the PipeConfig
	pipe, exists := t.strmConf.Pipes[id]
	if !exists {
		return getError(req, 101)
	}

	// Using reflection to set field dynamically and case-insensitively
	pipeValue := reflect.ValueOf(&pipe).Elem()
	pipeType := pipeValue.Type()

	// Normalize incoming data keys to lowercase
	normalizedData := make(map[string]string)
	for key, value := range req.Data {
		normalizedKey := strings.ToLower(key)
		normalizedData[normalizedKey] = value
	}

	for i := 0; i < pipeType.NumField(); i++ {
		field := pipeValue.Field(i)
		fieldName := strings.ToLower(pipeType.Field(i).Name)

		// Skip id field, normalize field name to lowercase
		if fieldName == "id" {
			continue
		}

		if fieldValue, ok := normalizedData[fieldName]; ok && field.CanSet() {
			// Check and set field by type
			switch field.Kind() {
			case reflect.String:
				field.SetString(fieldValue)
			default:
				fmt.Println("unsupported field type:", field.Type())
				continue
			}
		}
	}

	// Save back the modified PipeConfig
	t.strmConf.Pipes[id] = pipe
	ConfigSync(t)
	return Message{
		Name: "success",
		Data: map[string]string{"status": "pipe updated successfully"},
	}, 0
}

func ApiSetRtp(api *API, req *Message) (Message, error) {
	response := Message{req.Corr, "msg", "res", req.Verb, req.Noun, make(map[string]string)}
	response.Data["result"] = "OK"
	return response, nil
}

func ApiSetRtsp(api *API, req *Message) (Message, error) {
	response := Message{req.Corr, "msg", "res", req.Verb, req.Noun, make(map[string]string)}
	response.Data["result"] = "OK"
	return response, nil
}

// ApiGetPipe retrieves the values of specified fields from a PipeConfig in api.strmConf.Pipes by ID.
func ApiGetPipe(api *ApiServer, req *Message) (Message, int) {
	// Extract the pipe ID from the request
	id, err := ExtractID(req)
	if err != nil {
		return getError(req, 100)
	}

	// Retrieve the specific PipeConfig
	pipe, exists := api.strmConf.Pipes[id]
	if !exists {
		return getError(req, 101)
	}

	responseData := map[string]string{}

	// Iterate over each data key in the request (these are field names)
	v := reflect.ValueOf(&pipe).Elem()
	fieldsMap := make(map[string]reflect.Value)
	fType := v.Type()
	for i := 0; i < fType.NumField(); i++ {
		fieldName := fType.Field(i).Name
		fieldsMap[strings.ToLower(fieldName)] = v.FieldByName(fieldName)
	}
	for key := range req.Data {
		lowerKey := strings.ToLower(key)
		if lowerKey == "id" {
			continue // skip the ID field
		}

		fieldName := strings.ToLower(lowerKey) // Assume field names are in correct case
		fieldValue := fieldsMap[lowerKey]
		if !fieldValue.IsValid() {
			return getError(req, 104)
		}

		// Convert the field value to a string representation
		var valueStr string
		switch fieldValue.Kind() {
		case reflect.String:
			valueStr = fieldValue.String()
		case reflect.Int, reflect.Int32, reflect.Int64:
			valueStr = strconv.FormatInt(fieldValue.Int(), 10)
		// case reflect.Float32, reflect.Float64:
		// 	valueStr = strconv.FormatFloat(fieldValue.Float(), 'f', -1, 64)
		case reflect.Bool:
			valueStr = strconv.FormatBool(fieldValue.Bool())
		default:
			return getError(req, 104)
		}

		responseData[strings.ToLower(fieldName)] = valueStr
		fmt.Println("responseData", responseData)
	}

	// Return the field values as part of the response
	return Message{
		Name: "success",
		Data: responseData,
	}, 0
}

func ApiGetSubConfigField(api *ApiServer, req *Message, configType string) (Message, int) {
	id, err := ExtractID(req)
	if err != nil {
		return getError(req, 100)
	}

	pipe, exists := api.strmConf.Pipes[id]
	if !exists {
		return getError(req, 101)
	}

	v := reflect.ValueOf(&pipe).Elem()
	subConfigField := v.FieldByName(configType)
	//fmt.Println("subConfigField: ", subConfigField)
	if !subConfigField.IsValid() {
		return getError(req, 104)
	}

	responseData := make(map[string]string)
	// Create a map to match lowercased field names to reflect.Value fields
	fieldsMap := make(map[string]reflect.Value)
	fType := subConfigField.Type()
	for i := 0; i < fType.NumField(); i++ {
		fieldName := fType.Field(i).Name
		fieldsMap[strings.ToLower(fieldName)] = subConfigField.FieldByName(fieldName)
	}
	// Iterate over each data key in the request (these are field names)
	for key := range req.Data {
		lowerKey := strings.ToLower(key)
		if lowerKey == "id" {
			continue // skip the ID field
		}

		fieldName := strings.ToLower(lowerKey) // Assume field names are in correct case
		//fmt.Println("fieldName: ", fieldName)
		fieldValue := fieldsMap[lowerKey]
		if !fieldValue.IsValid() {
			return getError(req, 104)
		}

		// Convert the field value to a string representation
		var valueStr string
		switch fieldValue.Kind() {
		case reflect.String:
			valueStr = fieldValue.String()
		case reflect.Int, reflect.Int32, reflect.Int64:
			valueStr = strconv.FormatInt(fieldValue.Int(), 10)
		case reflect.Bool:
			valueStr = strconv.FormatBool(fieldValue.Bool())
		default:
			return getError(req, 105)
		}

		responseData[strings.ToLower(fieldName)] = valueStr
	}

	return Message{
		Name: "success",
		Data: responseData,
	}, 0
}

func ApiGetRtsp(api *API, req *Message) (Message, error) {
	response := Message{req.Corr, "msg", "res", req.Verb, req.Noun, make(map[string]string)}
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

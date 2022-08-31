package main

import (
	"broker/event"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/rpc"
)

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

func (app *Config) Broker(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "Hit the broker",
	}

	_ = app.writeJSON(w, http.StatusOK, payload)
}

func (app *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var request RequestPayload

	err := app.readJSON(w, r, &request)
	if err != nil {
		app.errJSON(w, err)
		return
	}

	switch request.Action {
	case "auth":
		app.Authenticate(w, request.Auth)
	case "log":
		app.MakeRPCLog(w, request.Log)
	case "mail":
		app.SendMail(w, request.Mail)
	default:
		app.errJSON(w, errors.New("Unknown action"))
	}
}

func (app *Config) Authenticate(w http.ResponseWriter, a AuthPayload) {
	//	create json to auth service
	jsonData, _ := json.MarshalIndent(a, "", "\t")

	//	call service
	request, err := http.NewRequest("POST", fmt.Sprintf("http://auth-service:%s/authenticate", webPort), bytes.NewBuffer(jsonData))
	if err != nil {
		app.errJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.errJSON(w, err)
		return
	}
	defer response.Body.Close()

	//	make sure to return correct status code
	if response.StatusCode == http.StatusUnauthorized {
		app.errJSON(w, errors.New("invalid credentials"))
		return
	} else if response.StatusCode != http.StatusAccepted {
		app.errJSON(w, errors.New("error calling auth service"))
		return
	}

	var jsonFromService jsonResponse

	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		app.errJSON(w, err)
		return
	}

	if jsonFromService.Error {
		app.errJSON(w, errors.New("error calling auth service"), http.StatusUnauthorized)
		return
	}

	var payload jsonResponse

	payload.Error = false
	payload.Message = "Authenticated"
	payload.Data = jsonFromService.Data

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) LogItem(w http.ResponseWriter, l LogPayload) {
	jsonData, _ := json.MarshalIndent(l, "", "\t")

	logServiceUrl := fmt.Sprintf("http://log-service:%s/log", webPort)

	request, err := http.NewRequest("POST", logServiceUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		app.errJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.errJSON(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.errJSON(w, errors.New("error calling logs service"))
		return
	}

	var jsonFromService jsonResponse

	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		app.errJSON(w, err)
		return
	}

	var payload jsonResponse

	payload.Error = false
	payload.Message = jsonFromService.Message
	payload.Data = jsonFromService.Data

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) SendMail(w http.ResponseWriter, m MailPayload) {
	jsonData, _ := json.MarshalIndent(m, "", "\t")

	mailServiceUrl := fmt.Sprintf("http://mail-service:%s/send", webPort)

	request, err := http.NewRequest("POST", mailServiceUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		app.errJSON(w, err)
		return
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		app.errJSON(w, err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusAccepted {
		app.errJSON(w, errors.New("error calling mail service"))
		return
	}

	var jsonFromService jsonResponse

	err = json.NewDecoder(response.Body).Decode(&jsonFromService)
	if err != nil {
		app.errJSON(w, err)
		return
	}

	var payload jsonResponse

	payload.Error = false
	payload.Message = jsonFromService.Message
	payload.Data = jsonFromService.Data

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) MakeLogQueueEvent(w http.ResponseWriter, l LogPayload) {
	err := app.pushToQueue(l.Name, l.Data)

	if err != nil {
		app.errJSON(w, err)
		return
	}

	var payload jsonResponse

	payload.Error = false
	payload.Message = "RabbitMQ Log event pushed"

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) pushToQueue(name, message string) error {
	emitter, err := event.NewEventEmmiter(app.Rabbit)
	if err != nil {
		return err
	}

	payload := LogPayload{
		Name: name,
		Data: message,
	}

	j, _ := json.MarshalIndent(payload, "", "\t")
	err = emitter.Push(string(j), "log.INFO")

	if err != nil {
		return err
	}

	return nil
}

type RPCPayload struct {
	Name string
	Data string
}

func (app *Config) MakeRPCLog(w http.ResponseWriter, l LogPayload) {
	client, err := rpc.Dial("tcp", "log-service:5001")
	if err != nil {
		app.errJSON(w, err)
		return
	}

	payload := RPCPayload{
		Name: l.Name,
		Data: l.Data,
	}

	var result string

	err = client.Call("RPCServer.LogInfo", payload, &result)

	if err != nil {
		app.errJSON(w, err)
		return
	}

	response := jsonResponse{
		Error:   false,
		Message: result,
	}

	app.writeJSON(w, http.StatusAccepted, response)
}

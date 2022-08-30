package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

func (app *Config) SendMail(w http.ResponseWriter, r *http.Request) {
	type mailMessage struct {
		From    string `json:"from"`
		To      string `json:"to"`
		Subject string `json:"subject"`
		Message string `json:"message"`
	}

	var requestPayload mailMessage

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errJSON(w, err)
		return
	}

	msg := Message{
		From:    requestPayload.From,
		To:      requestPayload.To,
		Subject: requestPayload.Subject,
		Data:    requestPayload.Message,
	}

	err = app.Mailer.SendSMTPMessage(msg)
	if err != nil {
		app.errJSON(w, err)
		return
	}

	payloadMessage := fmt.Sprintf("Mail with subject '%s' sent to: %s", msg.Subject, msg.To)

	err = app.logRequest("SendMail", payloadMessage)

	payload := jsonResponse{
		Error:   false,
		Message: payloadMessage,
	}

	app.writeJSON(w, http.StatusAccepted, payload)
}

func (app *Config) logRequest(method, data string) error {
	var entry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	entry.Name = fmt.Sprintf("Mail-Service func: %s", method)
	entry.Data = data

	jsonData, _ := json.MarshalIndent(entry, "", "\t")
	logServiceUrl := fmt.Sprintf("http://log-service:%s/log", webPort)

	request, err := http.NewRequest("POST", logServiceUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	_, err = client.Do(request)
	if err != nil {
		return err
	}

	return nil
}

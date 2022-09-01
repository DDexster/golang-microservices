package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func (app *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	err := app.readJSON(w, r, &payload)
	if err != nil {
		app.errJSON(w, err, http.StatusBadRequest)
		return
	}

	//check if admin exist and generate one if not
	exists, _ := app.Models.User.IsAdminExist()

	if !exists {
		err = app.Models.User.GenerateAdminUser()
		if err != nil {
			app.errJSON(w, err, http.StatusInternalServerError)
			return
		}
	}

	//	validate user
	user, err := app.Models.User.GetByEmail(payload.Email)
	if err != nil {
		app.errJSON(w, errors.New("invalid Credentials"), http.StatusUnauthorized)
		return
	}

	valid, err := user.PasswordMatches(payload.Password)
	if err != nil || !valid {
		app.errJSON(w, errors.New("invalid password"), http.StatusUnauthorized)
		return
	}

	message := fmt.Sprintf("Logged in user %s", user.Email)
	// log authentication
	err = app.logRequest("Authenticate", message)
	if err != nil {
		app.errJSON(w, err)
		return
	}

	response := jsonResponse{
		Error:   false,
		Message: message,
		Data:    user,
	}

	app.writeJSON(w, http.StatusAccepted, response)
}

func (app *Config) logRequest(method, data string) error {
	var entry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}

	entry.Name = fmt.Sprintf("Auth-Service func %s", method)
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

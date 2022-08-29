package main

import (
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

	//	validate user
	user, err := app.Models.User.GetByEmail(payload.Email)
	if err != nil {
		app.errJSON(w, errors.New("invalid Credentials"), http.StatusUnauthorized)
		return
	}

	valid, err := user.PasswordMatches(payload.Password)
	if err != nil || !valid {
		app.errJSON(w, errors.New("invalid Credentials"), http.StatusUnauthorized)
		return
	}

	response := jsonResponse{
		Error:   false,
		Message: fmt.Sprintf("Logged in user %s", user.Email),
		Data:    user,
	}

	app.writeJSON(w, http.StatusAccepted, response)
}

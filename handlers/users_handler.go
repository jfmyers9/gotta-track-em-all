package handlers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/jfmyers9/gotta-track-em-all/db"
	"github.com/pivotal-golang/lager"
)

type UsersHandler struct {
	logger lager.Logger
	d      *db.DB
}

func NewUsersHandler(logger lager.Logger, d *db.DB) UsersHandler {
	return UsersHandler{logger, d}
}

type CreateRequest struct {
	Username        string `json:"username"`
	TrackerAPIToken string `json:"tracker_api_token"`
}

func (u UsersHandler) CreateUser(w http.ResponseWriter, req *http.Request) {
	logger := u.logger.Session("create-user")

	request := &CreateRequest{}

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.Error("failed-to-read-body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(data, request)
	if err != nil {
		logger.Error("failed-to-parse-request", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = u.d.CreateUser(logger, request.Username, request.TrackerAPIToken)
	if err != nil {
		logger.Error("failed-to-create-user", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (u UsersHandler) GetUser(w http.ResponseWriter, req *http.Request) {
	logger := u.logger.Session("get-user")

	username := req.FormValue(":username")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := u.d.GetUser(logger, username)
	if err != nil {
		logger.Error("failed-to-get-user", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data, err := json.Marshal(&user)
	if err != nil {
		logger.Error("failed-marshalling-data", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

type UpdateRequest struct {
	TrackerAPIToken string `json:"tracker_api_token"`
}

func (u UsersHandler) UpdateUser(w http.ResponseWriter, req *http.Request) {
	logger := u.logger.Session("update-user")

	request := &UpdateRequest{}

	data, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logger.Error("failed-to-read-body", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(data, request)
	if err != nil {
		logger.Error("failed-to-parse-request", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	username := req.FormValue(":username")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err = u.d.UpdateUser(logger, username, request.TrackerAPIToken)
	if err != nil {
		logger.Error("failed-to-update-user", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (u UsersHandler) DeleteUser(w http.ResponseWriter, req *http.Request) {
	logger := u.logger.Session("delete-user")

	username := req.FormValue(":username")
	if username == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	err := u.d.DeleteUser(logger, username)
	if err != nil {
		logger.Error("failed-to-delete-user", err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

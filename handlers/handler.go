package handlers

import (
	"net/http"

	"github.com/jfmyers9/gotta-track-em-all/db"
	"github.com/jfmyers9/gotta-track-em-all/routes"
	"github.com/pivotal-golang/lager"
	"github.com/tedsuo/rata"
)

func NewHandler(logger lager.Logger, d *db.DB) (http.Handler, error) {
	usersHandler := NewUsersHandler(logger, d)

	handlers := rata.Handlers{
		routes.CreateUser: http.HandlerFunc(usersHandler.CreateUser),
		routes.GetUser:    http.HandlerFunc(usersHandler.GetUser),
		routes.UpdateUser: http.HandlerFunc(usersHandler.UpdateUser),
		routes.DeleteUser: http.HandlerFunc(usersHandler.DeleteUser),
	}

	return rata.NewRouter(routes.Routes, handlers)
}

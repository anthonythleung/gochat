package main

import (
	"log"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/gorilla/mux"

	"github.com/AntsEclipse/gochat/protobuf/auth"
	"github.com/AntsEclipse/gochat/protobuf/user"
	"github.com/AntsEclipse/gochat/utils"
)

type server struct {
	userClient user.UserClient
	authClient auth.AuthClient
	router     *mux.Router
	log        *logrus.Entry
}

// LoginResult ... Result JSON from Login
type LoginResult struct {
	UserID int64  `json:"userid"`
	Token  string `json:"token"`
}

func main() {
	userClient := user.NewUserClient(helpers.Dial("user:50051"))
	authClient := auth.NewAuthClient(helpers.Dial("auth:50051"))

	s := server{
		router:     mux.NewRouter(),
		log:        helpers.Logger("login"),
		userClient: userClient,
		authClient: authClient,
	}
	s.routes()

	s.log.Info("Login Initialized")
	log.Fatal(http.ListenAndServe(":8080", helpers.CorsHandler(s.router)))
}

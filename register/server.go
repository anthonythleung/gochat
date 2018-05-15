package main

import (
	"log"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/AntsEclipse/gochat/protobuf/auth"
	"github.com/AntsEclipse/gochat/protobuf/user"
	"github.com/AntsEclipse/gochat/utils"
	"github.com/gorilla/mux"
)

type server struct {
	userClient user.UserClient
	authClient auth.AuthClient
	log        *logrus.Entry
	router     *mux.Router
}

// RegisterResult ... Result JSON from register
type RegisterResult struct {
	UserID int64 `json:"userid"`
}

func main() {
	userClient := user.NewUserClient(helpers.Dial("user:50051"))
	authClient := auth.NewAuthClient(helpers.Dial("auth:50051"))

	s := server{
		userClient: userClient,
		authClient: authClient,
		log:        helpers.Logger("register"),
		router:     mux.NewRouter(),
	}

	log.Fatal(http.ListenAndServe(":8080", helpers.CorsHandler(s.router)))
}

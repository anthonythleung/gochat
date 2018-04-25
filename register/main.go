package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/AntsEclipse/gochat/protobuf/auth"
	"github.com/AntsEclipse/gochat/protobuf/user"
	"github.com/AntsEclipse/gochat/utils"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

var (
	conn       *grpc.ClientConn
	err        error
	userClient user.UserClient
	authClient auth.AuthClient
)

// RegisterResult ... Result JSON from register
type RegisterResult struct {
	UserID int64 `json:"userid"`
}

/**
 * @api {post} /register Request User information
 * @apiName post
 * @apiGroup Register
 *
 * @apiParam {string} email Users email.
 * @apiParam {string} password Users password.
 *
 * @apiSuccess {String} userid ID of the User.
 */

func post(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(100)
	if err != nil {
		panic(err)
	}
	email := r.PostFormValue("email")
	password := r.PostFormValue("password")

	userResult, userErr := userClient.CreateUser(context.Background(), &user.Request{Email: email})
	if userErr != nil {
		http.Error(w, userErr.Error(), http.StatusBadRequest)
		return
	}
	authResult, authErr := authClient.CreateUser(context.Background(), &auth.Request{UserId: userResult.UserId, Password: password})
	if authErr != nil {
		http.Error(w, authErr.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(RegisterResult{UserID: authResult.UserId})
}

func handleUserRegister(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		post(w, r)
	}
}

func main() {
	userClient = user.NewUserClient(helpers.Dial("user:50051"))
	authClient = auth.NewAuthClient(helpers.Dial("auth:50051"))

	router := mux.NewRouter()
	router.HandleFunc("/", handleUserRegister).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))
}

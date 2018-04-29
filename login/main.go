package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/AntsEclipse/gochat/protobuf/auth"
	"github.com/AntsEclipse/gochat/protobuf/user"
	"github.com/AntsEclipse/gochat/utils"
	"google.golang.org/grpc"
)

var (
	conn       *grpc.ClientConn
	err        error
	userClient user.UserClient
	authClient auth.AuthClient
)

// LoginResult ... Result JSON from Login
type LoginResult struct {
	UserID int64  `json:"userid"`
	Token  string `json:"token"`
}

/**
 *
 * @api {post} /login Users login
 * @apiName post
 * @apiGroup Login
 *
 * @apiParam  {String} email A users email.
 * @apiParam  {String} password A users password.
 *
 * @apiSuccess  {number} userid A users unique id.
 * @apiSuccess  {string} token A users jwt token.
 */

func post(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(100)
	if err != nil {
		panic(err)
	}

	email := r.PostFormValue("email")
	password := r.PostFormValue("password")

	userResult, err := userClient.GetUserID(context.Background(), &user.Request{Email: email})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	authResult, err := authClient.GetToken(context.Background(), &auth.Request{UserId: userResult.UserId, Password: password})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(LoginResult{UserID: userResult.UserId, Token: authResult.Token})
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

	log.Fatal(http.ListenAndServe(":8080", helpers.CorsHandler(router)))
}

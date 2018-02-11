package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/AntsEclipse/gochat/protobuf/auth"
	"github.com/AntsEclipse/gochat/protobuf/user"
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

func post(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(100)
	if err != nil {
		panic(err)
	}

	username := r.PostFormValue("username")
	password := r.PostFormValue("password")

	userResult, err := userClient.GetUserID(context.Background(), &user.Request{UserName: username})
	if err != nil {
		panic(err)
	}
	authResult, err := authClient.GetToken(context.Background(), &auth.Request{UserId: userResult.UserId, Password: password})
	if err != nil {
		panic(err)
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
	userClient = user.NewUserClient(dial("user:50051"))
	authClient = auth.NewAuthClient(dial("auth:50051"))

	router := mux.NewRouter()
	router.HandleFunc("/", handleUserRegister).Methods("POST")
	log.Fatal(http.ListenAndServe(":8080", router))
}

func dial(addr string) *grpc.ClientConn {
	conn, err = grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	return conn
}

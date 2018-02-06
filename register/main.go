package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/AntsEclipse/gochat/protobuf/auth"
	"github.com/AntsEclipse/gochat/protobuf/user"
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

func post(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(100)
	if err != nil {
		panic(err)
	}
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")

	userResult, userErr := userClient.CreateUser(context.Background(), &user.Request{UserName: username})
	if userErr != nil {
		panic(err)
	}
	authResult, authErr := authClient.CreateUser(context.Background(), &auth.Request{UserId: userResult.UserId, Password: password})
	if authErr != nil {
		panic(authErr)
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

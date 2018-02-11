package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/AntsEclipse/gochat/protobuf/auth"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

func get(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(100)
	if err != nil {
		panic(err)
	}
	userID, err := strconv.ParseInt(r.PostFormValue("userID"), 10, 64)
	if err != nil {
		panic(err)
	}
	password := r.PostFormValue("password")
	token := requestJWT(userID, password)
	json.NewEncoder(w).Encode(token)
}

func handleAuth(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		get(w, r)
	}
}

type server struct{}

func (s *server) CreateUser(ctx context.Context, req *auth.Request) (*auth.Result, error) {
	newUser := createUserCredential(req.UserId, req.Password)
	return &auth.Result{UserId: newUser.UserID}, nil
}

func (s *server) GetToken(ctx context.Context, req *auth.Request) (*auth.Token, error) {
	newToken := requestJWT(req.GetUserId(), req.GetPassword())
	return &auth.Token{Token: newToken.Token}, nil
}

func main() {
	initDB()
	initKeys()

	lis, _ := net.Listen("tcp", ":50051")

	s := grpc.NewServer()
	auth.RegisterAuthServer(s, &server{})
	go s.Serve(lis)

	router := mux.NewRouter()
	router.HandleFunc("/", handleAuth).Methods("GET", "POST")
	log.Fatal(http.ListenAndServe(":8080", router))
}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/AntsEclipse/gochat/auth/utils"
	"github.com/AntsEclipse/gochat/protobuf/user"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

func handleUser(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		userGet(w, r)
	case "PUT":
		put(w, r)
	case "DELETE":
		delete(w, r)
	}
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		usersGet(w, r)
	}
}

func userGet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.ParseInt(params["userID"], 10, 64)
	if err != nil {
		panic(err)
	}
	user := getUser(id)
	json.NewEncoder(w).Encode(user)
}

func usersGet(w http.ResponseWriter, r *http.Request) {
	users := getUsers()
	json.NewEncoder(w).Encode(users)
}

func put(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "TODO")
}

func delete(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "TODO")
}

type server struct{}

// protobuf function to create a user
func (s *server) CreateUser(ctx context.Context, req *user.Request) (*user.Result, error) {
	newUser := createUser(req.GetUserName())
	return &user.Result{UserId: newUser.ID}, nil
}

func (s *server) GetUserID(ctx context.Context, req *user.Request) (*user.Result, error) {
	result := findUser(req.GetUserName())
	return &user.Result{UserId: result.ID}, nil
}

func main() {
	initDB()

	lis, _ := net.Listen("tcp", ":50051")

	s := grpc.NewServer()
	user.RegisterUserServer(s, &server{})
	go s.Serve(lis)

	router := mux.NewRouter()
	router.Use(authUtil.ValidateTokenMiddleware)
	router.HandleFunc("/", handleUsers).Methods("GET")
	router.HandleFunc("/{userID}", handleUser).Methods("GET", "PUT", "DELETE")
	log.Fatal(http.ListenAndServe(":8080", router))
}

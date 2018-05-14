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
	"github.com/AntsEclipse/gochat/utils"
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

/**
 * @api {get} /user/:userId Request User information
 * @apiName userGet
 * @apiGroup User
 *
 * @apiParam {Number} userId Users unique ID.
 *
 * @apiSuccess {String} id ID of the User.
 * @apiSuccess {String} username Username of the User.
 * @apiSuccess {String} email Email of the User.
 */
func userGet(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	id, err := strconv.ParseInt(params["userID"], 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	user, err := getUser(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(user)
}

/**
 * @api {get} /user/ Request Users information
 * @apiName usersGet
 * @apiGroup User
 *
 * @apiParam {Number} username Users username.
 * @apiParam {Number} email Users email.
 *
 * @apiSuccess {User[]} users List of user profile
 * @apiSuccess {String} users.id ID of the User.
 * @apiSuccess {String} users.username Username of the User.
 * @apiSuccess {String} users.email Email of the User.
 */

func usersGet(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	email := r.FormValue("email")
	username := r.FormValue("username")
	user := User{Username: username, Email: email}
	users, err := getUsers(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(users)
}

/**
 * @api {put} /user/:userId Update User information
 * @apiName put
 * @apiGroup User
 *
 * @apiParam {Number} userId Users unique ID.
 * @apiParam {String} username Users new username.
 *
 * @apiSuccess {String} id ID of the User.
 * @apiSuccess {String} username Username of the User.
 * @apiSuccess {String} email Email of the User.
 */
func put(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	err := r.ParseMultipartForm(100)
	id, err := strconv.ParseInt(params["userID"], 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	username := r.PostFormValue("username")
	user, err := updateUser(User{ID: id, Username: username})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	json.NewEncoder(w).Encode(user)
}

func delete(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "TODO")
}

type server struct{}

// protobuf function to create a user
func (s *server) CreateUser(ctx context.Context, req *user.Request) (*user.Result, error) {
	newUser, err := createUser(req.GetEmail())
	return &user.Result{UserId: newUser.ID}, err
}

// protobuf function to get a user's id from user email
func (s *server) GetUserID(ctx context.Context, req *user.Request) (*user.Result, error) {
	result, err := findUser(req.GetEmail())
	return &user.Result{UserId: result.ID}, err
}

func main() {
	initDB()

	lis, _ := net.Listen("tcp", ":50051")

	s := grpc.NewServer()
	user.RegisterUserServer(s, &server{})
	go s.Serve(lis)

	router := mux.NewRouter()
	router.Use(authutil.ValidateTokenMiddleware)
	router.HandleFunc("/", handleUsers).Methods("GET")
	router.HandleFunc("/{userID}", handleUser).Methods("GET", "PUT", "DELETE")

	log.Fatal(http.ListenAndServe(":8080", helpers.CorsHandler(router)))
}

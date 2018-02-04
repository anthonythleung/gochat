package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/AntsEclipse/gochat/protobuf/user"
	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

var db *pg.DB

// User ... The main user object
type User struct {
	ID        int64
	Username  string
	createdAt int64
	updatedAt int64
}

func (u User) String() string {
	return fmt.Sprintf("User<%d %s>", u.ID, u.Username)
}

func createSchema(db *pg.DB) error {
	for _, model := range []interface{}{&User{}} {
		err := db.CreateTable(model, nil)
		if err != nil {
			return nil
		}
	}
	return nil
}

func getUser(id int64) User {
	user := User{ID: id}
	err := db.Select(&user)
	if err != nil {
		panic(err)
	}

	return user
}

func getUsers() []User {
	var users []User
	err := db.Model(&users).Select()
	if err != nil {
		panic(err)
	}
	return users
}

func createUser(username string) User {
	newUser := &User{
		Username: username,
	}

	err := db.Insert(newUser)

	if err != nil {
		panic(err)
	}

	return *newUser
}

func deleteUser(id int64) {
	user := &User{
		ID: id,
	}

	err := db.Delete(user)

	if err != nil {
		panic(err)
	}
}

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
	return &user.Result{UserID: newUser.ID}, nil
}

func main() {
	db = pg.Connect(&pg.Options{
		User:     "postgres",
		Password: "password",
		Addr:     "db:5432",
	})
	createSchema(db)

	lis, _ := net.Listen("tcp", ":50051")

	s := grpc.NewServer()
	user.RegisterUserServer(s, &server{})
	go s.Serve(lis)

	router := mux.NewRouter()
	router.HandleFunc("/", handleUsers).Methods("GET")
	router.HandleFunc("/{userID}", handleUser).Methods("GET", "PUT", "DELETE")
	log.Fatal(http.ListenAndServe(":8080", router))
}

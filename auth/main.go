package main

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/AntsEclipse/gochat/protobuf/auth"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-pg/pg"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
)

const (
	privateKeyPath = "/go/src/github.com/AntsEclipse/gochat/auth/gochat.rsa"
	publicKeyPath  = "/go/src/github.com/AntsEclipse/gochat/auth/gochat.rsa.pub"
)

var (
	verifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
)

func initKeys() {
	signBytes, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		panic(err)
	}

	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		panic(err)
	}

	verifyBytes, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		panic(err)
	}

	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		panic(err)
	}

}

var db *pg.DB

// UserCredential ... A User Credential Object with UserID and Password
type UserCredential struct {
	ID       int64
	UserID   int64
	Password string
}

func createSchema(db *pg.DB) error {
	for _, model := range []interface{}{&UserCredential{}} {
		err := db.CreateTable(model, nil)
		if err != nil {
			return nil
		}
	}
	return nil
}

// Token ... A JWT Token
type Token struct {
	Token string `json:"token"`
}

func requestJWT(userID int64, password string) Token {
	var user UserCredential
	err := db.Model(&user).
		Where("user_id = ?", userID).
		Limit(1).
		Select()
	if err != nil {
		panic(err)
	}

	if !checkPasswordHash(password, user.Password) {
		return Token{}
	}

	token := jwt.New(jwt.SigningMethodRS256)
	claims := make(jwt.MapClaims)
	claims["userID"] = user.UserID
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	token.Claims = claims
	tokenString, err := token.SignedString(signKey)

	if err != nil {
		panic(err)
	}

	return Token{tokenString}
}

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

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func createUser(userID int64, password string) UserCredential {
	hash, err := hashPassword(password)
	if err != nil {
		panic(err)
	}
	newUser := &UserCredential{
		UserID:   userID,
		Password: hash,
	}

	err = db.Insert(newUser)

	if err != nil {
		panic(err)
	}

	return *newUser
}

func handleAuth(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		get(w, r)
	}
}

type server struct{}

func (s *server) CreateUser(ctx context.Context, req *auth.Request) (*auth.Result, error) {
	newUser := createUser(req.UserID, req.Password)
	return &auth.Result{UserID: newUser.UserID}, nil
}

func main() {
	db = pg.Connect(&pg.Options{
		User:     "postgres",
		Password: "password",
		Addr:     "db:5432",
	})
	db.OnQueryProcessed(func(event *pg.QueryProcessedEvent) {
		query, err := event.FormattedQuery()
		if err != nil {
			panic(err)
		}

		log.Printf("%s %s", time.Since(event.StartTime), query)
	})
	createSchema(db)

	initKeys()

	lis, _ := net.Listen("tcp", ":50051")

	s := grpc.NewServer()
	auth.RegisterAuthServer(s, &server{})
	go s.Serve(lis)

	router := mux.NewRouter()
	router.HandleFunc("/", handleAuth).Methods("GET", "POST")
	log.Fatal(http.ListenAndServe(":8080", router))
}

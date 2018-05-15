package main

import (
	"net"
	"net/http"
	"time"

	"github.com/AntsEclipse/gochat/protobuf/auth"
	"github.com/AntsEclipse/gochat/utils"
	jwt "github.com/dgrijalva/jwt-go"
	"google.golang.org/grpc"

	"github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
)

type server struct {
	dao    *dao
	crypto *crypto
	router *mux.Router
	log    *logrus.Entry
}

func main() {
	d := dao{}
	d.initDB()

	c := crypto{}
	c.initKeys()

	s := server{
		dao:    &d,
		crypto: &c,
		router: mux.NewRouter(),
		log:    helpers.Logger("auth"),
	}
	s.routes()

	grpcServer := grpc.NewServer()
	auth.RegisterAuthServer(grpcServer, &s)
	lis, _ := net.Listen("tcp", ":50051")
	go grpcServer.Serve(lis)

	s.log.Info("Auth Initialized")
	s.log.Fatal(http.ListenAndServe(":8080", helpers.CorsHandler(s.router)))
}

// Token ... A JWT Token
type Token struct {
	Token string `json:"token"`
}

func (s *server) requestJWT(userID int64, password string) Token {
	var user UserCredential
	err := s.dao.db.Model(&user).
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
	tokenString, err := token.SignedString(s.crypto.signKey)

	if err != nil {
		panic(err)
	}

	return Token{tokenString}
}

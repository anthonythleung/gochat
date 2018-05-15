package main

import (
	"net"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/AntsEclipse/gochat/protobuf/user"
	"github.com/AntsEclipse/gochat/utils"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

type server struct {
	log    *logrus.Entry
	router *mux.Router
	dao    *dao
}

func main() {
	logger := helpers.Logger("user")
	d := dao{
		log: logger,
	}
	d.initDB()

	s := server{
		log:    logger,
		router: mux.NewRouter(),
		dao:    &d,
	}
	s.routes()

	lis, _ := net.Listen("tcp", ":50051")
	grpcServer := grpc.NewServer()
	user.RegisterUserServer(grpcServer, &s)
	go grpcServer.Serve(lis)

	s.log.Info("Initialized User")
	s.log.Fatal(http.ListenAndServe(":8080", helpers.CorsHandler(s.router)))
}

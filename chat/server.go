package main

import (
	"net"
	"net/http"

	"github.com/sirupsen/logrus"

	"github.com/AntsEclipse/gochat/protobuf/chat"
	"github.com/AntsEclipse/gochat/utils"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

// TODO: Use Slice instead of array and dynamically resize
type server struct {
	hubs   [99]*Hub
	uuids  map[string]int
	count  int
	router *mux.Router
	log    *logrus.Entry
}

func main() {
	s := server{
		uuids:  make(map[string]int),
		router: mux.NewRouter(),
		log:    helpers.Logger("chat"),
	}
	s.routes()

	grpcServer := grpc.NewServer()
	chat.RegisterChatServer(grpcServer, &s)
	lis, _ := net.Listen("tcp", ":50051")
	go grpcServer.Serve(lis)

	s.log.Info("Chat Initialized")
	s.log.Fatal(http.ListenAndServe(":8080", helpers.CorsHandler(s.router)))
}

package main

import (
	"net"
	"net/http"

	"github.com/garyburd/redigo/redis"
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
	logger := helpers.Logger("chat")
	s := server{
		uuids:  make(map[string]int),
		router: mux.NewRouter(),
		log:    logger,
	}
	s.routes()

	grpcServer := grpc.NewServer()
	chat.RegisterChatServer(grpcServer, &s)
	lis, _ := net.Listen("tcp", ":50051")
	go grpcServer.Serve(lis)

	helpers.Wait("chat-redis:6379", logger)
	redisClient, err := redis.Dial("tcp", "chat-redis:6379")
	if err != nil {
		panic(err)
	}

	channelIDs, _ := redis.Strings(redisClient.Do("SMEMBERS", "channels:id"))

	for _, v := range channelIDs {
		s.RestoreChannel(v)
	}

	s.log.Info("Chat Initialized")
	s.log.Fatal(http.ListenAndServe(":8080", helpers.CorsHandler(s.router)))
}

func (s *server) RestoreChannel(channelID string) {
	hub := newHub(channelID, s.log)
	s.uuids[channelID] = s.count
	s.hubs[s.count] = hub
	s.count = s.count + 1
	s.log.WithFields(logrus.Fields{"id": channelID}).Info("Restoring Chat Server")
	go hub.run()
}

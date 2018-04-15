package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"

	"github.com/AntsEclipse/gochat/protobuf/chat"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

// TODO: Use Slice instead of array and dynamically resize
var (
	hubs  [99]*Hub
	uuids map[string]int
	count int
)

func handleChat(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	uuid := params["channelID"]
	fmt.Printf("Connecting: %s\n", uuid)
	fmt.Printf("ID: %s\n", strconv.Itoa(uuids[uuid]))
	serveWs(hubs[uuids[uuid]], w, r)
}

func createServer(uuid string) {
	hub := newHub(uuid)
	uuids[uuid] = count
	hubs[count] = hub
	count = count + 1
	log.Printf("Creating New Chat Server: %s\n", strconv.Itoa(count))
	go hub.run()
}

type server struct{}

// protobuf function to create a new server
func (s *server) CreateServer(ctx context.Context, req *chat.Request) (*chat.Result, error) {
	createServer(req.GetChannelId())
	return &chat.Result{ChannelId: req.GetChannelId()}, nil
}

func main() {
	lis, _ := net.Listen("tcp", ":50051")

	s := grpc.NewServer()
	chat.RegisterChatServer(s, &server{})
	go s.Serve(lis)

	uuids = make(map[string]int)

	router := mux.NewRouter()
	router.HandleFunc("/connect/{channelID}", handleChat)
	log.Fatal(http.ListenAndServe(":8080", router))
}

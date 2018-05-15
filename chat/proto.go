package main

import (
	"context"
	"log"
	"strconv"

	"github.com/AntsEclipse/gochat/protobuf/chat"
)

func (s *server) createServer(uuid string) {
	hub := newHub(uuid)
	s.uuids[uuid] = s.count
	s.hubs[s.count] = hub
	s.count = s.count + 1
	log.Printf("Creating New Chat Server: %s\n", strconv.Itoa(s.count))
	go hub.run()
}

// protobuf function to create a new server
func (s *server) CreateServer(ctx context.Context, req *chat.Request) (*chat.Result, error) {
	s.createServer(req.GetChannelId())
	return &chat.Result{ChannelId: req.GetChannelId()}, nil
}

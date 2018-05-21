package main

import (
	"context"
	"strconv"

	"github.com/AntsEclipse/gochat/protobuf/chat"
)

// protobuf function to create a new server
func (s *server) CreateServer(ctx context.Context, req *chat.Request) (*chat.Result, error) {
	uuid := req.GetChannelId()
	hub := newHub(uuid, s.log)
	s.uuids[uuid] = s.count
	s.hubs[s.count] = hub
	s.count = s.count + 1
	s.log.Printf("Creating New Chat Server: %s\n", strconv.Itoa(s.count))
	go hub.run()
	return &chat.Result{ChannelId: req.GetChannelId()}, nil
}

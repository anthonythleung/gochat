package main

import (
	"github.com/gocql/gocql"
)

// Hub ... WebSocket Hub
type Hub struct {
	id         string
	clients    map[*Client]bool
	messages   chan []byte
	session    *gocql.Session
	register   chan *Client
	unregister chan *Client
}

func newHub(id string) *Hub {
	return &Hub{
		id:         id,
		messages:   make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	// Setup Cassandra
	cluster := gocql.NewCluster("chat-cassandra1")
	cluster.Keyspace = "chat_log"
	cluster.Consistency = gocql.One
	session, err := cluster.CreateSession()
	if err != nil {
		panic(err)
	}
	defer session.Close()
	h.session = session

	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.messages:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

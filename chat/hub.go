package main

import (
	"github.com/gocql/gocql"
	"github.com/olivere/elastic"
)

// Hub ... WebSocket Hub
type Hub struct {
	id         string
	clients    map[*Client]bool
	messages   chan []byte
	session    *gocql.Session
	elastic    *elastic.Client
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
	cluster.Keyspace = "gochat"
	cluster.Consistency = gocql.One
	session, err := cluster.CreateSession()
	if err != nil {
		panic(err)
	}
	defer session.Close()
	h.session = session

	// Setup ElasticSearch
	client, err := elastic.NewClient(
		elastic.SetURL("http://chat-elasticsearch:9200"),
	)
	if err != nil {
		panic(err)
	}

	h.elastic = client

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

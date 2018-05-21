package main

import (
	"context"
	"net/http"

	"github.com/AntsEclipse/gochat/protobuf/chat"
	"github.com/AntsEclipse/gochat/utils"
	"github.com/garyburd/redigo/redis"
	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	"github.com/olivere/elastic"
	"github.com/sirupsen/logrus"
)

type server struct {
	router           *mux.Router
	log              *logrus.Entry
	chatClient       chat.ChatClient
	redisClient      redis.Conn
	cassandraSession *gocql.Session
	elasticClient    *elastic.Client
}

// Channel ... a channel lol
type Channel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Message ... a message
type Message struct {
	Type      string `json:"type"`
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

func main() {
	logger := helpers.Logger("channels")
	helpers.Wait("chat-redis:6379", logger)
	helpers.Wait("chat-cassandra1:9042", logger)

	// Setup gRPC
	chatClient := chat.NewChatClient(helpers.Dial("chat:50051"))

	// Setup Redis
	redisClient, err := redis.Dial("tcp", "chat-redis:6379")
	if err != nil {
		panic(err)
	}

	// Setup Cassandra
	cluster := gocql.NewCluster("chat-cassandra1")
	cluster.Keyspace = "gochat"
	cluster.Consistency = gocql.One
	cassandraSession, err := cluster.CreateSession()
	if err != nil {
		panic(err)
	}
	defer cassandraSession.Close()

	// Setup ElasticSearch
	elasticClient, err := elastic.NewClient(
		elastic.SetURL("http://chat-elasticsearch:9200"),
	)
	if err != nil {
		panic(err)
	}

	exists, err := elasticClient.IndexExists("messages").Do(context.Background())
	if err != nil {
		panic(err)
	}
	if !exists {
		_, err := elasticClient.CreateIndex("messages").Do(context.Background())
		if err != nil {
			panic(err)
		}
	}

	s := server{
		router:           mux.NewRouter(),
		log:              logger,
		chatClient:       chatClient,
		redisClient:      redisClient,
		cassandraSession: cassandraSession,
		elasticClient:    elasticClient,
	}
	s.routes()

	s.log.Info("Channels Initialized")
	s.log.Fatal(http.ListenAndServe(":8080", helpers.CorsHandler(s.router)))
}

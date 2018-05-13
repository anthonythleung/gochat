package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/AntsEclipse/gochat/auth/utils"
	"github.com/AntsEclipse/gochat/protobuf/chat"
	"github.com/AntsEclipse/gochat/utils"
	"github.com/garyburd/redigo/redis"
	"github.com/gocql/gocql"
	"github.com/gorilla/mux"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/olivere/elastic"
	"google.golang.org/grpc"
)

var (
	conn             *grpc.ClientConn
	err              error
	chatClient       chat.ChatClient
	redisClient      redis.Conn
	cassandraSession *gocql.Session
	elasticClient    *elastic.Client
)

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

func handleChannels(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		channelsGet(w, r)
	case "POST":
		channelsCreate(w, r)
	}
}

/**
 * @api {get} /channels/ Get a List of Channels
 * @apiName channelsGet
 * @apiGroup Channels
 *
 * @apiSuccess {Channel[]} channels List of channels
 */

func channelsGet(w http.ResponseWriter, r *http.Request) {
	channelIDs, _ := redis.Strings(redisClient.Do("SMEMBERS", "channels:id"))

	channels := make([]Channel, len(channelIDs))
	for i, v := range channelIDs {
		channelMap, _ := redis.StringMap(redisClient.Do("HGETALL", v))
		channels[i] = Channel{
			ID:   v,
			Name: channelMap["name"],
		}
	}
	json.NewEncoder(w).Encode(channels)
}

/**
 * @api {post} /channels/ Create a New Channel
 * @apiName channelsCreate
 * @apiGroup Channels
 *
 * @apiParam {string} channelName new channel"s name
 *
 * @apiSuccess {Channel} channel the new created channel
 */
func channelsCreate(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(100)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	channelName := r.PostFormValue("channelName")
	UUID, _ := uuid.NewV4()
	channelID := UUID.String()
	newChannel := Channel{
		ID:   channelID,
		Name: channelName,
	}
	chatClient.CreateServer(context.Background(), &chat.Request{ChannelId: channelID})
	redisClient.Do("SADD", "channels:id", channelID)
	redisClient.Do("HSET", channelID, "name", channelName)
	json.NewEncoder(w).Encode(newChannel)
}

func handleChannel(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		channelGet(w, r)
	case "PUT":
		channelUpdate(w, r)
	case "DELETE":
		channelDelete(w, r)
	}
}

func channelGet(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "NOT IMPLEMENTED", http.StatusMethodNotAllowed)
}

func channelUpdate(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "NOT IMPLEMENTED", http.StatusMethodNotAllowed)
}

func channelDelete(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "NOT IMPLEMENTED", http.StatusMethodNotAllowed)
}

func handleChannelHistory(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		history(w, r)
	}
}

func history(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	channelID := params["channelID"]

	iter := cassandraSession.Query(`select * from messages where channel_id = ? and type = 'MESSAGE' limit 100`, channelID).Iter()
	fmt.Println("got " + strconv.Itoa(iter.NumRows()))
	results := []Message{}
	m := map[string]interface{}{}

	for iter.MapScan(m) {
		results = append(results, Message{
			Type:      m["type"].(string),
			ID:        m["author_id"].(int64),
			Username:  "Test",
			Message:   m["content"].(string),
			Timestamp: int64(m["created_at"].(time.Time).UnixNano() / int64(time.Millisecond)),
		})
		m = map[string]interface{}{}
	}

	json.NewEncoder(w).Encode(results)
}

func handleChannelSearch(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		search(w, r)
	}
}

func search(w http.ResponseWriter, r *http.Request) {
	fmt.Println("New Query")
	// params := mux.Vars(r)
	// channelID := params["channelID"]
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	message := r.FormValue("message")
	query := elastic.NewFuzzyQuery("message", message)
	searchResult, err := elasticClient.Search().
		Index("messages").
		Query(query).
		From(0).Size(10).
		Pretty(true).
		Do(context.Background())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Printf("Query took %d milliseconds\n", searchResult.TookInMillis)
	fmt.Printf("Found a total of %d tweets\n", searchResult.TotalHits())
	results := []Message{}
	if searchResult.Hits.TotalHits > 0 {
		for _, hit := range searchResult.Hits.Hits {

			var m Message
			err := json.Unmarshal(*hit.Source, &m)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			results = append(results, m)
		}
	} else {
		fmt.Print("Found no message\n")
	}
	json.NewEncoder(w).Encode(results)

}

func main() {
	chatClient = chat.NewChatClient(helpers.Dial("chat:50051"))
	redisClient, err = redis.Dial("tcp", "chat-redis:6379")
	redisClient.Do("FLUSHDB")
	if err != nil {
		panic(err)
	}
	cluster := gocql.NewCluster("chat-cassandra1")
	cluster.Keyspace = "gochat"
	cluster.Consistency = gocql.One
	cassandraSession, err = cluster.CreateSession()
	if err != nil {
		panic(err)
	}
	defer cassandraSession.Close()

	// Setup ElasticSearch
	elasticClient, err = elastic.NewClient(
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

	router := mux.NewRouter()
	router.Use(authUtil.ValidateTokenMiddleware)
	router.HandleFunc("/", handleChannels)
	router.HandleFunc("/{channelID}", handleChannel)
	router.HandleFunc("/{channelID}/history", handleChannelHistory).Methods("GET")
	router.HandleFunc("/{channelID}/search", handleChannelSearch).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", helpers.CorsHandler(router)))
}

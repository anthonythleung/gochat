package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/AntsEclipse/gochat/auth/utils"
	"github.com/AntsEclipse/gochat/protobuf/chat"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
	uuid "github.com/nu7hatch/gouuid"
	"google.golang.org/grpc"
)

var (
	conn        *grpc.ClientConn
	err         error
	chatClient  chat.ChatClient
	redisClient redis.Conn
)

// Channel ... a channel lol
type Channel struct {
	ID   string `json:"id"`
	Name string `json:"name"`
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

func handleConnect(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		connect(w, r)
	}
}

func connect(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "NOT IMPLEMENTED", http.StatusMethodNotAllowed)
}

func main() {
	chatClient = chat.NewChatClient(dial("chat:50051"))
	redisClient, err = redis.Dial("tcp", "chat-redis:6379")
	if err != nil {
		panic(err)
	}

	router := mux.NewRouter()
	router.Use(authUtil.ValidateTokenMiddleware)
	router.HandleFunc("/", handleChannels)
	router.HandleFunc("/{channelID}", handleChannel)

	log.Fatal(http.ListenAndServe(":8080", router))
}

func dial(addr string) *grpc.ClientConn {
	conn, err = grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	return conn
}

package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/AntsEclipse/gochat/auth/utils"
	"github.com/AntsEclipse/gochat/protobuf/chat"
	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
	uuid "github.com/nu7hatch/gouuid"
	"github.com/olivere/elastic"
	"github.com/sirupsen/logrus"
)

func (s *server) routes() {
	s.router.Use(authutil.ValidateTokenMiddleware)
	s.router.HandleFunc("/", s.handleChannels()).Methods("GET", "POST")
	s.router.HandleFunc("/{channelID}", s.handleChannel()).Methods("GET", "PUT", "DELETE")
	s.router.HandleFunc("/{channelID}/history", s.handleChannelHistory()).Methods("GET")
	s.router.HandleFunc("/{channelID}/search", s.handleChannelSearch()).Methods("GET")
}

func (s *server) handleChannels() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			/**
			 * @api {get} /channels/ Get a List of Channels
			 * @apiName channelsGet
			 * @apiGroup Channels
			 *
			 * @apiSuccess {Channel[]} channels List of channels
			 */

			channelIDs, _ := redis.Strings(s.redisClient.Do("SMEMBERS", "channels:id"))

			channels := make([]Channel, len(channelIDs))
			for i, v := range channelIDs {
				channelMap, _ := redis.StringMap(s.redisClient.Do("HGETALL", v))
				channels[i] = Channel{
					ID:   v,
					Name: channelMap["name"],
				}
			}
			json.NewEncoder(w).Encode(channels)
		case "POST":
			/**
			 * @api {post} /channels/ Create a New Channel
			 * @apiName channelsCreate
			 * @apiGroup Channels
			 *
			 * @apiParam {string} channelName new channel"s name
			 *
			 * @apiSuccess {Channel} channel the new created channel
			 */
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
			s.chatClient.CreateServer(context.Background(), &chat.Request{ChannelId: channelID})
			s.redisClient.Do("SADD", "channels:id", channelID)
			s.redisClient.Do("HSET", channelID, "name", channelName)
			json.NewEncoder(w).Encode(newChannel)
		}
	}
}

func (s *server) handleChannel() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			http.Error(w, "NOT IMPLEMENTED", http.StatusMethodNotAllowed)
		case "PUT":
			http.Error(w, "NOT IMPLEMENTED", http.StatusMethodNotAllowed)
		case "DELETE":
			http.Error(w, "NOT IMPLEMENTED", http.StatusMethodNotAllowed)
		}
	}
}

func (s *server) handleChannelHistory() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			params := mux.Vars(r)
			channelID := params["channelID"]

			iter := s.cassandraSession.Query(`select * from messages where channel_id = ? and type = 'MESSAGE' limit 100`, channelID).Iter()
			s.log.WithFields(logrus.Fields{
				"count":   iter.NumRows(),
				"channel": channelID,
			}).Info("Executed Channel History Query")
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
	}
}

func (s *server) handleChannelSearch() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			params := mux.Vars(r)
			channelID := params["channelID"]
			err := r.ParseForm()
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			message := r.FormValue("message")
			query := elastic.NewFuzzyQuery("message", message)
			searchResult, err := s.elasticClient.Search().
				Index("messages").
				Query(query).
				From(0).Size(10).
				Pretty(true).
				Do(context.Background())
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			s.log.WithFields(logrus.Fields{
				"message": message,
				"count":   searchResult.TotalHits(),
				"channel": channelID,
				"tookms":  searchResult.TookInMillis,
			}).Info("Executed Channel Search Query")
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
			}
			json.NewEncoder(w).Encode(results)
		}
	}
}

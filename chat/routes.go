package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func (s *server) routes() {
	s.router.HandleFunc("/connect/{channelID}", s.handleChat())
}

func (s *server) handleChat() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		channelID := params["channelID"]
		s.log.WithFields(logrus.Fields{
			"channelID":   channelID,
			"channeluuid": s.uuids[channelID],
		}).Info("Connecting user to server")
		serveWs(s.hubs[s.uuids[channelID]], w, r)
	}
}

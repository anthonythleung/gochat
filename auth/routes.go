package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/sirupsen/logrus"
)

func (s *server) routes() {
	s.router.HandleFunc("/", s.handleAuth())
}

/**
 *
 * @api {get} /auth/:userID Get a users jwt token
 * @apiName get
 * @apiGroup Auth
 *
 * @apiParam  {number} userID A users unique ID.
 *
 * @apiSuccess {string} token A users JWT token
 */
func (s *server) handleAuth() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(100)
		if err != nil {
			panic(err)
		}
		userID, err := strconv.ParseInt(r.PostFormValue("userID"), 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			s.log.WithFields(logrus.Fields{"statusCode": http.StatusBadRequest}).Error(err.Error())
			return
		}
		password := r.PostFormValue("password")
		token := s.requestJWT(userID, password)
		json.NewEncoder(w).Encode(token)
	}
}

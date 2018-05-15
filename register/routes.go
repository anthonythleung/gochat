package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/AntsEclipse/gochat/protobuf/auth"
	"github.com/AntsEclipse/gochat/protobuf/user"
)

func (s *server) routes() {
	s.router.HandleFunc("/", s.handleUserRegister).Methods("POST")
}

func (s *server) handleUserRegister(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		/**
		 * @api {post} /register Request User information
		 * @apiName post
		 * @apiGroup Register
		 *
		 * @apiParam {string} email Users email.
		 * @apiParam {string} password Users password.
		 *
		 * @apiSuccess {String} userid ID of the User.
		 */
		err := r.ParseMultipartForm(100)
		if err != nil {
			panic(err)
		}
		email := r.PostFormValue("email")
		password := r.PostFormValue("password")

		userResult, userErr := s.userClient.CreateUser(context.Background(), &user.Request{Email: email})
		if userErr != nil {
			http.Error(w, userErr.Error(), http.StatusBadRequest)
			return
		}
		authResult, authErr := s.authClient.CreateUser(context.Background(), &auth.Request{UserId: userResult.UserId, Password: password})
		if authErr != nil {
			http.Error(w, authErr.Error(), http.StatusBadRequest)
			return
		}

		json.NewEncoder(w).Encode(RegisterResult{UserID: authResult.UserId})
	}
}

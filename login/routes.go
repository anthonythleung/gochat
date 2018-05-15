package main

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/AntsEclipse/gochat/protobuf/auth"
	"github.com/AntsEclipse/gochat/protobuf/user"
)

func (s *server) routes() {
	s.router.HandleFunc("/", s.handleUserRegister()).Methods("POST")
}

func (s *server) handleUserRegister() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			/**
			 *
			 * @api {post} /login Users login
			 * @apiName post
			 * @apiGroup Login
			 *
			 * @apiParam  {String} email A users email.
			 * @apiParam  {String} password A users password.
			 *
			 * @apiSuccess  {number} userid A users unique id.
			 * @apiSuccess  {string} token A users jwt token.
			 */
			err := r.ParseMultipartForm(100)
			if err != nil {
				panic(err)
			}

			email := r.PostFormValue("email")
			password := r.PostFormValue("password")

			userResult, err := s.userClient.GetUserID(context.Background(), &user.Request{Email: email})
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			authResult, err := s.authClient.GetToken(context.Background(), &auth.Request{UserId: userResult.UserId, Password: password})
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			json.NewEncoder(w).Encode(LoginResult{UserID: userResult.UserId, Token: authResult.Token})
		}
	}
}

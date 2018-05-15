package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/AntsEclipse/gochat/auth/utils"
	"github.com/gorilla/mux"
)

func (s *server) routes() {
	s.router.Use(authutil.ValidateTokenMiddleware)
	s.router.HandleFunc("/", s.handleUsers()).Methods("GET")
	s.router.HandleFunc("/{userID}", s.handleUser()).Methods("GET", "PUT", "DELETE")
}

func (s *server) handleUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			/**
			 * @api {get} /user/:userId Request User information
			 * @apiName userGet
			 * @apiGroup User
			 *
			 * @apiParam {Number} userId Users unique ID.
			 *
			 * @apiSuccess {String} id ID of the User.
			 * @apiSuccess {String} username Username of the User.
			 * @apiSuccess {String} email Email of the User.
			 */
			params := mux.Vars(r)
			id, err := strconv.ParseInt(params["userID"], 10, 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			user, err := s.dao.getUser(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			json.NewEncoder(w).Encode(user)
		case "PUT":
			/**
			 * @api {put} /user/:userId Update User information
			 * @apiName put
			 * @apiGroup User
			 *
			 * @apiParam {Number} userId Users unique ID.
			 * @apiParam {String} username Users new username.
			 *
			 * @apiSuccess {String} id ID of the User.
			 * @apiSuccess {String} username Username of the User.
			 * @apiSuccess {String} email Email of the User.
			 */
			params := mux.Vars(r)
			err := r.ParseMultipartForm(100)
			id, err := strconv.ParseInt(params["userID"], 10, 64)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			username := r.PostFormValue("username")
			user, err := s.dao.updateUser(User{ID: id, Username: username})
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			json.NewEncoder(w).Encode(user)
		case "DELETE":
			http.Error(w, "NOT IMPLEMENTED", http.StatusMethodNotAllowed)
			return
		}
	}
}

func (s *server) handleUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			/**
			 * @api {get} /user/ Request Users information
			 * @apiName usersGet
			 * @apiGroup User
			 *
			 * @apiParam {Number} username Users username.
			 * @apiParam {Number} email Users email.
			 *
			 * @apiSuccess {User[]} users List of user profile
			 * @apiSuccess {String} users.id ID of the User.
			 * @apiSuccess {String} users.username Username of the User.
			 * @apiSuccess {String} users.email Email of the User.
			 */

			err := r.ParseForm()
			email := r.FormValue("email")
			username := r.FormValue("username")
			user := User{Username: username, Email: email}
			users, err := s.dao.getUsers(user)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			json.NewEncoder(w).Encode(users)
		}
	}

}

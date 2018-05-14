package main

import (
	"log"
	"time"

	"github.com/AntsEclipse/gochat/utils"
	"github.com/go-pg/pg"
)

// User ... The main user object
type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email" sql:",unique"`
}

var db *pg.DB

func initDB() {
	helpers.Wait("db:5432")
	db = pg.Connect(&pg.Options{
		User:     "postgres",
		Password: "password",
		Addr:     "db:5432",
	})
	db.OnQueryProcessed(func(event *pg.QueryProcessedEvent) {
		query, err := event.FormattedQuery()
		if err != nil {
			panic(err)
		}
		log.Printf("%s %s", time.Since(event.StartTime), query)
	})
	createSchema(db)
}

func createSchema(db *pg.DB) error {
	for _, model := range []interface{}{&User{}} {
		err := db.CreateTable(model, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func getUser(id int64) (User, error) {
	user := User{ID: id}
	err := db.Select(&user)

	return user, err
}

func findUser(email string) (User, error) {
	var user User
	err := db.Model(&user).
		Where("email = ?", email).
		Limit(1).
		Select()
	return user, err
}

func createUser(email string) (User, error) {
	newUser := &User{
		Email:    email,
		Username: email,
	}

	err := db.Insert(newUser)

	return *newUser, err
}

func updateUser(user User) (User, error) {
	_, err := db.Model(&user).
		Column("username").
		Returning("email").
		Update()
	return user, err
}

func deleteUser(id int64) error {
	user := &User{
		ID: id,
	}

	err := db.Delete(user)

	return err
}

func getUsers(user User) ([]User, error) {
	var users []User
	err := db.Model(&users).
		Where("?0 = '' or email = ?0", user.Email).
		Where("?0 = '' or username = ?0", user.Username).
		Select()
	return users, err
}

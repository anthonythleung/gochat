package main

import (
	"log"
	"time"

	"github.com/go-pg/pg"
)

// UserCredential ... A User Credential Object with UserID and Password
type UserCredential struct {
	ID       int64
	UserID   int64
	Password string
}

var db *pg.DB

func initDB() {
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

// createSchema ... Create UserCredential Database Schema
func createSchema(db *pg.DB) error {
	for _, model := range []interface{}{&UserCredential{}} {
		err := db.CreateTable(model, nil)
		if err != nil {
			return nil
		}
	}
	return nil
}

// CreateUserCredential ... Create a new UserCredential
func createUserCredential(userID int64, password string) UserCredential {
	hash, err := hashPassword(password)
	if err != nil {
		panic(err)
	}
	newUser := &UserCredential{
		UserID:   userID,
		Password: hash,
	}

	err = db.Insert(newUser)

	if err != nil {
		panic(err)
	}

	return *newUser
}

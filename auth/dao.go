package main

import (
	"log"
	"time"

	"github.com/go-pg/pg"
)

type dao struct {
	db *pg.DB
}

// UserCredential ... A User Credential Object with UserID and Password
type UserCredential struct {
	ID       int64
	UserID   int64
	Password string
}

func (d *dao) initDB() {
	d.db = pg.Connect(&pg.Options{
		User:     "postgres",
		Password: "password",
		Addr:     "db:5432",
	})

	d.db.OnQueryProcessed(func(event *pg.QueryProcessedEvent) {
		query, err := event.FormattedQuery()
		if err != nil {
			panic(err)
		}
		log.Printf("%s %s", time.Since(event.StartTime), query)
	})

	d.createSchema()
}

// createSchema ... Create UserCredential Database Schema
func (d *dao) createSchema() error {
	for _, model := range []interface{}{&UserCredential{}} {
		err := d.db.CreateTable(model, nil)
		if err != nil {
			return nil
		}
	}
	return nil
}

// CreateUserCredential ... Create a new UserCredential
func (d *dao) createUserCredential(userID int64, password string) UserCredential {
	hash, err := hashPassword(password)
	if err != nil {
		panic(err)
	}
	newUser := &UserCredential{
		UserID:   userID,
		Password: hash,
	}

	err = d.db.Insert(newUser)

	if err != nil {
		panic(err)
	}

	return *newUser
}

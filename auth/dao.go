package main

import (
	"github.com/AntsEclipse/gochat/utils"
	"github.com/sirupsen/logrus"

	"github.com/go-pg/pg"
)

type dao struct {
	db  *pg.DB
	log *logrus.Entry
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

	d.db.OnQueryProcessed(helpers.CreateQueryLogger(d.log))
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

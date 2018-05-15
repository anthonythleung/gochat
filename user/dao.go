package main

import (
	"log"
	"time"

	"github.com/AntsEclipse/gochat/utils"
	"github.com/go-pg/pg"
)

type dao struct {
	db *pg.DB
}

// User ... The main user object
type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email" sql:",unique"`
}

func (d *dao) initDB() {
	logger := helpers.Logger("user")
	helpers.Wait("db:5432", logger)
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

func (d *dao) createSchema() error {
	for _, model := range []interface{}{&User{}} {
		err := d.db.CreateTable(model, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func (d *dao) getUser(id int64) (User, error) {
	user := User{ID: id}
	err := d.db.Select(&user)

	return user, err
}

func (d *dao) findUser(email string) (User, error) {
	var user User
	err := d.db.Model(&user).
		Where("email = ?", email).
		Limit(1).
		Select()
	return user, err
}

func (d *dao) createUser(email string) (User, error) {
	newUser := &User{
		Email:    email,
		Username: email,
	}

	err := d.db.Insert(newUser)

	return *newUser, err
}

func (d *dao) updateUser(user User) (User, error) {
	_, err := d.db.Model(&user).
		Column("username").
		Returning("email").
		Update()
	return user, err
}

func (d *dao) deleteUser(id int64) error {
	user := &User{
		ID: id,
	}

	err := d.db.Delete(user)

	return err
}

func (d *dao) getUsers(user User) ([]User, error) {
	var users []User
	err := d.db.Model(&users).
		Where("?0 = '' or email = ?0", user.Email).
		Where("?0 = '' or username = ?0", user.Username).
		Select()
	return users, err
}

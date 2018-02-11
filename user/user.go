package main

import "github.com/go-pg/pg"

// User ... The main user object
type User struct {
	ID        int64
	Username  string
	createdAt int64
	updatedAt int64
}

var db *pg.DB

func initDB() {
	db = pg.Connect(&pg.Options{
		User:     "postgres",
		Password: "password",
		Addr:     "db:5432",
	})
	createSchema(db)
}

func createSchema(db *pg.DB) error {
	for _, model := range []interface{}{&User{}} {
		err := db.CreateTable(model, nil)
		if err != nil {
			return nil
		}
	}
	return nil
}

func getUser(id int64) User {
	user := User{ID: id}
	err := db.Select(&user)
	if err != nil {
		panic(err)
	}

	return user
}

func findUser(username string) User {
	var user User
	err := db.Model(&user).
		Where("username = ?", username).
		Limit(1).
		Select()
	if err != nil {
		panic(err)
	}
	return user
}

func createUser(username string) User {
	newUser := &User{
		Username: username,
	}

	err := db.Insert(newUser)

	if err != nil {
		panic(err)
	}

	return *newUser
}

func deleteUser(id int64) {
	user := &User{
		ID: id,
	}

	err := db.Delete(user)

	if err != nil {
		panic(err)
	}
}

func getUsers() []User {
	var users []User
	err := db.Model(&users).Select()
	if err != nil {
		panic(err)
	}
	return users
}

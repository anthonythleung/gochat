package main

import (
	"crypto/rsa"
	"io/ioutil"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

const (
	privateKeyPath = "/go/src/github.com/AntsEclipse/gochat/auth/gochat.rsa"
	publicKeyPath  = "/go/src/github.com/AntsEclipse/gochat/auth/gochat.rsa.pub"
)

var (
	verifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
)

// InitKeys ... initialize public and private key
func initKeys() {
	signBytes, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		panic(err)
	}

	signKey, err = jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		panic(err)
	}

	verifyBytes, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		panic(err)
	}

	verifyKey, err = jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		panic(err)
	}

}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Token ... A JWT Token
type Token struct {
	Token string `json:"token"`
}

// RequestJWT ... Returns a JWT token
func requestJWT(userID int64, password string) Token {
	var user UserCredential
	err := db.Model(&user).
		Where("user_id = ?", userID).
		Limit(1).
		Select()
	if err != nil {
		panic(err)
	}

	if !checkPasswordHash(password, user.Password) {
		return Token{}
	}

	token := jwt.New(jwt.SigningMethodRS256)
	claims := make(jwt.MapClaims)
	claims["userID"] = user.UserID
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	token.Claims = claims
	tokenString, err := token.SignedString(signKey)

	if err != nil {
		panic(err)
	}

	return Token{tokenString}
}

package main

import (
	"crypto/rsa"
	"io/ioutil"

	jwt "github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

const (
	privateKeyPath = "/auth/gochat.rsa"
	publicKeyPath  = "/auth/gochat.rsa.pub"
)

type crypto struct {
	verifyKey *rsa.PublicKey
	signKey   *rsa.PrivateKey
}

// InitKeys ... initialize public and private key
func (c *crypto) initKeys() {
	signBytes, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		panic(err)
	}

	signKey, err := jwt.ParseRSAPrivateKeyFromPEM(signBytes)
	if err != nil {
		panic(err)
	}

	verifyBytes, err := ioutil.ReadFile(publicKeyPath)
	if err != nil {
		panic(err)
	}

	verifyKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		panic(err)
	}

	c.signKey = signKey
	c.verifyKey = verifyKey
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

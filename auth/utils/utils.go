package authutil

import (
	"io/ioutil"
	"net/http"

	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
)

const publicKeyPath = "/auth/gochat.rsa.pub"

var verifyBytes []byte
var initialized = false

// ValidateTokenMiddleware ... Valide Token Middleware for Protected API
func ValidateTokenMiddleware(next http.Handler) http.Handler {
	if !initialized {
		verifyBytes, _ = ioutil.ReadFile(publicKeyPath)
	}

	verifyKey, err := jwt.ParseRSAPublicKeyFromPEM(verifyBytes)
	if err != nil {
		panic(err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := request.ParseFromRequest(r, request.AuthorizationHeaderExtractor,
			func(token *jwt.Token) (interface{}, error) {
				return verifyKey, nil
			})

		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized Access", http.StatusUnauthorized)
		} else {
			next.ServeHTTP(w, r)
		}
	})
}

package authutil

import (
	"fmt"
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

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, "Unauthorized Access")
		} else {
			if token.Valid {
				next.ServeHTTP(w, r)
			} else {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprintf(w, "Unauthorized Access")
			}
		}
	})
}

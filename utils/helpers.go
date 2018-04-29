package helpers

import (
	"github.com/rs/cors"
	"google.golang.org/grpc"
	"net/http"
)

// Dial ... connect to grpc
func Dial(addr string) *grpc.ClientConn {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	return conn
}

// CorsHandler ... handle cors
func CorsHandler(router http.Handler) http.Handler {
	c := cors.New(cors.Options{
		AllowedOrigins:     []string{"*"},
		AllowCredentials:   true,
		AllowedMethods:     []string{"GET", "PUT", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:     []string{"Authorization"},
		OptionsPassthrough: false,
	})

	return c.Handler(router)
}

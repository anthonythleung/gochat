package helpers

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/rs/cors"
	"google.golang.org/grpc"
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

// Wait ... Wait for a port to be avilable
func Wait(addr string) {
	for {
		conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
		if err == nil {
			conn.Close()
			fmt.Printf("%s is now avilable\n", addr)
			return
		}
		fmt.Printf("waiting for %s ...\n", addr)
		time.Sleep(5 * time.Second)
	}
}

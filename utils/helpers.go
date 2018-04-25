package helpers

import (
	"google.golang.org/grpc"
)

func Dial(addr string) *grpc.ClientConn {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	return conn
}

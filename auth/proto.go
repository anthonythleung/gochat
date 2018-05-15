package main

import (
	"context"

	"github.com/AntsEclipse/gochat/protobuf/auth"
)

func (s *server) CreateUser(ctx context.Context, req *auth.Request) (*auth.Result, error) {
	newUser := s.dao.createUserCredential(req.UserId, req.Password)
	return &auth.Result{UserId: newUser.UserID}, nil
}

func (s *server) GetToken(ctx context.Context, req *auth.Request) (*auth.Token, error) {
	newToken := s.requestJWT(req.GetUserId(), req.GetPassword())
	return &auth.Token{Token: newToken.Token}, nil
}

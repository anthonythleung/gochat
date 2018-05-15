package main

import (
	"context"

	"github.com/AntsEclipse/gochat/protobuf/user"
)

func (s *server) CreateUser(ctx context.Context, req *user.Request) (*user.Result, error) {
	newUser, err := s.dao.createUser(req.GetEmail())
	return &user.Result{UserId: newUser.ID}, err
}

func (s *server) GetUserID(ctx context.Context, req *user.Request) (*user.Result, error) {
	result, err := s.dao.findUser(req.GetEmail())
	return &user.Result{UserId: result.ID}, err
}

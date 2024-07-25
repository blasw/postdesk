package server

import (
	"context"
	"encoding/json"
	"errors"
	"math/rand"
	"users-service/pb"
	"users-service/storage/entities"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/emptypb"
)

type signUpDto struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Server) SignUp(ctx context.Context, req *pb.SignUpRequest) (*pb.SignUpResponse, error) {
	var dto signUpDto

	if err := json.Unmarshal(req.Json, &dto); err != nil {
		logrus.WithError(err).Error("Unable to unmarshal JSON")
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		logrus.WithError(err).Error("Unable to generate hash from password")
		return nil, err
	}

	userID, err := s.storage.CreateUser(&entities.User{
		Email:    dto.Email,
		Username: dto.Username,
		Password: string(hashedPassword),
	})

	if err != nil {
		logrus.WithError(err).Error("Unable to create user")
		return nil, err
	}

	return &pb.SignUpResponse{
		UserId: userID,
	}, nil
}

type signInDto struct {
	Username string `json:"email"`
	Password string `json:"password"`
}

func (s *Server) SignIn(ctx context.Context, req *pb.SignInRequest) (*pb.SignInResponse, error) {
	var dto signInDto

	if err := json.Unmarshal(req.Json, &dto); err != nil {
		return nil, err
	}

	user, err := s.storage.GetUserByUsername(dto.Username)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(dto.Password)); err != nil {
		return nil, errors.New("Wrong password")
	}

	return &pb.SignInResponse{
		UserId: user.ID,
	}, nil
}

func (s *Server) CheckHealth(ctx context.Context, req *emptypb.Empty) (*pb.CheckHealthResponse, error) {
	val := rand.Float32()

	return &pb.CheckHealthResponse{
		Health: val,
	}, nil
}

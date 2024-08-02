package server

import (
	"context"
	"encoding/json"
	"math/rand"
	"net/http"
	"users-service/pb"
	"users-service/storage/entities"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	INTERNAL_ERR = http.StatusInternalServerError
	BAD_REQ_ERR  = http.StatusBadRequest
	NO_ERR       = http.StatusOK
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
		return &pb.SignUpResponse{
			UserId: -1,
			Status: INTERNAL_ERR,
		}, nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	if err != nil {
		logrus.WithError(err).Error("Unable to generate hash from password")
		return &pb.SignUpResponse{
			UserId: -1,
			Status: INTERNAL_ERR,
		}, nil
	}

	userID, err := s.storage.CreateUser(&entities.User{
		Email:    dto.Email,
		Username: dto.Username,
		Password: string(hashedPassword),
	})

	if err != nil {
		return &pb.SignUpResponse{
			UserId: -1,
			Status: BAD_REQ_ERR,
		}, nil
	}

	return &pb.SignUpResponse{
		UserId: userID,
		Status: NO_ERR,
	}, nil
}

type signInDto struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *Server) SignIn(ctx context.Context, req *pb.SignInRequest) (*pb.SignInResponse, error) {
	var dto signInDto

	if err := json.Unmarshal(req.Json, &dto); err != nil {
		logrus.WithError(err).Error("Unable to unmarshal JSON")
		return &pb.SignInResponse{
			UserId: -1,
			Status: INTERNAL_ERR,
		}, nil
	}

	user, err := s.storage.GetUserByUsername(dto.Username)
	if err != nil {
		return &pb.SignInResponse{
			UserId: -1,
			Status: BAD_REQ_ERR,
		}, nil
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(dto.Password)); err != nil {
		return &pb.SignInResponse{
			UserId: -1,
			Status: BAD_REQ_ERR,
		}, nil
	}

	return &pb.SignInResponse{
		UserId: user.ID,
		Status: NO_ERR,
	}, nil
}

func (s *Server) CheckHealth(ctx context.Context, req *emptypb.Empty) (*pb.CheckHealthResponse, error) {
	val := rand.Float32()

	return &pb.CheckHealthResponse{
		Health: val,
	}, nil
}

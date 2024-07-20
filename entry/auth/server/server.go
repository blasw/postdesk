package server

import (
	"auth/pb"
	"auth/tokenizer"
	"net"

	"google.golang.org/grpc"
)

type Storage interface {
	SetToken(userID uint64, token string) error
	GetToken(userID uint64) (string, error)
}

type Server struct {
	pb.UnimplementedAuthServiceServer

	port      string
	tokenizer tokenizer.Tokenizer
	storage   Storage
}

func NewServer(port string, storage Storage, tokenizer tokenizer.Tokenizer) *Server {
	return &Server{
		port:      port,
		storage:   storage,
		tokenizer: tokenizer,
	}
}

func (s *Server) Run() error {
	lis, err := net.Listen("tcp", s.port)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	pb.RegisterAuthServiceServer(grpcServer, s)

	if err := grpcServer.Serve(lis); err != nil {
		return err
	}

	return nil
}

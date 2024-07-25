package server

import (
	"net"
	"users-service/pb"
	"users-service/storage"

	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedUsersServiceServer

	port    string
	storage *storage.PostgreStorage
}

func NewServer(storage *storage.PostgreStorage, port string) *Server {
	return &Server{
		port:    port,
		storage: storage,
	}
}

func (s *Server) Run() error {
	lis, err := net.Listen("tcp", s.port)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	pb.RegisterUsersServiceServer(grpcServer, s)

	if err := grpcServer.Serve(lis); err != nil {
		return err
	}

	return nil
}

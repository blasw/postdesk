package users

import (
	"loadbalancer/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func NewUsersService(usersAddr string) (pb.UsersServiceClient, error) {
	options := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	client, err := grpc.NewClient(usersAddr, options...)
	if err != nil {
		return nil, err
	}

	usersClient := pb.NewUsersServiceClient(client)

	return usersClient, nil
}

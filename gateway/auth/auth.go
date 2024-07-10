package auth

import (
	"gateway/pb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Connect(authAddr string) (pb.AuthServiceClient, error) {
	options := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	client, err := grpc.NewClient(authAddr, options...)
	if err != nil {
		return nil, err
	}

	authClient := pb.NewAuthServiceClient(client)

	return authClient, nil
}

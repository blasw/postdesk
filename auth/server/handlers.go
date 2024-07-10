package server

import (
	"auth/pb"
	"auth/tokenizer"
	"context"

	"github.com/sirupsen/logrus"
)

// Verify checks if provided tokens are valid and creates new access and refresh tokens if needed.
func (s *Server) Verify(ctx context.Context, req *pb.VerifyRequest) (*pb.VerifyResponse, error) {
	logrus.Debug("Starting verification of access and refresh tokens")
	claims, err := s.tokenizer.ParseAccessToken(req.AccessToken)
	if err == nil {
		logrus.Debug("Access token is valid")
		return &pb.VerifyResponse{Valid: true}, nil
	}

	logrus.Debug("access token is invalid")
	//Access token is invalid or expired
	if claims == nil {
		//Access token is invalid
		logrus.Debug("Access token is invalid")
		return &pb.VerifyResponse{Valid: false}, nil
	}

	err = s.tokenizer.ParseRefreshToken(req.RefreshToken)
	if err != nil {
		// Access token is expired and refresh token is invalid
		logrus.WithError(err).Error("Both access is expired, refresh token is invalid or expired")
		return &pb.VerifyResponse{Valid: false}, nil
	}

	// Access token is expired and refresh token is valid
	storedRefreshToken, err := s.storage.GetToken(claims.UserID)
	if err != nil {
		logrus.WithError(err).Error("Error accured while trying to get token from storage")
		return nil, err
	}

	if storedRefreshToken != req.RefreshToken {
		logrus.Debug("Refresh tokens do not match")
		return &pb.VerifyResponse{Valid: false}, nil
	}

	//Refresh token is valid
	newAccessToken, err := s.tokenizer.NewAccessToken(tokenizer.UserClaims{UserID: claims.UserID})
	if err != nil {
		logrus.WithError(err).Error("Error accured while trying to create new access token")
		return nil, err
	}

	newRefreshToken, err := s.tokenizer.NewRefreshToken()
	if err != nil {
		logrus.WithError(err).Error("Error accured while trying to create new refresh token")
		return nil, err
	}

	err = s.storage.SetToken(claims.UserID, newRefreshToken)
	if err != nil {
		logrus.WithError(err).Error("Error accured while trying to set token in storage")
		return nil, err
	}

	return &pb.VerifyResponse{Valid: true, AccessToken: newAccessToken, RefreshToken: newRefreshToken}, nil
}

func (s *Server) CreateTokens(ctx context.Context, req *pb.CreateTokensRequest) (*pb.CreateTokensResponse, error) {
	accessToken, err := s.tokenizer.NewAccessToken(tokenizer.UserClaims{UserID: req.UserId})
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.tokenizer.NewRefreshToken()
	if err != nil {
		return nil, err
	}

	resp := &pb.CreateTokensResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	err = s.storage.SetToken(req.UserId, refreshToken)
	if err != nil {
		logrus.WithError(err).Error("Error accured while trying to set token in storage")
		return nil, err
	}

	return resp, nil
}

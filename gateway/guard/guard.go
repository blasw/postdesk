package guard

import (
	"context"
	"gateway/pb"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Guard struct {
	auth pb.AuthServiceClient
}

func NewGuard(client pb.AuthServiceClient) *Guard {
	return &Guard{
		auth: client,
	}
}

func (g *Guard) VerifyTokens(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	accessToken, err := c.Cookie("access_token")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := g.auth.Verify(ctx, &pb.VerifyRequest{AccessToken: accessToken, RefreshToken: refreshToken})
	if err != nil {
		logrus.WithError(err).Error("Error accured while trying to verify tokens in auth service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !resp.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid tokens"})
		return
	}

	c.Set("userID", resp.UserId)

	if resp.AccessToken != "" && resp.RefreshToken != "" {
		c.SetCookie("access_token", resp.AccessToken, 3600, "/", "localhost", false, true)
		c.SetCookie("refresh_token", resp.RefreshToken, 3600, "/", "localhost", false, true)
		// c.JSON(200, gin.H{"access_token": resp.AccessToken, "refresh_token": resp.RefreshToken})
		c.Next()
		return
	}

	c.Next()
}

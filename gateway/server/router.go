package server

import (
	"context"
	"gateway/pb"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Router struct {
	auth pb.AuthServiceClient
}

func (r *Router) CreatePost(c *gin.Context) {
	// accessToken, _ := c.Cookie("access_token")

	// refreshToken, err := c.Cookie("refresh_token")
	// if err != nil {
	// 	c.Status(http.StatusUnauthorized)
	// }

	//TODO: get user info

	//TODO: public a message for creating a post to broker with user info
}

func (r *Router) DeletePost(c *gin.Context) {}

func (r *Router) GetPosts(c *gin.Context) {}

func (r *Router) SignUp(c *gin.Context) {}

func (r *Router) SignIn(c *gin.Context) {}

func (r *Router) SignOut(c *gin.Context) {}

func (r *Router) LikePost(c *gin.Context) {}

func (r *Router) UnlikePost(c *gin.Context) {}

func (r *Router) GetUserInfo(c *gin.Context) {}

// TESTING ONLY
func (r *Router) VerifyTokens(c *gin.Context) {
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

	logrus.Debug("tokens acquired")

	//---------------------------------------------------
	// TODO: check if error is server side or client side AND add setting new cookies if needed
	resp, err := r.auth.Verify(ctx, &pb.VerifyRequest{AccessToken: accessToken, RefreshToken: refreshToken})
	if err != nil {
		logrus.WithError(err).Error("Error accured while trying to verify tokens in auth service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if !resp.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid tokens"})
		return
	}

	if resp.AccessToken != "" && resp.RefreshToken != "" {
		//TODO: set new cookies
		c.JSON(200, gin.H{"access_token": resp.AccessToken, "refresh_token": resp.RefreshToken})
	}

	c.Status(200)
	//---------------------------------------------------
}

type CreateTokensDto struct {
	UserID uint64 `json:"user_id" binding:"required"`
}

func (r *Router) CreateTokens(c *gin.Context) {
	var dto CreateTokensDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	logrus.Debug("userID acquired, creating tokens")

	resp, err := r.auth.CreateTokens(ctx, &pb.CreateTokensRequest{UserId: dto.UserID})
	if err != nil {
		logrus.WithError(err).Error("Error accured while trying to create tokens in auth service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"access_token": resp.AccessToken, "refresh_token": resp.RefreshToken})
}

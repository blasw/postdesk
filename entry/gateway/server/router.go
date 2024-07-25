package server

import (
	"context"
	"encoding/json"
	"fmt"
	"gateway/broker"
	"gateway/pb"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

//TODO: needs refactoring

type Router struct {
	auth   pb.AuthServiceClient
	broker broker.Broker
}

type CreatePostDto struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
}

func (r *Router) CreatePost(c *gin.Context) {
	var dto CreatePostDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetInt64("userID")

	data := struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		AuthorID string `json:"author_id"`
	}{
		Title:    dto.Title,
		Content:  dto.Content,
		AuthorID: fmt.Sprint(userID),
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		logrus.WithError(err).Error("Error accured while marshaling data")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error accured"})
		return
	}

	//TODO: FIX THIS _
	_, err = r.broker.Produce("create_post", jsonBytes)
	if err != nil {
		logrus.WithError(err).Error("Error accured while publishing to kafka")
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

type DeletePostDto struct {
	PostID string `json:"post_id" binding:"required"`
}

func (r *Router) DeletePost(c *gin.Context) {
	var dto DeletePostDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetInt64("userID")

	data := struct {
		PostID string `json:"post_id"`
		UserID string `json:"user_id"`
	}{
		PostID: dto.PostID,
		UserID: fmt.Sprint(userID),
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		logrus.WithError(err).Error("Error accured while marshaling data")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error accured"})
		return
	}

	//TODO FIX THIS _
	_, err = r.broker.Produce("delete_post", jsonBytes)
	if err != nil {
		logrus.WithError(err).Error("Error accured while publishing to kafka")
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

type GetPostsDto struct {
	Sort   string `form:"sort"`
	Amount int    `form:"amount"`
	Page   int    `form:"page"`
}

func (r *Router) GetPosts(c *gin.Context) {
	var dto GetPostsDto
	if err := c.ShouldBindQuery(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data := struct {
		Sort   string `json:"sort"`
		Amount int    `json:"amount"`
		Page   int    `json:"page"`
	}{
		Sort:   dto.Sort,
		Amount: dto.Amount,
		Page:   dto.Page,
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		logrus.WithError(err).Error("Error accured while marshaling data")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error accured"})
		return
	}

	//TODO FIX THIS _
	_, err = r.broker.Produce("get_posts", jsonBytes)
	if err != nil {
		logrus.WithError(err).Error("Error accured while publishing to kafka")
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

type SignUpDto struct {
	Email    string `json:"email" binding:"required"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserResponse struct {
	UserID int64 `json:"user_id"`
}

func (r *Router) SignUp(c *gin.Context) {
	var dto SignUpDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data := struct {
		Email    string `json:"email"`
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Email:    dto.Email,
		Username: dto.Username,
		Password: dto.Password,
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		logrus.WithError(err).Error("Error accured while marshaling data")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error accured"})
		return
	}

	logrus.Debug("Publishing to kafka")
	uniqueID, err := r.broker.Produce("sign_up", jsonBytes)
	if err != nil {
		logrus.WithError(err).Error("Error accured while publishing to kafka")
		c.Status(http.StatusInternalServerError)
		return
	}
	kafkaCtx, kafkaCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer kafkaCancel()

	respBytes, err := r.broker.WaitForMessage(kafkaCtx, "sign_up_response", uniqueID)
	if err != nil {
		logrus.WithError(err).Error("Error accured while waiting for message")
		c.Status(http.StatusInternalServerError)
		return
	}

	var resp UserResponse
	err = json.Unmarshal(respBytes, &resp)
	if err != nil {
		logrus.WithError(err).Error("Error accured while unmarshaling data")
		c.Status(http.StatusInternalServerError)
		return
	}

	authCtx, authCancel := context.WithTimeout(context.Background(), time.Second)
	defer authCancel()

	tokens, err := r.auth.CreateTokens(authCtx, &pb.CreateTokensRequest{UserId: uint64(resp.UserID)})
	if err != nil {
		logrus.WithError(err).Error("Error accured while trying to create tokens in auth service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie("access_token", tokens.AccessToken, 0, "/", "", false, true)
	c.SetCookie("refresh_token", tokens.RefreshToken, 0, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{"access_token": tokens.AccessToken, "refresh_token": tokens.RefreshToken, "user_id": resp.UserID})
}

type SignInDto struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (r *Router) SignIn(c *gin.Context) {
	var dto SignInDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	jsonBytes, err := json.Marshal(dto)
	if err != nil {
		logrus.WithError(err).Error("Error accured while marshaling data")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error accured"})
		return
	}

	uniqueID, err := r.broker.Produce("sign_in", jsonBytes)
	if err != nil {
		logrus.WithError(err).Error("Error accured while publishing to kafka")
		c.Status(http.StatusInternalServerError)
		return
	}

	kafkaCtx, kafkaCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer kafkaCancel()

	respBytes, err := r.broker.WaitForMessage(kafkaCtx, "sign_in_response", uniqueID)
	if err != nil {
		logrus.WithError(err).Error("Error accured while waiting for message")
		c.Status(http.StatusInternalServerError)
		return
	}

	var resp UserResponse
	err = json.Unmarshal(respBytes, &resp)
	if err != nil {
		logrus.WithError(err).Error("Error accured while unmarshaling data")
		c.Status(http.StatusInternalServerError)
		return
	}

	authCtx, authCancel := context.WithTimeout(context.Background(), time.Second)
	defer authCancel()

	tokens, err := r.auth.CreateTokens(authCtx, &pb.CreateTokensRequest{UserId: uint64(resp.UserID)})
	if err != nil {
		logrus.WithError(err).Error("Error accured while trying to create tokens in auth service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.SetCookie("access_token", tokens.AccessToken, 0, "/", "", false, true)
	c.SetCookie("refresh_token", tokens.RefreshToken, 0, "/", "", false, true)

	c.Status(http.StatusOK)
}

// TODO
func (r *Router) SignOut(c *gin.Context) {
	c.Status(http.StatusOK)
}

type PostDto struct {
	PostID string `form:"post_id" binding:"required"`
}

func (r *Router) LikePost(c *gin.Context) {
	var dto PostDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetInt64("userID")

	data := struct {
		PostID string `json:"post_id"`
		UserID string `json:"user_id"`
	}{
		PostID: dto.PostID,
		UserID: fmt.Sprint(userID),
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		logrus.WithError(err).Error("Error accured while marshaling data")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error accured"})
		return
	}

	//TODO FIX THIS _
	_, err = r.broker.Produce("like_post", jsonBytes)
	if err != nil {
		logrus.WithError(err).Error("Error accured while publishing to kafka")
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

func (r *Router) UnlikePost(c *gin.Context) {
	var dto PostDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetInt64("userID")

	data := struct {
		PostID string `json:"post_id"`
		UserID string `json:"user_id"`
	}{
		PostID: dto.PostID,
		UserID: fmt.Sprint(userID),
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		logrus.WithError(err).Error("Error accured while marshaling data")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error accured"})
		return
	}

	// TODO FIX THIS _
	_, err = r.broker.Produce("unlike_post", jsonBytes)
	if err != nil {
		logrus.WithError(err).Error("Error accured while publishing to kafka")
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

type GetUserInfoDto struct {
	UserID int64 `json:"user_id" binding:"required"`
}

func (r *Router) GetUserInfo(c *gin.Context) {
	var dto GetUserInfoDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetInt64("userID")

	if userID != dto.UserID {
		c.Status(http.StatusForbidden)
		return
	}

	data := struct {
		UserID string `json:"user_id"`
	}{
		UserID: fmt.Sprint(dto.UserID),
	}

	jsonBytes, err := json.Marshal(data)
	if err != nil {
		logrus.WithError(err).Error("Error accured while marshaling data")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error accured"})
		return
	}

	// TODO FIX THIS _
	_, err = r.broker.Produce("get_user_info", jsonBytes)
	if err != nil {
		logrus.WithError(err).Error("Error accured while publishing to kafka")
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
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

	resp, err := r.auth.CreateTokens(ctx, &pb.CreateTokensRequest{UserId: dto.UserID})
	if err != nil {
		logrus.WithError(err).Error("Error accured while trying to create tokens in auth service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"access_token": resp.AccessToken, "refresh_token": resp.RefreshToken})
}

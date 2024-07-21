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

	message := broker.KafkaMessage{
		Topic: "create_post",
		Data:  jsonBytes,
	}

	err = r.broker.Publish(message)
	if err != nil {
		logrus.WithError(err).Error("Error accured while publishing to kafka")
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
	//TODO: return created post
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

	message := broker.KafkaMessage{
		Topic: "delete_post",
		Data:  jsonBytes,
	}

	err = r.broker.Publish(message)
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

	message := broker.KafkaMessage{
		Topic: "get_posts",
		Data:  jsonBytes,
	}

	err = r.broker.Publish(message)
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

	message := broker.KafkaMessage{
		Topic: "sign_up",
		Data:  jsonBytes,
	}

	err = r.broker.Publish(message)
	if err != nil {
		logrus.WithError(err).Error("Error accured while publishing to kafka")
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Status(http.StatusOK)
}

type SignInDto struct {
	Username string `json:"email" binding:"required"`
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

	message := broker.KafkaMessage{
		Topic: "sign_in",
		Data:  jsonBytes,
	}

	err = r.broker.Publish(message)
	if err != nil {
		logrus.WithError(err).Error("Error accured while publishing to kafka")
		c.Status(http.StatusInternalServerError)
		return
	}

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

	message := broker.KafkaMessage{
		Topic: "like_post",
		Data:  jsonBytes,
	}

	err = r.broker.Publish(message)
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

	message := broker.KafkaMessage{
		Topic: "unlike_post",
		Data:  jsonBytes,
	}

	err = r.broker.Publish(message)
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

	message := broker.KafkaMessage{
		Topic: "get_user_info",
		Data:  jsonBytes,
	}

	err = r.broker.Publish(message)
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

package server

import (
	"gateway/broker"
	"gateway/guard"
	"gateway/pb"

	"github.com/gin-gonic/gin"
)

type GatewayServer struct {
	port   string
	engine *gin.Engine
	auth   pb.AuthServiceClient
	broker broker.Broker
}

func New(port string, authService pb.AuthServiceClient, broker broker.Broker) (*GatewayServer, error) {
	return &GatewayServer{
		port:   port,
		engine: gin.Default(),
		auth:   authService,
		broker: broker,
	}, nil
}

func (s *GatewayServer) SetupRoutes() {
	guard := guard.NewGuard(s.auth)

	r := &Router{
		auth:   s.auth,
		broker: s.broker,
	}

	// protected
	s.engine.POST("/posts/create", guard.VerifyTokens, r.CreatePost)
	// protected
	s.engine.DELETE("/posts/delete", guard.VerifyTokens, r.DeletePost)
	// public
	s.engine.GET("/posts/get", r.GetPosts)

	// public
	s.engine.POST("/users/signin", r.SignIn)
	// public
	s.engine.POST("/users/signup", r.SignUp)
	// protected
	s.engine.POST("/users/signout", guard.VerifyTokens, r.SignOut)
	// protected
	s.engine.GET("/users/info", guard.VerifyTokens, r.GetUserInfo)

	// protected
	s.engine.POST("/posts/like", guard.VerifyTokens, r.LikePost)
	//protected
	s.engine.POST("/posts/unlike", guard.VerifyTokens, r.UnlikePost)

	//TESTING ONLY
	s.engine.POST("/tokens/create", r.CreateTokens)
}

func (s *GatewayServer) Run() error {
	s.SetupRoutes()
	if err := s.engine.Run(s.port); err != nil {
		return err
	}

	return nil
}

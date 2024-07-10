package server

import (
	"gateway/pb"

	"github.com/gin-gonic/gin"
)

type GatewayServer struct {
	port   string
	engine *gin.Engine
	auth   pb.AuthServiceClient
}

func New(port string, authService pb.AuthServiceClient) (*GatewayServer, error) {
	return &GatewayServer{
		port:   port,
		engine: gin.Default(),
		auth:   authService,
	}, nil
}

func (s *GatewayServer) SetupRoutes() {
	r := &Router{
		auth: s.auth,
	}

	// protected
	s.engine.POST("/posts/create", r.CreatePost)
	// protected
	s.engine.DELETE("/posts/delete", r.DeletePost)
	// public
	s.engine.GET("/posts/get", r.GetPosts)

	// public
	s.engine.POST("/users/signin", r.SignIn)
	// public
	s.engine.POST("/users/signup", r.SignUp)
	// protected
	s.engine.POST("/users/signout", r.SignOut)
	// protected
	s.engine.GET("/users/info", r.GetUserInfo)

	// protected
	s.engine.POST("/posts/like", r.LikePost)
	//protected
	s.engine.POST("/posts/unlike", r.UnlikePost)

	//TESTING ONLY

	s.engine.GET("/tokens/verify", r.VerifyTokens)

	s.engine.POST("/tokens/create", r.CreateTokens)
}

func (s *GatewayServer) Run() error {
	s.SetupRoutes()
	if err := s.engine.Run(s.port); err != nil {
		return err
	}

	return nil
}

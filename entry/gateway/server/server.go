package server

import (
	"gateway/broker"
	"gateway/guard"
	"gateway/middleware"
	"gateway/pb"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	// metrics
	s.engine.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// protected
	s.engine.POST("/posts/create", middleware.AddMetric("create_post"), guard.VerifyTokens, r.CreatePost)
	// protected
	s.engine.DELETE("/posts/delete", middleware.AddMetric("delete_post"), guard.VerifyTokens, r.DeletePost)
	// public
	s.engine.GET("/posts/get", middleware.AddMetric("get_posts"), r.GetPosts)

	// public
	s.engine.POST("/users/signin", middleware.AddMetric("sign_in"), r.SignIn)
	// public
	s.engine.POST("/users/signup", middleware.AddMetric("sign_up"), r.SignUp)
	// protected
	s.engine.POST("/users/signout", middleware.AddMetric("sign_out"), guard.VerifyTokens, r.SignOut)
	// protected
	s.engine.GET("/users/info", middleware.AddMetric("get_user_info"), guard.VerifyTokens, r.GetUserInfo)

	// protected
	s.engine.POST("/posts/like", middleware.AddMetric("like_post"), guard.VerifyTokens, r.LikePost)
	//protected
	s.engine.POST("/posts/unlike", middleware.AddMetric("unlike_post"), guard.VerifyTokens, r.UnlikePost)

	//TESTING ONLY
	s.engine.POST("/tokens/create", middleware.AddMetric("create_tokens"), r.CreateTokens)
}

func (s *GatewayServer) Run() error {
	s.SetupRoutes()
	if err := s.engine.Run(s.port); err != nil {
		return err
	}

	return nil
}

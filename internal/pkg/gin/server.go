package gin

import (
	"time"

	"github.com/gin-gonic/gin"
)

type Config struct {
	Port         string
	Mode         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type Server struct {
	engine *gin.Engine
	config *Config
}

func New(config *Config) *Server {
	gin.SetMode(config.Mode)
	engine := gin.New()

	return &Server{
		engine: engine,
		config: config,
	}
}

func (s *Server) Engine() *gin.Engine {
	return s.engine
}

func (s *Server) Use(middleware ...gin.HandlerFunc) {
	s.engine.Use(middleware...)
}

func (s *Server) Run() error {
	s.engine.Use(gin.Recovery())
	return s.engine.Run(":" + s.config.Port)
}

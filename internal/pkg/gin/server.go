package gin

import (
	"context"
	"net/http"
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
	engine     *gin.Engine
	config     *Config
	httpServer *http.Server
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
	return s.server().ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}

	return s.httpServer.Shutdown(ctx)
}

func (s *Server) server() *http.Server {
	if s.httpServer == nil {
		s.httpServer = &http.Server{
			Addr:         ":" + s.config.Port,
			Handler:      s.engine,
			ReadTimeout:  s.config.ReadTimeout,
			WriteTimeout: s.config.WriteTimeout,
		}
	}

	return s.httpServer
}

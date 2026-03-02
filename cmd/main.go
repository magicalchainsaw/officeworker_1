package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"officeworker/internal/pkg/gin"
	"officeworker/internal/pkg/middleware"

	"go.uber.org/zap"
)

func main() {
	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("Starting officeworker server...")

	ginConfig := &gin.Config{
		Port:         "8080",
		Mode:         "debug",
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	server := gin.New(ginConfig)
	
	server.Use(
		middleware.CORS(),
		middleware.Logger(middleware.LoggerConfig{Logger: logger}),
		middleware.Recovery(middleware.RecoveryConfig{Logger: logger}),
	)

	router := gin.NewRouterGroup(server.Engine())
	router.SetupRoutes()

	go func() {
		if err := server.Run(); err != nil {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	logger.Info("Server started on port 8080")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	<-ch
	logger.Info("Shutting down server gracefully...")
	
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	
	<-shutdownCtx.Done()
	logger.Info("Server stopped")
}

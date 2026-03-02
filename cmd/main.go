package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"officeworker/internal/config"
	"officeworker/internal/pkg/gin"
	"officeworker/internal/pkg/middleware"
	"officeworker/internal/repository"
	"officeworker/models"

	"go.uber.org/zap"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info("Starting officeworker server...")

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	db, err := repository.NewMySQL(&repository.DatabaseConfig{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		DBName:          cfg.Database.DBName,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	})
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}

	if err := models.AutoMigrate(db); err != nil {
		logger.Fatal("Failed to migrate database", zap.Error(err))
	}
	logger.Info("Database migrated successfully")

	ginConfig := &gin.Config{
		Port:         cfg.Server.Port,
		Mode:         cfg.Server.Mode,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
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

	logger.Info("Server started", zap.String("port", cfg.Server.Port))

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)

	<-ch
	logger.Info("Shutting down server gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}

	<-shutdownCtx.Done()
	logger.Info("Server stopped")
}

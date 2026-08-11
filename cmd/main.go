package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/abeselom-personal/go-ai-service/internal/config"
	"github.com/abeselom-personal/go-ai-service/internal/database"
	models "github.com/abeselom-personal/go-ai-service/internal/model"
	"github.com/abeselom-personal/go-ai-service/internal/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	cfg, err := config.LoadConfig("./config")
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	db, err := database.NewPostgresDB(database.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		DBName:          cfg.Database.Name,
		SSLMode:         cfg.Database.SSLMode,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	})
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}

	if cfg.Database.MigrationEnabled {
		if err := db.AutoMigrate(&models.AIUsageLog{}, &models.RateLimit{}, &models.SystemPrompt{}); err != nil {
			logger.Fatal("failed to migrate database", zap.Error(err))
		}
	}

	redisClient, err := database.NewRedisClient(config.RedisConfig{
		Host:     cfg.Redis.Host,
		Port:     cfg.Redis.Port,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	if err != nil {
		logger.Fatal("Redis connection failed", zap.Error(err))
	}

	router := gin.Default()
	routes.RegisterRoutes(router, db, cfg, redisClient)

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("Starting server", zap.Int("port", cfg.Server.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed to start", zap.Error(err))
		}
	}()

	<-ctx.Done()
	logger.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	sqlDB, _ := db.DB()
	if sqlDB != nil {
		_ = sqlDB.Close()
	}
	_ = redisClient.Close()

	logger.Info("Server exited")
}

package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/abeselom-personal/go-ai-service/internal/config"
	"github.com/abeselom-personal/go-ai-service/internal/database"
	models "github.com/abeselom-personal/go-ai-service/internal/model"
	"github.com/abeselom-personal/go-ai-service/internal/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}
	cfg, err := config.LoadConfig("./config")

	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Setup database connection
	db, err := database.NewPostgresDB(database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.Name,
		SSLMode:  cfg.Database.SSLMode,
	})
	if err != nil {
		logger.Fatal("failed to connect to database", zap.Error(err))
	}

	// Auto migrate if enabled
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
	// Initialize Gin router
	router := gin.Default()

	routes.RegisterRoutes(router, db, cfg, redisClient)

	// Start server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	logger.Info("Starting server", zap.Int("port", cfg.Server.Port))

	if err := server.ListenAndServe(); err != nil {
		logger.Fatal("server failed to start", zap.Error(err))
	}
}

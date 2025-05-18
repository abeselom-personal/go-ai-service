package main

import (
	"fmt"
	"net/http"

	"github.com/abeselom-personal/go-ai-service/internal/config"
	"github.com/abeselom-personal/go-ai-service/internal/database"
	models "github.com/abeselom-personal/go-ai-service/internal/model"
	"github.com/abeselom-personal/go-ai-service/internal/repository"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("./config")
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	// Initialize logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// Setup database connection
	db, err := database.NewPostgresDB(database.Config{
		Host: cfg.Database.Host,
		//fix this string conversion issue
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
		if err := db.AutoMigrate(&models.Provider{}, &models.Model{}); err != nil {
			logger.Fatal("failed to migrate database", zap.Error(err))
		}
	}
	//inititalize repository
	_ = repository.NewProviderRepository(db)
	_ = repository.NewModelRepository(db)

	// Initialize Gin router
	router := gin.Default()

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

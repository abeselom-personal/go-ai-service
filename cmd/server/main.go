// cmd/main.go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/abeselom-personal/go-ai-service/internal/config"
	migrations "github.com/abeselom-personal/go-ai-service/internal/migration"
	"github.com/abeselom-personal/go-ai-service/pkg/logger"
	"github.com/gin-gonic/gin"
	"google.golang.org/genai"
	"gorm.io/gorm"
)

func main() {
	logg := logger.NewLogger()

	db, err := config.InitDBWithLogger(logg)
	if err != nil {
		logg.Error("Failed to initialize database: " + err.Error())
		return
	}
	defer config.Close(db)

	r := createRouter(db, logg)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logg.Error("Server failed: " + err.Error())
		}
	}()
	logg.Info("Server started on :8080")

	<-done
	logg.Info("Starting graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logg.Error("Server shutdown error: " + err.Error())
	}
	logg.Info("Server stopped")
}

func createRouter(db *gorm.DB, log *logger.Logger) *gin.Engine {
	r := gin.New()
	r.Use(ginLogger(log))
	r.Use(dbInjectMiddleware(db))
	migrations.RunMigrations(db)
	r.GET("/health", healthCheckHandler(db))
	r.GET("/ai", aiHandler)

	return r
}

func ginLogger(l *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		l.Info(fmt.Sprintf("%s %s %d %s",
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			time.Since(start),
		))
	}
}

func dbInjectMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	}
}

func healthCheckHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy"})
			return
		}
		if err := sqlDB.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "unhealthy"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	}
}

func aiHandler(c *gin.Context) {
	apiKey := os.Getenv("GENAI_API_KEY")
	if apiKey == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "API key not set"})
		return
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.0-flash",
		genai.Text("Explain how AI works in a few words"),
		nil,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"response": result.Text()})
}

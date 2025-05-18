# Source Code Context

Generated on: 2025-05-18T19:12:45Z

## Repository Overview
- Total Files: 11
- Total Size: 17938 bytes

## Directory Structure
```
config/
  config.go
context/
  images/
controller/
  system_prompt_controller.go
database/
  database.go
model/
  ai_usage.go
  rate_limit.go
  system_prompt.go
repository/
  repository.go
  system_prompt_repository.go
  system_prompt_repository_test.go
routes/
  routes.go
service/
  system_prompt_service.go

```

## File Contents


### File: config/config.go

```go
package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig
	Database  DatabaseConfig
	Security  SecurityConfig
	Defaults  DefaultConfig
	Logging   LoggingConfig
	RateLimit RateLimitConfig
}

type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

type DatabaseConfig struct {
	Host             string        `mapstructure:"host"`
	Port             int           `mapstructure:"port"`
	User             string        `mapstructure:"user"`
	Password         string        `mapstructure:"password"`
	Name             string        `mapstructure:"name"`
	SSLMode          string        `mapstructure:"ssl_mode"`
	MaxIdleConns     int           `mapstructure:"max_idle_conns"`
	MaxOpenConns     int           `mapstructure:"max_open_conns"`
	ConnMaxLifetime  time.Duration `mapstructure:"conn_max_lifetime"`
	MigrationEnabled bool          `mapstructure:"migration_enabled"`
}

type SecurityConfig struct {
	EncryptionKey        string `mapstructure:"encryption_key"`
	EncryptionKeyVersion int    `mapstructure:"encryption_key_version"`
}

type DefaultConfig struct {
	Provider  string           `mapstructure:"provider"`
	Model     string           `mapstructure:"model"`
	Providers []ProviderConfig `mapstructure:"providers"` // Changed from "default_providers"
}

type ProviderConfig struct {
	Name    string        `mapstructure:"name"`
	BaseURL string        `mapstructure:"base_url"`
	APIKey  string        `mapstructure:"api_key"`
	Default bool          `mapstructure:"default"`
	Models  []ModelConfig `mapstructure:"models"`
}

type ModelConfig struct {
	Name       string `mapstructure:"name"`
	Parameters string `mapstructure:"parameters"`
	Config     string `mapstructure:"config"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type RateLimitConfig struct {
	Enabled     bool     `mapstructure:"enabled"`
	Requests    float64  `mapstructure:"requests"`
	Window      string   `mapstructure:"window"`
	IPWhitelist []string `mapstructure:"ip_whitelist"`
}

func LoadConfig(path string) (*Config, error) {
	v := viper.New()

	// Set default values
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", 15*time.Second)
	v.SetDefault("server.write_timeout", 15*time.Second)
	v.SetDefault("server.idle_timeout", 60*time.Second)

	v.SetDefault("database.port", 5432)
	v.SetDefault("database.ssl_mode", "disable")
	v.SetDefault("database.max_idle_conns", 10)
	v.SetDefault("database.max_open_conns", 100)
	v.SetDefault("database.conn_max_lifetime", time.Hour)
	v.SetDefault("database.migration_enabled", true)

	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")

	v.SetDefault("rate_limit.enabled", false)
	v.SetDefault("rate_limit.requests", 100)
	v.SetDefault("rate_limit.window", "1m")

	// Bind environment variables to config paths
	_ = v.BindEnv("server.port", "PORT")
	_ = v.BindEnv("server.read_timeout", "READ_TIMEOUT")
	_ = v.BindEnv("server.write_timeout", "WRITE_TIMEOUT")
	_ = v.BindEnv("server.idle_timeout", "IDLE_TIMEOUT")

	_ = v.BindEnv("database.host", "DB_HOST")
	_ = v.BindEnv("database.port", "DB_PORT")
	_ = v.BindEnv("database.user", "DB_USER")
	_ = v.BindEnv("database.password", "DB_PASSWORD")
	_ = v.BindEnv("database.name", "DB_NAME")
	_ = v.BindEnv("database.ssl_mode", "DB_SSL_MODE")
	_ = v.BindEnv("database.max_idle_conns", "DB_MAX_IDLE_CONNS")
	_ = v.BindEnv("database.max_open_conns", "DB_MAX_OPEN_CONNS")
	_ = v.BindEnv("database.conn_max_lifetime", "DB_CONN_MAX_LIFETIME")
	_ = v.BindEnv("database.migration_enabled", "DB_MIGRATION_ENABLED")

	_ = v.BindEnv("security.encryption_key", "ENCRYPTION_KEY")
	_ = v.BindEnv("security.encryption_key_version", "ENCRYPTION_KEY_VERSION")

	_ = v.BindEnv("defaults.provider", "DEFAULT_PROVIDER")
	_ = v.BindEnv("defaults.model", "DEFAULT_MODEL")

	_ = v.BindEnv("logging.level", "LOG_LEVEL")
	_ = v.BindEnv("logging.format", "LOG_FORMAT")

	_ = v.BindEnv("rate_limit.enabled", "RATE_LIMIT_ENABLED")
	_ = v.BindEnv("rate_limit.requests", "RATE_LIMIT_REQUESTS")
	_ = v.BindEnv("rate_limit.window", "RATE_LIMIT_WINDOW")
	_ = v.BindEnv("rate_limit.ip_whitelist", "RATE_LIMIT_IP_WHITELIST")
	// Configuration sources
	v.AddConfigPath(path)
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AutomaticEnv()

	// Bind environment variables
	_ = v.BindEnv("security.encryption_key", "ENCRYPTION_KEY")
	_ = v.BindEnv("database.password", "DB_PASSWORD")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	if err := validateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	for i := range cfg.Defaults.Providers {
		cfg.Defaults.Providers[i].APIKey = os.Getenv(fmt.Sprintf("%s_API_KEY", strings.ToUpper(cfg.Defaults.Providers[i].Name)))
	}
	return &cfg, nil
}

func validateConfig(cfg *Config) error {
	if cfg.Security.EncryptionKey == "" {
		return fmt.Errorf("encryption key is required")
	}

	if len(cfg.Security.EncryptionKey) != 32 {
		return fmt.Errorf("encryption key must be 32 bytes")
	}

	if cfg.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if cfg.Database.User == "" {
		return fmt.Errorf("database user is required")
	}

	if cfg.Database.Name == "" {
		return fmt.Errorf("database name is required")
	}

	if cfg.Defaults.Provider == "" && len(cfg.Defaults.Providers) > 0 {
		cfg.Defaults.Provider = cfg.Defaults.Providers[0].Name
	}

	return nil
}

```





### File: controller/system_prompt_controller.go

```go
// controller/system_prompt_controller.go
package controller

import (
	"net/http"

	"github.com/abeselom-personal/go-ai-service/internal/service"
	"github.com/gin-gonic/gin"
)

type SystemPromptController struct {
	svc *service.SystemPromptService
}

func NewSystemPromptController(svc *service.SystemPromptService) *SystemPromptController {
	return &SystemPromptController{svc}
}

func (c *SystemPromptController) Create(ctx *gin.Context) {
	var req struct {
		ModuleName   string `json:"module_name" binding:"required"`
		Provider     string `json:"provider" binding:"required"`
		SystemPrompt string `json:"system_prompt" binding:"required"`
		UserPrompt   string `json:"user_prompt" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	prompt, err := c.svc.Create(ctx, req.ModuleName, req.Provider, req.SystemPrompt, req.UserPrompt)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, prompt)
}

func (c *SystemPromptController) Get(ctx *gin.Context) {
	hash := ctx.Param("hash")
	prompt, err := c.svc.Get(ctx, hash)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Prompt not found"})
		return
	}
	ctx.JSON(http.StatusOK, prompt)
}

func (c *SystemPromptController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var req struct {
		SystemPrompt string `json:"system_prompt" binding:"required"`
		UserPrompt   string `json:"user_prompt" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := c.svc.Update(ctx, id, req.SystemPrompt, req.UserPrompt); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusOK)
}

func (c *SystemPromptController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	if err := c.svc.Delete(ctx, id); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.Status(http.StatusNoContent)
}

func (c *SystemPromptController) Send(ctx *gin.Context) {
	var req struct {
		ModuleName   string `json:"module_name" binding:"required"`
		Provider     string `json:"provider" binding:"required"`
		SystemPrompt string `json:"system_prompt" binding:"required"`
		UserPrompt   string `json:"user_prompt" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	prompt, err := c.svc.SendPrompt(ctx, req.ModuleName, req.Provider, req.SystemPrompt, req.UserPrompt)
	if err != nil {
		ctx.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, prompt)
}

```





### File: database/database.go

```go
package database

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func NewPostgresDB(cfg Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Connection pool settings
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Enable UUID extension
	db.Exec(`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`)

	return db, nil
}

```





### File: model/ai_usage.go

```go
package models

import (
	"time"

	"github.com/google/uuid"
)

type AIUsageLog struct {
	ID         uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ModuleName string    `gorm:"index;not null"`
	Provider   string    `gorm:"index;not null"`
	PromptHash string    `gorm:"index;not null"`
	UsedAt     time.Time `gorm:"autoCreateTime"`
}

```





### File: model/rate_limit.go

```go
package models

import (
	"github.com/google/uuid"
)

type RateLimit struct {
	ID          uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ModuleName  string    `gorm:"index;not null"`
	Provider    string    `gorm:"index;not null"`
	MaxRequests int       `gorm:"not null"`
	PerSeconds  int       `gorm:"not null"`
}

```





### File: model/system_prompt.go

```go
// models/system_prompt.go
package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SystemPrompt struct {
	ID           uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ModuleName   string    `gorm:"index;not null"`
	Provider     string    `gorm:"index;not null"`
	SystemPrompt string    `gorm:"type:text;not null"`
	UserPrompt   string    `gorm:"type:text;not null"`
	PromptHash   string    `gorm:"uniqueIndex;not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

```





### File: repository/repository.go

```go
// internal/repository/repository.go
package repository

import (
	"context"

	"gorm.io/gorm"
)

type contextKey string

const (
	contextTxKey contextKey = "db_transaction"
)

// Used in repository methods to get transaction from context
func getDB(ctx context.Context, defaultDB *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(contextTxKey).(*gorm.DB); ok {
		return tx
	}
	return defaultDB
}

```





### File: repository/system_prompt_repository.go

```go
// internal/repository/system_prompt_repo.go
package repository

import (
	"context"

	models "github.com/abeselom-personal/go-ai-service/internal/model"
	"gorm.io/gorm"
)

type SystemPromptRepo struct {
	db *gorm.DB
}

func NewSystemPromptRepo(db *gorm.DB) *SystemPromptRepo {
	return &SystemPromptRepo{db}
}

func (r *SystemPromptRepo) WithTransaction(ctx context.Context, fn func(context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, contextTxKey, tx)
		return fn(txCtx)
	})
}

func (r *SystemPromptRepo) Create(ctx context.Context, sp *models.SystemPrompt) error {
	return getDB(ctx, r.db).WithContext(ctx).Create(sp).Error
}

func (r *SystemPromptRepo) GetByHash(ctx context.Context, hash string) (*models.SystemPrompt, error) {
	var sp models.SystemPrompt
	err := getDB(ctx, r.db).WithContext(ctx).Where("prompt_hash = ?", hash).First(&sp).Error
	return &sp, err
}

func (r *SystemPromptRepo) Update(ctx context.Context, sp *models.SystemPrompt) error {
	return getDB(ctx, r.db).WithContext(ctx).Save(sp).Error
}

func (r *SystemPromptRepo) Delete(ctx context.Context, id string) error {
	return getDB(ctx, r.db).WithContext(ctx).Delete(&models.SystemPrompt{}, "id = ?", id).Error
}

func (r *SystemPromptRepo) List(ctx context.Context) ([]models.SystemPrompt, error) {
	var prompts []models.SystemPrompt
	err := getDB(ctx, r.db).WithContext(ctx).Find(&prompts).Error
	return prompts, err
}

```





### File: repository/system_prompt_repository_test.go

```go
// internal/repository/system_prompt_repository_test.go
package repository_test

import (
	"context"
	"testing"

	models "github.com/abeselom-personal/go-ai-service/internal/model"
	"github.com/abeselom-personal/go-ai-service/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)
	err = db.AutoMigrate(&models.SystemPrompt{})
	assert.NoError(t, err)
	return db
}

func TestSystemPromptRepo_CRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := repository.NewSystemPromptRepo(db)
	ctx := context.Background()

	// Create
	prompt := &models.SystemPrompt{
		ID:           uuid.New(),
		ModuleName:   "test-module",
		Provider:     "ChatGPT",
		SystemPrompt: "You are a helper.",
		UserPrompt:   "What's the weather?",
		PromptHash:   "abc123",
	}
	err := repo.Create(ctx, prompt)
	assert.NoError(t, err)

	// Get
	fetched, err := repo.GetByHash(ctx, "abc123")
	assert.NoError(t, err)
	assert.Equal(t, prompt.ID, fetched.ID)

	// Update
	fetched.SystemPrompt = "You are a very helpful assistant."
	err = repo.Update(ctx, fetched)
	assert.NoError(t, err)

	updated, err := repo.GetByHash(ctx, "abc123")
	assert.NoError(t, err)
	assert.Equal(t, "You are a very helpful assistant.", updated.SystemPrompt)

	// List
	all, err := repo.List(ctx)
	assert.NoError(t, err)
	assert.Len(t, all, 1)

	// Delete
	err = repo.Delete(ctx, prompt.ID.String())
	assert.NoError(t, err)

	_, err = repo.GetByHash(ctx, "abc123")
	assert.Error(t, err)
}

```





### File: routes/routes.go

```go
// routes/routes.go
package routes

import (
	"github.com/abeselom-personal/go-ai-service/internal/config"
	"github.com/abeselom-personal/go-ai-service/internal/controller"
	"github.com/abeselom-personal/go-ai-service/internal/repository"
	"github.com/abeselom-personal/go-ai-service/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB, cfg *config.Config) {

	// register routes
	repo := repository.NewSystemPromptRepo(db)
	svc := service.NewSystemPromptService(db, repo)
	ctrl := controller.NewSystemPromptController(svc)

	api := r.Group("/api/system-prompts")
	{
		api.POST("/", ctrl.Create)
		api.GET("/:hash", ctrl.Get)
		api.PUT("/:id", ctrl.Update)
		api.DELETE("/:id", ctrl.Delete)
		api.POST("/send", ctrl.Send)
	}
}

```





### File: service/system_prompt_service.go

```go
package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	models "github.com/abeselom-personal/go-ai-service/internal/model"
	"github.com/abeselom-personal/go-ai-service/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SystemPromptService struct {
	repo *repository.SystemPromptRepo
	db   *gorm.DB
}

func NewSystemPromptService(db *gorm.DB, repo *repository.SystemPromptRepo) *SystemPromptService {
	return &SystemPromptService{repo: repo, db: db}
}

func hashPrompt(systemPrompt, userPrompt, moduleName string) string {
	raw := systemPrompt + userPrompt + moduleName
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (s *SystemPromptService) Create(ctx context.Context, module, provider, sys, user string) (*models.SystemPrompt, error) {
	hash := hashPrompt(sys, user, module)
	existing, _ := s.repo.GetByHash(ctx, hash)
	if existing != nil && existing.ID != uuid.Nil {
		return existing, nil
	}

	sp := &models.SystemPrompt{
		ModuleName:   module,
		Provider:     provider,
		SystemPrompt: sys,
		UserPrompt:   user,
		PromptHash:   hash,
	}
	err := s.repo.Create(ctx, sp)
	return sp, err
}

func (s *SystemPromptService) Get(ctx context.Context, hash string) (*models.SystemPrompt, error) {
	return s.repo.GetByHash(ctx, hash)
}

func (s *SystemPromptService) Update(ctx context.Context, id string, sys, user string) error {
	var sp models.SystemPrompt
	if err := s.db.WithContext(ctx).First(&sp, "id = ?", id).Error; err != nil {
		return err
	}
	sp.SystemPrompt = sys
	sp.UserPrompt = user
	sp.PromptHash = hashPrompt(sys, user, sp.ModuleName)
	return s.repo.Update(ctx, &sp)
}

func (s *SystemPromptService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *SystemPromptService) SendPrompt(ctx context.Context, module, provider, sys, user string) (*models.SystemPrompt, error) {
	hash := hashPrompt(sys, user, module)
	sp, err := s.repo.GetByHash(ctx, hash)
	if err != nil {
		sp, err = s.Create(ctx, module, provider, sys, user)
		if err != nil {
			return nil, err
		}
	}

	limit := models.RateLimit{}
	err = s.db.WithContext(ctx).
		Where("module_name = ? AND provider = ?", module, provider).
		First(&limit).Error
	if err != nil {
		return nil, err
	}

	var count int64
	start := time.Now().Add(-time.Duration(limit.PerSeconds) * time.Second)
	s.db.Model(&models.AIUsageLog{}).
		Where("module_name = ? AND provider = ? AND used_at >= ?", module, provider, start).
		Count(&count)

	if count >= int64(limit.MaxRequests) {
		return nil, errors.New("rate limit exceeded")
	}

	s.db.Create(&models.AIUsageLog{
		ModuleName: module,
		Provider:   provider,
		PromptHash: hash,
	})

	return sp, nil
}

```





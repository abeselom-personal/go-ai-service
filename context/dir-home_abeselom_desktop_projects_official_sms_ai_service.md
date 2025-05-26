# Source Code Context

Generated on: 2025-05-25T19:13:02Z

## Repository Overview
- Total Files: 23
- Total Size: 75671 bytes

## Directory Structure
```
.air.toml
.env
Dockerfile
Makefile
README.md
cmd/
  main.go
config/
  config.yml
context/
  images/
coverage.out
diagram.puml
docker-compose.yml
internal/
  config/
    config.go
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
result.log
templates/
  index.html
tmp/
  main

```

## File Contents


### File: .air.toml

```
root = "."
tmp_dir = "tmp"

[build]
  bin = "tmp/main"
  cmd = "go build -o tmp/main ./cmd"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata"]
  exclude_regex = ["_test.go"]
  include_ext = ["go", "tpl", "tmpl", "html"]
  log = "build-errors.log"
  stop_on_error = true

[color]
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  time = false

[misc]
  clean_on_exit = true

[screen]
  clear_on_rebuild = true

```





### File: .env

```
PORT=8080
GIN_MODE=release
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=mydb
DB_HOST=ai-database
DB_PORT=5432

OPENAI_API_KEY=your-key
DEEPSEEK_API_KEY=sk-264958b6af834253ab2b82f96acf7ba5
GEMINI_API_KEY=AIzaSyAjOEqn9i_uGY5-_emX5RHSCvQ4NBI-3MQ

ENCRYPTION_KEY=3f8a1c2d9e7b4f6130a5d6c8b2e4f1a9

```





### File: Dockerfile

```
FROM golang:1.24-alpine

WORKDIR /app

RUN go install github.com/air-verse/air@latest

COPY go.mod go.sum ./
RUN go mod download

COPY . .

EXPOSE 8080

CMD ["air", "-c", ".air.toml"]

```





### File: Makefile

```
.PHONY: test

TEST_RESULT=result.log

test:
	go test -v ./... | tee $(TEST_RESULT)
	@echo ""
	@echo "==== TEST SUMMARY ===="
	@grep -E "PASS|FAIL" $(TEST_RESULT)

clean:
	rm -f $(TEST_RESULT)

```





### File: README.md

```markdown
# Universal AI Prompt Service

A global AI prompt microservice built with Go (Gin), supporting REST, gRPC, and WebSocket communication. This service acts as a centralized, reusable prompt and response manager with intelligent caching and real-time communication capabilities.

## Features

- **Universal System Prompt Store**  
  Store and manage multiple system prompts for different tenants or contexts.

- **Prompt Request Caching with Hashing**  
  Avoid duplicate AI requests by generating an MD5 hash for identical input and reusing cached responses from the database.

- **AI Response Proxying**  
  Seamlessly integrates with OpenAI or similar APIs to forward prompts and retrieve completions.

- **WebSocket Chat Support**  
  Real-time bidirectional WebSocket communication with per-user rooms for live chat experiences.

- **REST & gRPC Support**  
  Expose endpoints for both REST and gRPC protocols to support any type of client.

- **Multi-Tenant Ready**  
  Designed with tenant-level prompt isolation for SaaS environments.

## Database Schema

- `prompts` — Stores reusable system prompts.
- `prompt_cache` — Caches hashed prompt+input requests and AI responses.
- `users` — Tracks users for WebSocket room management.

## Example Use Cases

- AI chatbots
- Prompt-as-a-service for internal teams
- Multi-tenant SaaS AI integrations
- Real-time WebSocket chat apps with AI

## Tech Stack

- Go (Gin Framework)
- WebSockets
- PostgreSQL
- gRPC
- REST (JSON API)
- Optional: Redis for faster cache lookup

## Future Additions

- Admin dashboard for prompt management  
- TTL-based cache cleanup  
- Rate limiting and auth middlewares  

```





### File: cmd/main.go

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/abeselom-personal/go-ai-service/internal/config"
	"github.com/abeselom-personal/go-ai-service/internal/database"
	models "github.com/abeselom-personal/go-ai-service/internal/model"
	"github.com/abeselom-personal/go-ai-service/internal/routes"
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

	// Initialize Gin router
	router := gin.Default()

	routes.RegisterRoutes(router, db, cfg)

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

```





### File: config/config.yml

```yaml
server:
  port: 8080
  read_timeout: 15s
  write_timeout: 15s
  idle_timeout: 60s

database:
  host: postgres
  port: 5432
  user: admin
  password: secret
  name: providers
  ssl_mode: disable
  max_idle_conns: 10
  max_open_conns: 100
  conn_max_lifetime: 1h
  migration_enabled: true 

security:
  encryption_key: ""
  encryption_key_version: 1

defaults:
  provider: "gemini"
  model: "gemini-2.0-flash"
  providers:
    - name: "gemini"
      base_url: "https://generativelanguage.googleapis.com/v1beta/models/"
      api_key: "${GEMINI_API_KEY}"
      auth_method: "query_param"
      models:
        - name: "gemini-2.0-flash"
          parameters: '{"temperature": 0.9, "maxOutputTokens": 100}'
          config: |
            {
              "contents": [{
                "parts": [
                  {"text": "{{.SystemPrompt}}"},
                  {"text": "{{.UserPrompt}}"}
                ]
              }]
            }
          response_path: "candidates.0.content.parts.0.text"

logging:
  level: info
  format: json

rate_limit:
  enabled: true
  requests: 100
  window: "1m"
  ip_whitelist:
    - "127.0.0.1"

```





### File: coverage.out

```
mode: set
github.com/abeselom-personal/go-ai-service/cmd/main.go:15.13,18.16 2 0
github.com/abeselom-personal/go-ai-service/cmd/main.go:18.16,19.55 1 0
github.com/abeselom-personal/go-ai-service/cmd/main.go:23.2,36.16 4 0
github.com/abeselom-personal/go-ai-service/cmd/main.go:36.16,38.3 1 0
github.com/abeselom-personal/go-ai-service/cmd/main.go:41.2,41.35 1 0
github.com/abeselom-personal/go-ai-service/cmd/main.go:41.35,42.77 1 0
github.com/abeselom-personal/go-ai-service/cmd/main.go:42.77,44.4 1 0
github.com/abeselom-personal/go-ai-service/cmd/main.go:47.2,62.48 6 0
github.com/abeselom-personal/go-ai-service/cmd/main.go:62.48,64.3 1 0
github.com/abeselom-personal/go-ai-service/internal/database/database.go:21.50,28.16 3 0
github.com/abeselom-personal/go-ai-service/internal/database/database.go:28.16,30.3 1 0
github.com/abeselom-personal/go-ai-service/internal/database/database.go:32.2,33.16 2 0
github.com/abeselom-personal/go-ai-service/internal/database/database.go:33.16,35.3 1 0
github.com/abeselom-personal/go-ai-service/internal/database/database.go:38.2,45.16 5 0
github.com/abeselom-personal/go-ai-service/internal/config/config.go:76.47,139.41 47 0
github.com/abeselom-personal/go-ai-service/internal/config/config.go:139.41,140.56 1 0
github.com/abeselom-personal/go-ai-service/internal/config/config.go:140.56,142.4 1 0
github.com/abeselom-personal/go-ai-service/internal/config/config.go:145.2,146.42 2 0
github.com/abeselom-personal/go-ai-service/internal/config/config.go:146.42,148.3 1 0
github.com/abeselom-personal/go-ai-service/internal/config/config.go:149.2,149.45 1 0
github.com/abeselom-personal/go-ai-service/internal/config/config.go:149.45,151.3 1 0
github.com/abeselom-personal/go-ai-service/internal/config/config.go:153.2,153.18 1 0
github.com/abeselom-personal/go-ai-service/internal/config/config.go:156.40,157.38 1 0
github.com/abeselom-personal/go-ai-service/internal/config/config.go:157.38,159.3 1 0
github.com/abeselom-personal/go-ai-service/internal/config/config.go:161.2,161.43 1 0
github.com/abeselom-personal/go-ai-service/internal/config/config.go:161.43,163.3 1 0
github.com/abeselom-personal/go-ai-service/internal/config/config.go:165.2,165.29 1 0
github.com/abeselom-personal/go-ai-service/internal/config/config.go:165.29,167.3 1 0
github.com/abeselom-personal/go-ai-service/internal/config/config.go:169.2,169.29 1 0
github.com/abeselom-personal/go-ai-service/internal/config/config.go:169.29,171.3 1 0
github.com/abeselom-personal/go-ai-service/internal/config/config.go:173.2,173.29 1 0
github.com/abeselom-personal/go-ai-service/internal/config/config.go:173.29,175.3 1 0
github.com/abeselom-personal/go-ai-service/internal/config/config.go:177.2,177.68 1 0
github.com/abeselom-personal/go-ai-service/internal/config/config.go:177.68,179.3 1 0
github.com/abeselom-personal/go-ai-service/internal/config/config.go:181.2,181.12 1 0
github.com/abeselom-personal/go-ai-service/internal/model/model.go:21.49,22.22 1 0
github.com/abeselom-personal/go-ai-service/internal/model/model.go:22.22,24.3 1 0
github.com/abeselom-personal/go-ai-service/internal/model/model.go:25.2,25.12 1 0
github.com/abeselom-personal/go-ai-service/internal/model/provider.go:22.52,23.22 1 0
github.com/abeselom-personal/go-ai-service/internal/model/provider.go:23.22,25.3 1 0
github.com/abeselom-personal/go-ai-service/internal/model/provider.go:26.2,26.12 1 0
github.com/abeselom-personal/go-ai-service/internal/repository/mocks/mock_provider_repository.go:33.81,37.2 3 0
github.com/abeselom-personal/go-ai-service/internal/repository/mocks/mock_provider_repository.go:40.79,42.2 1 0
github.com/abeselom-personal/go-ai-service/internal/repository/mocks/mock_provider_repository.go:45.94,50.2 4 0
github.com/abeselom-personal/go-ai-service/internal/repository/mocks/mock_provider_repository.go:53.86,56.2 2 0
github.com/abeselom-personal/go-ai-service/internal/repository/mocks/mock_provider_repository.go:59.79,64.2 4 0
github.com/abeselom-personal/go-ai-service/internal/repository/mocks/mock_provider_repository.go:67.80,70.2 2 0
github.com/abeselom-personal/go-ai-service/internal/repository/mocks/mock_provider_repository.go:73.99,79.2 5 0
github.com/abeselom-personal/go-ai-service/internal/repository/mocks/mock_provider_repository.go:82.81,85.2 2 0
github.com/abeselom-personal/go-ai-service/internal/repository/mocks/mock_provider_repository.go:88.103,94.2 5 0
github.com/abeselom-personal/go-ai-service/internal/repository/mocks/mock_provider_repository.go:97.85,100.2 2 0
github.com/abeselom-personal/go-ai-service/internal/repository/mocks/mock_provider_repository.go:103.89,109.2 5 0
github.com/abeselom-personal/go-ai-service/internal/repository/mocks/mock_provider_repository.go:112.77,115.2 2 0
github.com/abeselom-personal/go-ai-service/internal/repository/mocks/mock_provider_repository.go:118.94,123.2 4 0
github.com/abeselom-personal/go-ai-service/internal/repository/mocks/mock_provider_repository.go:126.86,129.2 2 0
github.com/abeselom-personal/go-ai-service/internal/repository/mocks/mock_provider_repository.go:132.109,137.2 4 0
github.com/abeselom-personal/go-ai-service/internal/repository/mocks/mock_provider_repository.go:140.89,143.2 2 0
github.com/abeselom-personal/go-ai-service/internal/repository/model_repository.go:32.54,34.2 1 1
github.com/abeselom-personal/go-ai-service/internal/repository/model_repository.go:36.82,37.34 1 1
github.com/abeselom-personal/go-ai-service/internal/repository/model_repository.go:37.34,39.3 1 1
github.com/abeselom-personal/go-ai-service/internal/repository/model_repository.go:41.2,42.43 2 1
github.com/abeselom-personal/go-ai-service/internal/repository/model_repository.go:42.43,44.3 1 1
github.com/abeselom-personal/go-ai-service/internal/repository/model_repository.go:45.2,45.12 1 1
github.com/abeselom-personal/go-ai-service/internal/repository/model_repository.go:48.90,54.44 3 1
github.com/abeselom-personal/go-ai-service/internal/repository/model_repository.go:54.44,56.3 1 1
github.com/abeselom-personal/go-ai-service/internal/repository/model_repository.go:57.2,57.20 1 1
github.com/abeselom-personal/go-ai-service/internal/repository/model_repository.go:60.117,66.44 3 1
github.com/abeselom-personal/go-ai-service/internal/repository/model_repository.go:66.44,68.3 1 1
github.com/abeselom-personal/go-ai-service/internal/repository/model_repository.go:69.2,69.20 1 1
github.com/abeselom-personal/go-ai-service/internal/repository/model_repository.go:72.82,73.26 1 1
github.com/abeselom-personal/go-ai-service/internal/repository/model_repository.go:73.26,75.3 1 0
github.com/abeselom-personal/go-ai-service/internal/repository/model_repository.go:77.2,82.25 2 1
github.com/abeselom-personal/go-ai-service/internal/repository/model_repository.go:82.25,84.3 1 1
github.com/abeselom-personal/go-ai-service/internal/repository/model_repository.go:85.2,85.30 1 1
github.com/abeselom-personal/go-ai-service/internal/repository/model_repository.go:85.30,87.3 1 0
github.com/abeselom-personal/go-ai-service/internal/repository/model_repository.go:88.2,88.12 1 1
github.com/abeselom-personal/go-ai-service/internal/repository/model_repository.go:91.72,96.25 2 1
github.com/abeselom-personal/go-ai-service/internal/repository/model_repository.go:96.25,98.3 1 1
github.com/abeselom-personal/go-ai-service/internal/repository/model_repository.go:99.2,99.30 1 1
github.com/abeselom-personal/go-ai-service/internal/repository/model_repository.go:99.30,101.3 1 0
github.com/abeselom-personal/go-ai-service/internal/repository/model_repository.go:102.2,102.12 1 1
github.com/abeselom-personal/go-ai-service/internal/repository/model_repository.go:105.106,111.2 3 1
github.com/abeselom-personal/go-ai-service/internal/repository/provider_repository.go:33.60,35.2 1 1
github.com/abeselom-personal/go-ai-service/internal/repository/provider_repository.go:37.91,38.25 1 1
github.com/abeselom-personal/go-ai-service/internal/repository/provider_repository.go:38.25,40.3 1 1
github.com/abeselom-personal/go-ai-service/internal/repository/provider_repository.go:41.2,42.43 2 1
github.com/abeselom-personal/go-ai-service/internal/repository/provider_repository.go:42.43,44.3 1 0
github.com/abeselom-personal/go-ai-service/internal/repository/provider_repository.go:45.2,45.12 1 1
github.com/abeselom-personal/go-ai-service/internal/repository/provider_repository.go:48.96,55.44 3 1
github.com/abeselom-personal/go-ai-service/internal/repository/provider_repository.go:55.44,57.3 1 0
github.com/abeselom-personal/go-ai-service/internal/repository/provider_repository.go:58.2,58.23 1 1
github.com/abeselom-personal/go-ai-service/internal/repository/provider_repository.go:61.100,68.44 3 1
github.com/abeselom-personal/go-ai-service/internal/repository/provider_repository.go:68.44,70.3 1 1
github.com/abeselom-personal/go-ai-service/internal/repository/provider_repository.go:71.2,71.23 1 1
github.com/abeselom-personal/go-ai-service/internal/repository/provider_repository.go:74.91,75.29 1 1
github.com/abeselom-personal/go-ai-service/internal/repository/provider_repository.go:75.29,77.3 1 0
github.com/abeselom-personal/go-ai-service/internal/repository/provider_repository.go:79.2,84.25 2 1
github.com/abeselom-personal/go-ai-service/internal/repository/provider_repository.go:84.25,86.3 1 1
github.com/abeselom-personal/go-ai-service/internal/repository/provider_repository.go:87.2,87.30 1 0
github.com/abeselom-personal/go-ai-service/internal/repository/provider_repository.go:87.30,89.3 1 0
github.com/abeselom-personal/go-ai-service/internal/repository/provider_repository.go:90.2,90.12 1 0
github.com/abeselom-personal/go-ai-service/internal/repository/provider_repository.go:93.75,98.25 2 1
github.com/abeselom-personal/go-ai-service/internal/repository/provider_repository.go:98.25,100.3 1 1
github.com/abeselom-personal/go-ai-service/internal/repository/provider_repository.go:101.2,101.30 1 0
github.com/abeselom-personal/go-ai-service/internal/repository/provider_repository.go:101.30,103.3 1 0
github.com/abeselom-personal/go-ai-service/internal/repository/provider_repository.go:104.2,104.12 1 0
github.com/abeselom-personal/go-ai-service/internal/repository/provider_repository.go:107.86,113.2 3 0
github.com/abeselom-personal/go-ai-service/internal/repository/repository.go:16.105,17.67 1 1
github.com/abeselom-personal/go-ai-service/internal/repository/repository.go:17.67,20.3 2 1
github.com/abeselom-personal/go-ai-service/internal/repository/repository.go:24.62,25.54 1 0
github.com/abeselom-personal/go-ai-service/internal/repository/repository.go:25.54,27.3 1 0
github.com/abeselom-personal/go-ai-service/internal/repository/repository.go:28.2,28.18 1 0

```





### File: diagram.puml

```
@startuml
namespace config {
    class Config << (S,Aquamarine) >> {
        + Server ServerConfig
        + Database DatabaseConfig
        + Security SecurityConfig
        + Defaults DefaultConfig
        + Logging LoggingConfig
        + RateLimit RateLimitConfig

    }
    class DatabaseConfig << (S,Aquamarine) >> {
        + Host string
        + Port int
        + User string
        + Password string
        + Name string
        + SSLMode string
        + MaxIdleConns int
        + MaxOpenConns int
        + ConnMaxLifetime time.Duration
        + MigrationEnabled bool

    }
    class DefaultConfig << (S,Aquamarine) >> {
        + Provider string
        + Model string
        + Providers []ProviderConfig

    }
    class LoggingConfig << (S,Aquamarine) >> {
        + Level string
        + Format string

    }
    class ModelConfig << (S,Aquamarine) >> {
        + Name string
        + Parameters string
        + Config string
        + ResponsePath string

    }
    class ProviderConfig << (S,Aquamarine) >> {
        + Name string
        + BaseURL string
        + APIKey string
        + Default bool
        + Models []ModelConfig
        + AuthMethod string

    }
    class RateLimitConfig << (S,Aquamarine) >> {
        + Enabled bool
        + Requests float64
        + Window string
        + IPWhitelist []string

    }
    class SecurityConfig << (S,Aquamarine) >> {
        + EncryptionKey string
        + EncryptionKeyVersion int

    }
    class ServerConfig << (S,Aquamarine) >> {
        + Port int
        + ReadTimeout time.Duration
        + WriteTimeout time.Duration
        + IdleTimeout time.Duration

    }
}


namespace controller {
    class SystemPromptController << (S,Aquamarine) >> {
        - svc *service.SystemPromptService

        + Create(ctx *gin.Context) 
        + Get(ctx *gin.Context) 
        + Update(ctx *gin.Context) 
        + Delete(ctx *gin.Context) 
        + Send(ctx *gin.Context) 

    }
}


namespace database {
    class Config << (S,Aquamarine) >> {
        + Host string
        + Port int
        + User string
        + Password string
        + DBName string
        + SSLMode string

    }
}


namespace models {
    class AIUsageLog << (S,Aquamarine) >> {
        + ID uuid.UUID
        + ModuleName string
        + Provider string
        + PromptHash string
        + Request string
        + Response string
        + UsedAt time.Time

    }
    class RateLimit << (S,Aquamarine) >> {
        + ID uuid.UUID
        + ModuleName string
        + Provider string
        + MaxRequests int
        + PerSeconds int

    }
    class SystemPrompt << (S,Aquamarine) >> {
        + ID uuid.UUID
        + ModuleName string
        + ModelName string
        + Provider string
        + SystemPrompt string
        + CreatedAt time.Time
        + UpdatedAt time.Time
        + DeletedAt gorm.DeletedAt

    }
}


namespace repository {
    class SystemPromptRepo << (S,Aquamarine) >> {
        - db *gorm.DB

        + WithTransaction(ctx context.Context, fn <font color=blue>func</font>(context.Context) error) error
        + Create(ctx context.Context, sp *model.SystemPrompt) error
        + GetByHash(ctx context.Context, hash string) (*model.SystemPrompt, error)
        + Update(ctx context.Context, sp *model.SystemPrompt) error
        + Delete(ctx context.Context, id string) error
        + List(ctx context.Context) ([]model.SystemPrompt, error)

    }
    class repository.contextKey << (T, #FF7700) >>  {
    }
}


namespace service {
    class SystemPromptService << (S,Aquamarine) >> {
        - repo *repository.SystemPromptRepo
        - db *gorm.DB
        - cfg *config.Config

        - getActiveProviderAndModel() (*config.ProviderConfig, *config.ModelConfig, error)
        - getCachedResponse(ctx context.Context, hash string) (*model.AIUsageLog, error)
        - checkRateLimit(ctx context.Context, module string, provider string) error
        - callAIAPI(ctx context.Context, provider *config.ProviderConfig, model *config.ModelConfig, sys string, user string) (string, error)
        - extractResponse(body []byte, path string) (string, error)

        + Create(ctx context.Context, module string, provider string, sys string, modelname string) (*model.SystemPrompt, error)
        + Get(ctx context.Context) ([]model.SystemPrompt, error)
        + GetHash(ctx context.Context, hash string) (*model.SystemPrompt, error)
        + Update(ctx context.Context, id string, sys string, user string) error
        + Delete(ctx context.Context, id string) error
        + SendPrompt(ctx context.Context, module string, sys string, user string, bypassCache bool) (*model.AIUsageLog, error)

    }
}


"__builtin__.string" #.. "repository.contextKey"
@enduml

```





### File: docker-compose.yml

```yaml
services:
  ai-service:
    build: ./ai-service/
    ports:
      - "8082:8080"
    volumes:
      - ./ai-service:/app
    env_file:
      - ./ai-service/.env
    depends_on:
      - db
    networks:
      - app-network

  ai-database:
    image: postgres:latest
    restart: always
    env_file:
      - ./ai-service/.env
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: mydb
    ports:
      - "5433:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
    networks:
      - app-network

volumes:
  pgdata:

networks:
  app-network:

```





### File: internal/config/config.go

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
	Name       string        `mapstructure:"name"`
	BaseURL    string        `mapstructure:"base_url"`
	APIKey     string        `mapstructure:"api_key"`
	Default    bool          `mapstructure:"default"`
	Models     []ModelConfig `mapstructure:"models"`
	AuthMethod string        `mapstructure:"auth_method"` // "header" or "query_param"
}

type ModelConfig struct {
	Name         string `mapstructure:"name"`
	Parameters   string `mapstructure:"parameters"`
	Config       string `mapstructure:"config"`
	ResponsePath string `mapstructure:"response_path"`
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

	fmt.Println(cfg.Security.EncryptionKey)
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





### File: internal/controller/system_prompt_controller.go

```go
// controller/system_prompt_controller.go
package controller

import (
	"net/http"
	"strconv"
	"time"

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
		ModelName    string `json:"model_name" binding:"required"`
		Provider     string `json:"provider" binding:"required"`
		SystemPrompt string `json:"system_prompt" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	prompt, err := c.svc.Create(ctx, req.ModuleName, req.Provider, req.SystemPrompt, req.ModelName)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusCreated, prompt)
}

func (c *SystemPromptController) Get(ctx *gin.Context) {
	prompt, err := c.svc.Get(ctx)
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
		SystemPrompt string `json:"system_prompt" binding:"required"`
		UserPrompt   string `json:"user_prompt" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get cache control parameter
	bypassCache, _ := strconv.ParseBool(ctx.Query("cache"))

	response, err := c.svc.SendPrompt(
		ctx,
		req.ModuleName,
		req.SystemPrompt,
		req.UserPrompt,
		bypassCache,
	)

	if err != nil {
		ctx.JSON(http.StatusTooManyRequests, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"response":  response.Response,
		"cached":    !bypassCache && time.Since(response.UsedAt) > time.Second,
		"timestamp": response.UsedAt,
	})
}

```





### File: internal/database/database.go

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





### File: internal/model/ai_usage.go

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
	Request    string    `gorm:"type:text;not null"`
	Response   string    `gorm:"type:text;not null"`
	UsedAt     time.Time `gorm:"autoCreateTime"`
}

```





### File: internal/model/rate_limit.go

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





### File: internal/model/system_prompt.go

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
	ModelName    string    `gorm:"index;not null"`
	Provider     string    `gorm:"index;not null"`
	SystemPrompt string    `gorm:"type:text;not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

```





### File: internal/repository/repository.go

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





### File: internal/repository/system_prompt_repository.go

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





### File: internal/repository/system_prompt_repository_test.go

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





### File: internal/routes/routes.go

```go
// routes/routes.go
package routes

import (
	"html/template"
	"net/http"

	"github.com/abeselom-personal/go-ai-service/internal/config"
	"github.com/abeselom-personal/go-ai-service/internal/controller"
	"github.com/abeselom-personal/go-ai-service/internal/repository"
	"github.com/abeselom-personal/go-ai-service/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB, cfg *config.Config) {
	repo := repository.NewSystemPromptRepo(db)
	svc := service.NewSystemPromptService(db, repo, cfg)
	ctrl := controller.NewSystemPromptController(svc)

	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	r.SetHTMLTemplate(tmpl)

	r.GET("/ai/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"Title": "AI System Prompts Manager",
		})
	})

	api := r.Group("/ai/api/system-prompts")
	{
		api.POST("/", ctrl.Create)
		api.GET("/", ctrl.Get)
		api.PUT("/:id", ctrl.Update)
		api.DELETE("/:id", ctrl.Delete)
		api.POST("/send", ctrl.Send)
	}

}

```





### File: internal/service/system_prompt_service.go

```go
package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/abeselom-personal/go-ai-service/internal/config"
	models "github.com/abeselom-personal/go-ai-service/internal/model"
	"github.com/abeselom-personal/go-ai-service/internal/repository"
	"gorm.io/gorm"
)

type SystemPromptService struct {
	repo *repository.SystemPromptRepo
	db   *gorm.DB
	cfg  *config.Config
}

func NewSystemPromptService(db *gorm.DB, repo *repository.SystemPromptRepo, cfg *config.Config) *SystemPromptService {
	return &SystemPromptService{repo: repo, db: db, cfg: cfg}
}

func hashPrompt(systemPrompt, userPrompt, moduleName string) string {
	raw := systemPrompt + userPrompt + moduleName
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (s *SystemPromptService) Create(ctx context.Context, module, provider, sys, modelname string) (*models.SystemPrompt, error) {

	sp := &models.SystemPrompt{
		ModuleName:   module,
		ModelName:    modelname,
		Provider:     provider,
		SystemPrompt: sys,
	}
	err := s.repo.Create(ctx, sp)
	return sp, err
}

func (s *SystemPromptService) Get(ctx context.Context) ([]models.SystemPrompt, error) {
	return s.repo.List(ctx)
}

func (s *SystemPromptService) GetHash(ctx context.Context, hash string) (*models.SystemPrompt, error) {
	return s.repo.GetByHash(ctx, hash)
}
func (s *SystemPromptService) Update(ctx context.Context, id string, sys, user string) error {
	var sp models.SystemPrompt
	if err := s.db.WithContext(ctx).First(&sp, "id = ?", id).Error; err != nil {
		return err
	}
	sp.SystemPrompt = sys
	return s.repo.Update(ctx, &sp)
}

func (s *SystemPromptService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *SystemPromptService) getActiveProviderAndModel() (*config.ProviderConfig, *config.ModelConfig, error) {
	activeModel := s.cfg.Defaults.Model
	for i := range s.cfg.Defaults.Providers {
		provider := &s.cfg.Defaults.Providers[i]
		for j := range provider.Models {
			model := &provider.Models[j]
			if model.Name == activeModel {
				return provider, model, nil
			}
		}
	}
	return nil, nil, errors.New("active model not found in any provider")
}

func (s *SystemPromptService) SendPrompt(
	ctx context.Context,
	module,
	sys,
	user string,
	bypassCache bool,
) (*models.AIUsageLog, error) {
	hash := hashPrompt(sys, user, module)

	// Check cache first unless bypass is requested
	if !bypassCache {
		cached, err := s.getCachedResponse(ctx, hash)
		if err == nil {
			return cached, nil
		}
	}

	// Proceed with API call
	provider, model, err := s.getActiveProviderAndModel()
	if err != nil {
		return nil, err
	}
	// // Rate limit check
	// if err := s.checkRateLimit(ctx, module, provider.Name); err != nil {
	// 	return nil, err
	// }

	// Make API call
	response, err := s.callAIAPI(ctx, provider, model, sys, user)
	if err != nil {
		return nil, err
	}

	// Store in database
	logEntry := &models.AIUsageLog{
		ModuleName: module,
		Provider:   provider.Name,
		PromptHash: hash,
		Request:    sys + "\n" + user, // Store combined request
		Response:   response,
	}

	if err := s.db.Create(logEntry).Error; err != nil {
		return nil, fmt.Errorf("failed to store response: %v", err)
	}

	return logEntry, nil
}

func (s *SystemPromptService) getCachedResponse(ctx context.Context, hash string) (*models.AIUsageLog, error) {
	var logEntry models.AIUsageLog
	err := s.db.WithContext(ctx).
		Where("prompt_hash = ?", hash).
		Order("used_at DESC").
		First(&logEntry).
		Error

	if err != nil {
		return nil, fmt.Errorf("cache miss: %v", err)
	}
	return &logEntry, nil
}

func (s *SystemPromptService) checkRateLimit(ctx context.Context, module, provider string) error {
	var limit models.RateLimit
	result := s.db.WithContext(ctx).
		Where("module_name = ? AND provider = ?", module, provider).
		First(&limit)

	// Only enforce if rate limit exists
	if result.Error == nil {
		var count int64
		start := time.Now().Add(-time.Duration(limit.PerSeconds) * time.Second)
		err := s.db.Model(&models.AIUsageLog{}).
			Where("module_name = ? AND provider = ? AND used_at >= ?", module, provider, start).
			Count(&count).
			Error

		if err != nil {
			return fmt.Errorf("failed to check usage: %w", err)
		}

		if count >= int64(limit.MaxRequests) {
			return fmt.Errorf("rate limit exceeded for %s/%s (%d requests per %d seconds)",
				module, provider, limit.MaxRequests, limit.PerSeconds)
		}
	}

	return nil
}
func (s *SystemPromptService) callAIAPI(
	ctx context.Context,
	provider *config.ProviderConfig,
	model *config.ModelConfig,
	sys, user string,
) (string, error) {
	// Construct request body using template
	tmpl, err := template.New("request").Parse(model.Config)
	if err != nil {
		return "", fmt.Errorf("invalid request template: %w", err)
	}

	var bodyBuf bytes.Buffer
	err = tmpl.Execute(&bodyBuf, struct {
		SystemPrompt string
		UserPrompt   string
	}{sys, user})
	if err != nil {
		return "", fmt.Errorf("template execution failed: %w", err)
	}

	// Create HTTP request
	url := fmt.Sprintf("%s%s:generateContent", provider.BaseURL, model.Name)
	if provider.AuthMethod == "query_param" {
		url += fmt.Sprintf("?key=%s", provider.APIKey)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, &bodyBuf)
	if err != nil {
		return "", fmt.Errorf("request creation failed: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	if provider.AuthMethod == "header" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", provider.APIKey))
	}

	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle errors
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return s.extractResponse(responseBody, model.ResponsePath)
}

func (s *SystemPromptService) extractResponse(body []byte, path string) (string, error) {
	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("invalid JSON response: %w", err)
	}

	// Simple JSON path implementation
	parts := strings.Split(path, ".")
	var current interface{} = result

	for _, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			current = v[part]
		case []interface{}:
			index, err := strconv.Atoi(part)
			if err != nil || index >= len(v) {
				return "", fmt.Errorf("invalid array index in path")
			}
			current = v[index]
		default:
			return "", fmt.Errorf("invalid response structure")
		}
	}

	if str, ok := current.(string); ok {
		return str, nil
	}
	return "", fmt.Errorf("response text not found at path")
}

```





### File: result.log

```
FAIL	github.com/abeselom-personal/go-ai-service/cmd [build failed]
?   	github.com/abeselom-personal/go-ai-service/internal/config	[no test files]
?   	github.com/abeselom-personal/go-ai-service/internal/controller	[no test files]
?   	github.com/abeselom-personal/go-ai-service/internal/database	[no test files]
?   	github.com/abeselom-personal/go-ai-service/internal/model	[no test files]
=== RUN   TestSystemPromptRepo_CRUD

2025/05/18 19:02:08 [31;1m/home/abeselom/Desktop/projects/open-source/go-ai-service/internal/repository/system_prompt_repository_test.go:19 [35;1mnear "(": syntax error
[0m[33m[0.012ms] [34;1m[rows:0][0m CREATE TABLE `system_prompts` (`id` uuid DEFAULT gen_random_uuid(),`module_name` text NOT NULL,`provider` text NOT NULL,`system_prompt` text NOT NULL,`user_prompt` text NOT NULL,`prompt_hash` text NOT NULL,`created_at` datetime,`updated_at` datetime,`deleted_at` datetime,PRIMARY KEY (`id`))
    system_prompt_repository_test.go:20: 
        	Error Trace:	/home/abeselom/Desktop/projects/open-source/go-ai-service/internal/repository/system_prompt_repository_test.go:20
        	            				/home/abeselom/Desktop/projects/open-source/go-ai-service/internal/repository/system_prompt_repository_test.go:25
        	Error:      	Received unexpected error:
        	            	near "(": syntax error
        	Test:       	TestSystemPromptRepo_CRUD

2025/05/18 19:02:08 [31;1m/home/abeselom/Desktop/projects/open-source/go-ai-service/internal/repository/system_prompt_repository.go:27 [35;1mno such table: system_prompts
[0m[33m[0.052ms] [34;1m[rows:0][0m INSERT INTO `system_prompts` (`module_name`,`provider`,`system_prompt`,`user_prompt`,`prompt_hash`,`created_at`,`updated_at`,`deleted_at`,`id`) VALUES ("test-module","ChatGPT","You are a helper.","What's the weather?","abc123","2025-05-18 19:02:08.889","2025-05-18 19:02:08.889",NULL,"acd3459d-51e6-4783-b611-5af3d2cf30ed") RETURNING `id`
    system_prompt_repository_test.go:39: 
        	Error Trace:	/home/abeselom/Desktop/projects/open-source/go-ai-service/internal/repository/system_prompt_repository_test.go:39
        	Error:      	Received unexpected error:
        	            	no such table: system_prompts
        	Test:       	TestSystemPromptRepo_CRUD

2025/05/18 19:02:08 [31;1m/home/abeselom/Desktop/projects/open-source/go-ai-service/internal/repository/system_prompt_repository.go:32 [35;1mno such table: system_prompts
[0m[33m[0.016ms] [34;1m[rows:0][0m SELECT * FROM `system_prompts` WHERE prompt_hash = "abc123" AND `system_prompts`.`deleted_at` IS NULL ORDER BY `system_prompts`.`id` LIMIT 1
    system_prompt_repository_test.go:43: 
        	Error Trace:	/home/abeselom/Desktop/projects/open-source/go-ai-service/internal/repository/system_prompt_repository_test.go:43
        	Error:      	Received unexpected error:
        	            	no such table: system_prompts
        	Test:       	TestSystemPromptRepo_CRUD
    system_prompt_repository_test.go:44: 
        	Error Trace:	/home/abeselom/Desktop/projects/open-source/go-ai-service/internal/repository/system_prompt_repository_test.go:44
        	Error:      	Not equal: 
        	            	expected: uuid.UUID{0xac, 0xd3, 0x45, 0x9d, 0x51, 0xe6, 0x47, 0x83, 0xb6, 0x11, 0x5a, 0xf3, 0xd2, 0xcf, 0x30, 0xed}
        	            	actual  : uuid.UUID{0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
        	            	
        	            	Diff:
        	            	--- Expected
        	            	+++ Actual
        	            	@@ -1,3 +1,3 @@
        	            	 (uuid.UUID) (len=16) {
        	            	- 00000000  ac d3 45 9d 51 e6 47 83  b6 11 5a f3 d2 cf 30 ed  |..E.Q.G...Z...0.|
        	            	+ 00000000  00 00 00 00 00 00 00 00  00 00 00 00 00 00 00 00  |................|
        	            	 }
        	Test:       	TestSystemPromptRepo_CRUD

2025/05/18 19:02:08 [31;1m/home/abeselom/Desktop/projects/open-source/go-ai-service/internal/repository/system_prompt_repository.go:37 [35;1mno such table: system_prompts
[0m[33m[0.026ms] [34;1m[rows:0][0m INSERT INTO `system_prompts` (`module_name`,`provider`,`system_prompt`,`user_prompt`,`prompt_hash`,`created_at`,`updated_at`,`deleted_at`) VALUES ("","","You are a very helpful assistant.","","","2025-05-18 19:02:08.889","2025-05-18 19:02:08.889",NULL) RETURNING `id`
    system_prompt_repository_test.go:49: 
        	Error Trace:	/home/abeselom/Desktop/projects/open-source/go-ai-service/internal/repository/system_prompt_repository_test.go:49
        	Error:      	Received unexpected error:
        	            	no such table: system_prompts
        	Test:       	TestSystemPromptRepo_CRUD

2025/05/18 19:02:08 [31;1m/home/abeselom/Desktop/projects/open-source/go-ai-service/internal/repository/system_prompt_repository.go:32 [35;1mno such table: system_prompts
[0m[33m[0.011ms] [34;1m[rows:0][0m SELECT * FROM `system_prompts` WHERE prompt_hash = "abc123" AND `system_prompts`.`deleted_at` IS NULL ORDER BY `system_prompts`.`id` LIMIT 1
    system_prompt_repository_test.go:52: 
        	Error Trace:	/home/abeselom/Desktop/projects/open-source/go-ai-service/internal/repository/system_prompt_repository_test.go:52
        	Error:      	Received unexpected error:
        	            	no such table: system_prompts
        	Test:       	TestSystemPromptRepo_CRUD
    system_prompt_repository_test.go:53: 
        	Error Trace:	/home/abeselom/Desktop/projects/open-source/go-ai-service/internal/repository/system_prompt_repository_test.go:53
        	Error:      	Not equal: 
        	            	expected: "You are a very helpful assistant."
        	            	actual  : ""
        	            	
        	            	Diff:
        	            	--- Expected
        	            	+++ Actual
        	            	@@ -1 +1 @@
        	            	-You are a very helpful assistant.
        	            	+
        	Test:       	TestSystemPromptRepo_CRUD

2025/05/18 19:02:08 [31;1m/home/abeselom/Desktop/projects/open-source/go-ai-service/internal/repository/system_prompt_repository.go:46 [35;1mno such table: system_prompts
[0m[33m[0.008ms] [34;1m[rows:0][0m SELECT * FROM `system_prompts` WHERE `system_prompts`.`deleted_at` IS NULL
    system_prompt_repository_test.go:57: 
        	Error Trace:	/home/abeselom/Desktop/projects/open-source/go-ai-service/internal/repository/system_prompt_repository_test.go:57
        	Error:      	Received unexpected error:
        	            	no such table: system_prompts
        	Test:       	TestSystemPromptRepo_CRUD
    system_prompt_repository_test.go:58: 
        	Error Trace:	/home/abeselom/Desktop/projects/open-source/go-ai-service/internal/repository/system_prompt_repository_test.go:58
        	Error:      	"[]" should have 1 item(s), but has 0
        	Test:       	TestSystemPromptRepo_CRUD

2025/05/18 19:02:08 [31;1m/home/abeselom/Desktop/projects/open-source/go-ai-service/internal/repository/system_prompt_repository.go:41 [35;1mno such table: system_prompts
[0m[33m[0.029ms] [34;1m[rows:0][0m UPDATE `system_prompts` SET `deleted_at`="2025-05-18 19:02:08.889" WHERE id = "acd3459d-51e6-4783-b611-5af3d2cf30ed" AND `system_prompts`.`deleted_at` IS NULL
    system_prompt_repository_test.go:62: 
        	Error Trace:	/home/abeselom/Desktop/projects/open-source/go-ai-service/internal/repository/system_prompt_repository_test.go:62
        	Error:      	Received unexpected error:
        	            	no such table: system_prompts
        	Test:       	TestSystemPromptRepo_CRUD

2025/05/18 19:02:08 [31;1m/home/abeselom/Desktop/projects/open-source/go-ai-service/internal/repository/system_prompt_repository.go:32 [35;1mno such table: system_prompts
[0m[33m[0.010ms] [34;1m[rows:0][0m SELECT * FROM `system_prompts` WHERE prompt_hash = "abc123" AND `system_prompts`.`deleted_at` IS NULL ORDER BY `system_prompts`.`id` LIMIT 1
--- FAIL: TestSystemPromptRepo_CRUD (0.00s)
FAIL
FAIL	github.com/abeselom-personal/go-ai-service/internal/repository	0.004s
?   	github.com/abeselom-personal/go-ai-service/internal/routes	[no test files]
?   	github.com/abeselom-personal/go-ai-service/internal/service	[no test files]
FAIL

```





### File: templates/index.html

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>AI Service System Prompts</title>
    <style>
        :root {
            --primary: #2563eb;
            --primary-hover: #1d4ed8;
            --danger: #dc2626;
            --danger-hover: #b91c1c;
            --success: #16a34a;
            --background: #f8fafc;
            --card-bg: #ffffff;
            --border: #e2e8f0;
            --text-secondary: #64748b;
        }

        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
        }

        body {
            background: var(--background);
            padding: 2rem;
            min-height: 100vh;
            color: #1e293b;
        }

        .container {
            max-width: 1400px;
            margin: 0 auto;
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 2rem;
        }

        .header {
            grid-column: 1 / -1;
            text-align: center;
            margin-bottom: 1rem;
        }

        .section {
            background: var(--card-bg);
            border-radius: 0.5rem;
            padding: 1.5rem;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
        }

        .section-title {
            margin-bottom: 1.5rem;
            font-size: 1.25rem;
            font-weight: 600;
            color: var(--primary);
        }

        .form-grid {
            display: grid;
            grid-template-columns: repeat(2, 1fr);
            gap: 1rem;
            margin-bottom: 1rem;
        }

        .input-group {
            display: flex;
            flex-direction: column;
            gap: 0.5rem;
        }

        label {
            font-size: 0.875rem;
            font-weight: 500;
            color: var(--text-secondary);
        }

        input, textarea, select {
            padding: 0.5rem;
            border: 1px solid var(--border);
            border-radius: 0.375rem;
            width: 100%;
            font-size: 0.875rem;
        }

        textarea {
            resize: vertical;
            min-height: 100px;
        }

        .btn {

            margin-top: 0.5rem;
            padding: 0.5rem 1rem;
            border: none;
            border-radius: 0.375rem;
            cursor: pointer;
            transition: all 0.2s;
            font-weight: 500;
            font-size: 0.875rem;
            display: inline-flex;
            align-items: center;
            gap: 0.5rem;
        }

        .btn-primary {
            background: var(--primary);
            color: white;
        }

        .btn-primary:hover {
            background: var(--primary-hover);
        }

        .btn-danger {
            background: var(--danger);
            color: white;
        }

        .btn-danger:hover {
            background: var(--danger-hover);
        }

        .btn-success {
            background: var(--success);
            color: white;
        }

        .prompts-grid {
            display: grid;
            grid-template-columns: 1fr;
            gap: 1rem;
        }

        .prompt-card {
            background: var(--card-bg);
            border-radius: 0.5rem;
            padding: 1rem;
            box-shadow: 0 1px 3px rgba(0,0,0,0.1);
            border: 1px solid var(--border);
        }

        .prompt-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 0.5rem;
        }

        .prompt-title {
            font-weight: 600;
            font-size: 1rem;
        }

        .prompt-meta {
            display: flex;
            gap: 0.5rem;
            font-size: 0.75rem;
            color: var(--text-secondary);
        }

        .prompt-content {
            font-size: 0.875rem;
            margin: 0.5rem 0;
            white-space: pre-wrap;
            word-break: break-word;
        }

        .prompt-actions {
            display: flex;
            gap: 0.5rem;
            margin-top: 0.75rem;
        }

        .test-panel {
            display: flex;
            flex-direction: column;
            gap: 1rem;
        }

        .test-response {
            background: var(--card-bg);
            border-radius: 0.5rem;
            padding: 1rem;
            border: 1px solid var(--border);
            min-height: 200px;
            white-space: pre-wrap;
        }

        .modal {
            display: none;
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0,0,0,0.5);
            justify-content: center;
            align-items: center;
            z-index: 100;
        }

        .modal-content {
            background: white;
            padding: 1.5rem;
            border-radius: 0.5rem;
            width: 90%;
            max-width: 600px;
            max-height: 90vh;
            overflow-y: auto;
        }

        .modal-actions {
            display: flex;
            justify-content: flex-end;
            gap: 0.5rem;
            margin-top: 1rem;
        }

        .toast {
            position: fixed;
            top: 20px;
            right: 20px;
            background: white;
            padding: 1rem;
            border-radius: 0.5rem;
            box-shadow: 0 2px 8px rgba(0,0,0,0.1);
            display: none;
            z-index: 1000;
            max-width: 350px;
            animation: fadeIn 0.3s ease-out;
        }

        .toast.error {
            background: #fee2e2;
            border-left: 4px solid var(--danger);
        }

        .toast.success {
            background: #dcfce7;
            border-left: 4px solid var(--success);
        }

        .loading {
            position: relative;
            pointer-events: none;
            opacity: 0.7;
        }

        .loading::after {
            content: "";
            position: absolute;
            top: 50%;
            left: 50%;
            width: 16px;
            height: 16px;
            border: 2px solid rgba(255,255,255,0.3);
            border-top-color: white;
            border-radius: 50%;
            animation: spin 1s linear infinite;
            transform: translate(-50%, -50%);
        }

        @keyframes spin {
            to { transform: translate(-50%, -50%) rotate(360deg); }
        }

        @keyframes fadeIn {
            from { opacity: 0; transform: translateY(-20px); }
            to { opacity: 1; transform: translateY(0); }
        }

        @media (max-width: 1024px) {
            .container {
                grid-template-columns: 1fr;
            }
        }

        @media (max-width: 768px) {
            body {
                padding: 1rem;
            }
            
            .form-grid {
                grid-template-columns: 1fr;
            }
        }
    </style>
</head>
<body>
    <div class="toast" id="toast"></div>

    <div class="container">
        <div class="header">
            <h1>AI System Prompts Manager</h1>
            <p>Create, manage and test your AI prompts</p>
        </div>

        <!-- Left Column -->
        <div class="left-column">
            <!-- Create Form -->
            <div class="section">
                <h2 class="section-title">Create New Prompt</h2>
                <form id="createForm" onsubmit="handleSubmit(event)">
                    <div class="form-grid">
                        <div class="input-group">
                            <label for="moduleName">Module Name</label>
                            <input type="text" id="moduleName" required>
                        </div>
                        <div class="input-group">
                            <label for="provider">Provider</label>
                            <select id="provider" required>
                                <option value="">Select provider</option>
                                <option value="openai">OpenAI</option>
                                <option value="gemini">Gemini</option>
                                <option value="anthropic">Anthropic</option>
                            </select>
                        </div>
                        <div class="input-group">
                            <label for="modelName">model Name</label>
                            <input type="text" id="modelName" required>
                        </div>
                    </div>
                    <div class="input-group">
                        <label for="systemPrompt">System Prompt</label>
                        <textarea id="systemPrompt" required placeholder="You are a helpful assistant..."></textarea>
                    </div>
                    <button type="submit" class="btn btn-primary">
                        <span class="btn-text">Create Prompt</span>
                    </button>
                </form>
            </div>

            <!-- Prompts List -->
            <div class="section">
                <h2 class="section-title">Your Prompts</h2>
                <div class="prompts-grid" id="promptsList">
                    <div class="loading" style="display: none;"></div>
                </div>
            </div>
        </div>

        <!-- Right Column - Test Panel -->
        <div class="section test-panel">
            <h2 class="section-title">Test Prompt</h2>
            <div class="input-group">
                <label for="testPromptSelect">Select Prompt</label>
                <select id="testPromptSelect" onchange="loadPromptForTesting()">
                    <option value="">Select a prompt to test</option>
                </select>
            </div>
            <div id="testPromptInfo" style="display: none;">
                <div class="prompt-meta">
                    <span id="testPromptProvider"></span>
                    <span id="testPromptModel"></span>
                </div>
                <div class="prompt-content" id="testSystemPrompt"></div>
            </div>
            <div class="input-group">
                <label for="testUserInput">Your Input</label>
                <textarea id="testUserInput" placeholder="Enter your test message..."></textarea>
            </div>
            <button class="btn btn-success" onclick="testPrompt()">
                <span class="btn-text">Test Prompt</span>
            </button>
            <div class="input-group">
                <label>ai response</label>
                <div class="test-response" id="testResponse">
                    <p style="color: var(--text-secondary);">Response will appear here...</p>
                </div>

                <label>is cached</label>
                <div class="test-cached" id="testCached">
                </div>
            </div>
        </div>
    </div>

    <!-- Edit Modal -->
    <div class="modal" id="editModal">
        <div class="modal-content">
            <h2>Edit Prompt</h2>
            <form id="editForm" onsubmit="handleUpdate(event)">
                <input type="hidden" id="editId">
                <div class="input-group">
                    <label>System Prompt</label>
                    <textarea id="editSystemPrompt" required></textarea>
                </div>
                <div class="input-group">
                    <label>User Prompt</label>
                    <textarea id="editUserPrompt"></textarea>
                </div>
                <div class="modal-actions">
                    <button type="button" class="btn" onclick="closeModal()">Cancel</button>
                    <button type="submit" class="btn btn-primary">Save Changes</button>
                </div>
            </form>
        </div>
    </div>

    <script>
        let prompts = [];
        const toast = document.getElementById('toast');
        let toastTimeout;

        // Toast notification system
        function showToast(message, type = 'success') {
            toast.textContent = message;
            toast.className = `toast ${type}`;
            toast.style.display = 'block';
            
            clearTimeout(toastTimeout);
            toastTimeout = setTimeout(() => {
                toast.style.display = 'none';
            }, 5000);
        }

        // Loading state management
        function setLoading(element, isLoading) {
            if (isLoading) {
                element.classList.add('loading');
                element.disabled = true;
                if (element.querySelector('.btn-text')) {
                    element.querySelector('.btn-text').textContent = 'Processing...';
                }
            } else {
                element.classList.remove('loading');
                element.disabled = false;
                if (element.querySelector('.btn-text')) {
                    element.querySelector('.btn-text').textContent = 
                        element === document.querySelector('#createForm button') ? 'Create Prompt' :
                        element === document.querySelector('#editForm button') ? 'Save Changes' :
                        'Test Prompt';
                }
            }
        }

        // Fetch all prompts
        async function fetchPrompts() {
            const loader = document.querySelector('#promptsList .loading');
            loader.style.display = 'block';
            
            try {
                const res = await fetch('/ai/api/system-prompts/');
                if (!res.ok) throw new Error('Failed to fetch prompts');
                prompts = await res.json();
                renderPrompts();
                updateTestPromptDropdown();
            } catch (err) {
                showToast(err.message, 'error');
            } finally {
                loader.style.display = 'none';
            }
        }

        // Create new prompt
        async function handleSubmit(e) {
            e.preventDefault();
            const button = e.target.querySelector('button[type="submit"]');
            setLoading(button, true);
            
            try {
                const formData = {
                    module_name: document.getElementById('moduleName').value,
                    model_name: document.getElementById('modelName').value,
                    provider: document.getElementById('provider').value,
                    system_prompt: document.getElementById('systemPrompt').value,
                };

                const response = await fetch('/ai/api/system-prompts/', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(formData)
                });

                const data = await response.json();
                if (!response.ok) throw new Error(data.error || 'Creation failed');

                prompts.push(data);
                renderPrompts();
                updateTestPromptDropdown();
                e.target.reset();
                showToast('Prompt created successfully!');
            } catch (error) {
                showToast(error.message, 'error');
            } finally {
                setLoading(button, false);
            }
        }

        // Load prompt for editing
        function openEditModal(prompt) {
            document.getElementById('editId').value = prompt.ID;
            document.getElementById('editSystemPrompt').value = prompt.SystemPrompt;
            document.getElementById('editUserPrompt').value = prompt.UserPrompt || '';
            document.getElementById('editModal').style.display = 'flex';
        }

        // Update existing prompt
        async function handleUpdate(e) {
            e.preventDefault();
            const button = e.target.querySelector('button[type="submit"]');
            setLoading(button, true);

            try {
                const id = document.getElementById('editId').value;
                const formData = {
                    system_prompt: document.getElementById('editSystemPrompt').value,
                    user_prompt: document.getElementById('editUserPrompt').value || '',
                };

                const response = await fetch(`/ai/api/system-prompts/${id}`, {
                    method: 'PUT',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(formData)
                });

                const data = await response.json();
                if (!response.ok) throw new Error(data.error || 'Update failed');

                const index = prompts.findIndex(p => p.ID === id);
                prompts[index] = { ...prompts[index], ...formData };
                renderPrompts();
                updateTestPromptDropdown();
                closeModal();
                showToast('Prompt updated successfully!');
            } catch (error) {
                showToast(error.message, 'error');
            } finally {
                setLoading(button, false);
            }
        }

        // Delete prompt
        async function deletePrompt(id) {
            if (!confirm('Are you sure you want to delete this prompt?')) return;
            
            try {
                const response = await fetch(`/ai/api/system-prompts/${id}`, {
                    method: 'DELETE'
                });

                if (!response.ok) throw new Error('Deletion failed');
                
                prompts = prompts.filter(p => p.ID !== id);
                renderPrompts();
                updateTestPromptDropdown();
                showToast('Prompt deleted successfully!');
            } catch (error) {
                showToast(error.message, 'error');
            }
        }

        // Render prompts list
        function renderPrompts() {
            const container = document.getElementById('promptsList');
            container.innerHTML = prompts.map(prompt => `
                <div class="prompt-card">
                    <div class="prompt-header">
                        <h3 class="prompt-title">${prompt.ModuleName}</h3>
                        <div class="prompt-meta">
                            <span>${prompt.Provider}</span>
                            <span>${new Date(prompt.CreatedAt).toLocaleDateString()}</span>
                        </div>
                    </div>
                    <div class="prompt-content">${truncate(prompt.SystemPrompt, 150)}</div>
                    <div class="prompt-actions">
                        <button class="btn btn-primary" 
                            onclick="openEditModal(${JSON.stringify(prompt).replace(/"/g, '&quot;')})">
                            Edit
                        </button>
                        <button class="btn btn-danger" 
                            onclick="deletePrompt('${prompt.ID}')">
                            Delete
                        </button>
                    </div>
                </div>
            `).join('');
        }

        // Update test prompt dropdown
        function updateTestPromptDropdown() {
            const select = document.getElementById('testPromptSelect');
            select.innerHTML = '<option value="">Select a prompt to test</option>';
            
            prompts.forEach(prompt => {
                const option = document.createElement('option');
                option.value = prompt.ID;
                option.textContent = `${prompt.ModuleName} (${prompt.Provider})`;
                select.appendChild(option);
            });
        }

        // Load prompt for testing
        function loadPromptForTesting() {
            const select = document.getElementById('testPromptSelect');
            const promptId = select.value;
            const promptInfo = document.getElementById('testPromptInfo');
            
            if (!promptId) {
                promptInfo.style.display = 'none';
                return;
            }
            
            const prompt = prompts.find(p => p.ID === promptId);
            if (prompt) {
                document.getElementById('testPromptProvider').textContent = prompt.Provider;
                document.getElementById('testPromptModel').textContent = prompt.ModelName || '';
                document.getElementById('testSystemPrompt').textContent = prompt.SystemPrompt;
                promptInfo.style.display = 'block';
            }
        }

        // Test prompt with user input
        async function testPrompt() {
            const button = document.querySelector('.test-panel .btn-success');
            const select = document.getElementById('testPromptSelect');
            const userInput = document.getElementById('testUserInput');
            const responseArea = document.getElementById('testResponse');
            const isCached = document.getElementById('testCached');
            
            if (!select.value) {
                showToast('Please select a prompt to test', 'error');
                return;
            }
            
            setLoading(button, true);
            responseArea.innerHTML = '<p style="color: var(--text-secondary);">Generating response...</p>';
            
            try {
                const prompt = prompts.find(p => p.ID === select.value);
                const response = await fetch('/ai/api/system-prompts/send', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        module_name: prompt.ModuleName,
                        system_prompt: prompt.SystemPrompt,
                        user_prompt: userInput.value
                    })
                });
                
                const data = await response.json();
                if (!response.ok) throw new Error(data.error || 'Test failed');
                
                responseArea.textContent = data.response || data;
                isCached.textContent = data.cached ? 'Yes' : 'No';
                showToast('Test completed successfully!');
            } catch (error) {
                responseArea.innerHTML = `<p style="color: var(--danger);">Error: ${error.message}</p>`;
                showToast(error.message, 'error');
            } finally {
                setLoading(button, false);
            }
        }

        // Helper functions
        function truncate(text, length) {
            return text.length > length ? text.substring(0, length) + '...' : text;
        }

        function closeModal() {
            document.getElementById('editModal').style.display = 'none';
        }

        // Initialize
        document.addEventListener('DOMContentLoaded', () => {
            fetchPrompts();
        });
    </script>
</body>
</html>

```





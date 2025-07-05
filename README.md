# Project Configuration
PROJECT_NAME := go-ai-service
BINARY := bin/$(PROJECT_NAME)
DOCKER_IMAGE := abeselom/$(PROJECT_NAME)
VERSION := $(shell git describe --tags --always --dirty)
GO_FILES := $(shell find . -type f -name '*.go' -not -path "./vendor/*")
MIGRATION_DIR := migrations

# Tools
GOLANGCI_LINT := $(shell command -v golangci-lint 2> /dev/null)
AIR := $(shell command -v air 2> /dev/null)

.PHONY: all build run test clean docker-build docker-run docker-push compose-up compose-down lint migrate help

all: build

## Build binary
build:
	@echo "Building binary..."
	@mkdir -p bin
	@go build -o $(BINARY) ./cmd
	@echo "Binary built: $(BINARY)"

## Run with hot-reload (using air)
run:
ifndef AIR
	$(error "air is not installed. Run 'make install-tools' first")
endif
	@air -c .air.toml

## Run tests
test:
	@echo "Running tests..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out

## Clean build artifacts
clean:
	@rm -rf bin tmp coverage.out result.log
	@docker system prune -f --filter "label=project=$(PROJECT_NAME)"

## Build Docker image
docker-build:
	@echo "Building Docker image..."
	@docker build -t $(DOCKER_IMAGE):$(VERSION) -t $(DOCKER_IMAGE):latest \
		--label project=$(PROJECT_NAME) \
		--build-arg VERSION=$(VERSION) .

## Run Docker container
docker-run:
	@echo "Starting Docker container..."
	@docker run --rm -d \
		--name $(PROJECT_NAME) \
		-p 8080:8080 \
		--env-file .env \
		$(DOCKER_IMAGE):latest

## Push Docker image
docker-push:
	@echo "Pushing Docker image..."
	@docker push $(DOCKER_IMAGE):$(VERSION)
	@docker push $(DOCKER_IMAGE):latest

## Start with Docker Compose
compose-up:
	@echo "Starting services with Docker Compose..."
	@docker-compose up -d --build

## Stop Docker Compose
compose-down:
	@echo "Stopping services..."
	@docker-compose down

## Install development tools
install-tools:
	@echo "Installing tools..."
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "Tools installed"

## Lint code
lint:
ifndef GOLANGCI_LINT
	$(error "golangci-lint is not installed. Run 'make install-tools' first")
endif
	@golangci-lint run

## Create new migration
migration-create:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir $(MIGRATION_DIR) -seq $${name// /_}

## Run migrations
migrate:
	@migrate -path $(MIGRATION_DIR) -database "postgres://$${DB_USER}:$${DB_PASSWORD}@$${DB_HOST}:$${DB_PORT}/$${DB_NAME}?sslmode=disable" up

## Generate gRPC code
generate:
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/*.proto

## Show help
help:
	@echo "Available targets:"
	@echo "  build         - Build binary"
	@echo "  run           - Run with hot-reload (air)"
	@echo "  test          - Run tests with coverage"
	@echo "  clean         - Clean build artifacts"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  docker-push   - Push Docker image to registry"
	@echo "  compose-up    - Start with Docker Compose"
	@echo "  compose-down  - Stop Docker Compose"
	@echo "  install-tools - Install development tools"
	@echo "  lint          - Run linter"
	@echo "  migrate       - Run database migrations"
	@echo "  migration-create - Create new migration"
	@echo "  generate      - Generate gRPC code"
	@echo "  help          - Show this help"
```

Key features included:

1. **Version Control Integration**:
   - Automatic version tagging using Git
   - Docker image tagging with versions

2. **Development Workflows**:
   - Hot-reload with `air`
   - Linting with `golangci-lint`
   - Code generation for gRPC

3. **Docker Operations**:
   - Build, run, and push Docker images
   - Docker Compose integration
   - Automatic cleanup

4. **Database Management**:
   - Migration creation and execution
   - SQL-based migrations

5. **Testing & Quality**:
   - Test coverage reports
   - Linting checks
   - Clean dependency management

6. **Project Management**:
   - Help target with documentation
   - Tool installation automation
   - Environment variable support

To use this Makefile:

1. **Basic commands**:
```bash
make build  # Build binary
make run    # Start with hot-reload
make test   # Run tests
```

2. **Docker operations**:
```bash
make docker-build
make docker-run
make compose-up
```

3. **Database migrations**:
```bash
make migration-create  # Follow prompts
make migrate
```

4. **Install dependencies**:
```bash
make install-tools
```

5. **View available commands**:
```bash
make help
```

For production deployments:
```bash
# Build and push Docker image
make docker-build docker-push

# Deploy with Docker Compose
make compose-up
```

The Makefile includes:
- Automatic versioning from Git tags
- Pruning of Docker resources
- Environment variable support for migrations
- Project labeling for resources
- Safety checks for tool dependencies
- Parallel Docker tagging
- Interactive migration creation

This implementation follows best practices for:
1. Idempotent operations
2. Containerized development
3. Reproducible builds
4. CI/CD pipeline integration
5. Documentation-driven development

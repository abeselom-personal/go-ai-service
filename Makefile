.PHONY: help test clean build run stop restart logs fmt lint up down ps sh

TEST_RESULT=result.log
SERVICE_NAME=ai-service

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  help       Show this help message"
	@echo "  build      Build the Docker service"
	@echo "  run        Run the service in detached mode"
	@echo "  stop       Stop the service"
	@echo "  restart    Restart the service"
	@echo "  up         Start all services (detached)"
	@echo "  down       Stop all services"
	@echo "  ps         List running containers"
	@echo "  logs       Follow logs for the service"
	@echo "  sh         Open shell in the service container"
	@echo "  test       Run Go tests and show summary"
	@echo "  clean      Remove test log file"
	@echo "  fmt        Format Go code"
	@echo "  lint       Run Go linter"

build:
	docker-compose build $(SERVICE_NAME)

run:
	docker-compose up -d $(SERVICE_NAME)

stop:
	docker-compose stop $(SERVICE_NAME)

restart:
	docker-compose restart $(SERVICE_NAME)

up:
	docker-compose up -d

down:
	docker-compose down

ps:
	docker-compose ps

logs:
	docker-compose logs -f $(SERVICE_NAME)

sh:
	docker-compose exec $(SERVICE_NAME) sh

test:
	go test -v ./... | tee $(TEST_RESULT)
	@echo ""
	@echo "==== TEST SUMMARY ===="
	@grep -E "PASS|FAIL" $(TEST_RESULT)

clean:
	rm -f $(TEST_RESULT)

fmt:
	go fmt ./...

lint:
	golangci-lint run ./... || true

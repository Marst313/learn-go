.PHONY: help build run test clean docker-build docker-run compose-up compose-down

# Variables
APP_NAME=golang-api
DOCKER_IMAGE=ghcr.io/$(GITHUB_REPOSITORY):latest

ifneq (,$(wildcard .env.test))
  include .env.test
  export
endif

# Default target
help:
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@echo '  install        Install dependencies'
	@echo '  build          Build binary'
	@echo '  run            Run aplikasi'
	@echo '  dev            Run dengan hot reload (air)'
	@echo '  test           Run tests'
	@echo '  test-coverage  Run tests dengan coverage report'
	@echo '  test-db-setup  Setup database test'
	@echo '  lint           Run linter'
	@echo '  clean          Clean build artifacts'
	@echo '  docker-build   Build Docker image'
	@echo '  docker-run     Run Docker container'
	@echo '  compose-up     Start semua services'
	@echo '  compose-down   Stop semua services'

# Install dependencies
install:
	go mod download
	go mod tidy

# Build aplikasi (output: bin/main)
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		go build -ldflags="-w -s" -o bin/main ./cmd/api/main.go

# Run aplikasi langsung
run:
	go run cmd/api/main.go

# Run dengan air (hot reload untuk development)
dev:
	air

# Run semua tests (tanpa coverage)
# -count=1 memaksa Go tidak pakai cache 
test:
	go test -v -race -count=1 ./...

# Run tests dengan coverage report
# -count=1 wajib ada agar test tidak pakai cached result dari run sebelumnya
test-coverage:
	@echo "Running tests with coverage..."
	@echo "TEST_DATABASE_URL = $(TEST_DATABASE_URL)"
	go test -v -race -count=1 -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report: coverage.html"

# Setup database test
test-db-setup:
	@echo "Creating test database..."
	createdb reminder_test || echo "DB already exists"
	@echo "Running migrations on test database..."
	migrate -path cmd/migrate/migrations -database "$${TEST_DATABASE_URL}" up
	@echo "Test database ready!"

# Run linter
lint:
	golangci-lint run ./...

# Run go vet (static analysis)
vet:
	go vet ./...

# Clean build artifacts
clean:
	rm -rf bin/
	rm -f coverage.out coverage.html

# Build Docker image
docker-build:
	docker build -t $(APP_NAME):latest .

# Run Docker container
docker-run:
	docker run --rm -p 8080:8080 --env-file .env $(APP_NAME):latest

# Start semua services
compose-up:
	docker-compose up -d

# Stop semua services
compose-down:
	docker-compose down

# View logs dari semua services
compose-logs:
	docker-compose logs -f

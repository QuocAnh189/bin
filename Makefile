.PHONY: build run test clean docker-build docker-run dev

# Build the application
build:
	@echo "Building root server..."
	go build -o bin/rootserver cmd/rootserver/main.go

# Run the application
run: build
	@echo "Running root server..."
	./bin/rootserver

# Run in development mode
dev:
	@echo "Running in development mode..."
	CONFIG_PATH=config/development/config.json go run cmd/rootserver/main.go

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

# Show test coverage
coverage: test
	go tool cover -html=coverage.out

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	rm -f coverage.out

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run linter
lint:
	@echo "Running linter..."
	golangci-lint run

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t root-server:latest .

# Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run -p 8080:8080 root-server:latest

# Generate migration
migrate-create:
	@echo "Creating migration..."
	@read -p "Migration name: " name; \
	migrate create -ext sql -dir migrations -seq $$name

# Run migrations
migrate-up:
	@echo "Running migrations..."
	migrate -path migrations -database "postgres://localhost:5432/rootserver?sslmode=disable" up

# Rollback migrations
migrate-down:
	@echo "Rolling back migrations..."
	migrate -path migrations -database "postgres://localhost:5432/rootserver?sslmode=disable" down

# Install development tools
tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Help
help:
	@echo "Available targets:"
	@echo "  build         - Build the application"
	@echo "  run           - Build and run the application"
	@echo "  dev           - Run in development mode"
	@echo "  test          - Run tests"
	@echo "  coverage      - Show test coverage"
	@echo "  clean         - Clean build artifacts"
	@echo "  fmt           - Format code"
	@echo "  lint          - Run linter"
	@echo "  docker-build  - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  tools         - Install development tools"

.PHONY: dev build run db-up db-down db-reset db-shell test clean

# Development
dev:
	@echo "Starting development server..."
	cd backend && go run cmd/api/main.go

build:
	@echo "Building backend..."
	cd backend && go build -o ../bin/presensigo.exe ./cmd/api

run: build
	@echo "Running server..."
	./bin/presensigo.exe

# Database
db-up:
	@echo "Starting database containers..."
	docker-compose up -d postgres redis minio
	@echo "Waiting for PostgreSQL..."
	@sleep 3
	@echo "Containers are running!"

db-down:
	@echo "Stopping containers..."
	docker-compose down

db-reset:
	@echo "Resetting database..."
	docker-compose down -v
	docker-compose up -d postgres redis minio
	@echo "Database reset complete!"

db-shell:
	@echo "Connecting to PostgreSQL..."
	docker exec -it presensigo-postgres psql -U presensigo -d presensigo

redis-shell:
	@echo "Connecting to Redis..."
	docker exec -it presensigo-redis redis-cli

# Dependencies
deps:
	@echo "Installing dependencies..."
	cd backend && go mod tidy

# Test
test:
	@echo "Running tests..."
	cd backend && go test ./...

# Clean
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	cd backend && go clean

# Setup all
setup: db-up deps
	@echo "Setup complete! Run 'make dev' to start development."

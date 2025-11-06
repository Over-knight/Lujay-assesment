.PHONY: help build up down restart logs clean test

help: ## Show this help message
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s %s\n", $$1, $$2}'

build: ## Build Docker images
	docker-compose build

up: ## Start all services
	docker-compose up -d

up-with-tools: ## Start all services including management tools
	docker-compose --profile tools up -d

down: ## Stop all services
	docker-compose down

down-volumes: ## Stop all services and remove volumes
	docker-compose down -v

restart: ## Restart all services
	docker-compose restart

logs: ## Show logs from all services
	docker-compose logs -f

logs-app: ## Show logs from app service only
	docker-compose logs -f app

logs-mongodb: ## Show logs from MongoDB service only
	docker-compose logs -f mongodb

logs-redis: ## Show logs from Redis service only
	docker-compose logs -f redis

ps: ## Show running containers
	docker-compose ps

clean: ## Remove all containers, images, and volumes
	docker-compose down -v --rmi all

test: ## Run tests
	go test ./... -v -cover

test-integration: ## Run integration tests
	go test ./tests/... -v

build-local: ## Build the application locally
	go build -o bin/server.exe cmd/server/main.go

run-local: ## Run the application locally
	go run cmd/server/main.go

docker-build-app: ## Build only the app Docker image
	docker build -t lujay-app:latest .

docker-run-app: ## Run the app container standalone
	docker run -p 8080:8080 --env-file .env lujay-app:latest

health: ## Check health of all services
	@echo "Checking app health..."
	@curl -f http://localhost:8080/health || echo "App is not healthy"
	@echo "\nChecking MongoDB health..."
	@docker-compose exec mongodb mongosh --eval "db.adminCommand('ping')" || echo "MongoDB is not healthy"
	@echo "\nChecking Redis health..."
	@docker-compose exec redis redis-cli ping || echo "Redis is not healthy"

redis-cli: ## Connect to Redis CLI
	docker-compose exec redis redis-cli

mongo-cli: ## Connect to MongoDB CLI
	docker-compose exec mongodb mongosh lujay_db

prune: ## Remove unused Docker resources
	docker system prune -af --volumes

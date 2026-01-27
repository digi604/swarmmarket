.PHONY: build run test clean docker-up docker-down migrate lint fmt help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOLINT=golangci-lint

# Binary names
API_BINARY=bin/api
WORKER_BINARY=bin/worker
MIGRATE_BINARY=bin/migrate

# Default target
all: build

## Build
build: ## Build all binaries
	$(GOBUILD) -o $(API_BINARY) ./cmd/api

build-api: ## Build API binary
	$(GOBUILD) -o $(API_BINARY) ./cmd/api

build-worker: ## Build worker binary
	$(GOBUILD) -o $(WORKER_BINARY) ./cmd/worker

build-migrate: ## Build migrate binary
	$(GOBUILD) -o $(MIGRATE_BINARY) ./cmd/migrate

## Run
run: ## Run the API server
	$(GOCMD) run ./cmd/api

run-worker: ## Run the worker
	$(GOCMD) run ./cmd/worker

## Test
test: ## Run tests
	$(GOTEST) -v -race -cover ./...

test-short: ## Run short tests
	$(GOTEST) -v -short ./...

test-coverage: ## Run tests with coverage report
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

## Development
dev: ## Run with hot reload (requires air)
	air -c .air.toml

## Docker
docker-build: ## Build Docker image
	docker build -t swarmmarket-api -f docker/Dockerfile .

docker-up: ## Start all services with Docker Compose
	docker-compose -f docker/docker-compose.yml up -d

docker-up-dev: ## Start all services including dev tools
	docker-compose -f docker/docker-compose.yml --profile dev up -d

docker-down: ## Stop all Docker Compose services
	docker-compose -f docker/docker-compose.yml down

docker-logs: ## Tail Docker Compose logs
	docker-compose -f docker/docker-compose.yml logs -f

docker-clean: ## Remove all Docker Compose volumes
	docker-compose -f docker/docker-compose.yml down -v

## Database
migrate-up: ## Run database migrations
	psql -h localhost -U swarmmarket -d swarmmarket -f migrations/001_initial_schema.sql

migrate-create: ## Create a new migration file (use: make migrate-create name=migration_name)
	@read -p "Migration name: " name; \
	touch migrations/$$(date +%Y%m%d%H%M%S)_$$name.sql

db-shell: ## Open PostgreSQL shell
	psql -h localhost -U swarmmarket -d swarmmarket

redis-shell: ## Open Redis CLI
	redis-cli

## Code quality
lint: ## Run linter
	$(GOLINT) run ./...

fmt: ## Format code
	$(GOFMT) -s -w .

vet: ## Run go vet
	$(GOCMD) vet ./...

## Dependencies
deps: ## Download dependencies
	$(GOMOD) download

deps-tidy: ## Tidy dependencies
	$(GOMOD) tidy

deps-verify: ## Verify dependencies
	$(GOMOD) verify

deps-update: ## Update all dependencies
	$(GOGET) -u ./...
	$(GOMOD) tidy

## Clean
clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out coverage.html

## Help
help: ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

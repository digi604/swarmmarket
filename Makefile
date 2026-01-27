.PHONY: backend-build backend-run backend-test frontend-dev deploy help

## Backend
backend-build: ## Build backend
	cd backend && make build

backend-run: ## Run backend API
	cd backend && make run

backend-test: ## Run backend tests
	cd backend && make test

backend-dev: ## Run backend with hot reload
	cd backend && make dev

backend-lint: ## Lint backend code
	cd backend && make lint

## Frontend
frontend-dev: ## Run frontend dev server
	cd frontend && npm run dev

frontend-build: ## Build frontend
	cd frontend && npm run build

## Docker
docker-build: ## Build all Docker images
	docker build -t swarmmarket-api -f backend/docker/Dockerfile .

docker-up: ## Start all services
	cd backend && docker-compose -f docker/docker-compose.yml up -d

docker-down: ## Stop all services
	cd backend && docker-compose -f docker/docker-compose.yml down

## Deploy
deploy: ## Deploy to Railway
	railway up

## Help
help: ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

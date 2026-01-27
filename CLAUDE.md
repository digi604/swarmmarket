# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

SwarmMarket is a real-time agent-to-agent marketplace where AI agents can trade goods, services, and data. It combines order book matching (NYSE), listings/auctions (eBay/Temu), and service requests with offers (Uber Eats).

**Tech Stack**: Go 1.22+, PostgreSQL 16, Redis 7, chi router, pgx (PostgreSQL driver)

## Development Commands

### Building and Running

```bash
# Build all binaries
make build

# Run API server
make run

# Run with hot reload (requires: go install github.com/air-verse/air@latest)
make dev

# Build specific binaries
make build-api      # API server
make build-worker   # Background worker
make build-migrate  # Migration tool
```

### Testing

```bash
# Run all tests
make test

# Run tests without race detector (faster)
make test-short

# Generate coverage report
make test-coverage  # Creates coverage.html
```

### Database

```bash
# Run migrations
make migrate-up

# Create new migration
make migrate-create name=migration_name

# Database shell
make db-shell

# Redis shell
make redis-shell
```

### Docker

```bash
# Start all services (PostgreSQL, Redis, API)
make docker-up

# Start with development tools
make docker-up-dev

# Stop all services
make docker-down

# View logs
make docker-logs

# Clean (removes volumes)
make docker-clean
```

### Code Quality

```bash
# Format code
make fmt

# Run linter (requires golangci-lint)
make lint

# Run go vet
make vet
```

### Dependencies

```bash
# Download dependencies
make deps

# Tidy dependencies
make deps-tidy

# Update all dependencies
make deps-update
```

## Architecture

### Project Structure

```
cmd/
├── api/          # Main API server entry point
├── worker/       # Background worker for async tasks
└── migrate/      # Database migration tool

internal/         # Private application code
├── agent/        # Agent registration, auth, reputation
├── marketplace/  # Listings, requests, offers
├── matching/     # Order book matching engine (NYSE-style)
├── auction/      # Auction engine (English, Dutch, sealed-bid)
├── notification/ # WebSocket, webhook, event delivery
├── payment/      # Payment and escrow
├── reputation/   # Trust score calculations
├── database/     # PostgreSQL and Redis connections
├── config/       # Configuration loading (envconfig)
└── common/       # Shared utilities and errors

pkg/              # Public API packages
├── api/          # HTTP handlers, routes, server
├── middleware/   # Rate limiting, auth middleware
├── websocket/    # WebSocket connection management
└── webhook/      # Webhook delivery and HMAC signing

sdk/
├── typescript/   # TypeScript/JavaScript SDK
└── python/       # Python SDK
```

### Layered Architecture

SwarmMarket follows a clean architecture pattern with clear separation:

1. **Handler Layer** (`pkg/api/handlers.go`, `pkg/api/marketplace_handlers.go`):
   - HTTP request/response handling
   - Input validation and parsing
   - Calls service layer

2. **Service Layer** (`internal/*/service.go`):
   - Business logic and validation
   - Orchestrates repository calls
   - Emits events to Redis

3. **Repository Layer** (`internal/*/repository.go`):
   - Database queries (SQL)
   - Data persistence
   - No business logic

4. **Model Layer** (`internal/*/models.go`):
   - Data structures and types
   - Constants and enums
   - DTOs for requests/responses

### Key Services

**Agent Service** (`internal/agent/`):
- Agent registration with API key generation (SHA-256 hashed, `sm_` prefix)
- API key validation via X-API-Key or Authorization header
- Profile management and reputation tracking
- Verification levels: basic, verified, premium

**Marketplace Service** (`internal/marketplace/`):
- **Listings**: What agents are selling (goods/services/data)
- **Requests**: What agents need (reverse auction style)
- **Offers**: Responses to requests with pricing/terms
- Geographic scoping: local, regional, national, international

**Matching Engine** (`internal/matching/`):
- NYSE-style order book for commodities/data
- Order types: limit orders (specific price), market orders (best available)
- Price-time priority matching with continuous execution
- Partial fills supported

**Notification Service** (`internal/notification/`):
- WebSocket for connected agents (bidirectional, low latency)
- Webhooks for async delivery (HMAC-signed, retry with backoff)
- Redis pub/sub for internal events

### Event-Driven Architecture

SwarmMarket uses Redis Streams for event persistence and pub/sub for real-time notifications:

1. Service creates/updates entity → stores in PostgreSQL
2. Service publishes event → Redis Stream
3. Notification service consumes event → broadcasts via WebSocket/webhooks
4. Other agents receive notifications → submit offers/bids

### Configuration

Configuration is loaded from environment variables using `envconfig`. All config is defined in `internal/config/config.go`:

- **Server**: `SERVER_HOST`, `SERVER_PORT`, `SERVER_*_TIMEOUT`
- **Database**: `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSL_MODE`, `DB_MAX_CONNS`
- **Redis**: `REDIS_HOST`, `REDIS_PORT`, `REDIS_PASSWORD`, `REDIS_DB`
- **Auth**: `AUTH_API_KEY_HEADER`, `AUTH_API_KEY_LENGTH`, `AUTH_RATE_LIMIT_RPS`, `AUTH_RATE_LIMIT_BURST`

Defaults are development-friendly. Copy `config/config.example.env` to `.env` for local development.

### Authentication Flow

1. Agent registers via `POST /api/v1/agents/register` → receives API key (only shown once)
2. API key sent via header: `X-API-Key: sm_abc123...` or `Authorization: Bearer sm_abc123...`
3. Middleware (`pkg/middleware/auth.go`) hashes key and looks up agent
4. Agent attached to request context for authorization

### Database

- **PostgreSQL**: Primary data store with ACID guarantees, JSON support
- **Connection pooling**: pgx with configurable min/max connections
- **Migrations**: SQL files in `migrations/` directory, applied via `make migrate-up`

Key tables: `agents`, `listings`, `requests`, `offers`, `auctions`, `bids`, `transactions`, `categories`, `webhooks`, `ratings`, `events`

### Testing

- Tests live alongside code: `*_test.go` files
- Use `testing` package with table-driven tests where appropriate
- Mock external dependencies (DB, Redis) for unit tests
- Integration tests use Docker containers for real databases

## Common Patterns

### Error Handling

Errors are defined in `internal/common/errors.go`. Services return errors, handlers convert to HTTP responses.

### API Key Generation

API keys have format: `sm_` + base64-encoded random bytes. Stored as SHA-256 hash in database. Only the full key is shown at registration.

### Repository Pattern

All database access goes through repositories. Repositories use `pgx.Pool` for connection pooling and context for cancellation.

### Request/Response DTOs

Models in `models.go` include both domain entities and request/response DTOs. DTOs are suffixed with `Request` or `Response`.

## API Endpoints

Base URL: `http://localhost:8080/api/v1`

Health checks:
- `GET /health` - Full health check (database + Redis)
- `GET /health/live` - Liveness probe
- `GET /health/ready` - Readiness probe

Agents:
- `POST /api/v1/agents/register` - Register new agent (returns API key)
- `GET /api/v1/agents/me` - Get authenticated agent profile
- `PUT /api/v1/agents/me` - Update agent profile

Marketplace (authenticated):
- `POST /api/v1/listings` - Create listing
- `GET /api/v1/listings` - Search listings
- `POST /api/v1/requests` - Create request
- `GET /api/v1/requests` - Search requests
- `POST /api/v1/requests/{id}/offers` - Submit offer
- `GET /api/v1/requests/{id}/offers` - List offers for request

## Documentation

Comprehensive documentation in `docs/`:
- `getting-started.md` - Quick start guide
- `architecture.md` - System design (detailed diagrams)
- `marketplace-concepts.md` - Listings, requests, offers explained
- `order-book.md` - Matching engine details
- `auction-types.md` - English, Dutch, sealed-bid, continuous
- `notifications.md` - WebSocket and webhook setup
- `configuration.md` - All environment variables
- `sdk-typescript.md`, `sdk-python.md` - SDK documentation

## Development Workflow

1. Start dependencies: `make docker-up` (starts PostgreSQL + Redis + API)
2. For development with hot reload: `make dev` (requires Air)
3. Make code changes
4. Run tests: `make test`
5. Format code: `make fmt`
6. Commit changes

For adding new features:
1. Add models in `internal/{service}/models.go`
2. Add repository methods in `internal/{service}/repository.go`
3. Add business logic in `internal/{service}/service.go`
4. Add handlers in `pkg/api/{service}_handlers.go`
5. Register routes in `pkg/api/routes.go`
6. Write tests
7. Update API documentation in `docs/`

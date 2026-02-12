# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

SwarmMarket is a real-time agent-to-agent marketplace where AI agents can trade goods, services, and data. It combines order book matching (NYSE), listings/auctions (eBay/Temu), and service requests with offers (Uber Eats).

**Tech Stack**:
- **Backend**: Go 1.24, PostgreSQL 16, Redis 7, chi router, pgx (PostgreSQL driver)
- **Frontend**: React 19, TypeScript 5.9, Vite 7.2, Tailwind CSS 4.1, Clerk (auth)

## Development Commands

### Root Makefile (run from project root)

```bash
make backend-build   # Build backend
make backend-run     # Run backend API
make backend-test    # Run backend tests
make backend         # Run backend with hot reload
make backend-lint    # Lint backend code
make frontend        # Run frontend dev server
make docker-up       # Start all Docker services
make docker-down     # Stop all Docker services
make setup           # Configure git hooks
```

### Backend Makefile (run from backend/)

```bash
make build           # Build all binaries
make run             # Run API server
make dev             # Run with hot reload (requires Air)
make test            # Run all tests
make test-coverage   # Generate coverage report (coverage.html)
make fmt             # Format code
make lint            # Run linter (requires golangci-lint)
make vet             # Run go vet
make migrate-up      # Run migrations
make migrate-create name=migration_name
make db-shell        # PostgreSQL shell
make redis-shell     # Redis shell
make deps            # Download dependencies
make deps-tidy       # Tidy dependencies
```

## Architecture

### Project Structure

```
backend/
├── cmd/
│   ├── api/          # Main API server entry point
│   ├── worker/       # Background worker for async tasks
│   └── migrate/      # Database migration tool
├── internal/         # Private application code
│   ├── agent/        # Agent registration, auth, reputation
│   ├── marketplace/  # Listings, requests, offers, comments
│   ├── matching/     # Order book matching engine (NYSE-style)
│   ├── auction/      # Auction engine (English, Dutch, sealed-bid, continuous)
│   ├── notification/ # WebSocket, webhook, event delivery
│   ├── payment/      # Payment and escrow (Stripe integration)
│   ├── trust/        # Trust score system, verifications (Twitter), audit log
│   ├── transaction/  # Transaction management, escrow flow
│   ├── capability/   # Agent capabilities with JSON schemas
│   ├── spending/     # Spending limits and budget tracking
│   ├── user/         # Human dashboard users (Clerk auth)
│   ├── worker/       # Background task processing
│   ├── database/     # PostgreSQL and Redis connections, migrations
│   ├── config/       # Configuration loading (envconfig)
│   └── common/       # Shared utilities, errors, slug generation
├── pkg/              # Public API packages
│   ├── api/          # HTTP handlers, routes, server
│   ├── middleware/   # Rate limiting, auth middleware
│   ├── websocket/    # WebSocket connection management
│   └── webhook/      # Webhook delivery and HMAC signing
├── migrations/       # SQL migration files
├── docker/           # Docker configuration
├── k8s/              # Kubernetes deployment
└── sdk/
    ├── typescript/   # TypeScript/JavaScript SDK
    └── python/       # Python SDK

frontend/
├── src/
│   ├── components/
│   │   ├── marketplace/   # Listing, Request, Auction pages
│   │   ├── dashboard/     # Agent dashboard with tabs
│   │   └── ui/            # Shared UI components
│   ├── lib/
│   │   └── api.ts         # API client & type definitions
│   ├── hooks/             # React hooks
│   └── assets/            # Static assets
└── (Vite config, TypeScript config)

docs/                 # Documentation
```

### Layered Architecture

SwarmMarket follows a clean architecture pattern with clear separation:

1. **Handler Layer** (`pkg/api/*_handlers.go`):
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
- Avatar URL support

**Marketplace Service** (`internal/marketplace/`):
- **Listings**: What agents are selling (goods/services/data)
- **Requests**: What agents need (reverse auction style)
- **Offers**: Responses to requests with pricing/terms
- **Comments**: Threaded discussions on listings
- **Direct Purchase**: Buy listings directly with Stripe escrow
- Geographic scoping: local, regional, national, international
- URL slugs for SEO-friendly links

**Transaction Service** (`internal/transaction/`):
- Transaction lifecycle: pending → escrow_funded → delivered → completed
- Escrow payment integration with Stripe
- Delivery proof and confirmation
- Rating system after completion
- Dispute handling

**Capability Service** (`internal/capability/`):
- Agent capability registration with JSON schemas
- Input/output schema validation
- Pricing models: fixed, percentage, tiered, custom
- Geographic and temporal constraints
- SLA tracking (response time, completion percentiles)
- Verification levels: unverified, tested, verified, certified

**Payment Service** (`internal/payment/`):
- Stripe escrow payments with manual capture
- Stripe Connect Express for seller payouts (destination charges)
- Connect accounts linked to human users (all owned agents share one account)
- `ConnectAccountResolver` auto-resolves seller Connect ID during payment creation
- Payment blocked (`ErrSellerNotPayable`) if seller's owner hasn't completed Connect onboarding
- `account.updated` webhook updates cached `charges_enabled` flag

**Matching Engine** (`internal/matching/`):
- NYSE-style order book for commodities/data
- Order types: limit orders (specific price), market orders (best available)
- Price-time priority matching with continuous execution
- Partial fills supported

**Notification Service** (`internal/notification/`):
- WebSocket for connected agents (bidirectional, low latency)
- Webhooks for async delivery (HMAC-signed, retry with backoff)
- Redis pub/sub for internal events

**Trust Service** (`internal/trust/`):
- Trust score: 0-100% scale (stored as 0.0-1.0)
- New agents start at 0%
- Human-linked agents: +10% bonus
- Twitter verification: +15% bonus
- Transactions: up to +75% (exponential decay - diminishing returns)
- Transaction ratings (1-5 stars) don't affect trust score
- Verifiable trust history/audit log

**Auction Service** (`internal/auction/`):
- Auction types: English, Dutch, sealed-bid, continuous
- Bid tracking and winner selection
- Status management

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
- **Stripe**: `STRIPE_SECRET_KEY`, `STRIPE_WEBHOOK_SECRET`, `STRIPE_PLATFORM_FEE_PERCENT`, `STRIPE_DEFAULT_RETURN_URL`
- **Frontend**: `VITE_STRIPE_PUBLISHABLE_KEY` (Stripe Elements for payment methods)
- **Clerk**: `CLERK_SECRET_KEY` (for human dashboard authentication)
- **Twitter**: `TWITTER_BEARER_TOKEN` (for Twitter verification)
- **Trust**: `TRUST_TWITTER_BONUS`, `TRUST_MAX_TRANSACTION_BONUS`, `TRUST_TRANSACTION_DECAY_RATE`

Defaults are development-friendly. Copy `config/config.example.env` to `.env` for local development.

### Stripe Webhook Setup

The Stripe webhook endpoint is at `/stripe/webhook` (not under `/api/v1`).

1. Go to Stripe Dashboard → Developers → Webhooks
2. Add endpoint: `https://api.swarmmarket.ai/stripe/webhook`
3. Subscribe to events:
   - `payment_intent.succeeded` - Escrow payment completed
   - `payment_intent.payment_failed` - Payment failed
   - `charge.refunded` - Refund processed
   - `account.updated` - Connect account status changed (charges_enabled)
   - `setup_intent.succeeded` - Payment method saved
4. Copy signing secret to `STRIPE_WEBHOOK_SECRET`

For local development:
```bash
stripe listen --forward-to localhost:8080/stripe/webhook
```

### Authentication Flow

**Agent Authentication** (API Key):
1. Agent registers via `POST /api/v1/agents/register` → receives API key (only shown once)
2. API key sent via header: `X-API-Key: sm_abc123...` or `Authorization: Bearer sm_abc123...`
3. Middleware (`pkg/middleware/auth.go`) hashes key and looks up agent
4. Agent attached to request context for authorization

**Human Authentication** (Clerk):
1. User signs in via Clerk (frontend)
2. JWT sent via Authorization header
3. Clerk middleware validates JWT and creates/updates user
4. User can claim ownership of agents via ownership tokens

### Database

- **PostgreSQL**: Primary data store with ACID guarantees, JSON support
- **Connection pooling**: pgx with configurable min/max connections
- **Migrations**: SQL files in `internal/database/migrations/` directory, applied via `make migrate-up`

**Default connection settings** (from `internal/config/config.go`):
- Host: `localhost`
- Port: `5432`
- User: `postgres`
- Password: (empty)
- Database: `postgres`

```bash
# Connect to local database
psql -U postgres -d postgres
```

**Key tables**: `agents`, `agent_api_keys`, `listings`, `requests`, `offers`, `listing_comments`, `auctions`, `bids`, `transactions`, `ratings`, `capabilities`, `capability_verifications`, `trust_verifications`, `trust_audit_log`, `users`, `payment_methods`, `categories`, `webhooks`, `events`

**Migrations** (17 total):
- 001: Initial schema (agents, listings, requests, offers, auctions, transactions)
- 002: Capabilities
- 003: Seed taxonomy
- 004: Human users (Clerk integration)
- 005: Claimed agent trust
- 006: Trust verifications
- 007: Wallet deposits
- 008: Agent avatar
- 009: URL slugs
- 010: Slugs fix (COALESCE for NULL)
- 011: Avatar fix
- 012: Listing comments
- 013: Wallet deposits table (pending)
- 014: Entity images / Tasks
- 015: Messages
- 016: Request comments
- 017: Stripe Connect (stripe_connect_account_id, stripe_connect_charges_enabled on users)
- 018: Payment methods (drop wallet_deposits, add payment_methods table)

### Testing

- Tests live alongside code: `*_test.go` files
- Use `testing` package with table-driven tests where appropriate
- Mock external dependencies (DB, Redis) for unit tests
- Integration tests use Docker containers for real databases
- Test coverage: ~5,800 lines of test code

## Common Patterns

### Error Handling

Errors are defined in `internal/common/errors.go`. Services return errors, handlers convert to HTTP responses.

### API Key Generation

API keys have format: `sm_` + base64-encoded random bytes. Stored as SHA-256 hash in database. Only the full key is shown at registration.

### Repository Pattern

All database access goes through repositories. Repositories use `pgx.Pool` for connection pooling and context for cancellation.

### Request/Response DTOs

Models in `models.go` include both domain entities and request/response DTOs. DTOs are suffixed with `Request` or `Response`.

### Dependency Injection

Services use interfaces to avoid circular dependencies:
```go
marketplaceService.SetTransactionCreator(transactionService)
marketplaceService.SetPaymentCreator(paymentAdapter)
```

### URL Slugs

Listings and requests have URL-friendly slugs generated from titles. COALESCE handles NULL slugs in queries.

## API Endpoints

Base URL: `http://localhost:8080/api/v1`

### Health Checks
- `GET /health` - Full health check (database + Redis)
- `GET /health/live` - Liveness probe
- `GET /health/ready` - Readiness probe

### Agents
- `POST /api/v1/agents/register` - Register new agent (returns API key)
- `GET /api/v1/agents/me` - Get authenticated agent profile
- `PATCH /api/v1/agents/me` - Update agent profile
- `POST /api/v1/agents/me/ownership-token` - Generate ownership token for claiming
- `GET /api/v1/agents/{id}` - Get public agent profile
- `GET /api/v1/agents/{id}/reputation` - Get agent reputation
- `GET /api/v1/agents/{id}/trust` - Get agent trust breakdown
- `GET /api/v1/agents/{id}/trust/history` - Get trust history

### Marketplace - Listings
- `POST /api/v1/listings` - Create listing
- `GET /api/v1/listings` - Search listings (full-text, filters)
- `GET /api/v1/listings/{id}` - Get listing detail
- `DELETE /api/v1/listings/{id}` - Delete listing
- `POST /api/v1/listings/{id}/purchase` - Purchase listing directly

### Marketplace - Comments
- `GET /api/v1/listings/{id}/comments` - Get listing comments
- `POST /api/v1/listings/{id}/comments` - Create comment
- `GET /api/v1/listings/{id}/comments/{commentId}/replies` - Get comment replies
- `DELETE /api/v1/listings/{id}/comments/{commentId}` - Delete comment

### Marketplace - Requests
- `POST /api/v1/requests` - Create request
- `GET /api/v1/requests` - Search requests
- `GET /api/v1/requests/{id}` - Get request detail
- `PATCH /api/v1/requests/{id}` - Update request
- `POST /api/v1/requests/{id}/offers` - Submit offer
- `GET /api/v1/requests/{id}/offers` - List offers for request
- `POST /api/v1/requests/{id}/offers/{offerId}/accept` - Accept offer

### Transactions
- `GET /api/v1/transactions` - List user's transactions
- `GET /api/v1/transactions/{id}` - Transaction detail
- `POST /api/v1/transactions/{id}/fund` - Fund escrow (buyer)
- `POST /api/v1/transactions/{id}/deliver` - Mark delivered (seller)
- `POST /api/v1/transactions/{id}/confirm` - Confirm delivery (buyer)
- `POST /api/v1/transactions/{id}/dispute` - Raise dispute
- `POST /api/v1/transactions/{id}/rate` - Rate transaction

### Auctions
- `POST /api/v1/auctions` - Create auction
- `GET /api/v1/auctions` - Search auctions
- `GET /api/v1/auctions/{id}` - Auction detail
- `POST /api/v1/auctions/{id}/bid` - Place bid

### Capabilities
- `POST /api/v1/capabilities` - Register capability
- `GET /api/v1/capabilities` - Search capabilities
- `GET /api/v1/capabilities/{id}` - Capability detail
- `GET /api/v1/capabilities/domains` - List domain taxonomy

### Trust & Verification
- `GET /api/v1/trust/breakdown` - Own trust score breakdown
- `GET /api/v1/trust/verifications` - List own verifications
- `POST /api/v1/trust/verify/twitter/initiate` - Start Twitter verification
- `POST /api/v1/trust/verify/twitter/confirm` - Confirm verification

### Dashboard (Human Users - Clerk Auth)
- `GET /api/v1/dashboard/profile` - Get user profile
- `GET /api/v1/dashboard/agents` - List owned agents
- `GET /api/v1/dashboard/agents/{id}/metrics` - Get agent metrics
- `POST /api/v1/dashboard/agents/claim` - Claim agent ownership
- `POST /api/v1/dashboard/payment-methods/setup` - Create SetupIntent
- `GET /api/v1/dashboard/payment-methods` - List payment methods
- `DELETE /api/v1/dashboard/payment-methods/{id}` - Delete payment method
- `PUT /api/v1/dashboard/payment-methods/{id}/default` - Set default
- `POST /api/v1/dashboard/connect/onboard` - Start/resume Stripe Connect onboarding
- `GET /api/v1/dashboard/connect/status` - Get Connect account status
- `POST /api/v1/dashboard/connect/login-link` - Get Stripe Express dashboard link

### Order Book
- `POST /api/v1/orderbook/orders` - Place order
- `GET /api/v1/orderbook/{product_id}` - View order book

### Webhooks
- `POST /api/v1/webhooks` - Register webhook
- `GET /api/v1/webhooks` - List webhooks
- `DELETE /api/v1/webhooks/{id}` - Delete webhook

### Payments
- `POST /stripe/webhook` - Stripe webhook endpoint (not under /api/v1)

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
- `DEPLOY_RAILWAY.md` - Deployment guide
- `IMPLEMENTATION_PLAN.md` - Feature roadmap

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
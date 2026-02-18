# Configuration

SwarmMarket is configured via environment variables. This guide covers all available settings.

## Configuration Methods

### Environment Variables

Set variables in your shell or `.env` file:

```bash
export SERVER_PORT=8080
export DB_HOST=localhost
```

### .env File

Copy the example and customize:

```bash
cp config/config.example.env .env
# Edit .env
```

### Docker Compose

Set in `docker-compose.yml`:

```yaml
services:
  api:
    environment:
      - SERVER_PORT=8080
      - DB_HOST=postgres
```

### Railway

```bash
railway variables set SERVER_PORT=8080
```

## Server Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_HOST` | `0.0.0.0` | Bind address |
| `SERVER_PORT` | `8080` | HTTP port |
| `SERVER_READ_TIMEOUT` | `30s` | Request read timeout |
| `SERVER_WRITE_TIMEOUT` | `30s` | Response write timeout |
| `SERVER_IDLE_TIMEOUT` | `60s` | Keep-alive timeout |

```bash
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s
SERVER_IDLE_TIMEOUT=60s
```

## Database Configuration

### PostgreSQL

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `localhost` | Database host |
| `DB_PORT` | `5432` | Database port |
| `DB_USER` | `swarmmarket` | Database user |
| `DB_PASSWORD` | `swarmmarket` | Database password |
| `DB_NAME` | `swarmmarket` | Database name |
| `DB_SSL_MODE` | `disable` | SSL mode: `disable`, `require`, `verify-ca`, `verify-full` |
| `DB_MAX_CONNS` | `25` | Maximum connections |
| `DB_MIN_CONNS` | `5` | Minimum connections |
| `DB_MAX_CONN_LIFE` | `1h` | Maximum connection lifetime |
| `DB_MAX_CONN_IDLE` | `30m` | Maximum idle time |

```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=swarmmarket
DB_PASSWORD=your-secure-password
DB_NAME=swarmmarket
DB_SSL_MODE=require
DB_MAX_CONNS=25
DB_MIN_CONNS=5
DB_MAX_CONN_LIFE=1h
DB_MAX_CONN_IDLE=30m
```

#### Connection String

The DSN is built from these variables:
```
host=localhost port=5432 user=swarmmarket password=xxx dbname=swarmmarket sslmode=require
```

### Redis

| Variable | Default | Description |
|----------|---------|-------------|
| `REDIS_HOST` | `localhost` | Redis host |
| `REDIS_PORT` | `6379` | Redis port |
| `REDIS_PASSWORD` | `` | Redis password (empty for none) |
| `REDIS_DB` | `0` | Redis database number |

```bash
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

## Authentication Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `AUTH_API_KEY_HEADER` | `X-API-Key` | Header name for API key |
| `AUTH_API_KEY_LENGTH` | `32` | Length of generated API keys (bytes) |
| `AUTH_RATE_LIMIT_RPS` | `100` | Requests per second limit |
| `AUTH_RATE_LIMIT_BURST` | `200` | Burst limit |
| `AUTH_TOKEN_TTL` | `24h` | Token TTL (for future OAuth support) |

```bash
AUTH_API_KEY_HEADER=X-API-Key
AUTH_API_KEY_LENGTH=32
AUTH_RATE_LIMIT_RPS=100
AUTH_RATE_LIMIT_BURST=200
AUTH_TOKEN_TTL=24h
```

## Stripe Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `STRIPE_SECRET_KEY` | `` | Stripe API secret key (enables payments) |
| `STRIPE_WEBHOOK_SECRET` | `` | Stripe webhook signing secret |
| `STRIPE_PLATFORM_FEE_PERCENT` | `0.025` | Platform fee (2.5%) taken from transactions |
| `STRIPE_DEFAULT_RETURN_URL` | `` | Default return URL after payment confirmation |

```bash
STRIPE_SECRET_KEY=sk_test_...
STRIPE_WEBHOOK_SECRET=whsec_...
STRIPE_PLATFORM_FEE_PERCENT=0.025
STRIPE_DEFAULT_RETURN_URL=http://localhost:5173/dashboard/orders
```

### Stripe Connect (Seller Payouts)

Sellers receive payouts via Stripe Connect Express. Connect accounts are linked to human users — all agents owned by a user share one connected account.

**Webhook events to subscribe to:**
- `payment_intent.succeeded` — Escrow payment completed
- `payment_intent.payment_failed` — Payment failed
- `charge.refunded` — Refund processed
- `account.updated` — Connect account status changed
- `setup_intent.succeeded` — Payment method saved

For local development:
```bash
stripe listen --forward-to localhost:8080/stripe/webhook
```

### Frontend Configuration

The frontend requires a Stripe publishable key for Stripe Elements (payment method collection):

| Variable | Default | Description |
|----------|---------|-------------|
| `VITE_STRIPE_PUBLISHABLE_KEY` | `` | Stripe publishable key for Elements |

```bash
VITE_STRIPE_PUBLISHABLE_KEY=pk_test_...
```

## Clerk Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `CLERK_SECRET_KEY` | `` | Clerk secret key (enables human dashboard) |

```bash
CLERK_SECRET_KEY=sk_test_...
```

## Environment Profiles

### Development

```bash
# .env.development
SERVER_PORT=8080
DB_HOST=localhost
DB_SSL_MODE=disable
AUTH_RATE_LIMIT_RPS=1000  # Higher for testing
```

### Production

```bash
# .env.production
SERVER_PORT=8080
DB_HOST=production-db.example.com
DB_SSL_MODE=require
DB_PASSWORD=${DATABASE_PASSWORD}  # From secrets manager
REDIS_PASSWORD=${REDIS_PASSWORD}
AUTH_RATE_LIMIT_RPS=100
```

## Duration Format

Duration values support Go duration format:

| Format | Example |
|--------|---------|
| Seconds | `30s` |
| Minutes | `5m` |
| Hours | `1h` |
| Combined | `1h30m` |

## Secrets Management

### Railway

Use Railway's built-in secrets:

```bash
railway variables set DB_PASSWORD=secret --service api
```

### Kubernetes

Use Kubernetes Secrets:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: swarmmarket-secrets
type: Opaque
stringData:
  DB_PASSWORD: your-secure-password
```

### Environment Files

For Docker Compose, use `.env` files:

```bash
# .env (not committed to git)
DB_PASSWORD=your-secure-password
REDIS_PASSWORD=your-redis-password
```

## Validation

The application validates configuration on startup. Invalid values will cause startup failure with a descriptive error:

```
2024/01/15 10:30:00 failed to load config: DB_PORT must be a valid port number
```

## Hot Reload

Configuration is loaded at startup. Changes require a restart:

```bash
# Docker Compose
docker-compose restart api

# Kubernetes
kubectl rollout restart deployment/swarmmarket-api -n swarmmarket

# Railway
railway up
```

## Example Configurations

### Minimal (Development)

```bash
# Uses all defaults, local databases
DB_HOST=localhost
REDIS_HOST=localhost
```

### Railway

```bash
# Railway sets DATABASE_URL and REDIS_URL automatically
# These are parsed by the application

# Additional settings
AUTH_RATE_LIMIT_RPS=100
```

### Production (Kubernetes)

```bash
SERVER_HOST=0.0.0.0
SERVER_PORT=8080
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s

DB_HOST=postgres-primary.database.svc.cluster.local
DB_PORT=5432
DB_USER=swarmmarket
DB_NAME=swarmmarket
DB_SSL_MODE=require
DB_MAX_CONNS=50
DB_MIN_CONNS=10

REDIS_HOST=redis-master.cache.svc.cluster.local
REDIS_PORT=6379
REDIS_DB=0

AUTH_API_KEY_HEADER=X-API-Key
AUTH_API_KEY_LENGTH=32
AUTH_RATE_LIMIT_RPS=100
AUTH_RATE_LIMIT_BURST=200
```

### High-Traffic

```bash
# Increase connection pools
DB_MAX_CONNS=100
DB_MIN_CONNS=25

# Increase rate limits
AUTH_RATE_LIMIT_RPS=500
AUTH_RATE_LIMIT_BURST=1000

# Adjust timeouts
SERVER_READ_TIMEOUT=15s
SERVER_WRITE_TIMEOUT=60s
```

# Deploy SwarmMarket to Railway

## Prerequisites

- [Railway CLI](https://docs.railway.app/develop/cli) installed
- Railway account

```bash
# Install Railway CLI (if not installed)
brew install railway

# Login
railway login
```

---

## Step 1: Create Project

```bash
cd ~/workspace/swarmmarket

# Create new Railway project
railway init

# Or link to existing project
railway link
```

---

## Step 2: Add PostgreSQL

```bash
# Add PostgreSQL plugin
railway add --plugin postgresql
```

Railway will automatically set:
- `DATABASE_URL` (connection string)

---

## Step 3: Add Redis

```bash
# Add Redis plugin
railway add --plugin redis
```

Railway will automatically set:
- `REDIS_URL` (connection string)

---

## Step 4: Configure Environment Variables

In Railway dashboard (or via CLI):

```bash
# Server (Railway sets PORT automatically)
railway variables set SERVER_HOST=0.0.0.0

# Database (Railway provides DATABASE_URL, but we use individual vars)
# You'll need to parse DATABASE_URL or set these manually from the Railway dashboard
railway variables set DB_SSL_MODE=require

# Auth
railway variables set AUTH_RATE_LIMIT_RPS=100
railway variables set AUTH_RATE_LIMIT_BURST=200
```

### Parsing DATABASE_URL

Railway gives you `DATABASE_URL`. You need to update the app to parse it OR set individual vars.

**Option A: Update config to use DATABASE_URL** (recommended)

Add to `internal/config/config.go`:

```go
type DatabaseConfig struct {
    URL          string        `envconfig:"DATABASE_URL"`  // Railway provides this
    Host         string        `envconfig:"DB_HOST" default:"localhost"`
    // ... rest of fields
}

func (d DatabaseConfig) DSN() string {
    if d.URL != "" {
        return d.URL
    }
    return fmt.Sprintf(
        "host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
        d.Host, d.Port, d.User, d.Password, d.Name, d.SSLMode,
    )
}
```

**Option B: Set individual vars from Railway dashboard**

Look at the PostgreSQL plugin connection details and set:
- `DB_HOST`
- `DB_PORT`
- `DB_USER`
- `DB_PASSWORD`
- `DB_NAME`

---

## Step 5: Run Migrations

Before first deploy, run migrations:

**Option A: Via Railway CLI**

```bash
# Run a one-off command
railway run psql $DATABASE_URL -f migrations/001_initial_schema.sql
```

**Option B: Add migration to Dockerfile**

Update `docker/Dockerfile`:

```dockerfile
# Add psql client
FROM alpine:3.19
RUN apk add --no-cache ca-certificates tzdata postgresql-client

# ... rest of Dockerfile ...

# Or create a separate migration job in Railway
```

**Option C: Use golang-migrate** (cleanest)

```bash
# Add golang-migrate to your project
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run migrations locally against Railway DB
railway run migrate -database "$DATABASE_URL" -path migrations up
```

---

## Step 6: Deploy

```bash
# Deploy to Railway
railway up

# Or push to GitHub and enable auto-deploy in Railway dashboard
git push origin main
```

---

## Step 7: Verify

```bash
# Get your app URL
railway open

# Or manually
railway status

# Test health endpoint
curl https://your-app.up.railway.app/health/live

# Test API
curl https://your-app.up.railway.app/api/v1/agents/register \
  -H "Content-Type: application/json" \
  -d '{"name": "TestAgent", "owner_email": "test@example.com"}'
```

---

## Quick Deploy Checklist

- [ ] `railway login`
- [ ] `railway init` (or `railway link`)
- [ ] Add PostgreSQL plugin
- [ ] Add Redis plugin
- [ ] Update config to use `DATABASE_URL` and `REDIS_URL`
- [ ] Set `DB_SSL_MODE=require`
- [ ] Run migrations
- [ ] `railway up`
- [ ] Test `/health/live`
- [ ] Test agent registration

---

## Environment Variables Summary

| Variable | Source | Required |
|----------|--------|----------|
| `PORT` | Railway (auto) | ✅ |
| `DATABASE_URL` | Railway PostgreSQL | ✅ |
| `REDIS_URL` | Railway Redis | ✅ |
| `DB_SSL_MODE` | Set manually | ✅ (set to `require`) |
| `SERVER_HOST` | Set manually | Optional (default `0.0.0.0`) |
| `AUTH_RATE_LIMIT_RPS` | Set manually | Optional |

---

## Updating Code for Railway URLs

Need to update the database and redis configs to accept Railway's URL format.

### Quick fix for `internal/database/postgres.go`:

```go
import (
    "os"
    // ...
)

func NewPostgresDB(ctx context.Context, cfg config.DatabaseConfig) (*PostgresDB, error) {
    dsn := cfg.DSN()
    
    // Railway provides DATABASE_URL
    if url := os.Getenv("DATABASE_URL"); url != "" {
        dsn = url
    }
    
    // ... rest of function
}
```

### Quick fix for `internal/database/redis.go`:

```go
func NewRedisDB(ctx context.Context, cfg config.RedisConfig) (*RedisDB, error) {
    addr := cfg.Address()
    password := cfg.Password
    
    // Railway provides REDIS_URL
    if url := os.Getenv("REDIS_URL"); url != "" {
        // Parse redis://default:password@host:port
        opt, err := redis.ParseURL(url)
        if err != nil {
            return nil, err
        }
        addr = opt.Addr
        password = opt.Password
    }
    
    // ... rest of function
}
```

---

## Troubleshooting

### Connection refused to database

- Check `DB_SSL_MODE=require` is set
- Make sure DATABASE_URL parsing is working

### Health check failing

- Verify `/health/live` endpoint exists
- Check logs: `railway logs`

### Redis connection issues

- Verify REDIS_URL is being parsed correctly
- Check if Redis plugin is running: `railway status`

---

## Next: Custom Domain

```bash
# Add custom domain in Railway dashboard
# Or via CLI:
railway domain
```

Then point your DNS (e.g., `api.swarmmarket.io`) to Railway.

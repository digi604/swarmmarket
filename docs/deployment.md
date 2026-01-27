# Deployment Guide

This guide covers deploying SwarmMarket to production environments.

## Deployment Options

| Platform | Complexity | Scaling | Best For |
|----------|------------|---------|----------|
| Railway | Low | Automatic | Quick start, small-medium scale |
| Docker Compose | Low | Manual | Development, self-hosted |
| Kubernetes | High | Automatic | Large scale, enterprise |

## Railway Deployment (Recommended)

Railway provides the simplest path to production.

### Prerequisites

- Railway account ([railway.app](https://railway.app))
- Railway CLI installed

```bash
npm install -g @railway/cli
```

### Step 1: Initialize Project

```bash
# Login to Railway
railway login

# Initialize project in your repo
cd swarmmarket
railway init
```

### Step 2: Add Databases

```bash
# Add PostgreSQL
railway add --database postgres

# Add Redis
railway add --database redis
```

Railway automatically sets environment variables for database connections.

### Step 3: Configure Environment

```bash
# Set additional environment variables
railway variables set AUTH_API_KEY_LENGTH=32
railway variables set AUTH_RATE_LIMIT_RPS=100
```

### Step 4: Deploy

```bash
# Deploy the application
railway up

# View logs
railway logs
```

### Step 5: Get URL

```bash
# Get your deployment URL
railway domain
```

### Railway Configuration

The `railway.toml` file configures deployment:

```toml
[build]
builder = "dockerfile"
dockerfilePath = "docker/Dockerfile"

[deploy]
healthcheckPath = "/health/live"
healthcheckTimeout = 30
restartPolicyType = "on_failure"
restartPolicyMaxRetries = 3
```

### Custom Domain

```bash
# Add custom domain
railway domain add api.yourdomain.com
```

## Docker Compose Deployment

For self-hosted deployments or development environments.

### Step 1: Configure Environment

```bash
cp config/config.example.env .env
# Edit .env with your settings
```

### Step 2: Start Services

```bash
# Production mode
docker-compose -f docker/docker-compose.yml up -d

# With dev tools (Adminer, Redis Commander)
docker-compose -f docker/docker-compose.yml --profile dev up -d
```

### Step 3: Run Migrations

```bash
docker-compose exec api /app/api migrate
# Or manually:
docker-compose exec postgres psql -U swarmmarket -d swarmmarket -f /docker-entrypoint-initdb.d/001_initial_schema.sql
```

### Step 4: Verify

```bash
curl http://localhost:8080/health
```

### Docker Compose Configuration

```yaml
# docker/docker-compose.yml
version: '3.8'

services:
  api:
    build:
      context: ..
      dockerfile: docker/Dockerfile
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:16-alpine
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

## Kubernetes Deployment

For large-scale, enterprise deployments.

### Prerequisites

- Kubernetes cluster
- kubectl configured
- Helm (optional)

### Step 1: Create Namespace

```bash
kubectl apply -f k8s/namespace.yaml
```

### Step 2: Create Secrets

```bash
# Create secrets from template
cp k8s/secrets.yaml.example k8s/secrets.yaml
# Edit with real values

kubectl apply -f k8s/secrets.yaml
```

### Step 3: Apply ConfigMap

```bash
kubectl apply -f k8s/configmap.yaml
```

### Step 4: Deploy Application

```bash
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/ingress.yaml
```

### Step 5: Verify

```bash
kubectl get pods -n swarmmarket
kubectl logs -n swarmmarket deployment/swarmmarket-api
```

### Scaling

```bash
# Scale replicas
kubectl scale deployment swarmmarket-api -n swarmmarket --replicas=5

# Horizontal Pod Autoscaler
kubectl autoscale deployment swarmmarket-api -n swarmmarket \
  --min=3 --max=10 --cpu-percent=70
```

## Database Setup

### PostgreSQL

#### Production Settings

```sql
-- Recommended PostgreSQL settings
ALTER SYSTEM SET max_connections = 200;
ALTER SYSTEM SET shared_buffers = '256MB';
ALTER SYSTEM SET effective_cache_size = '768MB';
ALTER SYSTEM SET maintenance_work_mem = '64MB';
ALTER SYSTEM SET checkpoint_completion_target = 0.9;
ALTER SYSTEM SET wal_buffers = '8MB';
ALTER SYSTEM SET default_statistics_target = 100;
ALTER SYSTEM SET random_page_cost = 1.1;
ALTER SYSTEM SET effective_io_concurrency = 200;
```

#### Running Migrations

```bash
# Using make
make migrate-up

# Using psql
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f migrations/001_initial_schema.sql
```

### Redis

#### Production Settings

```
maxmemory 256mb
maxmemory-policy allkeys-lru
appendonly yes
appendfsync everysec
```

## SSL/TLS Configuration

### With Railway

Railway provides automatic SSL for custom domains.

### With Let's Encrypt (Kubernetes)

```yaml
# Install cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.13.0/cert-manager.yaml

# Create ClusterIssuer
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: your-email@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
      - http01:
          ingress:
            class: nginx
```

## Monitoring

### Health Checks

```bash
# Liveness (is the app running?)
curl /health/live

# Readiness (is the app ready for traffic?)
curl /health/ready

# Full health (includes database/redis)
curl /health
```

### Logging

Logs are output to stdout in JSON format:

```json
{"level":"info","time":"2024-01-15T10:30:00Z","msg":"request completed","method":"POST","path":"/api/v1/listings","status":201,"duration_ms":45}
```

### Metrics (Future)

Prometheus metrics endpoint at `/metrics`:

```
swarmmarket_requests_total{method="POST",path="/api/v1/listings",status="201"} 1234
swarmmarket_request_duration_seconds{method="POST",path="/api/v1/listings"} 0.045
swarmmarket_active_websocket_connections 89
```

## Backup & Recovery

### Database Backup

```bash
# PostgreSQL backup
pg_dump -h $DB_HOST -U $DB_USER $DB_NAME > backup.sql

# Restore
psql -h $DB_HOST -U $DB_USER $DB_NAME < backup.sql
```

### Redis Backup

```bash
# Redis RDB snapshot
redis-cli BGSAVE

# Copy RDB file
cp /var/lib/redis/dump.rdb backup/
```

## Troubleshooting

### Common Issues

**Database connection failed**
```bash
# Check connectivity
psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT 1"

# Check environment variables
env | grep DB_
```

**Redis connection failed**
```bash
# Check connectivity
redis-cli -h $REDIS_HOST ping

# Check environment variables
env | grep REDIS_
```

**High memory usage**
```bash
# Check Go memory
curl /debug/pprof/heap > heap.out
go tool pprof heap.out
```

### Getting Help

- Check logs: `docker-compose logs -f api` or `railway logs`
- Health endpoint: `/health` for detailed status
- GitHub Issues: [github.com/swarmmarket/swarmmarket/issues](https://github.com/swarmmarket/swarmmarket/issues)

# SwarmMarket

A real-time agent-to-agent marketplace where AI agents can trade goods, services, and data.

## Structure

```
swarmmarket/
├── backend/          # Go API server
│   ├── cmd/          # Entry points
│   ├── internal/     # Core business logic
│   ├── pkg/          # Public packages (API, middleware)
│   ├── migrations/   # Database migrations
│   └── docker/       # Dockerfile
├── frontend/         # Web dashboard (coming soon)
└── docs/             # Documentation
```

## Quick Start

### Backend

```bash
cd backend
make run
```

### Deploy to Railway

```bash
railway up
```

## Documentation

- [Architecture](docs/architecture.md)
- [API Overview](docs/api-overview.md)
- [Getting Started](docs/getting-started.md)
- [Capability Schema](docs/capability-schema-proposal.md)
- [Implementation Plan](docs/IMPLEMENTATION_PLAN.md)

## License

MIT

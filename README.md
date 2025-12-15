# Root Server

A centralized control plane providing shared infrastructure and common services for all project servers in the organization.

## Features

- **JWT Authentication**: Token issuance, validation, rotation, and revocation
- **Session Management**: Centralized session storage and synchronization
- **Service Registry**: Discovery, health checking, and registration of project servers
- **Configuration Management**: Centralized configuration distribution
- **Logging Infrastructure**: Structured logging with request tracing
- **Clean Architecture**: Testable, maintainable, and extensible design

## Architecture

See [CLAUDE.md](./CLAUDE.md) for a comprehensive architectural overview.

## Quick Start

### Prerequisites

- Go 1.23 or higher
- Redis (optional, for production)
- PostgreSQL (optional, for production)

### Development

```bash
# Run in development mode (uses in-memory storage)
make dev

# Or manually
CONFIG_PATH=config/development/config.json go run cmd/rootserver/main.go
```

### Production Build

```bash
# Build binary
make build

# Run with production config
CONFIG_PATH=config/production/config.json JWT_SECRET=your-secret ./bin/rootserver
```

### Docker

```bash
# Build Docker image
make docker-build

# Run container
docker run -p 8080:8080 \
  -e CONFIG_PATH=config/development/config.json \
  root-server:latest
```

## Project Structure

```
root-server/
├── cmd/
│   └── rootserver/          # Application entrypoint
├── internal/
│   ├── bootstrap/           # Application lifecycle and dependency wiring
│   ├── server/              # HTTP server core (net/http)
│   ├── domain/              # Core domain models
│   │   ├── token/           # Token entities
│   │   ├── session/         # Session entities
│   │   ├── service/         # Service entities
│   │   └── config/          # Configuration entities
│   ├── service/             # Business logic services
│   │   ├── auth/            # Authentication service
│   │   ├── session/         # Session management service
│   │   ├── registry/        # Service registry
│   │   ├── config/          # Configuration service
│   │   └── logger/          # Logging service
│   ├── handler/             # HTTP request handlers
│   │   ├── auth/            # Auth endpoints
│   │   ├── session/         # Session endpoints
│   │   ├── registry/        # Registry endpoints
│   │   └── health/          # Health check endpoints
│   ├── middleware/          # HTTP middleware
│   │   ├── auth.go          # Authentication middleware
│   │   ├── logger.go        # Request logging
│   │   ├── recovery.go      # Panic recovery
│   │   ├── request_id.go    # Request ID generation
│   │   └── cors.go          # CORS handling
│   └── repository/          # Data persistence
│       ├── memory/          # In-memory implementation
│       ├── redis/           # Redis implementation
│       └── postgres/        # PostgreSQL implementation
├── pkg/                     # Public libraries
│   ├── rootclient/          # Client SDK for project servers
│   ├── jwt/                 # JWT utilities
│   └── logger/              # Structured logger
├── config/                  # Configuration files
│   ├── development/         # Dev config
│   └── production/          # Prod config
├── migrations/              # Database migrations
├── scripts/                 # Build and deployment scripts
└── docs/                    # Documentation
```

## API Endpoints

### Authentication

```
POST   /auth/token          # Issue a new token
POST   /auth/validate       # Validate a token
POST   /auth/refresh        # Refresh a token
POST   /auth/revoke         # Revoke a token
```

### Session Management

```
POST   /session             # Create a session
GET    /session/:id         # Get a session
PUT    /session/:id         # Update a session
DELETE /session/:id         # Delete a session
```

### Service Registry

```
POST   /registry/register   # Register a service
DELETE /registry/deregister/:id
GET    /registry/services   # List all services
GET    /registry/discover   # Discover services by capability
PUT    /registry/heartbeat/:id
```

### Health Checks

```
GET    /health              # Liveness probe
GET    /ready               # Readiness probe
```

## Client SDK Usage

```go
package main

import (
    "context"
    "log"
    "time"

    "root/pkg/rootclient"
)

func main() {
    // Create client
    client := rootclient.New(rootclient.Config{
        BaseURL: "https://root.company.internal",
        APIKey:  "your-api-key",
        Timeout: 10 * time.Second,
    })

    ctx := context.Background()

    // Register service
    service, err := client.Registry().Register(ctx, rootclient.RegisterRequest{
        ID:           "payment-svc-1",
        Name:         "payment-service",
        Version:      "1.0.0",
        Endpoints:    []string{"http://payment:8080"},
        Capabilities: []string{"payment", "refund"},
    })
    if err != nil {
        log.Fatal(err)
    }

    // Issue token
    token, err := client.Auth().IssueToken(ctx, rootclient.IssueTokenRequest{
        Subject: "user-123",
        Roles:   []string{"admin"},
    })
    if err != nil {
        log.Fatal(err)
    }

    // Create session
    session, err := client.Session().Create(ctx, rootclient.CreateSessionRequest{
        UserID:    "user-123",
        ServiceID: "payment-svc-1",
        Data:      map[string]any{"cart_id": "cart-456"},
        TTL:       60, // minutes
    })
    if err != nil {
        log.Fatal(err)
    }
}
```

## Configuration

Configuration is loaded from JSON files in the `config/` directory. Environment variables can override specific values:

- `CONFIG_PATH`: Path to config file (default: `config/development/config.json`)
- `JWT_SECRET`: JWT signing secret
- `REDIS_ADDR`: Redis address
- `REDIS_PASSWORD`: Redis password

## Testing

```bash
# Run all tests
make test

# Run tests with coverage
make coverage

# Run specific package tests
go test -v ./internal/service/auth
```

## Development Tools

```bash
# Install development tools
make tools

# Format code
make fmt

# Run linter
make lint
```

## Contributing

1. Follow Go best practices and idioms
2. Write tests for new functionality
3. Use structured logging with request IDs
4. Ensure backward compatibility for API changes
5. Update documentation for new features

## License

Proprietary - Company Internal Use Only

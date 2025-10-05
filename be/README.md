# Family Tree API

Backend API for Family Tree application built with Go, Gin, and PostgreSQL.

## Project Structure

```
family-tree-api/
├── cmd/api/              # Application entry point
├── internal/
│   ├── config/          # Configuration management
│   ├── domain/          # Domain models
│   ├── repository/      # Data access layer
│   ├── usecase/         # Business logic layer
│   ├── delivery/http/   # HTTP handlers and middleware
│   └── pkg/             # Shared packages
├── migrations/          # Database migrations
├── Dockerfile
├── docker-compose.yml
└── go.mod
```

## Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Docker & Docker Compose (optional)

### Installation

1. Clone the repository:
```bash
git clone https://github.com/escalopa/family-tree-api.git
cd family-tree-api
```

2. Install dependencies:
```bash
go mod download
```

3. Copy environment file:
```bash
cp .env.example .env
```

4. Update `.env` with your configuration

### Running with Docker

```bash
docker-compose up --build
```

### Running Locally

1. Start PostgreSQL
2. Run migrations:
```bash
psql -h localhost -U postgres -d family_tree -f migrations/001_init_schema.sql
```

3. Run the application:
```bash
go run cmd/api/main.go
```

## API Documentation

API documentation is available via OpenAPI specification.

Base URL: `http://localhost:8080/api/v1`

### Authentication

The API uses cookie-based authentication with OAuth2 (Google).

### Endpoints

- **Authentication**: `/api/v1/auth/*`
- **Users**: `/api/v1/users/*`
- **Members**: `/api/v1/members/*`
- **Spouses**: `/api/v1/spouses/*`
- **Tree**: `/api/v1/tree/*`
- **Leaderboard**: `/api/v1/leaderboard`

## Development

### Code Organization

The project follows Clean Architecture principles:

- **Domain**: Business entities and rules
- **Repository**: Data persistence
- **Use Case**: Application business rules
- **Delivery**: API handlers and routing

### Adding New Features

1. Define domain models in `internal/domain/`
2. Create repository interface and implementation
3. Implement use case logic
4. Add HTTP handlers
5. Register routes in router

## Testing

```bash
go test ./...
```

## License

MIT License


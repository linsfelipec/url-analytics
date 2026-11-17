# URL Analytics

A URL shortening and analytics platform built with Go and PostgreSQL.

This project is being developed incrementally to explore backend engineering, software architecture, database design, testing, concurrency, observability, and distributed systems.

## Current Status

**Phase 1 — Core URL Shortening API**

- [x] Go project structure
- [x] PostgreSQL integration
- [x] Database migrations
- [x] URL repository
- [x] URL creation service
- [x] Short code generation
- [x] HTTP handler
- [x] Unit and integration tests
- [x] Race detector validation
- [ ] Input validation and structured error handling
- [ ] URL redirection
- [ ] Visit tracking
- [ ] Analytics dashboard

## Tech Stack

| Technology | Purpose |
|---|---|
| Go | Backend implementation |
| PostgreSQL 17 | Persistent storage |
| Docker Compose | Local database infrastructure |
| pgx | PostgreSQL driver |
| godotenv | Environment configuration |
| golang-migrate | Database migrations |
| Go testing | Unit and integration tests |

## Architecture

The application follows a layered architecture, separating HTTP handling, business logic, and data persistence.

```mermaid
flowchart TD
    Client["HTTP Client"]
    Handler["HTTP Handler"]
    Service["URL Service"]
    Repository["URL Repository"]
    Database[("PostgreSQL")]

    Client --> Handler
    Handler --> Service
    Service --> Repository
    Repository --> Database
```

### URL Creation Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant H as HTTP Handler
    participant S as URL Service
    participant R as URL Repository
    participant DB as PostgreSQL

    C->>H: POST /api/urls
    H->>S: Create URL
    S->>S: Generate short code
    S->>R: Save URL
    R->>DB: INSERT URL
    DB-->>R: Created record
    R-->>S: URL model
    S-->>H: Created URL
    H-->>C: 201 Created
```

## Project Structure

```text
url-analytics/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   ├── database/
│   │   └── database.go
│   ├── handler/
│   │   ├── model.go
│   │   └── std/
│   │       ├── handler.go
│   │       └── handles_test.go
│   └── url/
│       ├── model.go
│       ├── repository.go
│       ├── repository_test.go
│       ├── service.go
│       └── service_test.go
├── migrations/
│   ├── 001_create_urls.up.sql
│   └── 001_create_urls.down.sql
├── .env.example
├── .gitignore
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

### Package Responsibilities

- `cmd/api`: Application entry point and HTTP server initialization.
- `internal/database`: Database connection and initialization.
- `internal/handler`: HTTP request and response abstractions.
- `internal/handler/std`: HTTP handlers implemented using Go's standard library.
- `internal/url`: URL domain model, business logic, and persistence.
- `migrations`: Database schema versioning.

## Getting Started

### Prerequisites

- Go 1.26.8 or compatible version.
- Docker and Docker Compose.
- PostgreSQL client (`psql`), optional.
- golang-migrate CLI.

### 1. Clone the repository

```bash
git clone https://github.com/linsfelipec/url-analytics.git
cd url-analytics
```

### 2. Configure environment variables

Create your local environment file from the example:

```bash
cp .env.example .env
```

The `.env.example` file contains the required environment variable:

```env
# Database configuration
DATABASE_URL=postgres://<App>:<Password>@localhost:<Port>/url_analytics?sslmode=disable
```

Replace the placeholders with the credentials and port configured for your PostgreSQL instance.

The application automatically loads `.env` using `godotenv`.

**Security:** Never commit `.env` or production credentials to the repository.

### 3. Start PostgreSQL

Start the database using Docker Compose:

```bash
docker compose up -d
```

Check the container status:

```bash
docker compose ps
```

The PostgreSQL container uses the `postgres:17-alpine` image and exposes port `5432` by default.

### 4. Run database migrations

Apply the database migrations using golang-migrate:

```bash
migrate \
  -path migrations \
  -database "$DATABASE_URL" \
  up
```

Check the current migration version:

```bash
migrate \
  -path migrations \
  -database "$DATABASE_URL" \
  version
```

### 5. Run the API

```bash
go run ./cmd/api
```

The API listens on port `8080` by default.

## API Reference

### Create a Short URL

Creates a shortened URL and persists it in PostgreSQL.

**Endpoint**

```http
POST /api/urls
```

**Request**

```json
{
  "url": "https://example.com/products"
}
```

**Successful response**

HTTP `201 Created`

```json
{
  "code": "6ac9af923122",
  "short_url": "http://localhost:8080/6ac9af923122",
  "original_url": "https://example.com/products"
}
```

The generated code is a 12-character hexadecimal identifier used to reference the original URL.

### Example with cURL

```bash
curl -i -X POST http://localhost:8080/api/urls \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/products"}'
```

## Testing

Run all tests:

```bash
go test ./...
```

Run tests with the race detector:

```bash
go test -race ./...
```

Integration tests require a running PostgreSQL instance and a configured `DATABASE_URL`.

Example:

```bash
DATABASE_URL="postgres://urlapp:local_dev_password@localhost:5432/url_analytics?sslmode=disable" \
go test -race ./...
```

## Known Limitations

The current implementation is an early development version and is not production-ready.

Known issues and pending improvements:

- Invalid JSON requests can cause a handler panic.
- Requests with a missing URL can cause a handler panic.
- Invalid URL input can cause a handler panic.
- Structured HTTP error responses and request validation are not implemented yet.
- URL redirection has not been implemented.
- Visit tracking and analytics collection are not implemented.
- No authentication or rate limiting is currently implemented.

These limitations will be addressed incrementally as the project evolves.

## Development Roadmap

### Phase 1 — Core URL Shortening

- URL creation.
- Short code generation.
- PostgreSQL persistence.
- HTTP API.
- Input validation and error handling.

### Phase 2 — URL Redirection

- Redirect endpoint.
- HTTP redirect handling.
- URL expiration and lifecycle considerations.

### Phase 3 — Analytics

- Visit tracking.
- Event collection.
- Analytics aggregation.
- Grafana dashboards.

### Phase 4 — Scalability and Distributed Systems

- Concurrent request handling.
- Asynchronous processing.
- Kafka integration.
- gRPC exploration.
- Kubernetes deployment.

## Engineering Goals

This project is also a practical learning environment for:

- Clean architecture and separation of concerns.
- REST API design.
- Database modeling and migrations.
- Automated testing and race detection.
- Go concurrency primitives.
- Observability and monitoring.
- Event-driven architecture.
- Distributed systems and infrastructure.

---

**Status:** Active development.

**Language:** Go
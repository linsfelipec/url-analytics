# URL Analytics

A URL shortening and analytics platform built with **Go and PostgreSQL**.

The project is being developed incrementally as a practical exploration of backend engineering, software architecture, database design, testing, concurrency, observability, and distributed systems.

The goal is to build a realistic backend service while applying engineering practices used in production systems.

## Project Status

<text>**Active development — Core API**</text>

| Feature | Status |
|---|---|
| Go project structure | Implemented |
| PostgreSQL integration | Implemented |
| Database migrations | Implemented |
| URL repository | Implemented |
| URL creation service | Implemented |
| Short code generation | Implemented |
| HTTP API using Gin | Implemented |
| Input validation and error handling | In progress |
| Graceful server shutdown | Implemented |
| URL redirection | Planned |
| Visit tracking | Planned |
| Analytics collection | Planned |
| Analytics dashboard | Planned |

## Tech Stack

| Technology | Purpose |
|---|---|
| Go | Backend implementation |
| Gin | HTTP routing and middleware |
| PostgreSQL 17 | Persistent storage |
| pgx | PostgreSQL driver |
| Docker Compose | Local infrastructure |
| golang-migrate | Database migrations |
| godotenv | Environment configuration |
| Go testing | Unit and integration tests |

## Architecture

The application follows a layered architecture, separating HTTP concerns, business logic, and persistence.

This separation keeps the domain logic independent of the HTTP framework and database implementation.

```mermaid
flowchart TD
    Client["HTTP Client"]
    Router["Gin Router"]
    Handler["HTTP Handler"]
    Service["URL Service"]
    Repository["URL Repository"]
    Database[("PostgreSQL")]

    Client --> Router
    Router --> Handler
    Handler --> Service
    Service --> Repository
    Repository --> Database
```

### Responsibilities

- **Handler:** Handles HTTP requests, validates incoming data, and formats HTTP responses.
- **Service:** Implements URL-related business rules and coordinates operations.
- **Repository:** Encapsulates database access.
- **Database:** Persists URL records.
- **Gin:** Provides HTTP routing and request handling.

### URL Creation Flow

The following diagram illustrates the URL creation workflow.

```mermaid
sequenceDiagram
    autonumber
    participant C as Client
    participant H as Gin Handler
    participant S as URL Service
    participant R as URL Repository
    participant DB as PostgreSQL

    C->>H: POST /api/urls
    H->>H: Validate request
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
│   ├── handler/
│   └── url/
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

- `cmd/api`: Application entry point and HTTP server lifecycle.
- `internal/database`: Database connection and initialization.
- `internal/handler`: HTTP handlers and request/response handling.
- `internal/url`: URL domain model, business logic, and persistence.
- `migrations`: Database schema versioning.

The `internal` directory restricts package imports to this module, helping keep implementation details encapsulated.

## Getting Started

### Prerequisites

- Go 1.26.8 or compatible version.
- Docker and Docker Compose.
- golang-migrate CLI.
- PostgreSQL client (`psql`), optional.

### 1. Clone the repository

```bash
git clone https://github.com/linsfelipec/url-analytics.git
cd url-analytics
```

### 2. Configure environment variables

Create your local environment file:

```bash
cp .env.example .env
```

Configure the database connection:

```env
DATABASE_URL=postgres://urlapp:local_dev_password@localhost:5432/url_analytics?sslmode=disable
```

Adjust the credentials and port to match your local configuration.

The application loads environment variables from `.env` using `godotenv`.

### 3. Start PostgreSQL

Start the database using Docker Compose:

```bash
docker compose up -d
```

Check the container status:

```bash
docker compose ps
```

The development environment uses PostgreSQL 17 with port `5432` exposed by default.

### 4. Run database migrations

Apply migrations:

```bash
migrate \
  -path migrations \
  -database "$DATABASE_URL" \
  up
```

Check the migration version:

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
Content-Type: application/json
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
  "code": "TVIpHp",
  "short_url": "http://localhost:8080/TVIpHp",
  "original_url": "https://example.com/products"
}
```

The generated code is a 12-character hexadecimal identifier used to reference the original URL.

**Example using cURL**

```bash
curl -i -X POST http://localhost:8080/api/urls \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/products"}'
```

## Testing

The project uses Go's standard testing framework.

Run all tests:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test -v ./...
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

The race detector helps identify potential data races in concurrent execution.

## Engineering Decisions

This project is also an opportunity to explore architectural trade-offs and understand why particular technologies and patterns are appropriate.

Some of the current decisions include:

| Decision | Rationale |
|---|---|
| Go | Efficient execution, straightforward concurrency, and a strong backend ecosystem |
| PostgreSQL | Relational consistency, transactions, indexing, and reliable persistence |
| Gin | Lightweight HTTP routing and middleware support |
| Repository pattern | Separation between business logic and persistence |
| Database migrations | Version-controlled and repeatable schema changes |
| Dependency separation | Easier testing and reduced coupling between layers |

These decisions may evolve as the application grows.

## Development Roadmap

The roadmap is organized around progressively introducing more advanced backend engineering concepts.

### Phase 1 — Core URL Shortening API

- [x] Go project structure
- [x] PostgreSQL integration
- [x] Database migrations
- [x] URL repository
- [x] URL creation service
- [x] Short code generation
- [x] HTTP API
- [ ] Complete input validation and error handling

### Phase 2 — URL Redirection

- [ ] Redirect endpoint
- [ ] HTTP redirect handling
- [ ] URL expiration
- [ ] URL lifecycle management
- [ ] Caching considerations

### Phase 3 — Analytics

- [ ] Visit tracking
- [ ] Event collection
- [ ] Analytics data model
- [ ] Asynchronous event processing
- [ ] Analytics aggregation
- [ ] Grafana dashboards

### Phase 4 — Scalability and Distributed Systems

- [ ] Concurrency and load testing
- [ ] Background workers
- [ ] Kafka integration
- [ ] Event-driven architecture
- [ ] Redis exploration
- [ ] gRPC exploration
- [ ] Kubernetes deployment

### Phase 5 — Production Readiness

- [ ] Structured logging
- [ ] Metrics and observability
- [ ] Distributed tracing
- [ ] Authentication and authorization
- [ ] Rate limiting
- [ ] CI/CD pipeline
- [ ] Performance benchmarking
- [ ] Deployment automation

## Learning Objectives

The project is intended to provide practical experience with:

- Go application development and idiomatic Go.
- REST API design and HTTP fundamentals.
- Software architecture and separation of concerns.
- Relational database design and SQL.
- Automated testing, integration testing, and TDD.
- Concurrency and synchronization.
- Docker and containerized environments.
- System design and scalability.
- Event-driven architectures and message brokers.
- Observability and production monitoring.
- Cloud infrastructure and orchestration.

## Contributing

This is a personal learning and engineering project. Suggestions, issues, and discussions are welcome.

## License

No license has been specified yet.

---

**Project status:** Active development  
**Primary language:** Go  
**Database:** PostgreSQL
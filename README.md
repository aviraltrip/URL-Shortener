# URL Shortener

A lightweight, production-oriented URL shortener built with Go, Fiber, PostgreSQL, and Redis. The service keeps the canonical link metadata in PostgreSQL and uses Redis as a fast read-through cache plus a rate-limiting backend.

## Tech Stack

- Go 1.x
- Fiber web framework
- PostgreSQL via `pgxpool`
- SQL code generation via `sqlc`
- Redis for caching and IP-based rate limiting
- Docker Compose for local orchestration
- Database migrations for schema versioning

## System Overview

```mermaid
flowchart LR
    Client[Client / Browser / API Consumer]
    API[Go Fiber API]
    SVC[LinkService]
    CACHE[(Redis Cache DB 0)]
    RL[(Redis Rate Limit DB 1)]
    DB[(PostgreSQL)]

    Client -->|POST /api/v1| API
    Client -->|GET /:short_code| API

    API --> SVC
    SVC -->|cache hit / miss| CACHE
    SVC -->|persist link metadata| DB
    API -->|IP quota enforcement| RL

    DB -->|canonical data source| SVC
    CACHE -->|fast redirect lookup| API
```

## Request Flow

### 1. Create short URL

```mermaid
sequenceDiagram
    participant C as Client
    participant A as Fiber API
    participant S as LinkService
    participant D as PostgreSQL
    participant R as Redis

    C->>A: POST /api/v1 { url, short?, expiry? }
    A->>S: Shorten(rawURL, customShort, expiry)
    S->>S: Validate URL + alias
    S->>D: INSERT link metadata
    D-->>S: Link record
    S->>R: Set cache entry for short code
    S-->>A: Short URL response
    A-->>C: 200 OK
```

### 2. Resolve short URL

```mermaid
sequenceDiagram
    participant C as Client
    participant A as Fiber API
    participant S as LinkService
    participant R as Redis
    participant D as PostgreSQL

    C->>A: GET /:short_code
    A->>S: Resolve(code)
    S->>R: Check cache
    alt cache hit
        R-->>S: original URL
        S-->>A: redirect target
        S->>D: Increment click count asynchronously
    else cache miss
        S->>D: Query link by short code
        D-->>S: original URL
        S->>R: Write cache entry
        S-->>A: redirect target
        S->>D: Increment click count asynchronously
    end
    A-->>C: 307 Temporary Redirect
```

## Core Responsibilities

- URL validation and canonicalization
- custom alias enforcement
- expiration handling
- cache-backed redirect resolution
- click tracking and link metadata persistence
- Redis-backed rate limiting per client IP
- DB health and readiness checks

## Local Setup

### Docker Compose

```bash
docker compose up -d --build
```

This starts:

- API: `http://localhost:3000`
- Redis: `localhost:6379`
- PostgreSQL: `localhost:5432`

### Direct Go Run

Set environment variables in `api/.env`:

```env
DATABASE_URL=postgresql://user:password@host:5432/urlshortener
REDIS_ADDR=localhost:6379
APP_PORT=:3000
DOMAIN=http://localhost:3000
API_QUOTA=10
```

Then run:

```bash
cd api
go run main.go
```

## API

### Health and Readiness

- `GET /health` — liveness endpoint
- `GET /ready` — checks PostgreSQL and Redis connectivity

### Shorten URL

```http
POST /api/v1
Content-Type: application/json
```

```json
{
  "url": "https://github.com/aviraltrip/urlshortener",
  "short": "repo",
  "expiry": 48
}
```

### Redirect

```http
GET /:short_code
```

Returns a `307 Temporary Redirect` to the original destination and records a click event.

### Link Management

- `GET /api/v1/links?page=1&limit=20`
- `GET /api/v1/links/:code/stats`
- `DELETE /api/v1/links/:code`

## Persistence Model

The application stores canonical link metadata in PostgreSQL, including:

- short code
- original URL
- expiration timestamp
- click count
- creation timestamp

Redis is used as a fast cache for redirect lookups and as a separate database for rate-limit counters.

## Operational Notes

- Redis cache entries are TTL-based and expire with the link lifetime.
- Rate limiting is enforced per IP address using a Redis key.
- PostgreSQL migrations are applied automatically on startup.
- The service uses graceful shutdown handling for clean termination.
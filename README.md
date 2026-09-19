# URL Shortener

A high-performance, resilient URL shortening service built with **Go (Fiber)**, **Neon PostgreSQL** (pgxpool + sqlc) as the durable source of truth, and **Redis** for sub-millisecond caching and IP rate limiting.

## Architecture

```text
Client
      │
      ▼
Go Fiber API
  ├──> Neon PostgreSQL (Durable links, metadata, click analytics)
  └──> Redis (Fast cache DB 0, Rate limiting DB 1)
```

## Running Locally

### 1. With Docker Compose

```bash
docker compose up -d --build
```

Runs the API at `http://localhost:3000`, Redis at `localhost:6379`, and local PostgreSQL at `localhost:5432`.

### 2. With Neon PostgreSQL (Cloud)

Set `DATABASE_URL` in `api/.env`:

```env
DATABASE_URL=postgresql://user:password@ep-example.us-east-2.aws.neon.tech/urlshortener?sslmode=require
REDIS_ADDR=localhost:6379
APP_PORT=:3000
DOMAIN=http://localhost:3000
API_QUOTA=10
```

Run directly:

```bash
cd api
go run main.go
```

---

## API Endpoints

### Health & Readiness
- `GET /health` - Service liveness probe.
- `GET /ready` - Verifies live PostgreSQL & Redis connectivity.

### Shorten URL
- `POST /api/v1`
```json
{
  "url": "https://github.com/aviraltrip/urlshortener",
  "short": "repo",
  "expiry": 48
}
```

### Redirect / Resolve
- `GET /:short_code` - Returns `307 Temporary Redirect` to original URL and updates click counts.

### Link Management & Analytics
- `GET /api/v1/links` - List all active links with pagination (`?page=1&limit=20`).
- `GET /api/v1/links/:code/stats` - Get click statistics and expiration details.
- `DELETE /api/v1/links/:code` - Delete link from database and evict from cache.
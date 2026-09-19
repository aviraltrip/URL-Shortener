# URL Shortener

A small URL-shortening API built with Go, Fiber, Redis, and Docker.

## Run locally

```bash
docker compose up -d --build
```

The API runs at `http://localhost:3000` and Redis runs on port `6379`.

Check the containers:

```bash
docker compose ps
```

## API

### Create a short URL

`POST http://localhost:3000/api/v1`

Header:

```text
Content-Type: application/json
```

Body:

```json
{
	"url": "https://example.com",
	"short": "example",
	"expiry": 24
}
```

`short` is optional. If omitted, the API generates a short code. `expiry` is in hours and defaults to `24`.

Example response:

```json
{
	"url": "https://example.com",
	"short": "localhost:3000/example",
	"expiry": 24,
	"rate_limit": 9
}
```

### Resolve a short URL

Open or request:

```text
GET http://localhost:3000/example
```

The API returns a `301` redirect to the original URL. In Postman, disable **Automatically follow redirects** to see the destination in the `Location` response header.

## Stop the services

```bash
docker compose down
```

Redis data is persisted through the `.data` directory mounted by Docker Compose.
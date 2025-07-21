# Currency Rate Backend Service

## Overview

This project is a Go backend service for fetching and serving currency exchange rates. It is designed with modularity, scalability, and extensibility in mind, using idiomatic Go patterns and best practices. The service supports caching (in-memory or Redis), background synchronization, and integration with external exchange rate providers.

## Features

- Fetches and serves currency exchange rates via a REST API
- Supports both in-memory and Redis caching for fast lookups
- Periodic background sync with external providers
- Modular architecture: clear separation of API, service, cache, provider, and database layers
- Logging middleware for request tracing
- Configurable via environment variables
- Auto-migrates database schema for rates

## Architecture

```
[Client] ⇄ [Gin HTTP API] ⇄ [Service Layer] ⇄ [Cache Layer] ⇄ [Database Layer]
                                         ⇄ [Provider Layer (External API)]
```

- **API Layer**: Handles HTTP requests and responses (see `api/`)
- **Service Layer**: Business logic for rate fetching, caching, and cross-rate calculation (see `service/`)
- **Cache Layer**: In-memory or Redis-based caching (see `cache/`)
- **Provider Layer**: Integrates with external exchange rate APIs (see `provider/`)
- **Database Layer**: Persists rates using PostgreSQL via GORM (see `db/`)

## Getting Started

### Prerequisites
- Go 1.24+
- PostgreSQL database
- (Optional) Redis server for distributed caching

### Installation
1. Clone the repository:
   ```sh
   git clone <repo-url>
   cd assignment1
   ```
2. Install dependencies:
   ```sh
   go mod download
   ```

### Configuration

Set the following environment variables (or create a `.env` file):

```
PORT=8080
DATABASE_URL=postgres://user:password@localhost:5432/dbname?sslmode=disable
OPENEXCHANGE_URL=https://openexchangerates.org/api/latest.json?app_id=
OPENEXCHANGE_APP_ID=your_app_id
CACHE_EXPIRY_SECONDS=300
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
BACKGROUND_TASK_TIMER=10
GLOBAL_BASE_CURRENCY=USD
```

- `PORT`: Port for the HTTP server
- `DATABASE_URL`: PostgreSQL DSN
- `OPENEXCHANGE_URL`/`OPENEXCHANGE_APP_ID`: External provider config
- `CACHE_EXPIRY_SECONDS`: Cache TTL in seconds
- `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB`: Redis config (if used)
- `BACKGROUND_TASK_TIMER`: Minutes between background syncs
- `GLOBAL_BASE_CURRENCY`: Usually `USD`

### Running the Application Locally

```sh
APP_ENV=development go run cmd/main.go //for dev
APP_ENV=beta go run cmd/main.go //for beta
APP_ENV=production go run cmd/main.go //for production
```

The service will start, auto-migrate the database, and begin serving requests.

---

## Docker Usage

You can run the entire application stack (Go app, PostgreSQL, Redis) using Docker and Docker Compose.

### Build and Run with Docker

1. **Build the Docker image:**
   ```sh
   docker build -t assignment1-go-app .
   ```
2. **Run the container:**
   ```sh
   docker run --env-file .env -p 8080:8080 assignment1-go-app
   ```

### Using Docker Compose (Recommended for Local Dev)

1. **Start all services (app, db, redis):**
   ```sh
   docker-compose up --build
   ```
   This will build the Go app, start PostgreSQL and Redis, and link them together.

2. **Stop all services:**
   ```sh
   docker-compose down
   ```

- The app will be available at `http://localhost:8080`.
- Database and Redis data are persisted in Docker volumes (`pgdata`, `redisdata`).

### Production Deployment

- Use `docker-compose.prod.yml` for production settings. Update the image name and environment variables as needed.
- Example:
  ```sh
  docker-compose -f docker-compose.prod.yml up -d
  ```

### .dockerignore

The `.dockerignore` file excludes unnecessary files (binaries, git, editor configs, etc.) from the Docker build context for faster and smaller builds.

## API Usage

### Get Exchange Rate

**Endpoint:** `GET /rate?base={BASE}&target={TARGET}`

**Query Parameters:**
- `base`: The base currency code (e.g., `USD`)
- `target`: The target currency code (e.g., `EUR`)

**Response:**
```json
{
  "success": true,
  "message": "Rate fetched successfully",
  "data": {
    "Base": "USD",
    "Target": "EUR",
    "Rate": 0.92,
    "UpdatedAt": 1718000000
  }
}
```

**Error Response:**
```json
{
  "success": false,
  "error": "Missing base or target parameter"
}
```

## Project Structure

```
assignment1/
  api/         # HTTP handlers and response formatting
  cache/       # Caching logic (memory, redis)
  cmd/         # Application entry point (main.go)
  config/      # Configuration loading/structs
  db/          # Database connection logic
  middleware/  # HTTP middleware (e.g., logging)
  models/      # Data models and DTOs
  provider/    # External service integrations
  service/     # Business logic/services
  setup/       # App setup/initialization
  utility/     # Utility/helper functions
```

## Models

### Rate
| Field     | Type    | Description                |
|-----------|---------|----------------------------|
| ID        | uint    | Primary key                |
| Base      | string  | Base currency code         |
| Target    | string  | Target currency code       |
| Rate      | float64 | Exchange rate              |
| UpdatedAt | int64   | Last update (epoch time)   |

### RateDto (API Response)
| Field     | Type    | Description                |
|-----------|---------|----------------------------|
| Base      | string  | Base currency code         |
| Target    | string  | Target currency code       |
| Rate      | float64 | Exchange rate (rounded)    |
| UpdatedAt | int64   | Last update (epoch time)   |

## Caching
- **In-memory**: Fast, local cache (see `cache/memory_cache.go`)
- **Redis**: Distributed cache for multi-instance deployments (default in code, see `cache/redis_cache.go`)

## Provider Integration
- Default: [Open Exchange Rates](https://openexchangerates.org/)
- Easily extendable via the `RateProvider` interface

## Adapter Pattern in Provider Layer

The provider layer uses the **Adapter Pattern** to allow integration with multiple external exchange rate APIs in a consistent way. This is achieved via the `ProviderAdapter` interface:

This design allows you to add new providers by implementing the `ProviderAdapter` interface, without changing the rest of the codebase.

## Logging
- All requests and important service actions are logged to stdout
- See `middleware/logger.go` and service logs

## Background Sync
- Periodically fetches and updates rates from the provider to DB and cache
- Interval controlled by `BACKGROUND_TASK_TIMER` env variable

## gRPC API

This service also exposes a gRPC API for fetching currency exchange rates, suitable for high-performance or non-HTTP clients.

### Proto Definition

The gRPC service is defined in [`proto/rate.proto`](proto/rate.proto):

```proto
syntax = "proto3";

package rate;

option go_package = "grpc/proto;ratepb";

service RateService {
  rpc GetRate (GetRateRequest) returns (GetRateResponse) {}
}

message GetRateRequest {
  string base = 1;
  string target = 2;
}

message GetRateResponse {
  string base = 1;
  string target = 2;
  double rate = 3;
  int64 updated_at = 4;
  string error = 5;
}
```

### Generating Go Code from Proto

To generate/update the Go code for the gRPC service, run:

```sh
protoc --go_out=. --go-grpc_out=. proto/rate.proto
```

This will update the generated files in `grpc/proto/`.

- Requires [`protoc`](https://grpc.io/docs/protoc-installation/) and the Go plugins:
  ```sh
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
  ```
  Ensure `$GOPATH/bin` is in your `$PATH`.

### Running the gRPC Server

The gRPC server is started automatically alongside the HTTP server. It listens on the port specified by the `GRPC_PORT` environment variable (default: `50051`).

Set in your `.env` or environment:
```
GRPC_PORT=50051
```

### gRPC Endpoint

- **Service:** `RateService`
- **Method:** `GetRate`
- **Request:**
  - `base` (string): Base currency code (e.g., `USD`)
  - `target` (string): Target currency code (e.g., `EUR`)
- **Response:**
  - `base` (string)
  - `target` (string)
  - `rate` (double)
  - `updated_at` (int64)
  - `error` (string, optional)

### Example: Calling with grpcurl

You can test the gRPC API using [`grpcurl`](https://github.com/fullstorydev/grpcurl`):

```sh
grpcurl -plaintext -d '{"base":"USD","target":"EUR"}' localhost:50051 rate.RateService/GetRate
```

### Notes
- The gRPC server runs concurrently with the HTTP server.
- See `api/rate_grpc_server.go` and `setup/setup.go` for implementation details.
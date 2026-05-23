# Go Hiring Challenge

This repository contains a Go application for managing products and their prices, including functionalities for CRUD operations and seeding the database with initial data.

## Project Structure

1. **cmd/**: Contains the main application and seed command entry points.
   - `server/main.go`: The main application entry point, serves the REST API.
   - `seed/main.go`: Command to seed the database with initial product data.
2. **app/**: Contains HTTP handlers and shared API utilities.
   - `catalog/`: Catalog list and product detail endpoints.
   - `categories/`: Category list and create endpoints.
   - `health/`: Liveness and readiness probes.
   - `middleware/`: Recovery and rate limiting.
   - `config/`: Environment variable loading and validation.
   - `api/`: Shared JSON response helpers.
   - `database/`: PostgreSQL connection setup with pool tuning and retries.
3. **sql/**: Database migration scripts (executed in lexical order by the seed command).
4. **models/**: GORM models, repositories, and store interfaces.

## Setup Code Repository

1. Create a github/bitbucket/gitlab repository and push all this code as-is.
2. Create a new branch, and provide a pull-request against the main branch with your changes. Instructions to follow.

## Application Setup

- Ensure you have Go installed on your machine.
- Ensure you have Docker installed on your machine.
- Copy `.env.example` to `.env` and adjust values if needed.
- Important makefile targets:
  - `make tidy`: will install all dependencies.
  - `make docker-up`: will start the required infrastructure services via docker containers.
  - `make seed`: ⚠️ Will destroy and re-create the database tables.
  - `make test`: Will run the unit tests.
  - `make integration-test`: Will run integration tests against Postgres (requires DB).
  - `make run`: Will start the application.
  - `make docker-down`: Will stop the docker containers.

Follow up for the assignment here: [ASSIGNMENT.md](ASSIGNMENT.md)

API contract: [openapi.yaml](openapi.yaml)

## API Endpoints

All responses are JSON. Errors use `{"error":"<message>"}`.

### `GET /health/live`

Liveness probe for process health checks. Always returns `200` when the server is running.

### `GET /health/ready`

Readiness probe. Returns `200` when the database is reachable, `503` otherwise.

### `GET /catalog`

Returns a paginated list of products.

| Query param        | Default | Constraints | Description                          |
|--------------------|---------|-------------|--------------------------------------|
| `offset`           | `0`     | `>= 0`      | Number of records to skip            |
| `limit`            | `10`    | `1–100`     | Maximum records to return            |
| `category`         | —       | —           | Filter by category code (404 if unknown) |
| `price_less_than`  | —       | `>= 0`      | Filter products below this price         |

**Example response:**

```json
{
  "products": [
    {
      "code": "PROD001",
      "price": 10.99,
      "category": {
        "code": "clothing",
        "name": "Clothing"
      }
    }
  ],
  "total": 8,
  "offset": 0,
  "limit": 10
}
```

### `GET /catalog/{code}`

Returns product details including category and variants. Variant prices inherit the product base price when not set in the database.

**Example response:**

```json
{
  "code": "PROD001",
  "price": 10.99,
  "category": {
    "code": "clothing",
    "name": "Clothing"
  },
  "variants": [
    {
      "name": "Variant A",
      "sku": "SKU001A",
      "price": 11.99
    },
    {
      "name": "Variant B",
      "sku": "SKU001B",
      "price": 10.99
    }
  ]
}
```

### `GET /categories`

Returns all categories.

```json
{
  "categories": [
    {
      "code": "clothing",
      "name": "Clothing"
    }
  ]
}
```

### `POST /categories`

Creates a new category.

**Request body:**

```json
{
  "code": "bags",
  "name": "Bags"
}
```

Returns `201 Created` with the created category. Returns `409 Conflict` when the category code already exists.

## Seed Data

Categories and product assignments after seeding:

| Category     | Code           | Products                          |
|--------------|----------------|-----------------------------------|
| Clothing     | `clothing`     | PROD001, PROD004, PROD007         |
| Shoes        | `shoes`        | PROD002, PROD006                  |
| Accessories  | `accessories`  | PROD003, PROD005, PROD008         |

## Architecture Notes

- Handlers depend on `models.ProductStore` and `models.CategoryStore` interfaces rather than concrete repositories, keeping HTTP layers testable and decoupled from persistence.
- Repositories accept `context.Context` for cancellation and timeout propagation.
- Monetary values use `shopspring/decimal` internally and are exposed as JSON floats for API compatibility with the starter project.
- Required environment variables are validated at startup; missing values fail fast with a clear error.
- Database connections use a configurable pool and retry on startup until Postgres is ready.
- Global middleware provides panic recovery and per-IP rate limiting (configurable via `RATE_LIMIT_RPS`, `0` disables it).

## Testing

| Command | Description |
|---------|-------------|
| `make test` | Unit tests only (writes `coverage.unit.out`) |
| `make integration-test` | Integration tests against Postgres (writes `coverage.integration.out`) |
| `make coverage` | Unit tests + prints coverage percentage per package |
| `make test-all` | Unit + integration tests, merges coverage, prints combined percentage |

**Requirements:** Integration tests and `make test-all` require Postgres running (`make docker-up`) and a configured `.env`.

**View coverage locally:**

```bash
make coverage          # unit test coverage summary
make test-all          # combined unit + integration summary

# HTML report from merged coverage
make test-all
go tool cover -html=coverage.out -o coverage.html
```

CI runs `go vet`, `staticcheck`, unit tests, integration tests, and a combined coverage report on every branch push/PR. Coverage percentages are printed in the GitHub Actions logs.

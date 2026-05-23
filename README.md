# Go Hiring Challenge

This repository contains a Go application for managing products and their prices, including functionalities for CRUD operations and seeding the database with initial data.

## Project Structure

1. **cmd/**: Contains the main application and seed command entry points.
   - `server/main.go`: The main application entry point, serves the REST API.
   - `seed/main.go`: Command to seed the database with initial product data.
2. **app/**: Contains HTTP handlers and shared API utilities.
   - `catalog/`: Catalog list and product detail endpoints.
   - `categories/`: Category list and create endpoints.
   - `api/`: Shared JSON response helpers.
   - `database/`: PostgreSQL connection setup.
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
  - `make test`: Will run the tests.
  - `make run`: Will start the application.
  - `make docker-down`: Will stop the docker containers.

Follow up for the assignment here: [ASSIGNMENT.md](ASSIGNMENT.md)

## API Endpoints

All responses are JSON. Errors use `{"error":"<message>"}`.

### `GET /catalog`

Returns a paginated list of products.

| Query param        | Default | Constraints | Description                          |
|--------------------|---------|-------------|--------------------------------------|
| `offset`           | `0`     | `>= 0`      | Number of records to skip            |
| `limit`            | `10`    | `1–100`     | Maximum records to return            |
| `category`         | —       | —           | Filter by category code              |
| `price_less_than`  | —       | decimal     | Filter products below this price     |

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
  "total": 8
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

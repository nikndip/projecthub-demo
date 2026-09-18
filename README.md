# ProjectHub Demo

A compact public demonstration of a project and finance workflow built with Go and PostgreSQL.

> [!IMPORTANT]
> This repository is a clean-room demo created for technical review. It is not the source code of the production CRM, does not contain client data, and does not reproduce confidential business rules. The production system is private because it was developed for a specific client.

## What this demo shows

- modular Go application with HTTP, application and persistence boundaries;
- PostgreSQL schema constraints and transactional writes;
- money represented as integer kopecks, never `float`;
- concurrency-safe budget reservation with `SELECT ... FOR UPDATE`;
- strict JSON decoding and domain-oriented HTTP errors;
- graceful shutdown, structured logging and health checks;
- unit and HTTP tests, race detection and CI;
- reproducible local environment with Docker Compose.

## Business scenario

A project has an approved budget. Users can create payment requests against that budget. Creating a request reserves money immediately. Concurrent requests cannot reserve more than the available amount because the project row is locked and checked inside one database transaction.

```text
HTTP request
    |
    v
Handler -> Application service -> Repository -> PostgreSQL
                                      |
                                      +-> lock project
                                      +-> validate available budget
                                      +-> create payment request
                                      +-> update reserved amount
                                      +-> commit
```

## Quick start

Requirements: Docker with Docker Compose.

```bash
docker compose up --build
```

The API will be available at `http://localhost:8080`.

If port `8080` is already in use, choose another host port:

```bash
HOST_PORT=18082 docker compose up --build
```

Create a project:

```bash
curl -i http://localhost:8080/api/v1/projects \
  -H 'Content-Type: application/json' \
  -d '{
    "name": "Product launch",
    "client": "Acme",
    "budget_kopecks": 15000000
  }'
```

Create a payment request:

```bash
curl -i http://localhost:8080/api/v1/projects/1/payment-requests \
  -H 'Content-Type: application/json' \
  -d '{
    "purpose": "Production services",
    "amount_kopecks": 2500000
  }'
```

List projects and their reserved amounts:

```bash
curl http://localhost:8080/api/v1/projects
```

## API

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/healthz` | database readiness check |
| `GET` | `/api/v1/projects` | list projects |
| `POST` | `/api/v1/projects` | create a project |
| `POST` | `/api/v1/projects/{id}/payment-requests` | reserve budget and create a payment request |

All monetary values use integer kopecks. For example, `2500000` means `25,000.00 RUB`.

## Run checks

```bash
go test -race ./...
go vet ./...
govulncheck ./...
```

## Repository layout

```text
cmd/api/                    application entry point
internal/app/               use cases and validation
internal/domain/            domain models and errors
internal/httpapi/           HTTP transport
internal/store/postgres/    PostgreSQL repository and migration
.github/workflows/          continuous integration
```

## Scope

The demo intentionally focuses on one complete workflow instead of recreating the full private product. Authentication, the browser UI, tenders, estimates, invoices, documents and client-specific processes are outside this public example.

## License

MIT. No client source code or confidential materials are included.

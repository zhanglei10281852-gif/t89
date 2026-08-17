# Wedding Management System

This repository contains a Vue frontend and a Go API for managing wedding customers, quotes, contracts, staff schedules, and operational statistics.

## Backend Architecture

The Go module lives in `backend/` and uses Gin, GORM, and SQLite.

- `routers/` owns HTTP request parsing and response/status mapping.
- `contracting/` owns contract-signing rules and coordinates the complete booking operation.
- `repositories/` implements transaction-scoped persistence with GORM.
- `models/` contains the persisted data model.
- `utils/` contains database, JWT, and password infrastructure.

Contract creation keeps the existing `POST /api/contracts` request and response format. A successful request is committed as one operation: the quote and participants are validated, the planner's date is reserved, the contract and schedule are created, and the customer moves to `preparing`. A confirmed quote can be contracted only once. Contract updates and deletions use the same transaction boundary, so related customer and schedule changes are also atomic.

The contract service accepts a repository and a clock. This keeps business behavior deterministic in tests and allows HTTP, service, and persistence behavior to be verified independently. SQLite runs with foreign keys, WAL, a write wait timeout, and immediate transactions so concurrent signing resolves to a stable business conflict rather than a partial write.

## Backend Verification

Run commands from the Go module directory:

```bash
cd backend
go test ./...
go test -race ./...
go build ./...
go vet ./...
```

The contract tests use an isolated temporary SQLite database. They cover successful booking, quote reuse, concurrent signing, planner conflicts, transactional rollback, request cancellation, cent-precision payment rules, update/delete behavior, and HTTP error mapping.

## Frontend Verification

```bash
cd frontend
npm ci
npm run build
```

## Local Run

The complete application can be started from the repository root:

```bash
docker compose up --build
```

The API listens on port `8627`; the frontend listens on port `5173`.
The SQLite database is created under `data/` on first start and populated
with the built-in development users, service items, lucky days, and packages.


# Lujay Auto — Backend Assessment Submission

Welcome! This README is written for the Lujay Auto engineering assessment. It explains the project purpose, how to run and test it locally (Docker-first), and — most importantly — the design decisions I made to demonstrate code quality, backend architecture understanding, and database/API efficiency.

This README is intentionally detailed, practical, and a little fun — think of it as the engineer's storybook: technical, honest, and easy to read.

## 🎯 Assessment Goals (What this repo demonstrates)

1. Code quality, readability, and maintainability
	- Clear project layout and small, focused packages
	- Idiomatic Go style and consistent patterns
	- Centralized error handling and input validation
	- Tests for core logic where safe and fast

2. Understanding of backend & architectural concepts
	- Layered architecture: handlers → services → storage
	- JWT authentication with RBAC (roles: admin, dealer, buyer)
	- External services integration (Cloudinary for file storage)
	- Caching (Redis) for rate limiting and query caching
	- Containerized for reproducible local testing (Docker)

3. Efficiency of database design and API performance
	- MongoDB models tuned with BSON tags and indexes
	- Pagination and selective projections for large lists
	- Cache layer to reduce DB load for frequently requested resources
	- Background-friendly, idempotent operations for uploads/deletes

## 🚀 Quick start (Docker — recommended)

Prereqs: Docker Desktop running.

1. From repository root:

```powershell
# build and start everything (app, mongodb, redis)
docker-compose up --build -d

# check containers
docker-compose ps

# view logs for app
docker-compose logs -f app
```

2. Open the API: http://localhost:8080
3. Health: http://localhost:8080/health
4. Use the included Postman collection (`Lujay_API_Collection.postman_collection.json`) for an end-to-end flow (register, create vehicle, upload images).


## 🧩 Project structure (high level)

- `cmd/server` — application entry
- `internal/handlers` — HTTP handlers (Gin)
- `internal/service` — business logic and domain operations
- `internal/storage` — MongoDB client and collections
- `internal/cache` — Redis wrapper and rate-limit helpers
- `internal/upload` — Cloudinary uploader abstraction
- `internal/middleware` — auth, validation, caching middlewares
- `internal/models` — request/response and DB models
- `docker-compose.yml` & `Dockerfile` — containerization
- `Lujay_API_Collection.postman_collection.json` — Postman collection

This separation keeps code testable and each layer focused on one responsibility.

## 🧠 Design highlights & rationale

Layered architecture
- Handlers do minimal HTTP-level work: parse request, call service, send response.
- Services implement domain logic and orchestrate storage + upload + cache interactions.
- Storage packages isolate MongoDB usage so models and services don't leak DB-specific code.

Why this matters: If we switch DBs or add a queue later, the changes are localized.

Authentication & Authorization
- JWT-based authentication.
- Middleware extracts user identity and does RBAC checks (roles: admin, dealer, buyer).

File uploads
- Cloudinary used as the external object store for images. The upload module validates file type/size and supports safe rollback if DB updates fail.

Caching & Rate limiting
- Redis is used for short-term caching (e.g., GET /vehicles cache) and rate-limiting counters. The middleware gracefully degrades if Redis is missing (useful for local dev without Redis).

Database efficiency
- Documents focused per collection (e.g., vehicles, users, inspections, transactions).
- Important fields have BSON tags and indexes are defined in initialization code (look for index creation in `internal/storage`).
- List endpoints use pagination and projection to avoid returning large documents.

Scalability considerations
- Stateless app process (container-friendly) so horizontal scaling behind a load balancer is straightforward.
- External services (MongoDB, Redis, Cloudinary) can be scaled independently.

## ✅ API overview (high-level)

- Auth: `/api/v1/auth/register`, `/api/v1/auth/login`, `/api/v1/auth/profile`
- Vehicles: CRUD on `/api/v1/vehicles` + `/api/v1/vehicles/:id/images` (upload/delete/set-primary)
- Inspections: `/api/v1/inspections` and `/api/v1/vehicles/:id/inspections`
- Transactions: `/api/v1/transactions` and `/api/v1/vehicles/:id/transactions`

Use the Postman collection for ready-to-run requests. Authentication requests automatically save tokens into collection variables.

## 📐 Database design (summary)

Collections:
- `users`: users with roles and secure password hashes
- `vehicles`: ownerId, metadata, array of `VehicleImage` objects (url/publicId/isPrimary)
- `inspections`: linked to `vehicleId`
- `transactions`: linked to `vehicleId` and user

Indexes (selected):
- `users.email` — unique
- `vehicles.ownerId` — to quickly find a user's vehicles
- `vehicles.vin` — optional, unique if present
- `inspections.vehicleId` & `transactions.vehicleId` — for fast lookups

Why BSON + indexes matter: queries use indexes and projection to minimize I/O and CPU on the DB.

## 🧪 Testing & verification

What I included:
- Unit tests for core model validation and business logic (see `internal/models` and `internal/service/*_test.go`) — aim to keep tests fast and deterministic.
- Integration tests are not included to avoid flakiness in CI (they require external services). Instead, containerized local tests with Docker are supported.

Quick manual test steps:

1. Start services via Docker (see Quick start)
2. Import `Lujay_API_Collection.postman_collection.json` into Postman
3. Run `Authentication -> Register User (Dealer)` then `Vehicles -> Create Vehicle` then `File Upload -> Upload Vehicle Images`

Automated tests (run locally):

```powershell
# run Go unit tests
go test ./... -run Test -v
```

Note: Running the full test suite requires Go toolchain installed.

## 🔒 Security considerations

- Passwords are hashed before storage. The API never returns passwords in responses.
- JWT secret is provided through environment variables and should be rotated in production.
- Input validation is enforced using binding tags and manual checks to avoid malformed data reaching the DB.
- File uploads are validated for type and size; Cloudinary public IDs are used to safely reference images.

## 📈 Performance & reliability notes

- Read-heavy endpoints use Redis caching to reduce DB load.
- List endpoints implement pagination and optional filtering to limit result set size.
- Background jobs and async queues are considered for heavy tasks (e.g., image processing) — not yet required for this assessment.

## 🔁 CI/CD and Deployment

- A minimal Docker / Docker Compose setup is provided for local dev and smoke testing.
- Recommended production deployment: build multi-arch images, run in Kubernetes or a container service, use managed MongoDB and Redis.
- Add an infra pipeline (GitHub Actions) to build, run lints, run tests, and push images on merge.

## 🛠️ Troubleshooting quick wins

- If you get `Invalid request payload` when testing endpoints, ensure JSON body matches model fields (use `firstName` + `lastName` for registration, and `location.city/state/country` for vehicle creation).
- Use `docker-compose logs -f app` to view live server logs.

## ✨ Final notes

Building backend systems is equal parts engineering and ergonomics: useful APIs, predictable behavior, and readable code. I designed this project to be clear to another engineer reading it five minutes from now — so they can fix, extend, or scale it without guessing.

If you like what you see, please run the Postman flow (it’s fun — try uploading three photos at once). If you want, I can also add a short video walkthrough or a CI pipeline next.

Thank you for reading — may your logs be green and your builds fast! 🚗💨


## Project Structure

```
LUJAY ASSESMENT/
├─ cmd/
│  └─ server/main.go          # Application entry point
├─ internal/
│  ├─ auth/                   # Authentication logic
│  ├─ handlers/               # HTTP handlers
│  ├─ models/                 # Data models
│  ├─ service/                # Business logic
│  ├─ storage/                # Database layer
│  └─ middleware/             # HTTP middleware
├─ pkg/
│  └─ utils/                  # Utility functions
├─ scripts/                   # Build and deployment scripts
├─ tests/                     # Integration tests
├─ Dockerfile
├─ docker-compose.yml
├─ .env.example
├─ go.mod
└─ README.md
```

## Getting Started

### Prerequisites

- Go 1.21 or higher
- Docker (optional)

### Installation

1. Clone the repository
```bash
git clone https://github.com/Over-knight/GO-BACKEND.git
cd "Lujay assesment"
```

2. Copy the environment file
```bash
cp .env.example .env
```

3. Install dependencies
```bash
go mod tidy
```

4. Run the server
```bash
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`

## Development

### Running Tests
```bash
go test ./...
```

### Building
```bash
go build -o bin/server cmd/server/main.go
```

## Docker

### Build and run with Docker
```bash
docker-compose up --build
```

## License

MIT

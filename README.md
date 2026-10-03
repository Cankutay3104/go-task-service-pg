# Production-Grade Task Service (Go & PostgreSQL)

A resilient, production-ready RESTful Task Management Service engineered with idiomatic Go (1.22+) and PostgreSQL. Built around clean architecture, dependency injection, and strict input sanitization.

---

## Architectural Highlights

- **Standard Library Routing:** Leverages Go 1.22+ method-matching ServeMux (e.g., "GET /tasks/{id}") with zero external web routing frameworks.
- **Dependency Injection:** Handlers depend strictly on abstract repository interfaces, allowing full unit-test coverage via in-memory mocks without touching network sockets.
- **Relational & JSONB Storage:** Backed by PostgreSQL (`pgx/v5` stdlib driver). Complex unstructured metadata is stored as binary `JSONB`, implementing `sql.Scanner` and `driver.Valuer`.
- **Hardened Payload Parsing:** Custom generic validator using `http.MaxBytesReader`, `DisallowUnknownFields()`, and trailing stream validation to prevent memory exhaustion (OOM) and malformed payload injection.
- **Graceful Shutdown:** Implements an asynchronous lifecycle listener trapping `os.Interrupt` and `syscall.SIGTERM` with a 10-second drain window to ensure in-flight HTTP requests complete before the database pool closes.

---

## Project Structure

```plaintext
task-service-pg/
├── cmd/
│   └── api/
│       └── main.go                 # Composition root & lifecycle coordinator
├── internal/
│   ├── database/
│   │   └── postgres.go             # Connection pool tuning & migrations
│   ├── handlers/
│   │   ├── handlers.go             # Transport controller & JSON serialization
│   │   └── handlers_test.go        # Unit tests using in-memory mock repository
│   ├── models/
│   │   └── task.go                 # Domain entity, validation, and JSONB mapping
│   ├── repository/
│   │   ├── repository.go           # Repository interface & sentinel errors
│   │   ├── postgres_task.go        # Concrete PostgreSQL CRUD implementation
│   │   └── postgres_test.go        # Transactional integration test suite
│   └── validator/
│       └── payload.go              # Generic JSON reader with stream boundary checks
├── .env.example                    # Blueprint for environment variables
├── .gitignore                      # Excludes credentials (.env) and compiled binaries
├── go.mod
└── go.sum
```

---

## Prerequisites

- **Go:** 1.22 or higher
- **PostgreSQL:** 14 or higher running locally or in Docker

---

## Environment Configuration

Copy the example environment file and configure your local PostgreSQL credentials:

```bash
cp .env.example .env
```

Ensure your `.env` contains:

```env
DATABASE_URL=postgres://<username>:<password>@localhost:5432/<database>?sslmode=disable
PORT=:8080
```

---

## Running the Application

1. **Download dependencies:**
   ```bash
   go mod tidy
   ```

2. **Start the server:**
   ```bash
   go run ./cmd/api/main.go
   ```

The application will automatically verify the connection pool, execute schema migrations (`CREATE TABLE IF NOT EXISTS`), and begin listening for requests on the specified port.

---

## Automated Test Suites

The project maintains comprehensive separation between pure unit tests and live integration tests.

### 1. Handler Unit Tests (Zero Database Required)
Tests HTTP routing, request decoding, and status code negotiation against an in-memory mock repository:

```bash
go test -v ./internal/handlers/...
```

### 2. Repository Integration Tests (Live PostgreSQL)
Executes parameterized CRUD queries and JSONB serialization against your live test database:

```bash
go test -v ./internal/repository/...
```

### 3. Run Entire Test Suite
```bash
go test -v ./...
```

---

## API Specification

### Endpoints

| Method | Endpoint | Description | Status Codes |
|---|---|---|---|
| `POST` | `/tasks` | Create a new task entity | `201 Created`, `400 Bad Request`, `409 Conflict` |
| `GET` | `/tasks/{id}` | Retrieve a task by its unique ID | `200 OK`, `404 Not Found` |
| `GET` | `/tasks` | List all tasks ordered by creation time | `200 OK`, `500 Internal Server Error` |
| `PUT` | `/tasks/{id}` | Update existing task details and metadata | `200 OK`, `400 Bad Request`, `404 Not Found` |
| `DELETE` | `/tasks/{id}` | Permanently remove a task record | `204 No Content`, `404 Not Found` |

---

### Request Payload Example (`POST /tasks`)

```json
{
  "id": "task-uuid-001",
  "title": "Implement Graceful Shutdown",
  "description": "Ensure context timeout protects in-flight operations",
  "status": "todo",
  "metadata": {
    "tags": ["backend", "go", "resilience"],
    "priority": 1
  }
}
```

---

## License

This project is open-source and available under the MIT License.
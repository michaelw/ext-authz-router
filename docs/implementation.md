# 🏗️ Ext AuthZ Routing Plugin Implementation Guide

This document describes the architecture and technical decisions.

## 📁 Project Structure

The project follows the standard Go project layout:

```
.
├── api/                   # OpenAPI spec and all generated code
├── db/
│   ├── generate.go        # go:generate directive for sqlc
│   ├── sqlc.yaml          # sqlc configuration
│   ├── schema.sql         # Predefined schema; must be compatible with sqlc
│   ├── queries/           # .sql files for sqlc query generation
│   ├── *.gen.go           # sqlc-generated data access layer
│   └── validation.go      # Entity validation based on OpenAPI constraints
├── internal/
│   └── server/
│       ├── main.go        # Entry point logic
│       └── handlers/      # One file per resource (e.g., organizations_handlers.go)
├── cmd/
│   └── ext-authz-router-service/
│       └── main.go        # Application bootstrap
├── Dockerfile
├── docker-compose.yml
├── .devcontainer/
│   └── devcontainer.json  # VS Code DevContainer setup
└── go.mod
```

## 🔧 Technical Stack

* **Web Framework:** `gin`
* **Database Access:** `pgx/v5` via `sqlc` (do not use `database/sql` or `pq`)
* **API Specification:** OpenAPI 3.0, integrated via `oapi-codegen` (with `StrictServerInterface`)
* **Database:** PostgreSQL (assume schema is available as raw SQL; migrations handled externally)

## 🧱 Architecture Layers

The service is structured in **three clean layers**:

1. **API Layer** (OpenAPI + transport layer)
   - Generated types and interfaces from OpenAPI spec
   - HTTP request/response handling
   - Input validation and serialization

2. **Business Logic Layer** (internal/server/handlers)
   - Core business logic implementation
   - Handler implementations for each resource
   - Business rule enforcement

3. **Data Access Layer** (db layer with `sqlc`-generated and hand-written code)
   - Database queries and operations
   - Data validation based on OpenAPI constraints
   - Transaction management

## 🔄 Data Flow

1. **Request** → API Layer (validation, deserialization)
2. **Processing** → Business Logic Layer (handlers)
3. **Persistence** → Data Access Layer (database operations)
4. **Response** ← API Layer (serialization, HTTP response)

## 🚦 Middleware Architecture

Implement middleware that:

* Acquires a request-scoped DB connection from a `pgxpool.Pool`
* Injects it into the `context.Context`
* Makes it available via a helper like `PgxConnFrom(ctx)`

## 📦 Dependency Injection

* Apply dependency injection for external services (DB, SDK clients, etc.)
* Use interfaces for external dependencies to enable testing
* Define **interfaces** in `types.go`; **implementations** in separate files

## 🔧 Code Generation

The project uses two main code generators:

### sqlc
- Generates Go code from SQL queries in `db/queries/`
- Uses `pgx/v5` driver
- Only generate queries actually needed for API implementation

### oapi-codegen
- Generates request/response types and handler interfaces
- Uses `StrictServerInterface` pattern
- All generated code placed under `api/`
- Actual implementations must override the default 501 handlers

## 🎯 Design Principles

* **Idiomatic Go**: Follow effectivego.dev guidelines
* **Single Responsibility**: Each handler handles one resource type
* **Interface Segregation**: Define focused interfaces
* **Dependency Inversion**: Depend on abstractions, not concretions
* **Testability**: All business logic must be unit testable

# 🚀 Ext AuthZ Routing Plugin Development Guide

This document provides comprehensive development guidelines.

## 📋 Pull Request Rules

* Use [Conventional Commit](https://www.conventionalcommits.org/en/v1.0.0/#specification) message format
  (`type` is `fix`, `feat`, `chore`, etc.)

  ```text
  <type>[optional scope in parenthesis]: <description>

  [optional body]

  [optional footer(s)]
  ```

  This helps with changelog generation.

---

## ⏱️ Development Phases

Follow these phases when implementing new features:

1. **Project Setup:** Generate the initial project structure and configuration files.
2. **OpenAPI Integration:** Generate API types and handlers from OpenAPI spec.
3. **Data Access Layer:** Implement the data access layer using `sqlc` with `pgx/v5`, but not yet the validations (this will come later).
4. **Business Logic:** Implement the core business logic in a modular way.
5. **Validation:** Implement validation rules in the data access layer.
6. **Testing:** Write comprehensive tests using Ginkgo and Gomega.
7. **Polish and Harden:** Ensure the service is production-ready with proper error handling, logging, and documentation. Ensure `go generate -x ./...` rebuilds all generated code cleanly.
8. **Dev Environment:** Set up Docker Compose and VS Code DevContainer for development.

---

## 🐳 Development Environment

### Local Development

1. Start the development environment:
   ```bash
   devspace dev
   ```

2. The service will automatically reload on file changes thanks to `air`.

3. Access the service at `http://localhost:8080` (or configured port).

---

## 🧪 Testing Strategy

### BDD Testing with Ginkgo + Gomega

* Use **Ginkgo** + **Gomega** for BDD-style tests
* All business logic must have unit tests
* Use **table-driven tests** (define struct above the test func)
* For Ginkgo tables, place the struct name on the same line as the Entry description

### Test Example

```go
Describe("FooHandler", func() {
  type testCase struct {
    input    string
    expected string
  }

  DescribeTable("does something",
    func(tc testCase) {
      ...
    },
    Entry("valid case", testCase{
      input: "bar",
      expected: "baz",
    }),
  )
})
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with Ginkgo
ginkgo -r

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## 🔧 Build and Deployment

### Building the Service

```bash
# Build all commands
go build ./cmd/...

# Generate all code
go generate -x ./...
```

### Code Generation

The project uses code generation for:
- **sqlc**: Database access layer from SQL queries
- **oapi-codegen**: API types and handlers from OpenAPI spec

Always use `go generate -x ./...` to regenerate code after changes.

---

## ⚠️ Important Constraints

* No schema migrations in the service itself
* Assume OpenAPI + schema are fixed inputs
* Code generation must be triggered via `go generate -x ./...`
* Builds must be reproducible with `go build ./cmd/...`
* No unused or stub code — everything must be connected and testable

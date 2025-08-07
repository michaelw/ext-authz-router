# 🏗️ Ext AuthZ Routing Plugin Implementation Guide

This document describes the architecture and technical decisions.

## 📁 Project Structure

The project follows the standard Go project layout:

```
.
├── api/                   # OpenAPI spec and all generated code
├── internal/
│   └── server/
│       └── handlers/      # One file per resource
├── cmd/
│   └── ext-authz-router-service/
│       └── main.go        # Application bootstrap
├── devspace.yaml
├── Dockerfile
├── docker-compose.yml
├── .devcontainer/
│   └── devcontainer.json  # VS Code DevContainer setup
└── go.mod
```

## 🔧 Technical Stack

* **Web Framework:** `gin`
* **API Specification:** OpenAPI 3.0, integrated via `oapi-codegen` (with `StrictServerInterface`)

## 🔧 Code Generation

The project uses these main code generators:

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

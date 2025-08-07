# Ext AuthZ Routing Plugin

Envoy Ext AuthZ Plugin for header-based routing to environments

## Project Overview

- `cmd/ext-authz-router-service/main.go` — main entrypoint
- `internal/server/` — server logic
- `api/openapi.yaml` — OpenAPI spec
- `db/` — Database related code
- `Dockerfile`, `docker-compose.yaml` — containerization

For a detailed overview, refer to the [Implementation Guide](docs/implementation.md).

## Running the Services

Users who want to run the service **without modifying the code** can use Docker Compose directly.

### Running with Docker Compose

1. Clone the repository.

2. Run:

   ```bash
   docker-compose up --build -d
   ```

3. This spins up all required services.

4. Service runs by default at http://localhost:3000/

**Notes**

* No source code or build dependencies are required locally.
* All services run as pre-built images.
* Configuration is driven by docker-compose.yaml and overrides.

### Running with DevSpace

If you have a running Kubernetes cluster, [DevSpace](https://devspace.sh/) is all you need:

```bash
devspace dev
```

Refer to the [DevSpace](docs/devspace.md) documentation for more details.


## Usage



## Development Quickstart Guide

This project uses Docker-based devcontainers and a multi-stage Docker build for development.

For further information, refer to the [Development Guide](docs/development.md).

- Build: `docker-compose build`
- Run: `docker-compose up --build`
- K8s: `devspace dev`
- API Spec: see `api/openapi.yaml`

### Getting Started in VSCode

1. **Open the project in VS Code.**
   Make sure you have the [Remote - Containers](https://code.visualstudio.com/docs/remote/containers) extension installed.

2. **Rebuild and open the devcontainer**
   Use the command palette:
   `Dev Containers: Rebuild and Reopen in Container`

3. **Initial setup runs automatically**:
   - `go generate ./...` runs to update any generated Go code
   - `go mod tidy` runs to clean up module dependencies

4. **Source code is mounted as a workspace volume**, allowing live editing without rebuilding the container.

5. **File watching and live reload:**
   The project uses [`air`](https://github.com/air-verse/air) for fast live rebuilds and restarts on code changes.
   - Generated files (e.g., `*.gen.go`) are excluded from triggering rebuilds to avoid infinite loops.

### Other IDEs

We can spin up the build environment and then connect to it like so:

```bash
docker compose -f docker-compose.yaml -f docker-compose.dev.override.yaml up --build -d
docker compose exec ext-authz-router-service bash
```

# TODO

* Tests
* CI builds
* Persist shell history in devcontainer, dotfiles, etc.
* Inject Github credentials

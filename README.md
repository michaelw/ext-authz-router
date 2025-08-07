# Ext AuthZ Routing Plugin

Envoy Ext-AuthZ Plugin for header-based routing to environments:

* Without `namespace` cookie or `x-namespace` header set, redirect to Namespace Selector UI (or 401 error for headless clients).
* Namespace Selector sets cookie for selected environment.
* When cookie or header is set, ext-authz plugin converts it to internal `x-backend` header that is matched by HTTPRoutes to select workload.

## Project Overview

- `cmd/ext-authz-router-service/main.go` — main entrypoint
- `internal/server/` — plugin logic (gRPC server and Web UI)

For a detailed overview, refer to the [Implementation Guide](docs/implementation.md).

## Running the Services

Users who want to run the service **without modifying the code** can use DevSpace directly.

### Running with DevSpace

If you have a local Kubernetes cluster available, [DevSpace](https://devspace.sh/) is all you need:

For a preview of what gets deployed:

```bash
devspace deploy --render --skip-build
```

If you do not yet have Gateway API CRDs, a cluster-wide gateway named `gateway`, external-dns, cert-manager, etc. installed:

```bash
devspace deploy -p with-infra
```

The command will setup a fully functionioning self-contained demo environment.

Refer to the [DevSpace](docs/devspace.md) and [devspace-starter-pack](https://github.com/michaelw/devspace-starter-pack) documentation
for more information.

## Usage

After deployments have settled (`dns-sd -q ns.dns.kube` should return an IP address eventually):

Open the demo application in a browser.  On first run, it should redirect to a namespace selector dialog:

- `devspace run open-envdemo`

To select again, either delete the cookie and reload the window, or open the namespace selector in an additional window:

- `devspace run open-namespaces`

### Using Curl

```shell
❯ curl https://envdemo.int.kube -fsSI
HTTP/2 401
www-authenticate: Custom realm="namespace-required", error="missing_namespace", error_description="Provide namespace via x-namespace header or namespace cookie"
content-length: 95
content-type: text/plain
date: Mon, 11 Aug 2025 22:26:53 GMT
server: istio-envoy

curl: (56) The requested URL returned error: 401

❯ curl https://envdemo.int.kube -fsSI -H 'x-namespace: cool-otter'
HTTP/2 200
accept-ranges: bytes
content-length: 1395
content-type: text/html; charset=utf-8
last-modified: Tue, 22 Jun 2021 05:40:33 GMT
date: Mon, 11 Aug 2025 22:27:06 GMT
x-envoy-upstream-service-time: 3
x-backend-processed: blue
server: istio-envoy
```

### Uninstall

- `devspace purge` or `devspace purge -p with-infra`

## Development Quickstart Guide

This project uses Docker-based devcontainers and a multi-stage Docker build for development.

For further information, refer to the [Development Guide](docs/development.md).

- K8s: `devspace dev`
- API Spec: see `api/openapi.yaml`

### Getting Started in VSCode

Deploy a development container and connect it to VSCode

- `devspace dev --vscode`

# TODO

* Tests
* CI builds
* Persist shell history in devcontainer, dotfiles, etc.
* Inject Github credentials

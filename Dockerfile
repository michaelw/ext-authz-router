# syntax=docker/dockerfile:1.4

# match relative path from the root of the repository
ARG WORKSPACE=/workspace

# Base stage
FROM golang:1.24-bookworm AS base

USER root

# Dev stage
FROM base AS dev
ARG WORKSPACE
ARG DEVSPACE_VERSION=v6.3.15
ARG HELM_VERSION=v3.18.4
ARG YQ_VERSION=v4.46.1

WORKDIR /tmp

RUN apt-get update && apt-get install -y \
    sudo curl git
RUN tee -a /etc/sudoers <<< 'devuser ALL=(ALL) NOPASSWD:ALL'

COPY scripts/docker-entrypoint.sh /entrypoint.sh

RUN mkdir -p /.devspace /usr/local/share/bash-completion/completions \
    && chown devuser:devuser /.devspace /usr/local/bin


RUN curl -fsSL -o yq https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_linux_amd64 \
    && install -c -m 0755 yq /usr/local/bin \
    && rm yq

RUN curl -fsSL -o devspace "https://github.com/devspace-sh/devspace/releases/download/${DEVSPACE_VERSION}/devspace-linux-amd64" \
    && install -c -m 0755 devspace /usr/local/bin \
    && rm devspace \
    && devspace completion bash > /usr/local/share/bash-completion/completions/devspace

RUN curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 \
    && chmod +x get_helm.sh \
    && ./get_helm.sh --version "${HELM_VERSION}" \
    && rm get_helm.sh \
    && helm completion bash > /usr/local/share/bash-completion/completions/helm

RUN --mount=type=cache,target=/go-cache \
    chown -v devuser:devuser /go-cache

USER devuser

ENV GOCACHE=/go-cache/build
ENV GOMODCACHE=/go-cache/mod
ENV CGO_ENABLED=0

WORKDIR ${WORKSPACE}

CMD ["sleep", "infinity"]

# Build stage
FROM base AS build
ARG WORKSPACE

RUN --mount=type=cache,target=/go-cache \
    chown -v devuser:devuser /go-cache

USER devuser

ENV GOCACHE=/go-cache/build
ENV GOMODCACHE=/go-cache/mod
ENV CGO_ENABLED=0

WORKDIR ${WORKSPACE}

COPY --chown=devuser:devuser go.mod go.sum ./
RUN --mount=type=cache,target=/go-cache \
    go mod download
COPY --chown=devuser:devuser api/ ./api/
RUN --mount=type=cache,target=/go-cache \
    go generate ./...
COPY --chown=devuser:devuser . .
RUN --mount=type=cache,target=/go-cache \
    go build -v ./cmd/...

# Run stage
FROM gcr.io/distroless/static-debian12 AS prod
ARG WORKSPACE
USER nobody

WORKDIR /app
RUN mkdir -p /app/keys && chown nobody:nogroup /app/keys
COPY --from=build ${WORKSPACE}/ext-authz-router-service /app/ext-authz-router-service

ENV GIN_MODE=release
EXPOSE 3000

ENTRYPOINT ["/app/ext-authz-router-service"]

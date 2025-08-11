# syntax=docker/dockerfile:1.4

# match relative path from the root of the repository
ARG WORKSPACE=/workspaces/ext-authz-router
ARG USER=devuser

# Base stage
FROM golang:1.24-bookworm AS base
ARG WORKSPACE
ARG USER

USER root

RUN adduser --disabled-password --gecos "" ${USER} && adduser ${USER} sudo

# Dev stage
FROM base AS dev
ARG TARGETOS
ARG TARGETARCH
ARG WORKSPACE
ARG USER
ARG DEVSPACE_VERSION=v6.3.15
ARG HELM_VERSION=v3.18.4
ARG YQ_VERSION=v4.46.1

WORKDIR /tmp

RUN apt-get update && apt-get install -y \
    sudo curl git
RUN echo "${USER} ALL=(ALL) NOPASSWD:ALL" | tee -a /etc/sudoers

COPY scripts/docker-entrypoint.sh /entrypoint.sh

RUN mkdir -p /.devspace /usr/local/share/bash-completion/completions \
    && chown ${USER}:${USER} /.devspace /usr/local/bin

RUN curl -fsSL -o yq https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_${TARGETOS}_${TARGETARCH} \
    && install -c -m 0755 yq /usr/local/bin \
    && rm yq

RUN curl -fsSL -o devspace "https://github.com/devspace-sh/devspace/releases/download/${DEVSPACE_VERSION}/devspace-${TARGETOS}-${TARGETARCH}" \
    && install -c -m 0755 devspace /usr/local/bin \
    && rm devspace \
    && devspace completion bash > /usr/local/share/bash-completion/completions/devspace

RUN curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 \
    && chmod +x get_helm.sh \
    && ./get_helm.sh --version "${HELM_VERSION}" \
    && rm get_helm.sh \
    && helm completion bash > /usr/local/share/bash-completion/completions/helm

RUN --mount=type=cache,target=/go-cache \
    chown -v ${USER}:${USER} /go-cache

USER ${USER}

ENV GOCACHE=/go-cache/build
ENV GOMODCACHE=/go-cache/mod
ENV CGO_ENABLED=0

WORKDIR ${WORKSPACE}

CMD ["sleep", "infinity"]

# Build stage
FROM base AS build
ARG WORKSPACE
ARG USER

RUN --mount=type=cache,target=/go-cache \
    chown -v ${USER}:${USER} /go-cache

USER ${USER}

ENV GOCACHE=/go-cache/build
ENV GOMODCACHE=/go-cache/mod
ENV CGO_ENABLED=0

WORKDIR ${WORKSPACE}

COPY --chown=${USER}:${USER} go.mod go.sum ./
RUN --mount=type=cache,target=/go-cache \
    go mod download
COPY --chown=${USER}:${USER} api/ ./api/
RUN --mount=type=cache,target=/go-cache \
    go generate ./...
COPY --chown=${USER}:${USER} . .
RUN --mount=type=cache,target=/go-cache \
    go build -v ./cmd/...

# Run stage
FROM gcr.io/distroless/static-debian12 AS prod
ARG WORKSPACE
USER nobody

WORKDIR /app
COPY --from=build ${WORKSPACE}/ext-authz-router-service /app/ext-authz-router-service

ENV GIN_MODE=release
EXPOSE 3000

ENTRYPOINT ["/app/ext-authz-router-service"]

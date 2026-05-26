#!/bin/bash
# DevLake local development helper
# Usage:
#   ./dev.sh build           - Build all plugins + server
#   ./dev.sh build <plugin>  - Build a specific plugin (e.g. ./dev.sh build gh-copilot)
#   ./dev.sh run             - Run the server (plugins must be built first)
#   ./dev.sh dev             - Build all + run server
#   ./dev.sh deps            - Start MySQL + Grafana via docker compose
#   ./dev.sh lint            - Run golangci-lint
#   ./dev.sh test            - Run unit tests

set -e

REPO_ROOT="$(cd "$(dirname "$0")" && pwd)"
BACKEND="$REPO_ROOT/backend"
GOPATH=$(go env GOPATH)

export PATH="$PATH:$GOPATH/bin"
export PKG_CONFIG_PATH="/usr/local/lib/pkgconfig:$PKG_CONFIG_PATH"

case "${1:-}" in
  build)
    cd "$BACKEND"
    if [ -n "$2" ]; then
      echo "Building plugin: $2"
      DEVLAKE_PLUGINS="$2" make build-plugin
    else
      echo "Building all plugins + server..."
      make build-plugin
      make build-server
    fi
    ;;
  run)
    cd "$BACKEND"
    echo "Starting DevLake server on :8080..."
    ./bin/lake
    ;;
  dev)
    cd "$BACKEND"
    make build-plugin
    make build-server
    echo "Starting DevLake server on :8080..."
    ./bin/lake
    ;;
  deps)
    echo "Starting MySQL + Grafana..."
    docker compose -f "$REPO_ROOT/docker-compose-dev.yml" up mysql grafana -d
    ;;
  lint)
    cd "$BACKEND"
    golangci-lint run ./...
    ;;
  test)
    cd "$BACKEND"
    go test ./...
    ;;
  *)
    echo "Usage: $0 {build [plugin]|run|dev|deps|lint|test}"
    exit 1
    ;;
esac

#!/usr/bin/env bash
# Idempotent environment bootstrap for simplybs.
# Runs after the repository is checked out. The dev loop is just `go run .`,
# so this only warms the Go module/build caches (which also validates the
# toolchain). go and gh are provided by the base image.
set -euo pipefail

cd "$(dirname "$0")/.."

echo "==> go version"
go version

echo "==> go mod download"
go mod download

echo "==> go build ./..."
go build ./...

echo "==> install complete"

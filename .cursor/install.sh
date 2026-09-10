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

# --- simplybs cache env vars --------------------------------------------
# Make the cache settings available to every command on the box. Two places:
#   * /etc/environment  -> picked up by PAM for all sessions (no shell
#     expansion allowed, so we write resolved values).
#   * /etc/profile.d/   -> sourced by login shells (keeps the dynamic form).
configure_cache_env() {
  local repo="mrcyjanek/simplybs_private"
  local tag="v0-sbs-${USER}-$(go env GOOS)-$(go env GOARCH)"

  echo "==> configuring SIMPLYBS_CACHE_* (tag=${tag}, repo=${repo})"

  # /etc/environment: replace any prior entries, then append resolved values.
  sudo touch /etc/environment
  sudo sed -i '/^SIMPLYBS_CACHE_TAG=/d;/^SIMPLYBS_CACHE_REPO=/d' /etc/environment
  printf 'SIMPLYBS_CACHE_TAG=%s\nSIMPLYBS_CACHE_REPO=%s\n' "$tag" "$repo" \
    | sudo tee -a /etc/environment >/dev/null

  # /etc/profile.d: dynamic exports for login shells (safe if go is missing).
  sudo tee /etc/profile.d/simplybs-cache.sh >/dev/null <<'EOF'
export SIMPLYBS_CACHE_REPO=mrcyjanek/simplybs_private
if command -v go >/dev/null 2>&1; then
  export SIMPLYBS_CACHE_TAG="v0-sbs-${USER}-$(go env GOOS)-$(go env GOARCH)"
else
  export SIMPLYBS_CACHE_TAG="v0-sbs-${USER}-linux-amd64"
fi
EOF
}

configure_cache_env

echo "==> install complete"

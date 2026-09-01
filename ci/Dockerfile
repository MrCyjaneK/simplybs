# CI toolchain for package-worker jobs (Go + git + gh).
# Package artifacts stay on the GitHub Release cache, not in this image.
#
# Seed/native binaries (e.g. clang21 ./bin/mkdir) need glibc >= 2.38.
# bookworm is Debian 12 (glibc 2.36) and fails with GLIBC_2.38 not found.
# trixie is Debian 13 (glibc 2.41).
FROM golang:1.25-trixie

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      bash \
      ca-certificates \
      curl \
      git \
      gnupg \
 && curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg \
      -o /usr/share/keyrings/githubcli-archive-keyring.gpg \
 && chmod go+r /usr/share/keyrings/githubcli-archive-keyring.gpg \
 && echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
      > /etc/apt/sources.list.d/github-cli.list \
 && apt-get update \
 && apt-get install -y --no-install-recommends gh \
 && rm -rf /var/lib/apt/lists/*

# GitHub-hosted ubuntu-latest uses uid 1001 (runner). Match it so the
# /__w workspace bind-mount is writable without running as root.
# crosstool-NG (native/ct-tools) refuses uid 0.
RUN groupadd --gid 1001 ci \
 && useradd --uid 1001 --gid 1001 --create-home --shell /bin/bash ci

ENV GOCACHE=/tmp/go-cache \
    GOMODCACHE=/tmp/go-modcache \
    GOTMPDIR=/tmp/go-tmp

USER 1001
RUN mkdir -p /tmp/go-cache /tmp/go-modcache /tmp/go-tmp
WORKDIR /workspace

# AGENTS.md

`simplybs` is a cross-compilation build system written in Go. Package definitions
live in `packages/*.json` (and `packages/native/*.json`); patches live in `patches/`.

## Dev loop

The tool is run directly from source — there is no separate build step for
day-to-day work:

```bash
go run . <flags>
```

Common invocations:

```bash
go run . -host x86_64-linux-gnu -package zlib -list      # show a package + its dep tree
go run . -host x86_64-linux-gnu -package zlib -download   # fetch sources (verifies sha256)
go run . -host x86_64-linux-gnu -package zlib -build      # build a package (+ its deps)
go run . -world -host <triplet> -build                    # build everything for a host
```

Supported host triplets are defined in `host/main.go` (e.g. `x86_64-linux-gnu`,
`aarch64-linux-android`, `x86_64-w64-mingw32`, `aarch64-apple-darwin`).

## After editing packages — always run lint

Whenever you add or change a package definition (new source, new git ref, bumped
version, new dependency, etc.), regenerate metadata and validate with:

```bash
go run . -lint
```

`-lint` reformats every `packages/**.json`, checks for invalid/cyclic
dependencies, and regenerates the checked-in `sources.json`. Commit the resulting
`sources.json` changes together with your package edits.

To generate git `download` entries from repos you've checked out locally:

```bash
go run ./cmd/gengitdeps        # prints download JSON to stdout
```

## Tests

```bash
go test ./...
```

Note: `host.TestHosts` is data-driven from `packages/native/_.json` and may
already fail on a clean checkout — unrelated to environment setup.

## Tooling

`go` and `gh` are provided by the base image; there is nothing else to install.

## Cache environment variables

`install.sh` exports the shared cache settings box-wide (via `/etc/environment`
and `/etc/profile.d/simplybs-cache.sh`), so every command sees them:

- `SIMPLYBS_CACHE_TAG=v0-sbs-<user>-<goos>-<goarch>` (e.g. `v0-sbs-ubuntu-linux-amd64`)
- `SIMPLYBS_CACHE_REPO=mrcyjanek/simplybs_private`

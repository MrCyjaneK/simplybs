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
go run . -host x86_64-linux-gnu -package zlib -list       # show a package + its dep tree
go run . -host x86_64-linux-gnu -package zlib -download   # fetch sources (verifies sha256)
go run . -host x86_64-linux-gnu -package zlib -build      # build a package (+ its deps)
go run . -world -host <triplet> -build                    # build everything for a host
```

Supported host triplets are defined in `host/main.go` (e.g. `x86_64-linux-gnu`,
`aarch64-linux-android`, `x86_64-w64-mingw32`, `aarch64-apple-darwin`).

## Making changes

There are many ways to solve problems, here are solutions, ordered by preference

1. Making a change to the .json file - adding an extra step.
2. Prefer an extra step over && - && use accepted mainly when you need to cd as each step starts at working dir
3. Adding existing dependency from the tree
4. Adding an extra dependnecy (**must be built from source**, unless we are moving a prebuilt dependency to an older version - in which event it is fine to add extra prebuilt)
5. Setting an ENV value depending on builder / host
6. Setting a custom command depending on builder / host (preference is on using ENV but it's just a preference)
7. Using `sed` to make a minimal inline change (replacing a header name, changing path)
8. Copying some binaries around (if a tool expects something on $PATH feel free to copy it to $NATIVEPREFIX/bin - every build gets a clean tree anyway)
9. Stubing a functionality (something expects output but can run without it? Why not run echo > output, expects nothing? just point to true)
10. Patching the code using .patch files


Things that are absolute nononononono

1. Pulling anything from the network in the build steps (no wget, curl, pip install, npm install that can reach the internet)
2. Pulling prebuilds because it is easier
3. When in doubt halt and ask human.


After changing sources run `go run . -lint`

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

## Cache

Cache is on. Do not dig for it. No `gh release`, no listing assets, no
inspecting tags, no probing `SIMPLYBS_CACHE_*`. `install.sh` already exports
`SIMPLYBS_CACHE_TAG` and `SIMPLYBS_CACHE_REPO` box-wide; `-build` auto-pulls
what it needs and auto-pushes what it built. It just works. Run the command:

```bash
go run . -host x86_64-linux-gnu -package zlib -build
```

## Clean

All builds start and end clean, if you need to inspect directory structure add exit 42 as last step (or other step if you ar einterested at certain point). Errors are not clean

## Logs

Never redirect to tail just to get last lines, always retain full logs to save time. Logs are very long.

## Dependencies

All dependencies are 1 level, no package pulls anything other than what's explicitly specified
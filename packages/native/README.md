# _native_ packages

These packages are intended to be ran on host when building.

## Windows (`windows_amd64`)

Builder userspace ABI is **Cygwin** (`NativeTriplet` = `x86_64-pc-cygwin`).

1. Provide a Cygwin+clang seed under `patches/native/_/windows_amd64_cygwin_clang/` (see README there).
2. `native/_/_` stages that seed into `$NATIVEPREFIX/_/`.
3. Rebuild `native/*` from source (make, bash, coreutils, …) with that compiler.
4. Build `native/crosstool-ng` then `native/ct-tools` from source for each `TOOL_TARGET`.

Suggested climb order toward ct-tools: `native/make` → `native/bash` → coreutils/sed/grep/… → autotools → `native/xz` → `native/crosstool-ng` → `native/ct-tools`.

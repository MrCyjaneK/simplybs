#!/bin/bash
ld64=@LD64_LLD@
sdk_max=@SDK_VERSION@
tmps=()
cleanup() {
  local f
  for f in "${tmps[@]}"; do
    rm -f "$f"
  done
}
trap cleanup EXIT

filter_response_file() {
  local in=$1 out
  out=$(mktemp)
  tmps+=("$out")
  python3 - "$in" "$out" "$sdk_max" <<'PY'
import shlex
import sys

src, dst, sdk_max = sys.argv[1:4]
data = open(src, "r", encoding="utf-8", errors="replace").read()
args = shlex.split(data)
out = []
i = 0
while i < len(args):
    arg = args[i]
    if arg in ("-undefined-version", "--undefined-version"):
        i += 1
        continue
    if arg in ("-soname", "-Wl,-soname"):
        i += 1
        if i < len(args):
            out.extend(["-install_name", args[i]])
        continue
    if arg.startswith("-Wl,-soname,"):
        out.extend(["-install_name", arg.split(",", 2)[2]])
        i += 1
        continue
    if arg in ("-retain-symbols-file", "-Wl,-retain-symbols-file", "-version-script", "-Wl,-version-script"):
        i += 2
        continue
    if arg.startswith("-Wl,-retain-symbols-file,") or arg.startswith("-Wl,-version-script,"):
        i += 1
        continue
    if arg == "-macosx_version_min":
        if i + 1 < len(args):
            out.extend(["-platform_version", "macos", args[i + 1], sdk_max])
            i += 2
            continue
    if arg.startswith("-macosx_version_min="):
        out.extend(["-platform_version", "macos", arg.split("=", 1)[1], sdk_max])
        i += 1
        continue
    out.append(arg)
    i += 1

with open(dst, "w", encoding="utf-8") as fh:
    for arg in out:
        if " " in arg or '"' in arg or "'" in arg:
            fh.write('"' + arg.replace("\\", "\\\\").replace('"', '\\"') + '" ')
        else:
            fh.write(arg + " ")
PY
  printf '@%s' "$out"
}

if [ $# -eq 1 ]; then
  case "$1" in
    -v)
      echo '@(#)PROGRAM:ld  PROJECT:ld64-711'
      exit 0
      ;;
    --version)
      exec -a ld64.lld "$ld64" --version
      ;;
  esac
fi

args=()
while [ $# -gt 0 ]; do
  case "$1" in
    -soname|-Wl,-soname)
      shift
      [ $# -gt 0 ] || { echo "host-ld-wrapper: -soname requires an argument" >&2; exit 1; }
      args+=(-install_name "$1")
      ;;
    -Wl,-soname,*)
      args+=(-install_name "${1#-Wl,-soname,}")
      ;;
    -retain-symbols-file|-Wl,-retain-symbols-file|-version-script|-Wl,-version-script)
      shift
      [ $# -gt 0 ] && shift
      continue
      ;;
    -Wl,-retain-symbols-file,*|-Wl,-version-script,*)
      ;;
    -macosx_version_min)
      shift
      args+=(-platform_version macos "$1" "$sdk_max")
      shift
      ;;
    -macosx_version_min=*)
      args+=(-platform_version macos "${1#-macosx_version_min=}" "$sdk_max")
      ;;
    -undefined-version|--undefined-version)
      ;;
    @*)
      f="${1#@}"
      if [ -f "$f" ]; then
        args+=("$(filter_response_file "$f")")
      else
        args+=("$1")
      fi
      ;;
    *)
      args+=("$1")
      ;;
  esac
  shift
done
exec -a ld64.lld "$ld64" "${args[@]}"

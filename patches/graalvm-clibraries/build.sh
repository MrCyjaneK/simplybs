#!/bin/bash
# Compile SVM clibraries (liblibchelper.a, libjvm.a, and on Darwin libdarwin.a)
# for $HOST from the Graal tree, then overlay them on the native/graalvm install.
# Analogous to rust-std staging into $NATIVEPREFIX/lib/rustlib/$RUST_TRIPLET.
set -euo pipefail
shopt -s nullglob

. "$PATCH_DIR/graalvm-clibraries/target.sh"

svm=substratevm/src
libchelper="$svm/com.oracle.svm.native.libchelper"
if [ ! -d "$libchelper/src" ]; then
	echo "graalvm-clibraries: Graal substratevm sources missing in $(pwd)" >&2
	ls -la
	exit 1
fi

if [ -z "${JAVA_HOME:-}" ] || [ ! -d "$JAVA_HOME/include" ]; then
	echo "graalvm-clibraries: JAVA_HOME must include JNI headers (JAVA_HOME=${JAVA_HOME:-})" >&2
	exit 1
fi

if [ -z "${CC:-}" ]; then
	echo "graalvm-clibraries: CC is unset" >&2
	exit 1
fi
if [ -z "${AR:-}" ]; then
	echo "graalvm-clibraries: AR is unset" >&2
	exit 1
fi

# $CC is often "clang ..." with flags baked in (Apple sysroot, Android API).
set -- $CC
CC_BIN=$1
shift
CC_EXTRA=("$@")

if [ ! -x "$CC_BIN" ] && ! command -v "$CC_BIN" >/dev/null 2>&1; then
	echo "graalvm-clibraries: C compiler not found: $CC_BIN (CC=$CC)" >&2
	exit 1
fi

# Stage a sibling of graalvm/ so ExtractTarGz does not strip the graalvm/ prefix
# (same idea as rust-std touching $NATIVEPREFIX/bin/.keep next to lib/rustlib).
mkdir -p "$STAGING_DIR$NATIVEPREFIX/bin"
: >"$STAGING_DIR$NATIVEPREFIX/bin/.graalvm-clibraries"

dest="$STAGING_DIR$NATIVEPREFIX/graalvm/lib/svm/clibraries/$GRAAL_TARGET"
objdir="$PWD/.clib-build"
mkdir -p "$dest/include" "$objdir"
cp -a "$libchelper/include/." "$dest/include/"

inc=(-I"$libchelper/include" -I"$JAVA_HOME/include")
for osinc in linux darwin win32 android; do
	if [ -d "$JAVA_HOME/include/$osinc" ]; then
		inc+=(-I"$JAVA_HOME/include/$osinc")
	fi
done

# Flags match substratevm/mx.substratevm/suite.py (non-Windows; mingw uses the
# posix set instead of MSVC -Zi/-MD).
case "$GRAAL_OS" in
	linux)
		helper_cflags=(-g -gdwarf-4 -fPIC -O2 -D_LITTLE_ENDIAN -ffunction-sections -fdata-sections -fvisibility=hidden -D_FORTIFY_SOURCE=0)
		jvm_cflags=(-g -gdwarf-4 -fPIC -O2 -ffunction-sections -fdata-sections -fvisibility=hidden -D_FORTIFY_SOURCE=0 -D_GNU_SOURCE)
		;;
	darwin)
		helper_cflags=(-g -fPIC -O2 -D_LITTLE_ENDIAN -ffunction-sections -fdata-sections -fvisibility=hidden -D_FORTIFY_SOURCE=0)
		jvm_cflags=(-g -fPIC -O2 -ffunction-sections -fdata-sections -fvisibility=hidden)
		darwin_cflags=(-x objective-c -fPIC -O1 -D_LITTLE_ENDIAN -ffunction-sections -fdata-sections -fvisibility=hidden -D_FORTIFY_SOURCE=0)
		;;
	windows)
		helper_cflags=(-g -O2 -D_LITTLE_ENDIAN)
		jvm_cflags=(-g -O2 -DJNIEXPORT=)
		;;
	*)
		echo "graalvm-clibraries: unhandled GRAAL_OS=$GRAAL_OS" >&2
		exit 1
		;;
esac

compile_archive() {
	local out=$1
	shift
	local -a cflags=("$@")
	local src obj
	local objs=()
	for src in "${srcs[@]}"; do
		[ -f "$src" ] || continue
		obj="$objdir/$(basename "$src" .c).o"
		# shellcheck disable=SC2086
		"$CC_BIN" "${CC_EXTRA[@]}" ${CFLAGS:-} ${CPPFLAGS:-} "${cflags[@]}" "${inc[@]}" -c "$src" -o "$obj"
		objs+=("$obj")
	done
	if [ ${#objs[@]} -eq 0 ]; then
		echo "graalvm-clibraries: no sources for $out" >&2
		exit 1
	fi
	"$AR" rcs "$out" "${objs[@]}"
}

srcs=("$libchelper/src"/*.c)
compile_archive "$dest/liblibchelper.a" "${helper_cflags[@]}"

if [ "$GRAAL_OS" = windows ]; then
	srcs=("$svm/com.oracle.svm.native.jvm.windows/src"/*.c)
else
	srcs=("$svm/com.oracle.svm.native.jvm.posix/src"/*.c)
fi
compile_archive "$dest/libjvm.a" "${jvm_cflags[@]}"

if [ "$GRAAL_OS" = darwin ]; then
	srcs=("$svm/com.oracle.svm.native.darwin/src"/*.c)
	compile_archive "$dest/libdarwin.a" "${darwin_cflags[@]}"
fi

echo "graalvm-clibraries: staged $GRAAL_TARGET ($HOST) into $dest"
ls -la "$dest"

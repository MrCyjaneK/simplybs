#!/bin/bash
# Compile Gluon HelloStaticLib to a native static library with GraalVM native-image
# and the simplybs C toolchain for $HOST.
# SVM clibraries come from graalvm-clibraries (overlay on $JAVA_HOME/lib/svm/clibraries).
set -euo pipefail
shopt -s globstar nullglob

. "$PATCH_DIR/graalvm-clibraries/target.sh"

if [ -z "${JAVA_HOME:-}" ] || [ ! -x "$JAVA_HOME/bin/javac" ] || [ ! -x "$JAVA_HOME/bin/native-image" ]; then
	echo "hellostaticlib: JAVA_HOME must point at GraalVM with javac and native-image (JAVA_HOME=${JAVA_HOME:-})" >&2
	exit 1
fi

clib_dest="$JAVA_HOME/lib/svm/clibraries/$GRAAL_TARGET"
if [ ! -f "$clib_dest/liblibchelper.a" ] || [ ! -f "$clib_dest/libjvm.a" ]; then
	echo "hellostaticlib: missing SVM clibraries in $clib_dest (build graalvm-clibraries for HOST=$HOST)" >&2
	ls -la "$JAVA_HOME/lib/svm/clibraries" 2>/dev/null || true
	exit 1
fi

if [ ! -d HelloStaticLib/src/main/java ]; then
	echo "hellostaticlib: HelloStaticLib sources not found in $(pwd)" >&2
	ls -la
	exit 1
fi

cd HelloStaticLib
mkdir -p build/classes gvm/tmp

# GraalVM 24 keeps IsolateThread / @CEntryPoint in library-support, not the
# default javac boot classpath.
libsupport="$JAVA_HOME/lib/svm/library-support.jar"
if [ ! -f "$libsupport" ]; then
	echo "hellostaticlib: missing $libsupport" >&2
	exit 1
fi
"$JAVA_HOME/bin/javac" -cp "$libsupport" -d build/classes src/main/java/hello/HelloStaticLib.java

# $CC is often "clang ..." / "gcc ..." with flags baked in (Apple sysroot, Android API).
# native-image wants the compiler binary and extra flags separately.
set -- $CC
CC_BIN=$1
shift
CC_EXTRA=("$@")

if [ ! -x "$CC_BIN" ] && ! command -v "$CC_BIN" >/dev/null 2>&1; then
	echo "hellostaticlib: C compiler not found: $CC_BIN (CC=$CC)" >&2
	exit 1
fi

ni_args=(
	--shared
	--no-fallback
	--target="$GRAAL_TARGET"
	--native-compiler-path="$CC_BIN"
	-H:+UnlockExperimentalVMOptions
	-H:+ExitAfterRelocatableImageWrite
	-H:+GenerateBuildArtifactsFile
	-H:Path="$PWD/gvm"
	-H:TempDirectory="$PWD/gvm/tmp"
	-H:Name=HelloStaticLib
	-H:NumberOfThreads="${NUM_CORES:-1}"
)

# Cross-arch: SVM defaults to UseCAPCache (cannot run target query code).
# Generate a cache instead by compiling query programs and running them
# (qemu-user-static binfmt on the builder, if registered).
case "$GRAAL_TARGET" in
	*aarch64*|*arm64*)
		if [ "$(uname -m)" != aarch64 ]; then
			mkdir -p "$PWD/gvm/cap"
			if [ ! -e /proc/sys/fs/binfmt_misc/qemu-aarch64 ] && [ -x /usr/bin/qemu-aarch64-static ]; then
				if [ -w /proc/sys/fs/binfmt_misc/register ] || [ -f /proc/sys/fs/binfmt_misc/register ]; then
					printf '%s\n' ':qemu-aarch64:M::\x7fELF\x02\x01\x01\x00\x00\x00\x00\x00\x00\x00\x00\x00\x02\x00\xb7\x00:\xff\xff\xff\xff\xff\xff\xff\x00\xff\xff\xff\xff\xff\xff\xff\xff\xfe\xff\xff\xff:/usr/bin/qemu-aarch64-static:' >/proc/sys/fs/binfmt_misc/register 2>/dev/null || true
				fi
			fi
			ni_args+=(-H:-UseCAPCache -H:+NewCAPCache -H:CAPCacheDir="$PWD/gvm/cap")
			# Query programs must be static; NDK default PIE needs /system/bin/linker64.
			ni_args+=(--native-compiler-options=-static)
		fi
		;;
esac

if [ -n "${GRAAL_LIBC:-}" ]; then
	ni_args+=(--libc="$GRAAL_LIBC")
fi

append_cc_opts() {
	# shellcheck disable=SC2086
	for opt in $1; do
		[ -n "$opt" ] || continue
		ni_args+=(--native-compiler-options="$opt")
	done
}
append_cc_opts "${CFLAGS:-}"
append_cc_opts "${CPPFLAGS:-}"
append_cc_opts "${CXXFLAGS:-}"
for opt in "${CC_EXTRA[@]}"; do
	ni_args+=(--native-compiler-options="$opt")
done

echo "hellostaticlib: native-image ${ni_args[*]} -cp build/classes:$libsupport hello.HelloStaticLib"
"$JAVA_HOME/bin/native-image" "${ni_args[@]}" -cp "build/classes:$libsupport" hello.HelloStaticLib

# Relocatable SVM .o files land in TempDirectory (SVM-*), not gvm/clib.
objs=(gvm/*.o gvm/*.obj gvm/tmp/SVM-*/*.o gvm/tmp/SVM-*/*.obj)
declare -A seen=()
uniq_objs=()
for f in "${objs[@]}"; do
	[ -f "$f" ] || continue
	case "$f" in
		*.o|*.obj) ;;
		*) continue ;;
	esac
	if [ -z "${seen[$f]:-}" ]; then
		seen[$f]=1
		uniq_objs+=("$f")
	fi
done
if [ ${#uniq_objs[@]} -eq 0 ]; then
	echo "hellostaticlib: native-image produced no relocatable objects" >&2
	ls -laR gvm || true
	if [ -n "${TMPDIR:-}" ]; then
		ls -laR "$TMPDIR" || true
	fi
	exit 1
fi
echo "hellostaticlib: archiving ${#uniq_objs[@]} object(s): ${uniq_objs[*]}"

incdir="$STAGING_DIR$PREFIX/include/hellostaticlib"
libdir="$STAGING_DIR$PREFIX/lib"
mkdir -p "$incdir" "$libdir"

"$AR" rcs "$libdir/libHelloStaticLib.a" "${uniq_objs[@]}"

for hdr in gvm/*.h gvm/**/*.h; do
	[ -f "$hdr" ] || continue
	cp "$hdr" "$incdir/"
done

ls -la "$libdir/libHelloStaticLib.a" "$incdir"

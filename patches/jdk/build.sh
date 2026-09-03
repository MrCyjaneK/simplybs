#!/bin/bash
# Build OpenJDK for the native builder. Required env: JDK_VERSION, BOOT_JDK.
set -euo pipefail
if [ -z "${JDK_VERSION:-}" ] || [ -z "${BOOT_JDK:-}" ]; then
	echo "JDK_VERSION and BOOT_JDK must be set" >&2
	exit 1
fi
if [ ! -x "$BOOT_JDK/bin/java" ]; then
	echo "boot JDK missing java: $BOOT_JDK/bin/java" >&2
	exit 1
fi

tools="$PWD/.jdk-tools"
mkdir -p "$tools"
cp "$PATCH_DIR/rust-std/xcrun" "$tools/xcrun"
chmod +x "$tools/xcrun"
if [ -x "$NATIVEPREFIX/_/bin/llvm-cxxfilt" ]; then
	ln -sf "$NATIVEPREFIX/_/bin/llvm-cxxfilt" "$tools/c++filt"
fi
if [ -x "$NATIVEPREFIX/_/bin/llvm-objdump" ]; then
	ln -sf "$NATIVEPREFIX/_/bin/llvm-objdump" "$tools/objdump"
fi
if [ ! -x "$NATIVEPREFIX/bin/which" ]; then
	printf '%s\n' '#!/bin/sh' 'command -v "$1"' >"$tools/which"
	chmod +x "$tools/which"
fi
export PATH="$NATIVEPREFIX/bin:$tools:$NATIVEPREFIX/_/bin:$PATH"
export AUTOCONF="${AUTOCONF:-$NATIVEPREFIX/bin/autoconf}"

config=(
	--with-boot-jdk="$BOOT_JDK"
	--with-toolchain-type=clang
	--disable-warnings-as-errors
	--enable-headless-only
	--with-native-debug-symbols=none
	--with-jvm-variants=server
	--with-debug-level=release
	--with-num-cores="$NUM_CORES"
	--with-jobs="$NUM_CORES"
	--with-zlib=system
	--with-extra-cflags="${CFLAGS:-}"
	--with-extra-cxxflags="${CXXFLAGS:-}"
)
case "$JDK_VERSION" in
11)
	config+=(--with-freetype=bundled)
	# In-tree gtest; removed in 16 (JDK-8245610). 20+ has no such flag.
	config+=(--disable-hotspot-gtest)
	;;
esac

os=$(uname -s)
extra_ldflags="${LDFLAGS:-}"
if [ "$os" = Linux ]; then
	# LLD 17+ errors on local symbols listed in Hotspot's generated mapfile.
	extra_ldflags="$extra_ldflags -Wl,--undefined-version"
fi
config+=(--with-extra-ldflags="$extra_ldflags")
if [ "$os" = Linux ]; then
	config+=(--with-alsa="$NATIVEPREFIX")
	config+=(--with-cups="$NATIVEPREFIX")
	config+=(--with-fontconfig="$NATIVEPREFIX")
	config+=(--with-fontconfig-include="$NATIVEPREFIX/include")
	config+=(--with-x="$NATIVEPREFIX")
	config+=(--x-includes="$NATIVEPREFIX/include")
	config+=(--x-libraries="$NATIVEPREFIX/lib")
fi
if [ "$os" = Darwin ]; then
	if [ -n "${SDK_PATH:-}" ]; then
		config+=(--with-sysroot="$SDK_PATH")
	fi
	if [ -n "${OSX_MIN_VERSION:-}" ]; then
		config+=(--with-macosx-version-min="$OSX_MIN_VERSION")
	fi
fi

bash configure "${config[@]}"
make JOBS="$NUM_CORES" images
if [ "${GRAAL_STATIC_LIBS:-}" = 1 ]; then
	# Graal native-image needs $JAVA_HOME/lib/static/<os>-<arch>/<libc>/*.a
	make JOBS="$NUM_CORES" static-libs-image
fi

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
os=$(uname -s)
# Hermetic xcrun: reports SDK_PATH. Never fall through to host /usr/bin.
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

# OpenJDK has --with-macosx-version-max only; MACOSX_VERSION_MIN is hardcoded.
# Honor simplybs OSX_MIN_VERSION via extra flags (JDK 20 configure rejects
# --with-macosx-version-min).
if [ "$os" = Darwin ] && [ -n "${OSX_MIN_VERSION:-}" ]; then
	min="-mmacosx-version-min=$OSX_MIN_VERSION"
	CFLAGS="${CFLAGS:+$CFLAGS }$min"
	CXXFLAGS="${CXXFLAGS:+$CXXFLAGS }$min"
	LDFLAGS="${LDFLAGS:+$LDFLAGS }$min"
fi
if [ "$os" = Darwin ] && [ -z "${SDK_PATH:-}" ]; then
	ver="${SDK_VERSION:-26.1}"
	for cand in \
		"$NATIVEPREFIX/SDK/MacOSX${ver}.sdk" \
		"$NATIVEPREFIX/_/SDK/MacOSX${ver}.sdk"; do
		if [ -d "$cand" ]; then
			SDK_PATH=$cand
			break
		fi
	done
	if [ -z "${SDK_PATH:-}" ]; then
		for cand in "$NATIVEPREFIX"/SDK/MacOSX*.sdk "$NATIVEPREFIX"/_/SDK/MacOSX*.sdk; do
			if [ -d "$cand" ]; then
				SDK_PATH=$cand
				break
			fi
		done
	fi
fi
if [ "$os" = Darwin ]; then
	if [ -z "${SDK_PATH:-}" ]; then
		echo "jdk: hermetic MacOSX SDK not found under $NATIVEPREFIX/SDK or $NATIVEPREFIX/_/SDK (need native/_ / osx-cross)" >&2
		exit 1
	fi
	# cctools-port does not ship mig. OpenJDK still requires mig, dsymutil,
	# xattr, and SetFile. Copy those binaries from Xcode/CLT by absolute
	# path into $tools so PATH stays hermetic (no /usr/bin).
	dev=""
	if [ -x /usr/bin/xcode-select ]; then
		dev=$(/usr/bin/xcode-select -p 2>/dev/null || true)
	fi
	stage_tool() {
		name=$1
		shift
		if [ -x "$tools/$name" ]; then
			return 0
		fi
		local src
		for src in "$@"; do
			if [ -n "$src" ] && [ -x "$src" ]; then
				cp -p "$src" "$tools/$name"
				chmod +x "$tools/$name"
				echo "jdk: staged $name <- $src"
				return 0
			fi
		done
		return 1
	}
	stage_tool migcom \
		"$NATIVEPREFIX/libexec/migcom" \
		"$NATIVEPREFIX/bin/migcom" \
		"${dev:+$dev/usr/libexec/migcom}" \
		"${dev:+$dev/Toolchains/XcodeDefault.xctoolchain/usr/libexec/migcom}" \
		/Library/Developer/CommandLineTools/usr/libexec/migcom \
		/usr/libexec/migcom || true
	cp "$PATCH_DIR/jdk/mig" "$tools/mig"
	chmod +x "$tools/mig"
	if [ -x "$NATIVEPREFIX/_/bin/llvm-dsymutil" ]; then
		ln -sf "$NATIVEPREFIX/_/bin/llvm-dsymutil" "$tools/dsymutil"
	elif [ -x "$NATIVEPREFIX/_/bin/llvm-dsymutil-21" ]; then
		ln -sf "$NATIVEPREFIX/_/bin/llvm-dsymutil-21" "$tools/dsymutil"
	elif [ -x "$NATIVEPREFIX/_/bin/dsymutil" ]; then
		ln -sf "$NATIVEPREFIX/_/bin/dsymutil" "$tools/dsymutil"
	else
		stage_tool dsymutil \
			"${dev:+$dev/Toolchains/XcodeDefault.xctoolchain/usr/bin/dsymutil}" \
			"${dev:+$dev/usr/bin/dsymutil}" \
			/Library/Developer/CommandLineTools/usr/bin/dsymutil \
			/usr/bin/dsymutil || true
	fi
	stage_tool xattr \
		"${dev:+$dev/usr/bin/xattr}" \
		/usr/bin/xattr || {
		printf '%s\n' '#!/bin/sh' 'exit 0' >"$tools/xattr"
		chmod +x "$tools/xattr"
		echo "jdk: staged no-op xattr"
	}
	stage_tool SetFile \
		"${dev:+$dev/usr/bin/SetFile}" \
		/usr/bin/SetFile || {
		printf '%s\n' '#!/bin/sh' 'exit 0' >"$tools/SetFile"
		chmod +x "$tools/SetFile"
		echo "jdk: staged no-op SetFile"
	}
	export SDK_PATH
	export SDKROOT=$SDK_PATH
	export MIGCOM=$tools/migcom
	export MIGCC="${MIGCC:-$(command -v clang || true)}"
	if [ ! -x "$tools/migcom" ]; then
		echo "jdk: hermetic migcom not staged (need Xcode/CLT libexec/migcom); PATH=$PATH" >&2
		exit 1
	fi
	if ! command -v mig >/dev/null 2>&1 || ! command -v dsymutil >/dev/null 2>&1; then
		echo "jdk: Darwin tools missing after staging; mig=$(command -v mig || true) dsymutil=$(command -v dsymutil || true) PATH=$PATH" >&2
		exit 1
	fi
	echo "jdk: Darwin SDK_PATH=$SDK_PATH mig=$(command -v mig) migcom=$MIGCOM dsymutil=$(command -v dsymutil)"
fi

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
fi

bash configure "${config[@]}"
make JOBS="$NUM_CORES" images
if [ "${GRAAL_STATIC_LIBS:-}" = 1 ]; then
	# Graal native-image needs $JAVA_HOME/lib/static/<os>-<arch>/<libc>/*.a
	make JOBS="$NUM_CORES" static-libs-image
fi

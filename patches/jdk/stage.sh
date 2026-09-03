#!/bin/bash
# Stage a JDK image as $NATIVEPREFIX/jdk-$JDK_VERSION plus a jlink JRE and PATH wrappers.
# Optional argument: directory that contains bin/java (or macOS Contents/Home).
set -euo pipefail
shopt -s nullglob
if [ -z "${JDK_VERSION:-}" ]; then
	echo "JDK_VERSION must be set" >&2
	exit 1
fi

find_java_home() {
	dir=$1
	if [ -x "$dir/bin/java" ]; then
		echo "$dir"
		return
	fi
	if [ -x "$dir/Contents/Home/bin/java" ]; then
		echo "$dir/Contents/Home"
		return
	fi
	return 1
}

src=""
if [ "${1:-}" != "" ]; then
	src=$(find_java_home "$1")
else
	for cand in build/*/images/jdk build/*/images/jdk-bundle/*.jdk/Contents/Home; do
		if [ -x "$cand/bin/java" ]; then
			src=$cand
			break
		fi
	done
fi
if [ -z "$src" ]; then
	echo "could not find JDK image to stage" >&2
	ls -la build/*/images 2>/dev/null || true
	exit 1
fi

dest="$STAGING_DIR$NATIVEPREFIX/jdk-$JDK_VERSION"
jre="$STAGING_DIR$NATIVEPREFIX/jre-$JDK_VERSION"
bindir="$STAGING_DIR$NATIVEPREFIX/bin"
mkdir -p "$dest" "$bindir"
cp -a "$src/." "$dest/"

if [ "${GRAAL_STATIC_LIBS:-}" = 1 ]; then
	static_src=""
	for cand in build/*/images/static-libs/lib build/*/images/static-libs; do
		if [ -d "$cand" ] && ls "$cand"/*.a >/dev/null 2>&1; then
			static_src=$cand
			break
		fi
	done
	if [ -z "$static_src" ]; then
		echo "jdk: GRAAL_STATIC_LIBS=1 but no static-libs image" >&2
		ls -la build/*/images 2>/dev/null || true
		exit 1
	fi
	machine=$(uname -m)
	case "$machine" in
		x86_64) graal_arch=amd64 ;;
		aarch64|arm64) graal_arch=aarch64 ;;
		*) graal_arch=$machine ;;
	esac
	graal_os=$(uname -s | tr '[:upper:]' '[:lower:]')
	# Graal native-image looks under lib/static/<os>-<arch>/<libc> on
	# Linux and lib/static/<os>-<arch> on Darwin.
	if [ "$graal_os" = linux ]; then
		static_dest="$dest/lib/static/${graal_os}-${graal_arch}/glibc"
	else
		static_dest="$dest/lib/static/${graal_os}-${graal_arch}"
	fi
	mkdir -p "$static_dest"
	cp -a "$static_src"/*.a "$static_dest/"
	echo "jdk: staged static libs in $static_dest"
	ls "$static_dest"
fi

if [ "${1:-}" = "" ]; then
	for cand in build/*/images/jre; do
		if [ -x "$cand/bin/java" ]; then
			mkdir -p "$jre"
			cp -a "$cand/." "$jre/"
			break
		fi
	done
fi
if [ ! -x "$jre/bin/java" ]; then
	"$dest/bin/jlink" \
		--module-path "$dest/jmods" \
		--add-modules ALL-MODULE-PATH \
		--strip-debug \
		--no-header-files \
		--no-man-pages \
		--output "$jre"
fi

for tool in java javac jar javadoc javap jlink jmod jdeps keytool; do
	if [ ! -x "$dest/bin/$tool" ]; then
		continue
	fi
	sed -e "s/@JDK_VERSION@/$JDK_VERSION/g" -e "s/@TOOL@/$tool/g" \
		"$PATCH_DIR/jdk/wrapper.sh" >"$bindir/$tool"
	chmod +x "$bindir/$tool"
	cp "$bindir/$tool" "$bindir/${tool}${JDK_VERSION}"
done
"$dest/bin/java" -version
"$dest/bin/javac" -version

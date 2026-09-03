#!/bin/bash
# Stage the mx GraalVM home as $NATIVEPREFIX/graalvm plus PATH wrappers.
set -euo pipefail
shopt -s nullglob

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
if [ -f .graalvm-home ]; then
	src=$(find_java_home "$(cat .graalvm-home)") || src=""
fi
if [ -z "$src" ] && [ "${1:-}" != "" ]; then
	src=$(find_java_home "$1") || src=""
fi
if [ -z "$src" ]; then
	for cand in sdk/latest_graalvm_home sdk/latest_graalvm vm/*/graalvm-*; do
		if src=$(find_java_home "$cand"); then
			break
		fi
		src=""
	done
fi
if [ -z "$src" ]; then
	echo "graalvm: could not find built GraalVM home to stage" >&2
	ls -la
	exit 1
fi

dest="$STAGING_DIR$NATIVEPREFIX/graalvm"
bindir="$STAGING_DIR$NATIVEPREFIX/bin"
mkdir -p "$dest" "$bindir"
cp -a "$src/." "$dest/"

for tool in java javac jar javadoc javap jlink jmod jdeps keytool native-image; do
	if [ ! -x "$dest/bin/$tool" ]; then
		continue
	fi
	sed -e "s/@TOOL@/$tool/g" "$PATCH_DIR/graalvm/wrapper.sh" >"$bindir/$tool"
	chmod +x "$bindir/$tool"
done

"$dest/bin/java" -version
"$dest/bin/javac" -version
"$dest/bin/native-image" --version

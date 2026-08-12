#!/bin/sh
# On Windows/Cygwin, several gnulib packages fail with wint_t / uchar issues.
# Stage the matching tools already present in the Cygwin seed ($NATIVEPREFIX/_/bin).
set -e
bin="$NATIVEPREFIX/_/bin"
dest="$STAGING_DIR$NATIVEPREFIX/bin"
mkdir -p "$dest"
copy_one() {
	name=$1
	if [ -f "$bin/$name.exe" ]; then
		cp -f "$bin/$name.exe" "$dest/$name.exe"
		return
	fi
	if [ -f "$bin/$name" ]; then
		cp -f "$bin/$name" "$dest/$name"
		return
	fi
	# Common aliases shipped as argv0 copies of another tool.
	case "$name" in
	egrep|fgrep)
		cp -f "$bin/grep.exe" "$dest/$name.exe"
		;;
	gunzip|zcat|gzcat)
		cp -f "$bin/gzip.exe" "$dest/$name.exe"
		;;
	*)
		echo "seed tool missing: $bin/$name(.exe)" >&2
		exit 1
		;;
	esac
}
for name in "$@"; do
	copy_one "$name"
done

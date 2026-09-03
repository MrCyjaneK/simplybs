# Map simplybs $HOST to GraalVM native-image --target / clibraries directory.
# Sourced by graalvm-clibraries and hellostaticlib.

if [ -z "${HOST:-}" ]; then
	echo "graalvm-clibraries: HOST is unset" >&2
	exit 1
fi

GRAAL_LIBC=
GRAAL_OS=
case "$HOST" in
	x86_64-linux-gnu)
		GRAAL_TARGET=linux-amd64
		GRAAL_LIBC=glibc
		GRAAL_OS=linux
		;;
	aarch64-linux-gnu)
		GRAAL_TARGET=linux-aarch64
		GRAAL_LIBC=glibc
		GRAAL_OS=linux
		;;
	aarch64-linux-android)
		GRAAL_TARGET=android-aarch64
		GRAAL_LIBC=bionic
		GRAAL_OS=linux
		;;
	x86_64-linux-android)
		GRAAL_TARGET=linux-amd64
		GRAAL_LIBC=bionic
		GRAAL_OS=linux
		;;
	x86_64-w64-mingw32)
		GRAAL_TARGET=windows-amd64
		GRAAL_OS=windows
		;;
	aarch64-apple-darwin)
		GRAAL_TARGET=darwin-aarch64
		GRAAL_OS=darwin
		;;
	x86_64-apple-darwin)
		GRAAL_TARGET=darwin-amd64
		GRAAL_OS=darwin
		;;
	aarch64-apple-ios)
		GRAAL_TARGET=ios-aarch64
		GRAAL_OS=darwin
		;;
	aarch64-apple-ios-simulator)
		GRAAL_TARGET=ios-aarch64
		GRAAL_OS=darwin
		;;
	*)
		echo "graalvm-clibraries: no GraalVM --target for HOST=$HOST" >&2
		exit 1
		;;
esac

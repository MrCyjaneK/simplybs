#!/bin/bash
# Build GraalVM CE (native-image) from source with native/jdk as JAVA_HOME.
set -euo pipefail

if [ -z "${JAVA_HOME:-}" ] || [ ! -x "$JAVA_HOME/bin/java" ]; then
	echo "graalvm: JAVA_HOME must be a JDK with bin/java (JAVA_HOME=${JAVA_HOME:-})" >&2
	exit 1
fi
if [ ! -x mx/mx ]; then
	echo "graalvm: mx/mx missing in $(pwd)" >&2
	ls -la
	exit 1
fi
if [ ! -d vm ]; then
	echo "graalvm: oracle/graal vm/ suite missing in $(pwd)" >&2
	ls -la
	exit 1
fi

# mx probes glibc with `ldd --version` (not on the hermetic PATH).
tools="$PWD/.graal-tools"
mkdir -p "$tools"
if [ -x /usr/bin/ldd ]; then
	ln -sf /usr/bin/ldd "$tools/ldd"
fi
if [ -x "$NATIVEPREFIX/_/bin/llvm-objdump" ]; then
	ln -sf "$NATIVEPREFIX/_/bin/llvm-objdump" "$tools/objdump"
fi
if [ -x "$NATIVEPREFIX/_/bin/llvm-nm" ]; then
	ln -sf "$NATIVEPREFIX/_/bin/llvm-nm" "$tools/nm"
fi
if [ -x "$NATIVEPREFIX/_/bin/llvm-ar" ]; then
	ln -sf "$NATIVEPREFIX/_/bin/llvm-ar" "$tools/ar"
fi
if [ -x "$NATIVEPREFIX/_/bin/llvm-cxxfilt" ]; then
	ln -sf "$NATIVEPREFIX/_/bin/llvm-cxxfilt" "$tools/c++filt"
fi
# simplybs $NATIVEPREFIX/_/bin/clang is a dash wrapper that execs the
# aarch64 gcc-toolchain binary for -print-multi-* and via -B. libffi's
# configure then runs that ELF and dies. Drive clang-21 + sysroot directly.
clang21="$NATIVEPREFIX/_/bin/clang-21"
sysroot="$NATIVEPREFIX/_/sysroot"
write_cc() {
	dest=$1
	mode=$2
	if [ "$mode" = g++ ]; then
		drv='--driver-mode=g++'
	else
		drv=''
	fi
	cat >"$dest" <<EOF
#!/bin/sh
for arg in "\$@"; do
	case "\$arg" in
		-print-multi-os-directory) echo ../lib; exit 0 ;;
		-print-multi-directory) echo .; exit 0 ;;
		-print-sysroot-headers-suffix) echo; exit 0 ;;
		-dumpmachine) echo x86_64-linux-gnu; exit 0 ;;
	esac
done
exec "$clang21" $drv -Qunused-arguments -target x86_64-linux-gnu --sysroot="$sysroot" -fuse-ld=lld "\$@"
EOF
	chmod +x "$dest"
}
if [ -x "$clang21" ]; then
	write_cc "$tools/cc" gcc
	write_cc "$tools/gcc" gcc
	write_cc "$tools/c++" g++
	write_cc "$tools/g++" g++
	write_cc "$tools/x86_64-linux-gnu-gcc" gcc
	write_cc "$tools/x86_64-linux-gnu-g++" g++
	gtbin="$NATIVEPREFIX/_/libexec/gcc-toolchain/bin"
	if [ -d "$gtbin" ]; then
		rm -f "$gtbin/gcc" "$gtbin/g++"
		write_cc "$gtbin/x86_64-linux-gnu-gcc" gcc
		write_cc "$gtbin/x86_64-linux-gnu-g++" g++
		write_cc "$gtbin/gcc" gcc
		write_cc "$gtbin/g++" g++
	fi
	export CC="$tools/gcc"
	export CXX="$tools/g++"
fi
if [ ! -x "$NATIVEPREFIX/bin/which" ]; then
	printf '%s\n' '#!/bin/sh' 'command -v "$1"' >"$tools/which"
	chmod +x "$tools/which"
fi
export PATH="$tools:$PWD/mx:$NATIVEPREFIX/bin:$NATIVEPREFIX/_/bin:$PATH"
export JAVA_HOME
export JVMCI_VERSION_CHECK="${JVMCI_VERSION_CHECK:-ignore}"
export MX_PYTHON="${MX_PYTHON:-$NATIVEPREFIX/bin/python3.14}"
export PYTHONPATH="$PATCH_DIR/graalvm/pylib${PYTHONPATH:+:$PYTHONPATH}"
# Graal's bundled LLVM clang++ probes host GCC. Ubuntu 24 has GCC 14 without
# libstdc++-14-dev, so #include <iostream> fails. Point at the hermetic
# toolchain (the same sysroot simplybs clang uses).
cxxinc=""
for d in "$NATIVEPREFIX/_/include/c++"/*; do
	if [ -f "$d/iostream" ]; then
		cxxinc=$d
		break
	fi
done
if [ -n "$cxxinc" ]; then
	export CPLUS_INCLUDE_PATH="$cxxinc:$cxxinc/x86_64-linux-gnu${CPLUS_INCLUDE_PATH:+:$CPLUS_INCLUDE_PATH}"
fi
if [ -d "$NATIVEPREFIX/_/sysroot/lib" ]; then
	export LIBRARY_PATH="$NATIVEPREFIX/_/sysroot/lib${LIBRARY_PATH:+:$LIBRARY_PATH}"
fi
if [ -z "${SSL_CERT_FILE:-}" ] && [ -f /etc/ssl/certs/ca-certificates.crt ]; then
	export SSL_CERT_FILE=/etc/ssl/certs/ca-certificates.crt
	export REQUESTS_CA_BUNDLE="$SSL_CERT_FILE"
	export CURL_CA_BUNDLE="$SSL_CERT_FILE"
fi

# libffi 3.4.4 tramp.c calls open_temp_exec_file(); clang rejects the
# implicit declaration. Disable static trampolines (Graal NFI does not need them).
"$MX_PYTHON" - <<'PY'
from pathlib import Path
p = Path("truffle/mx.truffle/mx_truffle.py")
t = p.read_text()
needle = "'--disable-shared',"
insert = "'--disable-shared',\n                              '--disable-exec-static-tramp',"
if "--disable-exec-static-tramp" not in t:
	if needle not in t:
		raise SystemExit("graalvm: --disable-shared not found in mx_truffle.py")
	p.write_text(t.replace(needle, insert, 1))
PY

cp "$PATCH_DIR/graalvm/ce.env" vm/mx.vm/simplybs

"$JAVA_HOME/bin/java" -version
mx/mx --version

cd vm
mx --java-home "$JAVA_HOME" --env simplybs build
mx --java-home "$JAVA_HOME" --env simplybs graalvm-home >"$PWD/../.graalvm-home"
home=$(cat "$PWD/../.graalvm-home")
echo "graalvm: built at $home"
test -x "$home/bin/native-image"
"$home/bin/native-image" --version

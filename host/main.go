package host

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mrcyjanek/simplybs/crash"
)

type Host struct {
	Triplet string
	Env     []string
}

func GetPackagesDir() string {
	if os.Getenv("SIMPLYBS_PACKAGES_DIR") != "" {
		return os.Getenv("SIMPLYBS_PACKAGES_DIR")
	}
	if os.Getenv("SIMPLYBS_DATA_DIR") != "" {
		guessPath := filepath.Join(os.Getenv("SIMPLYBS_DATA_DIR"), "..", "packages")
		if _, err := os.Stat(guessPath); err == nil {
			return guessPath
		}
	}
	wd, err := os.Getwd()
	crash.Handle(err)
	return filepath.Join(wd, "packages")
}

func DataDirRoot() string {
	if os.Getenv("SIMPLYBS_DATA_DIR") != "" {
		return os.Getenv("SIMPLYBS_DATA_DIR")
	}
	buildDir, err := os.Getwd()
	crash.Handle(err)
	return filepath.Join(buildDir, ".buildlib")
}

func DataDir() string {
	return filepath.Join(DataDirRoot(), runtime.GOOS+"_"+runtime.GOARCH)
}

func (h *Host) GetNativeEnvPath() string {
	if os.Getenv("SIMPLYBS_NATIVE_ENV_DIR") != "" {
		return filepath.Join(os.Getenv("SIMPLYBS_NATIVE_ENV_DIR"))
	}
	dir := h.GetEnvPath()
	return filepath.Join(dir, "native")
}

func (h *Host) GetEnvPath() string {
	if os.Getenv("SIMPLYBS_ENV_DIR") != "" {
		return filepath.Join(os.Getenv("SIMPLYBS_ENV_DIR"))
	}
	dir := DataDirRoot()
	return filepath.Join(dir, "env")
}

var SupportedHosts = map[string]*Host{
	"aarch64-apple-darwin": {
		Triplet: "aarch64-apple-darwin",
		Env: []string{
			"*:*:HOST=aarch64-apple-darwin",
			"*:*:TARGET=aarch64-apple-darwin",
			"*:*:RUST_TRIPLET=aarch64-apple-darwin",
			"*:*:ARCH=aarch64",
			"*:*:CMAKE_SYSTEM_NAME=Darwin",
			"*:*:SDK_VERSION=26.1",
			"*:*:SDK_PATH=$NATIVEPREFIX/SDK/MacOSX$SDK_VERSION.sdk",
			"*:*:OSX_MIN_VERSION=13.0",
			"*:*:LD64_VERSION=609",
			"*:*:CC_target=arm64-apple-darwin",
			"*:*:CC=aarch64-apple-darwin-clang -mmacosx-version-min=$OSX_MIN_VERSION -isysroot $SDK_PATH -I$PREFIX/include",
			"*:*:CXX=aarch64-apple-darwin-clang++ -mmacosx-version-min=$OSX_MIN_VERSION -isysroot $SDK_PATH -I$PREFIX/include",
			"*:*:CFLAGS=",
			"*:*:CXXFLAGS=$CFLAGS -stdlib=libc++",
			"*:*:CPPFLAGS=",
			"*:*:RANLIB=llvm-ranlib",
			"*:*:STRIP=strip",
			"*:*:AR=llvm-ar",
			"*:*:NM=llvm-nm",
			"*:*:LIBTOOL=llvm-libtool-darwin",
			"*:*:LDFLAGS=-L$PREFIX/lib -lc++ -lc++abi",
		},
	},
	"x86_64-apple-darwin": {
		Triplet: "x86_64-apple-darwin",
		Env: []string{
			"*:*:HOST=x86_64-apple-darwin",
			"*:*:TARGET=x86_64-apple-darwin",
			"*:*:RUST_TRIPLET=x86_64-apple-darwin",
			"*:*:ARCH=x86_64",
			"*:*:CMAKE_SYSTEM_NAME=Darwin",
			"*:*:SDK_VERSION=26.1",
			"*:*:SDK_PATH=$NATIVEPREFIX/SDK/MacOSX$SDK_VERSION.sdk",
			"*:*:OSX_MIN_VERSION=10.16",
			"*:*:LD64_VERSION=609",
			"*:*:CC_target=x86_64-apple-darwin",
			"*:*:CC=x86_64-apple-darwin-clang -mmacosx-version-min=$OSX_MIN_VERSION -isysroot $SDK_PATH -I$PREFIX/include",
			"*:*:CXX=x86_64-apple-darwin-clang++ -mmacosx-version-min=$OSX_MIN_VERSION -isysroot $SDK_PATH -I$PREFIX/include",
			"*:*:CFLAGS=",
			"*:*:CXXFLAGS=$CFLAGS -stdlib=libc++",
			"*:*:CPPFLAGS=",
			"*:*:RANLIB=ranlib",
			"*:*:AR=ar",
			"*:*:STRIP=strip",
			"*:*:NM=nm",
			"*:*:LIBTOOL=llvm-libtool-darwin",
			"*:*:LDFLAGS=-L$PREFIX/lib -lc++ -lc++abi",
		},
	},
	"aarch64-apple-ios": {
		Triplet: "aarch64-apple-ios",
		Env: []string{
			"*:*:HOST=aarch64-apple-ios",
			"*:*:TARGET=aarch64-apple-ios",
			"*:*:RUST_TRIPLET=aarch64-apple-ios",
			"*:*:ARCH=aarch64",
			"*:*:CMAKE_SYSTEM_NAME=iOS",
			"*:*:IOS_MIN_VERSION=12",
			"*:*:LD64_VERSION=609",
			"*:*:SDK_VERSION=26.1",
			"*:*:SDK_PATH=$NATIVEPREFIX/SDK/iPhoneOS$SDK_VERSION.sdk",
			"*:*:CC_target=aarch64-apple-ios",
			"*:*:CC=aarch64-apple-ios-clang -target $CC_target -mios-version-min=$IOS_MIN_VERSION -isysroot $SDK_PATH -I$PREFIX/include",
			"*:*:CXX=aarch64-apple-ios-clang++ -target $CC_target -mios-version-min=$IOS_MIN_VERSION -isysroot $SDK_PATH -I$PREFIX/include",
			"*:*:CFLAGS=",
			"*:*:CPPFLAGS=",
			"*:*:CXXFLAGS=$CFLAGS -stdlib=libc++",
			"*:*:RANLIB=ranlib",
			"*:*:AR=ar",
			"*:*:NM=nm",
			"*:*:STRIP=strip",
			"*:*:LIBTOOL=libtool",
			"*:*:LDFLAGS=-lc++ -lc++abi",
			"*:*:LD=lld",
		},
	},
	"aarch64-apple-ios-simulator": {
		Triplet: "aarch64-apple-ios-simulator",
		Env: []string{
			"*:*:HOST=aarch64-apple-ios-simulator",
			"*:*:TARGET=aarch64-apple-ios-simulator",
			"*:*:RUST_TRIPLET=aarch64-apple-ios-sim",
			"*:*:ARCH=aarch64",
			"*:*:CMAKE_SYSTEM_NAME=iOS",
			"*:*:IOS_MIN_VERSION=12",
			"*:*:LD64_VERSION=609",
			"*:*:SDK_VERSION=26.1",
			"*:*:SDK_PATH=$NATIVEPREFIX/SDK/iPhoneSimulator$SDK_VERSION.sdk",
			"*:*:CC_target=aarch64-apple-ios-simulator",
			"*:*:CC=aarch64-apple-ios-simulator-clang -target $CC_target -mios-version-min=$IOS_MIN_VERSION -isysroot $SDK_PATH -I$PREFIX/include",
			"*:*:CXX=aarch64-apple-ios-simulator-clang++ -target $CC_target -mios-version-min=$IOS_MIN_VERSION -isysroot $SDK_PATH -I$PREFIX/include",
			"*:*:CFLAGS=",
			"*:*:CPPFLAGS=",
			"*:*:CXXFLAGS=$CFLAGS -stdlib=libc++",
			"*:*:RANLIB=ranlib",
			"*:*:AR=ar",
			"*:*:NM=nm",
			"*:*:STRIP=strip",
			"*:*:LIBTOOL=libtool",
			"*:*:LDFLAGS=-lc++ -lc++abi",
		},
	},
	// "x86_64-linux-musl": {
	// 	Triplet: "x86_64-linux-musl",
	// 	Env: []string{
	// 		"*:*:HOST=x86_64-linux-musl",
	// 		"*:*:TARGET=x86_64-linux-musl",
	// 		"*:*:CC_target=x86_64-linux-musl",
	// 		"*:*:CC=x86_64-linux-musl-gcc",
	// 		"*:*:CXX=x86_64-linux-musl-g++",
	// 		"*:*:CXXFLAGS=$CFLAGS",
	// 		"*:*:RANLIB=x86_64-linux-musl-ranlib",
	// 		"*:*:AR=x86_64-linux-musl-ar",
	// 		"*:*:AS=x86_64-linux-musl-as",
	// 		"*:*:LIBTOOL=x86_64-linux-musl-libtool",
	// 		"*:*:OBJCOPY=x86_64-linux-musl-objcopy",
	// 		"*:*:OBJDUMP=x86_64-linux-musl-objdump",
	// 		"*:*:STRIP=x86_64-linux-musl-strip",
	// 		"*:*:READELF=x86_64-linux-musl-readelf",
	// 		"*:*:LD=x86_64-linux-musl-ld",
	// 		"*:*:NM=x86_64-linux-musl-nm",
	// 	},
	// },
	"x86_64-w64-mingw32": {
		Triplet: "x86_64-w64-mingw32",
		Env: []string{
			"*:*:HOST=x86_64-w64-mingw32",
			"*:*:TARGET=x86_64-w64-mingw32",
			"*:*:RUST_TRIPLET=x86_64-pc-windows-gnu",
			"*:*:ARCH=x86_64",
			"*:*:CMAKE_SYSTEM_NAME=Windows",
			"*:*:CC_target=x86_64-w64-mingw32",
			"*:*:RC=x86_64-w64-mingw32-windres",
			"*:*:RCFLAGS=-I$PREFIX/sysroot/usr/x86_64-w64-mingw32/include -I$PREFIX/include -I$PREFIX/usr/include",
			"*:*:CC=x86_64-w64-mingw32-gcc",
			"*:*:CFLAGS=--sysroot=$PREFIX/sysroot -I$PREFIX/sysroot/usr/x86_64-w64-mingw32/include -I$PREFIX/include -I$PREFIX/usr/include",
			"*:*:CXX=x86_64-w64-mingw32-g++",
			"*:*:CXXFLAGS=--sysroot=$PREFIX/sysroot -I$PREFIX/sysroot/usr/x86_64-w64-mingw32/include -I$PREFIX/include -I$PREFIX/include/c++/15.2.0 -I$PREFIX/include/c++/15.2.0/x86_64-w64-mingw32 -I$PREFIX/usr/include",
			"*:*:CPPFLAGS=-I$PREFIX/sysroot/usr/x86_64-w64-mingw32/include -I$PREFIX/include -I$PREFIX/usr/include",
			"*:*:RANLIB=x86_64-w64-mingw32-ranlib",
			"*:*:AR=x86_64-w64-mingw32-ar",
			"*:*:AS=x86_64-w64-mingw32-as",
			"*:*:LIBTOOL=x86_64-w64-mingw32-libtool",
			"*:*:OBJCOPY=x86_64-w64-mingw32-objcopy",
			"*:*:OBJDUMP=x86_64-w64-mingw32-objdump",
			"*:*:STRIP=x86_64-w64-mingw32-strip",
			"*:*:READELF=x86_64-w64-mingw32-readelf",
			"*:*:LD=x86_64-w64-mingw32-ld",
			"*:*:NM=x86_64-w64-mingw32-nm",
		},
	},
	"x86_64-linux-gnu": {
		Triplet: "x86_64-linux-gnu",
		Env: []string{
			"*:*:HOST=x86_64-linux-gnu",
			"*:*:TARGET=x86_64-linux-gnu",
			"*:*:RUST_TRIPLET=x86_64-unknown-linux-gnu",
			"*:*:ARCH=x86_64",
			"*:*:CMAKE_SYSTEM_NAME=Linux",
			"*:*:CC_target=x86_64-linux-gnu",
			"*:*:CC=x86_64-linux-gnu-gcc",
			"*:*:CFLAGS=--sysroot=$PREFIX/sysroot -I$PREFIX/include -I$PREFIX/usr/include",
			"*:*:CXX=x86_64-linux-gnu-g++",
			"*:*:CXXFLAGS=--sysroot=$PREFIX/sysroot -I$PREFIX/include -I$PREFIX/include/c++/15.2.0 -I$PREFIX/include/c++/15.2.0/x86_64-linux-gnu -I$PREFIX/usr/include",
			"*:*:CPPFLAGS=-I$PREFIX/include -I$PREFIX/usr/include",
			"*:*:RANLIB=x86_64-linux-gnu-ranlib",
			"*:*:AR=x86_64-linux-gnu-ar",
			"*:*:AS=x86_64-linux-gnu-as",
			"*:*:LIBTOOL=x86_64-linux-gnu-libtool",
			"*:*:OBJCOPY=x86_64-linux-gnu-objcopy",
			"*:*:OBJDUMP=x86_64-linux-gnu-objdump",
			"*:*:STRIP=x86_64-linux-gnu-strip",
			"*:*:READELF=x86_64-linux-gnu-readelf",
			"*:*:LD=x86_64-linux-gnu-ld",
			"*:*:NM=x86_64-linux-gnu-nm",
		},
	},
	"aarch64-linux-gnu": {
		Triplet: "aarch64-linux-gnu",
		Env: []string{
			"*:*:HOST=aarch64-linux-gnu",
			"*:*:TARGET=aarch64-linux-gnu",
			"*:*:RUST_TRIPLET=aarch64-unknown-linux-gnu",
			"*:*:ARCH=aarch64",
			"*:*:CMAKE_SYSTEM_NAME=Linux",
			"*:*:CC_target=aarch64-linux-gnu",
			"*:*:CC=aarch64-linux-gnu-gcc",
			"*:*:CXX=aarch64-linux-gnu-g++",
			"*:*:CXXFLAGS= --sysroot=$PREFIX/sysroot -I$PREFIX/include -I$PREFIX/include/c++/15.2.0 -I$PREFIX/include/c++/15.2.0/aarch64-linux-gnu -I$PREFIX/usr/include",
			"*:*:CFLAGS=--sysroot=$PREFIX/sysroot -I$PREFIX/include -I$PREFIX/usr/include",
			"*:*:CPPFLAGS=-I$PREFIX/include -I$PREFIX/usr/include",
			"*:*:RANLIB=aarch64-linux-gnu-ranlib",
			"*:*:AR=aarch64-linux-gnu-ar",
			"*:*:AS=aarch64-linux-gnu-as",
			"*:*:LIBTOOL=aarch64-linux-gnu-libtool",
			"*:*:OBJCOPY=aarch64-linux-gnu-objcopy",
			"*:*:OBJDUMP=aarch64-linux-gnu-objdump",
			"*:*:STRIP=aarch64-linux-gnu-strip",
			"*:*:READELF=aarch64-linux-gnu-readelf",
			"*:*:LD=aarch64-linux-gnu-ld",
			"*:*:NM=aarch64-linux-gnu-nm",
		},
	},
	"aarch64-linux-android": {
		Triplet: "aarch64-linux-android",
		Env: []string{
			"*:*:HOST=aarch64-linux-android",
			"*:*:TARGET=aarch64-linux-android",
			"*:*:RUST_TRIPLET=aarch64-linux-android",
			"*:*:ARCH=aarch64",
			"*:*:CMAKE_SYSTEM_NAME=Android",
			"*:*:CC_target=aarch64-linux-android",
			"*:*:CC=aarch64-linux-android21-clang",
			"*:*:CXX=aarch64-linux-android21-clang++",
			"*:*:CFLAGS=$CFLAGS",
			"*:*:CPPFLAGS=",
			"*:*:CXXFLAGS=$CXXFLAGS -stdlib=libc++",
			"*:*:RANLIB=llvm-ranlib",
			"*:*:AR=llvm-ar",
			"*:*:STRIP=llvm-strip",
			"*:*:NM=llvm-nm",
			"*:*:AS=llvm-as",
			"*:*:LIBTOOL=libtool",
			"*:*:ANDROID_NDK_HOME=$NATIVEPREFIX/",
			"*:*:LDFLAGS=$LDFLAGS -lc -lc++abi -lm",
		},
	},
	"x86_64-linux-android": {
		Triplet: "x86_64-linux-android",
		Env: []string{
			"*:*:HOST=x86_64-linux-android",
			"*:*:TARGET=x86_64-linux-android",
			"*:*:RUST_TRIPLET=x86_64-linux-android",
			"*:*:ARCH=x86_64",
			"*:*:CMAKE_SYSTEM_NAME=Android",
			"*:*:CC_target=x86_64-linux-android",
			"*:*:CC=x86_64-linux-android21-clang",
			"*:*:CXX=x86_64-linux-android21-clang++",
			"*:*:CFLAGS=$CFLAGS",
			"*:*:CPPFLAGS=",
			"*:*:CXXFLAGS=$CXXFLAGS -stdlib=libc++",
			"*:*:RANLIB=llvm-ranlib",
			"*:*:AR=llvm-ar",
			"*:*:NM=llvm-nm",
			"*:*:STRIP=llvm-strip",
			"*:*:AS=llvm-as",
			"*:*:LIBTOOL=libtool",
			"*:*:ANDROID_NDK_HOME=$NATIVEPREFIX/",
			"*:*:LDFLAGS=$LDFLAGS -lc -lc++abi -lm"},
	},
	"armv7a-linux-androideabi": {
		Triplet: "armv7a-linux-androideabi",
		Env: []string{
			"*:*:HOST=armv7a-linux-androideabi",
			"*:*:TARGET=armv7a-linux-androideabi",
			"*:*:RUST_TRIPLET=armv7-linux-androideabi",
			"*:*:ARCH=armv7a",
			"*:*:CMAKE_SYSTEM_NAME=Android",
			"*:*:CC_target=armv7a-linux-androideabi",
			"*:*:CC=armv7a-linux-androideabi21-clang",
			"*:*:CXX=armv7a-linux-androideabi21-clang++",
			"*:*:CFLAGS=$CFLAGS",
			"*:*:CPPFLAGS=",
			"*:*:CXXFLAGS=$CXXFLAGS -stdlib=libc++",
			"*:*:RANLIB=llvm-ranlib",
			"*:*:AR=llvm-ar",
			"*:*:NM=llvm-nm",
			"*:*:STRIP=llvm-strip",
			"*:*:AS=llvm-as",
			"*:*:LIBTOOL=libtool",
			"*:*:ANDROID_NDK_HOME=$NATIVEPREFIX/",
			"*:*:LDFLAGS=$LDFLAGS -lc -lc++abi -lm"},
	},
}

func shellOutput(cmd string) string {
	output, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return "$(" + cmd + ")"
	}
	return strings.TrimSpace(string(output))
}

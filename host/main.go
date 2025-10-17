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

func (h *Host) GetEnvPath() string {
	return filepath.Join(DataDir(), "env", h.Triplet)
}

var SupportedHosts = map[string]*Host{
	"aarch64-apple-darwin": {
		Triplet: "aarch64-apple-darwin",
		Env: []string{
			"*:HOST=aarch64-apple-darwin",
			"*:TARGET=aarch64-apple-darwin",
			"*:OSX_MIN_VERSION=10.15",
			"*:LD64_VERSION=609",
			"*:CC_target=arm64-apple-darwin",
			"*:CC=" + shellOutput("xcrun -f clang") + " -target $CC_target -mmacosx-version-min=$OSX_MIN_VERSION --sysroot " + shellOutput("xcrun --sdk macosx --show-sdk-path") + " -I" + shellOutput("xcrun --sdk macosx --show-sdk-path") + "/usr/include -I$PREFIX/include",
			"*:CXX=" + shellOutput("xcrun -f clang++") + " -target $CC_target -mmacosx-version-min=$OSX_MIN_VERSION --sysroot " + shellOutput("xcrun --sdk macosx --show-sdk-path") + " -I" + shellOutput("xcrun --sdk macosx --show-sdk-path") + "/usr/include -I$PREFIX/include",
			"*:CFLAGS=",
			"*:CXXFLAGS=$CFLAGS -stdlib=libc++",
			"*:RANLIB=" + shellOutput("xcrun -f ranlib"),
			"*:AR=" + shellOutput("xcrun -f ar"),
			"*:LIBTOOL=" + shellOutput("xcrun -f libtool"),
			"*:SDK_PATH=" + shellOutput("xcrun --sdk macosx --show-sdk-path"),
		},
	},
	"x86_64-apple-darwin": {
		Triplet: "x86_64-apple-darwin",
		Env: []string{
			"*:HOST=x86_64-apple-darwin",
			"*:TARGET=x86_64-apple-darwin",
			"*:OSX_MIN_VERSION=10.15",
			"*:LD64_VERSION=609",
			"*:CC_target=x86_64-apple-darwin",
			"*:CC=" + shellOutput("xcrun -f clang") + " -target $CC_target -mmacosx-version-min=$OSX_MIN_VERSION --sysroot " + shellOutput("xcrun --sdk macosx --show-sdk-path") + " -I" + shellOutput("xcrun --sdk macosx --show-sdk-path") + "/usr/include -I$PREFIX/include",
			"*:CXX=" + shellOutput("xcrun -f clang++") + " -target $CC_target -mmacosx-version-min=$OSX_MIN_VERSION --sysroot " + shellOutput("xcrun --sdk macosx --show-sdk-path") + " -I" + shellOutput("xcrun --sdk macosx --show-sdk-path") + "/usr/include -I$PREFIX/include",
			"*:CFLAGS=",
			"*:CXXFLAGS=$CFLAGS -stdlib=libc++",
			"*:RANLIB=" + shellOutput("xcrun -f ranlib"),
			"*:AR=" + shellOutput("xcrun -f ar"),
			"*:LIBTOOL=" + shellOutput("xcrun -f libtool"),
			"*:SDK_PATH=" + shellOutput("xcrun --sdk macosx --show-sdk-path"),
		},
	},
	"aarch64-apple-ios": {
		Triplet: "aarch64-apple-ios",
		Env: []string{
			"*:HOST=aarch64-apple-ios",
			"*:TARGET=aarch64-apple-ios",
			"*:IOS_MIN_VERSION=12",
			"*:LD64_VERSION=609",
			"*:CC_target=aarch64-apple-ios",
			"*:CC=" + shellOutput("xcrun -f clang") + " -target $CC_target -mios-version-min=$IOS_MIN_VERSION --sysroot " + shellOutput("xcrun --sdk iphoneos --show-sdk-path") + " -I" + shellOutput("xcrun --sdk iphoneos --show-sdk-path") + "/usr/include -I$PREFIX/include",
			"*:CXX=" + shellOutput("xcrun -f clang++") + " -target $CC_target -mios-version-min=$IOS_MIN_VERSION --sysroot " + shellOutput("xcrun --sdk iphoneos --show-sdk-path") + " -I" + shellOutput("xcrun --sdk iphoneos --show-sdk-path") + "/usr/include -I$PREFIX/include",
			"*:CFLAGS=",
			"*:CXXFLAGS=$CFLAGS -stdlib=libc++",
			"*:RANLIB=" + shellOutput("xcrun -f ranlib"),
			"*:AR=" + shellOutput("xcrun -f ar"),
			"*:LIBTOOL=" + shellOutput("xcrun -f libtool"),
			"*:SDK_PATH=" + shellOutput("xcrun --sdk iphoneos --show-sdk-path"),
		},
	},
	"aarch64-apple-ios-simulator": {
		Triplet: "aarch64-apple-ios-simulator",
		Env: []string{
			"*:HOST=aarch64-apple-ios-simulator",
			"*:TARGET=aarch64-apple-ios-simulator",
			"*:IOS_MIN_VERSION=12",
			"*:LD64_VERSION=609",
			"*:CC_target=aarch64-apple-ios-simulator",
			"*:CC=" + shellOutput("xcrun -f clang") + " -target $CC_target -mios-version-min=$IOS_MIN_VERSION --sysroot " + shellOutput("xcrun --sdk iphonesimulator --show-sdk-path") + " -I" + shellOutput("xcrun --sdk iphonesimulator --show-sdk-path") + "/usr/include -I$PREFIX/include",
			"*:CXX=" + shellOutput("xcrun -f clang++") + " -target $CC_target -mios-version-min=$IOS_MIN_VERSION --sysroot " + shellOutput("xcrun --sdk iphonesimulator --show-sdk-path") + " -I" + shellOutput("xcrun --sdk iphonesimulator --show-sdk-path") + "/usr/include -I$PREFIX/include",
			"*:CFLAGS=",
			"*:CXXFLAGS=$CFLAGS -stdlib=libc++",
			"*:RANLIB=" + shellOutput("xcrun -f ranlib"),
			"*:AR=" + shellOutput("xcrun -f ar"),
			"*:LIBTOOL=" + shellOutput("xcrun -f libtool"),
			"*:SDK_PATH=" + shellOutput("xcrun --sdk iphonesimulator --show-sdk-path"),
		},
	},
	// "x86_64-linux-musl": {
	// 	Triplet: "x86_64-linux-musl",
	// 	Env: []string{
	// 		"*:HOST=x86_64-linux-musl",
	// 		"*:TARGET=x86_64-linux-musl",
	// 		"*:CC_target=x86_64-linux-musl",
	// 		"*:CC=x86_64-linux-musl-gcc",
	// 		"*:CXX=x86_64-linux-musl-g++",
	// 		"*:CXXFLAGS=$CFLAGS",
	// 		"*:RANLIB=x86_64-linux-musl-ranlib",
	// 		"*:AR=x86_64-linux-musl-ar",
	// 		"*:AS=x86_64-linux-musl-as",
	// 		"*:LIBTOOL=x86_64-linux-musl-libtool",
	// 		"*:OBJCOPY=x86_64-linux-musl-objcopy",
	// 		"*:OBJDUMP=x86_64-linux-musl-objdump",
	// 		"*:STRIP=x86_64-linux-musl-strip",
	// 		"*:READELF=x86_64-linux-musl-readelf",
	// 		"*:LD=x86_64-linux-musl-ld",
	// 		"*:NM=x86_64-linux-musl-nm",
	// 	},
	// },
	// "x86_64-w64-mingw32": {
	// 	Triplet: "x86_64-w64-mingw32",
	// 	Env: []string{
	// 		"*:HOST=x86_64-w64-mingw32",
	// 		"*:TARGET=x86_64-w64-mingw32",
	// 		"*:CC_target=x86_64-w64-mingw32",
	// 		"*:CC=x86_64-w64-mingw32-gcc",
	// 		"*:CXX=x86_64-w64-mingw32-g++",
	// 		"*:CXXFLAGS=$CFLAGS",
	// 		"*:RANLIB=x86_64-w64-mingw32-ranlib",
	// 		"*:AR=x86_64-w64-mingw32-ar",
	// 		"*:AS=x86_64-w64-mingw32-as",
	// 		"*:LIBTOOL=x86_64-w64-mingw32-libtool",
	// 		"*:OBJCOPY=x86_64-w64-mingw32-objcopy",
	// 		"*:OBJDUMP=x86_64-w64-mingw32-objdump",
	// 		"*:STRIP=x86_64-w64-mingw32-strip",
	// 		"*:READELF=x86_64-w64-mingw32-readelf",
	// 		"*:LD=x86_64-w64-mingw32-ld",
	// 		"*:NM=x86_64-w64-mingw32-nm",
	// 	},
	// },
	"x86_64-linux-gnu": {
		Triplet: "x86_64-linux-gnu",
		Env: []string{
			"*:HOST=x86_64-linux-gnu",
			"*:TARGET=x86_64-linux-gnu",
			"*:CC_target=x86_64-linux-gnu",
			"*:CC=x86_64-unknown-linux-gnu-gcc",
			"*:CXX=x86_64-unknown-linux-gnu-g++",
			"*:CXXFLAGS=$CFLAGS",
			"*:RANLIB=x86_64-unknown-linux-gnu-ranlib",
			"*:AR=x86_64-unknown-linux-gnu-ar",
			"*:AS=x86_64-unknown-linux-gnu-as",
			"*:LIBTOOL=x86_64-unknown-linux-gnu-libtool",
			"*:OBJCOPY=x86_64-unknown-linux-gnu-objcopy",
			"*:OBJDUMP=x86_64-unknown-linux-gnu-objdump",
			"*:STRIP=x86_64-unknown-linux-gnu-strip",
			"*:READELF=x86_64-unknown-linux-gnu-readelf",
			"*:LD=x86_64-unknown-linux-gnu-ld",
			"*:NM=x86_64-unknown-linux-gnu-nm",
		},
	},
	"aarch64-linux-gnu": {
		Triplet: "aarch64-linux-gnu",
		Env: []string{
			"*:HOST=aarch64-linux-gnu",
			"*:TARGET=aarch64-linux-gnu",
			"*:CC_target=aarch64-linux-gnu",
			"*:CC=aarch64-unknown-linux-gnu-gcc",
			"*:CXX=aarch64-unknown-linux-gnu-g++",
			"*:CXXFLAGS=$CFLAGS",
			"*:RANLIB=aarch64-unknown-linux-gnu-ranlib",
			"*:AR=aarch64-unknown-linux-gnu-ar",
			"*:AS=aarch64-unknown-linux-gnu-as",
			"*:LIBTOOL=aarch64-unknown-linux-gnu-libtool",
			"*:OBJCOPY=aarch64-unknown-linux-gnu-objcopy",
			"*:OBJDUMP=aarch64-unknown-linux-gnu-objdump",
			"*:STRIP=aarch64-unknown-linux-gnu-strip",
			"*:READELF=aarch64-unknown-linux-gnu-readelf",
			"*:LD=aarch64-unknown-linux-gnu-ld",
			"*:NM=aarch64-unknown-linux-gnu-nm",
		},
	},
	"aarch64-linux-android": {
		Triplet: "aarch64-linux-android",
		Env: []string{
			"*:HOST=aarch64-linux-android",
			"*:TARGET=aarch64-linux-android",
			"*:CC_target=aarch64-linux-android",
			"*:CC=aarch64-linux-android21-clang",
			"*:CXX=aarch64-linux-android21-clang++",
			"*:CFLAGS=",
			"*:CXXFLAGS=$CFLAGS -stdlib=libc++",
			"*:RANLIB=llvm-ranlib",
			"*:AR=llvm-ar",
			"*:AS=llvm-as",
			"*:LIBTOOL=libtool",
			"*:ANDROID_NDK_HOME=$PREFIX/native/",
			"*:LDFLAGS=$PREFIX/usr/lib/libc++_static.a $PREFIX/usr/lib/libc++abi.a -lm -lc -llog",
		},
	},
	"x86_64-linux-android": {
		Triplet: "x86_64-linux-android",
		Env: []string{
			"*:HOST=x86_64-linux-android",
			"*:TARGET=x86_64-linux-android",
			"*:CC_target=x86_64-linux-android",
			"*:CC=x86_64-linux-android21-clang",
			"*:CXX=x86_64-linux-android21-clang++",
			"*:CFLAGS=",
			"*:CXXFLAGS=$CFLAGS",
			"*:RANLIB=llvm-ranlib",
			"*:AR=llvm-ar",
			"*:LIBTOOL=llvm-libtool",
			"*:ANDROID_NDK_HOME=$PREFIX/native/",
		},
	},
	"armv7a-linux-androideabi": {
		Triplet: "armv7a-linux-androideabi",
		Env: []string{
			"*:HOST=armv7a-linux-androideabi",
			"*:TARGET=armv7a-linux-androideabi",
			"*:CC_target=armv7a-linux-androideabi",
			"*:CC=armv7a-linux-androideabi21-clang",
			"*:CXX=armv7a-linux-androideabi21-clang++",
			"*:CFLAGS=",
			"*:CXXFLAGS=$CFLAGS",
			"*:RANLIB=llvm-ranlib",
			"*:AR=llvm-ar",
			"*:LIBTOOL=llvm-libtool",
			"*:ANDROID_NDK_HOME=$PREFIX/native/",
		},
	},
}

func shellOutput(cmd string) string {
	output, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return "$(" + cmd + ")"
	}
	return strings.TrimSpace(string(output))
}

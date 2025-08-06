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
			"all:HOST=aarch64-apple-darwin",
			"all:TARGET=aarch64-apple-darwin",
			"all:OSX_MIN_VERSION=10.15",
			"all:LD64_VERSION=609",
			"all:CC_target=arm64-apple-darwin",
			"all:CC=" + shellOutput("xcrun -f clang") + " -target $CC_target -mmacosx-version-min=$OSX_MIN_VERSION --sysroot " + shellOutput("xcrun --sdk macosx --show-sdk-path") + " -I" + shellOutput("xcrun --sdk macosx --show-sdk-path") + "/usr/include -I$PREFIX/include",
			"all:CXX=" + shellOutput("xcrun -f clang++") + " -target $CC_target -mmacosx-version-min=$OSX_MIN_VERSION --sysroot " + shellOutput("xcrun --sdk macosx --show-sdk-path") + " -I" + shellOutput("xcrun --sdk macosx --show-sdk-path") + "/usr/include -I$PREFIX/include",
			"all:CFLAGS=",
			"all:CXXFLAGS=$CFLAGS -stdlib=libc++",
			"all:RANLIB=" + shellOutput("xcrun -f ranlib"),
			"all:AR=" + shellOutput("xcrun -f ar"),
			"all:LIBTOOL=" + shellOutput("xcrun -f libtool"),
			"all:SDK_PATH=" + shellOutput("xcrun --sdk macosx --show-sdk-path"),
		},
	},
	"x86_64-apple-darwin": {
		Triplet: "x86_64-apple-darwin",
		Env: []string{
			"all:HOST=x86_64-apple-darwin",
			"all:TARGET=x86_64-apple-darwin",
			"all:OSX_MIN_VERSION=10.15",
			"all:LD64_VERSION=609",
			"all:CC_target=x86_64-apple-darwin",
			"all:CC=" + shellOutput("xcrun -f clang") + " -target $CC_target -mmacosx-version-min=$OSX_MIN_VERSION --sysroot " + shellOutput("xcrun --sdk macosx --show-sdk-path") + " -I" + shellOutput("xcrun --sdk macosx --show-sdk-path") + "/usr/include -I$PREFIX/include",
			"all:CXX=" + shellOutput("xcrun -f clang++") + " -target $CC_target -mmacosx-version-min=$OSX_MIN_VERSION --sysroot " + shellOutput("xcrun --sdk macosx --show-sdk-path") + " -I" + shellOutput("xcrun --sdk macosx --show-sdk-path") + "/usr/include -I$PREFIX/include",
			"all:CFLAGS=",
			"all:CXXFLAGS=$CFLAGS -stdlib=libc++",
			"all:RANLIB=" + shellOutput("xcrun -f ranlib"),
			"all:AR=" + shellOutput("xcrun -f ar"),
			"all:LIBTOOL=" + shellOutput("xcrun -f libtool"),
			"all:SDK_PATH=" + shellOutput("xcrun --sdk macosx --show-sdk-path"),
		},
	},
	"aarch64-apple-ios": {
		Triplet: "aarch64-apple-ios",
		Env: []string{
			"all:HOST=aarch64-apple-ios",
			"all:TARGET=aarch64-apple-ios",
			"all:IOS_MIN_VERSION=12",
			"all:LD64_VERSION=609",
			"all:CC_target=aarch64-apple-ios",
			"all:CC=" + shellOutput("xcrun -f clang") + " -target $CC_target -mios-version-min=$IOS_MIN_VERSION --sysroot " + shellOutput("xcrun --sdk iphoneos --show-sdk-path") + " -I" + shellOutput("xcrun --sdk iphoneos --show-sdk-path") + "/usr/include -I$PREFIX/include",
			"all:CXX=" + shellOutput("xcrun -f clang++") + " -target $CC_target -mios-version-min=$IOS_MIN_VERSION --sysroot " + shellOutput("xcrun --sdk iphoneos --show-sdk-path") + " -I" + shellOutput("xcrun --sdk iphoneos --show-sdk-path") + "/usr/include -I$PREFIX/include",
			"all:CFLAGS=",
			"all:CXXFLAGS=$CFLAGS -stdlib=libc++",
			"all:RANLIB=" + shellOutput("xcrun -f ranlib"),
			"all:AR=" + shellOutput("xcrun -f ar"),
			"all:LIBTOOL=" + shellOutput("xcrun -f libtool"),
			"all:SDK_PATH=" + shellOutput("xcrun --sdk iphoneos --show-sdk-path"),
		},
	},
	"aarch64-apple-ios-simulator": {
		Triplet: "aarch64-apple-ios-simulator",
		Env: []string{
			"all:HOST=aarch64-apple-ios-simulator",
			"all:TARGET=aarch64-apple-ios-simulator",
			"all:IOS_MIN_VERSION=12",
			"all:LD64_VERSION=609",
			"all:CC_target=aarch64-apple-ios-simulator",
			"all:CC=" + shellOutput("xcrun -f clang") + " -target $CC_target -mios-version-min=$IOS_MIN_VERSION --sysroot " + shellOutput("xcrun --sdk iphonesimulator --show-sdk-path") + " -I" + shellOutput("xcrun --sdk iphonesimulator --show-sdk-path") + "/usr/include -I$PREFIX/include",
			"all:CXX=" + shellOutput("xcrun -f clang++") + " -target $CC_target -mios-version-min=$IOS_MIN_VERSION --sysroot " + shellOutput("xcrun --sdk iphonesimulator --show-sdk-path") + " -I" + shellOutput("xcrun --sdk iphonesimulator --show-sdk-path") + "/usr/include -I$PREFIX/include",
			"all:CFLAGS=",
			"all:CXXFLAGS=$CFLAGS -stdlib=libc++",
			"all:RANLIB=" + shellOutput("xcrun -f ranlib"),
			"all:AR=" + shellOutput("xcrun -f ar"),
			"all:LIBTOOL=" + shellOutput("xcrun -f libtool"),
			"all:SDK_PATH=" + shellOutput("xcrun --sdk iphonesimulator --show-sdk-path"),
		},
	},
	"x86_64-linux-gnu": {
		Triplet: "x86_64-linux-gnu",
		Env: []string{
			"all:HOST=x86_64-linux-gnu",
			"all:TARGET=x86_64-linux-gnu",
			"all:CC_target=x86_64-linux-gnu",
			"all:CC=x86_64-linux-gnu-gcc",
			"all:CXX=x86_64-linux-gnu-g++",
			"all:CXXFLAGS=$CFLAGS",
			"all:RANLIB=x86_64-linux-gnu-ranlib",
			"all:AR=x86_64-linux-gnu-ar",
			"all:AS=x86_64-linux-gnu-as",
			"all:LIBTOOL=x86_64-linux-gnu-libtool",
			"all:OBJCOPY=x86_64-linux-gnu-objcopy",
			"all:OBJDUMP=x86_64-linux-gnu-objdump",
			"all:STRIP=x86_64-linux-gnu-strip",
			"all:READELF=x86_64-linux-gnu-readelf",
			"all:LD=x86_64-linux-gnu-ld",
			"all:NM=x86_64-linux-gnu-nm",
		},
	},
	"aarch64-linux-gnu": {
		Triplet: "aarch64-linux-gnu",
		Env: []string{
			"all:HOST=aarch64-linux-gnu",
			"all:TARGET=aarch64-linux-gnu",
			"all:CC_target=aarch64-linux-gnu",
			"all:CC=aarch64-linux-gnu-gcc",
			"all:CXX=aarch64-linux-gnu-g++",
			"all:CFLAGS=",
			"all:CXXFLAGS=$CFLAGS",
			"all:RANLIB=aarch64-linux-gnu-ranlib",
			"all:AR=aarch64-linux-gnu-ar",
			"all:LIBTOOL=aarch64-linux-gnu-libtool",
		},
	},
	"aarch64-linux-android": {
		Triplet: "aarch64-linux-android",
		Env: []string{
			"all:HOST=aarch64-linux-android",
			"all:TARGET=aarch64-linux-android",
			"all:CC_target=aarch64-linux-android",
			"all:CC=aarch64-linux-android21-clang",
			"all:CXX=aarch64-linux-android21-clang++",
			"all:CFLAGS=",
			"all:CXXFLAGS=$CFLAGS",
			"all:RANLIB=llvm-ranlib",
			"all:AR=llvm-ar",
			"all:AS=llvm-as",
			"all:LIBTOOL=libtool",
			"all:ANDROID_NDK_HOME=$PREFIX/native/",
		},
	},
	"x86_64-linux-android": {
		Triplet: "x86_64-linux-android",
		Env: []string{
			"all:HOST=x86_64-linux-android",
			"all:TARGET=x86_64-linux-android",
			"all:CC_target=x86_64-linux-android",
			"all:CC=x86_64-linux-android21-clang",
			"all:CXX=x86_64-linux-android21-clang++",
			"all:CFLAGS=",
			"all:CXXFLAGS=$CFLAGS",
			"all:RANLIB=llvm-ranlib",
			"all:AR=llvm-ar",
			"all:LIBTOOL=llvm-libtool",
			"all:ANDROID_NDK_HOME=$PREFIX/native/",
		},
	},
	"armv7a-linux-androideabi": {
		Triplet: "armv7a-linux-androideabi",
		Env: []string{
			"all:HOST=armv7a-linux-androideabi",
			"all:TARGET=armv7a-linux-androideabi",
			"all:CC_target=armv7a-linux-androideabi",
			"all:CC=armv7a-linux-androideabi21-clang",
			"all:CXX=armv7a-linux-androideabi21-clang++",
			"all:CFLAGS=",
			"all:CXXFLAGS=$CFLAGS",
			"all:RANLIB=llvm-ranlib",
			"all:AR=llvm-ar",
			"all:LIBTOOL=llvm-libtool",
			"all:ANDROID_NDK_HOME=$PREFIX/native/",
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

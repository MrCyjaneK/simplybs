package builder

var HostBuilder = Builder{
	GlobalEnv: []string{
		"all:HOST=aarch64-apple-darwin",
		"all:TARGET=aarch64-apple-darwin",
		"all:OSX_MIN_VERSION=10.15",
		"all:CC=" + shellOutput("xcrun -f clang") + " -target aarch64-apple-darwin -mmacosx-version-min=$OSX_MIN_VERSION --sysroot " + shellOutput("xcrun --sdk macosx --show-sdk-path") + " -I" + shellOutput("xcrun --sdk macosx --show-sdk-path") + "/usr/include -I$PREFIX/include",
		"all:CXX=" + shellOutput("xcrun -f clang++") + " -target aarch64-apple-darwin -mmacosx-version-min=$OSX_MIN_VERSION --sysroot " + shellOutput("xcrun --sdk macosx --show-sdk-path") + " -I" + shellOutput("xcrun --sdk macosx --show-sdk-path") + "/usr/include -I$PREFIX/include",
		"all:AR=" + shellOutput("xcrun --sdk macosx --find ar"),
		"all:RANLIB=" + shellOutput("xcrun --sdk macosx --find ranlib"),
		"all:STRIP=" + shellOutput("xcrun --sdk macosx --find strip"),
		"all:NM=" + shellOutput("xcrun --sdk macosx --find nm"),
		"all:OTOOL=" + shellOutput("xcrun --sdk macosx --find otool"),
		"all:INSTALL_NAME_TOOL=" + shellOutput("xcrun --sdk macosx --find install_name_tool"),
	},
}

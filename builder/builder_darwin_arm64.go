package builder

var HostBuilder = Builder{
	GlobalEnv: []string{
		"all:HOST=aarch64-apple-darwin",
		"all:TARGET=aarch64-apple-darwin",
		"all:OSX_MIN_VERSION=13",
		"all:CC=/usr/bin/clang",
		"all:CXX=/usr/bin/clang++",
		"all:AR=/usr/bin/ar",
		"all:RANLIB=/usr/bin/ranlib",
		"all:STRIP=/usr/bin/strip",
		"all:NM=/usr/bin/nm",
		"all:OTOOL=/usr/bin/otool",
		"all:INSTALL_NAME_TOOL=/usr/bin/install_name_tool",
	},
}

package builder

import "runtime"

var HostBuilder = Builder{
	GlobalEnv: []string{
		"*:*:HOST=aarch64-apple-darwin",
		"*:*:TARGET=aarch64-apple-darwin",
		"*:*:OSX_MIN_VERSION=13",
		"*:*:CC=/usr/bin/clang",
		"*:*:CXX=/usr/bin/clang++",
		"*:*:AR=/usr/bin/ar",
		"*:*:RANLIB=/usr/bin/ranlib",
		"*:*:STRIP=/usr/bin/strip",
		"*:*:NM=/usr/bin/nm",
		"*:*:OTOOL=/usr/bin/otool",
		"*:*:INSTALL_NAME_TOOL=/usr/bin/install_name_tool",
		"*:*:BUILDER_GOOS=" + runtime.GOOS,
		"*:*:BUILDER_GOARCH=" + runtime.GOARCH,
	},
}

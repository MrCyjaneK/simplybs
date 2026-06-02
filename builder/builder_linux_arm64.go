package builder

import "runtime"

var HostBuilder = Builder{
	GlobalEnv: []string{
		"*:*:HOST=aarch64-linux-gnu",
		"*:*:TARGET=aarch64-linux-gnu",
		"*:*:CC=/usr/bin/clang",
		"*:*:CXX=/usr/bin/clang++",
		"*:*:AR=/usr/bin/ar",
		"*:*:RANLIB=/usr/bin/ranlib",
		"*:*:INSTALL_NAME_TOOL=install_name_tool",
		"*:*:BUILDER_GOOS=" + runtime.GOOS,
		"*:*:BUILDER_GOARCH=" + runtime.GOARCH,
	},
}

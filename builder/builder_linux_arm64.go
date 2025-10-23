package builder

import "runtime"

var HostBuilder = Builder{
	GlobalEnv: []string{
		"*:HOST=aarch64-linux-gnu",
		"*:TARGET=aarch64-linux-gnu",
		"*:CC=aarch64-linux-gnu-gcc",
		"*:CXX=aarch64-linux-gnu-g++",
		"*:AR=ar",
		"*:RANLIB=ranlib",
		"*:STRIP=strip",
		"*:NM=nm",
		"*:OTOOL=otool",
		"*:AUTOMAKE=automake",
		"*:INSTALL_NAME_TOOL=install_name_tool",
		"*:BUILDER_GOOS=" + runtime.GOOS,
		"*:BUILDER_GOARCH=" + runtime.GOARCH,
	},
}

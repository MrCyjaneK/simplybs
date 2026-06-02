package builder

import "runtime"

var HostBuilder = Builder{
	GlobalEnv: []string{
		"*:*:HOST=x86_64-linux-gnu",
		"*:*:TARGET=x86_64-linux-gnu",
		"*:*:CC=/usr/bin/clang",
		"*:*:CXX=/usr/bin/clang++",
		"*:*:AR=/usr/bin/ar",
		"*:*:RANLIB=/usr/bin/ranlib",
		"*:*:BUILDER_GOOS=" + runtime.GOOS,
		"*:*:BUILDER_GOARCH=" + runtime.GOARCH,
	},
}

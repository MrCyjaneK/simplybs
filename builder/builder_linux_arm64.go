package builder

var HostBuilder = Builder{
	GlobalEnv: []string{
		"*:HOST=aarch64-linux-gnu",
		"*:TARGET=aarch64-linux-gnu",
		"*:CC=clang",
		"*:CXX=clang++",
		"*:AR=ar",
		"*:RANLIB=ranlib",
		"*:STRIP=strip",
		"*:NM=nm",
		"*:OTOOL=otool",
		"*:AUTOMAKE=automake",
		"*:INSTALL_NAME_TOOL=install_name_tool",
	},
}

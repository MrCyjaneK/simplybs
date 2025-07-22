package main

var hostBuilder = Builder{
	GlobalEnv: []string{
		"all:HOST=aarch64-apple-darwin",
		"all:TARGET=aarch64-apple-darwin",
		"all:CC=clang",
		"all:CXX=clang++",
		"all:AR=ar",
		"all:RANLIB=ranlib",
		"all:STRIP=strip",
		"all:NM=nm",
		"all:OTOOL=otool",
		"all:INSTALL_NAME_TOOL=install_name_tool",
	},
}

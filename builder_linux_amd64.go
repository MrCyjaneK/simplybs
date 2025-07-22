package main

var hostBuilder = Builder{
	GlobalEnv: []string{
		"all:HOST=x86_64-linux-gnu",
		"all:TARGET=x86_64-linux-gnu",
		"all:CC=clang",
		"all:CXX=clang++",
		"all:AR=ar",
		"all:RANLIB=ranlib",
		"all:STRIP=strip",
		"all:NM=nm",
		"all:OTOOL=otool",
		"all:AUTOMAKE=automake",
		"all:INSTALL_NAME_TOOL=install_name_tool",
	},
}

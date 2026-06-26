package builder

import "runtime"

var HostBuilder = Builder{
	GlobalEnv: []string{
		"*:*:HOST=aarch64-linux-gnu",
		"*:*:TARGET=aarch64-linux-gnu",
		"*:*:CC=$NATIVEPREFIX/_/bin/clang",
		"*:*:CXX=$NATIVEPREFIX/_/bin/clang++",
		"*:*:AR=$NATIVEPREFIX/_/bin/ar",
		"*:*:RANLIB=$NATIVEPREFIX/_/bin/ranlib",
		"*:*:STRIP=$NATIVEPREFIX/_/bin/strip",
		"*:*:NM=$NATIVEPREFIX/_/bin/nm",
		"*:*:BUILDER_GOOS=" + runtime.GOOS,
		"*:*:BUILDER_GOARCH=" + runtime.GOARCH,
		// lt_cv_prog_gnu_ld=no
		// helps with resolving ld64.lld pretending to be GNU on while it's actually not, this occured in libtool and pkgconf
		// so I assume this is better here as a general fix.
		// ld64.lld: error: unknown argument '-soname'
		// ld64.lld: error: unknown argument '-retain-symbols-file'
		// "*:*:lt_cv_prog_gnu_ld=no",
	},
}

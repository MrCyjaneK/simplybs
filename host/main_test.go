package host_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mrcyjanek/simplybs/host"
	"github.com/mrcyjanek/simplybs/utils/ifstring"
)

func TestHosts(t *testing.T) {
	for i := range host.SupportedHosts {
		t.Run("Env vars: "+i, func(t *testing.T) {
			assertEnv(t, i)
		})
	}
}

func assertEnv(t *testing.T, i string) {
	env := host.SupportedHosts[i].Env
	keys := []string{}
	for _, envVar := range env {
		parsed := ifstring.ParseIfString(envVar)
		key := strings.Split(parsed.Content, "=")[0]
		keys = append(keys, key)
	}
	var required = []string{
		"TARGET",
		"CC",
		"CXX",
		"CFLAGS",
		"AR",
		"RANLIB",
		"NM",
		"STRIP",
		"CXXFLAGS",
		"CPPFLAGS",
		"LDFLAGS",
		"CMAKE_SYSTEM_NAME",
		"ARCH",
	}
	for _, req := range required {
		exists := false
		for _, has := range keys {
			if has == req {
				exists = true
			}
		}
		if exists == false {
			fmt.Println(i, "doesn't have", req)
			t.Fail()
		}
	}
	// cat | sed -e "s|@HOST@|$(host)|g" \ <-- $TARGET
	//     -e "s|@CC@|$(host_CC)|g" \ <--- $CC
	//     -e "s|@CXX@|$(host_CXX)|g" \ <--- $CXX
	//     -e "s|@AR@|$(toolchain_path)$(host_AR)|g" \ <--- $AR
	//     -e "s|@RANLIB@|$(toolchain_path)$(host_RANLIB)|g" \ <--- $RANLIB
	//     -e "s|@NM@|$(toolchain_path)$(host_NM)|g" \ <--- $NM
	//     -e "s|@STRIP@|$(toolchain_path)$(host_STRIP)|g" \ <--- $STRIP
	//     -e "s|@CFLAGS@|$(strip $(host_CFLAGS) $(host_$(release_type)_CFLAGS))|g" \ <--- $CFLAGS
	//     -e "s|@CXXFLAGS@|$(strip $(host_CXXFLAGS) $(host_$(release_type)_CXXFLAGS))|g" \ <-- $CXXFLAGS
	//     -e "s|@CPPFLAGS@|$(strip $(host_CPPFLAGS) $(host_$(release_type)_CPPFLAGS))|g" \ <-- $CPPFLAGS
	//     -e "s|@LDFLAGS@|$(strip $(host_LDFLAGS) $(host_$(release_type)_LDFLAGS))|g" \ <-- $LDFLAGS
	//     -e "s|@cmake_system_name@|$($(host_os)_cmake_system)|g" \ $CMAKE_SYSTEM_NAME
	//     -e "s|@prefix@|$($(host_arch)_$(host_os)_prefix)|g"\ ---> $PREFIX
	//     -e "s|@arch@|$(host_arch)|g"\ ---> aarch64/arm64/x86_64
	// > $outfile <<EOF
}

package host_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mrcyjanek/simplybs/builder"
	"github.com/mrcyjanek/simplybs/host"
	"github.com/mrcyjanek/simplybs/pack"
	"github.com/mrcyjanek/simplybs/utils/ifstring"
)

func chdirRepoRoot(t *testing.T) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "packages", "native", "_.json")); err == nil {
			if err := os.Chdir(wd); err != nil {
				t.Fatal(err)
			}
			return
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatal("could not find repo root (packages/native/_.json)")
		}
		wd = parent
	}
}

func TestHosts(t *testing.T) {
	chdirRepoRoot(t)
	meta, err := pack.FindPackage("native/_")
	if err != nil {
		t.Fatal(err)
	}
	for triplet := range host.SupportedHosts {
		t.Run("Env vars: "+triplet, func(t *testing.T) {
			assertEnv(t, triplet, meta.ExportEnv)
		})
	}
}

func assertEnv(t *testing.T, triplet string, exportEnv []string) {
	keys := []string{}
	for _, envVar := range exportEnv {
		is := ifstring.ParseIfString(envVar)
		if !is.Matches(triplet, builder.GetName()) {
			continue
		}
		key := strings.Split(is.Content, "=")[0]
		keys = append(keys, key)
	}
	var required = []string{
		// HOST/TARGET are injected by pack.ExportEnv, not package JSON.
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
		if !exists {
			fmt.Println(triplet, "doesn't have", req)
			t.Fail()
		}
	}
}

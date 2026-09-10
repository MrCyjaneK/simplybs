package host

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/mrcyjanek/simplybs/crash"
)

type Host struct {
	Triplet string
}

func GetPackagesDir() string {
	if os.Getenv("SIMPLYBS_PACKAGES_DIR") != "" {
		return os.Getenv("SIMPLYBS_PACKAGES_DIR")
	}
	if os.Getenv("SIMPLYBS_DATA_DIR") != "" {
		guessPath := filepath.Join(os.Getenv("SIMPLYBS_DATA_DIR"), "..", "packages")
		if _, err := os.Stat(guessPath); err == nil {
			return guessPath
		}
	}
	wd, err := os.Getwd()
	crash.Handle(err)
	return filepath.Join(wd, "packages")
}

func DataDirRoot() string {
	if os.Getenv("SIMPLYBS_DATA_DIR") != "" {
		return os.Getenv("SIMPLYBS_DATA_DIR")
	}
	buildDir, err := os.Getwd()
	crash.Handle(err)
	return filepath.Join(buildDir, ".buildlib")
}

func DataDir() string {
	return filepath.Join(DataDirRoot(), runtime.GOOS+"_"+runtime.GOARCH)
}

func (h *Host) GetNativeEnvPath() string {
	if os.Getenv("SIMPLYBS_NATIVE_ENV_DIR") != "" {
		return filepath.Join(os.Getenv("SIMPLYBS_NATIVE_ENV_DIR"))
	}
	dir := DataDirRoot()
	return filepath.Join(dir, "env-native")
}

func (h *Host) GetEnvPath() string {
	if os.Getenv("SIMPLYBS_ENV_DIR") != "" {
		return filepath.Join(os.Getenv("SIMPLYBS_ENV_DIR"))
	}
	dir := DataDirRoot()
	return filepath.Join(dir, "env")
}

// aarch64-apple-darwin,x86_64-apple-darwin,aarch64-apple-ios,aarch64-apple-ios-simulator,x86_64-w64-mingw32,x86_64-linux-gnu,aarch64-linux-gnu,aarch64-linux-android,x86_64-linux-android,armv7a-linux-androideabi

var SupportedHosts = map[string]*Host{
	"aarch64-apple-darwin":        {Triplet: "aarch64-apple-darwin"},
	"x86_64-apple-darwin":         {Triplet: "x86_64-apple-darwin"},
	"aarch64-apple-ios":           {Triplet: "aarch64-apple-ios"},
	"aarch64-apple-ios-simulator": {Triplet: "aarch64-apple-ios-simulator"},
	"x86_64-w64-mingw32":          {Triplet: "x86_64-w64-mingw32"},
	"x86_64-linux-gnu":            {Triplet: "x86_64-linux-gnu"},
	"aarch64-linux-gnu":           {Triplet: "aarch64-linux-gnu"},
	"aarch64-linux-android":       {Triplet: "aarch64-linux-android"},
	"x86_64-linux-android":        {Triplet: "x86_64-linux-android"},
	"armv7a-linux-androideabi":    {Triplet: "armv7a-linux-androideabi"},
}

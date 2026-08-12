package builder

import (
	"log"
	"runtime"
)

func GetName() string {
	return runtime.GOOS + "_" + runtime.GOARCH
}

func NativeTriplet() string {
	switch GetName() {
	case "darwin_arm64":
		return "aarch64-apple-darwin"
	case "linux_amd64":
		return "x86_64-linux-gnu"
	case "linux_arm64":
		return "aarch64-linux-gnu"
	case "windows_amd64":
		return "x86_64-pc-cygwin"
	default:
		log.Fatalf("unknown builder: %s", GetName())
		return ""
	}
}

var Builders = []string{"darwin_arm64", "linux_amd64", "linux_arm64", "windows_amd64"}

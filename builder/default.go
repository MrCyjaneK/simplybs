package builder

import "runtime"

func GetName() string {
	return runtime.GOOS + "_" + runtime.GOARCH
}

var Builders = []string{"darwin_arm64", "linux_amd64", "linux_arm64"}

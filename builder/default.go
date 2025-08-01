package builder

type Builder struct {
	GlobalEnv []string
}

var Builders = []string{"darwin_arm64", "linux_amd64", "linux_arm64"}

package builder

import (
	"os/exec"
	"strings"
)

type Builder struct {
	GlobalEnv []string
}

var Builders = []string{"darwin_arm64", "linux_amd64", "linux_arm64"}

func shellOutput(cmd string) string {
	output, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return "$(" + cmd + ")"
	}
	return strings.TrimSpace(string(output))
}

func (b *Builder) GetCC() string {
	for _, env := range b.GlobalEnv {
		splitIndex := strings.Index(env, ":")
		envVar := env[splitIndex+1:]
		if strings.HasPrefix(envVar, "CC=") {
			return envVar[3:]
		}
	}
	return ""
}

func (b *Builder) GetCXX() string {
	for _, env := range b.GlobalEnv {
		splitIndex := strings.Index(env, ":")
		envVar := env[splitIndex+1:]
		if strings.HasPrefix(envVar, "CXX=") {
			return envVar[4:]
		}
	}
	return ""
}

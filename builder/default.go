package builder

import (
	"os/exec"
	"runtime"
	"strings"

	"github.com/mrcyjanek/simplybs/utils"
)

func GetName() string {
	return runtime.GOOS + "_" + runtime.GOARCH
}

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
		is := utils.ParseIfString(env)
		if strings.HasPrefix(is.Content, "CC=") {
			return is.Content[3:]
		}
	}
	return ""
}

func (b *Builder) GetCXX() string {
	for _, env := range b.GlobalEnv {
		is := utils.ParseIfString(env)
		if strings.HasPrefix(is.Content, "CXX=") {
			return is.Content[4:]
		}
	}
	return ""
}

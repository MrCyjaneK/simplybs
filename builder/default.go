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

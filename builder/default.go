package builder

import (
	"runtime"
	"strings"

	"github.com/mrcyjanek/simplybs/utils/ifstring"
)

func GetName() string {
	return runtime.GOOS + "_" + runtime.GOARCH
}

type Builder struct {
	GlobalEnv []string
}

var Builders = []string{"darwin_arm64", "linux_amd64", "linux_arm64"}

func (b *Builder) getToolFromEnv(prefix string) string {
	for _, env := range b.GlobalEnv {
		is := ifstring.ParseIfString(env)
		if strings.HasPrefix(is.Content, prefix) {
			return is.Content[len(prefix):]
		}
	}
	return ""
}

func (b *Builder) GetCC() string {
	return b.getToolFromEnv("CC=")
}

func (b *Builder) GetCXX() string {
	return b.getToolFromEnv("CXX=")
}

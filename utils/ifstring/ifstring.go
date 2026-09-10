package ifstring

import (
	"log"
	"strings"

	"github.com/gobwas/glob"
)

type IfString struct {
	Builder string
	Host    string
	Content string
}

func (is *IfString) BuilderGlob() glob.Glob {
	return glob.MustCompile(is.Builder)
}
func (is *IfString) HostGlob() glob.Glob {
	return glob.MustCompile(is.Host)
}

func (is *IfString) Matches(hostTriplet, builderName string) bool {
	return is.HostGlob().Match(hostTriplet) && is.BuilderGlob().Match(builderName)
}

func FilterContent(items []string, hostTriplet, builderName string) []string {
	filtered := make([]string, 0, len(items))
	for _, item := range items {
		is := ParseIfString(item)
		if !is.Matches(hostTriplet, builderName) {
			continue
		}
		filtered = append(filtered, is.Content)
	}
	return filtered
}

func ParseIfString(str string) *IfString {
	is := &IfString{}
	is.Builder = strings.Split(str, ":")[0]
	is.Host = strings.Split(str, ":")[1]
	offset := len(is.Builder) + len(is.Host) + 2
	if len(str) < offset {
		log.Fatalln("failed to parse:", str)
	}
	is.Content = str[offset:]
	return is
}

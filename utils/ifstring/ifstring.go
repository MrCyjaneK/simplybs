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

package utils

import (
	"log"
	"os"
	"strings"

	"github.com/mrcyjanek/simplybs/builder"
	"github.com/mrcyjanek/simplybs/host"
	"github.com/mrcyjanek/simplybs/utils/ifstring"
)

func ExpandEnvFromMap(s string, envMap map[string]string) string {
	ret := os.Expand(s, func(key string) string {
		if val, ok := envMap[key]; ok {
			return val
		}
		return ""
	})
	return ret
}

func AppendEnv(env map[string]string, newEnv []string, host *host.Host) map[string]string {
	for _, envVar := range newEnv {
		is := ifstring.ParseIfString(envVar)
		is.Content = ExpandEnvFromMap(is.Content, env)
		equalIndex := strings.Index(is.Content, "=")
		if equalIndex == -1 {
			log.Fatalf("Invalid env var: %s. Vars needs to be in the form of *:KEY=VALUE", envVar)
		}
		if !is.Matches(host.Triplet, builder.GetName()) {
			continue
		}
		k := is.Content[:equalIndex]
		v := is.Content[equalIndex+1:]
		env[k] = v
	}
	return env
}

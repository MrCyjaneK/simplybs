package utils

import (
	"log"
	"os"
	"strings"

	"github.com/mrcyjanek/simplybs/host"
	"github.com/ryanuber/go-glob"
)

func ExpandEnvFromMap(s string, envMap map[string]string) string {
	return os.Expand(s, func(key string) string {
		if val, ok := envMap[key]; ok {
			return val
		}
		return ""
	})
}

func AppendEnv(env map[string]string, newEnv []string, host *host.Host) map[string]string {
	for _, envVar := range newEnv {
		exp := ExpandEnvFromMap(envVar, env)
		colonIndex := strings.Index(exp, ":")
		equalIndex := strings.Index(exp, "=")
		if colonIndex == -1 || equalIndex == -1 {
			log.Fatalf("Invalid env var: %s. Vars needs to be in the form of all:KEY=VALUE", exp)
		}
		prefix := exp[0:colonIndex]
		if !glob.Glob(prefix, host.Triplet) && prefix != "all" {
			continue
		}
		k := exp[colonIndex+1 : equalIndex]
		v := exp[equalIndex+1:]
		env[k] = v
	}
	return env
}

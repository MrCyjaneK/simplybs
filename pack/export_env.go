package pack

import (
	"log"
	"strings"

	"github.com/mrcyjanek/simplybs/builder"
	"github.com/mrcyjanek/simplybs/host"
	"github.com/mrcyjanek/simplybs/utils"
	"github.com/mrcyjanek/simplybs/utils/ifstring"
)

type EnvKV struct {
	K string
	V string
}

type exportEnvContext struct {
	memo     map[string][]EnvKV
	visiting map[string]bool
}

func newExportEnvContext() *exportEnvContext {
	return &exportEnvContext{
		memo:     make(map[string][]EnvKV),
		visiting: make(map[string]bool),
	}
}

func copyEnv(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func envKeyFromLine(line string) string {
	idx := strings.Index(line, "=")
	if idx == -1 {
		return ""
	}
	return line[:idx]
}

func (ctx *exportEnvContext) resolvedExportEnv(p *Package, h *host.Host) []EnvKV {
	key := p.Package + "\x00" + h.Triplet
	if ctx.visiting[key] {
		log.Fatalf("export-env cycle involving package %s", p.Package)
	}
	if cached, ok := ctx.memo[key]; ok {
		return cached
	}
	ctx.visiting[key] = true
	defer delete(ctx.visiting, key)

	filtered := ifstring.FilterContent(p.ExportEnv, h.Triplet, builder.GetName())
	ownKeys := make(map[string]bool, len(filtered))
	for _, line := range filtered {
		ownKeys[envKeyFromLine(line)] = true
	}
	skipFromDeps := map[string]bool{}
	for _, k := range []string{"CFLAGS", "CXXFLAGS", "CPPFLAGS", "LDFLAGS", "LIBRARY_PATH", "PKG_CONFIG_PATH"} {
		if ownKeys[k] {
			skipFromDeps[k] = true
		}
	}

	base := copyEnv(p.minimalEnvWithStaging(h, p.baseBuildPath(h, "staging")))
	// Export-env describes the toolchain for h.Triplet; HOST/TARGET must match the
	// build target even when the exporting package is type "native".
	base["HOST"] = h.Triplet
	base["TARGET"] = h.Triplet
	for _, dep := range filteredDependencyPackages(p.Dependencies, h) {
		for _, e := range ctx.resolvedExportEnv(dep, h) {
			if skipFromDeps[e.K] {
				continue
			}
			base[e.K] = e.V
		}
	}

	after := utils.AppendEnv(copyEnv(base), p.ExportEnv, h)

	result := make([]EnvKV, 0, len(filtered))
	for _, line := range filtered {
		k := envKeyFromLine(line)
		if v, ok := after[k]; ok {
			result = append(result, EnvKV{K: k, V: v})
		}
	}
	ctx.memo[key] = result
	return result
}

func mergeResolvedExportEnv(env map[string]string, ctx *exportEnvContext, deps []*Package, h *host.Host) {
	for _, dep := range deps {
		for _, e := range ctx.resolvedExportEnv(dep, h) {
			env[e.K] = e.V
		}
	}
}

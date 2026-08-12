package pack

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/mrcyjanek/simplybs/builder"
	"github.com/mrcyjanek/simplybs/crash"
	"github.com/mrcyjanek/simplybs/host"
	"github.com/mrcyjanek/simplybs/utils"
	"github.com/mrcyjanek/simplybs/utils/ifstring"
)

func (p *Package) prefixPath(h *host.Host) string {
	if p.Type == "native" {
		return h.GetNativeEnvPath()
	}
	return h.GetEnvPath()
}

func (p *Package) homePath(h *host.Host) string {
	return filepath.Join(p.prefixPath(h), "home", "user")
}

func (p *Package) FilterForHost(h *host.Host) *Package {
	filtered := &Package{
		Package:      p.Package,
		Version:      p.Version,
		Type:         p.Type,
		Download:     p.Download,
		Dependencies: []string{},
	}
	builderName := builder.GetName()
	filtered.Dependencies = ifstring.FilterContent(p.Dependencies, h.Triplet, builderName)
	filtered.ExportEnv = ifstring.FilterContent(p.ExportEnv, h.Triplet, builderName)
	filtered.Build.Env = ifstring.FilterContent(p.Build.Env, h.Triplet, builderName)
	filtered.Build.Steps = ifstring.FilterContent(p.Build.Steps, h.Triplet, builderName)

	return filtered
}

func (p *Package) GeneratePackageInfo(h *host.Host) string {
	pkgs := map[string]interface{}{}
	pkgs["_target"] = p.FilterForHost(h)
	for _, dep := range p.Dependencies {
		is := ifstring.ParseIfString(dep)
		dep := is.Content
		pkg, err := FindPackage(dep)
		if err != nil {
			log.Fatalf("Package %s not found in info", dep)
		}
		pkgs[dep] = pkg.FilterForHost(h)
	}
	env := p.GetEnvForLogs(h)
	delete(env, "PATH")
	delete(env, "PREFIX")
	pkgs["_env"] = env
	pkgs["_prefix"] = p.prefixPath(h)
	info, err := json.MarshalIndent(pkgs, "", "  ")
	crash.Handle(err)
	return string(info)
}

func (p *Package) GeneratePackageInfoHash(h *host.Host) string {
	info := p.GeneratePackageInfo(h)
	hash := sha256.Sum256([]byte(info))
	return hex.EncodeToString(hash[:])
}

func (p *Package) GeneratePackageInfoShortHash(h *host.Host) string {
	hash := p.GeneratePackageInfoHash(h)
	return hash[:8]
}

func (p *Package) ShortName(h *host.Host) string {
	return p.Package + "-" + p.Version + "-" + p.GeneratePackageInfoShortHash(h)
}

func (p *Package) GenerateSourceBuildPath(download *Download) string {
	if download.Kind == "git" {
		return utils.SourcePathForGitURL(download.URL)
	}
	return utils.SourcePathForFileURL(download.URL)
}

func (p *Package) baseBuildPath(h *host.Host, kind string) string {
	safeName := strings.ReplaceAll(p.Package+"-"+p.Version, "/", "_")
	if p.Type == "native" {
		return filepath.Join(host.DataDir(), kind, safeName)
	}
	return filepath.Join(host.DataDir(), kind, h.Triplet, safeName)
}

func (p *Package) GenerateBuildPath(h *host.Host, kind string) string {
	if kind == "source" {
		log.Fatalf("Source build path is not supported")
	}
	safeName := strings.ReplaceAll(p.ShortName(h), "/", "_")
	if p.Type == "native" {
		return filepath.Join(host.DataDir(), kind, safeName)
	}
	return filepath.Join(host.DataDir(), kind, h.Triplet, safeName)
}

func getNumCores() int {
	cores, err := strconv.Atoi(os.Getenv("NUM_CORES"))
	if err != nil {
		cores = runtime.NumCPU()
	}
	return cores
}

func (p *Package) minimalEnvWithStaging(h *host.Host, stagingPath string) map[string]string {
	getwd, err := os.Getwd()
	crash.Handle(err)
	hostTriplet := h.Triplet
	targetTriplet := h.Triplet
	if p.Type == "native" {
		hostTriplet = builder.NativeTriplet()
		targetTriplet = builder.NativeTriplet()
	}
	return map[string]string{
		"PATH":                   utils.BuildStepPATH(h.GetNativeEnvPath(), h.GetEnvPath(), "", utils.GetHostPath()),
		"HOST":                   hostTriplet,
		"TARGET":                 targetTriplet,
		"PREFIX":                 utils.ToShellPath(h.GetEnvPath()),
		"NATIVEPREFIX":           utils.ToShellPath(h.GetNativeEnvPath()),
		"HOME":                   utils.ToShellPath(p.homePath(h)),
		"HOST_PREFIX":            utils.ToShellPath(h.GetEnvPath()),
		"NUM_CORES":              strconv.Itoa(getNumCores()),
		"PATCH_DIR":              utils.ToShellPath(filepath.Join(getwd, "patches")),
		"STAGING_DIR":            utils.ToShellPath(stagingPath),
		"PKG_CONFIG_ALLOW_CROSS": "1",
		"PKG_CONFIG_SYSROOT_DIR": utils.ToShellPath(h.GetEnvPath()),
		"BUILDER_GOOS":           runtime.GOOS,
		"BUILDER_GOARCH":         runtime.GOARCH,
	}
}

func (p *Package) minimalEnv(h *host.Host) map[string]string {
	return p.minimalEnvWithStaging(h, p.GenerateBuildPath(h, "staging"))
}

func (p *Package) GetEnv(h *host.Host) map[string]string {
	ctx := newExportEnvContext()
	env := p.minimalEnv(h)
	mergeResolvedExportEnv(env, ctx, filteredDependencyPackages(p.Dependencies, h), h)
	env = utils.AppendEnv(env, p.Build.Env, h)
	return env
}

func (p *Package) GetEnvForLogs(h *host.Host) map[string]string {
	ctx := newExportEnvContext()
	env := map[string]string{}
	mergeResolvedExportEnv(env, ctx, filteredDependencyPackages(p.Dependencies, h), h)
	env = utils.AppendEnv(env, p.Build.Env, h)
	return env
}

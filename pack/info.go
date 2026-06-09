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
	return p.prefixPath(h) + "/home/user"
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

func (p *Package) GetEnv(h *host.Host) map[string]string {
	getwd, err := os.Getwd()
	crash.Handle(err)
	stagingPath := p.GenerateBuildPath(h, "staging")
	env := map[string]string{
		"PATH":         h.GetNativeEnvPath() + "/bin:" + h.GetNativeEnvPath() + "/_/bin:" + utils.GetHostPath(),
		"HOST":         h.Triplet,
		"PREFIX":       h.GetEnvPath(),
		"NATIVEPREFIX": h.GetNativeEnvPath(),
		"HOME":         p.homePath(h),
		"HOST_PREFIX":  h.GetEnvPath(),
		"NUM_CORES":    strconv.Itoa(getNumCores()),
		"PATCH_DIR":    filepath.Join(getwd, "patches"),
		"STAGING_DIR":  stagingPath,
	}

	env = utils.AppendEnv(env, builder.HostBuilder.GlobalEnv, h)
	if p.Type == "native" {
		env = utils.AppendEnv(env, []string{
			"*:*:CFLAGS=$CFLAGS -I" + h.GetNativeEnvPath() + "/include",
			"*:*:CXXFLAGS=$CXXFLAGS -I" + h.GetNativeEnvPath() + "/include",
			"*:*:CPPFLAGS=$CPPFLAGS -I" + h.GetNativeEnvPath() + "/include",
			"*:*:LDFLAGS=$LDFLAGS -L" + h.GetNativeEnvPath() + "/lib",
			"*:*:LD_LIBRARY_PATH=$LD_LIBRARY_PATH:" + h.GetNativeEnvPath() + "/lib",
			"*:*:PKG_CONFIG_PATH=$PKG_CONFIG_PATH:" + h.GetNativeEnvPath() + "/lib/pkgconfig",
			"*:*:LIBRARY_PATH=$LIBRARY_PATH:" + h.GetNativeEnvPath() + "/lib",
		}, h)
	} else {
		env = utils.AppendEnv(env, []string{
			"*:*:CC_FOR_BUILD=" + builder.HostBuilder.GetCC(),
			"*:*:CXX_FOR_BUILD=" + builder.HostBuilder.GetCXX(),
			"*:*:CFLAGS=$CFLAGS -I" + h.GetEnvPath() + "/include",
			"*:*:CFLAGS=$CFLAGS -I" + h.GetEnvPath() + "/usr/include",
			"*:*:CXXFLAGS=$CXXFLAGS -I" + h.GetEnvPath() + "/include",
			"*:*:CXXFLAGS=$CXXFLAGS -I" + h.GetEnvPath() + "/usr/include",
			"*:*:CPPFLAGS=$CPPFLAGS -I" + h.GetEnvPath() + "/include",
			"*:*:CPPFLAGS=$CPPFLAGS -I" + h.GetEnvPath() + "/usr/include",
			"*:*:LDFLAGS=$LDFLAGS -L" + h.GetEnvPath() + "/lib",
			"*:*:LD_LIBRARY_PATH=" + h.GetNativeEnvPath() + "/lib",
			"*:*:PKG_CONFIG_PATH=$PKG_CONFIG_PATH:" + h.GetEnvPath() + "/lib/pkgconfig",
			"*:*:LIBRARY_PATH=$LIBRARY_PATH:" + h.GetEnvPath() + "/lib",
		}, h)
	}
	if p.Type != "native" {
		env = utils.AppendEnv(env, h.Env, h)
	}
	env = utils.AppendEnv(env, p.Build.Env, h)
	return env
}

func (p *Package) GetEnvForLogs(h *host.Host) map[string]string {
	env := map[string]string{}
	env = utils.AppendEnv(env, builder.HostBuilder.GlobalEnv, h)
	env = utils.AppendEnv(env, p.Build.Env, h)
	if p.Type != "native" {
		env = utils.AppendEnv(env, h.Env, h)
	}
	return env
}

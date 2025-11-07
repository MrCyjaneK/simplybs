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
)

func (p *Package) FilterForHost(h *host.Host) *Package {
	filtered := &Package{
		Package:      p.Package,
		Version:      p.Version,
		Type:         p.Type,
		Download:     p.Download,
		Dependencies: []string{},
	}
	filtered.Build.Env = []string{}
	filtered.Build.Steps = []string{}

	for _, dep := range p.Dependencies {
		is := utils.ParseIfString(dep)
		if !is.HostGlob().Match(h.Triplet) {
			continue
		}
		if !is.BuilderGlob().Match(builder.GetName()) {
			continue
		}
		filtered.Dependencies = append(filtered.Dependencies, is.Content)
	}

	for _, env := range p.Build.Env {
		is := utils.ParseIfString(env)
		if !is.HostGlob().Match(h.Triplet) {
			continue
		}
		if !is.BuilderGlob().Match(builder.GetName()) {
			continue
		}
		filtered.Build.Env = append(filtered.Build.Env, is.Content)
	}

	for _, step := range p.Build.Steps {
		is := utils.ParseIfString(step)
		if !is.HostGlob().Match(h.Triplet) {
			continue
		}
		if !is.BuilderGlob().Match(builder.GetName()) {
			continue
		}
		filtered.Build.Steps = append(filtered.Build.Steps, is.Content)
	}

	return filtered
}

func (p *Package) GeneratePackageInfo(h *host.Host) string {
	pkgs := map[string]interface{}{}
	pkgs["_target"] = p.FilterForHost(h)
	for _, dep := range p.Dependencies {
		is := utils.ParseIfString(dep)
		dep := is.Content
		pkg, err := FindPackage(dep)
		if err != nil {
			log.Fatalf("Package %s not found in info", dep)
		}
		pkgs[dep] = pkg.FilterForHost(h)
	}
	env := p.GetEnvForLogs(h)
	delete(env, "PATH")
	pkgs["_env"] = env
	pkgs["_prefix"] = h.GetEnvPath()
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
		urlPath, err := utils.URLToPath(download.URL)
		crash.Handle(err)
		urlPath = strings.TrimSuffix(urlPath, ".git")
		name := urlPath + ".bundle"
		return filepath.Join(host.DataDirRoot(), "source", name)
	}

	urlPath, err := utils.URLToPath(download.URL)
	if err != nil {
		return filepath.Join(host.DataDirRoot(), "source", filepath.Base(download.URL))
	}
	return filepath.Join(host.DataDirRoot(), "source", urlPath)
}

func (p *Package) GenerateBuildPath(h *host.Host, kind string) string {
	if kind == "source" {
		log.Fatalf("Source build path is not supported")
	}
	return filepath.Join(host.DataDir(), kind, h.Triplet, p.ShortName(h))
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
		"PATH":        h.GetEnvPath() + "/native/bin:" + utils.GetHostPath(),
		"HOST":        h.Triplet,
		"PREFIX":      h.GetEnvPath(),
		"HOME":        h.GetEnvPath() + "/home/user",
		"HOST_PREFIX": h.GetEnvPath(),
		"NUM_CORES":   strconv.Itoa(getNumCores()),
		"PATCH_DIR":   filepath.Join(getwd, "patches"),
		"STAGING_DIR": stagingPath,
	}

	env = utils.AppendEnv(env, builder.HostBuilder.GlobalEnv, h)
	if p.Type == "native" {
		env = utils.AppendEnv(env, []string{
			"*:CFLAGS=$CFLAGS -I" + h.GetEnvPath() + "/native/include",
			"*:CFLAGS=$CFLAGS -I" + h.GetEnvPath() + "/native/usr/include",
			"*:LDFLAGS=$LDFLAGS -L" + h.GetEnvPath() + "/native/lib",
			"*:LDFLAGS=$LDFLAGS -L" + h.GetEnvPath() + "/native/usr/lib",
			"*:LD_LIBRARY_PATH=$LD_LIBRARY_PATH:" + h.GetEnvPath() + "/native/lib",
			"*:PKG_CONFIG_PATH=$PKG_CONFIG_PATH:" + h.GetEnvPath() + "/native/lib/pkgconfig",
			"*:LIBRARY_PATH=$LIBRARY_PATH:" + h.GetEnvPath() + "/native/lib",
		}, h)
	} else {
		env = utils.AppendEnv(env, []string{
			"*:CC_FOR_BUILD=" + builder.HostBuilder.GetCC(),
			"*:CXX_FOR_BUILD=" + builder.HostBuilder.GetCXX(),
			"*:CFLAGS=$CFLAGS -I" + h.GetEnvPath() + "/include",
			"*:CFLAGS=$CFLAGS -I" + h.GetEnvPath() + "/usr/include",
			"*:LDFLAGS=$LDFLAGS -L" + h.GetEnvPath() + "/lib",
			"*:LDFLAGS=$LDFLAGS -L" + h.GetEnvPath() + "/usr/lib",
			"*:LD_LIBRARY_PATH=" + h.GetEnvPath() + "/native/lib",
			"*:PKG_CONFIG_PATH=$PKG_CONFIG_PATH:" + h.GetEnvPath() + "/lib/pkgconfig",
			"*:PKG_CONFIG_PATH=$PKG_CONFIG_PATH:" + h.GetEnvPath() + "/usr/lib/pkgconfig",
			"*:LIBRARY_PATH=$LIBRARY_PATH:" + h.GetEnvPath() + "/lib",
			"*:LIBRARY_PATH=$LIBRARY_PATH:" + h.GetEnvPath() + "/usr/lib",
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

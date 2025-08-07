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

func (p *Package) GeneratePackageInfo() string {
	pkgs := map[string]interface{}{}
	pkgs["_target"] = p
	for _, dep := range p.Dependencies {
		if strings.Contains(dep, ":") {
			dep = dep[strings.Index(dep, ":")+1:]
		}
		pkg, err := FindPackage(dep)
		if err != nil {
			log.Fatalf("Package %s not found in info", dep)
		}
		pkgs[dep] = pkg
	}
	env := p.GetEnvForLogs()
	delete(env, "PATH")
	pkgs["_env"] = env
	info, err := json.MarshalIndent(pkgs, "", "  ")
	crash.Handle(err)
	return string(info)
}

func (p *Package) GeneratePackageInfoHash() string {
	info := p.GeneratePackageInfo()
	hash := sha256.Sum256([]byte(info))
	return hex.EncodeToString(hash[:])
}

func (p *Package) GeneratePackageInfoShortHash() string {
	hash := p.GeneratePackageInfoHash()
	return hash[:8]
}

func (p *Package) ShortName() string {
	return p.Package + "-" + p.Version + "-" + p.GeneratePackageInfoShortHash()
}

func (p *Package) GenerateBuildPath(h *host.Host, kind string) string {
	if kind == "source" {
		var name string
		if p.Download.Kind == "git" {
			name = filepath.Base(p.Download.URL) + "-" + p.Download.Sha256[0:8] + ".git"
		} else {
			name = filepath.Base(p.Download.URL)
		}
		return filepath.Join(host.DataDir(), "..", kind, name)
	}
	return filepath.Join(host.DataDir(), kind, h.Triplet, p.ShortName())
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
	env := map[string]string{
		"PATH":        h.GetEnvPath() + "/native/bin:" + utils.GetHostPath(),
		"HOST":        h.Triplet,
		"PREFIX":      h.GetEnvPath(),
		"HOME":        h.GetEnvPath() + "/home/user",
		"HOST_PREFIX": h.GetEnvPath(),
		"NUM_CORES":   strconv.Itoa(runtime.NumCPU()),
		"PATCH_DIR":   filepath.Join(getwd, "patches"),
	}

	env = utils.AppendEnv(env, builder.HostBuilder.GlobalEnv, h)
	if p.Type == "native" {
		env = utils.AppendEnv(env, []string{
			"all:CFLAGS=$CFLAGS -I" + h.GetEnvPath() + "/native/include",
			"all:LDFLAGS=$LDFLAGS -L" + h.GetEnvPath() + "/native/lib",
			"all:LD_LIBRARY_PATH=$LD_LIBRARY_PATH:" + h.GetEnvPath() + "/native/lib",
			"all:PKG_CONFIG_PATH=$PKG_CONFIG_PATH:" + h.GetEnvPath() + "/native/lib/pkgconfig",
			"all:LIBRARY_PATH=$LIBRARY_PATH:" + h.GetEnvPath() + "/native/lib",
		}, h)
	} else {
		env = utils.AppendEnv(env, []string{
			"all:CFLAGS=-I" + h.GetEnvPath() + "/include",
			"all:LDFLAGS=-L" + h.GetEnvPath() + "/lib",
			"all:LD_LIBRARY_PATH=" + h.GetEnvPath() + "/lib",
			"all:PKG_CONFIG_PATH=" + h.GetEnvPath() + "/lib/pkgconfig",
			"all:LIBRARY_PATH=" + h.GetEnvPath() + "/lib",
		}, h)
	}
	if p.Type != "native" {
		env = utils.AppendEnv(env, h.Env, h)
	}
	env = utils.AppendEnv(env, p.Build.Env, h)
	return env
}

func (p *Package) GetEnvForLogs() map[string]string {
	env := map[string]string{}

	env = utils.AppendEnv(env, builder.HostBuilder.GlobalEnv, &host.Host{Triplet: "all"})
	env = utils.AppendEnv(env, p.Build.Env, &host.Host{Triplet: "all"})
	return env
}

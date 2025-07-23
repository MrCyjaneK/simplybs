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
	"github.com/ryanuber/go-glob"
)

func (p *Package) GeneratePackageInfo(h *host.Host) string {
	pkgs := map[string]interface{}{}
	pkgs["_target"] = p
	for _, dep := range p.Dependencies {
		if strings.Contains(dep, ":") {
			prefix := strings.Split(dep, ":")[0]
			if !glob.Glob(prefix, h.Triplet) && prefix != "all" {
				continue
			}
			dep = dep[strings.Index(dep, ":")+1:]
		}
		pkg, err := FindPackage(dep)
		if err != nil {
			log.Printf("Package %s not found in info", dep)
			continue
		}
		pkgs[dep] = pkg
	}
	env := p.GetEnv(h)
	delete(env, "PATH")
	pkgs["_env"] = env
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

func (p *Package) GenerateBuildPath(h *host.Host, kind string) string {
	if kind == "source" {
		return filepath.Join(host.DataDir(), "..", kind, p.Package+"-"+p.Version)
	}
	return filepath.Join(host.DataDir(), kind, h.Triplet, p.ShortName(h))
}

func (p *Package) GetEnv(h *host.Host) map[string]string {
	getwd, err := os.Getwd()
	crash.Handle(err)
	env := map[string]string{
		"PATH":        h.GetEnvPath() + "/native/bin:" + os.Getenv("PATH"),
		"HOST":        h.Triplet,
		"PREFIX":      h.GetEnvPath(),
		"HOST_PREFIX": h.GetEnvPath(),
		"NUM_CORES":   strconv.Itoa(runtime.NumCPU()),
		"PATCH_DIR":   filepath.Join(getwd, "patches", p.Package),
	}

	env = utils.AppendEnv(env, builder.HostBuilder.GlobalEnv, h)
	if p.Type == "native" {
		env = utils.AppendEnv(env, []string{
			"all:CFLAGS=-I" + h.GetEnvPath() + "/native/include",
			"all:LDFLAGS=-L" + h.GetEnvPath() + "/native/lib",
		}, h)
	}
	if p.Type != "native" {
		env = utils.AppendEnv(env, h.Env, h)
	}
	env = utils.AppendEnv(env, p.Build.Env, h)
	return env
}

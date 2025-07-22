package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/ryanuber/go-glob"
)

type Package struct {
	Package  string `json:"package"`
	Version  string `json:"version"`
	Type     string `json:"type"`
	Download struct {
		Kind   string `json:"kind"`
		URL    string `json:"url"`
		Sha256 string `json:"sha256"`
	} `json:"download"`
	Build struct {
		Env   []string `json:"env"`
		Steps []string `json:"steps"`
	} `json:"build"`
	Dependencies []string `json:"dependencies"`
}

func (p *Package) GeneratePackageInfo(h *Host) string {
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
		pkgs[dep] = FindPackage(dep)
	}
	env := p.GetEnv(h)
	delete(env, "PATH")
	pkgs["_env"] = env
	info, err := json.MarshalIndent(pkgs, "", "  ")
	crashErr(err)
	return string(info)
}

func (p *Package) GeneratePackageInfoHash(h *Host) string {
	info := p.GeneratePackageInfo(h)
	hash := sha256.Sum256([]byte(info))
	return hex.EncodeToString(hash[:])
}

func (p *Package) GeneratePackageInfoShortHash(h *Host) string {
	hash := p.GeneratePackageInfoHash(h)
	return hash[:8]
}

func (p *Package) ShortName(h *Host) string {
	return p.Package + "-" + p.Version + "-" + p.GeneratePackageInfoShortHash(h)
}

func (p *Package) GenerateBuildPath(h *Host, kind string) string {
	if kind == "source" {
		return filepath.Join(dataDir(), "..", kind, p.Package+"-"+p.Version)
	}
	return filepath.Join(dataDir(), kind, h.Triplet, p.ShortName(h))
}

func (p *Package) EnsureBuilt(h *Host, buildDependencies bool) {
	buildPath := p.GenerateBuildPath(h, "built") + ".info.txt"
	info, err := os.ReadFile(buildPath)
	if err != nil {
		log.Printf("[%s] No build cache found, building...", p.Package)
		p.BuildPackage(h, true)
		return
	}
	if string(info) == p.GeneratePackageInfo(h) {
		log.Printf("[%s] Build cache found, skipping build...", p.Package)
		return
	}
	log.Printf("[%s] Build cache found, but info mismatch, rebuilding...", p.Package)
	p.BuildPackage(h, true)
}

func (p *Package) ExtractEnv(host *Host, envPath string) {
	archive := p.GenerateBuildPath(host, "built") + ".tar.gz"
	err := ExtractTarGz(archive, envPath)
	if err != nil {
		log.Fatalf("Failed to extract archive %s: %v", archive, err)
	}
}

func (p *Package) DownloadSource(host *Host) {
	sourcePath := p.GenerateBuildPath(host, "source") + "." + p.Download.Kind
	os.MkdirAll(filepath.Dir(sourcePath), 0755)
	if p.Download.Kind == "none" {
		return
	}
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		if p.Download.Kind == "git" {
			DownloadGit(sourcePath, p.Download.URL, p.Download.Sha256)
		} else {
			DownloadFile(sourcePath, p.Download.URL, p.Download.Sha256)
		}
	}
}

func (p *Package) ExtractSource(host *Host, buildPath string) {
	sourcePath := p.GenerateBuildPath(host, "source") + "." + p.Download.Kind
	p.DownloadSource(host)
	var err error
	switch p.Download.Kind {
	case "tar.bz2":
		err = ExtractTarBz2(sourcePath, buildPath)
	case "tar.gz":
		err = ExtractTarGz(sourcePath, buildPath)
	case "git":
		os.MkdirAll(buildPath, 0755)
		err = os.CopyFS(buildPath, os.DirFS(sourcePath))
	case "none":
		return
	default:
		log.Fatalf("Unsupported archive kind: %s", p.Download.Kind)
	}
	if err != nil {
		log.Fatalf("Failed to extract archive %s: %v", sourcePath, err)
	}
}

func (p *Package) BuildPackage(h *Host, buildDependencies bool) {
	buildPath := p.GenerateBuildPath(h, "work")
	stagingPath := p.GenerateBuildPath(h, "staging")
	os.RemoveAll(buildPath)
	os.RemoveAll(stagingPath)
	os.MkdirAll(buildPath, 0755)
	os.MkdirAll(stagingPath, 0755)

	infoPath := filepath.Join(stagingPath, h.GetEnvPath(), "usr", "share", "buildlib", p.ShortName(h)+".txt")
	os.MkdirAll(filepath.Dir(infoPath), 0755)
	err := os.WriteFile(infoPath, []byte(p.GeneratePackageInfo(h)), 0644)
	if err != nil {
		log.Fatalf("Failed to write build info %s: %v", infoPath, err)
	}

	var deps []*Package
	if buildDependencies {
		for _, dep := range p.Dependencies {
			if strings.Contains(dep, ":") {
				prefix := strings.Split(dep, ":")[0]
				if !glob.Glob(prefix, h.Triplet) && prefix != "all" {
					continue
				}
				dep = dep[strings.Index(dep, ":")+1:]
			} else {
				log.Fatalf("Invalid dependency: %s", dep)
			}
			d := FindPackage(dep)
			deps = append(deps, d)
			d.EnsureBuilt(h, false)
		}
	}
	envPath := h.GetEnvPath()
	os.RemoveAll(envPath)
	os.MkdirAll(envPath, 0755)
	for _, dep := range deps {
		dep.ExtractEnv(h, envPath)
	}
	os.RemoveAll(buildPath)
	os.MkdirAll(buildPath, 0755)
	p.ExtractSource(h, buildPath)
	for _, step := range p.Build.Steps {
		if strings.Contains(step, ":") {
			prefix := strings.Split(step, ":")[0]
			if !glob.Glob(prefix, h.Triplet) && prefix != "all" {
				continue
			}
			step = step[strings.Index(step, ":")+1:]
		} else {
			log.Fatalf("Invalid step: %s", step)
		}

		cmd := exec.Command("sh", "-c", step)
		cmd.Dir = buildPath
		cmd.Env = os.Environ()
		env := p.GetEnv(h)

		cmd.Env = append(cmd.Env, []string{
			"STAGING_DIR=" + stagingPath,
			"HOST=" + h.Triplet,
			"PREFIX=" + h.GetEnvPath(),
			"PATH=" + h.GetEnvPath() + "/native/bin:" + env["PATH"],
		}...)
		for k, v := range env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}

		log.Printf("Executing step: %s", step)
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		err := cmd.Run()
		if err != nil {
			log.Fatalf("[%s] build step failed: %s, error: %v, %s", p.Package, step, err, cmd.Dir)
		}
	}

	builtArchivePath := p.GenerateBuildPath(h, "built") + ".tar.gz"
	os.MkdirAll(filepath.Dir(builtArchivePath), 0755)
	err = CreateTarGz(filepath.Join(stagingPath, h.GetEnvPath()), builtArchivePath)
	if err != nil {
		log.Fatalf("Failed to create archive %s: %v", builtArchivePath, err)
	}

	infoPath = p.GenerateBuildPath(h, "built") + ".info.txt"
	err = os.WriteFile(infoPath, []byte(p.GeneratePackageInfo(h)), 0644)
	if err != nil {
		log.Fatalf("Failed to write build info %s: %v", infoPath, err)
	}

	log.Printf("Package built successfully: %s", builtArchivePath)
}

func ExpandEnvFromMap(s string, envMap map[string]string) string {
	return os.Expand(s, func(key string) string {
		if val, ok := envMap[key]; ok {
			return val
		}
		return ""
	})
}

func appendEnv(env map[string]string, newEnv []string, host *Host) map[string]string {
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
func (p *Package) GetEnv(h *Host) map[string]string {
	getwd, err := os.Getwd()
	crashErr(err)
	env := map[string]string{
		"PATH":        h.GetEnvPath() + "/native/bin:" + os.Getenv("PATH"),
		"HOST":        h.Triplet,
		"PREFIX":      h.GetEnvPath(),
		"HOST_PREFIX": h.GetEnvPath(),
		"NUM_CORES":   strconv.Itoa(runtime.NumCPU()),
		"PATCH_DIR":   filepath.Join(getwd, "patches", p.Package),
	}

	env = appendEnv(env, hostBuilder.GlobalEnv, h)
	if p.Type != "native" {
		env = appendEnv(env, h.Env, h)
	}
	env = appendEnv(env, p.Build.Env, h)
	return env
}

func FindPackage(name string) *Package {
	buildDir, err := os.Getwd()
	crashErr(err)
	buildDir = filepath.Join(buildDir, "packages")
	buildDir = filepath.Join(buildDir, name+".json")
	info, err := os.ReadFile(buildDir)
	crashErr(err)
	var pkg Package
	err = json.Unmarshal(info, &pkg)
	crashErr(err)
	return &pkg
}

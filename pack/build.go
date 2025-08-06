package pack

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mrcyjanek/simplybs/host"
	"github.com/mrcyjanek/simplybs/utils"
	"github.com/ryanuber/go-glob"
)

func (p *Package) EnsureBuilt(h *host.Host, buildDependencies bool) {
	buildPath := p.GenerateBuildPath(h, "built") + ".info.txt"
	info, err := os.ReadFile(buildPath)
	if err != nil {
		log.Printf("[%s] No build cache found, building...", p.Package)
		p.BuildPackage(h, true)
		return
	}
	if string(info) == p.GeneratePackageInfo() {
		log.Printf("[%s] Build cache found, skipping build...", p.Package)
		return
	}
	log.Printf("[%s] Build cache found, but info mismatch, rebuilding...", p.Package)
	p.BuildPackage(h, true)
}

func (p *Package) ExtractEnv(host *host.Host, envPath string) {
	archive := p.GenerateBuildPath(host, "built") + ".tar.gz"
	err := utils.ExtractTarGz(archive, envPath)
	if err != nil {
		log.Fatalf("Failed to extract archive %s: %v", archive, err)
	}
}

func (p *Package) DownloadSource() {
	sourcePath := p.GenerateBuildPath(&host.Host{}, "source")
	os.MkdirAll(filepath.Dir(sourcePath), 0755)
	if p.Download.Kind == "none" {
		return
	}
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		var err error
		if p.Download.Kind == "git" {
			err = utils.DownloadGit(p.Package, sourcePath, p.Download.URL, p.Download.Sha256)
		} else {
			err = utils.DownloadFile(p.Package, sourcePath, p.Download.URL, p.Download.Sha256, false)
		}
		if err != nil {
			log.Fatalf("Failed to download source: %v", err)
		}
	}
}

func (p *Package) ExtractSource(host *host.Host, buildPath string) {
	sourcePath := p.GenerateBuildPath(host, "source")
	p.DownloadSource()
	var err error
	switch p.Download.Kind {
	case "tar.bz2":
		err = utils.ExtractTarBz2(sourcePath, buildPath)
	case "tar.gz":
		err = utils.ExtractTarGz(sourcePath, buildPath)
	case "tar.xz":
		err = utils.ExtractTarXz(sourcePath, buildPath)
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

func (p *Package) BuildPackage(h *host.Host, buildDependencies bool) {

	p.buildPackageInternal(h, buildDependencies)
}

func (p *Package) buildPackageInternal(h *host.Host, buildDependencies bool) {
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
			d, err := FindPackage(dep)
			if err != nil {
				log.Fatalf("Package %s not found in build", dep)
			}
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
	buildPath := p.GenerateBuildPath(h, "work")
	stagingPath := p.GenerateBuildPath(h, "staging")
	os.RemoveAll(buildPath)
	os.RemoveAll(stagingPath)
	os.MkdirAll(buildPath, 0755)
	os.MkdirAll(stagingPath, 0755)
	defer os.RemoveAll(buildPath)
	defer os.RemoveAll(stagingPath)

	p.ExtractSource(h, buildPath)

	infoPath := filepath.Join(stagingPath, h.GetEnvPath(), "usr", "share", "buildlib", p.ShortName()+".txt")
	os.MkdirAll(filepath.Dir(infoPath), 0755)
	err := os.WriteFile(infoPath, []byte(p.GeneratePackageInfo()), 0644)
	if err != nil {
		log.Fatalf("Failed to write build info %s: %v", infoPath, err)
	}

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
		pathEnv := utils.GetHostPath()
		env := p.GetEnv(h)

		cmd.Env = append(cmd.Env, []string{
			"STAGING_DIR=" + stagingPath,
			"HOST=" + h.Triplet,
			"PREFIX=" + h.GetEnvPath(),
			"PATH=" + h.GetEnvPath() + "/native/bin:" + env["PATH"] + ":" + pathEnv,
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
	err = utils.CreateTarGz(filepath.Join(stagingPath, h.GetEnvPath()), builtArchivePath)
	if err != nil {
		log.Fatalf("Failed to create archive %s: %v", builtArchivePath, err)
	}

	infoPath = p.GenerateBuildPath(h, "built") + ".info.txt"
	err = os.WriteFile(infoPath, []byte(p.GeneratePackageInfo()), 0644)
	if err != nil {
		log.Fatalf("Failed to write build info %s: %v", infoPath, err)
	}

	log.Printf("Package built successfully: %s", builtArchivePath)
}

func (p *Package) StartShell(h *host.Host) {
	log.Printf("Starting shell for package: %s for host %s", p.Package, h.Triplet)

	buildPath := p.GenerateBuildPath(h, "work")
	os.RemoveAll(buildPath)
	os.MkdirAll(buildPath, 0755)

	log.Printf("Extracting source for package: %s", p.Package)
	for _, depName := range p.Dependencies {
		if strings.Contains(depName, ":") {
			prefix := strings.Split(depName, ":")[0]
			if !glob.Glob(prefix, h.Triplet) && prefix != "all" {
				continue
			}
			depName = depName[strings.Index(depName, ":")+1:]
		} else {
			log.Fatalf("Invalid dependency: %s", depName)
		}
		dep, err := FindPackage(depName)
		if err != nil {
			log.Fatalf("Package %s not found in build", depName)
		}
		dep.ExtractEnv(h, h.GetEnvPath())
	}
	p.ExtractSource(h, buildPath)

	env := p.GetEnv(h)
	pathEnv := utils.GetHostPath()

	userShell := os.Getenv("SHELL")
	if userShell == "" {
		userShell = "/bin/sh"
	}

	for _, step := range p.Build.Steps {
		if strings.Contains(step, ":") {
			prefix := strings.Split(step, ":")[0]
			if !glob.Glob(prefix, h.Triplet) && prefix != "all" {
				log.Printf("[no match] %s", step)
				continue
			}
			log.Printf("   [match] %s", step)
		} else {
			log.Fatalf("Invalid step: %s", step)
		}
	}
	log.Printf("Starting %s in %s with build environment for %s", userShell, buildPath, h.Triplet)
	log.Printf("Type 'exit' to leave the shell")

	cmd := exec.Command(userShell)
	cmd.Dir = buildPath
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmd.Env = append(cmd.Env, []string{
		"HOST=" + h.Triplet,
		"PREFIX=" + h.GetEnvPath(),
		"PATH=" + h.GetEnvPath() + "/native/bin:" + env["PATH"] + ":" + pathEnv,
		"TERM=" + os.Getenv("TERM"),
	}...)

	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}

	err := cmd.Run()
	if err != nil {
		log.Printf("Shell exited with error: %v", err)
	} else {
		log.Printf("Shell session ended successfully")
	}
}

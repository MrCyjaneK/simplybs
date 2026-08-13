package pack

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/mrcyjanek/simplybs/builder"
	"github.com/mrcyjanek/simplybs/crash"
	"github.com/mrcyjanek/simplybs/host"
	"github.com/mrcyjanek/simplybs/utils"
	downloadpkg "github.com/mrcyjanek/simplybs/utils/download"
	"github.com/mrcyjanek/simplybs/utils/ifstring"
)

func (p *Package) EnsureBuilt(h *host.Host, buildDependencies bool) {
	buildPath := p.GenerateBuildPath(h, "built") + ".info.txt"
	info, err := os.ReadFile(buildPath)
	if err != nil {
		// Selective remote pull: only this package's assets (by short-hash name).
		if TryPullPackageCache(p, h) {
			info, err = os.ReadFile(buildPath)
		}
	}
	if err != nil {
		log.Printf("[%s][%s] No build cache found, building...", h.Triplet, p.Package)
		p.BuildPackage(h, true)
		return
	}
	if string(info) == p.GeneratePackageInfo(h) {
		log.Printf("[%s][%s] Build cache found, skipping build...", h.Triplet, p.Package)
		return
	}
	log.Printf("[%s][%s] Build cache found, but info mismatch, rebuilding...", h.Triplet, p.Package)
	p.BuildPackage(h, true)
}

func (p *Package) ExtractEnv(host *host.Host, envPath string, envNativePath string) {
	archive := p.GenerateBuildPath(host, "built") + ".tar.gz"
	archiveNative := p.GenerateBuildPath(host, "built") + "_native.tar.gz"
	err := utils.ExtractTarGz(archive, envPath)
	if err != nil {
		log.Panicf("Failed to extract archive %s: %v", archive, err)
	}
	err = utils.ExtractTarGz(archiveNative, envNativePath)
	if err != nil {
		log.Panicf("Failed to extract archive %s: %v", archive, err)
	}
	env := p.GetEnv(host)
	newEnv := []string{}
	for k, v := range env {
		newEnv = append(newEnv, k+"="+v)
	}
	exportEnv := p.GetExportEnv(host)
	exportEnvLines := []string{}
	for k, v := range exportEnv {
		exportEnvLines = append(exportEnvLines, k+"="+v)
	}
	os.MkdirAll(envPath, 0750)
	safeName := strings.ReplaceAll(p.Package, "/", "_")
	if p.Type == "native" {
		writeDotEnv(envNativePath+"/_source_me@"+safeName, newEnv)
		writeDotEnv(envNativePath+"/_source_me_export@"+safeName, exportEnvLines)
	} else {
		writeDotEnv(envPath+"/_source_me@"+safeName, newEnv)
		writeDotEnv(envPath+"/_source_me_export@"+safeName, exportEnvLines)
	}
}

func (p *Package) DownloadSource(download *Download) {
	sourcePath := p.GenerateSourceBuildPath(download)
	os.MkdirAll(filepath.Dir(sourcePath), 0755)
	if download.Kind == "none" {
		return
	}
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		var err error
		if download.Kind == "git" {
			err = utils.DownloadGit(p.Package, sourcePath, download.URL)
		} else {
			err = downloadpkg.DownloadFile(p.Package, sourcePath, download.URL, download.Sha256, false)
		}
		if err != nil {
			log.Fatalf("Failed to download source: %v", err)
		}
	}
}

func (p *Package) ExtractSource(host *host.Host, buildPath string) {
	for _, download := range p.Download {
		p.DownloadSource(download)
	}
	var err error
	for _, download := range p.Download {
		if download.Kind == "none" {
			continue
		}
		sourcePath := p.GenerateSourceBuildPath(download)
		actualBuildPath := buildPath
		if download.Path != "" {
			actualBuildPath = filepath.Join(buildPath, download.Path)
		}
		switch download.Kind {
		case "tar.bz2":
			err = utils.ExtractTarBz2(sourcePath, actualBuildPath)
		case "tar.gz":
			err = utils.ExtractTarGz(sourcePath, actualBuildPath)
		case "tar.xz":
			err = utils.ExtractTarXz(sourcePath, actualBuildPath)
		case "git":
			err = utils.ExtractGitCloneBundle(sourcePath, actualBuildPath, download.Sha256)
		case "blob":
			os.MkdirAll(filepath.Dir(actualBuildPath), 0755)
			err = utils.Copy(sourcePath, actualBuildPath)
		default:
			log.Fatalf("Unsupported archive kind: %s", download.Kind)
		}

		if err != nil {
			log.Fatalf("Failed to extract archive %s: %v", sourcePath, err)
		}
	}
}

func (p *Package) BuildPackage(h *host.Host, buildDependencies bool) {
	p.buildPackageInternal(h, buildDependencies)
}

func filteredDependencyPackages(deps []string, h *host.Host) []*Package {
	var pkgs []*Package
	for _, dep := range deps {
		is := ifstring.ParseIfString(dep)
		if !is.Matches(h.Triplet, builder.GetName()) {
			continue
		}
		d, err := FindPackage(is.Content)
		if err != nil {
			log.Fatalf("Package %s not found in build", is.Content)
		}
		pkgs = append(pkgs, d)
	}
	return pkgs
}

func wipeDirContents(dir string) {
	if entries, err := os.ReadDir(dir); err == nil {
		for _, entry := range entries {
			utils.RemoveAll(filepath.Join(dir, entry.Name()))
		}
	}
}

func resetBuildDirs() {
	for _, dir := range []string{
		filepath.Join(host.DataDir(), "work"),
		filepath.Join(host.DataDir(), "staging"),
	} {
		utils.RemoveAll(dir)
	}
}

func parseStepProfile(content string) (profile, cmd string) {
	if strings.HasPrefix(content, "@") {
		rest := content[1:]
		if idx := strings.IndexByte(rest, ' '); idx != -1 {
			return rest[:idx], rest[idx+1:]
		}
		return rest, ""
	}
	return "", content
}

func (p *Package) envForStep(h *host.Host, profileName string) map[string]string {
	return p.envForStepFrom(p.GetEnv(h), h, profileName)
}

func (p *Package) envForStepFrom(env map[string]string, h *host.Host, profileName string) map[string]string {
	env = copyEnv(env)
	if profileName == "" {
		return env
	}
	prof, ok := p.Build.Profiles[profileName]
	if !ok {
		log.Fatalf("[%s] unknown build profile %q", p.Package, profileName)
	}
	for _, k := range prof.Unset {
		delete(env, k)
	}
	if len(prof.Env) > 0 {
		env = utils.AppendEnv(env, prof.Env, h)
	}
	return env
}

func (p *Package) buildPackageInternal(h *host.Host, buildDependencies bool) {
	var deps []*Package
	if buildDependencies {
		deps = filteredDependencyPackages(p.Dependencies, h)
		for _, d := range deps {
			d.EnsureBuilt(h, false)
		}
	}
	envPath := h.GetEnvPath()
	wipeDirContents(envPath)
	nativeEnvPath := h.GetNativeEnvPath()
	wipeDirContents(nativeEnvPath)
	os.MkdirAll(envPath, 0755)
	for _, dep := range deps {
		dep.ExtractEnv(h, envPath, nativeEnvPath)
	}
	resetBuildDirs()
	buildPath := p.GenerateBuildPath(h, "work")
	stagingPath := p.GenerateBuildPath(h, "staging")
	os.MkdirAll(buildPath, 0755)
	os.MkdirAll(stagingPath, 0755)
	defer utils.RemoveAll(buildPath)
	defer utils.RemoveAll(stagingPath)

	p.ExtractSource(h, buildPath)

	infoPath := filepath.Join(utils.DestDirJoin(stagingPath, p.prefixPath(h)), ".buildlib", p.ShortName(h)+".txt")
	os.MkdirAll(filepath.Dir(infoPath), 0755)
	err := os.WriteFile(infoPath, []byte(p.GeneratePackageInfo(h)), 0644)
	if err != nil {
		log.Fatalf("Failed to write build info %s: %v", infoPath, err)
	}

	builderName := builder.GetName()
	for _, step := range ifstring.FilterContent(p.Build.Steps, h.Triplet, builderName) {
		profileName, stepCmd := parseStepProfile(step)
		baseEnv := p.GetEnv(h)
		profileName = utils.ExpandEnvFromMap(profileName, baseEnv)
		if profileName == "" && strings.HasPrefix(step, "@") {
			continue
		}
		nativePrefix := h.GetNativeEnvPath()
		tmpDir := filepath.Join(nativePrefix, "_", "tmp")
		os.MkdirAll(tmpDir, 0755)
		shell := utils.ResolveShell(nativePrefix)
		cmd := exec.Command(shell, "-c", stepCmd)
		cmd.Dir = buildPath
		pathEnv := utils.GetHostPath()
		env := p.envForStepFrom(baseEnv, h, profileName)

		hostTriplet := h.Triplet
		if p.Type == "native" {
			hostTriplet = builder.NativeTriplet()
		}

		// Explicit env only — do not inherit the parent process environment.
		overrides := map[string]string{
			"STAGING_DIR":  utils.ToShellPath(stagingPath),
			"HOST":         hostTriplet,
			"PREFIX":       utils.ToShellPath(h.GetEnvPath()),
			"NATIVEPREFIX": utils.ToShellPath(nativePrefix),
			"PATH":         utils.BuildStepPATH(nativePrefix, h.GetEnvPath(), env["PATH"], pathEnv),
			"TEMP":         utils.ToShellPath(tmpDir),
			"TMP":          utils.ToShellPath(tmpDir),
			"TMPDIR":       utils.ToShellPath(tmpDir),
			"CYGWIN":       "nodosfilewarning",
			"MSYS":         "nodosfilewarning",
		}
		for k, v := range env {
			overrides[k] = v
		}
		cmd.Env = utils.ProcessEnv(overrides)

		writeDotEnv(path.Join(buildPath, "_source_me"), cmd.Env)

		log.Printf("Executing step: %s", stepCmd)
		if profileName != "" {
			log.Printf("Using build profile: %s", profileName)
		}
		err := utils.RunLive(cmd)
		if err != nil {
			log.Fatalf("[%s] build step failed: %s, error: %v, %s", p.Package, stepCmd, err, cmd.Dir)
		}
	}

	builtArchivePath := p.GenerateBuildPath(h, "built") + ".tar.gz"
	builtNativeArchivePath := p.GenerateBuildPath(h, "built") + "_native.tar.gz"
	os.MkdirAll(filepath.Dir(builtArchivePath), 0755)
	os.MkdirAll(filepath.Dir(builtNativeArchivePath), 0755)
	stagedNative := utils.DestDirJoin(stagingPath, h.GetNativeEnvPath())
	stagedHost := utils.DestDirJoin(stagingPath, h.GetEnvPath())
	os.MkdirAll(stagedNative, 0755)
	os.MkdirAll(stagedHost, 0755)
	err = utils.CreateTarGz(stagedNative, builtNativeArchivePath)
	if err != nil {
		log.Fatalf("Failed to create archive %s: %v", builtArchivePath, err)
	}
	err = utils.CreateTarGz(stagedHost, builtArchivePath)
	if err != nil {
		log.Fatalf("Failed to create archive %s: %v", builtArchivePath, err)
	}

	infoPath = p.GenerateBuildPath(h, "built") + ".info.txt"
	err = os.WriteFile(infoPath, []byte(p.GeneratePackageInfo(h)), 0644)
	if err != nil {
		log.Fatalf("Failed to write build info %s: %v", infoPath, err)
	}

	log.Printf("Package built successfully: %s", builtArchivePath)
	TryPushPackageCache(p, h)
}

func writeDotEnv(path string, env []string) {
	f, err := os.Create(path)
	if err != nil {
		crash.Handle(err)
	}
	defer f.Close()

	var b strings.Builder
	b.WriteString("#!/bin/bash\n")

	for _, kv := range env {
		if kv == "" {
			continue
		}

		parts := strings.SplitN(kv, "=", 2)
		key := parts[0]
		val := ""
		if len(parts) > 1 {
			val = parts[1]
		}
		// Skip keys that are not valid bash identifiers (e.g. inherited Windows names).
		if !isBashIdent(key) {
			continue
		}

		escapedVal := shellEscape(val)

		b.WriteString(fmt.Sprintf("export %s=%s\n", key, escapedVal))
	}

	f.WriteString(b.String())
}

func isBashIdent(name string) bool {
	if name == "" {
		return false
	}
	for i, r := range name {
		if r == '_' || (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
			continue
		}
		if i > 0 && r >= '0' && r <= '9' {
			continue
		}
		return false
	}
	return true
}
func shellEscape(value string) string {
	if value == "" {
		return "\"\""
	}

	replacer := strings.NewReplacer(
		`\`, `\\`,
		`"`, `\"`,
		"`", "\\`",
		`$`, `\$`,
	)
	escaped := replacer.Replace(value)

	return "\"" + escaped + "\""
}

func (p *Package) StartShell(h *host.Host) {
	log.Printf("Starting shell for package: %s for host %s", p.Package, h.Triplet)

	resetBuildDirs()
	buildPath := p.GenerateBuildPath(h, "work")
	os.MkdirAll(buildPath, 0755)

	log.Printf("Extracting source for package: %s", p.Package)
	for _, dep := range filteredDependencyPackages(p.Dependencies, h) {
		dep.ExtractEnv(h, h.GetEnvPath(), h.GetNativeEnvPath())
	}
	p.ExtractSource(h, buildPath)

	env := p.GetEnv(h)
	pathEnv := utils.GetHostPath()
	nativePrefix := h.GetNativeEnvPath()
	os.MkdirAll(filepath.Join(nativePrefix, "_", "tmp"), 0755)

	userShell := os.Getenv("SHELL")
	if userShell == "" {
		userShell = utils.ResolveShell(nativePrefix)
	}
	fmt.Print("\n\n")
	builderName := builder.GetName()
	for _, step := range ifstring.FilterContent(p.Build.Steps, h.Triplet, builderName) {
		fmt.Printf("%s;\n", utils.ExpandEnvFromMap(step, env))
	}
	fmt.Print("\n\n")

	for _, step := range p.Build.Steps {
		is := ifstring.ParseIfString(step)
		if !is.Matches(h.Triplet, builderName) {
			log.Printf("[no match] %s", utils.ExpandEnvFromMap(step, env))
			continue
		}
		log.Printf("   [match] %s", utils.ExpandEnvFromMap(is.Content, env))
	}
	fmt.Print("\n\n")
	log.Printf("Starting %s in %s with build environment for %s", userShell, buildPath, h.Triplet)
	log.Printf("Type 'exit' to leave the shell")

	cmd := exec.Command(userShell)
	cmd.Dir = buildPath
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	hostTriplet := h.Triplet
	if p.Type == "native" {
		hostTriplet = builder.NativeTriplet()
	}

	overrides := map[string]string{
		"HOST":         hostTriplet,
		"PREFIX":       utils.ToShellPath(h.GetEnvPath()),
		"NATIVEPREFIX": utils.ToShellPath(nativePrefix),
		"PATH":         utils.BuildStepPATH(nativePrefix, h.GetEnvPath(), env["PATH"], pathEnv),
		"TEMP":         utils.ToShellPath(filepath.Join(nativePrefix, "_", "tmp")),
		"TMP":          utils.ToShellPath(filepath.Join(nativePrefix, "_", "tmp")),
		"TMPDIR":       utils.ToShellPath(filepath.Join(nativePrefix, "_", "tmp")),
		"CYGWIN":       "nodosfilewarning",
		"MSYS":         "nodosfilewarning",
	}
	if term := os.Getenv("TERM"); term != "" {
		overrides["TERM"] = term
	}
	for k, v := range env {
		overrides[k] = v
	}
	cmd.Env = utils.ProcessEnv(overrides)
	writeDotEnv(path.Join(buildPath, "_source_me"), cmd.Env)

	err := cmd.Run()
	if err != nil {
		log.Printf("Shell exited with error: %v", err)
	} else {
		log.Printf("Shell session ended successfully")
	}
}

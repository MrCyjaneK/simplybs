package utils

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// WindowsCygwinSeedRel is the repo-relative seed tree used before
// $NATIVEPREFIX/_ exists (PATCH_DIR / easybs-shaped Cygwin+clang).
const WindowsCygwinSeedRel = "patches/native/_/windows_amd64_cygwin_clang"

func WindowsCygwinSeedBin() string {
	wd, err := os.Getwd()
	if err != nil {
		return filepath.Join(WindowsCygwinSeedRel, "bin")
	}
	return filepath.Join(wd, WindowsCygwinSeedRel, "bin")
}

func windowsGitBins() []string {
	var out []string
	roots := []string{
		os.Getenv("ProgramFiles"),
		`C:\Program Files`,
	}
	if pf86 := os.Getenv("ProgramFiles(x86)"); pf86 != "" {
		roots = append(roots, pf86)
	}
	for _, root := range roots {
		if root == "" {
			continue
		}
		out = append(out,
			filepath.Join(root, "Git", "cmd"),
			filepath.Join(root, "Git", "bin"),
			filepath.Join(root, "Git", "mingw64", "bin"),
			filepath.Join(root, "Git", "usr", "bin"),
		)
	}
	return out
}

// ResolveGit returns an absolute path to git for host-side downloads.
// On Windows this finds Git-for-Windows without putting its MSYS usr/bin on
// the build-step PATH (which breaks autoconf/egrep).
func ResolveGit() string {
	if p, err := exec.LookPath("git"); err == nil {
		return p
	}
	if runtime.GOOS != "windows" {
		return "git"
	}
	for _, dir := range windowsGitBins() {
		cand := filepath.Join(dir, "git.exe")
		if st, err := os.Stat(cand); err == nil && !st.IsDir() {
			return cand
		}
	}
	return "git"
}

func guessNativePrefix() string {
	if v := os.Getenv("SIMPLYBS_NATIVE_ENV_DIR"); v != "" {
		return v
	}
	wd, err := os.Getwd()
	if err != nil {
		return filepath.Join(".buildlib", "env-native")
	}
	if v := os.Getenv("SIMPLYBS_DATA_DIR"); v != "" {
		return filepath.Join(v, "env-native")
	}
	return filepath.Join(wd, ".buildlib", "env-native")
}

func GetHostPath() string {
	switch runtime.GOOS {
	case "darwin":
		return ""
	case "linux":
		return ""
	case "windows":
		// Prefer the Cygwin seed only. Git-for-Windows on PATH makes
		// autoconf pick MSYS grep/sed (paths with spaces) and break egrep.
		return ToShellPath(WindowsCygwinSeedBin())
	default:
		log.Fatalln("Unsupported OS: ", runtime.GOOS)
		return ""
	}
}

// ToShellPath converts a native path into the form used inside build-step
// shells. On Windows this project always uses Cygwin (/cygdrive/...), even
// when ResolveShell temporarily falls back to Git Bash after env-native is
// wiped between EnsureBuilt and ExtractEnv. Hashing PREFIX/NATIVEPREFIX via
// Git's /c/... form would look for a different archive than the one built
// under Cygwin.
func ToShellPath(p string) string {
	if runtime.GOOS != "windows" {
		return p
	}
	if p == "" {
		return p
	}
	abs := p
	if !filepath.IsAbs(abs) {
		if a, err := filepath.Abs(abs); err == nil {
			abs = a
		}
	}
	vol := filepath.VolumeName(abs)
	if len(vol) >= 2 && vol[1] == ':' {
		letter := strings.ToLower(string(vol[0]))
		rest := filepath.ToSlash(abs[len(vol):])
		if rest == "" {
			rest = "/"
		}
		return "/cygdrive/" + letter + rest
	}
	return filepath.ToSlash(abs)
}

// DestDirJoin mirrors Unix string concat of $STAGING_DIR$PREFIX for archives
// and Go file ops so the staged tree matches what the shell wrote.
func DestDirJoin(staging, prefix string) string {
	if runtime.GOOS != "windows" {
		rel := strings.TrimPrefix(filepath.ToSlash(prefix), "/")
		return filepath.Join(staging, filepath.FromSlash(rel))
	}
	shellPrefix := ToShellPath(prefix)
	rel := strings.TrimPrefix(shellPrefix, "/")
	return filepath.Join(staging, filepath.FromSlash(rel))
}

// ResolveShell returns the shell used for build steps.
// Prefer Cygwin seed sh when present; otherwise Git for Windows sh.
func ResolveShell(nativePrefix string) string {
	if runtime.GOOS != "windows" {
		return "sh"
	}
	if nativePrefix == "" {
		nativePrefix = guessNativePrefix()
	}
	candidates := []string{
		filepath.Join(nativePrefix, "_", "bin", "sh.exe"),
		filepath.Join(nativePrefix, "_", "bin", "sh"),
		// After native/_ stages tools into bin/, prefer that Cygwin sh
		// over Git Bash so /cygdrive paths keep working.
		filepath.Join(nativePrefix, "bin", "sh.exe"),
		filepath.Join(nativePrefix, "bin", "sh"),
		filepath.Join(nativePrefix, "bin", "bash.exe"),
		filepath.Join(nativePrefix, "bin", "bash"),
		filepath.Join(WindowsCygwinSeedBin(), "sh.exe"),
		filepath.Join(WindowsCygwinSeedBin(), "sh"),
	}
	for _, b := range windowsGitBins() {
		candidates = append(candidates, filepath.Join(b, "sh.exe"))
	}
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c
		}
	}
	if p, err := exec.LookPath("sh.exe"); err == nil {
		return p
	}
	if p, err := exec.LookPath("sh"); err == nil {
		return p
	}
	return "sh"
}

// ProcessEnv builds an explicit process environment from overrides only.
// On Windows, SYSTEMROOT is copied from the host so Win32 APIs keep working;
// nothing else is inherited from the parent process.
func ProcessEnv(overrides map[string]string) []string {
	out := make([]string, 0, len(overrides)+1)
	if runtime.GOOS == "windows" {
		if !hasEnvKey(overrides, "SYSTEMROOT") {
			if v := os.Getenv("SYSTEMROOT"); v != "" {
				out = append(out, "SYSTEMROOT="+v)
			}
		}
	}
	for k, v := range overrides {
		out = append(out, k+"="+v)
	}
	return out
}

func hasEnvKey(env map[string]string, key string) bool {
	if _, ok := env[key]; ok {
		return true
	}
	if runtime.GOOS != "windows" {
		return false
	}
	want := strings.ToUpper(key)
	for k := range env {
		if strings.ToUpper(k) == want {
			return true
		}
	}
	return false
}

// BuildStepPATH prefers the simplybs native toolchain, then bootstrap host
// bins (seed / Git), then prefix bins. Always colon-separated for POSIX sh.
// $NATIVEPREFIX/bin (bash, etc.) must precede $NATIVEPREFIX/_/bin (clang seed
// toybox sh) so Configure scripts pick up bash for SHELL.
func BuildStepPATH(nativePrefix, prefix, envPath, hostPath string) string {
	parts := []string{
		ToShellPath(filepath.Join(nativePrefix, "bin")),
		ToShellPath(filepath.Join(nativePrefix, "_", "bin")),
	}
	if envPath != "" {
		parts = append(parts, envPath)
	}
	if hostPath != "" {
		parts = append(parts, hostPath)
	}
	if prefix != "" {
		parts = append(parts, ToShellPath(filepath.Join(prefix, "bin")))
	}
	out := parts[0]
	for _, p := range parts[1:] {
		if p == "" {
			continue
		}
		out += ":" + p
	}
	return out
}

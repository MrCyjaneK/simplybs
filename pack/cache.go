package pack

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/mrcyjanek/simplybs/host"
)

// Remote build-cache assets live on a GitHub release. Paths under
// DataDir()/built are flattened with '+' so release asset names stay flat:
//
//	x86_64-linux-gnu/zlib-1.3.1-deadbeef.tar.gz
//	  -> x86_64-linux-gnu+zlib-1.3.1-deadbeef.tar.gz
//
// Pull only requests assets for the packages needed by the current build
// (by short-hash filename). Push only uploads local files that are missing
// from the release (content changes produce a new short-hash name).
//
// Cache is enabled only when both SIMPLYBS_CACHE_TAG and SIMPLYBS_CACHE_REPO
// are set; EnsureBuilt then auto-pulls missing artifacts.

const cacheAssetSuffixes = 3 // .info.txt, .tar.gz, _native.tar.gz

var (
	remoteAssetsOnce sync.Once
	remoteAssets     map[string]bool
	remoteAssetsErr  error
)

// CacheTag returns SIMPLYBS_CACHE_TAG (empty if unset).
func CacheTag() string {
	return os.Getenv("SIMPLYBS_CACHE_TAG")
}

// CacheRepo returns SIMPLYBS_CACHE_REPO (owner/repo; empty if unset).
func CacheRepo() string {
	return os.Getenv("SIMPLYBS_CACHE_REPO")
}

// CacheEnabled reports whether remote build cache is configured.
// Both SIMPLYBS_CACHE_TAG and SIMPLYBS_CACHE_REPO must be set.
func CacheEnabled() bool {
	return CacheTag() != "" && CacheRepo() != ""
}

func requireCacheConfig() (tag, repo string, err error) {
	tag = CacheTag()
	repo = CacheRepo()
	if tag == "" || repo == "" {
		return "", "", fmt.Errorf("cache: set both SIMPLYBS_CACHE_TAG and SIMPLYBS_CACHE_REPO")
	}
	return tag, repo, nil
}

func ghCmd(args ...string) *exec.Cmd {
	bin := os.Getenv("SIMPLYBS_GH")
	if bin == "" {
		bin = "gh"
	}
	repo := CacheRepo()
	if repo == "" {
		return exec.Command(bin, args...)
	}
	return exec.Command(bin, append([]string{"-R", repo}, args...)...)
}

// AssetNameForRel converts a path relative to DataDir()/built into a flat
// release asset name.
func AssetNameForRel(rel string) string {
	rel = filepath.ToSlash(rel)
	if strings.Contains(rel, "/") {
		return strings.ReplaceAll(rel, "/", "+")
	}
	return rel
}

// RelFromAssetName reverses AssetNameForRel.
func RelFromAssetName(name string) string {
	return strings.ReplaceAll(name, "+", "/")
}

// BuiltRelPaths returns the paths (relative to DataDir()/built) for a
// package's cache artifacts on host h.
func (p *Package) BuiltRelPaths(h *host.Host) []string {
	base := p.GenerateBuildPath(h, "built")
	builtRoot := filepath.Join(host.DataDir(), "built")
	relBase, err := filepath.Rel(builtRoot, base)
	if err != nil {
		log.Fatalf("cache: relative path for %s: %v", base, err)
	}
	relBase = filepath.ToSlash(relBase)
	return []string{
		relBase + ".info.txt",
		relBase + ".tar.gz",
		relBase + "_native.tar.gz",
	}
}

// CollectNeededPackages returns pkg plus its transitive dependencies for h.
func CollectNeededPackages(pkgs []*Package, h *host.Host) []*Package {
	seen := map[string]*Package{}
	var walk func(*Package)
	walk = func(p *Package) {
		if p == nil {
			return
		}
		if _, ok := seen[p.Package]; ok {
			return
		}
		seen[p.Package] = p
		for _, dep := range filteredDependencyPackages(p.Dependencies, h) {
			walk(dep)
		}
	}
	for _, p := range pkgs {
		walk(p)
	}
	out := make([]*Package, 0, len(seen))
	for _, p := range seen {
		out = append(out, p)
	}
	return out
}

func neededAssetNames(pkgs []*Package, h *host.Host) []string {
	needed := CollectNeededPackages(pkgs, h)
	names := make([]string, 0, len(needed)*cacheAssetSuffixes)
	for _, p := range needed {
		for _, rel := range p.BuiltRelPaths(h) {
			names = append(names, AssetNameForRel(rel))
		}
	}
	return names
}

func loadRemoteAssets(tag string) (map[string]bool, error) {
	remoteAssetsOnce.Do(func() {
		remoteAssets = map[string]bool{}
		cmd := ghCmd("release", "view", tag, "--json", "assets")
		out, err := cmd.Output()
		if err != nil {
			stderr := ""
			if exitErr, ok := err.(*exec.ExitError); ok {
				stderr = string(exitErr.Stderr)
			}
			msg := strings.ToLower(stderr + string(out) + err.Error())
			if strings.Contains(msg, "not found") || strings.Contains(msg, "could not find") || strings.Contains(msg, "http 404") {
				log.Printf("cache: release %q not found; treating as empty", tag)
				return
			}
			remoteAssetsErr = fmt.Errorf("gh release view %s: %w%s", tag, err, formatStderr(stderr))
			return
		}
		var parsed struct {
			Assets []struct {
				Name string `json:"name"`
			} `json:"assets"`
		}
		if err := json.Unmarshal(out, &parsed); err != nil {
			remoteAssetsErr = err
			return
		}
		for _, a := range parsed.Assets {
			remoteAssets[a.Name] = true
		}
	})
	return remoteAssets, remoteAssetsErr
}

func formatStderr(stderr string) string {
	stderr = strings.TrimSpace(stderr)
	if stderr == "" {
		return ""
	}
	return ": " + stderr
}

func resetRemoteAssetsCache() {
	remoteAssetsOnce = sync.Once{}
	remoteAssets = nil
	remoteAssetsErr = nil
}

func ensureRelease(tag string) error {
	view := ghCmd("release", "view", tag)
	if err := view.Run(); err == nil {
		return nil
	}
	create := ghCmd("release", "create", tag,
		"--title", "simplybs build cache",
		"--notes", "Rolling cache of simplybs build artifacts (per-package, content-addressed by short hash).",
	)
	create.Stdout = os.Stdout
	create.Stderr = os.Stderr
	return create.Run()
}

func downloadAsset(tag, assetName, destPath string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}
	staging, err := os.MkdirTemp("", "simplybs-cache-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(staging)

	cmd := ghCmd("release", "download", tag, "-p", assetName, "-D", staging)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	src := filepath.Join(staging, assetName)
	// gh may write the literal asset name; if the pattern matched a file with
	// a different on-disk name, pick the sole file in staging.
	if _, err := os.Stat(src); err != nil {
		entries, readErr := os.ReadDir(staging)
		if readErr != nil {
			return err
		}
		if len(entries) != 1 || entries[0].IsDir() {
			return fmt.Errorf("cache: expected one downloaded file for %s", assetName)
		}
		src = filepath.Join(staging, entries[0].Name())
	}
	tmp := destPath + ".tmp"
	if err := copyFile(src, tmp); err != nil {
		return err
	}
	return os.Rename(tmp, destPath)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// TryPullPackageCache downloads missing built artifacts for one package from
// the GitHub release cache. Returns true when a complete local cache is present
// afterwards (either already was, or was fetched).
func TryPullPackageCache(p *Package, h *host.Host) bool {
	if !CacheEnabled() {
		return false
	}
	tag := CacheTag()
	assets, err := loadRemoteAssets(tag)
	if err != nil {
		log.Printf("cache: failed to list release assets: %v", err)
		return false
	}
	builtRoot := filepath.Join(host.DataDir(), "built")
	ok := true
	for _, rel := range p.BuiltRelPaths(h) {
		dest := filepath.Join(builtRoot, filepath.FromSlash(rel))
		if _, err := os.Stat(dest); err == nil {
			continue
		}
		name := AssetNameForRel(rel)
		if !assets[name] {
			ok = false
			continue
		}
		log.Printf("[%s][%s] cache pull: %s", h.Triplet, p.Package, name)
		if err := downloadAsset(tag, name, dest); err != nil {
			log.Printf("[%s][%s] cache pull failed for %s: %v", h.Triplet, p.Package, name, err)
			ok = false
		}
	}
	return ok
}

// CachePull downloads only the built artifacts needed for pkgs on host h.
func CachePull(pkgs []*Package, h *host.Host) error {
	tag, _, err := requireCacheConfig()
	if err != nil {
		return err
	}
	assets, err := loadRemoteAssets(tag)
	if err != nil {
		return fmt.Errorf("cache: list assets: %w", err)
	}

	builtRoot := filepath.Join(host.DataDir(), "built")
	names := neededAssetNames(pkgs, h)
	toDownload := 0
	skippedLocal := 0
	missingRemote := 0
	for _, name := range names {
		rel := RelFromAssetName(name)
		dest := filepath.Join(builtRoot, filepath.FromSlash(rel))
		if _, err := os.Stat(dest); err == nil {
			skippedLocal++
			continue
		}
		if !assets[name] {
			missingRemote++
			continue
		}
		log.Printf("cache: download %s", name)
		if err := downloadAsset(tag, name, dest); err != nil {
			return fmt.Errorf("cache: download %s: %w", name, err)
		}
		toDownload++
	}
	log.Printf("cache pull: downloaded=%d already-local=%d not-on-release=%d needed=%d tag=%s",
		toDownload, skippedLocal, missingRemote, len(names), tag)
	return nil
}

// TryPushPackageCache uploads this package's built artifacts that are missing
// from the release. No-op when cache is disabled or everything is already remote.
func TryPushPackageCache(p *Package, h *host.Host) {
	if !CacheEnabled() {
		return
	}
	tag := CacheTag()
	if err := ensureRelease(tag); err != nil {
		log.Printf("[%s][%s] cache push: ensure release: %v", h.Triplet, p.Package, err)
		return
	}
	assets, err := loadRemoteAssets(tag)
	if err != nil {
		log.Printf("[%s][%s] cache push: list assets: %v", h.Triplet, p.Package, err)
		return
	}
	builtRoot := filepath.Join(host.DataDir(), "built")
	var uploadPaths []string
	var uploadNames []string
	for _, rel := range p.BuiltRelPaths(h) {
		name := AssetNameForRel(rel)
		if assets[name] {
			continue
		}
		abs := filepath.Join(builtRoot, filepath.FromSlash(rel))
		if _, err := os.Stat(abs); err != nil {
			continue
		}
		uploadPaths = append(uploadPaths, abs)
		uploadNames = append(uploadNames, name)
	}
	if len(uploadPaths) == 0 {
		return
	}
	if err := uploadAssets(tag, uploadPaths, uploadNames); err != nil {
		log.Printf("[%s][%s] cache push failed: %v", h.Triplet, p.Package, err)
		return
	}
	for _, name := range uploadNames {
		assets[name] = true
		log.Printf("[%s][%s] cache push: %s", h.Triplet, p.Package, name)
	}
}

func uploadAssets(tag string, uploadPaths, uploadNames []string) error {
	staging, err := os.MkdirTemp("", "simplybs-cache-upload-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(staging)

	staged := make([]string, 0, len(uploadPaths))
	for i, src := range uploadPaths {
		dst := filepath.Join(staging, uploadNames[i])
		if err := copyFile(src, dst); err != nil {
			return err
		}
		staged = append(staged, dst)
	}

	log.Printf("cache push: uploading %d file(s) to %s...", len(staged), tag)
	args := append([]string{"release", "upload", tag}, staged...)
	cmd := ghCmd(args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cache: upload: %w", err)
	}
	log.Printf("cache push: uploaded=%d tag=%s", len(staged), tag)
	return nil
}

// CachePush uploads local built artifacts that are not yet on the release.
// When pkgs is non-empty, only artifacts for those packages (and their deps)
// on host h are considered; otherwise every file under DataDir()/built is.
func CachePush(pkgs []*Package, h *host.Host) error {
	tag, _, err := requireCacheConfig()
	if err != nil {
		return err
	}
	if err := ensureRelease(tag); err != nil {
		return fmt.Errorf("cache: ensure release: %w", err)
	}
	resetRemoteAssetsCache()
	assets, err := loadRemoteAssets(tag)
	if err != nil {
		return fmt.Errorf("cache: list assets: %w", err)
	}

	builtRoot := filepath.Join(host.DataDir(), "built")
	var uploadPaths []string
	var uploadNames []string

	consider := func(absPath, name string) {
		if assets[name] {
			return
		}
		uploadPaths = append(uploadPaths, absPath)
		uploadNames = append(uploadNames, name)
	}

	if len(pkgs) > 0 && h != nil {
		for _, name := range neededAssetNames(pkgs, h) {
			rel := RelFromAssetName(name)
			abs := filepath.Join(builtRoot, filepath.FromSlash(rel))
			if _, err := os.Stat(abs); err != nil {
				continue
			}
			consider(abs, name)
		}
	} else {
		if _, err := os.Stat(builtRoot); os.IsNotExist(err) {
			log.Printf("cache: built dir not found: %s", builtRoot)
			return nil
		}
		err := filepath.WalkDir(builtRoot, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			rel, err := filepath.Rel(builtRoot, path)
			if err != nil {
				return err
			}
			consider(path, AssetNameForRel(rel))
			return nil
		})
		if err != nil {
			return err
		}
	}

	if len(uploadPaths) == 0 {
		log.Printf("cache push: nothing to upload (tag=%s remote-assets=%d)", tag, len(assets))
		return nil
	}

	if err := uploadAssets(tag, uploadPaths, uploadNames); err != nil {
		return err
	}
	for _, name := range uploadNames {
		assets[name] = true
	}
	return nil
}

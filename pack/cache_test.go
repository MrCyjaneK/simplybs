package pack

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mrcyjanek/simplybs/host"
)

func chdirRepoRoot(t *testing.T) {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(wd, "packages", "native", "_.json")); err == nil {
			if err := os.Chdir(wd); err != nil {
				t.Fatal(err)
			}
			return
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			t.Fatal("could not find repo root (packages/native/_.json)")
		}
		wd = parent
	}
}

func TestAssetNameRoundTrip(t *testing.T) {
	cases := []string{
		"zlib-1.3.1-deadbeef.tar.gz",
		"x86_64-linux-gnu/zlib-1.3.1-deadbeef.tar.gz",
		"x86_64-linux-gnu/native_foo/bar-1.0-abcd1234.info.txt",
	}
	for _, rel := range cases {
		name := AssetNameForRel(rel)
		got := RelFromAssetName(name)
		if filepath.ToSlash(got) != filepath.ToSlash(rel) {
			t.Fatalf("round-trip %q -> %q -> %q", rel, name, got)
		}
		if rel != filepath.Base(rel) && !containsPlus(name) {
			t.Fatalf("expected '+' in asset name for nested path, got %q", name)
		}
	}
}

func containsPlus(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '+' {
			return true
		}
	}
	return false
}

func TestCacheEnabledRequiresTagAndRepo(t *testing.T) {
	t.Setenv("SIMPLYBS_CACHE_TAG", "")
	t.Setenv("SIMPLYBS_CACHE_REPO", "")
	if CacheEnabled() {
		t.Fatal("expected cache disabled with neither set")
	}
	t.Setenv("SIMPLYBS_CACHE_TAG", "v0-sbs-test")
	if CacheEnabled() {
		t.Fatal("expected cache disabled with only TAG set")
	}
	t.Setenv("SIMPLYBS_CACHE_TAG", "")
	t.Setenv("SIMPLYBS_CACHE_REPO", "owner/repo")
	if CacheEnabled() {
		t.Fatal("expected cache disabled with only REPO set")
	}
	t.Setenv("SIMPLYBS_CACHE_TAG", "v0-sbs-test")
	t.Setenv("SIMPLYBS_CACHE_REPO", "owner/repo")
	if !CacheEnabled() {
		t.Fatal("expected cache enabled when both set")
	}
	if CacheTag() != "v0-sbs-test" || CacheRepo() != "owner/repo" {
		t.Fatalf("CacheTag/CacheRepo = %q / %q", CacheTag(), CacheRepo())
	}
}

func TestPackageCacheOnReleaseDisabled(t *testing.T) {
	t.Setenv("SIMPLYBS_CACHE_TAG", "")
	t.Setenv("SIMPLYBS_CACHE_REPO", "")
	ok, err := PackageCacheOnRelease(&Package{Package: "zlib", Version: "1.3.1"}, host.SupportedHosts["x86_64-linux-gnu"])
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Fatal("expected miss when cache is disabled")
	}
}

func TestRequireCacheConfig(t *testing.T) {
	t.Setenv("SIMPLYBS_CACHE_TAG", "")
	t.Setenv("SIMPLYBS_CACHE_REPO", "owner/repo")
	if _, _, err := requireCacheConfig(); err == nil {
		t.Fatal("expected error when TAG missing")
	}
	t.Setenv("SIMPLYBS_CACHE_TAG", "v0-sbs-test")
	t.Setenv("SIMPLYBS_CACHE_REPO", "")
	if _, _, err := requireCacheConfig(); err == nil {
		t.Fatal("expected error when REPO missing")
	}
	t.Setenv("SIMPLYBS_CACHE_REPO", "owner/repo")
	tag, repo, err := requireCacheConfig()
	if err != nil || tag != "v0-sbs-test" || repo != "owner/repo" {
		t.Fatalf("got tag=%q repo=%q err=%v", tag, repo, err)
	}
}

func TestBuiltRelPathsNativeAndHost(t *testing.T) {
	chdirRepoRoot(t)
	t.Setenv("SIMPLYBS_DATA_DIR", t.TempDir())
	pkg, err := FindPackage("zlib")
	if err != nil {
		t.Fatal(err)
	}
	h := host.SupportedHosts["x86_64-linux-gnu"]
	rels := pkg.BuiltRelPaths(h)
	if len(rels) != 3 {
		t.Fatalf("expected 3 paths, got %v", rels)
	}
	for _, rel := range rels {
		if filepath.ToSlash(rel) != rel {
			t.Fatalf("expected slash paths, got %q", rel)
		}
		if rel == filepath.Base(rel) {
			t.Fatalf("host package should be under triplet dir, got %q", rel)
		}
		name := AssetNameForRel(rel)
		if RelFromAssetName(name) != rel {
			t.Fatalf("asset encoding broken for %q", rel)
		}
	}

	native, err := FindPackage("native/zlib")
	if err != nil {
		t.Fatal(err)
	}
	nrels := native.BuiltRelPaths(h)
	for _, rel := range nrels {
		if rel != filepath.Base(rel) {
			t.Fatalf("native package should be flat under built/, got %q", rel)
		}
	}
}

func TestNeededAssetNamesIsSelective(t *testing.T) {
	chdirRepoRoot(t)
	t.Setenv("SIMPLYBS_DATA_DIR", t.TempDir())
	pkg, err := FindPackage("zlib")
	if err != nil {
		t.Fatal(err)
	}
	h := host.SupportedHosts["x86_64-linux-gnu"]
	names := neededAssetNames([]*Package{pkg}, h)
	if len(names) < 3 {
		t.Fatalf("expected at least root package assets, got %d", len(names))
	}
	// Push still considers the full transitive tree, but remain far below a full-world dump.
	all := GetAllPackages()
	world := neededAssetNames(all, h)
	if len(names) >= len(world) {
		t.Fatalf("zlib tree (%d) should be smaller than world (%d)", len(names), len(world))
	}
	if len(names)%3 != 0 {
		t.Fatalf("asset count should be multiple of 3, got %d", len(names))
	}
}

func packageNameSet(pkgs []*Package) map[string]bool {
	found := map[string]bool{}
	for _, p := range pkgs {
		found[p.Package] = true
	}
	return found
}

func remoteAssetsFor(pkgs []*Package, h *host.Host, include func(*Package) bool) map[string]bool {
	remote := map[string]bool{}
	for _, p := range pkgs {
		if include != nil && !include(p) {
			continue
		}
		for _, rel := range p.BuiltRelPaths(h) {
			remote[AssetNameForRel(rel)] = true
		}
	}
	return remote
}

func TestCollectCachePullPackagesStopsAtCachedRoot(t *testing.T) {
	chdirRepoRoot(t)
	t.Setenv("SIMPLYBS_DATA_DIR", t.TempDir())
	pkg, err := FindPackage("zlib")
	if err != nil {
		t.Fatal(err)
	}
	h := host.SupportedHosts["x86_64-linux-gnu"]
	full := CollectNeededPackages([]*Package{pkg}, h)
	if len(full) < 3 {
		t.Fatalf("expected zlib to have transitive deps, got %d", len(full))
	}
	remote := remoteAssetsFor(full, h, nil)
	got := collectCachePullPackages([]*Package{pkg}, h, remote)
	found := packageNameSet(got)
	if len(got) != 1 || !found["zlib"] {
		t.Fatalf("cached root should pull only itself, got %v", found)
	}
}

func TestCollectCachePullPackagesStopsAtCachedDirectDeps(t *testing.T) {
	chdirRepoRoot(t)
	t.Setenv("SIMPLYBS_DATA_DIR", t.TempDir())
	pkg, err := FindPackage("native/rust@1_95_0")
	if err != nil {
		t.Fatal(err)
	}
	h := host.SupportedHosts["x86_64-linux-gnu"]
	full := CollectNeededPackages([]*Package{pkg}, h)
	foundFull := packageNameSet(full)
	if !foundFull["native/rust@1_94_0"] || !foundFull["native/rust@1_93_0"] {
		t.Fatalf("expected rust bootstrap chain in full tree, got %v", foundFull)
	}

	direct := packageNameSet(filteredDependencyPackages(pkg.Dependencies, h))
	if direct["native/rust@1_93_0"] {
		t.Fatal("native/rust@1_93_0 should not be a direct dep of rust@1_95_0")
	}

	// Everything except the root is on the release: pull the root + direct
	// deps (including rust@1_94_0), but not older bootstraps.
	remote := remoteAssetsFor(full, h, func(p *Package) bool {
		return p.Package != pkg.Package
	})
	got := collectCachePullPackages([]*Package{pkg}, h, remote)
	found := packageNameSet(got)
	t.Logf("native/rust@1_95_0: full tree=%d packages, pull with cached deps=%d", len(full), len(got))
	if !found[pkg.Package] {
		t.Fatal("missing root package")
	}
	if !found["native/rust@1_94_0"] {
		t.Fatal("expected direct dep native/rust@1_94_0")
	}
	if found["native/rust@1_93_0"] || found["native/rust@1_92_0"] || found["native/rust@1_91_0"] {
		t.Fatalf("pulled rust bootstraps past a cache hit: %v", found)
	}
	for name := range direct {
		if !found[name] {
			t.Fatalf("missing direct dep %s", name)
		}
	}
}

func TestCollectCachePullPackagesWalksPastCacheMiss(t *testing.T) {
	chdirRepoRoot(t)
	t.Setenv("SIMPLYBS_DATA_DIR", t.TempDir())
	pkg, err := FindPackage("native/rust@1_95_0")
	if err != nil {
		t.Fatal(err)
	}
	h := host.SupportedHosts["x86_64-linux-gnu"]
	full := CollectNeededPackages([]*Package{pkg}, h)
	// rust@1_95_0 and rust@1_94_0 are both misses, so walk to rust@1_93_0.
	// rust@1_93_0 is cached, so rust@1_92_0 is not needed.
	skip := map[string]bool{
		"native/rust@1_95_0": true,
		"native/rust@1_94_0": true,
	}
	remote := remoteAssetsFor(full, h, func(p *Package) bool {
		return !skip[p.Package]
	})
	got := collectCachePullPackages([]*Package{pkg}, h, remote)
	found := packageNameSet(got)
	t.Logf("native/rust@1_95_0: full tree=%d, pull walking past rust@1_94_0 miss=%d", len(full), len(got))
	for _, name := range []string{"native/rust@1_95_0", "native/rust@1_94_0", "native/rust@1_93_0"} {
		if !found[name] {
			t.Fatalf("expected %s when walking past misses, got %v", name, found)
		}
	}
	if found["native/rust@1_92_0"] || found["native/rust@1_91_0"] {
		t.Fatalf("pulled rust bootstraps past rust@1_93_0 cache hit: %v", found)
	}
}

func TestCollectCachePullPackagesEmptyRemoteIsFullTree(t *testing.T) {
	chdirRepoRoot(t)
	t.Setenv("SIMPLYBS_DATA_DIR", t.TempDir())
	pkg, err := FindPackage("zlib")
	if err != nil {
		t.Fatal(err)
	}
	h := host.SupportedHosts["x86_64-linux-gnu"]
	full := packageNameSet(CollectNeededPackages([]*Package{pkg}, h))
	got := packageNameSet(collectCachePullPackages([]*Package{pkg}, h, map[string]bool{}))
	if len(got) != len(full) {
		t.Fatalf("empty remote should walk the full tree: pull=%d full=%d", len(got), len(full))
	}
	for name := range full {
		if !got[name] {
			t.Fatalf("missing %s from uncached pull walk", name)
		}
	}
}

func TestCollectCachePullPackagesStopsAtLocalComplete(t *testing.T) {
	chdirRepoRoot(t)
	t.Setenv("SIMPLYBS_DATA_DIR", t.TempDir())
	pkg, err := FindPackage("zlib")
	if err != nil {
		t.Fatal(err)
	}
	h := host.SupportedHosts["x86_64-linux-gnu"]
	builtRoot := filepath.Join(host.DataDir(), "built")
	for _, rel := range pkg.BuiltRelPaths(h) {
		path := filepath.Join(builtRoot, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("local"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	got := collectCachePullPackages([]*Package{pkg}, h, map[string]bool{})
	found := packageNameSet(got)
	if len(got) != 1 || !found["zlib"] {
		t.Fatalf("local complete root should not walk deps, got %v", found)
	}
}

func TestCachePullDownloadsOnlyNeededAssets(t *testing.T) {
	chdirRepoRoot(t)
	dataDir := t.TempDir()
	t.Setenv("SIMPLYBS_DATA_DIR", dataDir)
	t.Setenv("SIMPLYBS_CACHE_TAG", "v0-sbs-test-linux-amd64")
	t.Setenv("SIMPLYBS_CACHE_REPO", "owner/simplybs-cache")

	pkg, err := FindPackage("zlib")
	if err != nil {
		t.Fatal(err)
	}
	h := host.SupportedHosts["x86_64-linux-gnu"]
	full := neededAssetNames([]*Package{pkg}, h)
	if len(full) < 6 {
		t.Fatalf("expected a non-trivial zlib tree, got %d assets", len(full))
	}
	needed := make([]string, 0, cacheAssetSuffixes)
	for _, rel := range pkg.BuiltRelPaths(h) {
		needed = append(needed, AssetNameForRel(rel))
	}

	// Release contains the full zlib tree plus a large pile of unrelated
	// assets. Pull should stop at the cached root and not fetch deps.
	type asset struct {
		Name string `json:"name"`
	}
	payload := struct {
		Assets []asset `json:"assets"`
	}{}
	for _, name := range full {
		payload.Assets = append(payload.Assets, asset{Name: name})
	}
	for i := 0; i < 1000; i++ {
		payload.Assets = append(payload.Assets, asset{Name: fmt.Sprintf("unrelated-pkg-%d-deadbeef.tar.gz", i)})
	}
	assetsPath := filepath.Join(t.TempDir(), "assets.json")
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(assetsPath, raw, 0644); err != nil {
		t.Fatal(err)
	}

	logPath := filepath.Join(t.TempDir(), "gh.log")
	fakeGh := filepath.Join(t.TempDir(), "gh")
	script := `#!/bin/bash
set -euo pipefail
echo "$*" >> "$FAKE_GH_LOG"
cmd=$1; shift
case "$cmd" in
  -R) shift; cmd=$1; shift ;;
esac
case "$cmd" in
  release)
    sub=$1; shift
    case "$sub" in
      view) cat "$FAKE_GH_ASSETS" ;;
      download)
        pattern=""; dest="."
        while [[ $# -gt 0 ]]; do
          case "$1" in
            -p) pattern=$2; shift 2 ;;
            -D) dest=$2; shift 2 ;;
            *) shift ;;
          esac
        done
        echo "DOWNLOAD $pattern" >> "$FAKE_GH_LOG"
        printf 'ok' > "$dest/$pattern"
        ;;
      *) echo "bad sub $sub" >&2; exit 1 ;;
    esac
    ;;
  *) echo "bad $cmd" >&2; exit 1 ;;
esac
`
	if err := os.WriteFile(fakeGh, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SIMPLYBS_GH", fakeGh)
	t.Setenv("FAKE_GH_LOG", logPath)
	t.Setenv("FAKE_GH_ASSETS", assetsPath)

	resetRemoteAssetsCache()
	if err := CachePull([]*Package{pkg}, h); err != nil {
		t.Fatal(err)
	}

	logData, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	downloads := 0
	for _, line := range strings.Split(string(logData), "\n") {
		if strings.HasPrefix(line, "DOWNLOAD ") {
			downloads++
			name := strings.TrimPrefix(line, "DOWNLOAD ")
			if strings.HasPrefix(name, "unrelated-pkg-") {
				t.Fatalf("downloaded unrelated asset %q", name)
			}
		}
	}
	if downloads != len(needed) {
		t.Fatalf("downloaded %d assets, want %d needed", downloads, len(needed))
	}

	builtRoot := filepath.Join(host.DataDir(), "built")
	for _, name := range needed {
		dest := filepath.Join(builtRoot, filepath.FromSlash(RelFromAssetName(name)))
		if _, err := os.Stat(dest); err != nil {
			t.Fatalf("missing local cache file for %s: %v", name, err)
		}
	}

	// Second pull should not hit release download again.
	resetRemoteAssetsCache()
	if err := os.WriteFile(logPath, nil, 0644); err != nil {
		t.Fatal(err)
	}
	if err := CachePull([]*Package{pkg}, h); err != nil {
		t.Fatal(err)
	}
	logData, err = os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(logData), "\n") {
		if strings.HasPrefix(line, "DOWNLOAD ") {
			t.Fatalf("second pull re-downloaded %q", line)
		}
	}
}

func TestTryPushPackageCacheUploadsOnlyMissing(t *testing.T) {
	chdirRepoRoot(t)
	dataDir := t.TempDir()
	t.Setenv("SIMPLYBS_DATA_DIR", dataDir)
	t.Setenv("SIMPLYBS_CACHE_TAG", "v0-sbs-test-linux-amd64")
	t.Setenv("SIMPLYBS_CACHE_REPO", "owner/simplybs-cache")

	pkg, err := FindPackage("native/zlib")
	if err != nil {
		t.Fatal(err)
	}
	h := host.SupportedHosts["x86_64-linux-gnu"]
	rels := pkg.BuiltRelPaths(h)
	builtRoot := filepath.Join(host.DataDir(), "built")
	for _, rel := range rels {
		path := filepath.Join(builtRoot, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("payload-"+rel), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Remote already has the .info.txt; the two archives should upload.
	already := AssetNameForRel(rels[0])
	assetsPath := filepath.Join(t.TempDir(), "assets.json")
	if err := os.WriteFile(assetsPath, []byte(`{"assets":[{"name":"`+already+`"}]}`), 0644); err != nil {
		t.Fatal(err)
	}
	logPath := filepath.Join(t.TempDir(), "gh.log")
	fakeGh := filepath.Join(t.TempDir(), "gh")
	script := `#!/bin/bash
set -euo pipefail
echo "$*" >> "$FAKE_GH_LOG"
cmd=$1; shift
case "$cmd" in
  -R) shift; cmd=$1; shift ;;
esac
case "$cmd" in
  release)
    sub=$1; shift
    case "$sub" in
      view) cat "$FAKE_GH_ASSETS" ;;
      create) echo CREATE >> "$FAKE_GH_LOG" ;;
      upload)
        echo "UPLOAD $*" >> "$FAKE_GH_LOG"
        ;;
      *) echo "bad sub $sub" >&2; exit 1 ;;
    esac
    ;;
  *) echo "bad $cmd" >&2; exit 1 ;;
esac
`
	if err := os.WriteFile(fakeGh, []byte(script), 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SIMPLYBS_GH", fakeGh)
	t.Setenv("FAKE_GH_LOG", logPath)
	t.Setenv("FAKE_GH_ASSETS", assetsPath)

	resetRemoteAssetsCache()
	TryPushPackageCache(pkg, h)

	logData, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	logStr := string(logData)
	if !strings.Contains(logStr, "UPLOAD ") {
		t.Fatalf("expected upload, log=%s", logStr)
	}
	if strings.Contains(logStr, already) {
		t.Fatalf("should not re-upload existing asset %s; log=%s", already, logStr)
	}
	wantTar := AssetNameForRel(rels[1])
	wantNative := AssetNameForRel(rels[2])
	if !strings.Contains(logStr, wantTar) || !strings.Contains(logStr, wantNative) {
		t.Fatalf("expected upload of %s and %s; log=%s", wantTar, wantNative, logStr)
	}

	// Second push should upload nothing.
	if err := os.WriteFile(logPath, nil, 0644); err != nil {
		t.Fatal(err)
	}
	TryPushPackageCache(pkg, h)
	logData, err = os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(logData), "UPLOAD ") {
		t.Fatalf("second push should be a no-op, log=%s", logData)
	}
}

func TestTryPushDisabledWithoutCacheEnv(t *testing.T) {
	t.Setenv("SIMPLYBS_CACHE_TAG", "")
	t.Setenv("SIMPLYBS_CACHE_REPO", "")
	// Must not panic / call gh when disabled.
	TryPushPackageCache(&Package{Package: "zlib", Version: "1.0"}, host.SupportedHosts["x86_64-linux-gnu"])
}

func TestCollectNeededPackagesIncludesDeps(t *testing.T) {
	chdirRepoRoot(t)
	pkg, err := FindPackage("zlib")
	if err != nil {
		t.Fatal(err)
	}
	h := host.SupportedHosts["x86_64-linux-gnu"]
	needed := CollectNeededPackages([]*Package{pkg}, h)
	found := map[string]bool{}
	for _, p := range needed {
		found[p.Package] = true
	}
	if !found["zlib"] {
		t.Fatal("missing zlib")
	}
	if !found["native/make"] {
		t.Fatal("expected transitive native/make dependency")
	}
}

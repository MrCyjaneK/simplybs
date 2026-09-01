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
	// Must include transitive deps, but remain far below a full-world dump.
	all := GetAllPackages()
	world := neededAssetNames(all, h)
	if len(names) >= len(world) {
		t.Fatalf("zlib tree (%d) should be smaller than world (%d)", len(names), len(world))
	}
	if len(names)%3 != 0 {
		t.Fatalf("asset count should be multiple of 3, got %d", len(names))
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
	needed := neededAssetNames([]*Package{pkg}, h)
	if len(needed) < 6 {
		t.Fatalf("expected a non-trivial zlib tree, got %d assets", len(needed))
	}

	// Release contains needed assets plus a large pile of unrelated ones.
	type asset struct {
		Name string `json:"name"`
	}
	payload := struct {
		Assets []asset `json:"assets"`
	}{}
	for _, name := range needed {
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

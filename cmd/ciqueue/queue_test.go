package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mrcyjanek/simplybs/host"
	"github.com/mrcyjanek/simplybs/pack"
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
			t.Fatal("could not find repo root")
		}
		wd = parent
	}
}

func missCache(t *testing.T) cacheFn {
	t.Helper()
	return func(*pack.Package, *host.Host) (bool, error) { return false, nil }
}

func TestPackagesFromPaths(t *testing.T) {
	chdirRepoRoot(t)
	names := []string{}
	for _, p := range pack.GetAllPackages() {
		names = append(names, p.Package)
	}
	got := packagesFromPaths([]string{
		"packages/zlib.json",
		"patches/zlib/does-not-need-to-exist.patch",
		"patches/native/make/foo.patch",
		"README.md",
		"cmd/ciqueue/queue.go",
	}, names)
	for _, want := range []string{"zlib", "native/make"} {
		if !got[want] {
			t.Fatalf("expected %s in %v", want, sortedKeys(got))
		}
	}
	if got["cmd/ciqueue"] {
		t.Fatal("Go source should not map to a package")
	}
}

func TestNextQueueEmptyDiffIsDone(t *testing.T) {
	chdirRepoRoot(t)
	res, err := nextQueue(queueOpts{
		changedFiles: []string{"README.md", ".github/workflows/ci.yml"},
		hosts:        DefaultHosts,
		cached:       missCache(t),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != "done" {
		t.Fatalf("status=%s pkg=%s", res.Status, res.Package)
	}
}

func TestNextQueueZlibPicksNativeFirst(t *testing.T) {
	chdirRepoRoot(t)
	res, err := nextQueue(queueOpts{
		changedFiles: []string{"packages/zlib.json"},
		hosts:        DefaultHosts,
		cached:       missCache(t),
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != "next" {
		t.Fatalf("status=%s msg=%s", res.Status, res.Message)
	}
	if res.Package == "zlib" {
		t.Fatal("expected a native dependency before zlib")
	}
	pkg, err := pack.FindPackage(res.Package)
	if err != nil {
		t.Fatal(err)
	}
	if pkg.Type != "native" {
		t.Fatalf("expected native seed, got %s type=%s", res.Package, pkg.Type)
	}
	if res.Needed < 2 {
		t.Fatalf("expected zlib plus deps, needed=%d", res.Needed)
	}
}

func TestNextQueueZlibWhenDepsCached(t *testing.T) {
	chdirRepoRoot(t)
	res, err := nextQueue(queueOpts{
		changedFiles: []string{"packages/zlib.json"},
		hosts:        DefaultHosts,
		cached: func(p *pack.Package, h *host.Host) (bool, error) {
			return p.Package != "zlib", nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != "next" || res.Package != "zlib" {
		t.Fatalf("got %+v", res)
	}
	if res.Host != "x86_64-linux-gnu" {
		t.Fatalf("expected linux-gnu first, got %s", res.Host)
	}
}

func TestNextQueueZlibSecondHostAfterFirstCached(t *testing.T) {
	chdirRepoRoot(t)
	res, err := nextQueue(queueOpts{
		changedFiles: []string{"packages/zlib.json"},
		hosts:        DefaultHosts,
		cached: func(p *pack.Package, h *host.Host) (bool, error) {
			if p.Package != "zlib" {
				return true, nil
			}
			return h.Triplet == "x86_64-linux-gnu", nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != "next" || res.Package != "zlib" || res.Host != "aarch64-linux-gnu" {
		t.Fatalf("got %+v", res)
	}
}

func TestDefaultHostsAreLinuxCIHosts(t *testing.T) {
	skip := map[string]bool{"armv7a-linux-androideabi": true}
	if DefaultHosts[0] != "x86_64-linux-gnu" {
		t.Fatalf("linux-gnu should stay first, got %s", DefaultHosts[0])
	}
	seen := map[string]bool{}
	for i, h := range DefaultHosts {
		if skip[h] {
			t.Fatalf("DefaultHosts[%d]=%q should not be in Linux CI yet", i, h)
		}
		if host.SupportedHosts[h] == nil {
			t.Fatalf("DefaultHosts[%d]=%q is not in SupportedHosts", i, h)
		}
		if seen[h] {
			t.Fatalf("duplicate host %q", h)
		}
		seen[h] = true
	}
	for triplet := range host.SupportedHosts {
		if skip[triplet] {
			continue
		}
		if !seen[triplet] {
			t.Fatalf("SupportedHosts %q missing from DefaultHosts", triplet)
		}
	}
}

func TestPackagesYmlHostsMatchDefaultHosts(t *testing.T) {
	chdirRepoRoot(t)
	b, err := os.ReadFile(filepath.Join(".github", "workflows", "packages.yml"))
	if err != nil {
		t.Fatal(err)
	}
	want := "hosts: " + strings.Join(DefaultHosts, ",")
	if !bytes.Contains(b, []byte(want)) {
		t.Fatalf("packages.yml missing %q", want)
	}
	if n := bytes.Count(b, []byte(want)); n != 3 {
		t.Fatalf("expected 3 batch host lists, found %d", n)
	}
}

func TestNextQueueZlibDarwinHost(t *testing.T) {
	chdirRepoRoot(t)
	res, err := nextQueue(queueOpts{
		changedFiles: []string{"packages/zlib.json"},
		hosts:        []string{"aarch64-apple-darwin"},
		cached: func(p *pack.Package, h *host.Host) (bool, error) {
			return p.Package != "zlib", nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if res.Status != "next" || res.Package != "zlib" || res.Host != "aarch64-apple-darwin" {
		t.Fatalf("got %+v", res)
	}
}

func TestRenderCommentContainsMarker(t *testing.T) {
	body := renderComment(CommentState{
		SHA: "abc",
		Runs: []CommentRun{{
			Package:    "zlib",
			Host:       "x86_64-linux-gnu",
			Conclusion: "success",
			RunURL:     "https://example.test/run/1",
		}},
		Remaining: []Item{{Package: "curl", Host: "x86_64-linux-gnu"}},
	}, "")
	st, ok := parseState(body, commentMarker)
	if !ok {
		t.Fatalf("parse failed:\n%s", body)
	}
	if st.SHA != "abc" || len(st.Runs) != 1 || st.Runs[0].Package != "zlib" {
		t.Fatalf("%+v", st)
	}
	if len(st.Remaining) != 1 || st.Remaining[0].Package != "curl" {
		t.Fatalf("remaining %+v", st.Remaining)
	}
}

func TestRenderCommentQueueMarkersDoNotCollide(t *testing.T) {
	linux := renderComment(CommentState{SHA: "abc"}, "")
	macos := renderComment(CommentState{SHA: "abc"}, "macos")
	if !strings.Contains(linux, commentMarker) {
		t.Fatalf("linux comment missing default marker:\n%s", linux)
	}
	if !strings.Contains(macos, stateMarker("macos")) {
		t.Fatalf("macos comment missing marker:\n%s", macos)
	}
	if !strings.Contains(macos, "## simplybs package queue (macos)") {
		t.Fatalf("macos title missing:\n%s", macos)
	}
	if _, ok := parseState(macos, stateMarker("macos")); !ok {
		t.Fatalf("macos parser missed macos comment:\n%s", macos)
	}
	if _, ok := parseState(linux, stateMarker("macos")); ok {
		t.Fatal("macos parser matched linux comment")
	}
	if _, ok := parseState(macos, commentMarker); ok {
		t.Fatal("linux parser matched macos comment")
	}
}

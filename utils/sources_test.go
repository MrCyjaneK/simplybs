package utils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLoadSourcesFileBackwardCompat(t *testing.T) {
	fixture := `{
    "repositories": {
        "https://example.com/foo.git": {
            "refs": ["abc123"]
        }
    }
}`

	sources, err := SourcesFileFromRepositoriesOnlyJSON([]byte(fixture))
	if err != nil {
		t.Fatalf("unmarshal fixture: %v", err)
	}
	if len(sources.Repositories) != 1 {
		t.Fatalf("expected 1 repository, got %d", len(sources.Repositories))
	}
	if len(sources.Downloads) != 0 {
		t.Fatalf("expected no downloads, got %d", len(sources.Downloads))
	}

	repoRoot := filepath.Join("..", "sources.json")
	if _, err := os.Stat(repoRoot); err == nil {
		content, err := os.ReadFile(repoRoot)
		if err != nil {
			t.Fatalf("read sources.json: %v", err)
		}
		loaded, err := SourcesFileFromRepositoriesOnlyJSON(content)
		if err != nil {
			t.Fatalf("unmarshal sources.json: %v", err)
		}
		if len(loaded.Repositories) == 0 {
			t.Fatal("expected repositories in sources.json")
		}
	}
}

func TestIgnoredGitRefs(t *testing.T) {
	s := NewSourcesFile()

	if AddGitRef(s, "https://example.com/foo.git", "504e4711cee972ff751ab20ac8c259bbcaa22bb3") {
		t.Fatal("expected ignored ref to be skipped")
	}
	if _, exists := s.Repositories["https://example.com/foo.git"]; exists {
		t.Fatal("expected ignored ref not to create repository entry")
	}

	s.Repositories["https://example.com/foo.git"] = RepositoryInfo{
		Refs: []string{
			"abc123",
			"504e4711cee972ff751ab20ac8c259bbcaa22bb3",
			"c2b3d1e9ed841474b04671f06b0f005d03dede82",
		},
	}
	encoded, err := MarshalSourcesFile(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(encoded), "504e4711cee972ff751ab20ac8c259bbcaa22bb3") {
		t.Fatal("expected ignored ref to be filtered from marshaled output")
	}
	if strings.Contains(string(encoded), "c2b3d1e9ed841474b04671f06b0f005d03dede82") {
		t.Fatal("expected ignored ref to be filtered from marshaled output")
	}
	if !strings.Contains(string(encoded), "abc123") {
		t.Fatal("expected non-ignored ref to remain in marshaled output")
	}
}

func TestMergeGitRefDedup(t *testing.T) {
	s := NewSourcesFile()

	if !AddGitRef(s, "https://example.com/foo.git", "abc123") {
		t.Fatal("expected first ref to be added")
	}
	if AddGitRef(s, "https://example.com/foo.git", "abc123") {
		t.Fatal("expected duplicate ref to be ignored")
	}
	if len(s.Repositories["https://example.com/foo.git"].Refs) != 1 {
		t.Fatalf("expected 1 ref, got %d", len(s.Repositories["https://example.com/foo.git"].Refs))
	}
}

func TestMergeDownloadDedup(t *testing.T) {
	s := NewSourcesFile()

	entry := DownloadEntry{
		Kind:   "tar.gz",
		URL:    "https://example.com/foo.tar.gz",
		Sha256: "deadbeef",
	}
	if !AddDownload(s, entry) {
		t.Fatal("expected first download to be added")
	}
	if AddDownload(s, entry) {
		t.Fatal("expected duplicate download to be ignored")
	}

	other := DownloadEntry{
		Kind:   "tar.gz",
		URL:    "https://example.com/foo.tar.gz",
		Sha256: "cafebabe",
	}
	if !AddDownload(s, other) {
		t.Fatal("expected differing sha256 to be added")
	}
	if len(s.Downloads) != 2 {
		t.Fatalf("expected 2 downloads, got %d", len(s.Downloads))
	}
}

func TestSourcesFileRoundTrip(t *testing.T) {
	original := &SourcesFile{
		Repositories: map[string]RepositoryInfo{
			"https://example.com/foo.git": {
				Refs: []string{"abc123", "def456"},
			},
		},
		Downloads: []DownloadEntry{
			{
				Kind:   "tar.gz",
				URL:    "https://example.com/foo.tar.gz",
				Sha256: "deadbeef",
			},
		},
	}

	encoded, err := MarshalSourcesFile(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var decoded SourcesFile
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if !reflect.DeepEqual(original.Repositories, decoded.Repositories) {
		t.Fatalf("repositories mismatch:\noriginal: %#v\ndecoded: %#v", original.Repositories, decoded.Repositories)
	}
	if !reflect.DeepEqual(original.Downloads, decoded.Downloads) {
		t.Fatalf("downloads mismatch:\noriginal: %#v\ndecoded: %#v", original.Downloads, decoded.Downloads)
	}
}

func TestUniqueDownloads(t *testing.T) {
	entries := []DownloadEntry{
		{Kind: "blob", URL: "https://example.com/a.tar.xz", Sha256: "aaa", Path: ".build/a.tar.xz"},
		{Kind: "blob", URL: "https://example.com/a.tar.xz", Sha256: "aaa", Path: "work/a.tar.xz"},
		{Kind: "tar.gz", URL: "https://example.com/b.tar.gz", Sha256: "bbb"},
	}

	unique := UniqueDownloads(entries)
	if len(unique) != 2 {
		t.Fatalf("expected 2 unique downloads, got %d", len(unique))
	}
	if unique[0].URL != "https://example.com/a.tar.xz" || unique[1].URL != "https://example.com/b.tar.gz" {
		t.Fatalf("unexpected sort order: %#v", unique)
	}
}

func TestPackageGitRefsCollected(t *testing.T) {
	merged := NewSourcesFile()
	packagesDir := filepath.Join("..", "packages")
	err := filepath.WalkDir(packagesDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}

		pkgContent, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		var data map[string]interface{}
		if err := json.Unmarshal(pkgContent, &data); err != nil {
			return err
		}

		downloads, ok := data["download"].([]interface{})
		if !ok {
			return nil
		}

		for _, downloadInterface := range downloads {
			download, ok := downloadInterface.(map[string]interface{})
			if !ok {
				continue
			}
			kind, _ := download["kind"].(string)
			if kind != "git" {
				continue
			}
			url, _ := download["url"].(string)
			ref, _ := download["sha256"].(string)
			if url == "" || ref == "" {
				continue
			}

			MergeDownloadMap(merged, download)
			if !ContainsGitRef(merged, url, ref) {
				t.Fatalf("missing ref %s for %s after merging %s", ref, url, path)
			}
		}

		return nil
	})
	if err != nil {
		t.Fatalf("walk packages dir: %v", err)
	}
}

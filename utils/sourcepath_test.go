package utils

import (
	"path/filepath"
	"testing"
)

func TestSourcePathForGitURL(t *testing.T) {
	t.Setenv("SIMPLYBS_DATA_DIR", "/tmp/simplybs-test")

	path := SourcePathForGitURL("https://github.com/example/foo.git")
	expected := filepath.Join("/tmp/simplybs-test", "source", "github.com", "example", "foo.bundle")
	if path != expected {
		t.Fatalf("expected %q, got %q", expected, path)
	}
}

func TestSourcePathForFileURL(t *testing.T) {
	t.Setenv("SIMPLYBS_DATA_DIR", "/tmp/simplybs-test")

	path := SourcePathForFileURL("https://example.com/releases/foo.tar.gz")
	expected := filepath.Join("/tmp/simplybs-test", "source", "example.com", "releases", "foo.tar.gz")
	if path != expected {
		t.Fatalf("expected %q, got %q", expected, path)
	}
}

package download

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureDownloadFileSkipsValidFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	content := []byte("hello")
	if err := os.WriteFile(path, content, 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	expectedHash, err := FileSHA256(path)
	if err != nil {
		t.Fatalf("hash file: %v", err)
	}

	if err := EnsureDownloadFile("test", path, "https://example.com/test.txt", expectedHash); err != nil {
		t.Fatalf("expected skip, got error: %v", err)
	}
}

func TestEnsureDownloadFileRejectsInvalidHash(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(path, []byte("hello"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	err := EnsureDownloadFile("test", path, "https://invalid.test/no-such-file", "0000000000000000000000000000000000000000000000000000000000000000")
	if err == nil {
		t.Fatal("expected download error for invalid hash and unreachable URL")
	}

	if _, statErr := os.Stat(path); statErr == nil {
		t.Fatal("expected stale file to be removed before re-download attempt")
	}
}

func TestShouldResetRetryCount(t *testing.T) {
	dir := t.TempDir()
	partial := filepath.Join(dir, "partial.bin")

	if shouldResetRetryCount(fmt.Errorf("network"), partial, 0, 0) {
		t.Fatal("expected no reset without progress")
	}

	if err := os.WriteFile(partial, make([]byte, 128), 0644); err != nil {
		t.Fatalf("write partial: %v", err)
	}
	if !shouldResetRetryCount(fmt.Errorf("network"), partial, 0, 128) {
		t.Fatal("expected reset when partial file remains")
	}

	if err := os.Remove(partial); err != nil {
		t.Fatalf("remove partial: %v", err)
	}
	if shouldResetRetryCount(fmt.Errorf("%w: expected x, got y", ErrHashMismatch), partial, 0, 128) {
		t.Fatal("hash mismatch must not reset retry count")
	}
}

package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type ProgressWriter struct {
	writer     io.Writer
	total      int64
	written    int64
	lastUpdate time.Time
	filename   string
}

func NewProgressWriter(writer io.Writer, total int64, filename string) *ProgressWriter {
	return &ProgressWriter{
		writer:     writer,
		total:      total,
		filename:   filename,
		lastUpdate: time.Now(),
	}
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	if err != nil {
		return n, err
	}

	pw.written += int64(n)

	if time.Since(pw.lastUpdate) > 100*time.Millisecond {
		pw.displayProgress()
		pw.lastUpdate = time.Now()
	}

	return n, err
}

func (pw *ProgressWriter) displayProgress() {
	if pw.total <= 0 {
		fmt.Printf("\r%s: Downloaded %s", pw.filename, formatBytes(pw.written))
	} else {
		percentage := float64(pw.written) / float64(pw.total) * 100
		fmt.Printf("\r%s: %.1f%% (%s / %s)", pw.filename, percentage, formatBytes(pw.written), formatBytes(pw.total))
	}
}

func (pw *ProgressWriter) finish() {
	fmt.Printf("\r%s: Complete (%s)\n", pw.filename, formatBytes(pw.written))
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp%len("KMGTPE")])
}

func DownloadFile(path, url, expectedSha256 string, isMirror bool) error {
	log.Printf("Downloading %s to %s", url, path)

	if !isMirror {
		baseName := filepath.Base(path)
		err := DownloadFile(path, "https://static.mrcyjanek.net/lfs/simplybs/sources/"+baseName, expectedSha256, true)
		if err != nil {
			log.Printf("Failed to download file from mirror: %v, trying original URL", err)
		} else {
			log.Printf("Downloaded file from mirror: %s", path)
			return nil
		}
	}

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("Failed to download file from %s: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Failed to download file: HTTP %d %s", resp.StatusCode, resp.Status)
	}

	var totalSize int64
	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		if size, err := strconv.ParseInt(contentLength, 10, 64); err == nil {
			totalSize = size
		}
	}

	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("Failed to create file %s: %v", path, err)
	}
	defer out.Close()

	hasher := sha256.New()

	filename := path
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		filename = path[idx+1:]
	}

	progressWriter := NewProgressWriter(out, totalSize, filename)
	multiWriter := io.MultiWriter(progressWriter, hasher)

	_, err = io.Copy(multiWriter, resp.Body)
	if err != nil {
		return fmt.Errorf("Failed to write file %s: %v", path, err)
	}

	progressWriter.finish()

	actualHash := hex.EncodeToString(hasher.Sum(nil))
	if actualHash != expectedSha256 {
		os.Remove(path)
		return fmt.Errorf("SHA256 hash mismatch for %s: expected %s, got %s", path, expectedSha256, actualHash)
	}

	log.Printf("Successfully downloaded and verified %s", path)
	return nil
}

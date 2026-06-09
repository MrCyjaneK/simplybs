package download

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
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

func GetMirrors() []string {
	mirrors := []string{
		"http://static.mrcyjanek.net/lfs/simplybs/source/",
	}
	if customMirror := os.Getenv("SIMPLYBS_MIRROR"); customMirror != "" {
		mirrors = append([]string{customMirror}, mirrors...)
	}
	return mirrors
}

func URLToPath(urlStr string) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %v", err)
	}

	path := strings.TrimPrefix(parsedURL.Path, "/")
	return filepath.Join(parsedURL.Host, path), nil
}

func DownloadFile(packageName, path, url, expectedSha256 string, isMirror bool) error {
	log.Printf("Downloading %s to %s", url, path)

	if !isMirror {
		urlPath, err := URLToPath(url)
		if err != nil {
			log.Printf("Failed to convert URL to path: %v", err)
		} else {
			mirrors := GetMirrors()
			for _, mirror := range mirrors {
				mirrorURL := mirror + urlPath
				err := DownloadFile(packageName, path, mirrorURL, expectedSha256, true)
				if err != nil {
					log.Printf("Failed to download file from mirror %s: %v", mirror, err)
					continue
				}
				log.Printf("Downloaded file from mirror: %s", path)
				return nil
			}
			log.Printf("All mirrors failed, trying original URL")
		}
	}

	os.MkdirAll(filepath.Dir(path), 0755)

	maxRetries := 15
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			log.Printf("Retry attempt %d/%d for %s", attempt, maxRetries, url)
			time.Sleep(5 * time.Second)
		}

		bytesReceived, err := downloadWithResume(path, url, expectedSha256)
		if err != nil {
			// If no data was received on first try, give up immediately
			if bytesReceived == 0 && attempt == 0 {
				return fmt.Errorf("no data received from server: %v", err)
			}
			if attempt == maxRetries {
				return fmt.Errorf("failed after %d retries: %v", maxRetries, err)
			}
			log.Printf("Download failed: %v", err)
			continue
		}

		log.Printf("Successfully downloaded and verified %s", path)
		return nil
	}

	return fmt.Errorf("download failed after all retries")
}

func downloadWithResume(path, url, expectedSha256 string) (int64, error) {
	var fileSize int64
	var out *os.File
	var err error

	// Check if partial file exists
	if info, err := os.Stat(path); err == nil {
		fileSize = info.Size()
		out, err = os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return 0, fmt.Errorf("failed to open file for append: %v", err)
		}
	} else {
		out, err = os.Create(path)
		if err != nil {
			return 0, fmt.Errorf("failed to create file: %v", err)
		}
	}
	defer out.Close()

	// Create request with Range header for resume
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %v", err)
	}

	if fileSize > 0 {
		req.Header.Set("Range", fmt.Sprintf("bytes=%d-", fileSize))
		log.Printf("Resuming download from byte %d", fileSize)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to download file: %v", err)
	}
	defer resp.Body.Close()

	// Accept both 200 (full content) and 206 (partial content)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return 0, fmt.Errorf("HTTP %d %s", resp.StatusCode, resp.Status)
	}

	// If server doesn't support resume (200 instead of 206), start over
	if resp.StatusCode == http.StatusOK && fileSize > 0 {
		out.Close()
		out, err = os.Create(path)
		if err != nil {
			return 0, fmt.Errorf("failed to recreate file: %v", err)
		}
		fileSize = 0
		log.Printf("Server doesn't support resume, starting from beginning")
	}

	var totalSize int64
	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		if size, err := strconv.ParseInt(contentLength, 10, 64); err == nil {
			totalSize = fileSize + size
		}
	}

	filename := filepath.Base(path)
	progressWriter := NewProgressWriter(out, totalSize, filename)

	bytesRead, err := io.Copy(progressWriter, resp.Body)
	if err != nil {
		return bytesRead, fmt.Errorf("failed to write file: %v", err)
	}

	progressWriter.finish()

	// Verify hash only if download is complete
	file, err := os.Open(path)
	if err != nil {
		return bytesRead, fmt.Errorf("failed to open file for verification: %v", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return bytesRead, fmt.Errorf("failed to calculate hash: %v", err)
	}

	actualHash := hex.EncodeToString(hasher.Sum(nil))
	if actualHash != expectedSha256 {
		os.Remove(path)
		return bytesRead, fmt.Errorf("SHA256 hash mismatch: expected %s, got %s", expectedSha256, actualHash)
	}

	return bytesRead, nil
}

func FileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func EnsureDownloadFile(packageName, path, url, expectedSha256 string) error {
	if actualHash, err := FileSHA256(path); err == nil {
		if actualHash == expectedSha256 {
			log.Printf("File already exists with correct hash: %s", path)
			return nil
		}
		log.Printf("File exists but hash mismatch (expected %s, got %s), re-downloading: %s", expectedSha256, actualHash, path)
		os.Remove(path)
	}

	return DownloadFile(packageName, path, url, expectedSha256, false)
}

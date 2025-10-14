package utils

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func verifyBundleHasRef(bundlePath, ref string) bool {
	// Try to verify the bundle contains our ref by attempting to fetch it
	tempVerifyDir := bundlePath + ".verify.tmp"
	defer os.RemoveAll(tempVerifyDir)

	os.MkdirAll(tempVerifyDir, 0755)

	// Initialize a temporary git repo
	initCmd := exec.Command("git", "init")
	initCmd.Dir = tempVerifyDir
	if err := initCmd.Run(); err != nil {
		return false
	}

	// Try to fetch the specific ref from the bundle
	fetchCmd := exec.Command("git", "fetch", bundlePath, ref)
	fetchCmd.Dir = tempVerifyDir
	if err := fetchCmd.Run(); err != nil {
		log.Printf("Bundle does not contain ref %s", ref)
		return false
	}

	log.Printf("Bundle contains ref %s", ref)
	return true
}

func downloadBundleFromMirror(bundlePath, ref string) error {
	mirrorURL := "http://static.mrcyjanek.net/lfs/simplybs/source/" + filepath.Base(bundlePath)
	log.Printf("Trying to download bundle from mirror: %s", mirrorURL)

	resp, err := http.Get(mirrorURL)
	if err != nil {
		return fmt.Errorf("failed to download from mirror: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("mirror returned HTTP %d", resp.StatusCode)
	}

	tempFile := bundlePath + ".download.tmp"
	defer os.Remove(tempFile)

	out, err := os.Create(tempFile)
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}

	_, err = out.ReadFrom(resp.Body)
	out.Close()
	if err != nil {
		return fmt.Errorf("failed to download bundle: %w", err)
	}

	if !verifyBundleHasRef(tempFile, ref) {
		return fmt.Errorf("mirror bundle does not contain ref %s", ref)
	}

	os.MkdirAll(filepath.Dir(bundlePath), 0755)
	if err := os.Rename(tempFile, bundlePath); err != nil {
		return fmt.Errorf("failed to move bundle to final location: %w", err)
	}

	log.Printf("Successfully downloaded bundle from mirror")
	return nil
}

func createBundleFromRepo(bundlePath, url, ref string) error {
	log.Printf("Creating bundle from repository %s", url)

	tempDir := bundlePath + ".clone.tmp"
	defer os.RemoveAll(tempDir)

	log.Printf("Cloning repository from %s", url)
	cloneCmd := exec.Command("git", "clone", url, tempDir)
	cloneCmd.Stdout = os.Stdout
	cloneCmd.Stderr = os.Stderr
	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	log.Printf("Checking out reference %s", ref)
	checkoutCmd := exec.Command("git", "checkout", ref)
	checkoutCmd.Dir = tempDir
	checkoutCmd.Stdout = os.Stdout
	checkoutCmd.Stderr = os.Stderr
	if err := checkoutCmd.Run(); err != nil {
		return fmt.Errorf("failed to checkout ref %s: %w", ref, err)
	}

	revParseCmd := exec.Command("git", "rev-parse", "HEAD")
	revParseCmd.Dir = tempDir
	commitHash, err := revParseCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to get commit hash: %w", err)
	}
	log.Printf("Checked out commit: %s", strings.TrimSpace(string(commitHash)))

	log.Printf("Creating bundle file...")
	os.MkdirAll(filepath.Dir(bundlePath), 0755)
	bundleCmd := exec.Command("git", "bundle", "create", bundlePath, "HEAD")
	bundleCmd.Dir = tempDir
	bundleCmd.Stdout = os.Stdout
	bundleCmd.Stderr = os.Stderr
	if err := bundleCmd.Run(); err != nil {
		return fmt.Errorf("failed to create bundle: %w", err)
	}

	log.Printf("Successfully created bundle at %s", bundlePath)
	return nil
}

func DownloadGit(packageName, bundlePath, url, ref string) error {
	log.Printf("Downloading %s to bundle %s", url, bundlePath)

	if _, err := os.Stat(bundlePath); err == nil {
		log.Printf("Bundle already exists, verifying it contains ref %s", ref)
		if verifyBundleHasRef(bundlePath, ref) {
			log.Printf("Bundle is valid and contains the required ref")
			return nil
		}
		log.Printf("Bundle does not contain ref %s, removing and re-downloading", ref)
		os.Remove(bundlePath)
	}

	err := downloadBundleFromMirror(bundlePath, ref)
	if err != nil {
		log.Printf("Failed to download from mirror: %v, creating bundle from original URL", err)
		return createBundleFromRepo(bundlePath, url, ref)
	}

	return nil
}

func ExtractGitCloneBundle(bundlePath, destPath string) error {
	log.Printf("Extracting git bundle from %s to %s", bundlePath, destPath)

	// Ensure destination directory exists
	os.MkdirAll(destPath, 0755)

	// Clone from bundle
	cloneCmd := exec.Command("git", "clone", bundlePath, destPath)
	cloneCmd.Stdout = os.Stdout
	cloneCmd.Stderr = os.Stderr
	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("failed to clone from bundle: %w", err)
	}

	log.Printf("Successfully extracted bundle to %s", destPath)
	return nil
}

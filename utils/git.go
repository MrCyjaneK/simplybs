package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mrcyjanek/simplybs/utils/download"
)

func verifyBundleHasRef(bundlePath string, refs []string) bool {
	tempVerifyDir := bundlePath + ".verify.tmp"
	defer os.RemoveAll(tempVerifyDir)

	os.MkdirAll(tempVerifyDir, 0755)

	// Initialize a temporary git repo
	initCmd := exec.Command("git", "init")
	initCmd.Dir = tempVerifyDir
	if err := initCmd.Run(); err != nil {
		return false
	}

	for _, ref := range refs {
		fetchCmd := exec.Command("git", "fetch", bundlePath, ref)
		fetchCmd.Dir = tempVerifyDir
		if err := fetchCmd.Run(); err != nil {
			log.Printf("Bundle does not contain ref %s", ref)
			return false
		}
		log.Printf("Bundle contains ref %s", ref)
	}

	return true
}

func downloadBundleFromMirrors(bundlePath, originalURL string, refs []string) error {
	urlPath, err := download.URLToPath(originalURL)
	if err != nil {
		return fmt.Errorf("failed to convert URL to path: %w", err)
	}

	bundleFilename := filepath.Base(bundlePath)
	mirrorPath := filepath.Join(filepath.Dir(urlPath), bundleFilename)

	mirrors := download.GetMirrors()
	for _, mirror := range mirrors {
		mirrorURL := mirror + mirrorPath
		log.Printf("Trying to download bundle from mirror: %s", mirrorURL)

		resp, err := http.Get(mirrorURL)
		if err != nil {
			log.Printf("Failed to download from mirror %s: %v", mirror, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Mirror %s returned HTTP %d", mirror, resp.StatusCode)
			continue
		}

		tempFile := bundlePath + ".download.tmp"
		defer os.Remove(tempFile)

		out, err := os.Create(tempFile)
		if err != nil {
			log.Printf("Failed to create temp file: %v", err)
			continue
		}

		_, err = out.ReadFrom(resp.Body)
		out.Close()
		if err != nil {
			log.Printf("Failed to download bundle: %v", err)
			continue
		}

		if !verifyBundleHasRef(tempFile, refs) {
			log.Printf("Mirror bundle does not contain all refs %v", refs)
			continue
		}

		os.MkdirAll(filepath.Dir(bundlePath), 0755)
		if err := os.Rename(tempFile, bundlePath); err != nil {
			log.Printf("Failed to move bundle to final location: %v", err)
			continue
		}

		log.Printf("Successfully downloaded bundle from mirror %s", mirror)
		return nil
	}

	return fmt.Errorf("all mirrors failed")
}

func createBundleFromRepo(bundlePath, url string, refs []string) error {
	log.Printf("Creating bundle from repository %s with %d refs", url, len(refs))

	tempDir := bundlePath + ".clone.tmp"
	defer os.RemoveAll(tempDir)

	log.Printf("Cloning repository from %s", url)
	cloneCmd := exec.Command("git", "clone", url, tempDir)
	cloneCmd.Stdout = os.Stdout
	cloneCmd.Stderr = os.Stderr
	if err := cloneCmd.Run(); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	log.Printf("Creating branches for refs: %v", refs)
	os.MkdirAll(filepath.Dir(bundlePath), 0755)

	branchNames := []string{}
	for i, ref := range refs {
		branchName := fmt.Sprintf("bundle-ref-%d", i)
		branchNames = append(branchNames, branchName)

		branchCmd := exec.Command("git", "branch", branchName, ref)
		branchCmd.Dir = tempDir
		branchCmd.Stdout = os.Stdout
		branchCmd.Stderr = os.Stderr
		if err := branchCmd.Run(); err != nil {
			log.Printf("Warning: failed to create branch %s for ref %s: %v", branchName, ref, err)
			branchNames[len(branchNames)-1] = ref
		}
	}

	bundleArgs := append([]string{"bundle", "create", bundlePath}, branchNames...)
	bundleCmd := exec.Command("git", bundleArgs...)
	bundleCmd.Dir = tempDir
	bundleCmd.Stdout = os.Stdout
	bundleCmd.Stderr = os.Stderr
	if err := bundleCmd.Run(); err != nil {
		return fmt.Errorf("failed to create bundle: %w", err)
	}

	log.Printf("Successfully created bundle at %s with %d refs", bundlePath, len(refs))
	return nil
}

func DownloadGit(packageName, bundlePath, url string) error {
	log.Printf("Downloading %s to bundle %s", url, bundlePath)

	sources, err := loadSourcesFile()
	if err != nil {
		log.Fatalf("Failed to load sources.json: %v, falling back to single ref mode", err)
	}

	refs, exists := sources.Repositories[url]
	if !exists {
		log.Fatalf("Repository %s not found in sources.json", url)
	}

	if _, err := os.Stat(bundlePath); err == nil {
		log.Printf("Bundle already exists, verifying it contains all refs %v", refs.Refs)
		if verifyBundleHasRef(bundlePath, refs.Refs) {
			log.Printf("Bundle is valid and contains all required refs")
			return nil
		}
		log.Printf("Bundle does not contain all refs %v, removing and re-downloading", refs.Refs)
		os.Remove(bundlePath)
	}

	err = downloadBundleFromMirrors(bundlePath, url, refs.Refs)
	if err != nil {
		log.Printf("Failed to download from mirrors: %v, creating bundle from original URL", err)
		return createBundleFromRepo(bundlePath, url, refs.Refs)
	}

	return nil
}

type SourcesFile struct {
	Repositories map[string]RepositoryInfo `json:"repositories"`
}

type RepositoryInfo struct {
	Refs     []string `json:"refs"`
	Packages []string `json:"packages"`
}

func loadSourcesFile() (*SourcesFile, error) {
	content, err := os.ReadFile("sources.json")
	if err != nil {
		return nil, err
	}

	var sources SourcesFile
	if err := json.Unmarshal(content, &sources); err != nil {
		return nil, err
	}

	return &sources, nil
}

func ExtractGitCloneBundle(bundlePath, destPath, ref string) error {
	log.Printf("Extracting git bundle from %s to %s", bundlePath, destPath)

	os.MkdirAll(destPath, 0755)

	steps := [][]string{
		{"git", "init"},
		{"git", "fetch", bundlePath, ref},
		{"git", "checkout", "--force", ref},
	}

	for i, step := range steps {
		cmd := exec.Command(step[0], step[1:]...)
		cmd.Dir = destPath
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			// If fetch failed, the bundle doesn't have this ref
			// Try to get the repo URL from the bundle path and re-create it
			if i == 1 { // fetch step failed
				log.Printf("Bundle does not contain ref %s, attempting to re-create bundle with updated refs", ref)
				os.RemoveAll(destPath) // Clean up failed extraction

				// Get the repository URL from sources.json by finding which repo this bundle belongs to
				url, err := getRepoURLFromBundlePath(bundlePath)
				if err != nil {
					return fmt.Errorf("failed to find repository URL for bundle: %w", err)
				}

				// Remove old bundle and re-create with all current refs
				log.Printf("Removing old bundle: %s", bundlePath)
				os.Remove(bundlePath)

				// Load sources and re-create bundle
				sources, err := loadSourcesFile()
				if err != nil {
					return fmt.Errorf("failed to load sources.json: %w", err)
				}

				repoInfo, exists := sources.Repositories[url]
				if !exists {
					return fmt.Errorf("repository %s not found in sources.json", url)
				}

				log.Printf("Re-creating bundle with %d refs from sources.json", len(repoInfo.Refs))
				if err := createBundleFromRepo(bundlePath, url, repoInfo.Refs); err != nil {
					return fmt.Errorf("failed to re-create bundle: %w", err)
				}

				// Now retry the extraction
				return ExtractGitCloneBundle(bundlePath, destPath, ref)
			}
			return fmt.Errorf("failed to run command: %w", err)
		}
	}
	log.Printf("Successfully extracted bundle to %s", destPath)
	return nil
}

func getRepoURLFromBundlePath(bundlePath string) (string, error) {
	sources, err := loadSourcesFile()
	if err != nil {
		return "", err
	}

	// Extract the URL pattern from the bundle path
	// Bundle path format: .buildlib/source/<url-path>.bundle
	bundleName := filepath.Base(bundlePath)
	bundleName = strings.TrimSuffix(bundleName, ".bundle")

	// Try to match against known repositories
	for url := range sources.Repositories {
		urlPath, err := download.URLToPath(url)
		if err != nil {
			continue
		}
		urlPath = strings.TrimSuffix(urlPath, ".git")
		if strings.HasSuffix(urlPath, bundleName) || strings.Contains(bundlePath, urlPath) {
			return url, nil
		}
	}

	return "", fmt.Errorf("could not find repository URL for bundle path: %s", bundlePath)
}

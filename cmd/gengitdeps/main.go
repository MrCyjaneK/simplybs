package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Download struct {
	Kind   string `json:"kind"`
	Path   string `json:"path,omitempty"`
	SHA256 string `json:"sha256"`
	URL    string `json:"url"`
}

type Output struct {
	Download []Download `json:"download"`
}

func main() {
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	var downloads []Download

	// Walk the directory tree
	err = filepath.Walk(currentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Check if this is a .git directory (or symlink to one)
		if info.Name() == ".git" {
			// Check if it's a directory or a symlink to a directory
			isGitDir := info.IsDir()
			if !isGitDir && info.Mode()&os.ModeSymlink != 0 {
				// It's a symlink, check if it points to a directory
				target, err := os.Stat(path)
				if err == nil && target.IsDir() {
					isGitDir = true
				}
			}

			if !isGitDir {
				return filepath.SkipDir
			}

			// Get the repository directory (parent of .git)
			repoDir := filepath.Dir(path)

			// Get git commit hash
			commitHash, err := getGitCommit(repoDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not get commit for %s: %v\n", repoDir, err)
				return filepath.SkipDir
			}

			// Get git remote URL
			remoteURL, err := getGitRemoteURL(repoDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not get remote URL for %s: %v\n", repoDir, err)
				return filepath.SkipDir
			}

			// Calculate relative path from current directory
			relPath, err := filepath.Rel(currentDir, repoDir)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Warning: Could not get relative path for %s: %v\n", repoDir, err)
				return filepath.SkipDir
			}

			download := Download{
				Kind:   "git",
				SHA256: commitHash,
				URL:    remoteURL,
			}

			// Add path only if it's not the current directory
			if relPath != "." {
				download.Path = relPath
			}

			downloads = append(downloads, download)

			// Skip further traversal into this .git directory
			return filepath.SkipDir
		}

		return nil
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking directory: %v\n", err)
		os.Exit(1)
	}

	// Output JSON
	output := Output{Download: downloads}
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "    ")
	if err := encoder.Encode(output); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding JSON: %v\n", err)
		os.Exit(1)
	}
}

func getGitCommit(repoDir string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoDir
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func getGitRemoteURL(repoDir string) (string, error) {
	// Try origin first
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	cmd.Dir = repoDir
	output, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output)), nil
	}

	// If origin doesn't exist, get the first available remote
	cmd = exec.Command("git", "remote")
	cmd.Dir = repoDir
	output, err = cmd.Output()
	if err != nil {
		return "", err
	}

	remotes := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(remotes) == 0 || remotes[0] == "" {
		return "", fmt.Errorf("no remotes found")
	}

	// Get the URL of the first remote
	firstRemote := remotes[0]
	cmd = exec.Command("git", "config", "--get", fmt.Sprintf("remote.%s.url", firstRemote))
	cmd.Dir = repoDir
	output, err = cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

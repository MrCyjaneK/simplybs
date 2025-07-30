package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

const (
	binaryBaseName = "simplybs"
	releaseVersion = "v0.1.0"
	releaseDir     = "release"

	dirPermissions = 0755
)

type Target struct {
	GOOS   string
	GOARCH string
}

func main() {
	targets := []Target{
		{"linux", "amd64"},
		{"linux", "arm64"},
		{"darwin", "arm64"},
	}

	if err := createReleaseDirectory(); err != nil {
		log.Fatalf("Failed to create release directory: %v", err)
	}

	for _, target := range targets {
		fmt.Printf("\n=== Building for %s/%s ===\n", target.GOOS, target.GOARCH)

		if err := buildAndPackage(target); err != nil {
			log.Fatalf("Failed to build and package for %s/%s: %v", target.GOOS, target.GOARCH, err)
		}

		fmt.Printf("Successfully packaged %s/%s\n", target.GOOS, target.GOARCH)
	}

	fmt.Println("\nAll releases built successfully!")
}

func createReleaseDirectory() error {
	return os.MkdirAll(releaseDir, dirPermissions)
}

func buildAndPackage(target Target) error {
	binaryName := generateBinaryName(target)

	if err := buildBinary(target, binaryName); err != nil {
		return fmt.Errorf("failed to build binary: %w", err)
	}
	defer cleanupBinary(binaryName)

	archiveName := generateArchiveName(target)
	if err := createArchive(archiveName, binaryName); err != nil {
		return fmt.Errorf("failed to create archive: %w", err)
	}

	return nil
}

func generateBinaryName(target Target) string {
	binaryName := fmt.Sprintf("%s_%s_%s", binaryBaseName, target.GOOS, target.GOARCH)
	if target.GOOS == "windows" {
		binaryName += ".exe"
	}
	return binaryName
}

func generateArchiveName(target Target) string {
	return fmt.Sprintf("%s/%s_%s_simplybs_%s.tar.gz",
		releaseDir, target.GOOS, target.GOARCH, releaseVersion)
}

func buildBinary(target Target, binaryName string) error {
	fmt.Printf("Building binary: %s\n", binaryName)

	cmd := exec.Command("go", "build", "-o", binaryName, ".")
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GOOS=%s", target.GOOS),
		fmt.Sprintf("GOARCH=%s", target.GOARCH),
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func cleanupBinary(binaryName string) {
	if err := os.Remove(binaryName); err != nil {
		log.Printf("Warning: Failed to remove binary %s: %v", binaryName, err)
	}
}

func createArchive(filename, binaryName string) error {
	fmt.Printf("Creating archive: %s\n", filename)

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create archive file: %w", err)
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	archiveItems := getArchiveItems(binaryName)

	for _, item := range archiveItems {
		if err := addToArchive(tarWriter, item); err != nil {
			return fmt.Errorf("failed to add %s to archive: %w", item, err)
		}
		fmt.Printf("  Added: %s\n", item)
	}

	return nil
}

func getArchiveItems(binaryName string) []string {
	return []string{
		binaryName,
		"patches/",
		"packages/",
		"README.md",
	}
}

func addToArchive(tarWriter *tar.Writer, path string) error {
	return filepath.WalkDir(path, func(file string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error accessing path %s: %w", file, err)
		}

		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("failed to get file info for %s: %w", file, err)
		}

		header, err := createTarHeader(info, file)
		if err != nil {
			return fmt.Errorf("failed to create tar header for %s: %w", file, err)
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			return fmt.Errorf("failed to write tar header for %s: %w", file, err)
		}

		if info.Mode().IsRegular() {
			if err := writeFileToArchive(tarWriter, file); err != nil {
				return fmt.Errorf("failed to write file content for %s: %w", file, err)
			}
		}

		return nil
	})
}

func createTarHeader(info os.FileInfo, filePath string) (*tar.Header, error) {
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return nil, err
	}

	baseName := filepath.Base(filePath)
	if isBuildlibBinary(baseName) {
		header.Name = normalizeBinaryName(baseName)
	} else {
		header.Name = filePath
	}

	return header, nil
}

func isBuildlibBinary(fileName string) bool {
	return len(fileName) >= len(binaryBaseName)+1 &&
		fileName[:len(binaryBaseName)+1] == binaryBaseName+"_"
}

func normalizeBinaryName(fileName string) string {
	if filepath.Ext(fileName) == ".exe" {
		return binaryBaseName + ".exe"
	}
	return binaryBaseName
}

func writeFileToArchive(tarWriter *tar.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(tarWriter, file)
	return err
}

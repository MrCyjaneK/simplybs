package buildweb

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mrcyjanek/simplybs/host"
	"github.com/mrcyjanek/simplybs/pack"
)

func ListArchiveContents(pkg *pack.Package, download *pack.Download) ([]ArchiveFileInfo, error) {
	sourcePath := pkg.GenerateSourceBuildPath(download)

	info, err := os.Stat(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat archive: %w", err)
	}
	if info.Size() == 0 {
		return nil, fmt.Errorf("archive is 0 bytes (incomplete download)")
	}

	if download.Kind == "blob" || download.Kind == "none" {
		return nil, fmt.Errorf("cannot list contents of %s type", download.Kind)
	}

	tempDir, err := os.MkdirTemp("", "buildweb-archive-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	tempPkg := &pack.Package{
		Package:  pkg.Package,
		Version:  pkg.Version,
		Download: []*pack.Download{download},
	}

	tempPkg.ExtractSource(&host.Host{}, tempDir)

	var files []ArchiveFileInfo
	err = filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, _ := filepath.Rel(tempDir, path)
			files = append(files, ArchiveFileInfo{
				Name: relPath,
				Size: info.Size(),
			})
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk extracted files: %w", err)
	}

	return files, nil
}

func GetArchiveInfo(pkg *pack.Package, download *pack.Download) ArchiveInfo {
	files, err := ListArchiveContents(pkg, download)
	if err != nil {
		return ArchiveInfo{Files: []ArchiveFileInfo{}, TotalSize: 0, FileCount: 0}
	}

	var totalSize int64
	for _, file := range files {
		totalSize += file.Size
	}

	return ArchiveInfo{
		Files:     files,
		TotalSize: totalSize,
		FileCount: len(files),
	}
}

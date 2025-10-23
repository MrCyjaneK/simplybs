package buildweb

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mrcyjanek/simplybs/crash"
	"github.com/mrcyjanek/simplybs/host"
	"github.com/mrcyjanek/simplybs/pack"
)

func GenerateIndexJSON(packagesWithBuilds []*pack.PackageWithBuilds, webDir string) {
	path := filepath.Join(webDir, "index.json")

	if _, err := os.Stat(path); err == nil {
		return
	}

	index := GetWebsiteIndex(packagesWithBuilds)

	data, err := json.MarshalIndent(index, "", "  ")
	crash.Handle(err)

	err = os.WriteFile(path, data, 0644)
	crash.Handle(err)
}

func GeneratePackageJSON(pkg *pack.PackageWithBuilds, webDir string) {
	path := filepath.Join(webDir, pkg.Package.Package+".json")

	if _, err := os.Stat(path); err == nil {
		return
	}

	metadata := ConvertToPackageMetadata(pkg)

	data, err := json.MarshalIndent(metadata, "", "  ")
	crash.Handle(err)

	err = os.MkdirAll(filepath.Dir(path), 0755)
	crash.Handle(err)

	err = os.WriteFile(path, data, 0644)
	crash.Handle(err)
}

func GenerateBuilderJSON(builderName string, metadata *BuildMetadata, webDir string) {
	path := filepath.Join(webDir, "builder_"+builderName+".json")

	if _, err := os.Stat(path); err == nil {
		return
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	crash.Handle(err)

	err = os.WriteFile(path, data, 0644)
	crash.Handle(err)
}

func GenerateBuilderMetadataFiles(metadata map[string]*BuildMetadata) {
	baseBuildDir := host.DataDirRoot()

	for builderName, meta := range metadata {
		builderDir := filepath.Join(baseBuildDir, builderName, "built")
		err := os.MkdirAll(builderDir, 0755)
		crash.Handle(err)

		metadataPath := filepath.Join(builderDir, "metadata.json")

		if _, err := os.Stat(metadataPath); err == nil {
			continue
		}

		data, err := json.MarshalIndent(meta, "", "  ")
		crash.Handle(err)

		err = os.WriteFile(metadataPath, data, 0644)
		crash.Handle(err)
	}
}

func GenerateFileDetailsJSON(pkg *pack.PackageWithBuilds, builtFile *pack.BuiltFile, webDir string) {
	fileName := pkg.Package.Package + "-" + pkg.Package.Version + "-" + builtFile.ID + ".json"
	path := filepath.Join(webDir, "files", builtFile.Builder, builtFile.Target, fileName)

	if _, err := os.Stat(path); err == nil {
		return
	}

	baseBuildDir := host.DataDirRoot()
	fullArchPath := filepath.Join(baseBuildDir, builtFile.ArchPath)

	info, err := os.Stat(fullArchPath)
	if os.IsNotExist(err) {
		return
	}

	if info.Size() == 0 {
		crash.Handle(fmt.Errorf("built archive is 0 bytes: %s", fullArchPath))
	}

	tempDownload := &pack.Download{Kind: "tar.gz"}
	tempPkg := &pack.Package{
		Package: pkg.Package.Package,
		Version: pkg.Package.Version,
	}
	archiveInfo := GetArchiveInfo(tempPkg, tempDownload)

	data := struct {
		Package     string      `json:"package"`
		Version     string      `json:"version"`
		Builder     string      `json:"builder"`
		Target      string      `json:"target"`
		BuildID     string      `json:"build_id"`
		FileSize    int64       `json:"file_size"`
		ArchivePath string      `json:"archive_path"`
		InfoPath    string      `json:"info_path"`
		Archive     ArchiveInfo `json:"archive"`
	}{
		Package:     pkg.Package.Package,
		Version:     pkg.Package.Version,
		Builder:     builtFile.Builder,
		Target:      builtFile.Target,
		BuildID:     builtFile.ID,
		FileSize:    builtFile.FileSize,
		ArchivePath: builtFile.ArchPath,
		InfoPath:    builtFile.InfoPath,
		Archive:     archiveInfo,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	crash.Handle(err)

	err = os.MkdirAll(filepath.Dir(path), 0755)
	crash.Handle(err)

	err = os.WriteFile(path, jsonData, 0644)
	crash.Handle(err)
}

func GenerateSourceDetailsJSON(pkg *pack.PackageWithBuilds, download *pack.Download, downloadIndex int, webDir string) {
	fileName := pkg.Package.Package + "-download-" + fmt.Sprintf("%d", downloadIndex) + ".json"
	path := filepath.Join(webDir, "source", fileName)

	if _, err := os.Stat(path); err == nil {
		return
	}

	sourcePath := pkg.Package.GenerateSourceBuildPath(download)
	relSourcePath, _ := filepath.Rel(host.DataDirRoot(), sourcePath)

	var fileSize int64
	var archiveInfo ArchiveInfo
	if info, err := os.Stat(sourcePath); err == nil {
		fileSize = info.Size()
		if download.Kind != "blob" && download.Kind != "none" {
			archiveInfo = GetArchiveInfo(pkg.Package, download)
		}
	}

	data := struct {
		Package    string      `json:"package"`
		Version    string      `json:"version"`
		Kind       string      `json:"kind"`
		URL        string      `json:"url"`
		SHA256     string      `json:"sha256"`
		Path       string      `json:"path,omitempty"`
		FileSize   int64       `json:"file_size"`
		SourcePath string      `json:"source_path"`
		Archive    ArchiveInfo `json:"archive,omitempty"`
	}{
		Package:    pkg.Package.Package,
		Version:    pkg.Package.Version,
		Kind:       download.Kind,
		URL:        download.URL,
		SHA256:     download.Sha256,
		Path:       download.Path,
		FileSize:   fileSize,
		SourcePath: relSourcePath,
		Archive:    archiveInfo,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	crash.Handle(err)

	err = os.MkdirAll(filepath.Dir(path), 0755)
	crash.Handle(err)

	err = os.WriteFile(path, jsonData, 0644)
	crash.Handle(err)
}

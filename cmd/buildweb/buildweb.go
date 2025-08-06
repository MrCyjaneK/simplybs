package buildweb

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"html/template"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	_ "embed"

	"github.com/mrcyjanek/simplybs/builder"
	"github.com/mrcyjanek/simplybs/crash"
	"github.com/mrcyjanek/simplybs/host"
	"github.com/mrcyjanek/simplybs/pack"
)

func getGlobPattern(entry string) string {
	if strings.Contains(entry, ":") {
		parts := strings.SplitN(entry, ":", 2)
		return parts[0]
	}
	return ""
}

func getGlobContent(entry string) string {
	if strings.Contains(entry, ":") {
		parts := strings.SplitN(entry, ":", 2)
		return parts[1]
	}
	return entry
}

func dependencyExists(dep string, packages []*pack.Package) bool {
	for _, pkg := range packages {
		if pkg.Package == dep {
			return true
		}
	}
	return false
}

func formatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

type ArchiveFileInfo struct {
	Name string
	Size int64
}

func listArchiveContents(archivePath string) ([]ArchiveFileInfo, error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer gzReader.Close()

	tarReader := tar.NewReader(gzReader)
	var files []ArchiveFileInfo

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		if header.Typeflag == tar.TypeReg {
			files = append(files, ArchiveFileInfo{
				Name: header.Name,
				Size: header.Size,
			})
		}
	}

	return files, nil
}

func scanBuiltFilesAllPlatforms(packageName string, packageVersion string) []pack.BuiltFile {
	var builtFiles []pack.BuiltFile
	baseBuildDir := host.DataDirRoot()

	targets := make([]string, 0, len(host.SupportedHosts))
	for k := range host.SupportedHosts {
		targets = append(targets, k)
	}

	for _, builderName := range builder.Builders {
		builderDir := filepath.Join(baseBuildDir, builderName)
		if _, err := os.Stat(builderDir); os.IsNotExist(err) {
			continue
		}

		for _, target := range targets {
			buildOutputDir := filepath.Join(builderDir, "built", target, packageName)
			buildOutputDir = filepath.Dir(buildOutputDir)

			if _, err := os.Stat(buildOutputDir); os.IsNotExist(err) {
				continue
			}

			files, err := os.ReadDir(buildOutputDir)
			if err != nil {
				continue
			}

			for _, file := range files {
				fileName := file.Name()
				fsFilepath := filepath.Join(buildOutputDir, fileName)
				if !strings.Contains(fsFilepath, packageName) {
					continue
				}
				if !strings.HasSuffix(fileName, ".tar.gz") {
					continue
				}

				nameWithoutExt := strings.TrimSuffix(fileName, ".tar.gz")
				parts := strings.Split(nameWithoutExt, "-")
				if len(parts) < 3 {
					continue
				}

				archPathRelative := filepath.Join(builderName, "built", target, filepath.Dir(packageName), fileName)
				infoPathRelative := filepath.Join(builderName, "built", target, filepath.Dir(packageName), strings.TrimSuffix(fileName, ".tar.gz")+".info.txt")

				fullArchPath := filepath.Join(baseBuildDir, archPathRelative)
				info, err := os.Stat(fullArchPath)
				var fileSize int64
				if err == nil {
					fileSize = info.Size()
				}
				id := strings.Split(parts[len(parts)-1], ".")[0]

				builtFiles = append(builtFiles, pack.BuiltFile{
					Builder:  builderName,
					Target:   target,
					ID:       id,
					InfoPath: infoPathRelative,
					ArchPath: archPathRelative,
					FileSize: fileSize,
				})
			}
		}
	}

	return builtFiles
}

func getAllPackagesWithBuildsAllPlatforms() []*pack.PackageWithBuilds {
	packages := pack.GetAllPackages()
	packagesWithBuilds := make([]*pack.PackageWithBuilds, len(packages))

	for i, pkg := range packages {
		builtFiles := scanBuiltFilesAllPlatforms(pkg.Package, pkg.Version)
		packagesWithBuilds[i] = &pack.PackageWithBuilds{
			Package:    pkg,
			BuiltFiles: builtFiles,
		}
	}

	return packagesWithBuilds
}

func BuildWeb() {

	webDir := filepath.Join(host.DataDirRoot(), "web")

	if _, err := os.Stat(webDir); !os.IsNotExist(err) {
		err := os.RemoveAll(webDir)
		crash.Handle(err)
	}
	err := os.MkdirAll(webDir, 0755)
	crash.Handle(err)

	packagesWithBuilds := getAllPackagesWithBuildsAllPlatforms()

	for i, pkg := range packagesWithBuilds {
		builderTargetMap := make(map[string]pack.BuiltFile)
		for _, builtFile := range pkg.BuiltFiles {
			key := builtFile.Builder + "/" + builtFile.Target
			if _, exists := builderTargetMap[key]; !exists {
				builderTargetMap[key] = builtFile
			}
		}

		var filteredBuilds []pack.BuiltFile
		for _, builtFile := range builderTargetMap {
			filteredBuilds = append(filteredBuilds, builtFile)
		}
		packagesWithBuilds[i].BuiltFiles = filteredBuilds
	}

	packages := make([]*pack.Package, len(packagesWithBuilds))
	for i, pwb := range packagesWithBuilds {
		packages[i] = pwb.Package
	}

	funcMap := template.FuncMap{
		"getGlobPattern": getGlobPattern,
		"getGlobContent": getGlobContent,
		"depExists": func(dep string) bool {
			return dependencyExists(dep, packages)
		},
		"getRelativePath": func(fromPackage, toPackage string) string {
			fromDepth := strings.Count(fromPackage, "/")
			upPath := strings.Repeat("../", fromDepth)
			if toPackage == "index" {
				return upPath + "index.html"
			}
			return upPath + toPackage + ".html"
		},
		"formatFileSize": formatFileSize,
		"getMirrorPath": func(pkg *pack.PackageWithBuilds) string {
			packageDepth := strings.Count(pkg.Package.Package, "/")
			upPath := strings.Repeat("../", packageDepth+1) // +1 to get out of web directory

			return fmt.Sprintf("%ssource/%s-%s.%s", upPath, pkg.Package.Package, pkg.Package.Version, pkg.Package.Download.Kind)
		},
		"getBuiltFilePath": func(packageName, filePath string) string {
			packageDepth := strings.Count(packageName, "/")
			upPath := strings.Repeat("../", packageDepth+1) // +1 to get out of web directory
			return upPath + filePath
		},
		"getBuildMatrix": func(pkg *pack.PackageWithBuilds) map[string]map[string]*pack.BuiltFile {
			builders := make([]string, len(builder.Builders))
			copy(builders, builder.Builders)
			sort.Strings(builders)

			targets := make([]string, 0, len(host.SupportedHosts))
			for k := range host.SupportedHosts {
				targets = append(targets, k)
			}
			sort.Strings(targets)

			matrix := make(map[string]map[string]*pack.BuiltFile)
			for _, builder := range builders {
				matrix[builder] = make(map[string]*pack.BuiltFile)
				for _, target := range targets {
					matrix[builder][target] = nil
				}
			}

			for i := range pkg.BuiltFiles {
				bf := &pkg.BuiltFiles[i]
				if matrix[bf.Builder] != nil {
					matrix[bf.Builder][bf.Target] = bf
				}
			}

			return matrix
		},
		"getBuilders": func() []string {
			builders := make([]string, len(builder.Builders))
			copy(builders, builder.Builders)
			sort.Strings(builders)
			return builders
		},
		"getTargets": func() []string {
			targets := make([]string, 0, len(host.SupportedHosts))
			for k := range host.SupportedHosts {
				targets = append(targets, k)
			}
			sort.Strings(targets)
			return targets
		},
		"getBuildProgress": func(pkg *pack.PackageWithBuilds) int {
			totalCombinations := len(builder.Builders) * len(host.SupportedHosts)
			if totalCombinations == 0 {
				return 0
			}

			actualBuilds := len(pkg.BuiltFiles)
			return (actualBuilds * 100) / totalCombinations
		},
	}

	log.Printf("Generating index page...")
	generateIndexPage(packagesWithBuilds, webDir, funcMap)

	log.Printf("Generating %d package pages...", len(packagesWithBuilds))
	for i, pkgWithBuilds := range packagesWithBuilds {
		log.Printf("\tProgress: %d/%d packages\n", i+1, len(packagesWithBuilds))
		generatePackagePage(pkgWithBuilds, webDir, funcMap)
	}

	log.Printf("Generating %d builder matrix pages...\n", len(builder.Builders))
	for i, builderName := range builder.Builders {
		log.Printf("\tGenerating matrix for %s (%d/%d)\n", builderName, i+1, len(builder.Builders))
		generateBuilderMatrixPage(builderName, packagesWithBuilds, webDir, funcMap)
	}

	totalFiles := 0
	for _, pkgWithBuilds := range packagesWithBuilds {
		totalFiles += len(pkgWithBuilds.BuiltFiles)
	}

	log.Printf("Generating %d file detail pages...\n", totalFiles)
	fileCount := 0
	for _, pkgWithBuilds := range packagesWithBuilds {
		for _, builtFile := range pkgWithBuilds.BuiltFiles {
			fileCount++
			log.Printf("\tProgress: %d/%d file pages\n", fileCount, totalFiles)
			generateFilePage(pkgWithBuilds, &builtFile, webDir, funcMap)
		}
	}

	log.Printf("Generated static website with %d packages, %d builder matrices, and %d file details in %s\n", len(packages), len(builder.Builders), totalFiles, webDir)
}

//go:embed index.tpl
var indexTemplate string

func generateIndexPage(packagesWithBuilds []*pack.PackageWithBuilds, webDir string, funcMap template.FuncMap) {

	tmpl, err := template.New("index").Funcs(funcMap).Parse(indexTemplate)
	crash.Handle(err)

	file, err := os.Create(filepath.Join(webDir, "index.html"))
	crash.Handle(err)
	defer file.Close()

	err = tmpl.Execute(file, packagesWithBuilds)
	crash.Handle(err)
}

//go:embed package.tpl
var packageTemplate string

func generatePackagePage(pkgWithBuilds *pack.PackageWithBuilds, webDir string, funcMap template.FuncMap) {
	funcMap["add"] = func(a, b int) int {
		return a + b
	}

	tmpl, err := template.New("package").Funcs(funcMap).Parse(packageTemplate)
	crash.Handle(err)

	path := filepath.Join(webDir, pkgWithBuilds.Package.Package+".html")
	err = os.MkdirAll(filepath.Dir(path), 0755)
	crash.Handle(err)

	file, err := os.Create(path)
	crash.Handle(err)
	defer file.Close()

	err = tmpl.Execute(file, pkgWithBuilds)
	crash.Handle(err)
}

//go:embed builder_matrix.tpl
var builderMatrixTemplate string

func generateBuilderMatrixPage(builderName string, packagesWithBuilds []*pack.PackageWithBuilds, webDir string, funcMap template.FuncMap) {
	tmpl, err := template.New("builder_matrix").Funcs(funcMap).Parse(builderMatrixTemplate)
	crash.Handle(err)

	path := filepath.Join(webDir, "builder_"+builderName+".html")
	err = os.MkdirAll(filepath.Dir(path), 0755)
	crash.Handle(err)

	file, err := os.Create(path)
	crash.Handle(err)
	defer file.Close()

	data := struct {
		Builder  string
		Packages []*pack.PackageWithBuilds
	}{
		Builder:  builderName,
		Packages: packagesWithBuilds,
	}

	err = tmpl.Execute(file, data)
	crash.Handle(err)
}

//go:embed file_details.tpl
var fileDetailsTemplate string

type ArchiveInfo struct {
	Files     []ArchiveFileInfo
	TotalSize int64
}

func generateFilePage(pkg *pack.PackageWithBuilds, builtFile *pack.BuiltFile, webDir string, funcMap template.FuncMap) {
	funcMap["getArchiveInfo"] = func(archPath string) ArchiveInfo {
		baseBuildDir := host.DataDirRoot()
		fullPath := filepath.Join(baseBuildDir, archPath)
		files, err := listArchiveContents(fullPath)
		if err != nil {
			return ArchiveInfo{Files: []ArchiveFileInfo{}, TotalSize: 0}
		}

		var totalSize int64
		for _, file := range files {
			totalSize += file.Size
		}

		return ArchiveInfo{Files: files, TotalSize: totalSize}
	}

	tmpl, err := template.New("file_details").Funcs(funcMap).Parse(fileDetailsTemplate)
	crash.Handle(err)

	fileName := fmt.Sprintf("%s-%s-%s.html", pkg.Package.Package, pkg.Package.Version, builtFile.ID)
	path := filepath.Join(webDir, "files", builtFile.Builder, builtFile.Target, fileName)
	err = os.MkdirAll(filepath.Dir(path), 0755)
	crash.Handle(err)

	file, err := os.Create(path)
	crash.Handle(err)
	defer file.Close()

	data := struct {
		Package   *pack.PackageWithBuilds
		BuiltFile *pack.BuiltFile
	}{
		Package:   pkg,
		BuiltFile: builtFile,
	}

	err = tmpl.Execute(file, data)
	crash.Handle(err)
}

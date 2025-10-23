package buildweb

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sync"

	_ "embed"

	"github.com/mrcyjanek/simplybs/builder"
	"github.com/mrcyjanek/simplybs/crash"
	"github.com/mrcyjanek/simplybs/host"
	"github.com/mrcyjanek/simplybs/pack"
)

//go:embed index.tpl
var indexTemplate string

//go:embed package.tpl
var packageTemplate string

//go:embed builder_matrix.tpl
var builderMatrixTemplate string

//go:embed file_details.tpl
var fileDetailsTemplate string

//go:embed source_details.tpl
var sourceDetailsTemplate string

func GenerateIndexHTML(packagesWithBuilds []*pack.PackageWithBuilds, webDir string, funcMap template.FuncMap) {
	path := filepath.Join(webDir, "index.html")

	tmpl, err := template.New("index").Funcs(funcMap).Parse(indexTemplate)
	crash.Handle(err)

	file, err := os.Create(path)
	crash.Handle(err)
	defer file.Close()

	err = tmpl.Execute(file, packagesWithBuilds)
	crash.Handle(err)
}

func GeneratePackageHTML(pkgWithBuilds *pack.PackageWithBuilds, webDir string, funcMap template.FuncMap) {
	path := filepath.Join(webDir, pkgWithBuilds.Package.Package+".html")

	funcMap["add"] = func(a, b int) int {
		return a + b
	}

	tmpl, err := template.New("package").Funcs(funcMap).Parse(packageTemplate)
	crash.Handle(err)

	err = os.MkdirAll(filepath.Dir(path), 0755)
	crash.Handle(err)

	file, err := os.Create(path)
	crash.Handle(err)
	defer file.Close()

	err = tmpl.Execute(file, pkgWithBuilds)
	crash.Handle(err)
}

func GenerateBuilderMatrixHTML(builderName string, packagesWithBuilds []*pack.PackageWithBuilds, webDir string, funcMap template.FuncMap) {
	path := filepath.Join(webDir, "builder_"+builderName+".html")

	tmpl, err := template.New("builder_matrix").Funcs(funcMap).Parse(builderMatrixTemplate)
	crash.Handle(err)

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

func GenerateFileDetailsHTML(pkg *pack.PackageWithBuilds, builtFile *pack.BuiltFile, webDir string, funcMap template.FuncMap) {
	fileName := fmt.Sprintf("%s-%s-%s.html", pkg.Package.Package, pkg.Package.Version, builtFile.ID)
	path := filepath.Join(webDir, "files", builtFile.Builder, builtFile.Target, fileName)

	tmpl, err := template.New("file_details").Funcs(funcMap).Parse(fileDetailsTemplate)
	crash.Handle(err)

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

func CreateTemplateFuncMap(packages []*pack.Package) template.FuncMap {
	return template.FuncMap{
		"getGlobPattern": GetGlobPattern,
		"getGlobContent": GetGlobContent,
		"depExists": func(dep string) bool {
			return DependencyExists(dep, packages)
		},
		"getRelativePath":      GetRelativePath,
		"formatFileSize":       FormatFileSize,
		"getMirrorPath":        GetMirrorPath,
		"getBuiltFilePath":     GetBuiltFilePath,
		"getFileDetailsPath":   GetFileDetailsPath,
		"getSourceFilePath":    GetSourceFilePath,
		"getSourceDetailsPath": GetSourceDetailsPath,
		"getBuildMatrix":       GetBuildMatrix,
		"getBuilders":          getSortedBuilders,
		"getTargets":           getSortedTargets,
		"getBuildProgress": func(pkg *pack.PackageWithBuilds) int {
			return GetBuildProgress(pkg, len(builder.Builders), len(host.SupportedHosts))
		},
		"getArchiveInfo": func(pkg *pack.Package, download *pack.Download) ArchiveInfo {
			return GetArchiveInfo(pkg, download)
		},
	}
}

func GenerateSourceDetailsHTML(pkg *pack.PackageWithBuilds, download *pack.Download, downloadIndex int, webDir string, funcMap template.FuncMap) {
	fileName := fmt.Sprintf("%s-download-%d.html", pkg.Package.Package, downloadIndex)
	path := filepath.Join(webDir, "source", fileName)

	tmpl, err := template.New("source_details").Funcs(funcMap).Parse(sourceDetailsTemplate)
	crash.Handle(err)

	err = os.MkdirAll(filepath.Dir(path), 0755)
	crash.Handle(err)

	file, err := os.Create(path)
	crash.Handle(err)
	defer file.Close()

	sourcePath := pkg.Package.GenerateSourceBuildPath(download)
	relSourcePath, _ := filepath.Rel(host.DataDirRoot(), sourcePath)

	var fileSize int64
	if info, err := os.Stat(sourcePath); err == nil {
		fileSize = info.Size()
	}

	data := struct {
		Package    *pack.Package
		Download   *pack.Download
		SourcePath string
		FileSize   int64
	}{
		Package:    pkg.Package,
		Download:   download,
		SourcePath: relSourcePath,
		FileSize:   fileSize,
	}

	err = tmpl.Execute(file, data)
	crash.Handle(err)
}

func GenerateAllHTML(packagesWithBuilds []*pack.PackageWithBuilds, webDir string) {
	packages := make([]*pack.Package, len(packagesWithBuilds))
	for i, pwb := range packagesWithBuilds {
		packages[i] = pwb.Package
	}

	funcMap := CreateTemplateFuncMap(packages)

	GenerateIndexHTML(packagesWithBuilds, webDir, funcMap)

	for _, pkgWithBuilds := range packagesWithBuilds {
		GeneratePackageHTML(pkgWithBuilds, webDir, funcMap)
	}

	for _, builderName := range builder.Builders {
		GenerateBuilderMatrixHTML(builderName, packagesWithBuilds, webDir, funcMap)
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, 10)

	for _, pkgWithBuilds := range packagesWithBuilds {
		for _, builtFile := range pkgWithBuilds.BuiltFiles {
			wg.Add(1)
			semaphore <- struct{}{}

			go func(pkg *pack.PackageWithBuilds, bf pack.BuiltFile) {
				defer wg.Done()
				defer func() { <-semaphore }()

				GenerateFileDetailsHTML(pkg, &bf, webDir, funcMap)
			}(pkgWithBuilds, builtFile)
		}

		for i, download := range pkgWithBuilds.Package.Download {
			wg.Add(1)
			semaphore <- struct{}{}

			go func(pkg *pack.PackageWithBuilds, dl *pack.Download, idx int) {
				defer wg.Done()
				defer func() { <-semaphore }()

				GenerateSourceDetailsHTML(pkg, dl, idx, webDir, funcMap)
			}(pkgWithBuilds, download, i)
		}
	}

	wg.Wait()
}

package buildweb

import (
	"fmt"
	"html/template"
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

func BuildWeb() {
	webDir := filepath.Join(host.DataDirRoot(), "web")

	if _, err := os.Stat(webDir); !os.IsNotExist(err) {
		err := os.RemoveAll(webDir)
		crash.Handle(err)
	}
	err := os.MkdirAll(webDir, 0755)
	crash.Handle(err)

	packagesWithBuilds := pack.GetAllPackagesWithBuilds()

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
			actualBuilds := len(pkg.BuiltFiles)
			if totalCombinations == 0 {
				return 0
			}
			return (actualBuilds * 100) / totalCombinations
		},
	}

	generateIndexPage(packagesWithBuilds, webDir, funcMap)

	for _, pkgWithBuilds := range packagesWithBuilds {
		generatePackagePage(pkgWithBuilds, webDir, funcMap)
	}

	fmt.Printf("Generated static website with %d packages in %s\n", len(packages), webDir)
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

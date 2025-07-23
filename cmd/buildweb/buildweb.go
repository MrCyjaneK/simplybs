package buildweb

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	_ "embed"

	"github.com/mrcyjanek/simplybs/crash"
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

func BuildWeb() {
	webDir := "./.buildlib/web"

	if _, err := os.Stat(webDir); !os.IsNotExist(err) {
		err := os.RemoveAll(webDir)
		crash.Handle(err)
	}
	err := os.MkdirAll(webDir, 0755)
	crash.Handle(err)

	packages := pack.GetAllPackages()

	funcMap := template.FuncMap{
		"getGlobPattern": getGlobPattern,
		"getGlobContent": getGlobContent,
		"depExists": func(dep string) bool {
			return dependencyExists(dep, packages)
		},
	}

	generateIndexPage(packages, webDir)

	for _, pkg := range packages {
		generatePackagePage(pkg, webDir, funcMap)
	}

	fmt.Printf("Generated static website with %d packages in %s\n", len(packages), webDir)
}

//go:embed index.tpl
var indexTemplate string

func generateIndexPage(packages []*pack.Package, webDir string) {

	tmpl, err := template.New("index").Parse(indexTemplate)
	crash.Handle(err)

	file, err := os.Create(filepath.Join(webDir, "index.html"))
	crash.Handle(err)
	defer file.Close()

	err = tmpl.Execute(file, packages)
	crash.Handle(err)
}

//go:embed package.tpl
var packageTemplate string

func generatePackagePage(pkg *pack.Package, webDir string, funcMap template.FuncMap) {
	funcMap["add"] = func(a, b int) int {
		return a + b
	}

	tmpl, err := template.New("package").Funcs(funcMap).Parse(packageTemplate)
	crash.Handle(err)

	file, err := os.Create(filepath.Join(webDir, pkg.Package+".html"))
	crash.Handle(err)
	defer file.Close()

	err = tmpl.Execute(file, pkg)
	crash.Handle(err)
}

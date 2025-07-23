package pack

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/mrcyjanek/simplybs/crash"
	"github.com/mrcyjanek/simplybs/host"
)

func GetPackagesByList(list string) []*Package {
	packageNames := strings.Split(list, ",")
	for i, name := range packageNames {
		packageNames[i] = strings.TrimSpace(name)
	}
	packages := []*Package{}
	for _, name := range packageNames {
		pkg, err := FindPackage(name)
		if err != nil {
			log.Printf("Package %s not found in %s", name, host.GetPackagesDir())
			continue
		}
		packages = append(packages, pkg)
	}
	return packages
}
func GetAllPackages() []*Package {
	packages := []*Package{}
	files, err := filepath.Glob("packages/*.json")
	crash.Handle(err)
	for _, file := range files {
		file = filepath.Base(file)
		pkgName := file[:len(file)-len(filepath.Ext(file))]
		pkg, err := FindPackage(pkgName)
		if err != nil {
			log.Printf("Package %s not found", pkgName)
			continue
		}
		packages = append(packages, pkg)
	}
	return packages
}

package pack

import (
	"log"
	"os"
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
	packagesDir := host.GetPackagesDir()

	err := filepath.WalkDir(packagesDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || filepath.Ext(path) != ".json" {
			return nil
		}

		relPath, err := filepath.Rel(packagesDir, path)
		if err != nil {
			log.Printf("Failed to get relative path for %s: %v", path, err)
			return nil
		}

		pkgName := relPath[:len(relPath)-len(filepath.Ext(relPath))]

		pkg, err := FindPackage(pkgName)
		if err != nil {
			log.Printf("Package %s not found", pkgName)
			return nil
		}
		packages = append(packages, pkg)
		return nil
	})
	crash.Handle(err)

	return packages
}

func GetAllPackagesWithBuilds() []*PackageWithBuilds {
	packages := GetAllPackages()
	packagesWithBuilds := make([]*PackageWithBuilds, len(packages))

	for i, pkg := range packages {
		builtFiles := ScanBuiltFiles(pkg.Package, pkg.Version)
		packagesWithBuilds[i] = &PackageWithBuilds{
			Package:    pkg,
			BuiltFiles: builtFiles,
		}
	}

	return packagesWithBuilds
}

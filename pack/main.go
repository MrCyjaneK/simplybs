package pack

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mrcyjanek/simplybs/builder"
	"github.com/mrcyjanek/simplybs/crash"
	"github.com/mrcyjanek/simplybs/host"
	"github.com/ryanuber/go-glob"
)

type Package struct {
	Package  string `json:"package"`
	Version  string `json:"version"`
	Type     string `json:"type"`
	Download struct {
		Kind   string `json:"kind"`
		URL    string `json:"url"`
		Sha256 string `json:"sha256"`
	} `json:"download"`
	Build struct {
		Env   []string `json:"env"`
		Steps []string `json:"steps"`
	} `json:"build"`
	Dependencies []string `json:"dependencies"`
}

type BuiltFile struct {
	Builder  string `json:"builder"`   // e.g. "darwin_arm64"
	Target   string `json:"target"`    // e.g. "aarch64-apple-ios"
	ID       string `json:"id"`        // short hash
	InfoPath string `json:"info_path"` // relative path to .info.txt
	ArchPath string `json:"arch_path"` // relative path to .tar.gz
	FileSize int64  `json:"file_size"` // size in bytes
}

type PackageWithBuilds struct {
	*Package
	BuiltFiles []BuiltFile `json:"built_files"`
}

func FindPackage(name string) (*Package, error) {
	pkgPath := filepath.Join(host.GetPackagesDir(), name+".json")
	info, err := os.ReadFile(pkgPath)
	if err != nil {
		return nil, err
	}
	var pkg Package
	err = json.Unmarshal(info, &pkg)
	if err != nil {
		return nil, err
	}
	return &pkg, nil
}

func PrintPackage(pkgName string, host string) {
	depsByLevel := collectDependenciesByLevel(pkgName, host)

	userPkg, err := FindPackage(pkgName)
	crash.Handle(err)
	fmt.Printf("0: %s (version: %s)\n", pkgName, userPkg.Version)

	for level := 1; level < len(depsByLevel); level++ {
		deps := depsByLevel[level]
		if len(deps) == 0 {
			continue
		}

		for i := 0; i < len(deps); i++ {
			for j := i + 1; j < len(deps); j++ {
				if deps[i] > deps[j] {
					deps[i], deps[j] = deps[j], deps[i]
				}
			}
		}

		for _, dep := range deps {
			pkg, err := FindPackage(dep)
			if err != nil {
				fmt.Printf("%d: %s (ERROR: %v)\n", level, dep, err)
			} else {
				fmt.Printf("%d: %s (version: %s)\n", level, dep, pkg.Version)
			}
		}
	}
}

func ScanBuiltFiles(packageName string, packageVersion string) []BuiltFile {
	var builtFiles []BuiltFile
	buildlibDir := host.DataDirRoot()

	targets := make([]string, 0, len(host.SupportedHosts))
	for k := range host.SupportedHosts {
		targets = append(targets, k)
	}

	for _, builder := range builder.Builders {
		for _, target := range targets {
			// For packages like "native/make", files are stored in the "native" directory
			// but the filename is just "make-version-id.tar.gz"
			packageParts := strings.Split(packageName, "/")
			var builtDir string
			var expectedPackageName string

			if len(packageParts) > 1 {
				// Package like "native/make" - files are in "native/" dir, filename starts with "make"
				builtDir = filepath.Join(buildlibDir, builder, "built", target, packageParts[0])
				expectedPackageName = packageParts[1]
			} else {
				// Package like "boost" - files are in root, filename starts with "boost"
				builtDir = filepath.Join(buildlibDir, builder, "built", target)
				expectedPackageName = packageName
			}

			if _, err := os.Stat(builtDir); os.IsNotExist(err) {
				continue
			}

			files, err := os.ReadDir(builtDir)
			if err != nil {
				continue
			}

			for _, file := range files {
				fileName := file.Name()
				if !strings.HasSuffix(fileName, ".tar.gz") {
					continue
				}

				// Parse filename: ${package}-${version}-${id}.tar.gz
				nameWithoutExt := strings.TrimSuffix(fileName, ".tar.gz")
				parts := strings.Split(nameWithoutExt, "-")
				if len(parts) < 3 {
					continue
				}

				// Reconstruct package name and check if this file belongs to our package
				reconstructedPkg := strings.Join(parts[:len(parts)-2], "-")
				reconstructedVersion := parts[len(parts)-2]
				id := parts[len(parts)-1]

				if reconstructedPkg == expectedPackageName && reconstructedVersion == packageVersion {
					var archPathRelative, infoPathRelative string
					if len(packageParts) > 1 {
						archPathRelative = filepath.Join(builder, "built", target, packageParts[0], fileName)
						infoPathRelative = filepath.Join(builder, "built", target, packageParts[0], strings.TrimSuffix(fileName, ".tar.gz")+".info.txt")
					} else {
						archPathRelative = filepath.Join(builder, "built", target, fileName)
						infoPathRelative = filepath.Join(builder, "built", target, strings.TrimSuffix(fileName, ".tar.gz")+".info.txt")
					}
					archPath := archPathRelative
					infoPath := infoPathRelative

					fullArchPath := filepath.Join(buildlibDir, archPath)
					info, err := os.Stat(fullArchPath)
					var fileSize int64
					if err == nil {
						fileSize = info.Size()
					}

					builtFiles = append(builtFiles, BuiltFile{
						Builder:  builder,
						Target:   target,
						ID:       id,
						InfoPath: infoPath,
						ArchPath: archPath,
						FileSize: fileSize,
					})
				}
			}
		}
	}

	return builtFiles
}

func Cleanup() {
	buildlibDir := host.DataDirRoot()

	packages := GetAllPackages()

	keepFiles := make(map[string]bool)

	builders := builder.Builders
	targets := make([]string, 0, len(host.SupportedHosts))
	for k := range host.SupportedHosts {
		targets = append(targets, k)
	}

	for _, pkg := range packages {
		for _, builder := range builders {
			for _, target := range targets {
				h := &host.Host{Triplet: target}

				currentBuildID := pkg.GeneratePackageInfoShortHash(h)

				packageParts := strings.Split(pkg.Package, "/")
				var expectedPackageName string
				var baseDir string

				if len(packageParts) > 1 {
					expectedPackageName = packageParts[1]
					baseDir = filepath.Join(builder, "built", target, packageParts[0])
				} else {
					expectedPackageName = pkg.Package
					baseDir = filepath.Join(builder, "built", target)
				}

				currentFileName := fmt.Sprintf("%s-%s-%s", expectedPackageName, pkg.Version, currentBuildID)
				archPath := filepath.Join(baseDir, currentFileName+".tar.gz")
				infoPath := filepath.Join(baseDir, currentFileName+".info.txt")

				keepFiles[archPath] = true
				keepFiles[infoPath] = true
			}
		}
	}

	fmt.Printf("Cleanup: Will keep %d current build files\n", len(keepFiles))

	for _, builder := range builders {
		builderDir := filepath.Join(buildlibDir, builder)
		if _, err := os.Stat(builderDir); os.IsNotExist(err) {
			continue
		}

		builtDir := filepath.Join(builderDir, "built")
		if _, err := os.Stat(builtDir); !os.IsNotExist(err) {
			err := filepath.WalkDir(builtDir, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}

				if d.IsDir() {
					return nil
				}

				relPath, err := filepath.Rel(buildlibDir, path)
				if err != nil {
					return nil
				}

				relPath = filepath.ToSlash(relPath)

				if !keepFiles[relPath] {
					fmt.Printf("Removing old build file: %s\n", relPath)
					os.Remove(path)
				}

				return nil
			})
			if err != nil {
				fmt.Printf("Error walking built directory %s: %v\n", builtDir, err)
			}
		}

		workDir := filepath.Join(builderDir, "work")
		stagingDir := filepath.Join(builderDir, "staging")
		envDir := filepath.Join(builderDir, "env")

		for _, dir := range []string{workDir, stagingDir, envDir} {
			if _, err := os.Stat(dir); !os.IsNotExist(err) {
				fmt.Printf("Removing directory: %s\n", filepath.Base(dir))
				os.RemoveAll(dir)
			}
		}
	}

	webDir := filepath.Join(buildlibDir, "web")
	if _, err := os.Stat(webDir); !os.IsNotExist(err) {
		fmt.Printf("Removing web directory\n")
		os.RemoveAll(webDir)
	}

	fmt.Println("Cleanup completed!")
}

func collectDependenciesByLevel(pkgName string, host string) [][]string {
	levels := [][]string{}
	visited := make(map[string]bool)

	currentLevel := []string{pkgName}
	visited[pkgName] = true
	levels = append(levels, []string{})

	for len(currentLevel) > 0 {
		nextLevel := []string{}

		for _, currentPkg := range currentLevel {
			pkg, err := FindPackage(currentPkg)
			if err != nil {
				continue
			}

			for _, dep := range pkg.Dependencies {
				var actualDep string
				if strings.Contains(dep, ":") {
					prefix := strings.Split(dep, ":")[0]
					if !glob.Glob(prefix, host) && prefix != "all" {
						continue
					}
					actualDep = dep[strings.Index(dep, ":")+1:]
				} else {
					actualDep = dep
				}

				if !visited[actualDep] {
					visited[actualDep] = true
					nextLevel = append(nextLevel, actualDep)
				}
			}
		}

		if len(nextLevel) > 0 {
			levels = append(levels, nextLevel)
			currentLevel = nextLevel
		} else {
			break
		}
	}

	return levels
}

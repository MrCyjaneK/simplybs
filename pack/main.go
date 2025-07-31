package pack

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

func collectDependenciesByLevel(pkgName string, host string) [][]string {
	levels := [][]string{}
	visited := make(map[string]bool)

	// Start with the root package
	currentLevel := []string{pkgName}
	visited[pkgName] = true
	levels = append(levels, []string{}) // Level 0 is handled separately in PrintPackage

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

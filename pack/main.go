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

func PrintPackage(pkgName string, host string, depth int) {
	pkg, err := FindPackage(pkgName)
	crash.Handle(err)
	fmt.Printf("%s%s: %s\n", strings.Repeat("  ", depth+1), pkg.Package, pkg.Version)
	for _, dep := range pkg.Dependencies {
		if strings.Contains(dep, ":") {
			prefix := strings.Split(dep, ":")[0]
			if !glob.Glob(prefix, host) && prefix != "all" {
				continue
			}
			dep = dep[strings.Index(dep, ":")+1:]
		}
		PrintPackage(dep, host, depth+1)
	}
}

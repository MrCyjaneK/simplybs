package lint

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mrcyjanek/simplybs/crash"
	"github.com/mrcyjanek/simplybs/host"
	"github.com/mrcyjanek/simplybs/pack"
	"github.com/ryanuber/go-glob"
)

type OrderedPackage struct {
	Package      string                 `json:"package"`
	Version      string                 `json:"version"`
	Type         string                 `json:"type"`
	Download     map[string]interface{} `json:"download,omitempty"`
	Dependencies []string               `json:"dependencies,omitempty"`
	Patches      []string               `json:"patches,omitempty"`
	Build        map[string]interface{} `json:"build,omitempty"`
}

func Lint() {
	fixFormatting()
	ensureSaneDependencies()

}

func fixFormatting() {
	files, err := filepath.Glob("packages/*.json")
	crash.Handle(err)
	for _, file := range files {
		contentInitial, err := os.ReadFile(file)
		crash.Handle(err)

		var data map[string]interface{}
		json.Unmarshal(contentInitial, &data)

		ordered := OrderedPackage{}

		if v, ok := data["package"].(string); ok {
			ordered.Package = v
		}
		if v, ok := data["version"].(string); ok {
			ordered.Version = v
		}
		if v, ok := data["type"].(string); ok {
			ordered.Type = v
		}
		if v, ok := data["download"].(map[string]interface{}); ok {
			ordered.Download = v
		}
		if v, ok := data["dependencies"].([]interface{}); ok {
			deps := make([]string, len(v))
			for i, dep := range v {
				if s, ok := dep.(string); ok {
					deps[i] = s
				}
			}
			ordered.Dependencies = deps
		}
		if v, ok := data["patches"].([]interface{}); ok {
			patches := make([]string, len(v))
			for i, patch := range v {
				if s, ok := patch.(string); ok {
					patches[i] = s
				}
			}
			ordered.Patches = patches
		}
		if v, ok := data["build"].(map[string]interface{}); ok {
			ordered.Build = v
		}

		var buf bytes.Buffer
		encoder := json.NewEncoder(&buf)
		encoder.SetEscapeHTML(false)
		encoder.SetIndent("", "    ")
		err = encoder.Encode(ordered)
		crash.Handle(err)

		contentNew := bytes.TrimSuffix(buf.Bytes(), []byte("\n"))

		if !bytes.Equal(contentNew, contentInitial) {
			log.Printf("Formatting %s", file)
			os.WriteFile(file, contentNew, 0644)
		}
	}
}

func ensureSaneDependencies() {
	pkgs := pack.GetAllPackages()
	for _, pkg := range pkgs {
		ensureValidName(pkg)
		ensureValidDependencies(pkg)
	}
}

func ensureValidName(pkg *pack.Package) {
	content, err := os.ReadFile(filepath.Join("packages", pkg.Package+".json"))
	if err != nil {
		log.Println(pkg.Package, "not found")
		return
	}
	var foundPackage pack.Package
	json.Unmarshal(content, &foundPackage)
	if foundPackage.Package != pkg.Package {
		log.Fatalf("Package %s has invalid name", pkg.Package)
	}
}

func ensureValidDependencies(pkg *pack.Package) {
	for _, dep := range pkg.Dependencies {
		split := strings.Split(dep, ":")
		if len(split) <= 1 {
			log.Printf("Package %s has invalid dependency %s", pkg.Package, dep)
			continue
		}
		prefix := split[0]

		usedIn := 0
		for _, host := range host.SupportedHosts {
			if glob.Glob(prefix, host.Triplet) {
				usedIn++
			}
		}
		if usedIn == 0 && prefix != "all" && prefix != "none" {
			log.Printf("Package %s is not used in any of host.SupportedHosts %s", pkg.Package, prefix)
		}

		_, err := pack.FindPackage(split[1])
		if err != nil {
			log.Printf("Package %s has invalid dependency %s: %v", pkg.Package, dep, err)
		}
	}
}

package lint

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gobwas/glob"
	"github.com/mrcyjanek/simplybs/crash"
	"github.com/mrcyjanek/simplybs/host"
	"github.com/mrcyjanek/simplybs/pack"
)

type RepositoryInfo struct {
	Refs []string `json:"refs"`
}

type SourcesFile struct {
	Repositories map[string]RepositoryInfo `json:"repositories"`
}

type OrderedPackage struct {
	Package      string                   `json:"package"`
	Version      string                   `json:"version"`
	Type         string                   `json:"type"`
	Download     []map[string]interface{} `json:"download,omitempty"`
	Dependencies []string                 `json:"dependencies,omitempty"`
	Patches      []string                 `json:"patches,omitempty"`
	Build        map[string]interface{}   `json:"build,omitempty"`
}

func Lint() {
	fixFormatting()
	ensureSaneDependencies()
	createSourcesFile()
}

func fixFormatting() {
	var files []string
	err := filepath.WalkDir("packages", func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".json") {
			files = append(files, path)
		}
		return nil
	})
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
		if v, ok := data["download"].([]interface{}); ok {
			ordered.Download = make([]map[string]interface{}, len(v))
			for i, download := range v {
				if m, ok := download.(map[string]interface{}); ok {
					ordered.Download[i] = m
				}
			}
		}
		if v, ok := data["dependencies"].([]interface{}); ok {
			nativeDeps := []string{}
			otherDeps := []string{}

			for _, dep := range v {
				if s, ok := dep.(string); ok {
					parts := strings.Split(s, ":")
					depName := parts[len(parts)-1]
					if strings.HasPrefix(depName, "native") {
						nativeDeps = append(nativeDeps, s)
					} else {
						otherDeps = append(otherDeps, s)
					}
				}
			}

			sort.Strings(nativeDeps)
			sort.Strings(otherDeps)
			ordered.Dependencies = append(nativeDeps, otherDeps...)
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
	ensureNoCyclicDependencies(pkgs)
}

func ensureValidName(pkg *pack.Package) {
	content, err := os.ReadFile(filepath.Join(host.GetPackagesDir(), pkg.Package+".json"))
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
			g := glob.MustCompile(prefix)
			if g.Match(host.Triplet) {
				usedIn++
			}
		}
		if usedIn == 0 && prefix != "none" {
			log.Printf("Package %s is not used in any of host.SupportedHosts %s", pkg.Package, prefix)
		}

		_, err := pack.FindPackage(split[1])
		if err != nil {
			log.Printf("Package %s has invalid dependency %s: %v", pkg.Package, dep, err)
		}
	}
}

func ensureNoCyclicDependencies(pkgs []*pack.Package) {
	for _, hostInfo := range host.SupportedHosts {
		checkCyclesForHost(pkgs, hostInfo.Triplet)
	}
	checkCyclesForHost(pkgs, "all")
}

func checkCyclesForHost(pkgs []*pack.Package, hostTriplet string) {
	graph := make(map[string][]string)
	allPackages := make(map[string]bool)

	for _, pkg := range pkgs {
		allPackages[pkg.Package] = true
		graph[pkg.Package] = []string{}

		for _, dep := range pkg.Dependencies {
			var actualDep string
			if strings.Contains(dep, ":") {
				prefix := strings.Split(dep, ":")[0]
				g := glob.MustCompile(prefix)
				if !g.Match(hostTriplet) {
					continue
				}
				actualDep = dep[strings.Index(dep, ":")+1:]
			} else {
				actualDep = dep
			}
			graph[pkg.Package] = append(graph[pkg.Package], actualDep)
		}
	}

	color := make(map[string]int)
	parent := make(map[string]string)

	for packageName := range allPackages {
		if color[packageName] == 0 {
			if dfsCycleDetection(packageName, graph, color, parent, hostTriplet) {
				return
			}
		}
	}
}

func dfsCycleDetection(packageName string, graph map[string][]string, color map[string]int, parent map[string]string, hostTriplet string) bool {
	color[packageName] = 1

	for _, neighbor := range graph[packageName] {
		if color[neighbor] == 1 {
			cycle := []string{neighbor}
			current := packageName
			for current != neighbor {
				cycle = append(cycle, current)
				current = parent[current]
			}
			cycle = append(cycle, neighbor)

			for i, j := 0, len(cycle)-1; i < j; i, j = i+1, j-1 {
				cycle[i], cycle[j] = cycle[j], cycle[i]
			}

			log.Fatalf("Cyclic dependency detected for host %s: %s", hostTriplet, strings.Join(cycle, " -> "))
			return true
		}

		if color[neighbor] == 0 {
			parent[neighbor] = packageName
			if dfsCycleDetection(neighbor, graph, color, parent, hostTriplet) {
				return true
			}
		}
	}

	color[packageName] = 2
	return false
}

func createSourcesFile() {
	log.Println("Creating sources.json file...")
	sources := SourcesFile{
		Repositories: make(map[string]RepositoryInfo),
	}

	pkgs := pack.GetAllPackages()
	for _, pkg := range pkgs {
		content, err := os.ReadFile(filepath.Join(host.GetPackagesDir(), pkg.Package+".json"))
		if err != nil {
			log.Printf("Failed to read package %s: %v", pkg.Package, err)
			continue
		}

		var packageData map[string]interface{}
		if err := json.Unmarshal(content, &packageData); err != nil {
			log.Printf("Failed to parse package %s: %v", pkg.Package, err)
			continue
		}

		downloads, ok := packageData["download"].([]interface{})
		if !ok {
			continue
		}

		for _, downloadInterface := range downloads {
			download, ok := downloadInterface.(map[string]interface{})
			if !ok {
				continue
			}

			kind, ok := download["kind"].(string)
			if !ok || kind != "git" {
				continue
			}

			url, ok := download["url"].(string)
			if !ok {
				continue
			}

			sha256, ok := download["sha256"].(string)
			if !ok || sha256 == "" {
				continue
			}

			if repoInfo, exists := sources.Repositories[url]; exists {
				refExists := false
				for _, ref := range repoInfo.Refs {
					if ref == sha256 {
						refExists = true
						break
					}
				}
				if !refExists {
					repoInfo.Refs = append(repoInfo.Refs, sha256)
				}

				sources.Repositories[url] = repoInfo
			} else {
				sources.Repositories[url] = RepositoryInfo{
					Refs: []string{sha256},
				}
			}
		}
	}

	sourcesJSON, err := json.MarshalIndent(sources, "", "    ")
	if err != nil {
		log.Fatalf("Failed to marshal sources.json: %v", err)
	}

	if err := os.WriteFile("sources.json", sourcesJSON, 0644); err != nil {
		log.Fatalf("Failed to write sources.json: %v", err)
	}

	log.Printf("Created sources.json with %d repositories", len(sources.Repositories))
}

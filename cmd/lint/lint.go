package lint

import (
	"bytes"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mrcyjanek/simplybs/builder"
	"github.com/mrcyjanek/simplybs/crash"
	"github.com/mrcyjanek/simplybs/host"
	"github.com/mrcyjanek/simplybs/pack"
	"github.com/mrcyjanek/simplybs/utils"
	"github.com/mrcyjanek/simplybs/utils/ifstring"
)

type OrderedPackage struct {
	Package      string                   `json:"package"`
	Version      string                   `json:"version"`
	Type         string                   `json:"type"`
	Download     []map[string]interface{} `json:"download,omitempty"`
	Dependencies []string                 `json:"dependencies,omitempty"`
	Patches      []string                 `json:"patches,omitempty"`
	ExportEnv    *[]string                `json:"export-env,omitempty"`
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
			ordered.Dependencies = make([]string, 0, len(v))
			for _, dep := range v {
				if s, ok := dep.(string); ok {
					ordered.Dependencies = append(ordered.Dependencies, s)
				}
			}
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
		if _, exists := data["export-env"]; exists {
			exportEnv := []string{}
			if v, ok := data["export-env"].([]interface{}); ok {
				for _, env := range v {
					if s, ok := env.(string); ok {
						exportEnv = append(exportEnv, s)
					}
				}
			}
			ordered.ExportEnv = &exportEnv
		}
		if v, ok := data["build"].(map[string]interface{}); ok {
			ordered.Build = v
		}

		var buf bytes.Buffer
		err = utils.NewIndentedEncoder(&buf).Encode(ordered)
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
		is := ifstring.ParseIfString(dep)

		usedIn := 0
		for _, host := range host.SupportedHosts {
			if is.HostGlob().Match(host.Triplet) {
				usedIn++
			}
		}
		usedInBuilder := 0
		for _, b := range builder.Builders {
			if is.BuilderGlob().Match(b) {
				usedInBuilder++
			}
		}
		if usedIn == 0 && is.Host != "none" {
			log.Printf("Package %s is not used in any of host.SupportedHosts %s", pkg.Package, is.Host)
		}
		if usedInBuilder == 0 && is.Builder != "none" {
			log.Printf("Package %s is not used in any of builder.Builders %s", pkg.Package, is.Builder)
		}

		_, err := pack.FindPackage(is.Content)
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
			is := ifstring.ParseIfString(dep)
			if is.HostGlob().Match(hostTriplet) {
				continue
			}
			actualDep = is.Content
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
	sources, err := utils.LoadSourcesFile()
	if err != nil {
		if os.IsNotExist(err) {
			sources = utils.NewSourcesFile()
		} else {
			log.Fatalf("Failed to load sources.json: %v", err)
		}
	} else if sources.Repositories == nil {
		sources.Repositories = make(map[string]utils.RepositoryInfo)
	}

	currentRefs := 0
	currentDownloads := 0

	pkgs := pack.GetAllPackages()
	for _, pkg := range pkgs {
		content, err := os.ReadFile(filepath.Join(host.GetPackagesDir(), pkg.Package+".json"))
		if err != nil {
			log.Printf("Failed to read package %s: %v", pkg.Package, err)
			continue
		}

		beforeRefs := utils.CountGitRefs(sources)
		beforeDownloads := len(sources.Downloads)
		utils.MergeDownloadsFromJSON(sources, content)
		currentRefs += utils.CountGitRefs(sources) - beforeRefs
		currentDownloads += len(sources.Downloads) - beforeDownloads
	}

	beforeHistoricRefs := utils.CountGitRefs(sources)
	beforeHistoricDownloads := len(sources.Downloads)

	if err := utils.CollectHistoricDownloads(sources); err != nil {
		log.Printf("Warning: historic source collection skipped: %v", err)
	}

	historicRefs := utils.CountGitRefs(sources) - beforeHistoricRefs
	historicDownloads := len(sources.Downloads) - beforeHistoricDownloads

	sourcesJSON, err := utils.MarshalSourcesFile(sources)
	if err != nil {
		log.Fatalf("Failed to marshal sources.json: %v", err)
	}

	if err := os.WriteFile("sources.json", sourcesJSON, 0644); err != nil {
		log.Fatalf("Failed to write sources.json: %v", err)
	}

	log.Printf(
		"Created sources.json with %d repositories (%d git refs, %d downloads); current packages added %d refs and %d downloads, history added %d refs and %d downloads",
		len(sources.Repositories),
		utils.CountGitRefs(sources),
		len(sources.Downloads),
		currentRefs,
		currentDownloads,
		historicRefs,
		historicDownloads,
	)
}

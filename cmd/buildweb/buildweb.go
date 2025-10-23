package buildweb

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"github.com/mrcyjanek/simplybs/builder"
	"github.com/mrcyjanek/simplybs/crash"
	"github.com/mrcyjanek/simplybs/host"
	"github.com/mrcyjanek/simplybs/pack"
)

func BuildWeb() {
	log.Println("Starting BuildWeb generation...")

	webDir := filepath.Join(host.DataDirRoot(), "web")
	if err := setupWebDirectory(webDir); err != nil {
		crash.Handle(err)
	}

	log.Println("Collecting package and build data...")
	packagesWithBuilds := GetAllPackagesWithBuilds()

	packagesWithBuilds = deduplicateBuilds(packagesWithBuilds)

	log.Println("Generating builder metadata...")
	buildersMetadata := GetBuildersMetadata(packagesWithBuilds)

	log.Println("Generating JSON files...")
	generateAllJSON(packagesWithBuilds, buildersMetadata, webDir)

	log.Println("Generating HTML pages...")
	GenerateAllHTML(packagesWithBuilds, webDir)

	printSummary(packagesWithBuilds)

	log.Printf("Website generated successfully in %s\n", webDir)
}

func setupWebDirectory(webDir string) error {
	// Don't clean the directory - just ensure it exists
	// This allows us to reuse cached JSON and HTML files
	return os.MkdirAll(webDir, 0755)
}

func deduplicateBuilds(packagesWithBuilds []*pack.PackageWithBuilds) []*pack.PackageWithBuilds {
	for i, pkg := range packagesWithBuilds {
		builderTargetMap := make(map[string]pack.BuiltFile)
		for _, builtFile := range pkg.BuiltFiles {
			key := builtFile.Builder + "/" + builtFile.Target
			if _, exists := builderTargetMap[key]; !exists {
				builderTargetMap[key] = builtFile
			}
		}

		var filteredBuilds []pack.BuiltFile
		for _, builtFile := range builderTargetMap {
			filteredBuilds = append(filteredBuilds, builtFile)
		}
		packagesWithBuilds[i].BuiltFiles = filteredBuilds
	}

	return packagesWithBuilds
}

func generateAllJSON(packagesWithBuilds []*pack.PackageWithBuilds, buildersMetadata map[string]*BuildMetadata, webDir string) {
	log.Println("  - Generating index.json...")
	GenerateIndexJSON(packagesWithBuilds, webDir)

	log.Printf("  - Generating %d package JSON files...", len(packagesWithBuilds))
	for _, pkg := range packagesWithBuilds {
		GeneratePackageJSON(pkg, webDir)
	}

	log.Printf("  - Generating %d builder JSON files...", len(builder.Builders))
	for builderName, metadata := range buildersMetadata {
		GenerateBuilderJSON(builderName, metadata, webDir)
	}

	log.Println("  - Generating builder metadata in built/ directories...")
	GenerateBuilderMetadataFiles(buildersMetadata)

	log.Println("  - Generating file details and source details JSON...")
	totalFiles := 0
	totalSources := 0
	for _, pkg := range packagesWithBuilds {
		totalFiles += len(pkg.BuiltFiles)
		totalSources += len(pkg.Package.Download)
	}

	var wg sync.WaitGroup
	semaphore := make(chan struct{}, runtime.NumCPU())

	for _, pkg := range packagesWithBuilds {
		// Generate built file details JSON
		for _, builtFile := range pkg.BuiltFiles {
			wg.Add(1)
			semaphore <- struct{}{}

			go func(p *pack.PackageWithBuilds, bf pack.BuiltFile) {
				defer wg.Done()
				defer func() { <-semaphore }()

				GenerateFileDetailsJSON(p, &bf, webDir)
			}(pkg, builtFile)
		}

		// Generate source file details JSON
		for i, download := range pkg.Package.Download {
			wg.Add(1)
			semaphore <- struct{}{}

			go func(p *pack.PackageWithBuilds, dl *pack.Download, idx int) {
				defer wg.Done()
				defer func() { <-semaphore }()

				GenerateSourceDetailsJSON(p, dl, idx, webDir)
			}(pkg, download, i)
		}
	}

	wg.Wait()
	log.Printf("  - Generated %d file detail JSON files and %d source detail JSON files", totalFiles, totalSources)
}

func printSummary(packagesWithBuilds []*pack.PackageWithBuilds) {
	totalFiles := 0
	for _, pkg := range packagesWithBuilds {
		totalFiles += len(pkg.BuiltFiles)
	}

	log.Println("\n=== Generation Summary ===")
	log.Printf("Total packages: %d", len(packagesWithBuilds))
	log.Printf("Total builders: %d", len(builder.Builders))
	log.Printf("Total targets: %d", len(host.SupportedHosts))
	log.Printf("Total built files: %d", totalFiles)
	log.Printf("HTML pages: %d", len(packagesWithBuilds)+len(builder.Builders)+totalFiles+1)
	log.Printf("JSON files: %d", len(packagesWithBuilds)+len(builder.Builders)+totalFiles+1+len(builder.Builders))
}

package buildweb

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/mrcyjanek/simplybs/builder"
	"github.com/mrcyjanek/simplybs/host"
	"github.com/mrcyjanek/simplybs/pack"
)

func GetAllPackagesWithBuilds() []*pack.PackageWithBuilds {
	packages := pack.GetAllPackages()
	packagesWithBuilds := make([]*pack.PackageWithBuilds, len(packages))

	for i, pkg := range packages {
		builtFiles := scanBuiltFilesForPackage(pkg.Package, pkg.Version)
		packagesWithBuilds[i] = &pack.PackageWithBuilds{
			Package:    pkg,
			BuiltFiles: builtFiles,
		}
	}

	return packagesWithBuilds
}

func scanBuiltFilesForPackage(packageName string, packageVersion string) []pack.BuiltFile {
	var builtFiles []pack.BuiltFile
	baseBuildDir := host.DataDirRoot()

	targets := getSortedTargets()

	for _, builderName := range builder.Builders {
		builderDir := filepath.Join(baseBuildDir, builderName)
		if _, err := os.Stat(builderDir); os.IsNotExist(err) {
			continue
		}

		for _, target := range targets {
			buildOutputDir := filepath.Join(builderDir, "built", target, packageName)
			buildOutputDir = filepath.Dir(buildOutputDir)

			if _, err := os.Stat(buildOutputDir); os.IsNotExist(err) {
				continue
			}

			files, err := os.ReadDir(buildOutputDir)
			if err != nil {
				continue
			}

			for _, file := range files {
				fileName := file.Name()
				fsFilepath := filepath.Join(buildOutputDir, fileName)
				if !strings.Contains(fsFilepath, packageName) {
					continue
				}
				if !strings.HasSuffix(fileName, ".tar.gz") {
					continue
				}

				nameWithoutExt := strings.TrimSuffix(fileName, ".tar.gz")
				parts := strings.Split(nameWithoutExt, "-")
				if len(parts) < 3 {
					continue
				}

				archPathRelative := filepath.Join(builderName, "built", target, filepath.Dir(packageName), fileName)
				infoPathRelative := filepath.Join(builderName, "built", target, filepath.Dir(packageName), strings.TrimSuffix(fileName, ".tar.gz")+".info.txt")

				fullArchPath := filepath.Join(baseBuildDir, archPathRelative)
				info, err := os.Stat(fullArchPath)
				var fileSize int64
				if err == nil {
					fileSize = info.Size()
				}
				id := strings.Split(parts[len(parts)-1], ".")[0]

				builtFiles = append(builtFiles, pack.BuiltFile{
					Builder:  builderName,
					Target:   target,
					ID:       id,
					InfoPath: infoPathRelative,
					ArchPath: archPathRelative,
					FileSize: fileSize,
				})
			}
		}
	}

	return builtFiles
}

func GetBuildersMetadata(packagesWithBuilds []*pack.PackageWithBuilds) map[string]*BuildMetadata {
	metadata := make(map[string]*BuildMetadata)
	targets := getSortedTargets()

	for _, builderName := range builder.Builders {
		meta := &BuildMetadata{
			Builder:  builderName,
			Targets:  targets,
			Packages: []PackageBuildStatus{},
			Stats: BuilderStats{
				TotalPackages: len(packagesWithBuilds),
				TotalTargets:  len(targets),
			},
		}

		for _, pkg := range packagesWithBuilds {
			status := PackageBuildStatus{
				Package: pkg.Package.Package,
				Version: pkg.Package.Version,
				Type:    pkg.Package.Type,
				Targets: make(map[string]*TargetBuildInfo),
			}

			for _, target := range targets {
				status.Targets[target] = &TargetBuildInfo{Built: false}
			}

			for _, builtFile := range pkg.BuiltFiles {
				if builtFile.Builder == builderName {
					status.Targets[builtFile.Target] = &TargetBuildInfo{
						Built:    true,
						BuildID:  builtFile.ID,
						FileSize: builtFile.FileSize,
						ArchPath: builtFile.ArchPath,
						InfoPath: builtFile.InfoPath,
					}
					meta.Stats.TotalBuilds++
				}
			}

			meta.Packages = append(meta.Packages, status)
		}

		totalPossible := meta.Stats.TotalPackages * meta.Stats.TotalTargets
		if totalPossible > 0 {
			meta.Stats.CompletionRate = float64(meta.Stats.TotalBuilds) / float64(totalPossible) * 100
		}

		metadata[builderName] = meta
	}

	return metadata
}

func ConvertToPackageMetadata(pkg *pack.PackageWithBuilds) *PackageMetadata {
	meta := &PackageMetadata{
		Package:      pkg.Package.Package,
		Version:      pkg.Package.Version,
		Type:         pkg.Package.Type,
		Downloads:    pkg.Package.Download,
		Dependencies: pkg.Package.Dependencies,
		BuildEnv:     pkg.Package.Build.Env,
		BuildSteps:   pkg.Package.Build.Steps,
		Builds:       make(map[string]map[string]*BuildInfo),
	}

	for _, builtFile := range pkg.BuiltFiles {
		if meta.Builds[builtFile.Builder] == nil {
			meta.Builds[builtFile.Builder] = make(map[string]*BuildInfo)
		}

		basePath := "../"
		depth := strings.Count(pkg.Package.Package, "/")
		for i := 0; i < depth; i++ {
			basePath += "../"
		}

		meta.Builds[builtFile.Builder][builtFile.Target] = &BuildInfo{
			Builder:     builtFile.Builder,
			Target:      builtFile.Target,
			BuildID:     builtFile.ID,
			FileSize:    builtFile.FileSize,
			ArchivePath: builtFile.ArchPath,
			InfoPath:    builtFile.InfoPath,
			DownloadURL: basePath + builtFile.ArchPath,
		}
	}

	return meta
}

func GetWebsiteIndex(packagesWithBuilds []*pack.PackageWithBuilds) *WebsiteIndex {
	index := &WebsiteIndex{
		TotalPackages: len(packagesWithBuilds),
		Builders:      getSortedBuilders(),
		Targets:       getSortedTargets(),
		Packages:      []PackageIndexEntry{},
	}

	totalCombinations := len(builder.Builders) * len(host.SupportedHosts)

	for _, pkg := range packagesWithBuilds {
		builderTargetMap := make(map[string]pack.BuiltFile)
		for _, builtFile := range pkg.BuiltFiles {
			key := builtFile.Builder + "/" + builtFile.Target
			if _, exists := builderTargetMap[key]; !exists {
				builderTargetMap[key] = builtFile
			}
		}

		actualBuilds := len(builderTargetMap)
		progress := 0
		if totalCombinations > 0 {
			progress = (actualBuilds * 100) / totalCombinations
		}

		entry := PackageIndexEntry{
			Package:        pkg.Package.Package,
			Version:        pkg.Package.Version,
			Type:           pkg.Package.Type,
			BuildProgress:  progress,
			TotalBuilds:    actualBuilds,
			PossibleBuilds: totalCombinations,
		}

		index.Packages = append(index.Packages, entry)
	}

	return index
}

func getSortedBuilders() []string {
	builders := make([]string, len(builder.Builders))
	copy(builders, builder.Builders)
	sort.Strings(builders)
	return builders
}

func getSortedTargets() []string {
	targets := make([]string, 0, len(host.SupportedHosts))
	for k := range host.SupportedHosts {
		targets = append(targets, k)
	}
	sort.Strings(targets)
	return targets
}

func DependencyExists(dep string, packages []*pack.Package) bool {
	for _, pkg := range packages {
		if pkg.Package == dep {
			return true
		}
	}
	return false
}

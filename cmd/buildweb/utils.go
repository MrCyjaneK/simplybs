package buildweb

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mrcyjanek/simplybs/pack"
)

func FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func GetGlobPattern(entry string) string {
	if strings.Contains(entry, ":") {
		parts := strings.SplitN(entry, ":", 2)
		return parts[0]
	}
	return ""
}

func GetGlobContent(entry string) string {
	if strings.Contains(entry, ":") {
		parts := strings.SplitN(entry, ":", 2)
		return parts[1]
	}
	return entry
}

func GetRelativePath(fromPackage, toPackage string) string {
	fromDepth := strings.Count(fromPackage, "/")
	upPath := strings.Repeat("../", fromDepth)
	if toPackage == "index" {
		return upPath + "index.html"
	}
	return upPath + toPackage + ".html"
}

func GetMirrorPath(pkg *pack.PackageWithBuilds, download *pack.Download) string {
	packageDepth := strings.Count(pkg.Package.Package, "/")
	upPath := strings.Repeat("../", packageDepth+1)

	if download.Kind == "git" {
		bundleName := filepath.Base(download.URL) + "-" + download.Sha256[0:8] + ".bundle"
		return fmt.Sprintf("%ssource/%s", upPath, bundleName)
	}
	return fmt.Sprintf("%ssource/%s", upPath, filepath.Base(download.URL))
}

func GetBuiltFilePath(packageName, filePath string) string {
	packageDepth := strings.Count(packageName, "/")
	upPath := strings.Repeat("../", packageDepth+1)
	return upPath + filePath
}

func GetFileDetailsPath(packageName, builder, target, packageVersion, buildID string) string {
	packageDepth := strings.Count(packageName, "/")
	upPath := strings.Repeat("../", packageDepth)
	fileName := fmt.Sprintf("%s-%s-%s.html", packageName, packageVersion, buildID)
	return fmt.Sprintf("%sfiles/%s/%s/%s", upPath, builder, target, fileName)
}

func GetBuildMatrix(pkg *pack.PackageWithBuilds) map[string]map[string]*pack.BuiltFile {
	builders := getSortedBuilders()
	targets := getSortedTargets()

	matrix := make(map[string]map[string]*pack.BuiltFile)
	for _, builder := range builders {
		matrix[builder] = make(map[string]*pack.BuiltFile)
		for _, target := range targets {
			matrix[builder][target] = nil
		}
	}

	for i := range pkg.BuiltFiles {
		bf := &pkg.BuiltFiles[i]
		if matrix[bf.Builder] != nil {
			matrix[bf.Builder][bf.Target] = bf
		}
	}

	return matrix
}

func GetBuildProgress(pkg *pack.PackageWithBuilds, totalBuilders, totalTargets int) int {
	totalCombinations := totalBuilders * totalTargets
	if totalCombinations == 0 {
		return 0
	}

	builderTargetMap := make(map[string]bool)
	for _, builtFile := range pkg.BuiltFiles {
		key := builtFile.Builder + "/" + builtFile.Target
		builderTargetMap[key] = true
	}

	actualBuilds := len(builderTargetMap)
	return (actualBuilds * 100) / totalCombinations
}

func GetSourceFilePath(packageName, sourcePath string) string {
	packageDepth := strings.Count(packageName, "/")
	upPath := strings.Repeat("../", packageDepth+1)
	return upPath + sourcePath
}

func GetSourceDetailsPath(packageName string, downloadIndex int) string {
	packageDepth := strings.Count(packageName, "/")
	upPath := strings.Repeat("../", packageDepth)
	return fmt.Sprintf("%ssource/%s-download-%d.html", upPath, packageName, downloadIndex)
}

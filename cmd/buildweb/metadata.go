package buildweb

import "github.com/mrcyjanek/simplybs/pack"

type BuildMetadata struct {
	Builder  string               `json:"builder"`
	Targets  []string             `json:"targets"`
	Packages []PackageBuildStatus `json:"packages"`
	Stats    BuilderStats         `json:"stats"`
}

type BuilderStats struct {
	TotalPackages  int     `json:"total_packages"`
	TotalTargets   int     `json:"total_targets"`
	TotalBuilds    int     `json:"total_builds"`
	CompletionRate float64 `json:"completion_rate"`
}

type PackageBuildStatus struct {
	Package string                      `json:"package"`
	Version string                      `json:"version"`
	Type    string                      `json:"type"`
	Targets map[string]*TargetBuildInfo `json:"targets"`
}

type TargetBuildInfo struct {
	Built    bool   `json:"built"`
	BuildID  string `json:"build_id,omitempty"`
	FileSize int64  `json:"file_size,omitempty"`
	ArchPath string `json:"arch_path,omitempty"`
	InfoPath string `json:"info_path,omitempty"`
}

type PackageMetadata struct {
	Package      string                           `json:"package"`
	Version      string                           `json:"version"`
	Type         string                           `json:"type"`
	Downloads    []*pack.Download                 `json:"downloads"`
	Dependencies []string                         `json:"dependencies"`
	BuildEnv     []string                         `json:"build_env"`
	BuildSteps   []string                         `json:"build_steps"`
	Builds       map[string]map[string]*BuildInfo `json:"builds"` // builder -> target -> BuildInfo
}

type BuildInfo struct {
	Builder     string `json:"builder"`
	Target      string `json:"target"`
	BuildID     string `json:"build_id"`
	FileSize    int64  `json:"file_size"`
	ArchivePath string `json:"archive_path"`
	InfoPath    string `json:"info_path"`
	DownloadURL string `json:"download_url"`
}

type ArchiveFileInfo struct {
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type ArchiveInfo struct {
	Files     []ArchiveFileInfo `json:"files"`
	TotalSize int64             `json:"total_size"`
	FileCount int               `json:"file_count"`
}

type WebsiteIndex struct {
	TotalPackages int                 `json:"total_packages"`
	Builders      []string            `json:"builders"`
	Targets       []string            `json:"targets"`
	Packages      []PackageIndexEntry `json:"packages"`
}

type PackageIndexEntry struct {
	Package        string `json:"package"`
	Version        string `json:"version"`
	Type           string `json:"type"`
	BuildProgress  int    `json:"build_progress"`
	TotalBuilds    int    `json:"total_builds"`
	PossibleBuilds int    `json:"possible_builds"`
}

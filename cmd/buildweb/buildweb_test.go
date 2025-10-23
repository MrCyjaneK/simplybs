package buildweb

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mrcyjanek/simplybs/pack"
)

func TestFormatFileSize(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{1073741824, "1.0 GB"},
		{1099511627776, "1.0 TB"},
	}

	for _, tt := range tests {
		result := FormatFileSize(tt.bytes)
		if result != tt.expected {
			t.Errorf("FormatFileSize(%d) = %s; want %s", tt.bytes, result, tt.expected)
		}
	}
}

func TestGetGlobPattern(t *testing.T) {
	tests := []struct {
		entry    string
		expected string
	}{
		{"*:content", "*"},
		{"pattern:value", "pattern"},
		{"nopattern", ""},
		{"x86*:CC=gcc", "x86*"},
	}

	for _, tt := range tests {
		result := GetGlobPattern(tt.entry)
		if result != tt.expected {
			t.Errorf("GetGlobPattern(%q) = %q; want %q", tt.entry, result, tt.expected)
		}
	}
}

func TestGetGlobContent(t *testing.T) {
	tests := []struct {
		entry    string
		expected string
	}{
		{"*:content", "content"},
		{"pattern:value", "value"},
		{"nopattern", "nopattern"},
		{"x86*:CC=gcc", "CC=gcc"},
	}

	for _, tt := range tests {
		result := GetGlobContent(tt.entry)
		if result != tt.expected {
			t.Errorf("GetGlobContent(%q) = %q; want %q", tt.entry, result, tt.expected)
		}
	}
}

func TestGetRelativePath(t *testing.T) {
	tests := []struct {
		from     string
		to       string
		expected string
	}{
		{"boost", "index", "index.html"},                       // web/boost.html -> web/index.html
		{"native/cmake", "index", "../index.html"},             // web/native/cmake.html -> web/index.html
		{"boost", "zlib", "zlib.html"},                         // web/boost.html -> web/zlib.html
		{"native/cmake", "native/perl", "../native/perl.html"}, // web/native/cmake.html -> web/native/perl.html
	}

	for _, tt := range tests {
		result := GetRelativePath(tt.from, tt.to)
		if result != tt.expected {
			t.Errorf("GetRelativePath(%q, %q) = %q; want %q", tt.from, tt.to, result, tt.expected)
		}
	}
}

func TestGetBuildProgress(t *testing.T) {
	tests := []struct {
		name          string
		builtFiles    []pack.BuiltFile
		totalBuilders int
		totalTargets  int
		expected      int
	}{
		{
			name:          "no builds",
			builtFiles:    []pack.BuiltFile{},
			totalBuilders: 3,
			totalTargets:  10,
			expected:      0,
		},
		{
			name: "50% complete",
			builtFiles: []pack.BuiltFile{
				{Builder: "darwin_arm64", Target: "aarch64-apple-ios"},
				{Builder: "darwin_arm64", Target: "aarch64-apple-darwin"},
				{Builder: "linux_amd64", Target: "aarch64-apple-ios"},
				{Builder: "linux_amd64", Target: "aarch64-apple-darwin"},
				{Builder: "linux_arm64", Target: "aarch64-apple-ios"},
				{Builder: "linux_arm64", Target: "aarch64-apple-darwin"},
			},
			totalBuilders: 3,
			totalTargets:  4,
			expected:      50,
		},
		{
			name: "100% complete",
			builtFiles: []pack.BuiltFile{
				{Builder: "darwin_arm64", Target: "aarch64-apple-ios"},
				{Builder: "darwin_arm64", Target: "aarch64-apple-darwin"},
				{Builder: "linux_amd64", Target: "aarch64-apple-ios"},
				{Builder: "linux_amd64", Target: "aarch64-apple-darwin"},
			},
			totalBuilders: 2,
			totalTargets:  2,
			expected:      100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkg := &pack.PackageWithBuilds{
				Package:    &pack.Package{},
				BuiltFiles: tt.builtFiles,
			}
			result := GetBuildProgress(pkg, tt.totalBuilders, tt.totalTargets)
			if result != tt.expected {
				t.Errorf("GetBuildProgress() = %d; want %d", result, tt.expected)
			}
		})
	}
}

func TestDependencyExists(t *testing.T) {
	packages := []*pack.Package{
		{Package: "boost"},
		{Package: "zlib"},
		{Package: "native/cmake"},
	}

	tests := []struct {
		dep      string
		expected bool
	}{
		{"boost", true},
		{"zlib", true},
		{"native/cmake", true},
		{"nonexistent", false},
		{"native/perl", false},
	}

	for _, tt := range tests {
		result := DependencyExists(tt.dep, packages)
		if result != tt.expected {
			t.Errorf("DependencyExists(%q) = %v; want %v", tt.dep, result, tt.expected)
		}
	}
}

func TestGetBuiltFilePath(t *testing.T) {
	tests := []struct {
		packageName string
		filePath    string
		expected    string
	}{
		// From web/boost.html, go up to root: ../
		{"boost", "darwin_arm64/built/target/boost-1.0-abc.tar.gz", "../darwin_arm64/built/target/boost-1.0-abc.tar.gz"},
		// From web/native/cmake.html, go up to root: ../../
		{"native/cmake", "linux_amd64/built/target/file.tar.gz", "../../linux_amd64/built/target/file.tar.gz"},
	}

	for _, tt := range tests {
		result := GetBuiltFilePath(tt.packageName, tt.filePath)
		if result != tt.expected {
			t.Errorf("GetBuiltFilePath(%q, %q) = %q; want %q", tt.packageName, tt.filePath, result, tt.expected)
		}
	}
}

func TestConvertToPackageMetadata(t *testing.T) {
	pkg := &pack.PackageWithBuilds{
		Package: &pack.Package{
			Package: "boost",
			Version: "1.0.0",
			Type:    "host",
			Download: []*pack.Download{
				{Kind: "http", URL: "https://example.com/boost.tar.gz", Sha256: "abc123"},
			},
			Dependencies: []string{"zlib"},
		},
		BuiltFiles: []pack.BuiltFile{
			{
				Builder:  "darwin_arm64",
				Target:   "aarch64-apple-ios",
				ID:       "abc12345",
				InfoPath: "darwin_arm64/built/aarch64-apple-ios/boost-1.0.0-abc12345.info.txt",
				ArchPath: "darwin_arm64/built/aarch64-apple-ios/boost-1.0.0-abc12345.tar.gz",
				FileSize: 1024,
			},
		},
	}

	metadata := ConvertToPackageMetadata(pkg)

	if metadata.Package != "boost" {
		t.Errorf("Package name = %q; want %q", metadata.Package, "boost")
	}

	if metadata.Version != "1.0.0" {
		t.Errorf("Version = %q; want %q", metadata.Version, "1.0.0")
	}

	if len(metadata.Builds) != 1 {
		t.Errorf("Number of builders = %d; want 1", len(metadata.Builds))
	}

	if _, exists := metadata.Builds["darwin_arm64"]; !exists {
		t.Error("darwin_arm64 builder not found in builds")
	}

	if len(metadata.Builds["darwin_arm64"]) != 1 {
		t.Errorf("Number of targets for darwin_arm64 = %d; want 1", len(metadata.Builds["darwin_arm64"]))
	}
}

func TestGetBuildMatrix(t *testing.T) {
	pkg := &pack.PackageWithBuilds{
		Package: &pack.Package{Package: "boost"},
		BuiltFiles: []pack.BuiltFile{
			{Builder: "darwin_arm64", Target: "aarch64-apple-ios", ID: "abc"},
			{Builder: "linux_amd64", Target: "x86_64-linux-gnu", ID: "def"},
		},
	}

	matrix := GetBuildMatrix(pkg)

	if len(matrix) == 0 {
		t.Error("Matrix should not be empty")
	}

	if matrix["darwin_arm64"]["aarch64-apple-ios"] == nil {
		t.Error("Expected darwin_arm64/aarch64-apple-ios to be in matrix")
	}

	if matrix["linux_amd64"]["x86_64-linux-gnu"] == nil {
		t.Error("Expected linux_amd64/x86_64-linux-gnu to be in matrix")
	}
}

func TestGetWebsiteIndex(t *testing.T) {
	packagesWithBuilds := []*pack.PackageWithBuilds{
		{
			Package: &pack.Package{
				Package: "boost",
				Version: "1.0.0",
				Type:    "host",
			},
			BuiltFiles: []pack.BuiltFile{
				{Builder: "darwin_arm64", Target: "aarch64-apple-ios"},
			},
		},
		{
			Package: &pack.Package{
				Package: "zlib",
				Version: "1.2.11",
				Type:    "host",
			},
			BuiltFiles: []pack.BuiltFile{},
		},
	}

	index := GetWebsiteIndex(packagesWithBuilds)

	if index.TotalPackages != 2 {
		t.Errorf("TotalPackages = %d; want 2", index.TotalPackages)
	}

	if len(index.Packages) != 2 {
		t.Errorf("Number of packages = %d; want 2", len(index.Packages))
	}

	if len(index.Builders) == 0 {
		t.Error("Builders list should not be empty")
	}

	if len(index.Targets) == 0 {
		t.Error("Targets list should not be empty")
	}
}

func TestGetArchiveInfo(t *testing.T) {
	testPkg := &pack.Package{Package: "test", Version: "1.0"}
	testDownload := &pack.Download{Kind: "tar.gz", URL: "file:///nonexistent/file.tar.gz"}
	info := GetArchiveInfo(testPkg, testDownload)

	if info.FileCount != 0 {
		t.Errorf("FileCount for nonexistent file = %d; want 0", info.FileCount)
	}

	if info.TotalSize != 0 {
		t.Errorf("TotalSize for nonexistent file = %d; want 0", info.TotalSize)
	}

	if len(info.Files) != 0 {
		t.Errorf("Files length for nonexistent file = %d; want 0", len(info.Files))
	}
}

func TestDeduplicateBuilds(t *testing.T) {
	packagesWithBuilds := []*pack.PackageWithBuilds{
		{
			Package: &pack.Package{Package: "boost"},
			BuiltFiles: []pack.BuiltFile{
				{Builder: "darwin_arm64", Target: "aarch64-apple-ios", ID: "abc"},
				{Builder: "darwin_arm64", Target: "aarch64-apple-ios", ID: "def"}, // Duplicate
				{Builder: "linux_amd64", Target: "x86_64-linux-gnu", ID: "ghi"},
			},
		},
	}

	result := deduplicateBuilds(packagesWithBuilds)

	if len(result[0].BuiltFiles) != 2 {
		t.Errorf("After deduplication, expected 2 builds, got %d", len(result[0].BuiltFiles))
	}
}

func TestGetBuildersMetadata(t *testing.T) {
	packagesWithBuilds := []*pack.PackageWithBuilds{
		{
			Package: &pack.Package{
				Package: "boost",
				Version: "1.0.0",
				Type:    "host",
			},
			BuiltFiles: []pack.BuiltFile{
				{Builder: "darwin_arm64", Target: "aarch64-apple-ios", FileSize: 1024},
			},
		},
	}

	metadata := GetBuildersMetadata(packagesWithBuilds)

	if len(metadata) == 0 {
		t.Error("Metadata should not be empty")
	}

	for _, builderName := range []string{"darwin_arm64", "linux_amd64", "linux_arm64"} {
		if meta, exists := metadata[builderName]; !exists {
			t.Errorf("Metadata for builder %s not found", builderName)
		} else {
			if meta.Builder != builderName {
				t.Errorf("Builder name in metadata = %q; want %q", meta.Builder, builderName)
			}
			if len(meta.Packages) != 1 {
				t.Errorf("Number of packages in metadata = %d; want 1", len(meta.Packages))
			}
		}
	}
}

func TestSetupWebDirectory(t *testing.T) {
	tmpDir := filepath.Join(os.TempDir(), "buildweb_test")
	defer os.RemoveAll(tmpDir)

	err := setupWebDirectory(tmpDir)
	if err != nil {
		t.Fatalf("setupWebDirectory failed: %v", err)
	}

	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Error("Web directory was not created")
	}

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	err = setupWebDirectory(tmpDir)
	if err != nil {
		t.Fatalf("setupWebDirectory failed on second call: %v", err)
	}

	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Existing files should be preserved for caching")
	}

	if _, err := os.Stat(tmpDir); os.IsNotExist(err) {
		t.Error("Web directory does not exist after second call")
	}
}

func TestGetSortedBuilders(t *testing.T) {
	builders := getSortedBuilders()

	if len(builders) == 0 {
		t.Error("Builders list should not be empty")
	}

	for i := 1; i < len(builders); i++ {
		if builders[i-1] > builders[i] {
			t.Errorf("Builders list is not sorted: %v", builders)
			break
		}
	}
}

func TestGetSortedTargets(t *testing.T) {
	targets := getSortedTargets()

	if len(targets) == 0 {
		t.Error("Targets list should not be empty")
	}

	for i := 1; i < len(targets); i++ {
		if targets[i-1] > targets[i] {
			t.Errorf("Targets list is not sorted: %v", targets)
			break
		}
	}
}

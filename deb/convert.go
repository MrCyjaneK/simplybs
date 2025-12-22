package deb

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mrcyjanek/simplybs/host"
)

func ConvertTarGzToDeb(tarGzPath, debPath, packageName, version string, h *host.Host) error {
	return ConvertTarGzToDebWithSources(tarGzPath, debPath, packageName, version, nil, h)
}

func ConvertTarGzToDebWithSources(tarGzPath, debPath, packageName, version string, sources []SourceInfo, h *host.Host) error {
	tmpDir, err := os.MkdirTemp("", "deb-build-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	debRoot := filepath.Join(tmpDir, "deb")
	dataDir := filepath.Join(debRoot, "data")
	controlDir := filepath.Join(debRoot, "control")

	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}
	if err := os.MkdirAll(controlDir, 0755); err != nil {
		return fmt.Errorf("failed to create control directory: %w", err)
	}

	installPrefix := h.GetEnvPath()
	if err := extractTarGzToDebData(tarGzPath, dataDir, installPrefix); err != nil {
		return fmt.Errorf("failed to extract tar.gz: %w", err)
	}

	if err := createControlFiles(controlDir, packageName, version, sources, h.Triplet); err != nil {
		return fmt.Errorf("failed to create control files: %w", err)
	}

	if err := createDebArchive(debRoot, debPath); err != nil {
		return fmt.Errorf("failed to create deb archive: %w", err)
	}

	return nil
}

type SourceInfo struct {
	Kind string
	URL  string
}

func extractTarGzToDebData(tarGzPath, dataDir, installPrefix string) error {
	file, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if header.Name == "." || header.Name == "" {
			continue
		}

		targetPath := filepath.Join(dataDir, installPrefix, header.Name)

		if header.Typeflag == tar.TypeDir {
			if err := os.MkdirAll(targetPath, os.FileMode(header.Mode)); err != nil {
				return err
			}
			continue
		}

		if header.Typeflag == tar.TypeSymlink {
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}
			if err := os.Symlink(header.Linkname, targetPath); err != nil {
				return err
			}
			continue
		}

		if header.Typeflag == tar.TypeReg || header.Typeflag == tar.TypeRegA {
			if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
				return err
			}

			outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}

func createControlFiles(controlDir, packageName, version string, sources []SourceInfo, triplet string) error {
	safePackageName := strings.ReplaceAll(packageName, "/", "-")
	safePackageName = strings.ReplaceAll(safePackageName, "@", "-")
	safePackageName = "sbs-" + safePackageName

	debVersion := sanitizeDebVersion(version)

	description := buildDescription(packageName, triplet, sources)

	controlContent := fmt.Sprintf(`Package: %s
Version: %s
Architecture: %s
Maintainer: simplybs <simplybs@mrcyjanek.net>
Description: %s
Section: devel
Priority: optional
`,
		safePackageName,
		debVersion,
		getDebArchitecture(triplet),
		description,
	)

	controlPath := filepath.Join(controlDir, "control")
	if err := os.WriteFile(controlPath, []byte(controlContent), 0644); err != nil {
		return err
	}

	return nil
}

func sanitizeDebVersion(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return "1.0"
	}

	if idx := strings.Index(version, ":"); idx != -1 {
		upstream := version[idx+1:]
		if len(upstream) > 0 && !(upstream[0] >= '0' && upstream[0] <= '9') {
			return version[:idx+1] + "0" + upstream
		}
		return version
	}

	if !(version[0] >= '0' && version[0] <= '9') {
		return "0:0" + version
	}
	return version
}

func isAlphanumeric(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

func buildDescription(packageName, triplet string, sources []SourceInfo) string {
	var desc strings.Builder

	desc.WriteString(fmt.Sprintf("%s built for %s", packageName, triplet))

	if len(sources) > 0 {
		desc.WriteString("\n")
		desc.WriteString(" Built from sources:")
		for _, src := range sources {
			desc.WriteString("\n")
			switch src.Kind {
			case "git":
				desc.WriteString(fmt.Sprintf("  - Git repository: %s", src.URL))
			case "tar.gz", "tar.bz2", "tar.xz":
				desc.WriteString(fmt.Sprintf("  - Archive: %s", src.URL))
			case "blob":
				desc.WriteString(fmt.Sprintf("  - File: %s", src.URL))
			default:
				if src.Kind != "" {
					desc.WriteString(fmt.Sprintf("  - %s: %s", src.Kind, src.URL))
				} else {
					desc.WriteString(fmt.Sprintf("  - %s", src.URL))
				}
			}
		}
	}

	desc.WriteString("\n")
	desc.WriteString(" This package was automatically generated by simplybs.")

	return desc.String()
}

func getDebArchitecture(triplet string) string {
	if strings.Contains(triplet, "aarch64") {
		if strings.Contains(triplet, "apple") {
			return "arm64"
		}
		return "arm64"
	}
	if strings.Contains(triplet, "x86_64") {
		return "amd64"
	}
	if strings.Contains(triplet, "armv7") {
		return "armhf"
	}
	return "all"
}

func createDebArchive(debRoot, debPath string) error {
	if err := os.MkdirAll(filepath.Dir(debPath), 0755); err != nil {
		return err
	}

	outFile, err := os.Create(debPath)
	if err != nil {
		return err
	}
	defer outFile.Close()

	if _, err := outFile.WriteString("!<arch>\n"); err != nil {
		return err
	}

	debianBinary := "2.0\n"
	if err := writeArEntryString(outFile, "debian-binary", debianBinary); err != nil {
		return err
	}

	controlTarPath := filepath.Join(debRoot, "control.tar.gz")
	if err := createTarGz(filepath.Join(debRoot, "control"), controlTarPath); err != nil {
		return err
	}
	if err := writeArEntry(outFile, "control.tar.gz", controlTarPath); err != nil {
		return err
	}

	dataTarPath := filepath.Join(debRoot, "data.tar.gz")
	if err := createTarGz(filepath.Join(debRoot, "data"), dataTarPath); err != nil {
		return err
	}
	if err := writeArEntry(outFile, "data.tar.gz", dataTarPath); err != nil {
		return err
	}

	return nil
}

func writeArEntry(outFile *os.File, name, filePath string) error {
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	return writeArEntryFromReader(outFile, name, file, info.Size())
}

func writeArEntryString(outFile *os.File, name, content string) error {
	return writeArEntryFromReader(outFile, name, strings.NewReader(content), int64(len(content)))
}

func writeArEntryFromReader(outFile *os.File, name string, reader io.Reader, size int64) error {
	// AR header format: name (16 bytes), timestamp (12 bytes), owner/group (6+6 bytes), mode (8 bytes), size (10 bytes), magic (2 bytes)
	header := make([]byte, 60)

	// Name (16 bytes, space-padded on the right)
	nameBytes := []byte(name)
	if len(nameBytes) > 16 {
		nameBytes = nameBytes[:16]
	}
	copy(header[0:16], nameBytes)

	// Pad with spaces
	for i := len(nameBytes); i < 16; i++ {
		header[i] = ' '
	}

	// Timestamp (12 bytes)
	timestamp := fmt.Sprintf("%-12d", time.Now().Unix())
	copy(header[16:28], []byte(timestamp))

	// Owner/Group (6+6 bytes, default to 0)
	copy(header[28:34], []byte("0     "))
	copy(header[34:40], []byte("0     "))

	// Mode (8 bytes, default to 644)
	copy(header[40:48], []byte("644     "))

	// Size (10 bytes)
	sizeStr := fmt.Sprintf("%-10d", size)
	copy(header[48:58], []byte(sizeStr))

	// Magic (2 bytes)
	copy(header[58:60], []byte("`\n"))

	if _, err := outFile.Write(header); err != nil {
		return err
	}

	if _, err := io.Copy(outFile, reader); err != nil {
		return err
	}

	if size%2 != 0 {
		if _, err := outFile.Write([]byte("\n")); err != nil {
			return err
		}
	}

	return nil
}

func createTarGz(sourceDir, archivePath string) error {
	file, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzw, err := gzip.NewWriterLevel(file, gzip.BestCompression)
	if err != nil {
		return err
	}
	defer gzw.Close()

	tw := tar.NewWriter(gzw)
	defer tw.Close()

	fixedTime := time.Unix(1, 0)

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == sourceDir {
			return nil
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(relPath)
		header.ModTime = fixedTime
		header.AccessTime = fixedTime
		header.ChangeTime = fixedTime

		if info.Mode()&os.ModeSymlink != 0 {
			linkTarget, err := os.Readlink(path)
			if err != nil {
				return err
			}
			header.Typeflag = tar.TypeSymlink
			header.Linkname = linkTarget
			header.Size = 0
		}

		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		if info.Mode().IsRegular() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			if _, err := io.Copy(tw, file); err != nil {
				return err
			}
		}

		return nil
	})
}

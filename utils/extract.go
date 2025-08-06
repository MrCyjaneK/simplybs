package utils

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/ulikunitz/xz"
)

type readerFactory func() (*tar.Reader, func(), error)

func detectCommonPrefix(readerFactory readerFactory) (string, error) {
	tr, cleanup, err := readerFactory()
	if err != nil {
		return "", err
	}
	defer cleanup()

	firstLevelDirs := make(map[string]int)
	rootFiles := 0

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		if header.Name == "pax_global_header" || header.Name == "." {
			continue
		}

		parts := strings.Split(header.Name, "/")
		if len(parts) == 1 {
			if !header.FileInfo().IsDir() {
				rootFiles++
			}
		} else {
			firstLevelDirs[parts[0]]++
		}
	}

	if len(firstLevelDirs) == 1 && rootFiles == 0 {
		for dirName := range firstLevelDirs {
			return dirName + "/", nil
		}
	}

	return "", nil
}

func extractTar(tr *tar.Reader, destPath, commonPrefix string) error {
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		targetName := header.Name
		if commonPrefix != "" && strings.HasPrefix(header.Name, commonPrefix) {
			targetName = strings.TrimPrefix(header.Name, commonPrefix)
		}

		if targetName == "" {
			continue
		}

		target := filepath.Join(destPath, targetName)

		if !filepath.HasPrefix(target, filepath.Clean(destPath)+string(os.PathSeparator)) {
			log.Printf("Skipping entry outside target directory: %s", header.Name)
			continue
		}

		switch header.Typeflag {
		case tar.TypeDir:
			dirMode := os.FileMode(header.Mode) & 0777
			if dirMode == 0 {
				dirMode = 0755
			}
			if dirMode&0111 == 0 {
				dirMode |= 0755
			}

			if err := os.MkdirAll(target, dirMode); err != nil {
				return err
			}

			if err := os.Chtimes(target, header.AccessTime, header.ModTime); err != nil {
				log.Printf("Warning: Failed to set timestamps for directory %s: %v", target, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}

			fileMode := os.FileMode(header.Mode) & 0777
			if fileMode == 0 {
				fileMode = 0644
			}
			fileMode &^= (os.ModeSetuid | os.ModeSetgid | os.ModeSticky)

			os.Remove(target)

			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, fileMode)
			if err != nil {
				return err
			}

			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()

			if err := os.Chtimes(target, header.AccessTime, header.ModTime); err != nil {
				log.Printf("Warning: Failed to set timestamps for file %s: %v", target, err)
			}
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}

			os.Remove(target)

			if err := os.Symlink(header.Linkname, target); err != nil {
				log.Printf("Warning: Failed to create symbolic link %s -> %s: %v", target, header.Linkname, err)
			} else {
				if err := os.Chtimes(target, header.AccessTime, header.ModTime); err != nil {
					log.Printf("Warning: Failed to set timestamps for symlink %s: %v", target, err)
				}
			}
		}
	}

	return nil
}

func createGzipTarReader(archivePath string) (*tar.Reader, func(), error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return nil, nil, err
	}

	gzr, err := gzip.NewReader(file)
	if err != nil {
		file.Close()
		return nil, nil, err
	}

	tr := tar.NewReader(gzr)
	cleanup := func() {
		gzr.Close()
		file.Close()
	}

	return tr, cleanup, nil
}

func createBzip2TarReader(archivePath string) (*tar.Reader, func(), error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return nil, nil, err
	}

	bzr := bzip2.NewReader(file)
	tr := tar.NewReader(bzr)
	cleanup := func() {
		file.Close()
	}

	return tr, cleanup, nil
}

func createXzTarReader(archivePath string) (*tar.Reader, func(), error) {
	file, err := os.Open(archivePath)
	if err != nil {
		return nil, nil, err
	}

	xzr, err := xz.NewReader(file)
	if err != nil {
		file.Close()
		return nil, nil, err
	}

	tr := tar.NewReader(xzr)
	cleanup := func() {
		file.Close()
	}

	return tr, cleanup, nil
}

func ExtractTarGz(archivePath, destPath string) error {
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		log.Printf("Archive not found: %s", archivePath)
		return err
	}

	log.Printf("Extracting archive: %s into %s", archivePath, destPath)

	readerFactory := func() (*tar.Reader, func(), error) {
		return createGzipTarReader(archivePath)
	}

	commonPrefix, err := detectCommonPrefix(readerFactory)
	if err != nil {
		return err
	}

	tr, cleanup, err := readerFactory()
	if err != nil {
		return err
	}
	defer cleanup()

	if err := extractTar(tr, destPath, commonPrefix); err != nil {
		return err
	}

	if commonPrefix != "" {
		log.Printf("Stripped common directory prefix: %s", commonPrefix)
	}

	return nil
}

func ExtractTarBz2(archivePath, destPath string) error {
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		log.Printf("Archive not found: %s", archivePath)
		return err
	}

	log.Printf("Extracting bz2 archive: %s into %s", archivePath, destPath)

	readerFactory := func() (*tar.Reader, func(), error) {
		return createBzip2TarReader(archivePath)
	}

	commonPrefix, err := detectCommonPrefix(readerFactory)
	if err != nil {
		return err
	}

	tr, cleanup, err := readerFactory()
	if err != nil {
		return err
	}
	defer cleanup()

	if err := extractTar(tr, destPath, commonPrefix); err != nil {
		return err
	}

	if commonPrefix != "" {
		log.Printf("Stripped common directory prefix: %s", commonPrefix)
	}

	return nil
}

func ExtractTarXz(archivePath, destPath string) error {
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		log.Printf("Archive not found: %s", archivePath)
		return err
	}

	log.Printf("Extracting xz archive: %s into %s", archivePath, destPath)

	readerFactory := func() (*tar.Reader, func(), error) {
		return createXzTarReader(archivePath)
	}

	commonPrefix, err := detectCommonPrefix(readerFactory)
	if err != nil {
		return err
	}

	tr, cleanup, err := readerFactory()
	if err != nil {
		return err
	}
	defer cleanup()

	if err := extractTar(tr, destPath, commonPrefix); err != nil {
		return err
	}

	if commonPrefix != "" {
		log.Printf("Stripped common directory prefix: %s", commonPrefix)
	}

	return nil
}

func writeFileToTar(tw *tar.Writer, header *tar.Header, filePath string) error {
	if err := tw.WriteHeader(header); err != nil {
		return err
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(tw, file)
	return err
}

func CreateTarGz(sourcePath, archivePath string) error {
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

	log.Printf("Creating archive: %s from %s", archivePath, sourcePath)

	var filePaths []string
	err = filepath.Walk(sourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if path == sourcePath {
			return nil
		}

		filePaths = append(filePaths, path)
		return nil
	})
	if err != nil {
		return err
	}

	sort.Strings(filePaths)

	fixedTime := time.Unix(1, 0)

	for _, path := range filePaths {
		info, err := os.Lstat(path)
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourcePath, path)
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

		if info.Mode().IsRegular() {
			var filePath string
			var cleanup func()

			if strings.HasSuffix(strings.ToLower(relPath), ".a") {
				// TODO: repack static libraries
				filePath = path
				cleanup = func() {}
			} else {
				filePath = path
				cleanup = func() {}
			}

			defer cleanup()
			if err := writeFileToTar(tw, header, filePath); err != nil {
				return err
			}
		} else {
			if err := tw.WriteHeader(header); err != nil {
				return err
			}
		}
	}

	return nil
}

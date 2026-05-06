package utils

import (
	"log"
	"os"
	"path/filepath"
)

func RemoveAll(path string) error {
	filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		os.Chmod(p, 0755)
		return nil
	})

	if err := os.RemoveAll(path); err != nil {
		log.Fatalf("removeAll %s: %w", path, err)
	}
	return nil
}

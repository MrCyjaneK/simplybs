package utils

import (
	"log"
	"os"
	"path/filepath"
)

func RemoveAll(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}

	filepath.WalkDir(path, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		os.Chmod(p, 0755)
		return nil
	})

	if err := os.RemoveAll(path); err != nil {
		log.Fatalf("removeAll %s: %v", path, err)
	}

	if _, err := os.Stat(path); err == nil {
		log.Fatalf("removeAll %s: still exists after removal", path)
	} else if !os.IsNotExist(err) {
		log.Fatalf("removeAll %s: %v", path, err)
	}
	return nil
}

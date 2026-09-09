// Command mvnpom moves blob-POM install packages from
// packages/native/jdk@22/bootstrap-maven/<name>.json to
// packages/native/jdk@22/bootstrap-maven/pom/<name>.json and rewrites
// "package" / dependency names accordingly.
//
//	go run ./cmd/mvnpom
package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	srcDir    = "packages/native/jdk@22/bootstrap-maven"
	dstDir    = "packages/native/jdk@22/bootstrap-maven/pom"
	oldPrefix = "native/jdk@22/bootstrap-maven/"
	newPrefix = "native/jdk@22/bootstrap-maven/pom/"
)

type pkgFile struct {
	Package      string                   `json:"package"`
	Version      string                   `json:"version"`
	Type         string                   `json:"type"`
	Download     []map[string]interface{} `json:"download,omitempty"`
	Dependencies []string                 `json:"dependencies,omitempty"`
	Build        map[string]interface{}   `json:"build,omitempty"`
}

type move struct {
	oldPath, newPath string
	oldName, newName string
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run() error {
	entries, err := os.ReadDir(srcDir)
	if err != nil {
		return err
	}

	var moves []move
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		oldPath := filepath.Join(srcDir, e.Name())
		raw, err := os.ReadFile(oldPath)
		if err != nil {
			return err
		}
		if !isPomInstall(raw) {
			continue
		}
		var pkg pkgFile
		if err := json.Unmarshal(raw, &pkg); err != nil {
			return fmt.Errorf("%s: %w", oldPath, err)
		}
		if !strings.HasPrefix(pkg.Package, oldPrefix) || strings.HasPrefix(pkg.Package, newPrefix) {
			return fmt.Errorf("%s: unexpected package name %q", oldPath, pkg.Package)
		}
		name := strings.TrimPrefix(pkg.Package, oldPrefix)
		if name != strings.TrimSuffix(e.Name(), ".json") {
			return fmt.Errorf("%s: package %q does not match filename", oldPath, pkg.Package)
		}
		moves = append(moves, move{
			oldPath: oldPath,
			newPath: filepath.Join(dstDir, e.Name()),
			oldName: pkg.Package,
			newName: newPrefix + name,
		})
	}
	sort.Slice(moves, func(i, j int) bool { return moves[i].oldName < moves[j].oldName })
	if len(moves) == 0 {
		fmt.Println("no pom-install packages left to move")
		return nil
	}

	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}
	for _, m := range moves {
		raw, err := os.ReadFile(m.oldPath)
		if err != nil {
			return err
		}
		raw = rewriteNames(raw, []move{m})
		if err := os.WriteFile(m.newPath, raw, 0644); err != nil {
			return err
		}
		if err := os.Remove(m.oldPath); err != nil {
			return err
		}
		fmt.Printf("moved %s -> %s\n", m.oldName, m.newName)
	}

	n, err := rewriteTree("packages", moves)
	if err != nil {
		return err
	}
	fmt.Printf("rewrote dependencies in %d files (%d packages)\n", n, len(moves))
	return nil
}

func rewriteNames(raw []byte, moves []move) []byte {
	// Longest names first so commons-parent@9 does not eat @91/@98.
	sort.Slice(moves, func(i, j int) bool { return len(moves[i].oldName) > len(moves[j].oldName) })
	for _, m := range moves {
		raw = bytes.ReplaceAll(raw, []byte(m.oldName+`"`), []byte(m.newName+`"`))
	}
	return raw
}

func rewriteTree(root string, moves []move) (int, error) {
	changed := 0
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".json") {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		next := rewriteNames(raw, moves)
		if bytes.Equal(raw, next) {
			return nil
		}
		if err := os.WriteFile(path, next, 0644); err != nil {
			return err
		}
		changed++
		return nil
	})
	return changed, err
}

func isPomInstall(raw []byte) bool {
	var pkg pkgFile
	if err := json.Unmarshal(raw, &pkg); err != nil {
		return false
	}
	if len(pkg.Download) == 0 {
		return false
	}
	for _, d := range pkg.Download {
		kind, _ := d["kind"].(string)
		path, _ := d["path"].(string)
		url, _ := d["url"].(string)
		if kind != "blob" || !(strings.HasSuffix(path, ".pom") || strings.HasSuffix(url, ".pom")) {
			return false
		}
	}
	steps, _ := pkg.Build["steps"].([]interface{})
	if len(steps) == 0 {
		return false
	}
	for _, s := range steps {
		step, _ := s.(string)
		cmd := strings.TrimPrefix(step, "*:*:")
		if !strings.HasPrefix(cmd, "mkdir ") && !strings.HasPrefix(cmd, "cp ") {
			return false
		}
	}
	return true
}

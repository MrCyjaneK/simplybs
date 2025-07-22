package main

import (
	"flag"
	"fmt"
	"log"
	"path/filepath"
	"slices"
	"strings"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	argList := flag.Bool("list", false, "List all supported hosts")
	argHost := flag.String("host", "aarch64-apple-darwin", "The host to build for")
	argPkg := flag.String("package", "", "The package(s) to build (comma-separated)")
	argWorld := flag.Bool("world", false, "Build all packages")
	argExtract := flag.Bool("extract", false, "Extract packages")
	argDownload := flag.Bool("download", false, "Download package sources")
	flag.Parse()

	host := supportedHosts[*argHost]
	if host == nil {
		crashErr(fmt.Errorf("host %s not supported", *argHost))
	}

	packageNames := strings.Split(*argPkg, ",")

	for i, name := range packageNames {
		packageNames[i] = strings.TrimSpace(name)
	}

	var validPackages []string
	for _, name := range packageNames {
		if name != "" {
			validPackages = append(validPackages, name)
		}
	}

	if *argList {
		files, err := filepath.Glob("packages/*.json")
		if err != nil {
			crashErr(fmt.Errorf("failed to list packages: %v", err))
		}

		if !*argWorld && len(validPackages) == 0 {
			log.Fatalln("No -package or -world provided")
		}

		for _, file := range files {
			if !*argWorld && !slices.Contains(validPackages, filepath.Base(file)[:len(filepath.Base(file))-len(filepath.Ext(filepath.Base(file)))]) {
				continue
			}
			file = filepath.Base(file)
			pkgName := file[:len(file)-len(filepath.Ext(file))]
			printPackage(pkgName, *argHost, 0)
		}
		return
	}

	if *argWorld {
		files, err := filepath.Glob("packages/*.json")
		if err != nil {
			crashErr(fmt.Errorf("failed to list packages: %v", err))
		}
		for _, file := range files {
			file = filepath.Base(file)
			pkgName := file[:len(file)-len(filepath.Ext(file))]
			pkg := FindPackage(pkgName)
			if *argDownload {
				pkg.DownloadSource(host)
			} else {
				pkg.EnsureBuilt(host, true)
			}
		}
		if !*argExtract {
			return
		}
		for _, file := range files {
			file = filepath.Base(file)
			pkgName := file[:len(file)-len(filepath.Ext(file))]
			pkg := FindPackage(pkgName)
			pkg.ExtractEnv(host, host.GetEnvPath())
		}
		return
	}

	log.Printf("Building for host: %s", *argHost)

	if len(validPackages) == 0 {
		crashErr(fmt.Errorf("no valid package names provided"))
	}

	log.Printf("Building packages: %s", strings.Join(validPackages, ", "))

	for _, packageName := range validPackages {
		log.Printf("Processing package: %s", packageName)

		pkg := FindPackage(packageName)
		pkg.EnsureBuilt(host, true)
	}
	if !*argExtract {
		return
	}
	for _, packageName := range validPackages {
		log.Printf("Extracting env for package: %s", packageName)

		pkg := FindPackage(packageName)
		pkg.ExtractEnv(host, host.GetEnvPath())
	}
}

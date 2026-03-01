package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/mrcyjanek/simplybs/cmd/lint"
	"github.com/mrcyjanek/simplybs/crash"
	"github.com/mrcyjanek/simplybs/host"
	"github.com/mrcyjanek/simplybs/pack"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	argList := flag.Bool("list", false, "List all supported hosts (value is depth)")
	argHost := flag.String("host", "", "The host to build for")
	argPkg := flag.String("package", "", "The package(s) to build (comma-separated)")
	argWorld := flag.Bool("world", false, "Build all packages")
	argExtract := flag.Bool("extract", false, "Extract packages")
	argDownload := flag.Bool("download", false, "Download package sources")
	argBuild := flag.Bool("build", false, "Build packages")
	argLint := flag.Bool("lint", false, "Lint packages")
	argVersion := flag.Bool("v", false, "Show version")
	argShell := flag.Bool("shell", false, "Extract source and start shell with build environment")
	argCleanup := flag.Bool("cleanup", false, "Remove everything except current built archives")
	// argQuiet := flag.Bool("quiet", false, "Redirect stdout and stderr to /dev/null")
	flag.Parse()
	if *argVersion {
		fmt.Println("simplybs version 0.0.0")
		return
	}
	if *argCleanup {
		pack.Cleanup()
		return
	}
	if *argLint {
		lint.Lint()
		return
	}

	packageNames := []*pack.Package{}
	if *argWorld {
		packageNames = pack.GetAllPackages()
	} else {
		packageNames = pack.GetPackagesByList(*argPkg)
	}

	if len(packageNames) == 0 {
		crash.Handle(fmt.Errorf("no valid -package names or -world provided"))
	}
	if *argDownload {
		for _, pkg := range packageNames {
			for _, download := range pkg.Download {
				pkg.DownloadSource(download)
			}
		}
		log.Println("Downloaded all sources")
		return
	}

	hosts := strings.SplitSeq(*argHost, ",")
	for h := range hosts {
		host := host.SupportedHosts[h]
		if host == nil {
			crash.Handle(fmt.Errorf("host %s not supported", h))
		}
		buildForHost(host, packageNames, *argList, *argExtract, *argBuild, *argShell)
	}
}

func buildForHost(host *host.Host, packageNames []*pack.Package, list bool, extract bool, build bool, shell bool) {
	if list {
		for _, pkg := range packageNames {
			pack.PrintPackage(pkg.Package, host.Triplet)
		}
		return
	}

	if build {
		for _, pkg := range packageNames {
			pkg.EnsureBuilt(host, true)
		}
	}
	if extract {
		for _, pkg := range packageNames {
			log.Printf("Extracting env for package: %s", pkg.Package)
			pkg.ExtractEnv(host, host.GetEnvPath(), host.GetNativeEnvPath())
		}
	}

	if shell {
		if len(packageNames) != 1 {
			crash.Handle(fmt.Errorf("shell option requires exactly one package, got %d", len(packageNames)))
		}
		packageNames[0].StartShell(host)
	}
}

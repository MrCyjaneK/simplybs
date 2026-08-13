package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/mrcyjanek/simplybs/cmd/archive"
	"github.com/mrcyjanek/simplybs/cmd/lint"
	"github.com/mrcyjanek/simplybs/crash"
	"github.com/mrcyjanek/simplybs/host"
	"github.com/mrcyjanek/simplybs/pack"
)

func main() {
	log.SetFlags(log.Lshortfile)
	argList := flag.Bool("list", false, "List all supported hosts (value is depth)")
	argHost := flag.String("host", "", "The host to build for")
	argPkg := flag.String("package", "", "The package(s) to build (comma-separated)")
	argWorld := flag.Bool("world", false, "Build all packages")
	argExtract := flag.Bool("extract", false, "Extract packages")
	argDownload := flag.Bool("download", false, "Download package sources")
	argBuild := flag.Bool("build", false, "Build packages")
	argLint := flag.Bool("lint", false, "Lint packages")
	argArchive := flag.Bool("archive", false, "Download all sources from sources.json")
	argVersion := flag.Bool("v", false, "Show version")
	argShell := flag.Bool("shell", false, "Extract source and start shell with build environment")
	argCleanup := flag.Bool("cleanup", false, "Remove everything except current built archives")
	argCachePull := flag.Bool("cache-pull", false, "Download needed build-cache artifacts (requires SIMPLYBS_CACHE_TAG and SIMPLYBS_CACHE_REPO)")
	argCachePush := flag.Bool("cache-push", false, "Upload new/changed local build-cache artifacts (requires SIMPLYBS_CACHE_TAG and SIMPLYBS_CACHE_REPO)")
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
	if *argArchive {
		archive.Archive()
		return
	}

	packageNames := []*pack.Package{}
	if *argWorld {
		packageNames = pack.GetAllPackages()
	} else if *argPkg != "" {
		packageNames = pack.GetPackagesByList(*argPkg)
	}

	if *argCachePush && *argPkg == "" && !*argWorld {
		// Push every local built artifact for this builder that is not on the release.
		if err := pack.CachePush(nil, nil); err != nil {
			crash.Handle(err)
		}
		return
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

	if *argCachePull || *argCachePush {
		if *argHost == "" {
			crash.Handle(fmt.Errorf("-cache-pull/-cache-push with -package requires -host"))
		}
		hosts := strings.SplitSeq(*argHost, ",")
		for h := range hosts {
			host := host.SupportedHosts[h]
			if host == nil {
				crash.Handle(fmt.Errorf("host %s not supported", h))
			}
			if *argCachePull {
				if err := pack.CachePull(packageNames, host); err != nil {
					crash.Handle(err)
				}
			}
			if *argCachePush {
				if err := pack.CachePush(packageNames, host); err != nil {
					crash.Handle(err)
				}
			}
		}
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
		if pack.CacheEnabled() {
			log.Printf("cache: enabled (tag=%s repo=%s); auto pull/push", pack.CacheTag(), pack.CacheRepo())
			if err := pack.CachePull(packageNames, host); err != nil {
				log.Printf("cache: auto-pull failed: %v", err)
			}
		}
		for _, pkg := range packageNames {
			pkg.EnsureBuilt(host, true)
		}
		if pack.CacheEnabled() {
			if err := pack.CachePush(packageNames, host); err != nil {
				log.Printf("cache: auto-push failed: %v", err)
			}
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

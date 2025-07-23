package main

import (
	"flag"
	"fmt"
	"log"

	cmd "github.com/mrcyjanek/simplybs/cmd/buildweb"
	"github.com/mrcyjanek/simplybs/cmd/lint"
	"github.com/mrcyjanek/simplybs/crash"
	"github.com/mrcyjanek/simplybs/host"
	"github.com/mrcyjanek/simplybs/pack"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	argList := flag.Bool("list", false, "List all supported hosts")
	argHost := flag.String("host", "", "The host to build for")
	argPkg := flag.String("package", "", "The package(s) to build (comma-separated)")
	argWorld := flag.Bool("world", false, "Build all packages")
	argExtract := flag.Bool("extract", false, "Extract packages")
	argDownload := flag.Bool("download", false, "Download package sources")
	argBuild := flag.Bool("build", false, "Build packages")
	argBuildWeb := flag.Bool("buildweb", false, "Generate static website with package information")
	argLint := flag.Bool("lint", false, "Lint packages")
	argVersion := flag.Bool("v", false, "Show version")
	flag.Parse()
	if *argVersion {
		fmt.Println("simplybs version 0.0.0")
		return
	}
	if *argBuildWeb {
		cmd.BuildWeb()
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
			pkg.DownloadSource()
		}
		log.Println("Downloaded all sources")
		return
	}
	host := host.SupportedHosts[*argHost]
	if host == nil {
		crash.Handle(fmt.Errorf("host %s not supported", *argHost))
	}

	if *argList {
		for _, pkg := range packageNames {
			pack.PrintPackage(pkg.Package, *argHost, 0)
		}
		return
	}

	if *argExtract {
		for _, pkg := range packageNames {
			pkg.ExtractEnv(host, host.GetEnvPath())
		}
	}

	if *argBuild {
		for _, pkg := range packageNames {
			pkg.EnsureBuilt(host, true)
		}
	}
	if *argExtract {
		for _, pkg := range packageNames {
			log.Printf("Extracting env for package: %s", pkg.Package)
			pkg.ExtractEnv(host, host.GetEnvPath())
		}
	}
}

package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/ryanuber/go-glob"
)

func crashErr(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func printPackage(pkgName string, host string, depth int) {
	pkg := FindPackage(pkgName)
	fmt.Printf("%s%s: %s\n", strings.Repeat("  ", depth+1), pkg.Package, pkg.Version)
	for _, dep := range pkg.Dependencies {
		if strings.Contains(dep, ":") {
			prefix := strings.Split(dep, ":")[0]
			if !glob.Glob(prefix, host) && prefix != "all" {
				continue
			}
			dep = dep[strings.Index(dep, ":")+1:]
		}
		printPackage(dep, host, depth+1)
	}
}

package main

import (
	"fmt"
	"path"
	"sort"
	"strings"

	"github.com/mrcyjanek/simplybs/host"
	"github.com/mrcyjanek/simplybs/pack"
)

// DefaultHosts are the first targets the package worker builds.
var DefaultHosts = []string{
	"x86_64-linux-gnu",
	"aarch64-linux-android",
}

// Item is one (package, host) node in the build queue.
type Item struct {
	Package string `json:"package"`
	Host    string `json:"host"`
}

func (it Item) key() string {
	return it.Package + "\x00" + it.Host
}

// Result is JSON written to stdout for the Actions worker.
type Result struct {
	Status     string   `json:"status"` // next | done | blocked
	Package    string   `json:"package,omitempty"`
	Host       string   `json:"host,omitempty"`
	Needed     int      `json:"needed"`
	Remaining  []Item   `json:"remaining,omitempty"`
	ChangedPkg []string `json:"changed_packages,omitempty"`
	Message    string   `json:"message,omitempty"`
}

type cacheFn func(*pack.Package, *host.Host) (bool, error)

type queueOpts struct {
	changedFiles []string
	hosts        []string
	cached       cacheFn
	extraRoots   []string
}

func nextQueue(opts queueOpts) (Result, error) {
	hosts, err := resolveHosts(opts.hosts)
	if err != nil {
		return Result{}, err
	}
	cached := opts.cached
	if cached == nil {
		cached = pack.PackageCacheOnRelease
	}

	allPkgs := pack.GetAllPackages()
	names := make([]string, 0, len(allPkgs))
	byName := map[string]*pack.Package{}
	for _, p := range allPkgs {
		names = append(names, p.Package)
		byName[p.Package] = p
	}

	changed := packagesFromPaths(opts.changedFiles, names)
	for _, extra := range opts.extraRoots {
		extra = strings.TrimSpace(extra)
		if extra != "" {
			changed[extra] = true
		}
	}
	changedList := sortedKeys(changed)

	needed := map[string]Item{}
	for _, h := range hosts {
		ht := host.SupportedHosts[h]
		for name := range changed {
			root, ok := byName[name]
			if !ok {
				continue
			}
			for _, p := range pack.CollectNeededPackages([]*pack.Package{root}, ht) {
				it := Item{Package: p.Package, Host: h}
				onRelease, err := cached(p, ht)
				if err != nil {
					return Result{}, fmt.Errorf("cache lookup %s %s: %w", p.Package, h, err)
				}
				if onRelease {
					continue
				}
				needed[it.key()] = it
			}
		}
	}

	if len(needed) == 0 {
		return Result{
			Status:     "done",
			Needed:     0,
			ChangedPkg: changedList,
			Message:    "no uncached work for changed packages",
		}, nil
	}

	depSet := func(it Item) ([]string, error) {
		p, err := pack.FindPackage(it.Package)
		if err != nil {
			return nil, err
		}
		ht := host.SupportedHosts[it.Host]
		var deps []string
		for _, d := range pack.CollectNeededPackages([]*pack.Package{p}, ht) {
			if d.Package == p.Package {
				continue
			}
			deps = append(deps, d.Package)
		}
		return deps, nil
	}

	var ready []Item
	var blocked []Item
	for _, it := range needed {
		deps, err := depSet(it)
		if err != nil {
			return Result{}, err
		}
		ok := true
		for _, dep := range deps {
			depItem := Item{Package: dep, Host: it.Host}
			if _, still := needed[depItem.key()]; still {
				ok = false
				break
			}
		}
		if ok {
			ready = append(ready, it)
		} else {
			blocked = append(blocked, it)
		}
	}

	if len(ready) == 0 {
		return Result{
			Status:     "blocked",
			Needed:     len(needed),
			Remaining:  itemsSorted(needed, hosts, byName),
			ChangedPkg: changedList,
			Message:    fmt.Sprintf("needed %d items but none are ready (%d blocked)", len(needed), len(blocked)),
		}, nil
	}

	sortItems(ready, hosts, byName)
	pick := ready[0]
	delete(needed, pick.key())
	return Result{
		Status:     "next",
		Package:    pick.Package,
		Host:       pick.Host,
		Needed:     len(needed) + 1,
		Remaining:  itemsSorted(needed, hosts, byName),
		ChangedPkg: changedList,
	}, nil
}

func resolveHosts(hosts []string) ([]string, error) {
	if len(hosts) == 0 {
		hosts = append([]string{}, DefaultHosts...)
	}
	out := make([]string, 0, len(hosts))
	seen := map[string]bool{}
	for _, h := range hosts {
		h = strings.TrimSpace(h)
		if h == "" || seen[h] {
			continue
		}
		if host.SupportedHosts[h] == nil {
			return nil, fmt.Errorf("unsupported host %q", h)
		}
		seen[h] = true
		out = append(out, h)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no hosts")
	}
	return out, nil
}

func itemsSorted(m map[string]Item, hosts []string, byName map[string]*pack.Package) []Item {
	out := make([]Item, 0, len(m))
	for _, it := range m {
		out = append(out, it)
	}
	sortItems(out, hosts, byName)
	return out
}

func sortItems(items []Item, hosts []string, byName map[string]*pack.Package) {
	hostIdx := map[string]int{}
	for i, h := range hosts {
		hostIdx[h] = i
	}
	sort.Slice(items, func(i, j int) bool {
		a, b := items[i], items[j]
		ap, bp := byName[a.Package], byName[b.Package]
		aNative, bNative := ap != nil && ap.Type == "native", bp != nil && bp.Type == "native"
		if aNative != bNative {
			return aNative
		}
		if a.Package != b.Package {
			return a.Package < b.Package
		}
		return hostIdx[a.Host] < hostIdx[b.Host]
	})
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func packagesFromPaths(files, packageNames []string) map[string]bool {
	nameSet := map[string]bool{}
	for _, n := range packageNames {
		nameSet[n] = true
	}
	out := map[string]bool{}
	for _, f := range files {
		f = strings.TrimSpace(strings.ReplaceAll(f, "\\", "/"))
		if f == "" {
			continue
		}
		for _, name := range packagesAffectedByPath(f, packageNames, nameSet) {
			out[name] = true
		}
	}
	return out
}

func packagesAffectedByPath(file string, packageNames []string, nameSet map[string]bool) []string {
	file = strings.TrimPrefix(file, "./")
	if strings.HasPrefix(file, "packages/") && strings.HasSuffix(file, ".json") {
		name := strings.TrimSuffix(strings.TrimPrefix(file, "packages/"), ".json")
		if nameSet[name] {
			return []string{name}
		}
		return nil
	}
	if !strings.HasPrefix(file, "patches/") {
		return nil
	}
	rest := strings.TrimPrefix(file, "patches/")
	dir := rest
	if i := strings.IndexByte(rest, '/'); i >= 0 {
		dir = rest[:i]
	}
	if dir == "" {
		return nil
	}

	// patches/native/foo/... → native/foo (possibly nested)
	if dir == "native" {
		rel := strings.TrimPrefix(rest, "native/")
		if rel == "" || rel == rest {
			return matchPatchPrefix("native", packageNames)
		}
		// longest package-name prefix of the patch path
		best := ""
		candidate := "native/" + rel
		for _, name := range packageNames {
			if name == "native" || strings.HasPrefix(name, "native/") {
				prefix := name
				if candidate == prefix || strings.HasPrefix(candidate, prefix+"/") {
					if len(name) > len(best) {
						best = name
					}
				}
			}
		}
		if best != "" {
			return []string{best}
		}
		return matchPatchPrefix("native/"+strings.Split(rel, "/")[0], packageNames)
	}

	if nameSet[dir] {
		return []string{dir}
	}
	if nameSet["native/"+dir] {
		return []string{"native/" + dir}
	}
	return matchPatchPrefix(dir, packageNames)
}

func matchPatchPrefix(prefix string, packageNames []string) []string {
	var out []string
	for _, name := range packageNames {
		if name == prefix || strings.HasPrefix(name, prefix+"/") || strings.HasPrefix(name, prefix+"@") || path.Base(name) == prefix {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out
}

package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

type coord struct {
	group      string
	artifact   string
	classifier string
	version    string
}

func (c coord) String() string {
	if c.classifier != "" {
		return fmt.Sprintf("%s:%s:%s:%s", c.group, c.artifact, c.version, c.classifier)
	}
	return fmt.Sprintf("%s:%s:%s", c.group, c.artifact, c.version)
}

func (c coord) packageName() string {
	return fmt.Sprintf("native/jdk@22/bootstrap-maven/%s@%s", c.artifact, c.version)
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <dependency-tree.txt>\n", os.Args[0])
		os.Exit(1)
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "open: %v\n", err)
		os.Exit(1)
	}
	defer f.Close()

	lineRe := regexp.MustCompile(`([a-zA-Z0-9][a-zA-Z0-9._-]*):([a-zA-Z0-9][a-zA-Z0-9._-]*):([a-zA-Z0-9][a-zA-Z0-9._-]*)`)

	parseCoord := func(line string) (coord, bool) {
		idx := strings.IndexAny(line, "+-\\")
		if idx < 0 {
			return coord{}, false
		}
		rest := strings.TrimSpace(line[idx+1:])
		if strings.HasPrefix(rest, "- ") {
			rest = strings.TrimSpace(rest[2:])
		} else if strings.HasPrefix(rest, "+- ") {
			rest = strings.TrimSpace(rest[3:])
		}
		m := lineRe.FindStringSubmatch(rest)
		if m == nil {
			return coord{}, false
		}
		tail := strings.TrimPrefix(rest, m[0]+":")
		parts := strings.Split(tail, ":")
		if len(parts) < 2 {
			return coord{}, false
		}
		c := coord{group: m[1], artifact: m[2]}
		if len(parts) == 2 {
			c.version = parts[0]
			return c, true
		}
		if len(parts) == 3 {
			c.classifier = parts[0]
			c.version = parts[1]
			return c, true
		}
		return coord{}, false
	}

	all := map[coord]bool{}
	children := map[coord]bool{}
	reactor := map[coord]bool{}

	scanner := bufio.NewScanner(f)
	type stackEntry struct {
		indent int
		c      coord
	}
	var stack []stackEntry

	for scanner.Scan() {
		line := scanner.Text()
		c, ok := parseCoord(line)
		if !ok {
			continue
		}
		all[c] = true
		if c.group == "org.apache.maven" && c.version == "3.9.16" {
			reactor[c] = true
		}

		indent := 0
		for i := 0; i < len(line); i++ {
			ch := line[i]
			if ch == ' ' || ch == '|' || ch == '+' || ch == '-' || ch == '\\' {
				indent++
				continue
			}
			break
		}

		for len(stack) > 0 && stack[len(stack)-1].indent >= indent {
			stack = stack[:len(stack)-1]
		}
		if len(stack) > 0 {
			children[stack[len(stack)-1].c] = true
		}
		stack = append(stack, stackEntry{indent: indent, c: c})
	}

	external := []coord{}
	leaves := []coord{}
	for c := range all {
		if reactor[c] {
			continue
		}
		external = append(external, c)
		if !children[c] {
			leaves = append(leaves, c)
		}
	}
	sort.Slice(external, func(i, j int) bool { return external[i].String() < external[j].String() })
	sort.Slice(leaves, func(i, j int) bool { return leaves[i].String() < leaves[j].String() })

	fmt.Printf("reactor modules (org.apache.maven:*:3.9.16, built by native/maven@3.9): %d\n", len(reactor))
	fmt.Printf("external compile coordinates: %d\n", len(external))
	fmt.Printf("external leaves: %d\n\n", len(leaves))

	fmt.Println("== leaves (build these first) ==")
	for _, c := range leaves {
		fmt.Printf("%s\n  package: %s\n", c.String(), c.packageName())
	}

	fmt.Println("\n== all external ==")
	for _, c := range external {
		fmt.Println(c.String())
	}
}

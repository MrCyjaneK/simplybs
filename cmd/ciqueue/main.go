package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	cmd := os.Args[1]
	args := os.Args[2:]
	var err error
	switch cmd {
	case "next":
		err = cmdNext(args)
	case "comment":
		err = cmdComment(args)
	default:
		usage()
		os.Exit(2)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "ciqueue %s: %v\n", cmd, err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `usage:
  ciqueue next    [-base SHA] [-sha SHA] [-hosts list] [-changed-files PATH] [-package name]
  ciqueue comment [-pr N] [-sha SHA] [-package name] [-host triplet] [-conclusion text] [-run-url URL] [-remaining JSON]
`)
}

func cmdNext(args []string) error {
	fs := flag.NewFlagSet("next", flag.ContinueOnError)
	base := fs.String("base", "", "git base SHA (diff start)")
	sha := fs.String("sha", "", "git head SHA (diff end); default HEAD")
	hosts := fs.String("hosts", strings.Join(DefaultHosts, ","), "comma-separated -host triplets")
	changedFile := fs.String("changed-files", "", "file with changed paths, one per line (skips git)")
	extraPkg := fs.String("package", "", "extra dirty root (workflow_dispatch)")
	if err := fs.Parse(args); err != nil {
		return err
	}

	var files []string
	if *changedFile != "" {
		var err error
		files, err = readLines(*changedFile)
		if err != nil {
			return err
		}
	} else {
		end := *sha
		if end == "" {
			end = "HEAD"
		}
		if *base == "" {
			return fmt.Errorf("provide -base or -changed-files")
		}
		var err error
		files, err = gitChanged(*base, end)
		if err != nil {
			return err
		}
	}

	var extra []string
	if *extraPkg != "" {
		extra = strings.Split(*extraPkg, ",")
		for i := range extra {
			extra[i] = strings.TrimSpace(extra[i])
		}
	}

	res, err := nextQueue(queueOpts{
		changedFiles: files,
		hosts:        splitCSV(*hosts),
		extraRoots:   extra,
	})
	if err != nil {
		return err
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(res); err != nil {
		return err
	}
	return writeGitHubOutput(res)
}

func writeGitHubOutput(res Result) error {
	path := os.Getenv("GITHUB_OUTPUT")
	if path == "" {
		return nil
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	fmt.Fprintf(f, "status=%s\n", res.Status)
	fmt.Fprintf(f, "package=%s\n", res.Package)
	fmt.Fprintf(f, "host=%s\n", res.Host)
	rem, err := json.Marshal(res.Remaining)
	if err != nil {
		return err
	}
	fmt.Fprintf(f, "remaining<<EOF\n%s\nEOF\n", rem)
	return nil
}

func cmdComment(args []string) error {
	fs := flag.NewFlagSet("comment", flag.ContinueOnError)
	pr := fs.Int("pr", 0, "pull request number (0 = skip comment)")
	sha := fs.String("sha", "", "commit SHA")
	pkg := fs.String("package", "", "package just built (empty on start/done)")
	hostName := fs.String("host", "", "host triplet")
	conclusion := fs.String("conclusion", "", "success|failure|skipped|pending|done")
	runURL := fs.String("run-url", "", "Actions run URL")
	jobURL := fs.String("job-url", "", "optional job URL")
	remaining := fs.String("remaining", "", "JSON array of remaining items")
	repo := fs.String("repo", os.Getenv("GITHUB_REPOSITORY"), "owner/repo")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *pr <= 0 {
		fmt.Fprintln(os.Stderr, "ciqueue comment: no -pr, skipping")
		return nil
	}
	if *repo == "" {
		return fmt.Errorf("set GITHUB_REPOSITORY or -repo")
	}

	var rem []Item
	if strings.TrimSpace(*remaining) != "" {
		if err := json.Unmarshal([]byte(*remaining), &rem); err != nil {
			return fmt.Errorf("parse -remaining: %w", err)
		}
	}

	st, err := loadCommentState(*repo, *pr)
	if err != nil {
		return err
	}
	if st.SHA != "" && *sha != "" && st.SHA != *sha {
		st = CommentState{SHA: *sha, Runs: []CommentRun{}}
	}
	if st.SHA == "" {
		st.SHA = *sha
	}
	if *conclusion != "" {
		st.Runs = append(st.Runs, CommentRun{
			Package:    *pkg,
			Host:       *hostName,
			Conclusion: *conclusion,
			RunURL:     *runURL,
			JobURL:     *jobURL,
		})
	}
	st.Remaining = rem
	body := renderComment(st)
	return upsertComment(*repo, *pr, body)
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func readLines(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var lines []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, sc.Err()
}

func gitChanged(base, sha string) ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", "-z", base+"..."+sha)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("git diff %s...%s: %w: %s", base, sha, err, strings.TrimSpace(string(exitErr.Stderr)))
		}
		return nil, err
	}
	var files []string
	for _, p := range strings.Split(string(out), "\x00") {
		if p != "" {
			files = append(files, p)
		}
	}
	return files, nil
}

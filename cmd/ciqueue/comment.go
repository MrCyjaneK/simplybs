package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const commentMarker = "<!-- simplybs-ci-state"

func stateMarker(queue string) string {
	q := strings.TrimSpace(queue)
	if q == "" || q == "linux" {
		return commentMarker
	}
	return "<!-- simplybs-ci-" + q + "-state"
}

// CommentState is stored inside the sticky PR comment.
type CommentState struct {
	SHA       string       `json:"sha"`
	Runs      []CommentRun `json:"runs"`
	Remaining []Item       `json:"remaining,omitempty"`
}

// CommentRun is one slot/package attempt.
type CommentRun struct {
	Package    string `json:"package,omitempty"`
	Host       string `json:"host,omitempty"`
	Conclusion string `json:"conclusion"`
	RunURL     string `json:"run_url,omitempty"`
	JobURL     string `json:"job_url,omitempty"`
}

func renderComment(st CommentState, queue string) string {
	var b strings.Builder
	title := "## simplybs package queue"
	q := strings.TrimSpace(queue)
	if q != "" && q != "linux" {
		title = "## simplybs package queue (" + q + ")"
	}
	b.WriteString(title)
	b.WriteString("\n\n")
	if st.SHA != "" {
		fmt.Fprintf(&b, "SHA `%s`\n\n", st.SHA)
	}
	b.WriteString("| Package | Host | Status | Log |\n| --- | --- | --- | --- |\n")
	if len(st.Runs) == 0 {
		b.WriteString("| — | — | pending | — |\n")
	}
	for _, r := range st.Runs {
		pkg := r.Package
		if pkg == "" {
			pkg = "—"
		}
		h := r.Host
		if h == "" {
			h = "—"
		}
		log := "—"
		url := r.JobURL
		if url == "" {
			url = r.RunURL
		}
		if url != "" {
			log = "[run](" + url + ")"
		}
		fmt.Fprintf(&b, "| `%s` | `%s` | %s | %s |\n", pkg, h, r.Conclusion, log)
	}
	if len(st.Remaining) > 0 {
		fmt.Fprintf(&b, "\n%d still queued:\n", len(st.Remaining))
		max := 20
		for i, it := range st.Remaining {
			if i == max {
				fmt.Fprintf(&b, "- … %d more\n", len(st.Remaining)-max)
				break
			}
			fmt.Fprintf(&b, "- `%s` / `%s`\n", it.Package, it.Host)
		}
	}
	raw, _ := json.Marshal(st)
	b.WriteString("\n")
	b.WriteString(stateMarker(queue))
	b.WriteByte('\n')
	b.Write(raw)
	b.WriteString("\n-->\n")
	return b.String()
}

func loadCommentState(repo string, pr int, marker string) (CommentState, error) {
	cmd := ghAPI("repos/" + repo + "/issues/" + strconv.Itoa(pr) + "/comments?per_page=100")
	out, err := cmd.Output()
	if err != nil {
		return CommentState{}, fmt.Errorf("list comments: %w", formatCmdErr(err))
	}
	var comments []struct {
		ID   int64  `json:"id"`
		Body string `json:"body"`
	}
	if err := json.Unmarshal(out, &comments); err != nil {
		return CommentState{}, err
	}
	for _, c := range comments {
		if st, ok := parseState(c.Body, marker); ok {
			return st, nil
		}
	}
	return CommentState{}, nil
}

func parseState(body, marker string) (CommentState, bool) {
	if marker == "" {
		marker = commentMarker
	}
	i := strings.Index(body, marker)
	if i < 0 {
		return CommentState{}, false
	}
	rest := body[i+len(marker):]
	rest = strings.TrimSpace(rest)
	end := strings.Index(rest, "-->")
	if end < 0 {
		return CommentState{}, false
	}
	raw := strings.TrimSpace(rest[:end])
	var st CommentState
	if err := json.Unmarshal([]byte(raw), &st); err != nil {
		return CommentState{}, false
	}
	return st, true
}

func upsertComment(repo string, pr int, body, marker string) error {
	cmd := ghAPI("repos/" + repo + "/issues/" + strconv.Itoa(pr) + "/comments?per_page=100")
	out, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("list comments: %w", formatCmdErr(err))
	}
	var comments []struct {
		ID   int64  `json:"id"`
		Body string `json:"body"`
	}
	if err := json.Unmarshal(out, &comments); err != nil {
		return err
	}
	var existing int64
	for _, c := range comments {
		if strings.Contains(c.Body, marker) {
			existing = c.ID
			break
		}
	}
	payload, err := json.Marshal(map[string]string{"body": body})
	if err != nil {
		return err
	}
	if existing == 0 {
		post := ghAPI("-X", "POST", "repos/"+repo+"/issues/"+strconv.Itoa(pr)+"/comments", "--input", "-")
		post.Stdin = bytes.NewReader(payload)
		post.Stdout = os.Stdout
		post.Stderr = os.Stderr
		return post.Run()
	}
	patch := ghAPI("-X", "PATCH", "repos/"+repo+"/issues/comments/"+strconv.FormatInt(existing, 10), "--input", "-")
	patch.Stdin = bytes.NewReader(payload)
	patch.Stdout = os.Stdout
	patch.Stderr = os.Stderr
	return patch.Run()
}

func ghAPI(args ...string) *exec.Cmd {
	all := append([]string{"api"}, args...)
	return exec.Command("gh", all...)
}

func formatCmdErr(err error) error {
	if exitErr, ok := err.(*exec.ExitError); ok {
		return fmt.Errorf("%w: %s", err, strings.TrimSpace(string(exitErr.Stderr)))
	}
	return err
}

package utils

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

func isGitRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

func gitOutput(args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	return cmd.Output()
}

func CollectHistoricDownloads(s *SourcesFile) error {
	if !isGitRepo() {
		return fmt.Errorf("not inside a git repository")
	}

	output, err := gitOutput("log", "--all", "--format=%H")
	if err != nil {
		return fmt.Errorf("git log failed: %w", err)
	}

	commits := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(commits) == 1 && commits[0] == "" {
		return nil
	}

	log.Printf("Scanning %d commits for historic download entries...", len(commits))

	for i, commit := range commits {
		if commit == "" {
			continue
		}
		if (i+1)%10 == 0 || i == len(commits)-1 {
			log.Printf("  Historic scan: commit %d/%d", i+1, len(commits))
		}

		treeOutput, err := gitOutput("ls-tree", "-r", "--name-only", commit)
		if err != nil {
			continue
		}

		for _, filepath := range strings.Split(strings.TrimSpace(string(treeOutput)), "\n") {
			if filepath == "" || !strings.HasSuffix(filepath, ".json") {
				continue
			}

			content, err := gitOutput("show", commit+":"+filepath)
			if err != nil {
				continue
			}

			MergeDownloadsFromJSON(s, content)
		}
	}

	return nil
}

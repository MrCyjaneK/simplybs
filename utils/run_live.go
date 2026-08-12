package utils

import (
	"os"
	"os/exec"

	"golang.org/x/term"
)

// RunLive runs cmd with stdout/stderr streamed as they are produced.
//
// When our own stdout is not a terminal (piped, PowerShell 2>&1, CI capture),
// libc fully-buffers the child's stdout, so tools like configure look silent
// until a large flush. Merging both streams onto stderr avoids that in hosts
// that flush the error stream more eagerly (notably PowerShell).
//
// We intentionally do not allocate a PTY here: if the sink stops reading,
// a PTY writer can fill its buffer and deadlock the whole build step.
func RunLive(cmd *exec.Cmd) error {
	if term.IsTerminal(int(os.Stdout.Fd())) {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

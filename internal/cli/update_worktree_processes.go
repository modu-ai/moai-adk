package cli

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// activeProcessCWDs is a conservative process-level guard for worktree moves.
// The session registry does not cover every active shell or Codex process.
// When lsof is unavailable or fails, the migration skips all legacy trees.
var activeProcessCWDs = func() ([]string, error) {
	command := exec.Command("lsof", "-a", "-d", "cwd", "-F0n")
	out, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("lsof cwd scan: %w", err)
	}
	var directories []string
	for _, field := range bytes.Split(out, []byte{0}) {
		value := strings.TrimLeft(string(field), "\r\n")
		if strings.HasPrefix(value, "n/") {
			directories = append(directories, strings.TrimSuffix(value[1:], " (deleted)"))
		}
	}
	return directories, nil
}

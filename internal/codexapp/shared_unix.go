//go:build !windows

package codexapp

import (
	"os"
	"os/exec"
)

// Keep the same flock open-file description alive in the actual auth owner
// even if its supervisor is killed without running cleanup.
func inheritSharedLease(cmd *exec.Cmd, lease *os.File) { cmd.ExtraFiles = []*os.File{lease} }
